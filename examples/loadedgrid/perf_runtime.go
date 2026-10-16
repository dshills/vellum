package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/text/fonts"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type perfProgram struct {
	owner      *app.Application
	grid       *virtual.Grid
	source     *fullSource
	projection virtual.GridProjection
	callbacks  uint64 // owner only
}

func perfReleaseBuild() (map[string]string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("loadedgrid perf: unavailable build settings")
	}
	settings := map[string]string{}
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	if settings["-race"] == "true" || strings.Contains(settings["-gcflags"], "-N") || strings.Contains(settings["-gcflags"], "-l") {
		return nil, errors.New("loadedgrid perf requires an optimized release build without -race")
	}
	return settings, nil
}
func perfEmit(event string, value map[string]any) error {
	value["event"] = event
	return json.NewEncoder(os.Stdout).Encode(value)
}
func newPerfProgram(ctx context.Context, c perfConfig) (*perfProgram, map[string]any, error) {
	if err := c.validate(); err != nil {
		return nil, nil, err
	}
	settings, err := perfReleaseBuild()
	if err != nil {
		return nil, nil, err
	}
	fontBytes, err := fs.ReadFile(fonts.Default, "noto/NotoSans-Regular.ttf")
	if err != nil {
		return nil, nil, err
	}
	fontHash := sha256.Sum256(fontBytes)
	before := readPerfMemory()
	started := time.Now()
	d, err := generate(ctx, c.Rows, nil)
	if err != nil {
		return nil, nil, err
	}
	storedAt := time.Now()
	s, err := source("perf-stored", d, c.Rows)
	if err != nil {
		return nil, nil, err
	}
	projection, err := virtual.BuildGridProjection(ctx, perfInput(s, c.Variable), 2, virtual.GridQuery{}, limits())
	if err != nil {
		return nil, nil, err
	}
	projectedAt := time.Now()
	// Baseline retains the stored data/index but no mounted native controls.
	// GC is explicit outside native initialization and every measured frame.
	runtime.GC()
	afterData := readPerfMemory()
	a, err := app.New(app.Config{QueueCapacity: 64, CompletionCapacity: 8})
	if err != nil {
		return nil, nil, err
	}
	p := &perfProgram{owner: a, source: s, projection: projection}
	p.grid, err = virtual.NewGrid(projection, perfColumns(), virtual.GridConfig{RowHeight: 12, HeaderHeight: 24, FrozenColumns: 2, OverscanRows: 2, OverscanColumns: 1, MaxRows: 256, MaxCells: 4096}, virtual.GridViewport{Width: 768, Height: 624})
	if err != nil {
		a.Close()
		return nil, nil, err
	}
	err = a.Open(app.Window{ID: "perf", Title: "Vellum stored-grid performance", Width: 1500, Height: 1250, View: p.describe})
	if err != nil {
		a.Close()
		return nil, nil, err
	}
	metadata := map[string]any{"geometry_persistence_disabled_in_fixture": core.DebugSettings.DisableWindowGeometrySaver, "rows": c.Rows, "variable": c.Variable, "gpu_scene_rendering": c.GPUSceneRendering, "novel": c.Novel, "alloc_profile": c.AllocationProfile != "", "cpu_profile": c.CPUProfile != "", "source_details": c.SourceDetails, "input_details": c.InputDetails != "", "warmup_guard": 1, "cpu_profile_path": c.CPUProfile, "alloc_profile_prefix": c.AllocationProfile, "mem_profile_rate": runtime.MemProfileRate, "release": true, "seed": 7, "requested_font_family": "Noto Sans", "embedded_regular_font_sha256": hex.EncodeToString(fontHash[:]), "stored_value_bytes": s.logicalBytes(), "projection_stats": projection.Stats(), "memory_before_data": before, "memory_after_data_index_gc": afterData, "generate_ns": storedAt.Sub(started).Nanoseconds(), "projection_ns": projectedAt.Sub(storedAt).Nanoseconds(), "go_version": runtime.Version(), "build_settings": settings, "scope": "stored generation/projection excluded; GC baseline before native initialization; retained Go/process observations do not attribute native/GPU allocations"}
	return p, metadata, nil
}
func (p *perfProgram) describe() ui.Node {
	return ui.Node{ID: "root", Kind: ui.Column, Style: ui.Style{Padding: ui.Some(ui.Insets{}), Gap: ui.Some(float32(0))}, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Stored numeric table · explicit warm scroll trace"},
		{ID: "data", Kind: ui.DataGrid, Grid: p.grid, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 1299, Height: 1060}), MaxSize: ui.Some(ui.Size{Width: 1299, Height: 1060})}, GridChange: func(ui.GridEvent) { p.callbacks++ }, Accessibility: ui.Accessibility{Name: "Stored performance table"}},
	}}
}

func (p *perfProgram) owned(fn func() error) error {
	var result error
	if err := p.owner.Do(func() { result = fn() }); err != nil {
		return err
	}
	return result
}
func (p *perfProgram) inspect(ctx context.Context) (cogent.GridInspection, error) {
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		var measured cogent.GridInspection
		err := p.owned(func() error { var err error; measured, err = cogent.InspectGrid(p.owner, p.grid); return err })
		if err == nil {
			return measured, nil
		}
		if !errors.Is(err, cogent.ErrGridLayout) && !errors.Is(err, cogent.ErrGridClosed) {
			return measured, err
		}
		select {
		case <-ctx.Done():
			return measured, ctx.Err()
		case <-tick.C:
		}
	}
}
func (p *perfProgram) scroll(ctx context.Context, m *cogent.FrameMetrics, dx, dy int, tag uint64) (cogent.FrameSample, cogent.GridInspection, error) {
	var empty cogent.FrameSample
	err := p.owned(func() error { return cogent.EnqueueGridScroll(ctx, p.owner, p.grid, float64(dx), float64(dy), tag) })
	if err != nil {
		return empty, cogent.GridInspection{}, err
	}
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		poll := m.Poll(tag)
		if poll.Failure != "" {
			return empty, cogent.GridInspection{}, errors.New(poll.Failure)
		}
		if poll.Dropped != 0 {
			return empty, cogent.GridInspection{}, errors.New("loadedgrid perf: frame recorder overflow")
		}
		if poll.Found {
			s := poll.Sample
			if !s.Completed || s.Failure != "" {
				return s, cogent.GridInspection{}, fmt.Errorf("loadedgrid perf: incomplete/refused frame: %s", s.Failure)
			}
			g, err := p.inspect(ctx)
			if err != nil {
				return s, g, err
			}
			if g.Generation != s.Generation {
				return s, g, errors.New("loadedgrid perf: measured frame generation mismatch")
			}
			return s, g, nil
		}
		select {
		case <-ctx.Done():
			return empty, cogent.GridInspection{}, ctx.Err()
		case <-ticker.C:
		}
	}
}
func perfMetrics(s cogent.FrameSample) map[string]int64 {
	return map[string]int64{"ui_work_ns": s.TotalUIWorkNS, "input_ns": s.InputNS, "owner_ns": s.OwnerNS, "reconcile_ns": s.ReconcileNS, "layout_paint_ns": s.LayoutPaintNS, "gather_ns": s.GatherNS, "raster_upload_ns": s.RasterUploadNS, "encode_submit_ns": s.EncodeSubmitNS, "queued_ns": s.QueuedNS, "acquire_ns": s.AcquireNS, "present_call_ns": s.PresentCallNS}
}
func perfCountDelta(before, after countSnapshot) (countSnapshot, error) {
	if after.Keys < before.Keys || after.Inverse < before.Inverse || after.Cells < before.Cells {
		return countSnapshot{}, errors.New("loadedgrid perf: source counter overflow")
	}
	return countSnapshot{after.Keys - before.Keys, after.Inverse - before.Inverse, after.Cells - before.Cells}, nil
}
func (p *perfProgram) measure(ctx context.Context, m *cogent.FrameMetrics, c perfConfig) (result error) {
	initial, err := p.inspect(ctx)
	if err != nil {
		return err
	}
	if err = verifyPerfCells(p.source, initial); err != nil {
		return err
	}
	if _, err = checkPerfGeometry(initial); err != nil {
		return err
	}
	var warmStartCallbacks uint64
	if err = p.owned(func() error { warmStartCallbacks = p.callbacks; return nil }); err != nil {
		return err
	}
	warmFinal := initial
	for i := 0; i < c.Warmup; i++ {
		dx, dy := perfDelta(i)
		_, g, err := p.scroll(ctx, m, dx, dy, (1<<32)+uint64(i+1))
		if err != nil {
			return err
		}
		if _, err = checkPerfGeometry(g); err != nil {
			return err
		}
		warmFinal = g
	}
	var warmCallbacks uint64
	if err = p.owned(func() error { warmCallbacks = p.callbacks - warmStartCallbacks; return nil }); err != nil {
		return err
	}
	if err = checkPerfWarmup(initial, warmFinal, warmCallbacks, c.Warmup); err != nil {
		return err
	}
	if c.Novel {
		// A distinct, untouched source range. This setup frame is unmeasured; each
		// later vertical step introduces new row strings while retaining warm fonts.
		err = p.owned(func() error {
			if e := p.grid.JumpToKey(key(1000), column(0)); e != nil {
				return e
			}
			return p.owner.Invalidate("perf")
		})
		if err != nil {
			return err
		}
		if _, err = p.inspect(ctx); err != nil {
			return err
		}
	}
	// Explicit post-warm GC retained snapshot: excluded from frame timings.
	runtime.GC()
	warmMemory := readPerfMemory()
	var startCallbacks uint64
	err = p.owned(func() error { startCallbacks = p.callbacks; return nil })
	if err != nil {
		return err
	}
	if err = perfEmit("grid.perf.ready", map[string]any{"rows": c.Rows, "variable": c.Variable, "gpu_scene_rendering": c.GPUSceneRendering, "novel": c.Novel, "alloc_profile": c.AllocationProfile != "", "cpu_profile": c.CPUProfile != "", "source_details": c.SourceDetails, "input_details": c.InputDetails != "", "warmup_guard": 1, "warmup": c.Warmup, "memory_after_warm_gc": warmMemory, "initial_geometry": initial.Bounds, "warm_start_viewport": initial.Viewport, "warm_end_viewport": warmFinal.Viewport, "warm_start_generation": initial.Generation, "warm_end_generation": warmFinal.Generation, "warm_callbacks": warmCallbacks}); err != nil {
		return err
	}
	stopAllocation, err := startPerfAllocationProfile(c.AllocationProfile)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, stopAllocation()) }()
	stopProfile, err := startPerfCPUProfile(c.CPUProfile)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, stopProfile()) }()
	for i := 0; i < c.Samples; i++ {
		dx, dy := perfDelta(i)
		if c.Novel {
			dx, dy = 0, 12
		}
		beforeMemory := readPerfMemory()
		beforeCounts := p.source.counts.snapshot()
		var callbacks uint64
		if err = p.owned(func() error { callbacks = p.callbacks; return nil }); err != nil {
			return err
		}
		sample, g, err := p.scroll(ctx, m, dx, dy, uint64(i+1))
		if err != nil {
			return err
		}
		if err = verifyPerfCells(p.source, g); err != nil {
			return err
		}
		geometry, err := checkPerfGeometry(g)
		if err != nil {
			return err
		}
		counts, err := perfCountDelta(beforeCounts, p.source.counts.snapshot())
		if err != nil {
			return err
		}
		var changed uint64
		if err = p.owned(func() error { changed = p.callbacks - callbacks; return nil }); err != nil {
			return err
		}
		if changed != 1 {
			return fmt.Errorf("loadedgrid perf: scroll callback count%d", changed)
		}
		afterMemory := readPerfMemory()
		if afterMemory.TotalAlloc < beforeMemory.TotalAlloc || afterMemory.Mallocs < beforeMemory.Mallocs {
			return errors.New("loadedgrid perf: allocation counter overflow")
		}
		if err = perfEmit("grid.perf.sample", map[string]any{"gpu_scene_rendering": c.GPUSceneRendering, "alloc_profile": c.AllocationProfile != "", "source_details": c.SourceDetails, "input_details": c.InputDetails != "", "warmup_guard": 1, "index": i, "tag": sample.Tag, "generation": sample.Generation, "frame": g.Frames, "callbacks": changed, "geometry": geometry, "metrics": perfMetrics(sample), "native_sample": sample, "source_callbacks": counts, "go_process_alloc_bytes": afterMemory.TotalAlloc - beforeMemory.TotalAlloc, "go_process_mallocs": afterMemory.Mallocs - beforeMemory.Mallocs, "go_gc_cycles": afterMemory.NumGC - beforeMemory.NumGC}); err != nil {
			return err
		}
	}
	final := m.Snapshot()
	var callbacks uint64
	if err = p.owned(func() error { callbacks = p.callbacks - startCallbacks; return nil }); err != nil {
		return err
	}
	if final.Failure != "" || final.Pending != 0 || final.Dropped != 0 {
		return errors.New("loadedgrid perf: recorder not terminal")
	}
	if err = stopAllocation(); err != nil {
		return err
	}
	if err = stopProfile(); err != nil {
		return err
	}
	runtime.GC()
	return perfEmit("grid.perf.end", map[string]any{"rows": c.Rows, "variable": c.Variable, "gpu_scene_rendering": c.GPUSceneRendering, "novel": c.Novel, "alloc_profile": c.AllocationProfile != "", "cpu_profile": c.CPUProfile != "", "source_details": c.SourceDetails, "input_details": c.InputDetails != "", "warmup_guard": 1, "samples": c.Samples, "callbacks": callbacks, "dropped": final.Dropped, "pending": final.Pending, "memory_after_measured_gc": readPerfMemory(), "native_gpu_memory": "unavailable", "presentation_timestamps": "unavailable", "allocation_scope": "process-wide Go allocation between observations includes measurement helpers; not framework-only allocation attribution"})
}

// Detailed records are diagnostic instrumentation. The ordinary constructor
// and its larger capacity remain unchanged when the mode is off.
func newPerfFrameMetrics(c perfConfig) (*cogent.FrameMetrics, error) {
	capacity := c.Warmup + c.Samples + 2
	if c.SourceDetails {
		return cogent.NewFrameMetricsWithSourceDetails(capacity)
	}
	if c.InputDetails != "" {
		return cogent.NewFrameMetricsWithGridInputDetails(capacity)
	}
	return cogent.NewFrameMetrics(capacity)
}
func runPerf(c perfConfig) (terminal error) {
	// Disposable measurements must neither restore nor persist application
	// window geometry. Restore this process-local switch after native teardown.
	previousGeometrySaver := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	defer func() { core.DebugSettings.DisableWindowGeometrySaver = previousGeometrySaver }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	p, metadata, err := newPerfProgram(ctx, c)
	if err != nil {
		return err
	}
	defer p.owner.Close()
	m, err := newPerfFrameMetrics(c)
	if err != nil {
		return err
	}
	finishInput, err := startPerfInputDetails(c.InputDetails, m)
	if err != nil {
		return err
	}
	defer func() { terminal = errors.Join(terminal, finishInput(terminal)) }()
	if err = perfEmit("grid.perf.setup", metadata); err != nil {
		return err
	}
	failure := make(chan error, 1)
	d := cogent.Driver{FrameMetrics: m, GPUSceneRendering: c.GPUSceneRendering, Theme: ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 15}), Foreground: ui.Some(color.NRGBA{R: 20, G: 20, B: 20, A: 255}), Background: ui.Some(color.NRGBA{R: 250, G: 250, B: 250, A: 255})}}, OnError: func(err error) {
		select {
		case failure <- err:
		default:
		}
		cancel()
	}}
	result := make(chan error, 1)
	go func() { result <- p.measure(ctx, m, c); p.owner.Close() }()
	native := p.owner.Run(ctx, d)
	var measured error
	select {
	case measured = <-result:
	case <-ctx.Done():
		measured = ctx.Err()
	case <-time.After(2 * time.Second):
		cancel()
		measured = errors.New("loadedgrid perf: native runtime ended before measured completion")
	}
	select {
	case e := <-failure:
		native = errors.Join(native, e)
	default:
	}
	return errors.Join(native, measured)
}
