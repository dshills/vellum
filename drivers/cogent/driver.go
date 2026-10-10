// Package cogent implements Vellum's native driver using pinned Cogent Core.
// It is experimental; one process main loop owns multiple desktop windows.
package cogent

import (
	"context"
	"fmt"
	"html"
	"reflect"
	"sync"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"example.com/vellum/app"
	"example.com/vellum/command"
	"example.com/vellum/ui"
)

// Extension is implemented by optional adapter widgets (workbench, canvas, etc.).
// Sync runs under the UI owner and preserves the native widget's resources.
type Extension interface {
	Mount(tree.Node) core.Widget
	Sync(core.Widget) error
}

// ThemeExtension receives a private application theme snapshot for managed subtrees.
type ThemeExtension interface{ SetTheme(ui.Theme) }

type Driver struct {
	Commands *command.Registry
	OnError  func(error)
	Theme    ui.Theme
	// ContinuousRendering opts into refresh-rate paint ticks; default uses invalidation.
	ContinuousRendering bool
}
type record struct {
	widget   core.Widget
	node     ui.Node
	children []*record
	theme    ui.Theme
	control  *controlState
}

var processRun atomic.Bool

type nativeWindow struct {
	window           app.Window
	body             *core.Body
	host             *core.Frame
	root             *record
	rootMu           sync.Mutex
	lastFocus        ui.FocusRequest
	stopLoss         func()
	lossMu           sync.Mutex
	source           atomic.Pointer[events.Source]
	pending, closing atomic.Bool
}

func (d Driver) Run(ctx context.Context, a *app.Application) error {
	if err := d.Theme.Validate(); err != nil {
		return err
	}
	d.Theme = d.Theme.Clone()
	if len(a.Windows()) == 0 {
		return fmt.Errorf("cogent: at least one registered window required")
	}
	if !processRun.CompareAndSwap(false, true) {
		return fmt.Errorf("cogent: native main loop runs once per process")
	}
	var started, shown atomic.Bool
	defer func() {
		if !started.Load() {
			processRun.Store(false)
		}
	}()
	var failure atomic.Pointer[error]
	var mu sync.Mutex
	windows := map[string]*nativeWindow{}
	snapshot := func() []*nativeWindow {
		mu.Lock()
		defer mu.Unlock()
		out := make([]*nativeWindow, 0, len(windows))
		for _, w := range windows {
			out = append(out, w)
		}
		return out
	}
	services := &platformServices{owner: a, commands: d.Commands, onError: d.OnError}
	var quitting atomic.Bool
	quit := func() {
		if !shown.Load() {
			return
		}
		if quitting.CompareAndSwap(false, true) {
			go func() { services.shutdown(); core.TheApp.QuitReq() }()
		}
	}
	defer func() {
		if started.Load() {
			services.shutdown()
		}
	}()
	fail := func(err error) {
		if failure.CompareAndSwap(nil, &err) {
			a.Close()
			quit()
		}
	}
	wake := func() {
		for _, w := range snapshot() {
			if s := w.source.Load(); s != nil && !w.closing.Load() && w.pending.CompareAndSwap(false, true) {
				s.Custom("vellum.wake")
			}
		}
	}
	var custom func(*nativeWindow, events.Event)
	refresh := func(w *nativeWindow) {
		if w.closing.Load() {
			return
		}
		a.Do(func() {
			w.rootMu.Lock()
			defer w.rootMu.Unlock()
			if w.closing.Load() {
				return
			}
			if !a.TakeDirty(w.window.ID) {
				return
			}
			n := w.window.View()
			if err := ui.Validate(n); err != nil {
				fail(err)
				return
			}
			var err error
			w.root, err = reconcile(w.host, w.root, n, a, w.window.ID, d.Theme)
			if err != nil {
				fail(err)
				return
			}
			if w.root == nil {
				fail(fmt.Errorf("cogent: reconciliation returned no root"))
				return
			}
			bindControlWake(w.root, func(e events.Event) { custom(w, e) }, d.OnError)
			flushDialogs(w.root)
			if err := applyFocus(w.root, n.Focus, &w.lastFocus); err != nil {
				fail(err)
			}

		})
	}
	var topologyMu sync.Mutex
	var syncTopology func()
	syncTopology = func() {
		// No other window's native render lock is acquired while holding the owner.
		// Native CloseReq runs asynchronously so Close events can finish independently.
		topologyMu.Lock()
		defer topologyMu.Unlock()
		var launch []*nativeWindow
		// If construction or refresh fails, discard every tree that never entered
		// a native loop. Close callbacks cannot clean up unlaunched bodies.
		defer func() {
			for _, w := range launch {
				w.closing.Store(true)
				cleanup := func() {
					w.rootMu.Lock()
					defer w.rootMu.Unlock()
					stopDeviceLoss(w)
					disposeControlTree(w.root)
					// NewBody assigns a Scene parent before adding the body to
					// its children, so Delete cannot find an unlaunched body.
					scene := w.body.Scene
					w.body.Destroy()
					scene.Destroy()
				}
				if err := a.Do(cleanup); err != nil {
					cleanup()
				}
				mu.Lock()
				delete(windows, w.window.ID)
				mu.Unlock()
			}
		}()
		desired := a.Windows()
		active := map[string]bool{}
		for _, w := range desired {
			active[w.ID] = true
		}
		for _, w := range snapshot() {
			if !active[w.window.ID] && w.closing.CompareAndSwap(false, true) {
				if w.body.Scene.RenderWindow() != nil {
					win := w.body.Scene.RenderWindow().SystemWindow
					go win.CloseReq()
				}
			}
		}
		for _, spec := range desired {
			mu.Lock()
			_, exists := windows[spec.ID]
			mu.Unlock()
			if exists || ctx.Err() != nil {
				continue
			}
			w := &nativeWindow{window: spec}
			buildErr := a.Do(func() {
				w.body = core.NewBody(spec.Title)
				w.body.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(spec.Width), units.Dp(spec.Height)); s.Grow.Set(1, 1) })
				w.host = core.NewFrame(w.body)
				w.host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Direction = styles.Column })
			})
			if buildErr != nil {
				return
			}
			mu.Lock()
			windows[spec.ID] = w
			mu.Unlock()
			launch = append(launch, w)
			w.body.OnShow(func(events.Event) {
				rw := w.body.Scene.RenderWindow()
				if rw == nil || rw.SystemWindow == nil {
					if !w.closing.Load() {
						fail(fmt.Errorf("cogent: render window unavailable on show"))
					}
					return
				}
				shown.Store(true)
				if w.closing.Load() {
					go rw.SystemWindow.CloseReq()
					return
				}
				if err := configureRendering(rw.SystemWindow, d.ContinuousRendering); err != nil {
					fail(err)
					return
				}
				watchDeviceLoss(w, fail)
				lifetime, lifetimeErr := a.WindowContext(spec.ID)
				if w.closing.Load() || lifetimeErr != nil || lifetime.Err() != nil {
					w.closing.Store(true)
					if native := w.body.Scene.RenderWindow(); native != nil {
						go native.SystemWindow.CloseReq()
					}
					return
				}
				w.source.Store(w.body.SystemEvents())
				a.Do(func() { w.rootMu.Lock(); defer w.rootMu.Unlock(); flushDialogs(w.root) })
				if ctx.Err() != nil || failure.Load() != nil {
					quit()
					return
				}
				a.SetServices(services)
				wake()
			})
			w.body.Scene.OnFirst(events.Custom, func(e events.Event) { custom(w, e) })
			if d.Commands != nil {
				w.body.Scene.OnFirst(events.KeyChord, func(e events.Event) {
					a.Do(func() {
						lifetime, err := a.WindowContext(spec.ID)
						if err != nil || lifetime.Err() != nil {
							return
						}
						matched, err := d.Commands.Shortcut(lifetime, string(e.KeyChord()))
						if matched {
							e.SetHandled()
							if err != nil && d.OnError != nil {
								d.OnError(err)
							}
							a.Invalidate("")
						}
					})
				})
			}
			w.body.OnClose(func(events.Event) {
				w.closing.Store(true)
				stopDeviceLoss(w)
				a.CloseWindow(spec.ID)
				cleanup := func() { w.rootMu.Lock(); defer w.rootMu.Unlock(); disposeControlTree(w.root) }
				if err := a.Do(cleanup); err != nil {
					cleanup()
				}
				w.source.Store(nil)
				mu.Lock()
				delete(windows, spec.ID)
				mu.Unlock()
				// A new registration may have been accepted just before the last native
				// window closes, before its queued Custom wake could create that window.
				// Launch remaining registrations before returning from Close, keeping
				// Cogent's window wait group alive without requiring a shown event source.
				if ctx.Err() == nil {
					syncTopology()
				}
			})
			refresh(w)
			if failure.Load() != nil {
				return
			}
		}
		for len(launch) > 0 {
			w := launch[0]
			launch = launch[1:]
			started.Store(true)
			w.body.NewWindow().SetUseMinSize(true).Run()
		}
	}
	custom = func(w *nativeWindow, e events.Event) {
		w.pending.Store(false)
		a.Drain()
		syncTopology()
		refresh(w)
		if ce, ok := e.(*events.CustomEvent); ok && ce.Data == "vellum.wake" {
			e.SetHandled()
		}
	}
	a.SetWake(wake)
	defer a.SetWake(nil)
	defer a.SetServices(nil)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	syncTopology()
	if err := failure.Load(); err != nil && !started.Load() {
		return *err
	}
	if ctx.Err() != nil && !started.Load() {
		return ctx.Err()
	}
	if !started.Load() {
		return fmt.Errorf("cogent: no native window started")
	}
	stop := context.AfterFunc(ctx, quit)
	defer stop()
	core.Wait()
	if err := failure.Load(); err != nil {
		return *err
	}
	return nil
}
func sameShape(r *record, n ui.Node) bool {
	if r.node.ID != n.ID || r.node.Kind != n.Kind || len(r.children) != len(n.Children) {
		return false
	}
	if n.Kind == ui.Extension {
		if reflect.TypeOf(r.node.Custom) != reflect.TypeOf(n.Custom) {
			return false
		}
		a, b := reflect.ValueOf(r.node.Custom), reflect.ValueOf(n.Custom)
		if a.IsValid() && a.Kind() == reflect.Pointer && a.Pointer() != b.Pointer() {
			return false
		}
	}
	for i, c := range n.Children {
		if !sameShape(r.children[i], c) {
			return false
		}
	}
	return true
}
func mount(parent tree.Node, n ui.Node, a *app.Application, window string, themes ...ui.Theme) (*record, error) {
	r, err := mountShallow(parent, n, a, window, themes...)
	if err != nil {
		return nil, err
	}
	for _, c := range n.Children {
		child, err := mount(childHost(r, c.ID), c, a, window, r.theme)
		if err != nil {
			disposeControlTree(r)
			r.widget.AsTree().Delete()
			return nil, err
		}
		r.children = append(r.children, child)
	}
	return r, nil
}

func mountShallow(parent tree.Node, n ui.Node, a *app.Application, window string, themes ...ui.Theme) (*record, error) {
	var theme ui.Theme
	if len(themes) > 0 {
		theme = themes[0]
	}
	if _, err := ui.ResolveStyle(n, theme); err != nil {
		return nil, err
	}
	r := &record{node: n, theme: theme}
	if ext, ok := n.Custom.(ThemeExtension); n.Kind == ui.Extension && ok {
		ext.SetTheme(theme.Clone())
	}
	switch n.Kind {
	case ui.Column, ui.Row:
		r.widget = core.NewFrame(parent)
	case ui.Label:
		r.widget = core.NewText(parent)
	case ui.Button:
		btn := core.NewButton(parent).SetText(html.EscapeString(n.Text))
		btn.SetState(n.Disabled, states.Disabled)
		r.widget = btn
		btn.OnClick(func(e events.Event) {
			a.Do(func() {
				if !controlDisabled(r) && r.node.Activate != nil {
					r.node.Activate()
					a.Invalidate(window)
				}
			})
		})
	case ui.Extension:
		if ext, ok := n.Custom.(Extension); ok {
			r.widget = ext.Mount(parent)
		} else {
			return nil, fmt.Errorf("cogent: unsupported extension %q", n.ID)
		}
	default:
		var ok bool
		var err error
		r.widget, ok, err = mountControl(parent, r, a, window)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("cogent: unsupported kind %q", n.Kind)
		}
	}
	if r.widget == nil || (reflect.ValueOf(r.widget).Kind() == reflect.Pointer && reflect.ValueOf(r.widget).IsNil()) {
		return nil, fmt.Errorf("cogent: extension %q returned nil widget", n.ID)
	}
	r.widget.AsWidget().SetName(n.ID)
	syncSemantics(r)
	r.widget.AsWidget().Styler(func(s *styles.Style) { applyStyle(s, r.node, r.theme) })
	return r, nil
}
func syncRecord(r *record, n ui.Node) error {
	if _, err := ui.ResolveStyle(n, r.theme); err != nil {
		return err
	}
	if err := ui.ValidateControl(n); err != nil {
		return err
	}
	previous := r.node
	r.node = n
	if previous.Style != n.Style || previous.Grow != n.Grow || previous.MinWidth != n.MinWidth || previous.MinHeight != n.MinHeight {
		r.widget.AsWidget().Update()
	}
	switch n.Kind {
	case ui.Label:
		if previous.Text != n.Text || r.widget.(*core.Text).Text != html.EscapeString(n.Text) {
			r.widget.(*core.Text).SetText(html.EscapeString(n.Text)).Update()
		}
	case ui.Button:
		if previous.Text != n.Text || previous.Disabled != n.Disabled {
			r.widget.(*core.Button).SetText(html.EscapeString(n.Text)).SetState(n.Disabled, states.Disabled)
			r.widget.AsWidget().Update()
		}
	case ui.Extension:
		ext, ok := n.Custom.(Extension)
		if !ok {
			return fmt.Errorf("cogent: unsupported extension %q", n.ID)
		}
		if err := ext.Sync(r.widget); err != nil {
			return err
		}
	}
	syncSemantics(r)
	if err := syncControl(r, previous); err != nil {
		return err
	}
	for i, c := range n.Children {
		if err := syncRecord(r.children[i], c); err != nil {
			return err
		}
	}
	return nil
}

func drainCustom(a *app.Application, pending *atomic.Bool, refresh func(), e events.Event) {
	pending.Store(false)
	a.Drain()
	refresh()
	if ce, ok := e.(*events.CustomEvent); ok && ce.Data == "vellum.wake" {
		e.SetHandled()
	}
}
