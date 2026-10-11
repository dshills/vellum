// Package app owns application lifetime and serializes UI mutations.
package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/ui"
	"math"
	"strings"
	"sync"
	"unicode/utf8"
)

var (
	ErrClosed       = errors.New("vellum: application closed")
	ErrQueueFull    = errors.New("vellum: UI queue full")
	ErrWindowClosed = errors.New("vellum: window closed")
	ErrCloseRefused = errors.New("vellum: window close refused")
	ErrClosePending = errors.New("vellum: window close callback is already executing")
)

type Config struct{ QueueCapacity, CompletionCapacity int }
type Window struct {
	ID, Title     string
	Width, Height float32
	View          func() ui.Node
	// CloseRequest runs on the UI owner for user/native close requests. Return
	// false to keep the window open while showing an asynchronous confirmation.
	// Once confirmed, CloseWindow can force teardown. Never block or nest Do.
	CloseRequest func() bool
}

// Driver.Run must execute on the native main thread. It installs a wakeup,
// drains posted work on its event owner, and reconciles Views after invalidation.
type Driver interface {
	Run(context.Context, *Application) error
}
type Application struct {
	changed                                    chan struct{}
	owner                                      sync.Mutex
	mu                                         sync.Mutex
	ctx                                        context.Context
	cancel                                     context.CancelFunc
	queue                                      []envelope
	capacity                                   int
	completionCapacity                         int
	normalPending, dispatching, cleanupPending int
	nextArrival, nextReservation               uint64
	unpublished                                int
	completions                                map[uint64]*Completion
	windows                                    []Window
	lifetimes                                  map[string]context.Context
	windowCancels                              map[string]context.CancelFunc
	closeRequests                              map[string]bool
	services                                   Services
	dirty                                      map[string]bool
	wake                                       func()
	running, closed                            bool
}

func New(c Config) (*Application, error) {
	if c.QueueCapacity == 0 {
		c.QueueCapacity = 256
	}
	if c.QueueCapacity < 1 {
		return nil, fmt.Errorf("vellum: queue capacity must be positive")
	}
	if c.CompletionCapacity == 0 {
		c.CompletionCapacity = 256
	}
	if c.CompletionCapacity < 1 {
		return nil, fmt.Errorf("vellum: completion capacity must be positive")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{ctx: ctx, cancel: cancel, capacity: c.QueueCapacity, completionCapacity: c.CompletionCapacity, completions: map[uint64]*Completion{}, dirty: map[string]bool{}, lifetimes: map[string]context.Context{}, windowCancels: map[string]context.CancelFunc{}, closeRequests: map[string]bool{}, changed: make(chan struct{})}, nil
}
func (a *Application) Context() context.Context { return a.ctx }

// Open registers a window before or during Run. IDs are unique for the entire
// application lifetime: a closed ID cannot be reopened. View runs on the owner.
func (a *Application) Open(w Window) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}
	if w.ID == "" || w.Title == "" || !utf8.ValidString(w.Title) || strings.ContainsRune(w.Title, 0) || w.View == nil || w.Width <= 0 || w.Height <= 0 || math.IsNaN(float64(w.Width)) || math.IsNaN(float64(w.Height)) || math.IsInf(float64(w.Width), 0) || math.IsInf(float64(w.Height), 0) {
		a.mu.Unlock()
		return fmt.Errorf("vellum: invalid window")
	}
	if _, exists := a.lifetimes[w.ID]; exists {
		a.mu.Unlock()
		return fmt.Errorf("vellum: duplicate window %q", w.ID)
	}
	lifetime, cancel := context.WithCancel(a.ctx)
	a.lifetimes[w.ID], a.windowCancels[w.ID] = lifetime, cancel
	a.windows = append(a.windows, w)
	a.dirty[w.ID] = true
	wake := a.wake
	a.mu.Unlock()
	if wake != nil {
		wake()
	}
	return nil
}

// WindowContext scopes work to a particular window. Its cancellation does not
// cancel other windows; closing the last registered window closes the app.
func (a *Application) WindowContext(id string) (context.Context, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ctx, ok := a.lifetimes[id]
	if !ok {
		return nil, fmt.Errorf("vellum: unknown window %q", id)
	}
	return ctx, nil
}

// RequestCloseWindow invokes the window's close policy and closes on acceptance.
// Call only on the UI owner (Do/Drain/widget callbacks), never inside nested Do.
// Refusal preserves registration and lifetime; CloseWindow and Close explicitly
// force cancellation/teardown without prompting. Recursive requests refuse.
func (a *Application) RequestCloseWindow(id string) error {
	if err := a.CheckCloseWindow(id); err != nil {
		return err
	}
	return a.CloseWindow(id)
}

// CheckCloseWindow runs the same owner-held policy without committing teardown.
// Drivers use it before native close events; registration changes only when native
// close is accepted. As with RequestCloseWindow, recursive policy calls refuse.
func (a *Application) CheckCloseWindow(id string) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}
	var callback func() bool
	found := false
	for _, window := range a.windows {
		if window.ID == id {
			callback = window.CloseRequest
			found = true
			break
		}
	}
	if !found {
		a.mu.Unlock()
		return fmt.Errorf("vellum: unknown window %q", id)
	}
	if a.closeRequests[id] {
		a.mu.Unlock()
		return ErrClosePending
	}
	a.closeRequests[id] = true
	a.mu.Unlock()
	defer func() { a.mu.Lock(); delete(a.closeRequests, id); a.mu.Unlock() }()
	accepted := callback == nil || callback()
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return ErrClosed
	}
	if a.lifetimes[id].Err() != nil {
		return ErrWindowClosed
	}
	if !accepted {
		return ErrCloseRefused
	}
	return nil
}

// CloseWindow removes a registered window and requests native teardown. Closing
// the last window ends the app. It is safe from workers and native callbacks.
func (a *Application) CloseWindow(id string) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}
	index := -1
	for i, w := range a.windows {
		if w.ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		a.mu.Unlock()
		return fmt.Errorf("vellum: unknown window %q", id)
	}
	a.windowCancels[id]()
	a.windows = append(a.windows[:index], a.windows[index+1:]...)
	delete(a.dirty, id)
	wake := a.wake
	hooks := a.revokeLocked(id, ErrWindowClosed)
	if len(a.windows) == 0 {
		wake = a.closeLocked()
		hooks = append(hooks, a.revokeLocked("", ErrClosed)...)
	}
	a.mu.Unlock()
	defer func() {
		if wake != nil {
			wake()
		}
	}()
	runCompletionHooks(hooks)
	return nil
}

// PostWindow discards delivery if its window closes before dispatch. It shares
// Post's bounded queue and does not transfer mutable data ownership implicitly.
func (a *Application) PostWindow(id string, fn func()) error {
	if fn == nil {
		return fmt.Errorf("vellum: nil UI callback")
	}
	ctx, err := a.WindowContext(id)
	if err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ErrWindowClosed
	}
	return a.Post(func() {
		if ctx.Err() == nil {
			fn()
		}
	})
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
	returned := true // Validation refusal never entered the callback.
	defer func() {
		p := recover()
		a.owner.Unlock()
		if !returned {
			a.Close()
		}
		if p != nil {
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
	returned = false
	fn()
	returned = true
	return nil
}

// Post admits bounded worker-to-UI delivery. Callers handle ErrQueueFull;
// posting does not transfer arbitrary mutable data safely by itself.
func (a *Application) Post(fn func()) error {
	if fn == nil {
		return fmt.Errorf("vellum: nil UI callback")
	}
	a.mu.Lock()
	err := a.postLocked(fn)
	wake := a.wake
	a.mu.Unlock()
	if err == nil && wake != nil {
		wake()
	}
	return err
}

// Drain consumes the initial arrival batch. Newly posted callbacks never enter
// that batch, even if a callback removes another initial completion.
func (a *Application) Drain() error {
	_, err := a.DrainN(math.MaxInt)
	return err
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
func (a *Application) closeLocked() func() {
	if a.closed {
		return nil
	}
	wake := a.wake
	a.closed = true
	a.queue = nil
	a.normalPending = 0
	close(a.changed)
	a.changed = make(chan struct{})
	a.wake = nil
	a.cancel()
	return wake
}
func (a *Application) Close() {
	a.mu.Lock()
	wake := a.closeLocked()
	hooks := a.revokeLocked("", ErrClosed)
	a.mu.Unlock()
	defer func() {
		if wake != nil {
			wake()
		}
	}()
	runCompletionHooks(hooks)
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
