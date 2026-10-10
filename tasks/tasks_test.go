package tasks

import (
	"context"
	"errors"
	"example.com/vellum/app"
	"testing"
	"time"
)

func TestBoundedCompletionAndCancellation(t *testing.T) {
	a, _ := app.New(app.Config{QueueCapacity: 1})
	defer a.Close()
	m, _ := New(a, 1, nil)
	a.Post(func() {})
	commits := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	task, err := Start(m, ctx, func(context.Context, func(float64, string)) (int, error) { return 42, nil }, func(v int, err error) {
		if v != 42 || err != nil {
			t.Errorf("%d %v", v, err)
		}
		commits++
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Start(m, ctx, func(context.Context, func(float64, string)) (int, error) { return 0, nil }, func(int, error) {}); !errors.Is(err, ErrCapacity) {
		t.Fatal(err)
	}
	a.Drain()
	select {
	case <-task.Done():
	case <-ctx.Done():
		t.Fatal("completion stuck")
	}
	a.Drain()
	if commits != 1 {
		t.Fatal(commits)
	}
	task, err = Start(m, ctx, func(ctx context.Context, _ func(float64, string)) (int, error) { <-ctx.Done(); return 0, ctx.Err() }, func(_ int, err error) {
		if !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	task.Cancel()
	<-task.Done()
	a.Drain()
}
func TestOwnerShutdownUnblocksWorker(t *testing.T) {
	a, _ := app.New(app.Config{QueueCapacity: 1})
	m, _ := New(a, 1, nil)
	a.Post(func() {})
	task, err := Start(m, context.Background(), func(context.Context, func(float64, string)) (int, error) { return 0, nil }, func(int, error) { t.Error("commit after close") })
	if err != nil {
		t.Fatal(err)
	}
	a.Close()
	select {
	case <-task.Done():
	case <-time.After(time.Second):
		t.Fatal("worker leaked")
	}
}

func TestCancellationAfterAdmissionIsRecheckedAtCommit(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	m, err := New(a, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	commits := 0
	task, err := Start(m, ctx, func(context.Context, func(float64, string)) (int, error) { return 17, nil }, func(_ int, err error) {
		commits++
		if !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	<-task.Done()
	cancel()
	a.Drain()
	if commits != 1 {
		t.Fatal(commits)
	}
	if _, err := Start(m, a.Context(), func(context.Context, func(float64, string)) (int, error) { return 0, nil }, func(int, error) {}); err != nil {
		t.Fatal("slot not released", err)
	}
}
