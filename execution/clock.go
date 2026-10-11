package execution

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

var (
	ErrDuration       = errors.New("execution: duration must not be negative")
	ErrClockExhausted = errors.New("execution: clock position or timer sequence exhausted")
	ErrAdvancing      = errors.New("execution: manual clock already advancing")
	ErrAdvanceBudget  = errors.New("execution: manual callback budget exhausted")
)

// Clock schedules cancelable timer callbacks. A zero delay is valid: a real
// clock runs it asynchronously, a manual clock requires explicit Advance(0).
type Clock interface {
	Now() time.Time
	AfterFunc(context.Context, time.Duration, func()) (Timer, error)
}

type Timer interface {
	Stop() bool
	Reset(time.Duration) error
	Snapshot() TimerSnapshot
}

type TimerSnapshot struct {
	ID       ID
	Pending  bool
	Running  int
	Deadline time.Time
	LastErr  error
}

type ClockSnapshot struct {
	Pending, Running                int
	Armed, Fired, Completed, Failed uint64
	Stopped, Refused                uint64
	Closed                          bool
	Advancing                       bool
}

func (s ClockSnapshot) Idle() bool { return s.Pending == 0 && s.Running == 0 && !s.Advancing }

type clockBase struct {
	mu        sync.Mutex
	capacity  int
	real      bool
	anchor    time.Time
	position  int64
	next      ID
	sequence  uint64
	s         ClockSnapshot
	timers    map[*clockTimer]struct{}
	advancing bool
	detached  bool
}

type clockTimer struct {
	clock       *clockBase
	id          ID
	ctx         context.Context
	fn          func()
	pending     bool
	running     int
	deadline    time.Time
	position    int64
	sequence    uint64
	lastErr     error
	native      *time.Timer
	stopContext func() bool
}

func (c *clockBase) nowLocked() time.Time {
	if c.real {
		return time.Now()
	}
	return c.anchor.Add(time.Duration(c.position))
}

func (c *clockBase) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nowLocked()
}

func (c *clockBase) snapshot() ClockSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.s
	s.Advancing = c.advancing
	return s
}

func (c *clockBase) after(ctx context.Context, delay time.Duration, fn func()) (Timer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ctx == nil {
		c.s.Refused++
		return nil, ErrContext
	}
	if fn == nil {
		c.s.Refused++
		return nil, ErrInvalidJob
	}
	if c.next == ID(math.MaxUint64) {
		c.s.Refused++
		return nil, ErrIDExhausted
	}
	t := &clockTimer{clock: c, ctx: ctx, fn: fn}
	if err := t.armLocked(delay); err != nil {
		return nil, err
	}
	c.next++
	t.id = c.next
	return t, nil
}

// armLocked validates all refusal conditions before changing an existing timer.
// A running callback and its reset pending occurrence consume separate slots.
func (t *clockTimer) armLocked(delay time.Duration) error {
	c := t.clock
	refuse := func(err error) error { c.s.Refused++; return err }
	if delay < 0 {
		return refuse(ErrDuration)
	}
	if c.s.Closed {
		return refuse(ErrClosed)
	}
	if err := t.ctx.Err(); err != nil {
		return refuse(err)
	}
	if !t.pending && c.s.Pending+c.s.Running >= c.capacity {
		return refuse(ErrCapacity)
	}
	if c.sequence == math.MaxUint64 || (!c.real && int64(delay) > math.MaxInt64-c.position) {
		return refuse(ErrClockExhausted)
	}
	c.sequence++
	if t.pending {
		if t.native != nil {
			t.native.Stop()
		}
	} else {
		t.pending = true
		c.s.Pending++
	}
	t.sequence = c.sequence
	t.position = c.position + int64(delay)
	t.deadline = c.nowLocked().Add(delay)
	c.s.Armed++
	c.timers[t] = struct{}{}
	if t.stopContext == nil {
		t.stopContext = context.AfterFunc(t.ctx, func() { t.Stop() })
	}
	if c.real {
		sequence := t.sequence
		t.native = time.AfterFunc(delay, func() {
			c.mu.Lock()
			claimed := t.claimLocked(sequence)
			c.mu.Unlock()
			if claimed {
				t.run(nil)
			}
		})
	}
	return nil
}

func (t *clockTimer) Reset(delay time.Duration) error {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	return t.armLocked(delay)
}

func (t *clockTimer) Snapshot() TimerSnapshot {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	return TimerSnapshot{ID: t.id, Pending: t.pending, Running: t.running, Deadline: t.deadline, LastErr: t.lastErr}
}

func (t *clockTimer) Stop() bool {
	c := t.clock
	c.mu.Lock()
	defer c.mu.Unlock()
	return t.stopLocked()
}

func (t *clockTimer) stopLocked() bool {
	if !t.pending {
		return false
	}
	c := t.clock
	t.pending = false
	c.s.Pending--
	c.s.Stopped++
	if t.native != nil {
		t.native.Stop()
		t.native = nil
	}
	t.retireLocked()
	return true
}

func (t *clockTimer) retireLocked() {
	if t.pending || t.running != 0 {
		return
	}
	delete(t.clock.timers, t)
	if t.stopContext != nil {
		t.stopContext()
		t.stopContext = nil
	}
}

func (t *clockTimer) claimLocked(sequence uint64) bool {
	if !t.pending || t.sequence != sequence {
		return false
	}
	if t.clock.s.Closed || t.ctx.Err() != nil {
		t.stopLocked()
		return false
	}
	t.pending = false
	t.native = nil
	t.running++
	t.clock.s.Pending--
	t.clock.s.Running++
	t.clock.s.Fired++
	return true
}

func (t *clockTimer) run(done chan<- error) {
	returned := false
	defer func() {
		var err error
		if p := recover(); p != nil {
			err = &PanicError{Value: p}
		} else if !returned {
			err = ErrGoexit
		}
		c := t.clock
		c.mu.Lock()
		t.running--
		t.lastErr = err
		c.s.Running--
		c.s.Completed++
		if err != nil {
			c.s.Failed++
		}
		t.retireLocked()
		if c.detached {
			// A canceled waiter has returned. Keep advance admission until
			// its running occurrence actually retires; do not start more work.
			c.detached = false
			c.advancing = false
		}
		c.mu.Unlock()
		if done != nil {
			done <- err
		}
	}()
	t.fn()
	returned = true
}

func (c *clockBase) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.s.Closed = true
	for t := range c.timers {
		t.stopLocked()
	}
}

// RealClock uses Go's real timer scheduler. The positive capacity bounds pending
// plus running occurrences. A Reset during a callback may run another occurrence
// concurrently, as with time.AfterFunc, but cannot exceed that bound.
type RealClock struct{ base clockBase }

var _ Clock = (*RealClock)(nil)

func NewRealClock(capacity int) (*RealClock, error) {
	if capacity < 1 {
		return nil, ErrInvalidSize
	}
	return &RealClock{base: clockBase{capacity: capacity, real: true, timers: make(map[*clockTimer]struct{})}}, nil
}

func (c *RealClock) Now() time.Time { return c.base.now() }
func (c *RealClock) AfterFunc(ctx context.Context, d time.Duration, fn func()) (Timer, error) {
	return c.base.after(ctx, d, fn)
}
func (c *RealClock) Snapshot() ClockSnapshot { return c.base.snapshot() }
func (c *RealClock) Close()                  { c.base.close() }

type TimerFailure struct {
	ID  ID
	Err error
}

type AdvanceResult struct {
	Target, Now time.Time
	Fired       int
	Failures    []TimerFailure
}

// ManualClock moves only when Advance is explicitly called. A positive callback
// budget bounds callbacks (and sequential isolated goroutine starts) per Advance,
// including callbacks that continuously reset themselves to the current time.
// At most one callback is running through a manual clock at any instant.
type ManualClock struct {
	base     clockBase
	maxFired int
}

var _ Clock = (*ManualClock)(nil)

func NewManualClock(start time.Time, capacity, maxCallbacks int) (*ManualClock, error) {
	if capacity < 1 || maxCallbacks < 1 {
		return nil, ErrInvalidSize
	}
	return &ManualClock{base: clockBase{capacity: capacity, anchor: start, timers: make(map[*clockTimer]struct{})}, maxFired: maxCallbacks}, nil
}

func (c *ManualClock) Now() time.Time { return c.base.now() }
func (c *ManualClock) AfterFunc(ctx context.Context, d time.Duration, fn func()) (Timer, error) {
	return c.base.after(ctx, d, fn)
}
func (c *ManualClock) Snapshot() ClockSnapshot { return c.base.snapshot() }
func (c *ManualClock) Close()                  { c.base.close() }

// Advance uses a noncancelable wait and is suitable only for terminating
// callbacks. Use AdvanceContext for a harness with a bounded/cancelable wait.
func (c *ManualClock) Advance(delay time.Duration) (AdvanceResult, error) {
	return c.AdvanceContext(context.Background(), delay)
}

// AdvanceContext executes due occurrences by (deadline, latest-arm sequence). A
// callback observes its actual scheduled instant and may schedule/reset nested
// work due before the target. The final Now reaches Target only after all such
// work completes. Callback failures are collected and do not silently discard
// later due timers. A budget error leaves Now at the last reached instant and
// retains every remaining pending occurrence; the caller decides the next step.
// Canceling the wait returns without running later due timers or moving past
// the reached instant. An already running callback retains its slot and the
// advance admission until it exits; later Advance calls refuse meanwhile.
// Advance is never reentrant: nested/concurrent calls return ErrAdvancing.
func (c *ManualClock) AdvanceContext(ctx context.Context, delay time.Duration) (result AdvanceResult, err error) {
	if ctx == nil {
		return result, ErrContext
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	b := &c.base
	b.mu.Lock()
	if delay < 0 {
		b.mu.Unlock()
		return result, ErrDuration
	}
	if b.s.Closed {
		b.mu.Unlock()
		return result, ErrClosed
	}
	if b.advancing {
		b.mu.Unlock()
		return result, ErrAdvancing
	}
	if int64(delay) > math.MaxInt64-b.position {
		b.mu.Unlock()
		return result, ErrClockExhausted
	}
	target := b.position + int64(delay)
	result.Target = b.anchor.Add(time.Duration(target))
	b.advancing = true
	b.mu.Unlock()
	detached := false
	defer func() {
		if detached {
			return // Cancellation captured this operation's reached instant.
		}
		b.mu.Lock()
		b.advancing = false
		result.Now = b.nowLocked()
		b.mu.Unlock()
	}()
	for {
		b.mu.Lock()
		if canceled := ctx.Err(); canceled != nil {
			b.mu.Unlock()
			return result, errors.Join(err, canceled)
		}
		if b.s.Closed {
			b.mu.Unlock()
			return result, errors.Join(err, ErrClosed)
		}
		var next *clockTimer
		for t := range b.timers {
			if !t.pending || t.position > target {
				continue
			}
			if t.ctx.Err() != nil {
				t.stopLocked()
				continue
			}
			if next == nil || t.position < next.position || t.position == next.position && t.sequence < next.sequence {
				next = t
			}
		}
		if next == nil {
			b.position = target
			b.mu.Unlock()
			return result, err
		}
		if result.Fired == c.maxFired {
			b.mu.Unlock()
			return result, errors.Join(err, ErrAdvanceBudget)
		}
		b.position = next.position
		claimed := next.claimLocked(next.sequence)
		b.mu.Unlock()
		if !claimed {
			continue
		}
		result.Fired++
		done := make(chan error, 1)
		go next.run(done)
		select {
		case callbackErr := <-done:
			if callbackErr != nil {
				result.Failures = append(result.Failures, TimerFailure{ID: next.id, Err: callbackErr})
				err = errors.Join(err, callbackErr)
			}
		case <-ctx.Done():
			b.mu.Lock()
			result.Now = b.nowLocked()
			if next.running != 0 {
				b.detached = true
			} else {
				b.advancing = false
				if next.lastErr != nil {
					result.Failures = append(result.Failures, TimerFailure{ID: next.id, Err: next.lastErr})
					err = errors.Join(err, next.lastErr)
				}
			}
			b.mu.Unlock()
			detached = true
			return result, errors.Join(err, ctx.Err())
		}
	}
}
