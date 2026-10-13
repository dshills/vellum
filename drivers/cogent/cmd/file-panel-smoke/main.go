//go:build darwin && cgo && !offscreen && vellum_native_checks

// file-panel-smoke observes real Cocoa panels before canceling callers and
// shutting down with one panel still active. It never selects or writes a file.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type config struct {
	cycles   int
	shutdown string
	timeout  time.Duration
}

func (c config) validate() error {
	if c.cycles < 1 || c.cycles > 100 || c.shutdown != "open" && c.shutdown != "save" || c.timeout < 10*time.Second || c.timeout > 5*time.Minute {
		return errors.New("file-panel fixture requires cycles 1..100, shutdown-mode open/save, timeout 10s..5m")
	}
	return nil
}

type reply struct {
	path string
	err  error
}
type event struct {
	Stage string
	Cycle int
	Mode  app.FileMode
	State cogent.NativeServiceCheckState
}
type recorder struct {
	sync.Mutex
	events []event
}

func (r *recorder) add(e event) { r.Lock(); defer r.Unlock(); r.events = append(r.events, e) }
func (r *recorder) snapshot() []event {
	r.Lock()
	defer r.Unlock()
	return append([]event(nil), r.events...)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	var c config
	flag.IntVar(&c.cycles, "cycles", 10, "sequential caller-cancel cycles, alternating Open and Save, 1..100")
	flag.StringVar(&c.shutdown, "shutdown-mode", "open", "active panel during application shutdown: open or save")
	flag.DurationVar(&c.timeout, "timeout", 90*time.Second, "native loop deadline, 10s..5m")
	flag.Parse()
	if err := c.validate(); err != nil {
		return err
	}
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
	if err = a.Open(app.Window{ID: "panels", Title: "Vellum file panel lifecycle", Width: 420, Height: 180, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "help", Kind: ui.Label, Text: "Temporary Open/Save cancellation and shutdown check"}}}
	}}); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	record := &recorder{}
	finished := make(chan error, 1)
	failures := make(chan error, 1)
	go func() { finished <- check(ctx, a, c, record); a.Close() }()
	runErr := a.Run(ctx, cogent.Driver{OnError: func(err error) {
		select {
		case failures <- err:
		default:
		}
	}})
	// Shutdown is an intentional stage: its worker must join the service caller
	// after the native loop has retired. Preserve partial events on every exit.
	cancel()
	select {
	case err = <-finished:
	case <-time.After(6 * time.Second):
		err = errors.New("file-panel worker did not join after native shutdown")
	}
	select {
	case e := <-failures:
		err = errors.Join(err, e)
	default:
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.FilePanels != (cogent.NativeServiceFilePanelCheck{}) {
		err = errors.Join(err, fmt.Errorf("file-panel terminal inventory: observed=%t state=%+v", observed, cleanup))
	}
	result := "FAILED_NATIVE_FILE_PANEL_LIFECYCLE"
	if errors.Join(runErr, err) == nil {
		result = "PASS_NATIVE_FILE_PANEL_LIFECYCLE"
	}
	report := map[string]any{"result": result, "requested_cycles": c.cycles, "shutdown_mode": c.shutdown, "events": record.snapshot(), "native_cleanup": cleanup, "selection_qualified": false, "physical_cancel_qualified": false, "clipboard_access": false, "preferences_changed": false}
	encodeErr := json.NewEncoder(os.Stdout).Encode(report)
	return errors.Join(runErr, err, encodeErr)
}

func panelMatches(s cogent.NativeServiceCheckState, mode app.FileMode, active bool) bool {
	p := s.FilePanels
	if !s.MainThread {
		return false
	}
	if !active {
		return p == (cogent.NativeServiceFilePanelCheck{}) && s.NativeOutstanding == 0
	}
	return p.Tracked == 1 && p.Visible == 1 && p.Open+p.Save == 1 && ((mode == app.OpenFile && p.Open == 1 && p.Save == 0) || (mode == app.SaveFile && p.Save == 1 && p.Open == 0)) && s.NativeOutstanding == 1
}
func waitPanel(ctx context.Context, mode app.FileMode, active bool) (cogent.NativeServiceCheckState, error) {
	limit, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	var state cogent.NativeServiceCheckState
	for {
		var err error
		state, err = cogent.CheckNativeServices(limit, cogent.NativeCheckSnapshot, 0)
		if err != nil {
			return state, err
		}
		if panelMatches(state, mode, active) {
			return state, nil
		}
		select {
		case <-limit.Done():
			return state, fmt.Errorf("panel active=%t mode=%d never settled: %+v: %w", active, mode, state, limit.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
}
func receive(ctx context.Context, replies <-chan reply) (reply, error) {
	select {
	case value := <-replies:
		return value, nil
	case <-ctx.Done():
		return reply{}, ctx.Err()
	}
}

// In-flight calls retain their Context contract when application lifetime ends.
// New admission after that boundary has the separate app.ErrClosed contract.
func canceledReply(value reply) bool {
	return value.path == "" && errors.Is(value.err, context.Canceled)
}
func choose(ctx context.Context, service app.Services, mode app.FileMode, title string) <-chan reply {
	replies := make(chan reply, 1)
	go func() {
		path, err := service.ChooseFile(ctx, app.FileOptions{Mode: mode, Title: title, Filename: "vellum-unused-destination.txt", Extensions: []string{"txt"}})
		replies <- reply{path, err}
	}()
	return replies
}
func check(ctx context.Context, a *app.Application, c config, r *recorder) error {
	for {
		s, e := cogent.InspectAccessibility(ctx, "panels")
		if e == nil && s.Reachable && len(s.Elements) > 0 && s.Elements[0].Frame[2] > 0 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	service, err := a.Services()
	if err != nil {
		return err
	}
	state, err := waitPanel(ctx, app.OpenFile, false)
	if err != nil {
		return err
	}
	r.add(event{Stage: "initial_empty", State: state})
	for i := 0; i < c.cycles; i++ {
		mode := app.FileMode(i % 2)
		child, stop := context.WithCancel(ctx)
		replies := choose(child, service, mode, fmt.Sprintf("Vellum caller cancellation %d", i))
		state, err = waitPanel(ctx, mode, true)
		if err != nil {
			stop()
			return err
		}
		r.add(event{Stage: "visible_before_cancel", Cycle: i, Mode: mode, State: state})
		// A completed selection/cancel before our caller cancellation cannot be
		// misreported as the intended context-cancel lifecycle observation.
		select {
		case value := <-replies:
			stop()
			return fmt.Errorf("panel completed before caller cancellation: path-present=%t error=%v", value.path != "", value.err)
		default:
		}
		stop()
		value, e := receive(ctx, replies)
		if e != nil || !canceledReply(value) {
			return errors.Join(e, fmt.Errorf("caller cancellation: path-present=%t error=%v", value.path != "", value.err))
		}
		state, err = waitPanel(ctx, mode, false)
		if err != nil {
			return err
		}
		r.add(event{Stage: "retired_after_cancel", Cycle: i, Mode: mode, State: state})
	}
	mode := app.OpenFile
	if c.shutdown == "save" {
		mode = app.SaveFile
	}
	// Background belongs to the caller; application ownership must cancel this
	// operation even without an explicit caller cancellation or deadline.
	replies := choose(context.Background(), service, mode, "Vellum active panel shutdown")
	state, err = waitPanel(ctx, mode, true)
	if err != nil {
		return err
	}
	r.add(event{Stage: "visible_before_shutdown", Mode: mode, State: state})
	select {
	case value := <-replies:
		return fmt.Errorf("shutdown panel completed early: path-present=%t error=%v", value.path != "", value.err)
	default:
	}
	a.Close()
	join, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	value, e := receive(join, replies)
	if e != nil || !canceledReply(value) {
		return errors.Join(e, fmt.Errorf("owner shutdown: path-present=%t error=%v", value.path != "", value.err))
	}
	if path, e := service.ChooseFile(join, app.FileOptions{}); path != "" || !errors.Is(e, app.ErrClosed) {
		return fmt.Errorf("service reuse after shutdown: path-present=%t error=%v", path != "", e)
	}
	r.add(event{Stage: "owner_caller_joined", Mode: mode})
	return nil
}
