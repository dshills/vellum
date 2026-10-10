// Package app owns application lifetime and serializes UI mutations.
package app

import (
	"context"
	"errors"
	"example.com/vellum/ui"
	"fmt"
	"sync"
)

var (
	ErrClosed    = errors.New("vellum: application closed")
	ErrQueueFull = errors.New("vellum: UI queue full")
)

type Config struct{ QueueCapacity int }
type Window struct {
	ID, Title     string
	Width, Height float32
	View          func() ui.Node
}

// Driver.Run must execute on the native main thread. It installs a wakeup,
// drains posted work on its event owner, and reconciles Views after invalidation.
type Driver interface {
	Run(context.Context, *Application) error
}
type Application struct {
	changed         chan struct{}
	owner           sync.Mutex
	mu              sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
	queue           []func()
	capacity        int
	windows         []Window
	dirty           map[string]bool
	wake            func()
	running, closed bool
}

func New(c Config) (*Application, error) {
	if c.QueueCapacity == 0 {
		c.QueueCapacity = 256
	}
	if c.QueueCapacity < 1 {
		return nil, fmt.Errorf("vellum: queue capacity must be positive")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{ctx: ctx, cancel: cancel, capacity: c.QueueCapacity, dirty: map[string]bool{}, changed: make(chan struct{})}, nil
}
func (a *Application) Context() context.Context { return a.ctx }

// Open registers a window before Run. Dynamic native window creation is deferred.
func (a *Application) Open(w Window) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return ErrClosed
	}
	if a.running {
		return fmt.Errorf("vellum: register windows before Run")
	}
	if w.ID == "" || w.Title == "" || w.View == nil || w.Width <= 0 || w.Height <= 0 {
		return fmt.Errorf("vellum: invalid window")
	}
	for _, old := range a.windows {
		if old.ID == w.ID {
			return fmt.Errorf("vellum: duplicate window %q", w.ID)
		}
	}
	a.windows = append(a.windows, w)
	a.dirty[w.ID] = true
	return nil
}
func (a *Application) Windows() []Window {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]Window(nil), a.windows...)
}

// Do serializes a native callback or headless action. Do/Drain must not be nested.
// Application model state is accessed only inside Do, Drain, or View execution.
func (a *Application) Do(fn func()) error {
	a.owner.Lock()
	defer func() {
		a.owner.Unlock()
		if p := recover(); p != nil {
			a.Close()
			panic(p)
		}
	}()
	a.mu.Lock()
	closed := a.closed
	a.mu.Unlock()
	if closed {
		return ErrClosed
	}
	if fn == nil {
		return fmt.Errorf("vellum: nil UI callback")
	}
	fn()
	return nil
}

// Post admits bounded worker-to-UI delivery. Callers handle ErrQueueFull;
// posting does not transfer arbitrary mutable data safely by itself.
func (a *Application) Post(fn func()) error {
	if fn == nil {
		return fmt.Errorf("vellum: nil UI callback")
	}
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}
	if len(a.queue) >= a.capacity {
		a.mu.Unlock()
		return ErrQueueFull
	}
	a.queue = append(a.queue, fn)
	wake := a.wake
	a.mu.Unlock()
	if wake != nil {
		wake()
	}
	return nil
}

// A panic in a UI callback closes the application and propagates; recovery
// does not leave an open application with silently lost accepted work.
// Drain consumes only the current batch, so self-posting work cannot starve input.
func (a *Application) Drain() error {
	return a.Do(func() {
		a.mu.Lock()
		count := len(a.queue)
		a.mu.Unlock()
		for i := 0; i < count; i++ {
			a.mu.Lock()
			if a.closed || len(a.queue) == 0 {
				a.mu.Unlock()
				break
			}
			fn := a.queue[0]
			a.queue[0] = nil
			a.queue = a.queue[1:]
			close(a.changed)
			a.changed = make(chan struct{})
			a.mu.Unlock()
			fn()
		}
	})
}
func (a *Application) Invalidate(id string) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}
	matched := false
	for _, w := range a.windows {
		if id == "" || w.ID == id {
			a.dirty[w.ID] = true
			matched = true
		}
	}
	wake := a.wake
	a.mu.Unlock()
	if !matched {
		return fmt.Errorf("vellum: unknown window %q", id)
	}
	if wake != nil {
		wake()
	}
	return nil
}

// TakeDirty is called by a driver while holding Do's owner.
func (a *Application) TakeDirty(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	d := a.dirty[id]
	delete(a.dirty, id)
	return d
}

// SetWake installs a nonblocking notification; it must not enter Do or Drain.
func (a *Application) SetWake(fn func()) {
	a.mu.Lock()
	a.wake = fn
	pending := len(a.queue) > 0 || len(a.dirty) > 0
	closed := a.closed
	a.mu.Unlock()
	if fn != nil && pending && !closed {
		fn()
	}
}

// Close is idempotent and cancels lifetime-scoped workers. Already-running work
// cooperates with context cancellation; admitted but undispatched callbacks drop.
func (a *Application) Close() {
	a.mu.Lock()
	wake := a.wake
	if !a.closed {
		a.closed = true
		a.queue = nil
		close(a.changed)
		a.changed = make(chan struct{})
		a.wake = nil
		a.cancel()
	} else {
		wake = nil
	}
	a.mu.Unlock()
	if wake != nil {
		wake()
	}
}
func (a *Application) Run(ctx context.Context, d Driver) error {
	if ctx == nil || d == nil {
		return fmt.Errorf("vellum: context and driver required")
	}
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}
	if a.running || len(a.windows) == 0 {
		a.mu.Unlock()
		return fmt.Errorf("vellum: Run requires registered windows and runs once")
	}
	a.running = true
	a.mu.Unlock()
	stop := context.AfterFunc(ctx, a.Close)
	defer stop()
	defer a.Close()
	return d.Run(a.ctx, a)
}
