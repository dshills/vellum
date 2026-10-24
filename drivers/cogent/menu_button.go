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
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

type menuEpoch struct{ marker byte }
type menuState struct {
	epoch         *menuEpoch
	options       []ui.Option
	popup         *menuPopup
	closing       *menuPopup // owner revokes first; originating reconciliation destroys native scene
	request       bool
	branchRequest string
	position      *image.Point
}
type menuPopup struct {
	parent, child     *menuPopup
	options           []ui.Option
	branchID          string
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
			if p.parent == nil && p.owner.control.menu.popup == p {
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
			if p.owner.control.menu.popup == p.top() {
				_ = p.owner.control.a.Invalidate(p.owner.control.window)
			}
		})
	}
}

// Retain legacy chooser behavior for large catalogues or unscoped themes.
func managedSelector(r *record) bool {
	return r.node.Kind == ui.Selector && r.theme.Palette.Set && len(r.node.Options) > 0 && len(r.node.Options) <= ui.MaxMenuOptions
}
func menuDescriptor(r *record) ui.Node {
	if r.node.Kind == ui.Selector {
		options := slices.Clone(r.node.Options)
		for i := range options {
			options[i].Checkable = true
			options[i].Checked = options[i].ID == r.node.SelectedID
		}
		n := ui.Node{ID: r.node.ID, Kind: ui.MenuButton, Text: "Choose value", Disabled: r.node.Disabled, Options: options}
		if r.node.Change != nil {
			n.Change = func(e ui.ChangeEvent) { e.Kind = ui.Selector; e.Checked = false; r.node.Change(e) }
		}
		return n
	}

	if r.node.Kind == ui.MenuButton {
		n := r.node
		if owner, _ := r.widget.AsWidget().Property(toolbarOverflowOwnerProperty).(*record); owner != nil && owner.node.ToolbarOverflow {
			n.Options = toolbarOverflowOptions(owner, n.Options)
		}
		return n
	}
	n := ui.Node{ID: r.node.ID, Kind: ui.MenuButton, Text: "Context actions", Disabled: r.node.Disabled}
	if r.node.ContextMenu != nil {
		n.Options = r.node.ContextMenu.Options
		n.Change = r.node.ContextMenu.Change
	}
	return n
}

const toolbarContextOwnerProperty = "vellum.toolbar-context-owner"

func bindContextMenu(r *record, w core.Widget) {
	request := func(e events.Event) {
		if e.IsHandled() {
			return
		}
		host := r
		if host.node.ContextMenu == nil {
			// Existing button/space routing forwards to its nearest toolbar. A child's
			// explicit context menu still wins. Traversal stays under native ownership.
			host = nil
			for p := w.AsTree().Parent; p != nil; p = p.AsTree().Parent {
				owner, _ := p.AsTree().Property(toolbarContextOwnerProperty).(*record)
				if owner != nil && owner.node.ContextMenu != nil {
					host = owner
					break
				}
			}
		}
		if host == nil || host.node.ContextMenu == nil || host.control == nil || host.control.menu == nil {
			return
		}
		e.SetHandled()
		var position *image.Point
		if e.HasPos() {
			p := e.WindowPos()
			position = &p
		}
		epoch := host.control.menu.epoch
		postControl(host, func() {
			if menuAvailable(host) && host.control.menu.epoch == epoch {
				host.control.menu.position = position
				host.control.menu.request = true
				_ = host.control.a.Invalidate(host.control.window)
			}
		})
	}
	w.AsWidget().OnFirst(events.ContextMenu, request)
	w.AsWidget().OnFirst(events.KeyChord, func(e events.Event) {
		mods := e.Modifiers() & ^((1 << key.CapsLock) | (1 << key.NumLock))
		if e.KeyCode() == key.CodeF10 && mods == 1<<key.Shift {
			request(e)
		}
	})
}

func menuAvailable(r *record) bool {
	if r.control == nil || r.control.retired || r.control.disabled || toolbarOverflowHidden(r) || menuDescriptor(r).Change == nil {
		return false
	}
	return ui.MenuEnabled(menuDescriptor(r))
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
		if !e.IsHandled() && e.Modifiers() & ^((1<<key.CapsLock)|(1<<key.NumLock)) == 0 && e.KeyCode() == key.CodeDownArrow {
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
	changed := !slices.Equal(s.options, menuDescriptor(r).Options) || previous.MenuAnchorID != r.node.MenuAnchorID || previous.Text != r.node.Text || previous.Icon != r.node.Icon || previous.IconImage != r.node.IconImage || previous.Appearance != r.node.Appearance || previous.Style != r.node.Style || previous.Grow != r.node.Grow || previous.Disabled != r.node.Disabled || previous.MinWidth != r.node.MinWidth || previous.MinHeight != r.node.MinHeight
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
		_, overflow := r.widget.AsWidget().Property(toolbarOverflowOwnerProperty).(*record)
		if r.node.Text == "" || overflow {
			indicator = ""
		}
		if b.Indicator != indicator {
			b.SetIndicator(indicator).Update()
		}
		r.widget.AsWidget().SetState(!menuAvailable(r), states.Disabled)
	}
	if id := s.branchRequest; id != "" {
		s.branchRequest = ""
		if menuOpen(r) {
			if item := menuItems(s.popup)[id]; item != nil && item.Menu != nil {
				item.Send(events.Click)
			}
		}
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
	p := &menuPopup{owner: r, epoch: s.epoch, items: map[string]*menuItem{}, options: s.options}
	scene := core.NewScene("vellum-action-menu")
	core.StyleMenuScene(scene)
	scene.FinalStyler(func(style *styles.Style) {
		if r.theme.Palette.Set {
			style.Background = colors.Uniform(r.theme.Palette.Value.Popup)
		} else if background := r.theme.Kinds[ui.FormGroup].Background; background.Set {
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
	populateManagedMenu(r, p, scene, keyboard)

	menuWidth := 0
	pos := r.widget.AsWidget().ContextMenuPos(nil)
	if r.node.Kind == ui.MenuButton || r.node.Kind == ui.Selector {
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
	if r.node.Kind != ui.Toolbar {
		r.widget.AsWidget().SetFocus()
	}
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

func (p *menuPopup) top() *menuPopup {
	for p.parent != nil {
		p = p.parent
	}
	return p
}
func menuItems(p *menuPopup) map[string]*menuItem {
	out := map[string]*menuItem{}
	for ; p != nil && !p.dead.Load(); p = p.child {
		for id, item := range p.items {
			out[id] = item
		}
	}
	return out
}
func menuVisibleOptions(p *menuPopup) []ui.Option {
	var out []ui.Option
	for ; p != nil && !p.dead.Load(); p = p.child {
		out = append(out, p.options...)
	}
	return out
}

// Native-context only: each level releases children before its own stage.
func closeMenuLevel(p *menuPopup) {
	if p == nil {
		return
	}
	closeMenuLevel(p.child)
	p.child = nil
	p.dead.Store(true)
	st := p.stage
	if st == nil && p.scene != nil {
		st = p.scene.Stage
	}
	if st != nil {
		st.ClosePopup()
	}
	p.items = nil
	p.scene = nil
	p.stage = nil
	p.context = nil
}
func closeMenuChain(p *menuPopup) {
	closeMenuLevel(p)
	if p != nil {
		postControl(p.owner, func() {
			if p.owner.control.menu.popup == p {
				requestMenuClose(p.owner)
				_ = p.owner.control.a.Invalidate(p.owner.control.window)
			}
		})
	}
}
func populateManagedMenu(r *record, p *menuPopup, scene *core.Scene, keyboard func(events.Event)) {
	scene.FinalStyler(func(s *styles.Style) {
		if r.theme.Palette.Set {
			s.Background = colors.Uniform(r.theme.Palette.Value.Popup)
		}
	})
	// Menu rows inherit typography/ink, not ordinary push-button borders/fills.
	itemStyle := ui.Style{Typography: r.theme.Default.Typography, Foreground: r.theme.Default.Foreground}
	var first core.Widget
	for _, o := range p.options {
		o := o
		if o.Separator {
			core.NewSeparator(scene).SetName(o.ID)
			continue
		}
		item := tree.New[menuItem](scene)
		item.popup = p
		item.SetName(o.ID)
		item.SetText(html.EscapeString(o.Label))
		if o.Icon != "" {
			item.SetIcon(nativeSymbol(o.Icon))
		}
		if o.Checkable && o.Checked {
			if o.Icon != "" || o.IconImage != nil {
				item.SetIndicator(nativeSymbol(ui.SymbolCheck))
			} else {
				item.SetIcon(nativeSymbol(ui.SymbolCheck))
			}
		}
		installMenuArtwork(&item.Button, o)

		item.FinalStyler(func(style *styles.Style) {
			style.SetAbilities(!o.Disabled, abilities.Focusable, abilities.Activatable)
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
				style.Background = colors.Uniform(managedAccent(r))
				style.Color = colors.Uniform(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
				if r.theme.Palette.Set {
					style.Color = colors.Uniform(r.theme.Palette.Value.AccentInk)
				}
			}
		})
		item.SetState(o.Disabled, states.Disabled)
		p.items[o.ID] = item
		if (first == nil || r.node.Kind == ui.Selector && o.Checked) && !o.Disabled {
			first = item
		}
		if o.Shortcut != "" {
			item.Maker(func(plan *tree.Plan) {
				tree.AddAt(plan, "menu-caption-space", func(w *core.Stretch) {})
				tree.AddAt(plan, "menu-caption", func(w *core.Text) {
					w.SetText(html.EscapeString(o.Shortcut))
					w.Styler(func(s *styles.Style) {
						s.Margin.Left = units.Dp(24)
					})
				})
			})
		}
		if o.Branch != nil {
			item.SetIndicator(nativeSymbol(ui.SymbolRight))
			item.SetMenu(func(sc *core.Scene, pos image.Point) {
				if p.dead.Load() {
					return
				}
				closeMenuLevel(p.child)
				q := &menuPopup{owner: r, epoch: p.epoch, parent: p, branchID: o.ID, options: o.Branch.Items(), scene: sc, context: p.context, items: map[string]*menuItem{}}
				p.child = q
				populateManagedMenu(r, q, sc, keyboard)
			})
		} else {
			item.OnClick(func(events.Event) { queueMenuAction(r, p, o.ID) })
			item.OnFinal(events.Click, func(events.Event) { closeMenuChain(p.top()) })
		}

		item.OnFocus(func(events.Event) { postControl(r, func() { _ = r.control.a.Invalidate(r.control.window) }) })
	}
	scene.OnFirst(events.KeyChord, func(e events.Event) {
		modifiers := e.Modifiers() & ^((1 << key.CapsLock) | (1 << key.NumLock))
		if e.KeyCode() == key.CodeEscape || (e.KeyCode() == key.CodeTab && (modifiers == 0 || modifiers == 1<<key.Shift)) {
			closeMenuChain(p.top())
			if e.KeyCode() == key.CodeTab && keyboard != nil {
				keyboard(e) // same event boundary, before declaring handled
			}
			e.SetHandled()
		}
	})
	if first != nil {
		scene.Events.SetStartFocus(first)
	}
}
func queueMenuAction(r *record, p *menuPopup, id string) {
	if p == nil || p.dead.Load() {
		return
	}
	postControl(r, func() {
		if r.control.menu.popup != p.top() || r.control.menu.epoch != p.epoch || !menuAvailable(r) {
			return
		}
		if o, ok := ui.MenuAction(menuDescriptor(r).Options, id); ok {
			requestMenuClose(r)
			menuDescriptor(r).Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.MenuButton, SelectedID: id, Checked: o.Checkable && !o.Checked})
			_ = r.control.a.Invalidate(r.control.window)
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
	s.branchRequest = ""
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
	closeMenuChain(p)
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
	for _, o := range menuVisibleOptions(p) {
		if o.Separator {
			continue
		}
		id := menuActionID(r.node.ID, o.ID)
		optionParent := parent
		for q := p.child; q != nil && !q.dead.Load(); q = q.child {
			for _, child := range q.options {
				if child.ID == o.ID {
					optionParent = menuActionID(r.node.ID, q.branchID)
				}
			}
		}
		node := accessibilityNode{ID: id, Parent: optionParent, Role: "menu-item", Actions: []string{}, DomainID: o.ID, Name: o.Label, Disabled: o.Disabled, Focusable: !o.Disabled}
		if o.Checkable {
			node.Value = o.Checked
		}
		if item := menuItems(p)[o.ID]; item != nil {
			node.Focused = item.StateIs(states.Focused)
		}
		if !o.Disabled {
			node.Actions = []string{"focus", "press"}
			option := o.ID
			branch := o.Branch != nil
			nativeItem := menuItems(p)[o.ID]
			handlers[id] = func(a accessibilityAction) bool {
				if !menuOpen(r) || r.control.menu.popup != p || nativeItem == nil || nativeItem.popup.dead.Load() {
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
					if branch {
						postControl(r, func() {
							if menuOpen(r) && r.control.menu.popup == p && !nativeItem.popup.dead.Load() {
								r.control.menu.branchRequest = option
								_ = r.control.a.Invalidate(r.control.window)
							}
						})
					} else {
						queueMenuAction(r, nativeItem.popup, option)
					}
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
		for id, item := range menuItems(r.control.menu.popup) {
			bounds(menuActionID(r.node.ID, id), item)
		}
	}
}
