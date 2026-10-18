package main

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type catalogueFailDriver struct {
	admitted <-chan struct{}
	failure  error
}

func (d catalogueFailDriver) Run(ctx context.Context, a *app.Application) error {
	select {
	case <-d.admitted:
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		return errors.New("owner callback was not admitted")
	}
	if a.Pending().Ordinary != 1 {
		return errors.New("owner callback not queued")
	}
	return d.failure
}
func TestCatalogueRetiredOwnerWorkerJoins(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	admitted := make(chan struct{}, 1)
	a.SetWake(func() {
		select {
		case admitted <- struct{}{}:
		default:
		}
	})
	if err = a.Open(app.Window{ID: "gallery", Title: "Media retirement regression", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: "not launched"} }}); err != nil {
		t.Fatal(err)
	}
	<-admitted // Discard the registration wake, before the worker posts its callback.
	ref, err := a.WindowRef("gallery")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan catalogueOutcome[struct{}], 1)
	var called atomic.Bool
	go func() {
		e := catalogueOwner(ctx, a, ref, func() error { called.Store(true); return nil })
		done <- catalogueOutcome[struct{}]{err: e}
	}()
	failure := errors.New("injected driver exit before owner dispatch")
	if err = a.Run(context.Background(), catalogueFailDriver{admitted, failure}); !errors.Is(err, failure) {
		t.Fatal("actual Run failure lost", err)
	}
	if ref.Context().Err() == nil || a.Pending().Ordinary != 0 || called.Load() {
		t.Fatal("actual application did not retire and drop undispatched callback")
	}
	select {
	case <-done:
		t.Fatal("worker unexpectedly finished before scenario cancellation")
	default:
	}
	result, joined := catalogueWorkerResult(cancel, done)
	if !joined {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("regression cleanup worker did not join")
		}
		t.Fatal("retired owner wait outlived Run")
	}
	if !errors.Is(result.err, context.Canceled) || ctx.Err() != context.Canceled || called.Load() {
		t.Fatal("worker failed cancellation identity or dispatched retired callback", result.err)
	}
}

type catalogueClosingDriver struct {
	admitted <-chan struct{}
	drained  chan<- error
}

func (d catalogueClosingDriver) Run(ctx context.Context, a *app.Application) error {
	select {
	case <-d.admitted:
	case <-time.After(3 * time.Second):
		return errors.New("close was not admitted")
	}
	go func() { d.drained <- a.Drain() }()
	<-ctx.Done()
	return ctx.Err()
}
func TestCatalogueLastWindowAcknowledgesAfterRunCancellation(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	admitted := make(chan struct{}, 1)
	ack := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(ack) }) }
	defer release()
	a.SetWake(func() {
		if a.Context().Err() != nil {
			<-ack
			return
		}
		select {
		case admitted <- struct{}{}:
		default:
		}
	})
	if err = a.Open(app.Window{ID: "gallery", Title: "Last window retirement", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	<-admitted
	ref, err := a.WindowRef("gallery")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan catalogueOutcome[struct{}], 1)
	workerDone := make(chan struct{})
	go func() {
		closed, e := catalogueFinishWindow(ctx, a, cogent.NewDesktopSession(), ref, cogent.DesktopInput{Frame: 1, Target: "retired", Type: events.KeyChord, Code: key.CodeEnd})
		if e == nil && !closed {
			e = errors.New("close did not retire ref")
		}
		done <- catalogueOutcome[struct{}]{err: e}
		close(workerDone)
	}()
	drained := make(chan error, 1)
	if e := a.Run(context.Background(), catalogueClosingDriver{admitted, drained}); !errors.Is(e, context.Canceled) {
		t.Fatal("last window did not end actual Run", e)
	}
	if ref.Context().Err() == nil {
		t.Fatal("last-window lifetime remained live")
	}
	joined := make(chan catalogueOutcome[struct{}], 1)
	go func() {
		r, ok := catalogueWorkerResult(cancel, done)
		if !ok {
			r.err = errors.Join(r.err, errors.New("terminal worker did not join"))
		}
		joined <- r
	}()
	<-ctx.Done() // Run-return cancellation precedes the owner's close reply.
	select {
	case <-workerDone:
		release()
		t.Fatal("terminal probe canceled before its close acknowledgement")
	case <-time.After(50 * time.Millisecond):
	}
	release()
	select {
	case r := <-joined:
		if r.err != nil {
			t.Fatal("closed-reference probe lost terminal authority", r.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("bounded final probe did not join")
	}
	select {
	case e := <-drained:
		if e != nil && !errors.Is(e, app.ErrClosed) {
			t.Fatal("owner drain", e)
		}
	case <-time.After(time.Second):
		t.Fatal("owner did not finish after acknowledgement release")
	}
}

func TestCatalogueFinalCloseRefusesCanceledScenario(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "gallery", Title: "Canceled scenario", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	ref, err := a.WindowRef("gallery")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	closed, err := catalogueFinishWindow(ctx, a, cogent.NewDesktopSession(), ref, cogent.DesktopInput{})
	if closed || !errors.Is(err, context.Canceled) || ref.Context().Err() != nil || a.Pending().Ordinary != 0 {
		t.Fatal("canceled scenario acquired terminal authority", closed, err, a.Pending())
	}
}

func TestCatalogueFinalCloseDroppedWithRetirementIsBounded(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	admitted := make(chan struct{}, 1)
	a.SetWake(func() {
		select {
		case admitted <- struct{}{}:
		default:
		}
	})
	if err = a.Open(app.Window{ID: "gallery", Title: "Dropped final close", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	<-admitted
	ref, err := a.WindowRef("gallery")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan catalogueOutcome[struct{}], 1)
	go func() {
		_, e := catalogueFinishWindow(ctx, a, cogent.NewDesktopSession(), ref, cogent.DesktopInput{})
		done <- catalogueOutcome[struct{}]{err: e}
	}()
	failure := errors.New("retire before final close dispatch")
	if err = a.Run(context.Background(), catalogueFailDriver{admitted, failure}); !errors.Is(err, failure) {
		t.Fatal("actual driver failure lost", err)
	}
	if ref.Context().Err() == nil || a.Pending().Ordinary != 0 {
		t.Fatal("actual retirement did not drop terminal callback")
	}
	result, joined := catalogueWorkerResult(cancel, done)
	if !joined || !errors.Is(result.err, context.DeadlineExceeded) {
		t.Fatal("detached terminal close failed its independent deadline", joined, result.err)
	}
}
