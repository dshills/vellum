package platformio

import (
	"context"
	"errors"
	"testing"

	"example.com/vellum/app"
)

func TestDispatchValue(t *testing.T) {
	stopped := make(chan struct{})
	value, err := Dispatch(context.Background(), stopped, func(_ context.Context, run func()) error { run(); return nil }, func() (string, error) { return "unicode 雪", nil })
	if err != nil || value != "unicode 雪" {
		t.Fatalf("%q %v", value, err)
	}
}
func TestDispatchCancellationBeforeCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	var run func()
	called := false
	_, err := Dispatch(ctx, stopped, func(_ context.Context, f func()) error { run = f; cancel(); return nil }, func() (int, error) { called = true; return 42, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	run()
	if called {
		t.Fatal("canceled native action executed")
	}
}
func TestDispatchCancellationDuringCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	started, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	go func() {
		_, err := Dispatch(ctx, stopped, func(_ context.Context, run func()) error { go func() { run(); close(finished) }(); return nil }, func() (int, error) { close(started); <-release; return 42, nil })
		result <- err
	}()
	<-started
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	close(release)
	<-finished // private result cannot block after caller returned
}
func TestDispatchStoppedQueue(t *testing.T) {
	stopped := make(chan struct{})
	close(stopped)
	_, err := Dispatch(context.Background(), stopped, func(_ context.Context, run func()) error { run(); return nil }, func() (int, error) { t.Fatal("stopped native action executed"); return 0, nil })
	if !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
}

func TestDispatchStopBeforeQueuedCallback(t *testing.T) {
	stopped := make(chan struct{})
	var run func()
	called := false
	_, err := Dispatch(context.Background(), stopped, func(_ context.Context, f func()) error { run = f; close(stopped); return nil }, func() (int, error) { called = true; return 0, nil })
	if !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
	run()
	if called {
		t.Fatal("action after native shutdown")
	}
}
