//go:build darwin && cgo && !offscreen && vellum_native_checks

// This tagged driver-module diagnostic overlays cmd/window-lifetime-smoke.
// It uses public application/render APIs plus existing internal fixture and tagged
// cleanup diagnostics. It never suppresses native callbacks.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/system/driver/desktop"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/nativefixture"
	"github.com/dshills/vellum/ui"
)

const ordinaryClose = "UNTAGGED_ORDINARY_CLOSE"
const cpuDeadline = "UNTAGGED_CPU_BATCH_DEADLINE"
const preShowDeadline = "UNTAGGED_PRE_SHOW_CPU_BATCH_DEADLINE"

type lifecycleCanvas struct {
	canvas        *cogent.MeshCanvas
	widget        core.Widget
	shown, closed atomic.Bool
}

func (x *lifecycleCanvas) Mount(parent tree.Node) core.Widget {
	w := x.canvas.Mount(parent)
	x.widget = w
	if w != nil && w.AsWidget().Scene != nil && w.AsWidget().Scene.Body != nil {
		w.AsWidget().Scene.Body.OnShow(func(events.Event) { x.shown.Store(true) })
		w.AsWidget().Scene.Body.OnClose(func(events.Event) { x.closed.Store(true) })
	}
	return w
}

func (x *lifecycleCanvas) Sync(w core.Widget) error { return x.canvas.Sync(w) }

type preShowObservation struct {
	ShownAtEntry       bool
	Before             desktop.DisplayBackpressureStatus
	After              desktop.DisplayBackpressureStatus
	EntryRecordWritten bool
	Completed          bool
}

type report struct {
	Case                          string
	Result                        string
	Error                         string
	RunError                      string
	RunReturned                   bool
	WorkerJoined                  bool
	Shown, Closed                 bool
	PrepareCalls                  int64
	SlowCalls                     int64
	SlowWorkNS                    int64
	TypedRefusal                  bool
	FrameMetricsEnabled           bool
	PresentationMetricsEnabled    bool
	CleanupObserved               bool
	Cleanup                       cogent.NativeServiceCheckState
	LostPresentedCallbackVerified bool
	PreShow                       *preShowObservation `json:",omitempty"`
}

func waitReady(ctx context.Context, a *app.Application, x *lifecycleCanvas, visits *atomic.Int64) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if x.shown.Load() && visits.Load() > 0 {
			_, err := a.Services()
			if err == nil {
				return nil
			}
			if !errors.Is(err, app.ErrUnavailable) {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.Context().Done():
			return errors.New("application closed before native readiness")
		case <-ticker.C:
		}
	}
}

func run(which string) (out report, err error) {
	out.Case = which
	out.Result = "FAILED_NATIVE_UNTAGGED_LIFECYCLE"
	if which != ordinaryClose && which != cpuDeadline && which != preShowDeadline {
		return out, errors.New("case must be UNTAGGED_ORDINARY_CLOSE, UNTAGGED_CPU_BATCH_DEADLINE or UNTAGGED_PRE_SHOW_CPU_BATCH_DEADLINE")
	}
	// Load lazy settings before applying the process-local fixture policy.
	// This defer precedes a.Close, so restoration follows native/app teardown.
	restoreGeometry, err := nativefixture.IsolateGeometry()
	if err != nil {
		return out, err
	}
	defer restoreGeometry()
	watchdog := time.AfterFunc(10*time.Second, func() {
		fmt.Fprintln(os.Stderr, "FAIL native untagged lifecycle hard watchdog; cleanup unqualified")
		os.Exit(1)
	})
	defer watchdog.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	a, err := app.New(app.Config{})
	if err != nil {
		return out, err
	}
	defer a.Close()
	remove, err := cogent.ObserveNativeServiceCleanup()
	if err != nil {
		return out, err
	}
	defer remove()
	mesh, err := canvas.NewMesh(nil, nil)
	if err != nil {
		return out, err
	}
	view := canvas.Viewport{Width: 240, Height: 200, Zoom: 1}
	var visits, slowCalls, slowNS atomic.Int64
	var slowOnce sync.Once
	var x *lifecycleCanvas
	var preMu sync.Mutex
	var pre preShowObservation
	var entryErr atomic.Pointer[error]
	armed := false // accessed only on the application owner, including Prepare
	g := &cogent.MeshCanvas{Owner: a, Window: "probe", View: &view}
	g.Prepare = func(canvas.Viewport) (*canvas.Mesh, error) {
		visits.Add(1)
		if which == preShowDeadline {
			slowOnce.Do(func() {
				rw := x.widget.AsWidget().Scene.RenderWindow()
				if rw == nil || rw.SystemWindow == nil {
					return
				}
				observed, ok := rw.SystemWindow.(interface {
					DisplayBackpressureSnapshot() desktop.DisplayBackpressureStatus
				})
				if !ok {
					return
				}
				before := observed.DisplayBackpressureSnapshot()
				atEntry := x.shown.Load()
				preMu.Lock()
				pre.ShownAtEntry, pre.Before = atEntry, before
				preMu.Unlock()
				// This detached entry survives a failed old Driver.Run that never returns.
				e := json.NewEncoder(os.Stdout).Encode(struct {
					Event        string
					ShownAtEntry bool
					Credit       desktop.DisplayBackpressureStatus
				}{"vellum.untagged-pre-show.entry", atEntry, before})
				if e != nil {
					entryErr.Store(&e)
					return
				}
				preMu.Lock()
				pre.EntryRecordWritten = true
				preMu.Unlock()
				if atEntry || !before.Enabled || before.Active == 0 || before.NativePending != 0 || before.CPUFinished || before.Closed || before.Failure != "" {
					return
				}
				slowCalls.Add(1)
				started := time.Now()
				time.Sleep(250 * time.Millisecond)
				slowNS.Store(time.Since(started).Nanoseconds())
				after := observed.DisplayBackpressureSnapshot()
				preMu.Lock()
				pre.After, pre.Completed = after, true
				preMu.Unlock()
			})
		} else if armed {
			slowOnce.Do(func() {
				slowCalls.Add(1)
				started := time.Now()
				// Deliberate bounded application work after Core admission. This
				// exercises its CPU-batch deadline, not missing native callbacks.
				time.Sleep(250 * time.Millisecond)
				slowNS.Store(time.Since(started).Nanoseconds())
			})
		}
		return mesh, nil
	}
	x = &lifecycleCanvas{canvas: g}
	ref, err := a.OpenWindow(app.Window{ID: "probe", Title: "Vellum untagged lifecycle", Width: 360, Height: 260, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Extension, Grow: true, Custom: x}
	}})
	if err != nil {
		return out, err
	}
	worker := make(chan error, 1)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		if which == preShowDeadline {
			select {
			case <-a.Context().Done():
				worker <- nil
			case <-ctx.Done():
				worker <- ctx.Err()
				a.Close()
			}
			return
		}
		if e := waitReady(ctx, a, x, &visits); e != nil {
			worker <- e
			a.Close()
			return
		}
		if which == ordinaryClose {
			worker <- nil
			a.Close()
			return
		}
		// No Snapshot/Rebuild call is used: those could render outside the
		// ordinary admitted window batch. Sync only marks the widget dirty.
		e := a.PostWindow(ref, func() { armed = true; g.Revision++; a.Invalidate("probe") })
		worker <- e
		if e != nil {
			a.Close()
		}
	}()
	timeout := time.Second
	if which == cpuDeadline || which == preShowDeadline {
		timeout = 100 * time.Millisecond
	}
	driver := cogent.Driver{ExperimentalDisplayBackpressure: timeout}
	out.FrameMetricsEnabled = driver.FrameMetrics != nil
	out.PresentationMetricsEnabled = driver.PresentationMetrics != nil
	// Both metrics fields intentionally remain nil. No tagged FrameTrace or
	// presentation recorder is required for the real monitor's error path.
	runErr := a.Run(ctx, driver)
	out.RunReturned = true
	if runErr != nil {
		out.RunError = runErr.Error()
	}
	var refusal *cogent.DisplayBackpressureError
	out.TypedRefusal = errors.As(runErr, &refusal)
	select {
	case <-workerDone:
		out.WorkerJoined = true
		err = errors.Join(err, <-worker)
	case <-ctx.Done():
		err = errors.Join(err, errors.New("worker did not join before cooperative deadline"))
	}
	out.Shown, out.Closed = x.shown.Load(), x.closed.Load()
	out.PrepareCalls, out.SlowCalls, out.SlowWorkNS = visits.Load(), slowCalls.Load(), slowNS.Load()
	if which == preShowDeadline {
		preMu.Lock()
		snapshot := pre
		preMu.Unlock()
		out.PreShow = &snapshot
		if e := entryErr.Load(); e != nil {
			err = errors.Join(err, *e)
		}
	}

	out.Cleanup, out.CleanupObserved = cogent.NativeServiceCleanupObservation()
	c := out.Cleanup
	if ctx.Err() != nil {
		err = errors.Join(err, fmt.Errorf("cooperative deadline: %w", ctx.Err()))
	}
	if !out.WorkerJoined || (!out.Shown && which != preShowDeadline) || !out.Closed || out.PrepareCalls == 0 {
		err = errors.Join(err, errors.New("native lifetime/readiness observations incomplete"))
	}
	if !out.CleanupObserved || !c.MainThread || !c.CleanupObserved || !c.CallbacksStopped || c.PendingCallbacks != 0 || c.NativeOutstanding != 0 || c.AccessibilityWindows != 0 || len(c.WindowTitles) != 0 {
		err = errors.Join(err, errors.New("native terminal cleanup inventory incomplete"))
	}
	if which == ordinaryClose {
		if runErr != nil || out.SlowCalls != 0 {
			err = errors.Join(err, fmt.Errorf("ordinary close failed: %v", runErr))
		}
	} else {
		if !out.TypedRefusal || refusal.Reason != "presentation deadline exceeded" || (which == cpuDeadline && out.PrepareCalls < 2) || out.SlowCalls != 1 || out.SlowWorkNS < (250*time.Millisecond).Nanoseconds() {
			err = errors.Join(err, errors.New("actual untagged CPU-batch deadline was not observed"))
		}
	}
	if which == preShowDeadline {
		p := out.PreShow
		if p == nil || p.ShownAtEntry || !p.EntryRecordWritten || !p.Completed || !p.Before.Enabled || p.Before.Active == 0 || p.Before.NativePending != 0 || p.Before.CPUFinished || p.Before.Closed || p.Before.Failure != "" || p.After.Active != p.Before.Active || p.After.NativePending != 0 || p.After.CPUFinished || p.After.Failure != "presentation deadline exceeded" {
			err = errors.Join(err, errors.New("pre-show admitted CPU-batch deadline evidence incomplete"))
		}
	}
	if err == nil {
		out.Result = "PASS_NATIVE_UNTAGGED_LIFECYCLE"
	}
	return out, err
}

func main() {
	which := flag.String("case", "", "one explicit untagged lifecycle case per process")
	flag.Parse()
	out, err := run(*which)
	if err != nil {
		out.Error = err.Error()
	}
	if encodeErr := json.NewEncoder(os.Stdout).Encode(out); encodeErr != nil {
		fmt.Fprintln(os.Stderr, encodeErr)
		os.Exit(1)
	}
	if err != nil {
		os.Exit(1)
	}
}
