//go:build darwin && cgo && !offscreen && vellum_native_checks

// workspace-native consumes the full immutable workspace record through native
// Go event routing. It does not generate OS input or claim deterministic replay.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"time"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/replay/internal/smoke"
	"github.com/dshills/vellum/examples/replay/workspace"
	"github.com/dshills/vellum/execution"
)

type report struct {
	Event, Result, Error         string
	Workflow                     workspace.NativeReport
	Session                      cogent.DesktopState
	Executor                     execution.Snapshot
	Cleanup                      cogent.NativeServiceCheckState
	WorkerJoined, ScenarioJoined bool
	Seconds                      float64
}

func run(rows int, timeout time.Duration) (out report, err error) {
	out.Event = "vellum.workspace-native.result"
	out.Result = "FAILED_NATIVE_WORKSPACE"
	started := time.Now()
	defer func() {
		out.Seconds = time.Since(started).Seconds()
		if err != nil {
			out.Error = err.Error()
		}
	}()
	if (rows != 10000 && rows != workspace.MaxRows) || timeout < 10*time.Second || timeout > 5*time.Minute {
		return out, errors.New("rows 10000 or 1000000 and timeout 10s..5m required")
	}
	previous := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	defer func() { core.DebugSettings.DisableWindowGeometrySaver = previous }()
	a, e := app.New(app.Config{QueueCapacity: 128, CompletionCapacity: 8})
	if e != nil {
		return out, e
	}
	defer a.Close()
	ex, e := execution.NewManual(8)
	if e != nil {
		return out, e
	}
	defer ex.Close()
	f, e := workspace.New(a, ex, rows)
	if e != nil {
		return out, e
	}
	defer f.Palette.Close()
	remove, e := cogent.ObserveNativeServiceCleanup()
	if e != nil {
		return out, e
	}
	defer remove()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	session := cogent.NewDesktopSession()
	var workflow workspace.NativeReport
	finished, done := smoke.StartScenario(func() error { var e error; workflow, e = workspace.NativeWorkflow(ctx, f, session, ex, rows); return e }, a.Close)
	nativeErrors := make(chan error, 1)
	runErr := a.Run(ctx, cogent.Driver{Commands: f.Commands, Theme: workspace.FixtureTheme(), DesktopSession: session, OnError: func(e error) {
		select {
		case nativeErrors <- e:
		default:
		}
	}})
	cancel()
	a.Close()
	var scenarioErr error
	select {
	case <-done:
		out.Workflow = workflow
		out.ScenarioJoined = true
		scenarioErr = <-finished
	case <-time.After(6 * time.Second):
		out.Error = "fatal: native scenario failed to join; process exit avoids concurrent deferred cleanup"
		_ = json.NewEncoder(os.Stdout).Encode(out)
		os.Exit(3)
	}
	select {
	case e := <-nativeErrors:
		scenarioErr = errors.Join(scenarioErr, e)
	default:
	}
	joinCtx, joinCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer joinCancel()
	var joinErr error
	for _, id := range ex.PendingIDs() {
		joinErr = errors.Join(joinErr, ex.Step(joinCtx, id))
	}
	joinErr = errors.Join(joinErr, f.WaitWorker(joinCtx))
	out.WorkerJoined = joinErr == nil
	ex.Close()
	out.Executor = ex.Snapshot()
	out.Session = session.State()
	if !out.Executor.Idle() || !out.Session.Closed || out.Session.Pending || out.Session.LiveWindows != 0 {
		joinErr = errors.Join(joinErr, errors.New("session or worker references remain"))
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	out.Cleanup = cleanup
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 || cleanup.FilePanels.Tracked != 0 || cleanup.Menu.Owned || cleanup.Tray.Present {
		joinErr = errors.Join(joinErr, errors.New("native terminal inventory not retired"))
	}
	err = errors.Join(runErr, scenarioErr, joinErr)
	if err == nil {
		out.Result = "PASS_NATIVE_WORKSPACE_RECORD"
	}
	return out, err
}
func main() {
	rows := flag.Int("rows", 10000, "stored rows: 10000 or 1000000")
	timeout := flag.Duration("timeout", 3*time.Minute, "native cooperative deadline 10s..5m")
	flag.Parse()
	out, err := run(*rows, *timeout)
	if e := json.NewEncoder(os.Stdout).Encode(out); e != nil {
		os.Exit(2)
	}
	if err != nil {
		os.Exit(1)
	}
}
