package tasks

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

func candidate(t *testing.T, window string) (*app.Application, *Manager, *execution.Manual) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 1, CompletionCapacity: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	for _, id := range []string{"one", "two"} {
		if err := a.Open(app.Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
			t.Fatal(err)
		}
	}
	e, err := execution.NewManual(4)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewWithConfig(Config{Owner: a, Window: window, Limit: 1, Executor: e})
	if err != nil {
		t.Fatal(err)
	}
	return a, m, e
}
func bounded(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func closed(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	default:
		t.Fatal("expected synchronously observable terminal signal")
	}
}

func TestManualWorkerCompletesAdmissionWithFullOrdinaryQueue(t *testing.T) {
	a, m, e := candidate(t, "")
	ctx := bounded(t)
	var order []string
	if err := a.Post(func() { order = append(order, "input") }); err != nil {
		t.Fatal(err)
	}
	task, err := Start(m, ctx, func(context.Context, func(float64, string)) (int, error) { return 42, nil }, func(v int, err error) {
		if v != 42 || err != nil {
			t.Errorf("result %d %v", v, err)
		}
		order = append(order, "result")
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.StepNext(ctx); err != nil {
		t.Fatalf("worker blocked on mailbox: %v", err)
	}
	closed(t, task.Done())
	if s := task.Snapshot(); s.Running || !s.WorkerDone || s.OwnerDone || !s.AwaitingOwner {
		t.Fatalf("worker/delivery states %+v", s)
	}
	if m.Pending().Idle() || a.Pending().ReadyCompletions != 1 {
		t.Fatal("pending result reported idle")
	}
	if n, err := a.DrainN(1); err != nil || n != 1 || len(order) != 1 || order[0] != "input" {
		t.Fatalf("input n%d err%v %v", n, err, order)
	}
	if n, err := a.DrainN(1); err != nil || n != 1 {
		t.Fatalf("result n%d %v", n, err)
	}
	closed(t, task.OwnerDone())
	if !m.Pending().Idle() || !a.Pending().Idle() || len(order) != 2 || order[1] != "result" {
		t.Fatalf("terminal cleanup lag %+v %+v %v", m.Pending(), a.Pending(), order)
	}
}

type refusing struct {
	err   error
	calls atomic.Int32
}

func (r *refusing) Submit(func()) error { r.calls.Add(1); return r.err }
func TestRefusedExecutorAndReservationRefundEveryCreditWithoutWork(t *testing.T) {
	a, m, e := candidate(t, "")
	refusal := errors.New("executor refusal")
	r := &refusing{err: refusal}
	m.executor = r
	var work atomic.Int32
	if task, err := Start(m, bounded(t), func(context.Context, func(float64, string)) (int, error) { work.Add(1); return 0, nil }, func(int, error) { t.Error("refused commit ran") }); task != nil || !errors.Is(err, refusal) {
		t.Fatalf("refusal %v %v", task, err)
	}
	if work.Load() != 0 || !m.Pending().Idle() || !a.Pending().Idle() {
		t.Fatalf("refund lag work%d manager%+v app%+v", work.Load(), m.Pending(), a.Pending())
	}
	m.executor = e
	tickets := make([]*app.Completion, 4)
	for i := range tickets {
		var err error
		tickets[i], err = a.ReserveCompletion("", nil)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Start(m, bounded(t), func(context.Context, func(float64, string)) (int, error) { work.Add(1); return 0, nil }, func(int, error) {}); !errors.Is(err, app.ErrCompletionCapacity) {
		t.Fatalf("reservation refusal %v", err)
	}
	if !m.Pending().Idle() || len(e.PendingIDs()) != 0 {
		t.Fatal("failed reservation started worker or retained slot")
	}
	for _, ticket := range tickets {
		ticket.Abort()
	}
}

func TestCanceledAcceptedWorkerStillRunsCleanupAndOwnerTerminalResult(t *testing.T) {
	a, m, e := candidate(t, "")
	ctx := bounded(t)
	var clean atomic.Int32
	var commitErr error
	task, err := Start(m, ctx, func(ctx context.Context, _ func(float64, string)) (int, error) {
		defer clean.Add(1)
		return 0, ctx.Err()
	}, func(_ int, err error) { commitErr = err })
	if err != nil {
		t.Fatal(err)
	}
	task.Cancel()
	if _, err := e.StepNext(ctx); err != nil {
		t.Fatal(err)
	}
	if clean.Load() != 1 || !errors.Is(task.Snapshot().Err, context.Canceled) {
		t.Fatalf("canceled cleanup %d state%+v", clean.Load(), task.Snapshot())
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	closed(t, task.OwnerDone())
	if !errors.Is(commitErr, context.Canceled) || !m.Pending().Idle() {
		t.Fatalf("terminal cancel err%v %+v", commitErr, m.Pending())
	}
}

func TestPanicAndGoexitPublishFailureAndRetireBothLatches(t *testing.T) {
	for _, mode := range []string{"panic", "goexit"} {
		t.Run(mode, func(t *testing.T) {
			a, m, e := candidate(t, "")
			ctx := bounded(t)
			var cleanup atomic.Int32
			var received error
			task, err := Start(m, ctx, func(context.Context, func(float64, string)) (int, error) {
				defer cleanup.Add(1)
				if mode == "panic" {
					panic("worker")
				}
				runtime.Goexit()
				return 0, nil
			}, func(_ int, err error) { received = err })
			if err != nil {
				t.Fatal(err)
			}
			_, stepErr := e.StepNext(ctx)
			if mode == "goexit" && !errors.Is(stepErr, execution.ErrGoexit) {
				t.Fatalf("outer Goexit not observable %v", stepErr)
			}
			if mode == "panic" && stepErr != nil {
				t.Fatalf("task handled panic escaped worker %v", stepErr)
			}
			closed(t, task.Done())
			if cleanup.Load() != 1 || m.Pending().Idle() {
				t.Fatal("cleanup missing or owner result slot prematurely free")
			}
			if err := a.Drain(); err != nil {
				t.Fatal(err)
			}
			closed(t, task.OwnerDone())
			if received == nil || !m.Pending().Idle() || !a.Pending().Idle() {
				t.Fatalf("failure cleanup received%v manager%+v app%+v", received, m.Pending(), a.Pending())
			}
		})
	}
}

func TestWindowCloseRevokesDeliveryButRetainsRunningWorkerLimit(t *testing.T) {
	a, m, e := candidate(t, "one")
	ctx := bounded(t)
	started, release := make(chan struct{}), make(chan struct{})
	task, err := Start(m, ctx, func(ctx context.Context, _ func(float64, string)) (int, error) {
		close(started)
		<-release
		return 0, ctx.Err()
	}, func(int, error) { t.Error("closed window commit ran") })
	if err != nil {
		t.Fatal(err)
	}
	stepDone := make(chan error, 1)
	go func() { _, err := e.StepNext(ctx); stepDone <- err }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := a.CloseWindow("one"); err != nil {
		t.Fatal(err)
	}
	closed(t, task.OwnerDone())
	if s := task.Snapshot(); !s.Running || s.WorkerDone || !s.OwnerDone || m.Pending().Idle() {
		t.Fatalf("running work disappeared %+v manager%+v", s, m.Pending())
	}
	select {
	case <-task.Done():
		t.Fatal("worker done before actual termination")
	default:
	}
	var survived bool
	if err := a.PostWindow("two", func() { survived = true }); err != nil {
		t.Fatal(err)
	}
	if err := a.Drain(); err != nil || !survived {
		t.Fatalf("survivor blocked %v %v", survived, err)
	}
	close(release)
	select {
	case err := <-stepDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	closed(t, task.Done())
	if !m.Pending().Idle() || !a.Pending().Idle() {
		t.Fatalf("post-terminal accounting lag %+v %+v", m.Pending(), a.Pending())
	}
}

func TestApplicationCloseDoesNotSkipAlreadyAcceptedPendingWorker(t *testing.T) {
	a, m, e := candidate(t, "")
	ctx := bounded(t)
	var work atomic.Int32
	task, err := Start(m, ctx, func(ctx context.Context, _ func(float64, string)) (int, error) { work.Add(1); return 0, ctx.Err() }, func(int, error) { t.Error("app closed commit") })
	if err != nil {
		t.Fatal(err)
	}
	a.Close()
	closed(t, task.OwnerDone())
	if m.Pending().Idle() {
		t.Fatal("accepted pending worker limit released by closure")
	}
	if _, err := e.StepNext(ctx); err != nil {
		t.Fatal(err)
	}
	closed(t, task.Done())
	if work.Load() != 1 || !m.Pending().Idle() {
		t.Fatalf("accepted canceled worker dropped work%d %+v", work.Load(), m.Pending())
	}
}
