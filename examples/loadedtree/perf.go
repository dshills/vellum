package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

const treePerfPeriod = 32
const treePerfMinRows = 64

type treePerfConfig struct {
	Nodes, Samples, Warmup           int
	Deep                             bool
	GPUSceneRendering, SourceDetails bool
}

func (c treePerfConfig) validate() error {
	if (c.Nodes != 10_000 && c.Nodes != 100_000) || c.Samples < 120 || c.Samples > 2048 || c.Warmup < treePerfPeriod || c.Warmup > 2048 || c.Warmup%treePerfPeriod != 0 {
		return errors.New("loadedtree perf: nodes10000/100000, samples120..2048, warmup32..2048 in complete cycles required")
	}
	if c.SourceDetails && c.Warmup+c.Samples+2 > cogent.MaxSourceFrameMetricsCapacity {
		return errors.New("loadedtree perf: source diagnostics exceed bounded recorder capacity")
	}
	return nil
}
func treePerfDelta(i int) int {
	if i%treePerfPeriod < treePerfPeriod/2 {
		return 12
	}
	return -12
}

type treePerfGeometry struct {
	Viewport                                    virtual.OutlineViewport
	Bounds                                      image.Rectangle
	Scale                                       float32
	ClippedRows, MaterializedRows, RetainedRows int
	LayoutSHA256                                string
}

func treePerfGeometryCheck(g cogent.TreeInspection) (treePerfGeometry, error) {
	out := treePerfGeometry{Viewport: g.Viewport, Bounds: g.Bounds, Scale: g.Scale, MaterializedRows: len(g.Rows), RetainedRows: g.RetainedRows}
	if g.Pending || g.TotalNodes != g.VisibleRows || g.Generation == 0 || g.Frames == 0 || g.Scale <= 0 || g.Bounds.Empty() || g.RetainedRows > 500 || len(g.Rows) > 500 {
		return out, errors.New("loadedtree perf: invalid completed tree workload")
	}
	rectangles := make([]image.Rectangle, 0, len(g.Rows))
	seen := map[string]bool{}
	for _, row := range g.Rows {
		if row.Bounds.Empty() {
			continue
		}
		if seen[row.Key] || row.Bounds != row.Bounds.Intersect(g.Bounds) {
			return out, errors.New("loadedtree perf: duplicate row or escaped native clip")
		}
		seen[row.Key] = true
		out.ClippedRows++
		rectangles = append(rectangles, row.Bounds.Sub(g.Bounds.Min))
	}
	if out.ClippedRows < treePerfMinRows {
		return out, fmt.Errorf("loadedtree perf: only%dclipped rows; at least%drequired", out.ClippedRows, treePerfMinRows)
	}
	data, err := json.Marshal(rectangles)
	if err != nil {
		return out, err
	}
	hash := sha256.Sum256(data)
	out.LayoutSHA256 = hex.EncodeToString(hash[:])
	return out, nil
}

type treePerfProgram struct {
	owner     *app.Application
	outline   *virtual.Outline
	data      *virtual.TreeData
	callbacks uint64
	ready     bool // measurement worker only; a retired binding cannot become startup again
}

func (p *treePerfProgram) describe() ui.Node {
	return ui.Node{ID: "root", Kind: ui.Column, Style: ui.Style{Padding: ui.Some(ui.Insets{}), Gap: ui.Some(float32(0))}, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Stored tree · routed warm scrolling"},
		{ID: "data", Kind: ui.DataTree, Tree: p.outline, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 1299, Height: 1600}), MaxSize: ui.Some(ui.Size{Width: 1299, Height: 1600})}, TreeChange: func(ui.TreeEvent) { p.callbacks++ }},
	}}
}
func treePerfMemory() runtime.MemStats { var s runtime.MemStats; runtime.ReadMemStats(&s); return s }
func (c treePerfConfig) record(event string, value map[string]any) map[string]any {
	value["event"] = event
	value["gpu_scene_rendering"] = c.GPUSceneRendering
	value["source_details"] = c.SourceDetails
	return value
}
func (c treePerfConfig) emit(event string, value map[string]any) error {
	value = c.record(event, value)
	return json.NewEncoder(os.Stdout).Encode(value)
}

// Worker probes revoke queued work on timeout and join work already executing.
// The outer process watchdog bounds a broken runtime; callbacks never outlive
// resources merely because admission returned or the caller's deadline expired.
func (p *treePerfProgram) owned(ctx context.Context, fn func() error) error {
	reply := make(chan error, 1)
	var mu sync.Mutex
	phase := 0
	if err := p.owner.PostWait(ctx, func() {
		mu.Lock()
		if phase == 2 || ctx.Err() != nil {
			phase = 2
			mu.Unlock()
			reply <- ctx.Err()
			return
		}
		phase = 1
		mu.Unlock()
		result := fn()
		mu.Lock()
		phase = 3
		mu.Unlock()
		reply <- result
	}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		mu.Lock()
		if phase == 0 {
			phase = 2
			mu.Unlock()
			return ctx.Err()
		}
		running := phase == 1 || phase == 3
		mu.Unlock()
		if running {
			return errors.Join(<-reply, ctx.Err())
		}
		return ctx.Err()
	}
}
func (p *treePerfProgram) inspect(ctx context.Context) (cogent.TreeInspection, error) {
	// Missing initial native binding uses the same ErrTreeClosed as retirement.
	// Permit that startup condition for at most five seconds and only while the
	// originating logical window is live. Once observed, closure is terminal.
	if !p.ready {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}
	life, err := p.owner.WindowContext("perf")
	if err != nil {
		return cogent.TreeInspection{}, err
	}
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if err := life.Err(); err != nil {
			return cogent.TreeInspection{}, err
		}
		var g cogent.TreeInspection
		err := p.owned(ctx, func() error { var e error; g, e = cogent.InspectTree(p.owner, p.outline); return e })
		if err == nil {
			p.ready = true
			return g, nil
		}
		if !errors.Is(err, cogent.ErrTreeLayout) && (p.ready || !errors.Is(err, cogent.ErrTreeClosed)) {
			return g, err
		}
		select {
		case <-ctx.Done():
			return g, ctx.Err()
		case <-ticker.C:
		}
	}
}
func (p *treePerfProgram) scroll(ctx context.Context, m *cogent.FrameMetrics, dy int, tag uint64) (cogent.FrameSample, cogent.TreeInspection, error) {
	var empty cogent.FrameSample
	before, err := p.inspect(ctx)
	if err != nil {
		return empty, before, err
	}
	if err := p.owned(ctx, func() error { return cogent.EnqueueTreeScroll(ctx, p.owner, p.outline, float64(dy), tag) }); err != nil {
		return empty, cogent.TreeInspection{}, err
	}
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		poll := m.Poll(tag)
		if poll.Failure != "" || poll.Dropped != 0 {
			return empty, cogent.TreeInspection{}, errors.New("loadedtree perf: invalid or overflowing recorder")
		}
		if poll.Found {
			sample := poll.Sample
			if !sample.Completed || sample.Failure != "" {
				return sample, cogent.TreeInspection{}, fmt.Errorf("loadedtree perf: refused frame: %s", sample.Failure)
			}
			g, err := p.inspect(ctx)
			if err == nil && (g.Generation != sample.Generation || g.Generation != before.Generation+1 || g.Viewport.Width != before.Viewport.Width || g.Viewport.Height != before.Viewport.Height || g.Viewport.ScrollY != before.Viewport.ScrollY+dy || g.Bounds != before.Bounds || g.Scale != before.Scale || g.Stamp != before.Stamp) {
				err = errors.New("loadedtree perf: event offset/generation/geometry drift")
			}
			return sample, g, err
		}
		select {
		case <-ctx.Done():
			return empty, cogent.TreeInspection{}, ctx.Err()
		case <-ticker.C:
		}
	}
}
func (p *treePerfProgram) callbackCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := p.owned(ctx, func() error { count = p.callbacks; return nil })
	return count, err
}
func (p *treePerfProgram) measure(ctx context.Context, m *cogent.FrameMetrics, c treePerfConfig) error {
	initial, err := p.inspect(ctx)
	if err != nil {
		return err
	}
	if err = c.emit("tree.perf.geometry", map[string]any{"initial": initial}); err != nil {
		return err
	}
	geometry, err := treePerfGeometryCheck(initial)
	if err != nil {
		return err
	}
	startCallbacks, err := p.callbackCount(ctx)
	if err != nil {
		return err
	}
	// Keep the first scrolling frame separate from the measured warm trajectory.
	// Initial scene/fonts/data already exist; this is not cold-cache evidence.
	cold, coldGeometry, err := p.scroll(ctx, m, 12, 1)
	if err != nil {
		return err
	}
	if _, err = treePerfGeometryCheck(coldGeometry); err != nil {
		return err
	}
	if err = c.emit("tree.perf.first-scroll", map[string]any{"nodes": c.Nodes, "deep": c.Deep, "frame": cold, "geometry": coldGeometry}); err != nil {
		return err
	}
	if _, _, err = p.scroll(ctx, m, -12, 2); err != nil {
		return err
	}
	for i := 0; i < c.Warmup; i++ {
		if _, _, err = p.scroll(ctx, m, treePerfDelta(i), uint64(i+3)); err != nil {
			return err
		}
	}
	warm, err := p.inspect(ctx)
	if err != nil {
		return err
	}
	warmGeometry, err := treePerfGeometryCheck(warm)
	if err != nil {
		return err
	}
	count, err := p.callbackCount(ctx)
	if err != nil {
		return err
	}
	if warm.Viewport != initial.Viewport || warm.Bounds != initial.Bounds || warm.Scale != initial.Scale || warmGeometry.LayoutSHA256 != geometry.LayoutSHA256 || count-startCallbacks != uint64(c.Warmup+2) {
		return errors.New("loadedtree perf: warmup offset/geometry/callback drift")
	}
	runtime.GC()
	warmMemory := treePerfMemory()
	if err = c.emit("tree.perf.ready", map[string]any{"nodes": c.Nodes, "deep": c.Deep, "warmup": c.Warmup, "geometry": warmGeometry, "memory": warmMemory}); err != nil {
		return err
	}
	for i := 0; i < c.Samples; i++ {
		before := treePerfMemory()
		beforeCount, err := p.callbackCount(ctx)
		if err != nil {
			return err
		}
		sample, g, err := p.scroll(ctx, m, treePerfDelta(i), uint64(c.Warmup+i+3))
		if err != nil {
			return err
		}
		measured, err := treePerfGeometryCheck(g)
		if err != nil {
			return err
		}
		afterCount, err := p.callbackCount(ctx)
		if err != nil {
			return err
		}
		after := treePerfMemory()
		if afterCount-beforeCount != 1 || after.TotalAlloc < before.TotalAlloc || after.Mallocs < before.Mallocs {
			return errors.New("loadedtree perf: extra callback or allocation counter drift")
		}
		if err = c.emit("tree.perf.sample", map[string]any{"nodes": c.Nodes, "deep": c.Deep, "index": i, "phase": i % treePerfPeriod, "frame": sample, "geometry": measured, "tree_notifications": afterCount - beforeCount, "allocation_bytes": after.TotalAlloc - before.TotalAlloc, "mallocs": after.Mallocs - before.Mallocs, "gc_count": after.NumGC - before.NumGC}); err != nil {
			return err
		}
	}
	snapshot := m.Snapshot()
	if snapshot.Pending != 0 || snapshot.Dropped != 0 || snapshot.Failure != "" || len(snapshot.Samples) != c.Samples+c.Warmup+2 {
		return errors.New("loadedtree perf: missing/extra/pending frame tickets")
	}
	for _, s := range snapshot.Samples {
		if !s.Completed || s.Failure != "" {
			return errors.New("loadedtree perf: terminal recorder contains refused frame")
		}
	}
	runtime.GC()
	return c.emit("tree.perf.finished", map[string]any{"nodes": c.Nodes, "deep": c.Deep, "samples": c.Samples, "frame_tickets": len(snapshot.Samples), "memory": treePerfMemory(), "full_p6_qualified": false})
}
func runTreePerf(c treePerfConfig) error {
	if err := c.validate(); err != nil {
		return err
	}
	core.DebugSettings.DisableWindowGeometrySaver = true
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("loadedtree perf: release build metadata unavailable")
	}
	settings := map[string]string{}
	for _, s := range build.Settings {
		settings[s.Key] = s.Value
	}
	if settings["-race"] == "true" {
		return errors.New("loadedtree perf: optimized non-race build required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	before := treePerfMemory()
	start := time.Now()
	data, err := makeData(ctx, "tree-perf", c.Nodes)
	if err != nil {
		return err
	}
	dataAt := time.Now()
	index, err := virtual.BuildTreeIndex(ctx, data, virtual.TreeIndexOptions{ExpandAll: true})
	if err != nil {
		return err
	}
	indexAt := time.Now()
	outline, err := virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 9, MaxRows: 500}, virtual.OutlineViewport{Width: 768, Height: 768})
	if err != nil {
		return err
	}
	if c.Deep {
		err = outline.SetViewport(virtual.OutlineViewport{Width: 768, Height: 768, ScrollY: (c.Nodes*3/4 + 64) * 9})
		if err != nil {
			return err
		}
	}
	runtime.GC()
	afterData := treePerfMemory()
	owner, err := app.New(app.Config{QueueCapacity: 64, CompletionCapacity: 8})
	if err != nil {
		return err
	}
	defer owner.Close()
	p := &treePerfProgram{owner: owner, outline: outline, data: data}
	if err = owner.Open(app.Window{ID: "perf", Title: "Vellum stored-tree performance", Width: 1500, Height: 1100, View: p.describe}); err != nil {
		return err
	}
	font, err := fs.ReadFile(fonts.Default, "noto/NotoSans-Regular.ttf")
	if err != nil {
		return err
	}
	hash := sha256.Sum256(font)
	if err = c.emit("tree.perf.setup", map[string]any{"nodes": c.Nodes, "deep": c.Deep, "data": data.Stats(), "index": index.Stats(), "construction_ns": dataAt.Sub(start).Nanoseconds(), "index_ns": indexAt.Sub(dataAt).Nanoseconds(), "font_sha256": hex.EncodeToString(hash[:]), "go_version": runtime.Version(), "build": settings, "memory_before_data": before, "memory_after_data": afterData, "target_p95_ns": int64(4_000_000), "timing": "elapsed CPU-side work regions; no thread CPU/GPU/presentation attribution"}); err != nil {
		return err
	}
	newMetrics := cogent.NewFrameMetrics
	if c.SourceDetails {
		newMetrics = cogent.NewFrameMetricsWithSourceDetails
	}
	metrics, err := newMetrics(c.Warmup + c.Samples + 2)
	if err != nil {
		return err
	}
	result := make(chan error, 1)
	failures := make(chan error, 1)
	driver := cogent.Driver{GPUSceneRendering: c.GPUSceneRendering, FrameMetrics: metrics, Theme: ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 12}), Foreground: ui.Some(color.NRGBA{R: 20, G: 20, B: 20, A: 255}), Background: ui.Some(color.NRGBA{R: 250, G: 250, B: 250, A: 255})}}, OnError: func(e error) {
		select {
		case failures <- e:
		default:
		}
		cancel()
	}}
	go func() { err := p.measure(ctx, metrics, c); result <- err; owner.Close() }()
	native := owner.Run(ctx, driver)
	var measured error
	select {
	case measured = <-result:
	case <-time.After(2 * time.Second):
		cancel()
		measured = errors.New("loadedtree perf: runtime ended before worker completion")
	}
	select {
	case e := <-failures:
		native = errors.Join(native, e)
	default:
	}
	return errors.Join(native, measured)
}
