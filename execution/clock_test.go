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

func manualClock(t *testing.T, capacity, budget int) *ManualClock {
	t.Helper()
	c, err := NewManualClock(time.Unix(1000, 0), capacity, budget)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestManualTimeMovesOnlyOnExplicitAdvanceAndNestedOrderIsStable(t *testing.T) {
	c := manualClock(t, 8, 20)
	start := c.Now()
	var events []string
	var instants []time.Duration
	record := func(value string) {
		events = append(events, value)
		instants = append(instants, c.Now().Sub(start))
	}
	if _, err := c.AfterFunc(context.Background(), 5*time.Second, func() {
		record("first")
		if _, err := c.AfterFunc(context.Background(), 0, func() { record("nested-zero") }); err != nil {
			t.Error(err)
		}
		if _, err := c.AfterFunc(context.Background(), 2*time.Second, func() { record("nested-later") }); err != nil {
			t.Error(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AfterFunc(context.Background(), 5*time.Second, func() { record("second") }); err != nil {
		t.Fatal(err)
	}
	if !c.Now().Equal(start) || len(events) != 0 {
		t.Fatal("timer admission advanced time or invoked a callback")
	}
	r, err := c.Advance(10 * time.Second)
	if err != nil || r.Fired != 4 || !r.Now.Equal(start.Add(10*time.Second)) {
		t.Fatalf("advance %+v error %v", r, err)
	}
	if !reflect.DeepEqual(events, []string{"first", "second", "nested-zero", "nested-later"}) {
		t.Fatalf("same-time/nested order %v", events)
	}
	if !reflect.DeepEqual(instants, []time.Duration{5 * time.Second, 5 * time.Second, 5 * time.Second, 7 * time.Second}) {
		t.Fatalf("callback scheduled instants %v", instants)
	}
}

func TestStopResetAndRefusalLeaveExistingTimerAtomic(t *testing.T) {
	c := manualClock(t, 2, 10)
	var order []int
	a, _ := c.AfterFunc(context.Background(), time.Second, func() { order = append(order, 1) })
	b, _ := c.AfterFunc(context.Background(), time.Second, func() { order = append(order, 2) })
	before := a.Snapshot()
	if err := a.Reset(-time.Nanosecond); !errors.Is(err, ErrDuration) || a.Snapshot() != before {
		t.Fatalf("invalid reset changed timer: %+v error %v", a.Snapshot(), err)
	}
	if _, err := c.AfterFunc(context.Background(), 0, func() { order = append(order, 99) }); !errors.Is(err, ErrCapacity) {
		t.Fatalf("full timer admission %v", err)
	}
	// Re-arming a pending timer uses its existing slot but a new tie sequence.
	if err := a.Reset(time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Advance(time.Second); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []int{2, 1}) {
		t.Fatalf("reset tie order %v", order)
	}
	if a.Stop() || b.Stop() {
		t.Fatal("stop reported an already fired occurrence pending")
	}
	if err := b.Reset(0); err != nil || !b.Stop() || b.Stop() {
		t.Fatalf("reset/stop terminal timer: %v", err)
	}
	if r, err := c.Advance(0); err != nil || r.Fired != 0 {
		t.Fatalf("stopped timer fired %+v error %v", r, err)
	}
}

func TestZeroDelayRequiresAdvanceAndCanceledLifetimeCannotReset(t *testing.T) {
	c := manualClock(t, 2, 10)
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	timer, err := c.AfterFunc(ctx, 0, func() { calls.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	// Advance itself checks lifetime, independently of asynchronous AfterFunc cleanup.
	if r, err := c.Advance(0); err != nil || r.Fired != 0 || calls.Load() != 0 {
		t.Fatalf("canceled lifetime fired %+v error %v", r, err)
	}
	if err := timer.Reset(0); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled timer rearmed: %v", err)
	}
	if _, err := c.AfterFunc(ctx, 0, func() {}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled admission %v", err)
	}
	if !c.Snapshot().Idle() {
		t.Fatalf("canceled pending occurrence leaked %+v", c.Snapshot())
	}
	if _, err := c.AfterFunc(context.Background(), 0, func() { calls.Add(1) }); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Fatal("zero-delay manual timer ran without explicit Advance")
	}
	if r, err := c.Advance(0); err != nil || r.Fired != 1 || calls.Load() != 1 {
		t.Fatalf("zero advance %+v error %v", r, err)
	}
}

func TestBudgetBoundsImmediateSelfResetAndRetainsPendingOccurrence(t *testing.T) {
	c := manualClock(t, 2, 3)
	start := c.Now()
	var timer Timer
	var calls int
	timer, _ = c.AfterFunc(context.Background(), 0, func() {
		calls++
		if err := timer.Reset(0); err != nil {
			t.Error(err)
		}
	})
	r, err := c.Advance(time.Second)
	if !errors.Is(err, ErrAdvanceBudget) || r.Fired != 3 || calls != 3 || !r.Now.Equal(start) || !r.Target.Equal(start.Add(time.Second)) {
		t.Fatalf("budget result %+v calls %d error %v", r, calls, err)
	}
	if s := c.Snapshot(); s.Pending != 1 || s.Running != 0 || s.Idle() || s.Completed != 3 {
		t.Fatalf("budget discarded work %+v", s)
	}
	if !timer.Stop() {
		t.Fatal("retained occurrence not stoppable")
	}
	if r, err := c.Advance(time.Second); err != nil || r.Fired != 0 || !r.Now.Equal(start.Add(time.Second)) {
		t.Fatalf("explicit continuation %+v error %v", r, err)
	}
}

func TestManualCloseDuringRunningCallbackCancelsPendingButDoesNotHideRunning(t *testing.T) {
	c := manualClock(t, 3, 10)
	ctx := testContext(t)
	started, release := make(chan struct{}), make(chan struct{})
	timer, _ := c.AfterFunc(context.Background(), 0, func() { close(started); <-release })
	if _, err := c.AfterFunc(context.Background(), time.Second, func() { t.Error("closed pending callback ran") }); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := c.Advance(2 * time.Second); done <- err }()
	waitSignal(t, ctx, started)
	if _, err := c.Advance(0); !errors.Is(err, ErrAdvancing) {
		t.Fatalf("concurrent advance %v", err)
	}
	c.Close()
	if s := c.Snapshot(); s.Idle() || !s.Closed || s.Pending != 0 || s.Running != 1 {
		t.Fatalf("closed running clock hidden %+v", s)
	}
	if err := timer.Reset(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("reset after clock close %v", err)
	}
	close(release)
	select {
	case err := <-done:
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("advance after concurrent close %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if !c.Snapshot().Idle() || timer.Snapshot().Running != 0 {
		t.Fatal("callback completion did not retire running occurrence")
	}
}

func TestCanceledAdvanceWaitRetainsRunningAdmissionAndDoesNotSkipVirtualTime(t *testing.T) {
	c := manualClock(t, 2, 10)
	ctx := testContext(t)
	start := c.Now()
	started, release := make(chan struct{}), make(chan struct{})
	var laterCalls atomic.Int32
	if _, err := c.AfterFunc(context.Background(), time.Second, func() { close(started); <-release }); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AfterFunc(context.Background(), 2*time.Second, func() { laterCalls.Add(1) }); err != nil {
		t.Fatal(err)
	}
	waitCtx, cancel := context.WithCancel(ctx)
	type outcome struct {
		result AdvanceResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() { r, err := c.AdvanceContext(waitCtx, 10*time.Second); done <- outcome{r, err} }()
	waitSignal(t, ctx, started)
	cancel()
	select {
	case o := <-done:
		if !errors.Is(o.err, context.Canceled) || o.result.Fired != 1 || !o.result.Now.Equal(start.Add(time.Second)) {
			t.Fatalf("canceled result %+v error %v", o.result, o.err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if s := c.Snapshot(); s.Idle() || !s.Advancing || s.Running != 1 || s.Pending != 1 {
		t.Fatalf("canceled waiter hid active advance %+v", s)
	}
	for range 10 {
		if _, err := c.AdvanceContext(ctx, 10*time.Second); !errors.Is(err, ErrAdvancing) {
			t.Fatalf("canceled advance released admission prematurely: %v", err)
		}
	}
	if _, err := c.AfterFunc(context.Background(), 0, func() {}); !errors.Is(err, ErrCapacity) {
		t.Fatalf("canceled waiter released running slot: %v", err)
	}
	if !c.Now().Equal(start.Add(time.Second)) || laterCalls.Load() != 0 {
		t.Fatal("canceled advance skipped time or started a later callback")
	}
	close(release)
	for c.Snapshot().Advancing && ctx.Err() == nil {
		runtime.Gosched()
	}
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	if !c.Now().Equal(start.Add(time.Second)) || c.Snapshot().Pending != 1 || laterCalls.Load() != 0 {
		t.Fatal("detached terminal cleanup implicitly resumed advance")
	}
	if r, err := c.AdvanceContext(ctx, time.Second); err != nil || r.Fired != 1 || laterCalls.Load() != 1 {
		t.Fatalf("explicit continuation %+v error %v", r, err)
	}
}

func TestRefusedAdvanceContextsNeverRunDueCallbacks(t *testing.T) {
	c := manualClock(t, 1, 10)
	var calls atomic.Int32
	if _, err := c.AfterFunc(context.Background(), 0, func() { calls.Add(1) }); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.AdvanceContext(ctx, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-canceled advance %v", err)
	}
	if _, err := c.AdvanceContext(nil, time.Second); !errors.Is(err, ErrContext) {
		t.Fatalf("nil advance context %v", err)
	}
	if s := c.Snapshot(); calls.Load() != 0 || s.Pending != 1 || s.Running != 0 || s.Advancing {
		t.Fatalf("refused context started work %+v calls %d", s, calls.Load())
	}
	c.Close()
	if _, err := c.AdvanceContext(context.Background(), time.Second); !errors.Is(err, ErrClosed) || calls.Load() != 0 {
		t.Fatalf("closed advance ran callback: %v calls %d", err, calls.Load())
	}
}

func TestCanceledAdvanceLateGoexitRetainsObservableTerminalFailure(t *testing.T) {
	c := manualClock(t, 1, 10)
	ctx := testContext(t)
	started, release := make(chan struct{}), make(chan struct{})
	timer, err := c.AfterFunc(context.Background(), 0, func() {
		close(started)
		<-release
		runtime.Goexit()
	})
	if err != nil {
		t.Fatal(err)
	}
	waitCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { _, err := c.AdvanceContext(waitCtx, 0); done <- err }()
	waitSignal(t, ctx, started)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("advance wait cancellation %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	close(release)
	for c.Snapshot().Advancing && ctx.Err() == nil {
		runtime.Gosched()
	}
	if ctx.Err() != nil || !c.Snapshot().Idle() || c.Snapshot().Failed != 1 || !errors.Is(timer.Snapshot().LastErr, ErrGoexit) {
		t.Fatalf("late failure lost: clock %+v timer %+v wait %v", c.Snapshot(), timer.Snapshot(), ctx.Err())
	}
}

func TestManualCallbackPanicGoexitAndNestedAdvanceAreObservable(t *testing.T) {
	c := manualClock(t, 4, 10)
	var cleanup atomic.Int32
	if _, err := c.AfterFunc(context.Background(), 0, func() { defer cleanup.Add(1); panic("timer") }); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AfterFunc(context.Background(), 0, func() { defer cleanup.Add(1); runtime.Goexit() }); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AfterFunc(context.Background(), 0, func() {
		defer cleanup.Add(1)
		if _, err := c.Advance(0); !errors.Is(err, ErrAdvancing) {
			t.Errorf("nested advance %v", err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	r, err := c.Advance(0)
	var panicErr *PanicError
	if !errors.Is(err, ErrGoexit) || !errors.As(err, &panicErr) || r.Fired != 3 || len(r.Failures) != 2 || cleanup.Load() != 3 {
		t.Fatalf("callback failures result %+v error %v cleanup %d", r, err, cleanup.Load())
	}
	if s := c.Snapshot(); !s.Idle() || s.Failed != 2 || s.Completed != 3 {
		t.Fatalf("timer failure accounting %+v", s)
	}
}

func TestClockValidationAndExhaustionHaveNoPartialMutation(t *testing.T) {
	if _, err := NewRealClock(0); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("invalid real clock: %v", err)
	}
	if _, err := NewManualClock(time.Time{}, 1, 0); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("invalid callback budget: %v", err)
	}
	c := manualClock(t, 2, 10)
	if _, err := c.AfterFunc(nil, 0, func() {}); !errors.Is(err, ErrContext) {
		t.Fatalf("nil lifetime: %v", err)
	}
	if _, err := c.AfterFunc(context.Background(), 0, nil); !errors.Is(err, ErrInvalidJob) {
		t.Fatalf("nil callback: %v", err)
	}
	if _, err := c.AfterFunc(context.Background(), -1, func() {}); !errors.Is(err, ErrDuration) {
		t.Fatalf("negative duration: %v", err)
	}
	start := c.Now()
	if _, err := c.Advance(-1); !errors.Is(err, ErrDuration) || !c.Now().Equal(start) {
		t.Fatalf("negative advance mutated clock: %v", err)
	}
	c.base.position = math.MaxInt64
	if _, err := c.Advance(1); !errors.Is(err, ErrClockExhausted) {
		t.Fatalf("advance overflow: %v", err)
	}
	if _, err := c.AfterFunc(context.Background(), 1, func() {}); !errors.Is(err, ErrClockExhausted) {
		t.Fatalf("deadline overflow: %v", err)
	}
	c.base.position = 0
	timer, _ := c.AfterFunc(context.Background(), 10, func() {})
	before := timer.Snapshot()
	c.base.sequence = math.MaxUint64
	if err := timer.Reset(0); !errors.Is(err, ErrClockExhausted) || timer.Snapshot() != before {
		t.Fatalf("sequence overflow mutated timer: %v", err)
	}
	if !timer.Stop() {
		t.Fatal("overflow refusal lost pending timer")
	}
	c.base.next = ID(math.MaxUint64)
	if _, err := c.AfterFunc(context.Background(), 0, func() {}); !errors.Is(err, ErrIDExhausted) {
		t.Fatalf("timer ID overflow: %v", err)
	}
}

func TestResetRunningOccurrenceRequiresAnotherOutstandingSlot(t *testing.T) {
	c := manualClock(t, 1, 10)
	var timer Timer
	timer, _ = c.AfterFunc(context.Background(), 0, func() {
		if err := timer.Reset(0); !errors.Is(err, ErrCapacity) {
			t.Errorf("reset exceeded running slot bound: %v", err)
		}
	})
	if r, err := c.Advance(0); err != nil || r.Fired != 1 || !c.Snapshot().Idle() {
		t.Fatalf("running bound result %+v error %v", r, err)
	}
}

func TestRealClockLifetimeAndRunningBound(t *testing.T) {
	c, err := NewRealClock(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	ctx := testContext(t)
	lifetime, cancel := context.WithCancel(ctx)
	started, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	timer, err := c.AfterFunc(lifetime, 0, func() { close(started); <-release; close(finished) })
	if err != nil {
		t.Fatal(err)
	}
	waitSignal(t, ctx, started)
	if err := timer.Reset(0); !errors.Is(err, ErrCapacity) {
		t.Fatalf("real running bound: %v", err)
	}
	if timer.Stop() {
		t.Fatal("Stop claimed to interrupt running callback")
	}
	cancel()
	if s := c.Snapshot(); s.Idle() || s.Running != 1 {
		t.Fatalf("lifetime cancellation hid running callback %+v", s)
	}
	close(release)
	waitSignal(t, ctx, finished)
	for !c.Snapshot().Idle() && ctx.Err() == nil {
		runtime.Gosched()
	}
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	if err := timer.Reset(0); !errors.Is(err, context.Canceled) {
		t.Fatalf("real canceled reset %v", err)
	}
	if c.Snapshot().Completed != 1 {
		t.Fatalf("real completion %+v", c.Snapshot())
	}
}

func TestRealPendingLifetimeCancellationReturnsCreditWithoutClockPolling(t *testing.T) {
	c, err := NewRealClock(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	ctx := testContext(t)
	lifetime, cancel := context.WithCancel(ctx)
	if _, err := c.AfterFunc(lifetime, time.Hour, func() { t.Error("canceled pending timer ran") }); err != nil {
		t.Fatal(err)
	}
	cancel()
	for !c.Snapshot().Idle() && ctx.Err() == nil {
		runtime.Gosched()
	}
	if ctx.Err() != nil || c.Snapshot().Stopped != 1 {
		t.Fatalf("lifetime did not release pending credit: %+v error %v", c.Snapshot(), ctx.Err())
	}
	fired := make(chan struct{})
	if _, err := c.AfterFunc(context.Background(), 0, func() { close(fired) }); err != nil {
		t.Fatalf("canceled timer retained admission credit: %v", err)
	}
	waitSignal(t, ctx, fired)
}

func TestConcurrentTimerAdmissionStopResetAndExplicitAdvance(t *testing.T) {
	c := manualClock(t, 32, 64)
	ctx := testContext(t)
	var callbacks atomic.Uint64
	var producers sync.WaitGroup
	stopAdvance := make(chan struct{})
	doneAdvance := make(chan struct{})
	go func() {
		defer close(doneAdvance)
		for {
			select {
			case <-stopAdvance:
				return
			case <-ctx.Done():
				return
			default:
			}
			if _, err := c.Advance(0); err != nil && !errors.Is(err, ErrAdvanceBudget) {
				t.Errorf("concurrent advance %v", err)
				return
			}
			runtime.Gosched()
		}
	}()
	for range 4 {
		producers.Go(func() {
			for range 100 {
				if ctx.Err() != nil {
					return
				}
				timer, err := c.AfterFunc(ctx, 0, func() { callbacks.Add(1) })
				if errors.Is(err, ErrCapacity) {
					runtime.Gosched()
					continue
				}
				if err != nil {
					t.Error(err)
					return
				}
				if err := timer.Reset(0); err != nil && !errors.Is(err, ErrCapacity) {
					t.Error(err)
				}
				timer.Stop()
				if s := c.Snapshot(); s.Pending+s.Running > 32 {
					t.Errorf("timer bound exceeded %+v", s)
				}
			}
		})
	}
	producers.Wait()
	close(stopAdvance)
	waitSignal(t, ctx, doneAdvance)
	c.Close()
	if s := c.Snapshot(); !s.Idle() || s.Fired != callbacks.Load() || s.Fired != s.Completed || s.Failed != 0 {
		t.Fatalf("concurrent clock accounting %+v callbacks %d", s, callbacks.Load())
	}
}
