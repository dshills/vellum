// Package execution admits callbacks for asynchronous or explicitly stepped
// execution. It deliberately knows nothing about UI owners or application clocks.
package execution

import (
	"context"
	"errors"
	"math"
	"sync"
)

var (
	ErrClosed      = errors.New("execution: executor closed")
	ErrCapacity    = errors.New("execution: outstanding job limit reached")
	ErrInvalidJob  = errors.New("execution: nonnil callback required")
	ErrInvalidSize = errors.New("execution: positive capacity required")
	ErrContext     = errors.New("execution: nonnil context required")
	ErrNotPending  = errors.New("execution: job is not pending")
	ErrRunning     = errors.New("execution: job already running")
	ErrGoexit      = errors.New("execution: callback exited without returning")
	ErrIDExhausted = errors.New("execution: job ID exhausted")
)

// Executor owns scheduling only. Refusal never executes the callback; success
// admits it exactly once. Submit never executes inline or waits for execution,
// so it is safe to call from an owner callback. Once Submit succeeds, it is retained
// until it runs, even when a context captured by the callback is canceled. The
// callback remains responsible for its own cancellation and terminal cleanup.
type Executor interface {
	Submit(func()) error
}

// ID is a monotonic, executor-local job identity. Zero is never admitted.
type ID uint64

// PanicError preserves the panic payload without invoking its formatting methods.
// As with recover, panic(nil) under legacy GODEBUG=panicnil=1 cannot be separated
// from Goexit; both remain observable terminal failures.
type PanicError struct{ Value any }

func (*PanicError) Error() string { return "execution: callback panicked" }

type State uint8

const (
	Pending State = iota
	Running
	Succeeded
	Failed
)

// Result is terminal only when Job.Result's boolean is true.
type Result struct {
	ID    ID
	State State
	Err   error
}

// Job permits observing eventual completion after a Step wait is canceled.
// Its terminal result is retained by the handle, not in an executor history.
type Job struct {
	id    ID
	mu    sync.Mutex
	state State
	err   error
	fn    func()
	done  chan struct{}
}

func (j *Job) ID() ID                { return j.id }
func (j *Job) Done() <-chan struct{} { return j.done }

func (j *Job) State() State {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.state
}

func (j *Job) Result() (Result, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	r := Result{ID: j.id, State: j.state, Err: j.err}
	return r, j.state == Succeeded || j.state == Failed
}

// Wait returns the callback's terminal failure or a canceled wait. Canceling
// this wait never stops an accepted or already running callback. If completion
// was already published when Wait starts, its terminal result wins cancellation.
func (j *Job) Wait(ctx context.Context) error {
	if ctx == nil {
		return ErrContext
	}
	select {
	case <-j.done:
		return j.terminalError()
	default:
	}
	select {
	case <-j.done:
		return j.terminalError()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (j *Job) terminalError() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.err
}

// Snapshot accounts for every accepted job. A canceled waiter or Close cannot
// make running work disappear from this accounting.
type Snapshot struct {
	Pending, Running            int
	Accepted, Completed, Failed uint64
	Refused                     uint64
	Closed                      bool
}

func (s Snapshot) Idle() bool { return s.Pending == 0 && s.Running == 0 }

type ledger struct {
	mu   sync.Mutex
	next ID
	s    Snapshot
}

// admit is called with the ledger locked. capacity zero means no executor bound.
func (l *ledger) admit(fn func(), capacity int) (*Job, error) {
	refuse := func(err error) (*Job, error) {
		l.s.Refused++
		return nil, err
	}
	if fn == nil {
		return refuse(ErrInvalidJob)
	}
	if l.s.Closed {
		return refuse(ErrClosed)
	}
	if capacity > 0 && l.s.Pending+l.s.Running >= capacity {
		return refuse(ErrCapacity)
	}
	if l.next == ID(math.MaxUint64) {
		return refuse(ErrIDExhausted)
	}
	l.next++
	l.s.Accepted++
	return &Job{id: l.next, state: Pending, fn: fn, done: make(chan struct{})}, nil
}

// run isolates runtime.Goexit and panic from its caller. The completion defer
// executes even when fn calls Goexit, and releases executor accounting before
// publishing Done. User callback defers run before this terminal publication.
func (l *ledger) run(j *Job, retire func()) {
	returned := false
	defer func() {
		var err error
		if p := recover(); p != nil {
			err = &PanicError{Value: p}
		} else if !returned {
			err = ErrGoexit
		}
		l.mu.Lock()
		j.mu.Lock()
		j.fn = nil
		j.err = err
		j.state = Succeeded
		if err != nil {
			j.state = Failed
			l.s.Failed++
		}
		l.s.Running--
		l.s.Completed++
		if retire != nil {
			retire()
		}
		close(j.done)
		j.mu.Unlock()
		l.mu.Unlock()
	}()
	j.fn()
	returned = true
}

// Goroutine is the default executor. Its zero value is ready for use. It does
// not bound goroutine count; its application caller must bound admissions when
// required. Enqueue exposes a result handle; Submit intentionally only admits.
// Do not copy an executor after first use.
type Goroutine struct{ ledger ledger }

var _ Executor = (*Goroutine)(nil)

func (g *Goroutine) Submit(fn func()) error {
	_, err := g.Enqueue(fn)
	return err
}

func (g *Goroutine) Enqueue(fn func()) (*Job, error) {
	g.ledger.mu.Lock()
	j, err := g.ledger.admit(fn, 0)
	if err != nil {
		g.ledger.mu.Unlock()
		return nil, err
	}
	j.state = Running
	g.ledger.s.Running++
	g.ledger.mu.Unlock()
	go g.ledger.run(j, nil)
	return j, nil
}

func (g *Goroutine) Snapshot() Snapshot {
	g.ledger.mu.Lock()
	defer g.ledger.mu.Unlock()
	return g.ledger.s
}

// Close refuses future submissions. It neither cancels nor waits for admitted
// jobs. Snapshot remains non-idle until all running callbacks terminate.
func (g *Goroutine) Close() {
	g.ledger.mu.Lock()
	g.ledger.s.Closed = true
	g.ledger.mu.Unlock()
}

// Manual retains at most capacity outstanding jobs, including running ones.
// Accepted callbacks execute only when selected by Step or StepNext. Close does
// not discard them; a harness must continue stepping if it wants to drain them.
// Do not copy an executor after first use.
type Manual struct {
	ledger   ledger
	capacity int
	queue    []*Job
	jobs     map[ID]*Job
}

var _ Executor = (*Manual)(nil)

func NewManual(capacity int) (*Manual, error) {
	if capacity < 1 {
		return nil, ErrInvalidSize
	}
	return &Manual{capacity: capacity, jobs: make(map[ID]*Job)}, nil
}

func (m *Manual) Submit(fn func()) error {
	_, err := m.Enqueue(fn)
	return err
}

func (m *Manual) Enqueue(fn func()) (*Job, error) {
	m.ledger.mu.Lock()
	defer m.ledger.mu.Unlock()
	if m.capacity < 1 || m.jobs == nil {
		m.ledger.s.Refused++
		return nil, ErrInvalidSize
	}
	j, err := m.ledger.admit(fn, m.capacity)
	if err != nil {
		return nil, err
	}
	m.jobs[j.id] = j
	m.queue = append(m.queue, j)
	m.ledger.s.Pending++
	return j, nil
}

func (m *Manual) Snapshot() Snapshot {
	m.ledger.mu.Lock()
	defer m.ledger.mu.Unlock()
	return m.ledger.s
}

// PendingIDs reports admission order for jobs that have not started. It returns
// a fresh slice, never the queue's storage. Running jobs are reported separately.
func (m *Manual) PendingIDs() []ID {
	m.ledger.mu.Lock()
	defer m.ledger.mu.Unlock()
	ids := make([]ID, len(m.queue))
	for i, j := range m.queue {
		ids[i] = j.id
	}
	return ids
}

func (m *Manual) Close() {
	m.ledger.mu.Lock()
	m.ledger.s.Closed = true
	m.ledger.mu.Unlock()
}

// Step selects a pending job, starts it in an isolated goroutine, and waits for
// its terminal result. Wait cancellation does not cancel the callback. The
// Enqueue handle can still observe its completion. Concurrent Steps may run
// different jobs concurrently; the same ID can be started only once.
func (m *Manual) Step(ctx context.Context, id ID) error {
	_, err := m.step(ctx, id, false)
	return err
}

// StepNext selects the earliest pending job, without skipping it for any other
// semantic reason. It returns that ID even if the wait is canceled after start.
func (m *Manual) StepNext(ctx context.Context) (ID, error) {
	return m.step(ctx, 0, true)
}

func (m *Manual) step(ctx context.Context, id ID, next bool) (ID, error) {
	if ctx == nil {
		return 0, ErrContext
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	m.ledger.mu.Lock()
	if next {
		if len(m.queue) == 0 {
			m.ledger.mu.Unlock()
			return 0, ErrNotPending
		}
		id = m.queue[0].id
	}
	j := m.jobs[id]
	if j == nil {
		m.ledger.mu.Unlock()
		return id, ErrNotPending
	}
	j.mu.Lock()
	if j.state == Running {
		j.mu.Unlock()
		m.ledger.mu.Unlock()
		return id, ErrRunning
	}
	j.state = Running
	j.mu.Unlock()
	for i, pending := range m.queue {
		if pending == j {
			copy(m.queue[i:], m.queue[i+1:])
			m.queue[len(m.queue)-1] = nil
			m.queue = m.queue[:len(m.queue)-1]
			break
		}
	}
	m.ledger.s.Pending--
	m.ledger.s.Running++
	m.ledger.mu.Unlock()
	go m.ledger.run(j, func() { delete(m.jobs, j.id) })
	return id, j.Wait(ctx)
}
