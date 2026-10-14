package cogent

import (
	"encoding/base64"
	"html"
	"slices"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

type menuEpoch struct{ marker byte }
type menuState struct {
	epoch   *menuEpoch
	options []ui.Option
	popup   *menuPopup
	closing *menuPopup // owner revokes first; originating reconciliation destroys native scene
	request bool
}
type menuPopup struct {
	owner             *record
	epoch             *menuEpoch
	stage             *core.Stage
	scene             *core.Scene
	context           *core.Scene
	items             map[string]*menuItem
	dead              atomic.Bool
	geometryReady     atomic.Bool
	manualIncarnation uint64
}
type menuItem struct {
	core.Button
	popup *menuPopup
}

func (w *menuItem) Destroy() {
	p := w.popup
	w.Button.Destroy()
	if p != nil && p.dead.CompareAndSwap(false, true) {
		postControl(p.owner, func() {
			if p.owner.control.menu.popup == p {
				requestMenuClose(p.owner)
				_ = p.owner.control.a.Invalidate(p.owner.control.window)
			}
		})
	}
}
func (w *menuItem) Render() {
	w.Button.Render()
	p := w.popup
	if p != nil && !p.dead.Load() && p.geometryReady.CompareAndSwap(false, true) {
		postControl(p.owner, func() {
			if p.owner.control.menu.popup == p {
				_ = p.owner.control.a.Invalidate(p.owner.control.window)
			}
		})
	}
}
func menuAvailable(r *record) bool {
	if r.control == nil || r.control.retired || r.control.disabled || r.node.Change == nil {
		return false
	}
	for _, o := range r.node.Options {
		if !o.Disabled && !o.Separator {
			return true
		}
	}
	return false
}
func menuOpen(r *record) bool {
	return r != nil && r.control != nil && r.control.menu != nil && r.control.menu.popup != nil && !r.control.menu.popup.dead.Load()
}
func mountMenuButton(parent tree.Node, r *record) core.Widget {
	r.control.menu = &menuState{epoch: &menuEpoch{}}
	b := core.NewButton(parent)
	b.SetIndicator(nativeSymbol(ui.SymbolDown))
	b.OnClick(func(events.Event) {
		epoch := r.control.menu.epoch
		postControl(r, func() {
			if menuAvailable(r) && r.control.menu.epoch == epoch {
				r.control.menu.request = true
				_ = r.control.a.Invalidate(r.control.window)
			}
		})
	})
	b.OnFirst(events.KeyChord, func(e events.Event) {
		if !e.IsHandled() && e.Modifiers() == 0 && e.KeyCode() == key.CodeDownArrow {
			e.SetHandled()
			epoch := r.control.menu.epoch
			postControl(r, func() {
				if menuAvailable(r) && r.control.menu.epoch == epoch {
					r.control.menu.request = true
					_ = r.control.a.Invalidate(r.control.window)
				}
			})
		}
	})
	return b
}
func syncMenuButton(r *record, previous ui.Node) {
	s := r.control.menu
	flushMenuClose(r)
	changed := !slices.Equal(s.options, r.node.Options) || previous.Text != r.node.Text || previous.Icon != r.node.Icon || previous.Appearance != r.node.Appearance || previous.Style != r.node.Style || previous.Grow != r.node.Grow || previous.Disabled != r.node.Disabled || previous.MinWidth != r.node.MinWidth || previous.MinHeight != r.node.MinHeight
	if changed {
		requestMenuClose(r)
		s.epoch = &menuEpoch{}
		s.options = slices.Clone(r.node.Options)
	}
	if !menuAvailable(r) {
		requestMenuClose(r)
	}
	flushMenuClose(r)
	r.widget.AsWidget().SetState(!menuAvailable(r), states.Disabled)
	if s.request {
		s.request = false
		if menuAvailable(r) && !menuOpen(r) {
			openManagedMenu(r)
		}
	}
}
func openManagedMenu(r *record) {
	s := r.control.menu
	context := r.widget.AsWidget().Scene
	if context != nil && context.Stage != nil && context.Stage.Main != nil {
		context = context.Stage.Main.Scene
	}
	if context != nil {
		if old, ok := context.Property("vellum.menu.active").(*menuPopup); ok && old != nil {
			requestMenuClose(old.owner)
			flushMenuClose(old.owner)
		}
	}
	p := &menuPopup{owner: r, epoch: s.epoch, items: map[string]*menuItem{}}
	scene := core.NewScene("vellum-action-menu")
	core.StyleMenuScene(scene)
	p.scene = scene
	p.context = context
	scene.OnFirst(events.Custom, func(e events.Event) {
		if r.control.wake != nil {
			r.control.wake(e)
		}
	})
	var keyboard func(events.Event)
	for parent := r.widget.AsTree().Parent; !accessibilityTreeNil(parent); parent = parent.AsTree().Parent {
		if handler, ok := parent.AsTree().Property("vellum.accessibility-keyboard").(func(events.Event)); ok {
			keyboard = handler
			break
		}
	}
	itemStyle, _ := ui.ResolveStyle(ui.Node{Kind: ui.Button}, r.theme)
	var first core.Widget
	for _, o := range s.options {
		o := o
		if o.Separator {
			core.NewSeparator(scene).SetName(o.ID)
			continue
		}
		item := tree.New[menuItem](scene)
		item.popup = p
		item.SetName(o.ID)
		item.SetText(html.EscapeString(o.Label))
		if o.Checkable && o.Checked {
			item.SetIcon(nativeSymbol(ui.SymbolCheck))
		}
		item.FinalStyler(func(style *styles.Style) { applyStyle(style, ui.Node{Kind: ui.Button, Style: itemStyle}, ui.Theme{}) })
		item.SetState(o.Disabled, states.Disabled)
		p.items[o.ID] = item
		if first == nil && !o.Disabled {
			first = item
		}
		item.OnClick(func(events.Event) { queueMenuAction(r, p, o.ID) })
		item.OnFinal(events.Click, func(events.Event) {
			if p.stage != nil {
				p.stage.ClosePopup()
			}
		})
		item.OnFocus(func(events.Event) { postControl(r, func() { _ = r.control.a.Invalidate(r.control.window) }) })
	}
	scene.OnFirst(events.KeyChord, func(e events.Event) {
		if e.KeyCode() == key.CodeEscape || (e.KeyCode() == key.CodeTab && (e.Modifiers() == 0 || e.Modifiers() == key.Shift)) {
			if p.stage != nil {
				p.stage.ClosePopup() // native event owns this window's render context
			}
			if e.KeyCode() == key.CodeTab && keyboard != nil {
				keyboard(e) // same event boundary, before declaring handled
			}
			e.SetHandled()
		}
	})
	if first != nil {
		scene.Events.SetStartFocus(first)
	}
	p.stage = core.NewMenuStage(scene, r.widget, r.widget.AsWidget().ContextMenuPos(nil))
	if p.stage == nil {
		scene.DeleteChildren()
		return
	}
	if context != nil {
		context.SetProperty("vellum.menu.active", p)
	}
	s.popup = p
	r.widget.AsWidget().SetFocus()
	p.stage.Run()
}
func queueMenuAction(r *record, p *menuPopup, id string) {
	if p == nil || p.dead.Load() {
		return
	}
	postControl(r, func() {
		if r.control.menu.popup != p || r.control.menu.epoch != p.epoch || !menuAvailable(r) {
			return
		}
		for _, o := range r.node.Options {
			if o.ID == id && !o.Disabled && !o.Separator {
				requestMenuClose(r)
				r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.MenuButton, SelectedID: id, Checked: o.Checkable && !o.Checked})
				_ = r.control.a.Invalidate(r.control.window)
				return
			}
		}
	})
}

// requestMenuClose is owner-only. Global owner work can be drained from a peer
// window, so it must never delete an originating window's native popup scene.
func requestMenuClose(r *record) {
	s := r.control.menu
	p := s.popup
	s.popup = nil
	s.request = false
	if p != nil {
		p.dead.Store(true)
		s.closing = p
	}
}

// flushMenuClose runs only inside originating reconciliation or native teardown.
// Opening a replacement always flushes first, bounding retirement to one popup.
func flushMenuClose(r *record) {
	s := r.control.menu
	p := s.closing
	s.closing = nil
	if p == nil {
		return
	}
	if p.context != nil && p.context.Property("vellum.menu.active") == p {
		p.context.SetProperty("vellum.menu.active", nil)
	}
	p.context = nil
	if p.stage != nil {
		p.stage.ClosePopup()
	}
	p.items = nil
	p.scene = nil
	p.stage = nil
}
func menuActionID(parent, id string) string {
	return "menu:" + base64.RawURLEncoding.EncodeToString([]byte(parent)) + ":" + base64.RawURLEncoding.EncodeToString([]byte(id))
}
func menuAccessibility(r *record, parent string, b *accessibilityBridge) ([]accessibilityNode, map[string]accessibilityHandler) {
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	if !menuOpen(r) {
		return nodes, handlers
	}
	p := r.control.menu.popup
	for _, o := range r.control.menu.options {
		if o.Separator {
			continue
		}
		id := menuActionID(r.node.ID, o.ID)
		node := accessibilityNode{ID: id, Parent: parent, Role: "menu-item", Actions: []string{}, DomainID: o.ID, Name: o.Label, Disabled: o.Disabled, Focusable: !o.Disabled}
		if o.Checkable {
			node.Value = o.Checked
		}
		if item := p.items[o.ID]; item != nil {
			node.Focused = item.StateIs(states.Focused)
		}
		if !o.Disabled {
			node.Actions = []string{"focus", "press"}
			option := o.ID
			handlers[id] = func(a accessibilityAction) bool {
				if !menuOpen(r) || r.control.menu.popup != p {
					return false
				}
				switch a.Kind {
				case "focus":
					if b == nil {
						return false
					}
					accessibilityFocus(b, id)
					return true
				case "press":
					queueMenuAction(r, p, option)
					return true
				default:
					return false
				}
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, handlers
}
func menuPopupBounds(r *record, bounds func(string, core.Widget)) {
	if menuOpen(r) {
		for id, item := range r.control.menu.popup.items {
			bounds(menuActionID(r.node.ID, id), item)
		}
	}
}
