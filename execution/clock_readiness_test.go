package execution

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func readiness(t *testing.T, c *ManualClock) ClockReadiness {
	t.Helper()
	s, err := c.ReadySnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if s.Due+s.Future != s.Pending || s.HasDeadline != (s.Pending != 0) {
		t.Fatalf("inconsistent single-lock inspection %+v", s)
	}
	return s
}

func TestReadinessRefusesNilAndUninitializedClockButInspectsClosedClock(t *testing.T) {
	var manual *ManualClock
	var real *RealClock
	for _, inspect := range []func() (ClockReadiness, error){manual.ReadySnapshot, real.ReadySnapshot, (&ManualClock{}).ReadySnapshot, (&RealClock{}).ReadySnapshot} {
		if s, err := inspect(); !errors.Is(err, ErrInvalidClock) || s != (ClockReadiness{}) {
			t.Fatalf("invalid clock manufactured idle: %+v error%v", s, err)
		}
	}
	c := manualClock(t, 1, 1)
	c.Close()
	if s := readiness(t, c); !s.Closed || !s.CurrentIdle() || s.HasDeadline {
		t.Fatalf("closed clock not inspectable %+v", s)
	}
}

func TestReadinessSeparatesDueNowFromFutureWithoutAdvanceOrCallback(t *testing.T) {
	c := manualClock(t, 2, 10)
	t.Cleanup(c.Close)
	start := c.Now()
	var calls atomic.Int32
	if _, err := c.AfterFunc(t.Context(), 0, func() { calls.Add(1) }); err != nil {
		t.Fatal(err)
	}
	future, err := c.AfterFunc(t.Context(), time.Second, func() { calls.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	before := c.Snapshot()
	for range 10 {
		s := readiness(t, c)
		if s.CurrentIdle() || s.Idle() || s.Due != 1 || s.Future != 1 || !s.NextDeadline.Equal(start) || !s.Now.Equal(start) {
			t.Fatalf("due inspection %+v", s)
		}
	}
	if calls.Load() != 0 || c.Snapshot() != before || !c.Now().Equal(start) {
		t.Fatal("inspection mutated scheduler or ran callback")
	}
	if _, err := c.Advance(0); err != nil {
		t.Fatal(err)
	}
	s := readiness(t, c)
	if !s.CurrentIdle() || s.Idle() || s.Due != 0 || s.Future != 1 || !s.NextDeadline.Equal(start.Add(time.Second)) || calls.Load() != 1 {
		t.Fatalf("future-only inspection %+v calls%d", s, calls.Load())
	}
	if err := future.Reset(0); err != nil {
		t.Fatal(err)
	}
	if s := readiness(t, c); s.CurrentIdle() || s.Due != 1 || s.Future != 0 || !s.NextDeadline.Equal(start) {
		t.Fatalf("reset deadline not reflected %+v", s)
	}
	if !future.Stop() {
		t.Fatal("stop failed")
	}
	if s := readiness(t, c); !s.CurrentIdle() || !s.Idle() || s.HasDeadline || !s.NextDeadline.IsZero() {
		t.Fatalf("stopped deadline retained %+v", s)
	}
}

func TestReadinessDuringNestedScheduleAndRunningReset(t *testing.T) {
	c := manualClock(t, 3, 10)
	t.Cleanup(c.Close)
	start := c.Now()
	var timer Timer
	var inside ClockReadiness
	var callbackErr error
	timer, callbackErr = c.AfterFunc(t.Context(), 0, func() {
		_, callbackErr = c.AfterFunc(t.Context(), 2*time.Second, func() {})
		if callbackErr == nil {
			callbackErr = timer.Reset(5 * time.Second)
		}
		inside, _ = c.ReadySnapshot()
	})
	if callbackErr != nil {
		t.Fatal(callbackErr)
	}
	if _, err := c.Advance(0); err != nil || callbackErr != nil {
		t.Fatalf("nested callback: %v / %v", err, callbackErr)
	}
	if inside.Pending != 2 || inside.Future != 2 || inside.Due != 0 || inside.Running != 1 || !inside.Advancing || inside.CurrentIdle() || !inside.NextDeadline.Equal(start.Add(2*time.Second)) {
		t.Fatalf("running/nested inspection %+v", inside)
	}
	if s := readiness(t, c); !s.CurrentIdle() || s.Running != 0 || s.Future != 2 || s.Advancing {
		t.Fatalf("future work incorrectly runs or blocks current idle %+v", s)
	}
}

func TestReadinessRetainsCanceledAdvanceRunningAndDueWork(t *testing.T) {
	c := manualClock(t, 3, 10)
	t.Cleanup(c.Close)
	start := c.Now()
	ctx := testContext(t)
	advanceContext, cancel := context.WithCancel(ctx)
	defer cancel()
	started, release := make(chan struct{}), make(chan struct{})
	if _, err := c.AfterFunc(ctx, time.Second, func() { close(started); <-release }); err != nil {
		t.Fatal(err)
	}
	var later atomic.Int32
	if _, err := c.AfterFunc(ctx, time.Second, func() { later.Add(1) }); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AfterFunc(ctx, 2*time.Second, func() { later.Add(1) }); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := c.AdvanceContext(advanceContext, 3*time.Second); done <- err }()
	waitSignal(t, ctx, started)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled advance: %v", err)
	}
	s := readiness(t, c)
	if s.Running != 1 || !s.Advancing || s.Due != 1 || s.Future != 1 || s.CurrentIdle() || !s.Now.Equal(start.Add(time.Second)) || later.Load() != 0 {
		t.Fatalf("cancellation hid or advanced work %+v later%d", s, later.Load())
	}
	if _, err := c.Advance(0); !errors.Is(err, ErrAdvancing) {
		t.Fatalf("blocked advance was released: %v", err)
	}
	close(release)
	for {
		s = readiness(t, c)
		if s.Running == 0 && !s.Advancing {
			break
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		runtime.Gosched()
	}
	if s.CurrentIdle() || s.Due != 1 || later.Load() != 0 || !s.Now.Equal(start.Add(time.Second)) {
		t.Fatalf("callback retirement skipped due work or time %+v later%d", s, later.Load())
	}
	if _, err := c.Advance(0); err != nil {
		t.Fatal(err)
	}
	if s := readiness(t, c); !s.CurrentIdle() || s.Future != 1 || later.Load() != 1 {
		t.Fatalf("explicit current-instant advance %+v later%d", s, later.Load())
	}
}

func TestRealReadinessRetainsClosedRunningCallbackAndFutureDeadline(t *testing.T) {
	c, err := NewRealClock(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	ctx := testContext(t)
	started, release := make(chan struct{}), make(chan struct{})
	if _, err := c.AfterFunc(ctx, 0, func() { close(started); <-release }); err != nil {
		t.Fatal(err)
	}
	future, err := c.AfterFunc(ctx, time.Hour, func() { t.Error("future callback fired") })
	if err != nil {
		t.Fatal(err)
	}
	waitSignal(t, ctx, started)
	s, err := c.ReadySnapshot()
	if err != nil || s.CurrentIdle() || s.Running != 1 || s.Future != 1 || s.Due != 0 || !s.NextDeadline.Equal(future.Snapshot().Deadline) {
		t.Fatalf("real clock facts %+v error%v", s, err)
	}
	c.Close()
	s, err = c.ReadySnapshot()
	if err != nil || !s.Closed || s.Running != 1 || s.Pending != 0 || s.HasDeadline || s.CurrentIdle() {
		t.Fatalf("close hid running callback %+v error%v", s, err)
	}
	close(release)
	for !c.Snapshot().Idle() && ctx.Err() == nil {
		runtime.Gosched()
	}
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	s, err = c.ReadySnapshot()
	if err != nil || !s.CurrentIdle() || !s.Idle() {
		t.Fatalf("retirement %+v error%v", s, err)
	}
}

func TestConcurrentReadinessIsConsistentWithRegisteredTimerMutations(t *testing.T) {
	c := manualClock(t, 8, 10)
	t.Cleanup(c.Close)
	var producers sync.WaitGroup
	for range 4 {
		producers.Go(func() {
			for i := range 40 {
				timer, err := c.AfterFunc(t.Context(), time.Duration(i%2)*time.Second, func() { t.Error("inspection fired callback") })
				if err != nil {
					t.Error(err)
					return
				}
				if err := timer.Reset(time.Second); err != nil {
					t.Error(err)
					return
				}
				timer.Stop()
			}
		})
	}
	for range 100 {
		s := readiness(t, c)
		if s.Running != 0 || s.Advancing || !s.Now.Equal(time.Unix(1000, 0)) {
			t.Fatalf("inspection advanced work %+v", s)
		}
	}
	producers.Wait()
	if s := readiness(t, c); !s.CurrentIdle() || !s.Idle() || s.HasDeadline {
		t.Fatalf("mutations left stale readiness %+v", s)
	}
}
