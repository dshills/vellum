package execution

import (
	"context"
	"errors"
	"math"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func newManual(t *testing.T, size int) *Manual {
	t.Helper()
	m, err := NewManual(size)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func waitSignal(t *testing.T, ctx context.Context, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func TestCapacityRefusalNeverAdmitsOrExecutesCallback(t *testing.T) {
	if _, err := NewManual(0); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("zero capacity: %v", err)
	}
	m := newManual(t, 1)
	var calls atomic.Int32
	j, err := m.Enqueue(func() { calls.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Submit(func() { calls.Add(100) }); !errors.Is(err, ErrCapacity) {
		t.Fatalf("full admission: %v", err)
	}
	if err := m.Submit(nil); !errors.Is(err, ErrInvalidJob) {
		t.Fatalf("nil callback: %v", err)
	}
	if got := m.PendingIDs(); !reflect.DeepEqual(got, []ID{j.ID()}) {
		t.Fatalf("pending %v", got)
	}
	s := m.Snapshot()
	if s.Idle() || s.Pending != 1 || s.Running != 0 || s.Accepted != 1 || s.Refused != 2 {
		t.Fatalf("pending accounting %+v", s)
	}
	if r, complete := j.Result(); complete || r.State != Pending {
		t.Fatalf("pending result %+v terminal %v", r, complete)
	}
	if err := m.Step(testContext(t), j.ID()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || !m.Snapshot().Idle() {
		t.Fatalf("refused callback ran: calls %d snapshot %+v", calls.Load(), m.Snapshot())
	}
}

func TestSelectedOrderAndFIFOHaveNoHiddenExecution(t *testing.T) {
	m := newManual(t, 3)
	var order []int
	jobs := make([]*Job, 3)
	for i := range jobs {
		value := i
		var err error
		jobs[i], err = m.Enqueue(func() { order = append(order, value) })
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(order) != 0 {
		t.Fatal("manual executed before a step")
	}
	ids := m.PendingIDs()
	ids[0] = 1000
	if err := m.Step(testContext(t), jobs[2].ID()); err != nil {
		t.Fatal(err)
	}
	for i := range 2 {
		id, err := m.StepNext(testContext(t))
		if err != nil || id != jobs[i].ID() {
			t.Fatalf("FIFO %d: ID %d error %v", i, id, err)
		}
	}
	if !reflect.DeepEqual(order, []int{2, 0, 1}) {
		t.Fatalf("order %v", order)
	}
	if err := m.Step(testContext(t), jobs[2].ID()); !errors.Is(err, ErrNotPending) {
		t.Fatalf("repeat step: %v", err)
	}
	if _, err := m.StepNext(testContext(t)); !errors.Is(err, ErrNotPending) {
		t.Fatalf("empty FIFO: %v", err)
	}
}

func TestCloseRetainsAcceptedPendingJobsAndCleanupAfterCancellation(t *testing.T) {
	m := newManual(t, 2)
	workCtx, cancel := context.WithCancel(context.Background())
	var cleanup atomic.Int32
	j, err := m.Enqueue(func() {
		defer cleanup.Add(1)
		if workCtx.Err() == nil {
			t.Error("job did not observe its captured canceled context")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	m.Close()
	m.Close()
	if err := m.Submit(func() { t.Error("closed submission ran") }); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed admission: %v", err)
	}
	if s := m.Snapshot(); s.Idle() || !s.Closed || s.Pending != 1 || s.Completed != 0 {
		t.Fatalf("close discarded accepted work %+v", s)
	}
	if err := m.Step(testContext(t), j.ID()); err != nil {
		t.Fatal(err)
	}
	if cleanup.Load() != 1 || !m.Snapshot().Idle() || m.Snapshot().Completed != 1 {
		t.Fatalf("cleanup %d accounting %+v", cleanup.Load(), m.Snapshot())
	}
}

func TestCanceledStepWaitDoesNotReportRunningJobIdleOrReleaseCapacity(t *testing.T) {
	m := newManual(t, 1)
	ctx := testContext(t)
	started, release := make(chan struct{}), make(chan struct{})
	j, err := m.Enqueue(func() { close(started); <-release })
	if err != nil {
		t.Fatal(err)
	}
	waitCtx, cancel := context.WithCancel(ctx)
	stepDone := make(chan error, 1)
	go func() { stepDone <- m.Step(waitCtx, j.ID()) }()
	waitSignal(t, ctx, started)
	cancel()
	select {
	case err := <-stepDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("wait cancellation: %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	m.Close()
	if s := m.Snapshot(); s.Idle() || s.Pending != 0 || s.Running != 1 || s.Completed != 0 {
		t.Fatalf("running work hidden %+v", s)
	}
	if err := m.Step(ctx, j.ID()); !errors.Is(err, ErrRunning) {
		t.Fatalf("second start: %v", err)
	}
	select {
	case <-j.Done():
		t.Fatal("canceled wait terminated callback")
	default:
	}
	close(release)
	if err := j.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); !s.Idle() || s.Completed != 1 || !s.Closed {
		t.Fatalf("completion accounting %+v", s)
	}
}

func TestRunningJobsCountTowardOutstandingLimit(t *testing.T) {
	m := newManual(t, 1)
	ctx := testContext(t)
	started, release := make(chan struct{}), make(chan struct{})
	j, _ := m.Enqueue(func() { close(started); <-release })
	done := make(chan error, 1)
	go func() { done <- m.Step(ctx, j.ID()) }()
	waitSignal(t, ctx, started)
	if err := m.Submit(func() { t.Error("over-capacity callback ran") }); !errors.Is(err, ErrCapacity) {
		t.Fatalf("running slot admission: %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := m.Submit(func() {}); err != nil {
		t.Fatalf("terminal slot not released: %v", err)
	}
	if _, err := m.StepNext(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCanceledBeforeStepPreservesPendingAndDoesNotExecute(t *testing.T) {
	m := newManual(t, 1)
	var calls atomic.Int32
	j, _ := m.Enqueue(func() { calls.Add(1) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := m.Step(ctx, j.ID()); !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-canceled step: %v", err)
	}
	if err := m.Step(nil, j.ID()); !errors.Is(err, ErrContext) {
		t.Fatalf("nil wait context: %v", err)
	}
	if s := m.Snapshot(); calls.Load() != 0 || s.Pending != 1 || s.Running != 0 {
		t.Fatalf("pre-canceled work changed %+v calls %d", s, calls.Load())
	}
	if err := m.Step(testContext(t), j.ID()); err != nil {
		t.Fatal(err)
	}
	// An already published terminal result wins an unrelated canceled wait.
	if err := j.Wait(ctx); err != nil {
		t.Fatalf("completed wait returned cancellation: %v", err)
	}
}

func TestPanicAndGoexitAreTerminalObservableAndCallerSurvives(t *testing.T) {
	for _, mode := range []string{"panic", "goexit", "panic-nil"} {
		t.Run(mode, func(t *testing.T) {
			m := newManual(t, 1)
			var cleaned atomic.Bool
			payload := &struct{ Value string }{"preserved"}
			j, _ := m.Enqueue(func() {
				defer cleaned.Store(true)
				switch mode {
				case "panic":
					panic(payload)
				case "goexit":
					runtime.Goexit()
				default:
					panic(nil)
				}
			})
			err := m.Step(testContext(t), j.ID())
			if err == nil || !cleaned.Load() {
				t.Fatalf("failure/cleanup not observed: %v cleaned %v", err, cleaned.Load())
			}
			if mode == "goexit" && !errors.Is(err, ErrGoexit) {
				t.Fatalf("Goexit: %v", err)
			}
			if mode == "panic" {
				var p *PanicError
				if !errors.As(err, &p) || p.Value != payload {
					t.Fatalf("panic payload lost: %v", err)
				}
			}
			if r, complete := j.Result(); !complete || r.State != Failed || r.Err != err {
				t.Fatalf("terminal failure %+v terminal %v", r, complete)
			}
			if s := m.Snapshot(); !s.Idle() || s.Completed != 1 || s.Failed != 1 {
				t.Fatalf("failure accounting %+v", s)
			}
			// A failed callback must not consume the executor's only slot forever.
			if err := m.Submit(func() {}); err != nil {
				t.Fatal(err)
			}
			if _, err := m.StepNext(testContext(t)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCallbackReentrancyDoesNotHoldSchedulerLock(t *testing.T) {
	m := newManual(t, 2)
	ctx := testContext(t)
	var outer *Job
	var childRan atomic.Bool
	outer, _ = m.Enqueue(func() {
		if m.Snapshot().Running != 1 {
			t.Error("running callback missing")
		}
		if err := m.Step(ctx, outer.ID()); !errors.Is(err, ErrRunning) {
			t.Errorf("self-start was not refused: %v", err)
		}
		child, err := m.Enqueue(func() { childRan.Store(true) })
		if err != nil {
			t.Error(err)
			return
		}
		if err := m.Step(ctx, child.ID()); err != nil {
			t.Error(err)
		}
		m.Close()
		if err := m.Submit(func() {}); !errors.Is(err, ErrClosed) {
			t.Errorf("reentrant closed admission: %v", err)
		}
	})
	if err := m.Step(ctx, outer.ID()); err != nil {
		t.Fatal(err)
	}
	if !childRan.Load() || !m.Snapshot().Idle() {
		t.Fatalf("nested work did not terminate %+v", m.Snapshot())
	}
}

func TestConcurrentSubmitAndStepExecuteEveryAcceptedJobOnce(t *testing.T) {
	const producers, perProducer, steppers = 4, 80, 4
	m := newManual(t, 16)
	ctx := testContext(t)
	counts := make([]atomic.Int32, producers*perProducer)
	var submitters, runners sync.WaitGroup
	for range steppers {
		runners.Go(func() {
			for ctx.Err() == nil {
				_, err := m.StepNext(ctx)
				if err != nil && !errors.Is(err, ErrNotPending) {
					t.Errorf("concurrent step: %v", err)
					return
				}
				s := m.Snapshot()
				if s.Pending+s.Running > 16 {
					t.Errorf("outstanding bound exceeded: %+v", s)
				}
				if s.Closed && s.Idle() {
					return
				}
				runtime.Gosched()
			}
		})
	}
	for producer := range producers {
		submitters.Go(func() {
			for i := range perProducer {
				index := producer*perProducer + i
				for ctx.Err() == nil {
					err := m.Submit(func() { counts[index].Add(1) })
					if err == nil {
						break
					}
					if !errors.Is(err, ErrCapacity) {
						t.Errorf("concurrent admission: %v", err)
						return
					}
					runtime.Gosched()
				}
			}
		})
	}
	done := make(chan struct{})
	go func() { submitters.Wait(); m.Close(); runners.Wait(); close(done) }()
	waitSignal(t, ctx, done)
	for i := range counts {
		if got := counts[i].Load(); got != 1 {
			t.Errorf("callback %d executed %d times", i, got)
		}
	}
	if s := m.Snapshot(); !s.Idle() || s.Accepted != uint64(len(counts)) || s.Completed != s.Accepted || s.Failed != 0 {
		t.Fatalf("concurrent terminal accounting %+v", s)
	}
}

func TestDefaultGoroutineRetainsRunningJobsAfterCloseAndReportsFailure(t *testing.T) {
	var g Goroutine
	ctx := testContext(t)
	started, release := make(chan struct{}), make(chan struct{})
	j, err := g.Enqueue(func() { close(started); <-release; runtime.Goexit() })
	if err != nil {
		t.Fatal(err)
	}
	waitSignal(t, ctx, started)
	g.Close()
	if err := g.Submit(func() { t.Error("closed goroutine callback ran") }); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed goroutine admission: %v", err)
	}
	if s := g.Snapshot(); s.Idle() || s.Running != 1 || !s.Closed {
		t.Fatalf("running goroutine hidden %+v", s)
	}
	close(release)
	if err := j.Wait(ctx); !errors.Is(err, ErrGoexit) {
		t.Fatalf("default terminal failure: %v", err)
	}
	if s := g.Snapshot(); !s.Idle() || s.Completed != 1 || s.Failed != 1 {
		t.Fatalf("default completion %+v", s)
	}
}

func TestExhaustedIDsRefuseRatherThanReuseIdentity(t *testing.T) {
	m := newManual(t, 1)
	m.ledger.next = ID(math.MaxUint64)
	if err := m.Submit(func() { t.Error("exhausted admission ran") }); !errors.Is(err, ErrIDExhausted) {
		t.Fatalf("exhausted identity: %v", err)
	}
	if s := m.Snapshot(); !s.Idle() || s.Accepted != 0 || s.Refused != 1 {
		t.Fatalf("exhaustion accounting %+v", s)
	}
}
