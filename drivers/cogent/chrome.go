package cogent

import (
	"html"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

// Managed children occupy host, never the private heading/message structure.
type chromeState struct {
	host    *core.Frame
	caption *core.Text
}

func mountChrome(parent tree.Node, r *record) core.Widget {
	outer := core.NewFrame(parent)
	c := &chromeState{host: outer}
	r.control.chrome = c
	outer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Grow.Set(1, 0)
		s.Padding.Set(units.Dp(4), units.Dp(8))
		s.Gap.Set(units.Dp(8))
		s.Align.Items = styles.Center
		s.Background = colors.Scheme.SurfaceContainer
		if r.node.Kind == ui.Toolbar {
			if r.node.Orientation == ui.Vertical {
				s.Direction = styles.Column
				s.Grow.Set(0, 1)
			}
			// Scroll keeps the same native controls and their private editing state.
			s.Overflow.Set(styles.OverflowAuto, styles.OverflowVisible)
			if r.node.Orientation == ui.Vertical {
				s.Overflow.Set(styles.OverflowVisible, styles.OverflowAuto)
			}
		}
		if r.node.Kind == ui.FormGroup {
			s.Direction = styles.Column
			s.Align.Items = styles.Start
			s.Padding.Set(units.Dp(8))
			s.Border.Width.Set(units.Dp(1))
			s.Border.Color.Set(colors.Scheme.OutlineVariant)
			s.Border.Radius.Set(units.Dp(6))
		}
	})
	if r.node.Kind != ui.Toolbar {
		c.caption = core.NewText(outer)
		c.caption.SetName("vellum-caption")
		c.caption.FinalStyler(func(s *styles.Style) { s.Font.Weight = outer.Styles.Font.Weight })
		c.caption.Styler(func(s *styles.Style) {
			if r.node.Text == "" {
				s.Display = styles.DisplayNone
			}
		})
		c.host = core.NewFrame(outer)
		c.host.SetName("vellum-content")
		c.host.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Grow.Set(1, 0)
			s.Gap.Set(units.Dp(8))
			if r.node.Kind == ui.FormGroup {
				s.Direction = styles.Column
			}
			value, _ := ui.ResolveStyle(r.node, r.theme)
			if value.Gap.Set {
				s.Gap.Set(units.Dp(value.Gap.Value))
			}
			if value.Align.Set {
				s.Align.Items = nativeAlignment(value.Align.Value)
				s.Align.Content = nativeAlignment(value.Align.Value)
			}
			if value.Justify.Set {
				s.Justify.Content = nativeAlignment(value.Justify.Value)
			}
		})
	}
	return outer
}
func syncChrome(r *record, previous ui.Node) {
	c := r.control.chrome
	if c.caption != nil {
		// DisplayNone marks Invisible during Style; explicitly restore both ways
		// before Update so the native walker can reach a returning caption.
		c.caption.SetState(r.node.Text == "", states.Invisible)
		if previous.Text != r.node.Text || c.caption.Text != html.EscapeString(r.node.Text) {
			c.caption.SetText(html.EscapeString(r.node.Text)).Update()
		}
		c.caption.SetState(r.control.disabled, states.Disabled)
	}
}
func toolbarKey(r *record, e events.Event) {
	// macOS desktop input always includes NumLock; lock flags are not chords.
	modifiers := e.Modifiers() & ^((1 << key.CapsLock) | (1 << key.NumLock))
	if e.IsHandled() || modifiers != 0 || r.control.retired {
		return
	}
	step, boundary := 0, 0
	switch e.KeyCode() {
	case key.CodeHome:
		boundary = -1
	case key.CodeEnd:
		boundary = 1
	case key.CodeLeftArrow:
		if r.node.Orientation == ui.Vertical {
			return
		}
		step = -1
	case key.CodeRightArrow:
		if r.node.Orientation == ui.Vertical {
			return
		}
		step = 1
	case key.CodeUpArrow:
		if r.node.Orientation != ui.Vertical {
			return
		}
		step = -1
	case key.CodeDownArrow:
		if r.node.Orientation != ui.Vertical {
			return
		}
		step = 1
	default:
		return
	}
	// A known traversal chord with no enabled target must not fall through to
	// generic Frame navigation, which can focus disabled native children.
	if r.control.disabled {
		e.SetHandled()
		return
	}
	order := ui.FocusOrder(ui.FocusScope(r.node))
	if len(order) == 0 {
		e.SetHandled()
		return
	}
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(r, "", targets)
	target := ""
	if boundary < 0 {
		target = order[0]
	} else if boundary > 0 {
		target = order[len(order)-1]
	} else {
		index := -1
		for i, id := range order {
			if targets[accessibilityManagedID(id)].focused {
				index = i
				break
			}
		}
		if index < 0 && step < 0 {
			index = 0
		}
		target = order[(index+step+len(order))%len(order)]
	}
	if next, ok := targets[accessibilityManagedID(target)]; ok && !tree.IsNil(next.widget) {
		next.widget.AsWidget().SetFocus()
		e.SetHandled()
	}
}

// The native ScrollToThis path intentionally refuses an item larger than its
// viewport when it lies to the right/bottom. Keep oversized toolbar actions
// reachable by showing their leading edge, matching native scroll priority.
// This runs on the native owner after the widget's normal Focus handler;
// the event manager may subsequently apply its ordinary scroll margin.
func revealOversizedToolbarTarget(r *record, target *core.Frame) {
	if r == nil || r.control == nil || r.control.retired || r.control.disabled || r.control.chrome == nil || !target.StateIs(states.Focused) {
		return
	}
	host := r.control.chrome.host
	axis := math32.X
	if r.node.Orientation == ui.Vertical {
		axis = math32.Y
	}
	if host.Geom.ContentBBox.Empty() || !host.HasScroll[axis] {
		return
	}
	raw := math32.RectFromPosSizeMax(target.Geom.Pos.Total, target.Geom.Size.Actual.Total)
	if math32.PointDim(raw.Size(), axis) <= math32.PointDim(host.Geom.ContentBBox.Size(), axis) {
		return
	}
	host.ScrollDimToStart(axis, math32.PointDim(raw.Min, axis))
}

const toolbarScopeProperty = "vellum.toolbar-scope"
const toolbarHookProperty = "vellum.toolbar-hook"

// Bind fallback navigation on descendant layout frames, since each native frame
// handles its own arrows before the event bubbles to the outer toolbar. Focused
// interactive controls still get their normal key handlers first. Scope is read
// dynamically, so reparenting cannot retain a former toolbar's navigation.
func bindToolbarNavigation(root *record) {
	// Ordinary trees avoid both native traversal and binder map allocations.
	var requiresBinding func(*record) bool
	requiresBinding = func(r *record) bool {
		if r == nil || tree.IsNil(r.widget) {
			return false
		}
		if r.node.Kind == ui.Toolbar {
			return true
		}
		if frame := core.AsFrame(r.widget); frame != nil {
			if scope, _ := frame.Property(toolbarScopeProperty).(*record); scope != nil {
				return true
			}
		}
		for _, child := range r.children {
			if requiresBinding(child) {
				return true
			}
		}
		return false
	}
	if !requiresBinding(root) {
		return
	}

	owners := map[tree.Node]*record{}
	needed := map[*record]bool{}
	dialogs := []*record{}
	var collect func(*record) bool
	collect = func(r *record) bool {
		if r == nil || tree.IsNil(r.widget) {
			return false
		}
		owners[r.widget.AsTree().This] = r
		has := r.node.Kind == ui.Toolbar
		if frame := core.AsFrame(r.widget); frame != nil {
			scope, _ := frame.Property(toolbarScopeProperty).(*record)
			has = has || scope != nil
		}
		for _, child := range r.children {
			has = collect(child) || has
		}
		if r.control != nil && r.control.dialog != nil {
			dialogs = append(dialogs, r)
		}
		needed[r] = has
		return has
	}
	if !collect(root) {
		return
	}
	var bind func(tree.Node, *record)
	bind = func(node tree.Node, scope *record) {
		if tree.IsNil(node) {
			return
		}
		if owner := owners[node.AsTree().This]; owner != nil {
			if owner.node.Kind == ui.Extension {
				return
			} // extensions bind their own managed trees
			if scope == nil && !needed[owner] {
				return
			} // skip unrelated native pools/subtrees
			if owner.node.Kind == ui.Toolbar {
				scope = owner
			}
			if owner.node.Kind == ui.Dialog {
				scope = nil
			} // a modal is a separate input scope
		}
		if frame := core.AsFrame(node); frame != nil {
			if scope == nil {
				frame.DeleteProperty(toolbarScopeProperty)
			} else {
				frame.SetProperty(toolbarScopeProperty, scope)
				installed, _ := frame.Property(toolbarHookProperty).(bool)
				if !installed {
					frame.SetProperty(toolbarHookProperty, true)
					frame.OnFinal(events.Focus, func(events.Event) {
						current, _ := frame.Property(toolbarScopeProperty).(*record)
						revealOversizedToolbarTarget(current, frame)
					})
					frame.OnFinal(events.KeyChord, func(e events.Event) {
						current, _ := frame.Property(toolbarScopeProperty).(*record)
						if current != nil {
							toolbarKey(current, e)
						}
					})
				}
			}
		}
		for _, child := range node.AsTree().Children {
			bind(child, scope)
		}
	}
	bind(root.widget, nil)
	// Active modal hosts can be outside the placeholder's native subtree.
	for _, dialog := range dialogs {
		bind(dialog.control.dialog.host, nil)
	}
}
