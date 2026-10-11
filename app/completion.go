package app

import (
	"context"
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidCompletion   = errors.New("vellum: invalid or copied completion reservation")
	ErrCompletionCapacity  = errors.New("vellum: completion reservation limit reached")
	ErrCompletionUsed      = errors.New("vellum: completion reservation already used")
	ErrSequenceExhausted   = errors.New("vellum: mailbox identity or arrival sequence exhausted")
	ErrOwnerCallbackFailed = errors.New("vellum: owner callback failed without returning")
	ErrCompletionAborted   = errors.New("vellum: completion reservation aborted")
)

type envelope struct {
	arrival    uint64
	fn         func()
	completion *Completion
}

type completionState uint8

const (
	completionReserved completionState = iota
	completionReady
	completionRunning
	completionTerminal
)

// Completion is a bounded, lifetime-scoped result admission. It does not insert
// a queue placeholder; arrival order is assigned only by successful Publish.
type Completion struct {
	self     *Completion
	app      *Application
	id       uint64
	window   string
	lifetime context.Context
	state    completionState
	err      error
	done     chan struct{}
	cleanup  func(error)
}

type completionHook struct {
	ticket *Completion
	fn     func(error)
	err    error
}

// ReserveCompletion is nonblocking. cleanup is synchronous resource/accounting
// cleanup, called once outside app.mu; it must return normally without blocking,
// entering Do, or mutating UI/domain state. User owner callbacks belong in
// Publish, never in cleanup.
// Reservation credit retires at dispatch termination/revocation; Cleanup
// remains observable and prevents Pending.Idle until the hook has returned.
// An empty window scopes the reservation to the whole application.
func (a *Application) ReserveCompletion(window string, cleanup func(error)) (*Completion, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return nil, ErrClosed
	}
	lifetime := a.ctx
	if window != "" {
		var ok bool
		lifetime, ok = a.lifetimes[window]
		if !ok {
			return nil, fmt.Errorf("vellum: unknown window %q", window)
		}
		if lifetime.Err() != nil {
			return nil, ErrWindowClosed
		}
	}
	if len(a.completions) >= a.completionCapacity {
		return nil, ErrCompletionCapacity
	}
	if a.nextReservation == math.MaxUint64 || math.MaxUint64-a.nextArrival <= uint64(a.unpublished) {
		return nil, ErrSequenceExhausted
	}
	a.nextReservation++
	c := &Completion{app: a, id: a.nextReservation, window: window, lifetime: lifetime, done: make(chan struct{}), cleanup: cleanup}
	c.self = c
	a.completions[c.id] = c
	a.unpublished++
	return c, nil
}

// Publish never waits for ordinary queue space. Its sequence credit was held at
// reservation. Lifetime revocation or duplicate publication is explicit refusal.
func (c *Completion) Publish(fn func()) error {
	if c == nil || c.app == nil || c.self != c {
		return ErrInvalidCompletion
	}
	if fn == nil {
		return fmt.Errorf("vellum: completion and callback required")
	}
	a := c.app
	a.mu.Lock()
	if c.state != completionTerminal && a.completions[c.id] != c {
		a.mu.Unlock()
		return ErrInvalidCompletion
	}
	if c.state != completionReserved {
		err := c.err
		if err == nil {
			err = ErrCompletionUsed
		}
		a.mu.Unlock()
		return err
	}
	var closed error
	if a.closed {
		closed = ErrClosed
	} else if c.lifetime.Err() != nil {
		closed = ErrWindowClosed
	}
	if closed != nil {
		hook := a.terminalLocked(c, closed)
		a.mu.Unlock()
		runCompletionHooks([]completionHook{hook})
		return closed
	}
	a.unpublished--
	a.nextArrival++
	c.state = completionReady
	a.queue = append(a.queue, envelope{arrival: a.nextArrival, fn: fn, completion: c})
	wake := a.wake
	a.mu.Unlock()
	if wake != nil {
		wake()
	}
	return nil
}

func (c *Completion) Done() <-chan struct{} {
	if c == nil || c.self != c {
		return nil
	}
	return c.done
}
func (c *Completion) Err() error {
	if c == nil || c.self != c || c.app == nil {
		return ErrInvalidCompletion
	}
	c.app.mu.Lock()
	defer c.app.mu.Unlock()
	return c.err
}

// Abort withdraws a reserved/ready result; it does not interrupt dispatch.
func (c *Completion) Abort() bool {
	if c == nil || c.app == nil || c.self != c {
		return false
	}
	a := c.app
	a.mu.Lock()
	if c.state == completionRunning || c.state == completionTerminal {
		a.mu.Unlock()
		return false
	}
	a.unlinkLocked(c)
	hook := a.terminalLocked(c, ErrCompletionAborted)
	a.mu.Unlock()
	runCompletionHooks([]completionHook{hook})
	return true
}

func (a *Application) unlinkLocked(c *Completion) {
	for i, e := range a.queue {
		if e.completion == c {
			copy(a.queue[i:], a.queue[i+1:])
			a.queue[len(a.queue)-1] = envelope{}
			a.queue = a.queue[:len(a.queue)-1]
			return
		}
	}
}

func (a *Application) terminalLocked(c *Completion, err error) completionHook {
	if c.state == completionReserved {
		a.unpublished--
	}
	c.state = completionTerminal
	c.err = err
	delete(a.completions, c.id)
	a.cleanupPending++
	hook := completionHook{ticket: c, fn: c.cleanup, err: err}
	c.cleanup = nil
	return hook
}

func runCompletionHooks(hooks []completionHook) {
	next := 0
	// A cleanup's panic/Goexit must not strand unrelated already-terminal
	// reservations. Normal retirement remains iterative; only failure unwinding
	// resumes the remainder through a defer. Abnormal hooks follow ordinary Go
	// deferred-unwinding semantics: a later panic/Goexit can replace a failure.
	defer func() {
		if next < len(hooks) {
			runCompletionHooks(hooks[next:])
		}
	}()
	for next < len(hooks) {
		hook := hooks[next]
		next++
		func() {
			defer func() {
				a := hook.ticket.app
				a.mu.Lock()
				a.cleanupPending--
				close(hook.ticket.done)
				a.mu.Unlock()
			}()
			if hook.fn != nil {
				hook.fn(hook.err)
			}
		}()
	}
}

func (a *Application) revokeLocked(window string, err error) []completionHook {
	var hooks []completionHook
	for _, c := range a.completions {
		if c.state == completionRunning || window != "" && c.window != window {
			continue
		}
		a.unlinkLocked(c)
		hooks = append(hooks, a.terminalLocked(c, err))
	}
	return hooks
}

func (a *Application) postLocked(fn func()) error {
	if a.closed {
		return ErrClosed
	}
	if a.normalPending >= a.capacity {
		return ErrQueueFull
	}
	if math.MaxUint64-a.nextArrival <= uint64(a.unpublished) {
		return ErrSequenceExhausted
	}
	a.nextArrival++
	a.normalPending++
	a.queue = append(a.queue, envelope{arrival: a.nextArrival, fn: fn})
	return nil
}

type PendingState struct {
	DirtyWindows                                               int
	Ordinary, Reserved, ReadyCompletions, Dispatching, Cleanup int
	Closed                                                     bool
}

// Idle reports mailbox quiescence. DirtyWindows is separate render work and
// must also be checked by a frame/native runner before claiming UI quiescence.
func (p PendingState) Idle() bool {
	return p.Ordinary == 0 && p.Reserved == 0 && p.ReadyCompletions == 0 && p.Dispatching == 0 && p.Cleanup == 0
}

func (a *Application) Pending() PendingState {
	a.mu.Lock()
	defer a.mu.Unlock()
	p := PendingState{Ordinary: a.normalPending, Dispatching: a.dispatching, Cleanup: a.cleanupPending, Closed: a.closed, DirtyWindows: len(a.dirty)}
	for _, c := range a.completions {
		switch c.state {
		case completionReserved:
			p.Reserved++
		case completionReady:
			p.ReadyCompletions++
		}
	}
	return p
}

// DrainN runs at most max envelopes from the initial arrival batch, under owner.
// The cutoff, rather than initial count alone, prevents aborting B from allowing
// a newly posted C into the initial A/B batch. A zero budget is a checked no-op.
func (a *Application) DrainN(max int) (processed int, err error) {
	if max < 0 {
		return 0, fmt.Errorf("vellum: drain budget must not be negative")
	}
	err = a.Do(func() {
		a.mu.Lock()
		cutoff := a.nextArrival
		a.mu.Unlock()
		for processed < max {
			a.mu.Lock()
			if a.closed || len(a.queue) == 0 || a.queue[0].arrival > cutoff {
				a.mu.Unlock()
				return
			}
			e := a.queue[0]
			a.queue[0] = envelope{}
			a.queue = a.queue[1:]
			if e.completion == nil {
				a.normalPending--
			} else {
				e.completion.state = completionRunning
			}
			a.dispatching++
			close(a.changed)
			a.changed = make(chan struct{})
			a.mu.Unlock()
			processed++
			func() {
				returned := false
				defer func() {
					value := recover()
					var failure error
					if !returned {
						failure = ErrOwnerCallbackFailed
					}
					a.mu.Lock()
					a.dispatching--
					var hooks []completionHook
					if e.completion != nil {
						hooks = append(hooks, a.terminalLocked(e.completion, failure))
					}
					a.mu.Unlock()
					runCompletionHooks(hooks)
					if !returned {
						a.Close()
					}
					if value != nil {
						panic(value)
					}
				}()
				e.fn()
				returned = true
			}()
		}
	})
	return processed, err
}
