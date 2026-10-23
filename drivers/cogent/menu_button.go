package cogent

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
	"encoding/base64"
	"html"
	"image"
	"image/color"
	"slices"
	"sync/atomic"

	"cogentcore.org/core/colors"
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
	epoch    *menuEpoch
	options  []ui.Option
	popup    *menuPopup
	closing  *menuPopup // owner revokes first; originating reconciliation destroys native scene
	request  bool
	position *image.Point
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
func menuDescriptor(r *record) ui.Node {
	if r.node.Kind == ui.MenuButton {
		return r.node
	}
	n := ui.Node{ID: r.node.ID, Kind: ui.MenuButton, Text: "Context actions", Disabled: r.node.Disabled}
	if r.node.ContextMenu != nil {
		n.Options = r.node.ContextMenu.Options
		n.Change = r.node.ContextMenu.Change
	}
	return n
}
func bindContextMenu(r *record, w core.Widget) {
	request := func(e events.Event) {
		if e.IsHandled() || r.node.ContextMenu == nil {
			return
		}
		e.SetHandled()
		var position *image.Point
		if e.HasPos() {
			p := e.WindowPos()
			position = &p
		}
		epoch := r.control.menu.epoch
		postControl(r, func() {
			if menuAvailable(r) && r.control.menu.epoch == epoch {
				r.control.menu.position = position
				r.control.menu.request = true
				_ = r.control.a.Invalidate(r.control.window)
			}
		})
	}
	w.AsWidget().OnFirst(events.ContextMenu, request)
	w.AsWidget().OnFirst(events.KeyChord, func(e events.Event) {
		if e.KeyCode() == key.CodeF10 && e.Modifiers() == 1<<key.Shift {
			request(e)
		}
	})
}
func menuAvailable(r *record) bool {
	if r.control == nil || r.control.retired || r.control.disabled || menuDescriptor(r).Change == nil {
		return false
	}
	for _, o := range menuDescriptor(r).Options {
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
	changed := !slices.Equal(s.options, menuDescriptor(r).Options) || previous.MenuAnchorID != r.node.MenuAnchorID || previous.Text != r.node.Text || previous.Icon != r.node.Icon || previous.Appearance != r.node.Appearance || previous.Style != r.node.Style || previous.Grow != r.node.Grow || previous.Disabled != r.node.Disabled || previous.MinWidth != r.node.MinWidth || previous.MinHeight != r.node.MinHeight
	if changed {
		requestMenuClose(r)
		s.epoch = &menuEpoch{}
		s.options = slices.Clone(menuDescriptor(r).Options)
	}
	if !menuAvailable(r) {
		requestMenuClose(r)
	}
	flushMenuClose(r)
	if r.node.Kind == ui.MenuButton {
		// An icon-only affordance is already the disclosure indicator (e.g. combo).
		// A second indicator clips in a compact Mac field. Restore it with a label.
		b := r.widget.(*core.Button)
		indicator := nativeSymbol(ui.SymbolDown)
		if r.node.Text == "" {
			indicator = ""
		}
		if b.Indicator != indicator {
			b.SetIndicator(indicator).Update()
		}
		r.widget.AsWidget().SetState(!menuAvailable(r), states.Disabled)
	}
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
	scene.FinalStyler(func(style *styles.Style) {
		if background := r.theme.Kinds[ui.FormGroup].Background; background.Set {
			style.Background = colors.Uniform(background.Value)
		}
	})
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
	// Menu rows inherit typography/ink, not ordinary push-button borders/fills.
	itemStyle := ui.Style{Typography: r.theme.Default.Typography, Foreground: r.theme.Default.Foreground}
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
		item.FinalStyler(func(style *styles.Style) {
			applyStyle(style, ui.Node{Kind: ui.Label, Style: itemStyle}, ui.Theme{})
			style.Grow.Set(1, 0)
			style.Justify.Content = styles.Start
			style.Border.Width.Zero()
			style.MaxBorder.Width.Zero()
			style.Border.Radius.Set(units.Dp(4))
			style.Padding.Set(units.Dp(4), units.Dp(10))
			style.Min.Y = units.Dp(24)
			style.Background = colors.Uniform(color.NRGBA{})
			if style.Is(states.Focused) || style.Is(states.Hovered) || style.Is(states.Active) {
				style.Background = colors.Uniform(color.NRGBA{R: 0, G: 112, B: 235, A: 255})
				style.Color = colors.Uniform(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		})
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
		if e.KeyCode() == key.CodeEscape || (e.KeyCode() == key.CodeTab && (e.Modifiers() == 0 || e.Modifiers() == 1<<key.Shift)) {
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
	menuWidth := 0
	pos := r.widget.AsWidget().ContextMenuPos(nil)
	if r.node.Kind == ui.MenuButton {
		anchor := r.widget
		if r.node.MenuAnchorID != "" {
			anchor = nil
			if sc := r.widget.AsWidget().Scene; sc != nil {
				sc.WidgetWalkDown(func(w core.Widget, b *core.WidgetBase) bool {
					if sem, ok := b.Property(SemanticsProperty).(Semantics); ok && sem.ID == r.node.MenuAnchorID && !sem.Disabled && b.IsVisible() {
						anchor = w
						return false
					}
					return true
				})
			}
		}
		if anchor == nil {
			scene.DeleteChildren()
			s.popup = nil
			return
		}
		box := anchor.AsWidget().Geom.TotalBBox
		if r.node.MenuAnchorID != "" {
			menuWidth = box.Dx()
		}
		offset := image.Point{}
		if anchor.AsWidget().Scene != nil {
			offset = anchor.AsWidget().Scene.SceneGeom.Pos
		}
		pos = image.Pt(box.Min.X, box.Max.Y+2).Add(offset)
	}
	if s.position != nil {
		pos = *s.position
	}
	s.position = nil
	p.stage = core.NewMenuStage(scene, r.widget, pos)
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
	if menuWidth > 0 && context != nil {
		geom := scene.SceneGeom
		bounds := context.SceneGeom.Bounds()
		margin := scene.Styles.TotalMargin()
		geom.Size.X = min(menuWidth+int(margin.Left+margin.Right), bounds.Dx())
		geom.Pos.X -= int(margin.Left)
		geom.Pos.Y -= int(margin.Top)
		geom.Pos.X = max(bounds.Min.X, min(geom.Pos.X, bounds.Max.X-geom.Size.X))
		scene.Resize(math32.Geom2DInt{Pos: geom.Pos, Size: geom.Size})
		scene.LayoutScene()
	}
}
func queueMenuAction(r *record, p *menuPopup, id string) {
	if p == nil || p.dead.Load() {
		return
	}
	postControl(r, func() {
		if r.control.menu.popup != p || r.control.menu.epoch != p.epoch || !menuAvailable(r) {
			return
		}
		for _, o := range menuDescriptor(r).Options {
			if o.ID == id && !o.Disabled && !o.Separator {
				requestMenuClose(r)
				menuDescriptor(r).Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.MenuButton, SelectedID: id, Checked: o.Checkable && !o.Checked})
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
	s.position = nil
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
