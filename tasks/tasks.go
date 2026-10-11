// Package tasks runs bounded, lifetime-scoped workers with owner-thread commits.
package tasks

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"math"
	"sync"
	"sync/atomic"
)

var ErrCapacity = errors.New("tasks: active task limit reached")

type State struct {
	Progress              float64
	Message               string
	Running               bool
	AwaitingOwner         bool
	WorkerDone, OwnerDone bool
	Err                   error
}
type Task struct {
	mu              sync.Mutex
	state           State
	cancel          context.CancelFunc
	done, ownerDone chan struct{}
	release         func()
}

func (t *Task) Cancel() { t.cancel() }

// Done retains worker+completion-admission meaning; owner commit can still be pending.
func (t *Task) Done() <-chan struct{}      { return t.done }
func (t *Task) OwnerDone() <-chan struct{} { return t.ownerDone }
func (t *Task) Snapshot() State            { t.mu.Lock(); defer t.mu.Unlock(); return t.state }
func (t *Task) workerFinished(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state.Running = false
	t.state.WorkerDone = true
	t.state.AwaitingOwner = !t.state.OwnerDone
	t.state.Err = errors.Join(t.state.Err, err)
	if t.state.OwnerDone {
		t.release()
	}
	close(t.done)
}
func (t *Task) deliveryFinished(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cancel()
	t.state.OwnerDone = true
	t.state.AwaitingOwner = false
	t.state.Err = errors.Join(t.state.Err, err)
	if t.state.WorkerDone {
		t.release()
	}
	close(t.ownerDone)
}

type Manager struct {
	app        *app.Application
	window     string
	executor   execution.Executor
	slots      chan struct{}
	invalidate func()
}
type Config struct {
	Owner      *app.Application
	Window     string
	Limit      int
	Invalidate func()
	Executor   execution.Executor
}
type PendingState struct{ Active int }

func (s PendingState) Idle() bool        { return s.Active == 0 }
func (m *Manager) Pending() PendingState { return PendingState{Active: len(m.slots)} }
func New(a *app.Application, limit int, invalidate func()) (*Manager, error) {
	return NewWithConfig(Config{Owner: a, Limit: limit, Invalidate: invalidate})
}
func NewWithConfig(c Config) (*Manager, error) {
	if c.Owner == nil || c.Limit < 1 {
		return nil, fmt.Errorf("tasks: application and positive limit required")
	}
	if c.Window != "" {
		lifetime, err := c.Owner.WindowContext(c.Window)
		if err != nil {
			return nil, err
		}
		if lifetime.Err() != nil {
			return nil, app.ErrWindowClosed
		}
	}
	if c.Executor == nil {
		c.Executor = &execution.Goroutine{}
	}
	return &Manager{app: c.Owner, window: c.Window, executor: c.Executor, slots: make(chan struct{}, c.Limit), invalidate: c.Invalidate}, nil
}

// Start is owner-only. Its slot remains occupied until worker termination AND
// reserved owner completion/revocation. Canceled accepted workers still clean up.
func Start[T any](m *Manager, parent context.Context, work func(context.Context, func(float64, string)) (T, error), commit func(T, error)) (*Task, error) {
	if m == nil || m.app == nil || m.executor == nil || m.slots == nil || parent == nil || work == nil || commit == nil {
		return nil, fmt.Errorf("tasks: manager, context, work and commit required")
	}
	if err := m.app.Context().Err(); err != nil {
		return nil, err
	}
	if err := parent.Err(); err != nil {
		return nil, err
	}
	select {
	case m.slots <- struct{}{}:
	default:
		return nil, ErrCapacity
	}
	ctx, cancel := context.WithCancel(parent)
	var once sync.Once
	task := &Task{state: State{Running: true}, cancel: cancel, done: make(chan struct{}), ownerDone: make(chan struct{}), release: func() { once.Do(func() { <-m.slots }) }}
	ticket, err := m.app.ReserveCompletion(m.window, task.deliveryFinished)
	if err != nil {
		cancel()
		task.release()
		return nil, err
	}
	var progressPending atomic.Bool
	err = m.executor.Submit(func() {
		var result T
		var workerErr error
		returned := false
		defer func() {
			if p := recover(); p != nil {
				workerErr = &execution.PanicError{Value: p}
			} else if !returned {
				workerErr = execution.ErrGoexit
			}
			workerErr = errors.Join(workerErr, ctx.Err())
			capturedErr := workerErr
			delivery := ticket.Publish(func() {
				commitErr := errors.Join(capturedErr, ctx.Err())
				task.mu.Lock()
				task.state.Err = commitErr
				task.mu.Unlock()
				commit(result, commitErr)
				if m.invalidate != nil {
					m.invalidate()
				}
			})
			task.workerFinished(errors.Join(workerErr, delivery))
		}()
		result, workerErr = work(ctx, func(p float64, s string) {
			if math.IsNaN(p) {
				p = 0
			}
			p = math.Max(0, math.Min(1, p))
			task.mu.Lock()
			task.state.Progress = p
			task.state.Message = s
			task.mu.Unlock()
			if m.invalidate != nil && progressPending.CompareAndSwap(false, true) {
				fn := func() { progressPending.Store(false); m.invalidate() }
				var err error
				if m.window == "" {
					err = m.app.Post(fn)
				} else {
					err = m.app.PostWindow(m.window, fn)
				}
				if err != nil {
					progressPending.Store(false)
				}
			}
		})
		returned = true
	})
	if err != nil {
		ticket.Abort()
		task.workerFinished(err)
		return nil, err
	}
	return task, nil
}
