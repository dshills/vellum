// Package tasks runs bounded, lifetime-scoped workers with owner-thread commits.
package tasks

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

var ErrCapacity = errors.New("tasks: active task limit reached")

type State struct {
	Progress float64
	Message  string
	Running  bool
	Err      error
}
type Task struct {
	mu     sync.Mutex
	state  State
	cancel context.CancelFunc
	done   chan struct{}
}

func (t *Task) Cancel()               { t.cancel() }
func (t *Task) Done() <-chan struct{} { return t.done }
func (t *Task) Snapshot() State       { t.mu.Lock(); defer t.mu.Unlock(); return t.state }

type Manager struct {
	app        *app.Application
	slots      chan struct{}
	invalidate func()
}

func New(a *app.Application, limit int, invalidate func()) (*Manager, error) {
	if a == nil || limit < 1 {
		return nil, fmt.Errorf("tasks: application and positive limit required")
	}
	return &Manager{app: a, slots: make(chan struct{}, limit), invalidate: invalidate}, nil
}

// Start is called on the UI owner. A slot remains occupied until completion is
// delivered, bounding workers AND pending results. Work must honor cancellation.
// Result is owned by the worker until commit; commit runs once on the UI owner.
func Start[T any](m *Manager, parent context.Context, work func(context.Context, func(float64, string)) (T, error), commit func(T, error)) (*Task, error) {
	if m == nil || m.app == nil || m.slots == nil || parent == nil || work == nil || commit == nil {
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
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { <-m.slots }) }
	stop := context.AfterFunc(m.app.Context(), func() { cancel(); release() })
	t := &Task{state: State{Running: true}, cancel: cancel, done: make(chan struct{})}
	var progressPending atomic.Bool
	go func() {
		defer close(t.done)
		var result T
		var err error
		func() {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("tasks: worker panic: %v", p)
				}
			}()
			result, err = work(ctx, func(p float64, s string) {
				if math.IsNaN(p) {
					p = 0
				}
				if p < 0 {
					p = 0
				}
				if p > 1 {
					p = 1
				}
				t.mu.Lock()
				t.state.Progress = p
				t.state.Message = s
				t.mu.Unlock()
				if m.invalidate != nil && progressPending.CompareAndSwap(false, true) {
					if m.app.Post(func() { progressPending.Store(false); m.invalidate() }) != nil {
						progressPending.Store(false)
					}
				}
			})
		}()
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		t.mu.Lock()
		t.state.Running = false
		t.state.Err = err
		t.mu.Unlock()
		delivery := m.app.PostWait(m.app.Context(), func() {
			defer release()
			defer cancel()
			defer stop()
			commitErr := err
			if ctx.Err() != nil && commitErr == nil {
				commitErr = ctx.Err()
			}
			t.mu.Lock()
			t.state.Err = commitErr
			t.mu.Unlock()
			commit(result, commitErr)
			if m.invalidate != nil {
				m.invalidate()
			}
		})
		if delivery != nil {
			cancel()
			stop()
			release()
		}
	}()
	return t, nil
}
