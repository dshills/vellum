//go:build darwin && cgo && !offscreen && vellum_native_checks

// window-cycle-smoke repeatedly mounts controls and a GPU canvas in a fresh
// native window, verifies readback, and observes teardown with a survivor alive.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/system"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/acceptance"
	"github.com/dshills/vellum/drivers/cogent/internal/nativefixture"
	"github.com/dshills/vellum/ui"
)

type config struct {
	cycles, every int
	timeout       time.Duration
	heapProfile   string
	composed      bool
	sameName      bool
	soak          soakConfig
}

func (c config) validate() error {
	if c.heapProfile != "" && (!filepath.IsAbs(c.heapProfile) || filepath.Clean(c.heapProfile) != c.heapProfile) {
		return errors.New("heap-profile needs a clean absolute new file path")
	}
	if err := validateSameNameMode(c.sameName, c.composed, c.cycles, c.soak.Duration); err != nil {
		return err
	}
	if c.soak.Duration != 0 {
		return c.soak.validate(c)
	}
	maximum := 20 * time.Minute
	if c.composed {
		maximum = 4 * time.Hour
	}
	if c.cycles < 1 || c.cycles > 1000 || c.every < 1 || c.every > c.cycles || c.timeout < 10*time.Second || c.timeout > maximum {
		return errors.New("cycles 1..1000, sample-every 1..cycles, timeout 10s..20m (4h for composed) required")
	}
	return nil
}

type observation struct {
	Completed                                               int
	HeapAlloc, HeapInuse, GoSys, TotalAlloc, Mallocs, Frees uint64
	NumGC                                                   uint32
	Goroutines                                              int
	RSSBytes                                                uint64
	ElapsedSeconds                                          float64
}
type progress struct {
	sync.Mutex
	completed int
	samples   []observation
}

func (p *progress) record(s observation) error {
	p.Lock()
	defer p.Unlock()
	p.samples = append(p.samples, s)
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.window-cycle.sample", "sample": s})
}
func (p *progress) advance() { p.Lock(); defer p.Unlock(); p.completed++ }
func (p *progress) snapshot() (int, []observation) {
	p.Lock()
	defer p.Unlock()
	return p.completed, append([]observation(nil), p.samples...)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	var c config
	flag.IntVar(&c.cycles, "cycles", 1000, "completed native open/readback/close cycles, 1..1000")
	flag.IntVar(&c.every, "sample-every", 25, "post-GC Go/RSS observation interval")
	flag.DurationVar(&c.timeout, "timeout", 12*time.Minute, "cooperative native loop deadline")
	flag.StringVar(&c.heapProfile, "heap-profile", "", "optional new absolute diagnostic heap-profile path (never overwrites)")
	flag.BoolVar(&c.composed, "composed", false, "mount and exercise the all-component workspace with 1000 dense labels in each fresh cycle window")
	flag.BoolVar(&c.sameName, "same-name", false, "reopen one logical name across composed cycles, verifying retired references and resources; requires at least two cycles and no soak")
	flag.DurationVar(&c.soak.Duration, "soak", 0, "total visible/occluded/minimized/background soak duration; requires composed, cycles=1, sample-every=1")
	flag.DurationVar(&c.soak.Quiet, "soak-quiet", time.Minute, "quiet CPU/paint observation interval")
	flag.DurationVar(&c.soak.Settle, "soak-settle", 10*time.Second, "settling interval before each quiet sample")
	flag.DurationVar(&c.soak.Interval, "soak-interval", 5*time.Minute, "minimum spacing of quiet samples within a phase")
	flag.Parse()
	if err := c.validate(); err != nil {
		return err
	}
	// Disposable unique titles must not write 1,000 geometry records into the
	// user's application preferences. This is local fixture configuration.
	restoreGeometry, err := nativefixture.IsolateGeometry()
	if err != nil {
		return err
	}
	defer restoreGeometry()
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	remove, err := cogent.ObserveNativeServiceCleanup()
	if err != nil {
		return err
	}
	defer remove()
	survivor, model, shown, err := open(a, "survivor", "Vellum cycle survivor")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	p := &progress{}
	finished := make(chan error, 1)
	failures := make(chan error, 1)
	go func() {
		var e error
		if c.soak.Duration != 0 {
			e = checkSoak(ctx, a, survivor, model, shown, c, p)
		} else {
			e = check(ctx, a, survivor, model, shown, c, p)
		}
		if c.heapProfile != "" {
			e = errors.Join(e, writeHeapProfile(c.heapProfile))
		}
		finished <- e
		a.Close()
	}()
	runErr := a.Run(ctx, cogent.Driver{OnError: func(e error) {
		select {
		case failures <- e:
		default:
		}
	}})
	cancel()
	select {
	case err = <-finished:
	case <-time.After(6 * time.Second):
		err = errors.New("window-cycle worker did not join after native exit")
	}
	select {
	case e := <-failures:
		err = errors.Join(err, e)
	default:
	}
	done, samples := p.snapshot()
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 {
		err = errors.Join(err, fmt.Errorf("native terminal inventory: observed=%t state=%+v", observed, cleanup))
	}
	if done != c.cycles {
		err = errors.Join(err, fmt.Errorf("completed %d of %d cycles", done, c.cycles))
	}
	result := "FAILED_NATIVE_WINDOW_CYCLES"
	if errors.Join(runErr, err) == nil {
		result = "PASS_NATIVE_WINDOW_CYCLES"
	}
	report := map[string]any{"event": "vellum.window-cycle.result", "result": result, "requested_cycles": c.cycles, "completed_cycles": done, "samples": samples, "native_cleanup": cleanup, "heap_profile_requested": c.heapProfile != "", "rss_definition": "ps resident pages, bytes; not private footprint or native allocation attribution", "samples_after_forced_gc": true, "rss_guard_bytes": uint64(1 << 30), "go_heap_guard_bytes": uint64(512 << 20), "geometry_persistence_disabled_in_fixture": true, "historical_id_tombstones_retained": false, "same_name_mode": c.sameName, "same_name_replacements_verified": func() int {
		if c.sameName {
			return max(0, done-1)
		}
		return 0
	}(), "full_p11_qualified": false, "native_heap_qualified": false, "gpu_memory_qualified": false, "clipboard_access": false, "preferences_changed": false, "cooperative_timeout_seconds": c.timeout.Seconds(), "composed_workspace": c.composed, "native_action_plan": func() int {
		if c.composed {
			return nativeActionPlanVersion
		}
		return 0
	}(), "dense_labels_per_composed_window": func() int {
		if c.composed {
			return acceptance.DenseLabels
		}
		return 0
	}()}
	if c.soak.Duration != 0 {
		report["event"] = "vellum.window-soak.result"
		report["soak_config"] = c.soak
		report["result"] = "FAILED_NATIVE_WINDOW_SOAK"
		if result == "PASS_NATIVE_WINDOW_CYCLES" {
			report["result"] = "PASS_NATIVE_WINDOW_SOAK"
		}
	}
	encodeErr := json.NewEncoder(os.Stdout).Encode(report)
	return errors.Join(runErr, err, encodeErr)
}

type profileWriter struct {
	file      *os.File
	remaining int
}

func (w *profileWriter) Write(data []byte) (int, error) {
	if len(data) > w.remaining {
		return 0, errors.New("heap-profile exceeds 8 MiB diagnostic limit")
	}
	n, err := w.file.Write(data)
	w.remaining -= n
	return n, err
}
func writeHeapProfile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	runtime.GC()
	err = pprof.WriteHeapProfile(&profileWriter{file: f, remaining: 8 << 20})
	return errors.Join(err, f.Close())
}

type readyCanvas struct {
	renderer *cogent.GPUCanvas
	shown    *atomic.Bool
}

func (p *readyCanvas) Mount(parent tree.Node) core.Widget {
	w := p.renderer.Mount(parent)
	if w != nil {
		w.AsWidget().OnShow(func(events.Event) { p.shown.Store(true) })
	}
	return w
}
func (p *readyCanvas) Sync(w core.Widget) error { return p.renderer.Sync(w) }

func open(a *app.Application, id, title string) (*cogent.GPUCanvas, *canvas.Scene, *atomic.Bool, error) {
	g, m, shown, _, _, err := openMode(a, id, title, false)
	return g, m, shown, err
}
func openMode(a *app.Application, id, title string, composed bool) (*cogent.GPUCanvas, *canvas.Scene, *atomic.Bool, *acceptance.Workspace, app.WindowRef, error) {
	m, err := canvas.New([]canvas.Rect{{ID: "probe", X: 0, Y: 0, Width: 100, Height: 60, Color: color.RGBA{67, 197, 166, 255}}})
	if err != nil {
		return nil, nil, nil, nil, app.WindowRef{}, err
	}
	m.Fit()
	g := &cogent.GPUCanvas{Owner: a, Window: id, Model: m}
	shown := &atomic.Bool{}
	text := "controlled field"
	checked := false
	var suite *acceptance.Workspace
	if composed {
		suite, err = acceptance.New(a, id, &readyCanvas{g, shown})
		if err != nil {
			return nil, nil, nil, nil, app.WindowRef{}, err
		}
	}
	width, height := float32(480), float32(420)
	if composed {
		width, height = 1920, 1080
	}
	ref, err := a.OpenWindow(app.Window{ID: id, Title: title, Width: width, Height: height, CloseRequest: closeRequestDiagnostic(id, os.Stdout), View: func() ui.Node {
		if suite != nil {
			return suite.View()
		}
		children := []ui.Node{{ID: "label", Kind: ui.Label, Text: title}, {ID: "field", Kind: ui.TextField, Text: text, Change: func(e ui.ChangeEvent) { text = e.Text }}, {ID: "check", Kind: ui.Checkbox, Text: "Controlled flag", Checked: checked, Change: func(e ui.ChangeEvent) { checked = e.Checked }}, {ID: "button", Kind: ui.Button, Text: "Temporary action", Activate: func() {}}}
		children = append(children, ui.Node{ID: "canvas", Kind: ui.Extension, Grow: true, Custom: &readyCanvas{g, shown}})
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: children}
	}})
	return g, m, shown, suite, ref, err
}
func wait(ctx context.Context, check func(context.Context) (bool, error)) error {
	return waitWithin(ctx, 5*time.Second, check)
}
func waitWithin(ctx context.Context, duration time.Duration, check func(context.Context) (bool, error)) error {
	deadline, stop := context.WithTimeout(ctx, duration)
	defer stop()
	for {
		ok, err := check(deadline)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-deadline.Done():
			return deadline.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func readback(ctx context.Context, a *app.Application, g *cogent.GPUCanvas, m *canvas.Scene, shown *atomic.Bool) (system.Window, error) {
	var native system.Window
	err := wait(ctx, func(inner context.Context) (bool, error) {
		if !shown.Load() {
			return false, nil
		}
		value, err := ownerProbe(inner, a, func() ownerReply {
			img, e := g.Snapshot()
			if e != nil {
				return ownerReply{err: e}
			}
			if img == nil || img.Bounds().Min != (image.Point{}) || img.Bounds().Size() != image.Pt(int(m.View.Width), int(m.View.Height)) {
				return ownerReply{err: errors.New("completed GPU image/viewport mismatch")}
			}
			point := m.View.Screen(canvas.Point{X: 50, Y: 30})
			pixel := image.Pt(int(point.X), int(point.Y))
			if !pixel.In(img.Bounds()) {
				return ownerReply{err: errors.New("GPU probe outside completed image")}
			}
			v := img.NRGBAAt(pixel.X, pixel.Y)
			if v.G < 100 || v.R > 150 || v.B < 80 {
				return ownerReply{err: fmt.Errorf("GPU scene probe missing: %v", v)}
			}
			return ownerReply{native: g.NativeWindow()}
		})
		if err != nil {
			return false, err
		}
		if errors.Is(value.err, cogent.ErrFrameNotReady) {
			return false, nil
		}
		native = value.native
		return native != nil, value.err
	})
	return native, err
}
func inventory(ctx context.Context, want []string) error {
	return wait(ctx, func(inner context.Context) (bool, error) {
		s, err := cogent.CheckNativeServices(inner, cogent.NativeCheckSnapshot, 0)
		if err != nil {
			return false, err
		}
		titles := slices.Clone(s.WindowTitles)
		slices.Sort(titles)
		expected := slices.Clone(want)
		slices.Sort(expected)
		return s.MainThread && s.NativeOutstanding == 0 && s.AccessibilityWindows == len(want) && slices.Equal(titles, expected), nil
	})
}
func sample(ctx context.Context, completed int, started time.Time) (observation, error) {
	// Forced GC is diagnostic and intentionally changes timing. RSS includes
	// caches/driver pages and cannot identify native heap or GPU deallocation.
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	limit, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	data, err := exec.CommandContext(limit, "/bin/ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		return observation{}, err
	}
	kib, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || kib > ^uint64(0)/1024 {
		return observation{}, errors.New("invalid self RSS observation")
	}
	return observation{Completed: completed, HeapAlloc: m.HeapAlloc, HeapInuse: m.HeapInuse, GoSys: m.Sys, TotalAlloc: m.TotalAlloc, Mallocs: m.Mallocs, Frees: m.Frees, NumGC: m.NumGC, Goroutines: runtime.NumGoroutine(), RSSBytes: kib * 1024, ElapsedSeconds: time.Since(started).Seconds()}, nil
}
func check(ctx context.Context, a *app.Application, survivor *cogent.GPUCanvas, model *canvas.Scene, shown *atomic.Bool, c config, p *progress) error {
	started := time.Now()
	if _, err := readback(ctx, a, survivor, model, shown); err != nil {
		return err
	}
	if err := inventory(ctx, []string{"Vellum cycle survivor"}); err != nil {
		return err
	}
	first, err := sample(ctx, 0, started)
	if err != nil {
		return err
	}
	if err = p.record(first); err != nil {
		return err
	}
	if err = memoryGuard(first); err != nil {
		return err
	}
	var previous *retiredNativeWindow
	for i := 0; i < c.cycles; i++ {
		previous, err = cycle(ctx, a, i, c, previous)
		if err != nil {
			return fmt.Errorf("cycle %d: %w", i, err)
		}
		if i+1 == c.cycles {
			previous = nil
		} // No final probe retention in the terminal sample.
		p.advance()
		if (i+1)%c.every == 0 || i+1 == c.cycles {
			if _, err = readback(ctx, a, survivor, model, shown); err != nil {
				return fmt.Errorf("survivor after %d: %w", i+1, err)
			}
			s, err := sample(ctx, i+1, started)
			if err != nil {
				return err
			}
			if err = p.record(s); err != nil {
				return err
			}
			if err = memoryGuard(s); err != nil {
				return err
			}
		}
	}
	return nil
}
func cycle(ctx context.Context, a *app.Application, i int, c config, previous *retiredNativeWindow) (*retiredNativeWindow, error) {
	id := cycleWindowID(i, c.sameName)
	title := "Vellum " + id
	g, m, shown, suite, ref, err := openMode(a, id, title, c.composed)
	if err != nil {
		return nil, err
	}
	original := ref.Context()
	child, stop := context.WithCancel(original)
	defer stop()
	var cleaned atomic.Int32
	var badCleanup atomic.Bool
	ticket, err := a.ReserveWindowCompletion(ref, func(e error) {
		cleaned.Add(1)
		if !errors.Is(e, app.ErrWindowClosed) {
			badCleanup.Store(true)
		}
	})
	if err != nil {
		return nil, err
	}
	native, err := readback(ctx, a, g, m, shown)
	if err != nil {
		return nil, err
	}
	var reopened *reopenedLifetimeProbe
	if c.sameName {
		if (i > 0) != (previous != nil) {
			return nil, errors.New("same-name fixture lost its preceding opening")
		}
		if previous != nil {
			reply, e := ownerProbe(ctx, a, func() ownerReply {
				if native.IsClosed() {
					return ownerReply{err: errors.New("replacement native window is closed")}
				}
				var e error
				reopened, e = probeReopenedLifetime(a, previous.ref, ref, previous.completion)
				if e == nil {
					e = previous.verify(a)
				}
				return ownerReply{err: e}
			})
			if e != nil || reply.err != nil {
				return nil, errors.Join(e, reply.err)
			}
		}
	}
	if suite != nil {
		if err = exerciseComposed(ctx, a, id, suite, i); err != nil {
			return nil, err
		}
	}
	if err = inventory(ctx, []string{"Vellum cycle survivor", title}); err != nil {
		return nil, err
	}
	if reopened != nil {
		// This owner acknowledgment follows the native action acknowledgments and
		// any ordinary work incorrectly enqueued by the old-reference probes.
		reply, e := ownerProbe(ctx, a, func() ownerReply {
			return ownerReply{err: errors.Join(reopened.afterOwnerBarrier(a, ref), previous.verify(a))}
		})
		if e != nil || reply.err != nil {
			return nil, errors.Join(e, reply.err)
		}
	}
	if err = a.CloseWindow(ref); err != nil {
		return nil, err
	}
	if !errors.Is(original.Err(), context.Canceled) || !errors.Is(child.Err(), context.Canceled) || cleaned.Load() != 1 || badCleanup.Load() {
		return nil, errors.New("captured context or reserved completion not canceled")
	}
	if err = ticket.Publish(func() {}); !errors.Is(err, app.ErrWindowClosed) {
		return nil, fmt.Errorf("late result admission: %v", err)
	}
	if err = a.PostWindow(ref, func() {}); !errors.Is(err, app.ErrWindowClosed) {
		return nil, fmt.Errorf("late owner delivery: %v", err)
	}
	if err = wait(ctx, func(inner context.Context) (bool, error) {
		return native.IsClosed() && core.AllRenderWindows.FindName(title) == nil, nil
	}); err != nil {
		return nil, err
	}
	if err = inventory(ctx, []string{"Vellum cycle survivor"}); err != nil {
		return nil, err
	}
	retired := &retiredNativeWindow{ref: ref, completion: ticket, native: native, canvas: g, composed: c.composed}
	if suite != nil {
		retired.grid, retired.outline = suite.Grid, suite.Tree
	}
	limit, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	reply, e := ownerProbe(limit, a, func() ownerReply { return ownerReply{err: retired.verify(a)} })
	if e != nil || reply.err != nil {
		return nil, errors.Join(e, reply.err)
	}
	if c.sameName {
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.window-lifetime", "cycle": i + 1, "window": id, "prior_opening_checked": previous != nil, "old_references_refused": previous != nil, "old_native_resources_retired": previous != nil, "old_callbacks_absent_after_owner_barrier": previous != nil, "current_opening_retired": true, "current_native_resources_retired": true, "full_p11_qualified": false}); err != nil {
			return nil, err
		}
		return retired, nil
	}
	return nil, nil
}

type ownerReply struct {
	native system.Window
	err    error
}

func ownerProbe(ctx context.Context, a *app.Application, fn func() ownerReply) (ownerReply, error) {
	reply := make(chan ownerReply, 1)
	var mu sync.Mutex
	phase := 0 // queued, running, revoked, completed
	if err := a.PostWait(ctx, func() {
		mu.Lock()
		if phase == 2 || ctx.Err() != nil {
			phase = 2
			mu.Unlock()
			reply <- ownerReply{err: ctx.Err()}
			return
		}
		phase = 1
		mu.Unlock()
		value := fn()
		mu.Lock()
		phase = 3
		mu.Unlock()
		reply <- value
	}); err != nil {
		return ownerReply{}, err
	}
	select {
	case value := <-reply:
		return value, nil
	case <-ctx.Done():
		mu.Lock()
		if phase == 0 {
			phase = 2
			mu.Unlock()
			return ownerReply{}, ctx.Err()
		}
		running := phase == 1 || phase == 3
		mu.Unlock()
		if running {
			value := <-reply
			return value, ctx.Err()
		}
		return ownerReply{}, ctx.Err()
	}
}

func memoryGuard(s observation) error {
	if s.RSSBytes > 1<<30 || s.HeapAlloc > 512<<20 {
		return fmt.Errorf("window-cycle safety memory limit at %d: RSS=%d GoHeap=%d", s.Completed, s.RSSBytes, s.HeapAlloc)
	}
	return nil
}
