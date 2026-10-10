// Package cogent implements Vellum's native driver using pinned Cogent Core.
// It is experimental; the current adapter supports one desktop window per run.
package cogent

import (
	"context"
	"fmt"
	"html"
	"reflect"
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
}
type record struct {
	widget   core.Widget
	node     ui.Node
	children []*record
	theme    ui.Theme
}

var processRun atomic.Bool

func (d Driver) Run(ctx context.Context, a *app.Application) error {
	if err := d.Theme.Validate(); err != nil {
		return err
	}
	theme := d.Theme.Clone()
	ws := a.Windows()
	if len(ws) != 1 {
		return fmt.Errorf("cogent: one native window supported; got %d", len(ws))
	}

	if !processRun.CompareAndSwap(false, true) {
		return fmt.Errorf("cogent: native main loop runs once per process")
	}
	started := false
	defer func() {
		if !started {
			processRun.Store(false)
		}
	}()
	w := ws[0]
	b := core.NewBody(w.Title)
	b.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(w.Width), units.Dp(w.Height)); s.Grow.Set(1, 1) })
	host := core.NewFrame(b)
	host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Direction = styles.Column })
	var root *record
	var failure atomic.Pointer[error]
	var source atomic.Pointer[events.Source]
	var pending atomic.Bool
	var shown atomic.Bool
	quit := func() {
		if shown.Load() {
			go core.TheApp.QuitReq()
		}
	}
	refresh := func() {
		a.Do(func() {
			if !a.TakeDirty(w.ID) {
				return
			}
			n := w.View()
			if err := ui.Validate(n); err != nil {
				failure.CompareAndSwap(nil, &err)
				quit()
				return
			}
			var err error
			root, err = reconcile(host, root, n, a, w.ID, theme)
			if err != nil {
				failure.CompareAndSwap(nil, &err)
				quit()
			}
		})
	}
	wake := func() {
		if ctx.Err() != nil {
			return
		}
		s := source.Load()
		if s != nil && pending.CompareAndSwap(false, true) {
			s.Custom("vellum.wake")
		}
	}
	b.OnShow(func(e events.Event) {
		shown.Store(true)
		source.Store(b.SystemEvents())
		if ctx.Err() != nil || failure.Load() != nil {
			quit()
			return
		}
		a.SetWake(wake)
		wake()
	})
	// Cogent compresses consecutive Custom events by type, including events
	// from other producers. Every Custom event drains our coalesced work.
	b.Scene.OnFirst(events.Custom, func(e events.Event) { drainCustom(a, &pending, refresh, e) })
	if d.Commands != nil {
		b.Scene.OnFirst(events.KeyChord, func(e events.Event) {
			a.Do(func() {
				matched, err := d.Commands.Shortcut(ctx, string(e.KeyChord()))
				if matched {
					e.SetHandled()
					if err != nil && d.OnError != nil {
						d.OnError(err)
					}
					a.Invalidate(w.ID)
				}
			})
		})
	}
	b.OnClose(func(e events.Event) { a.Close(); source.Store(nil) })
	if ctx.Err() != nil {
		return ctx.Err()
	}
	refresh()
	if err := failure.Load(); err != nil {
		return *err
	}
	stop := context.AfterFunc(ctx, func() { quit() })
	defer stop()
	defer a.SetWake(nil)

	started = true
	b.NewWindow().SetUseMinSize(true).RunMain()
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
		child, err := mount(r.widget, c, a, window, r.theme)
		if err != nil {
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
				if !r.node.Disabled && r.node.Activate != nil {
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
		return nil, fmt.Errorf("cogent: unsupported kind %q", n.Kind)
	}
	if r.widget == nil || (reflect.ValueOf(r.widget).Kind() == reflect.Pointer && reflect.ValueOf(r.widget).IsNil()) {
		return nil, fmt.Errorf("cogent: extension %q returned nil widget", n.ID)
	}
	r.widget.AsWidget().SetName(n.ID)
	r.widget.AsWidget().Styler(func(s *styles.Style) { applyStyle(s, r.node, r.theme) })
	return r, nil
}
func syncRecord(r *record, n ui.Node) error {
	if _, err := ui.ResolveStyle(n, r.theme); err != nil {
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
