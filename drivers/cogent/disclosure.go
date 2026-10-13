package cogent

import (
	"html"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

// The heading and body are private stable hosts. Collapsing changes visibility,
// never child ownership, so field drafts and keyed native records survive.
type disclosureState struct {
	header *core.Button
	host   *core.Frame
}

func mountDisclosure(parent tree.Node, r *record) core.Widget {
	outer := core.NewFrame(parent)
	c := &disclosureState{}
	r.control.disclosure = c
	outer.Styler(func(s *styles.Style) { s.Direction = styles.Column; s.Grow.Set(1, 0); s.Gap.Zero() })
	c.header = core.NewButton(outer).SetType(core.ButtonText)
	c.header.SetName("vellum-disclosure-heading")
	c.header.Styler(func(s *styles.Style) { s.Grow.Set(1, 0); s.Padding.Set(units.Dp(6)); s.Justify.Content = styles.Start })
	c.header.OnClick(func(events.Event) { queueDisclosure(r, !r.node.Checked) })
	c.header.OnFirst(events.KeyChord, func(e events.Event) {
		if e.IsHandled() || e.Modifiers() != 0 {
			return
		}
		switch e.KeyCode() {
		case key.CodeLeftArrow:
			queueDisclosure(r, false)
		case key.CodeRightArrow:
			queueDisclosure(r, true)
		default:
			return
		}
		e.SetHandled()
	})
	c.host = core.NewFrame(outer)
	c.host.SetName("vellum-disclosure-content")
	c.host.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 0)
		s.Padding.Set(units.Dp(4), units.Dp(12))
		if !r.node.Checked {
			s.Display = styles.DisplayNone
		}
		value, _ := ui.ResolveStyle(r.node, r.theme)
		if value.Gap.Set {
			s.Gap.Set(units.Dp(value.Gap.Value))
		}
	})
	return outer
}

func queueDisclosure(r *record, expanded bool) {
	if r.control.retired || r.control.disabled || r.node.Change == nil || expanded == r.node.Checked {
		return
	}
	accepted := r.node.Checked
	postControl(r, func() {
		// A later programmatic state or disabled ancestor revokes an older request.
		// Owner delivery never inspects or changes native widgets from another window.
		if r.control.disabled || r.node.Change == nil || r.node.Checked != accepted {
			return
		}
		r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.Disclosure, Checked: expanded})
		r.control.a.Invalidate(r.control.window)
	})
}

func syncDisclosure(r *record) {
	c := r.control.disclosure
	if !r.node.Checked && !r.control.disabled && c.host.ContainsFocus() {
		c.header.SetFocus()
	}
	symbol := ui.SymbolRight
	if r.node.Checked {
		symbol = ui.SymbolDown
	}
	c.header.SetText(html.EscapeString(r.node.Text)).SetIcon(nativeSymbol(symbol))
	c.header.Tooltip = html.EscapeString(r.node.Accessibility.Help)
	c.header.SetState(r.control.disabled, states.Disabled)
	// This property is sampled by descendant owner delivery, modal admission,
	// text settlement and file-drop routing independently of stale native styles.
	c.host.SetProperty(disabledProperty, r.control.disabled || !r.node.Checked)
	c.host.SetState(!r.node.Checked, states.Invisible)
}

func accessibilityDisclosureID(id string) string { return "disclosure:" + accessibilityIDPart(id) }

func disclosureAccessibility(r *record, parent string, b *accessibilityBridge, disabled bool) (accessibilityNode, accessibilityHandler) {
	name := r.node.Accessibility.Name
	if name == "" {
		name = r.node.Text
	}
	expanded := r.node.Checked
	n := accessibilityNode{ID: accessibilityDisclosureID(r.node.ID), Parent: parent, DomainID: r.node.ID,
		Role: "disclosure-heading", Name: name, Help: r.node.Accessibility.Help, Value: expanded, Expanded: &expanded,
		Disabled: disabled, Focusable: !disabled, Focused: r.control.disclosure.header.ContainsFocus(), Actions: []string{}}
	if disabled {
		return n, nil
	}
	n.Actions = append(n.Actions, "focus")
	if r.node.Change != nil {
		n.Actions = append(n.Actions, "press", "expand", "collapse")
	}
	allowed := append([]string(nil), n.Actions...)
	return n, func(action accessibilityAction) bool {
		if r.control.retired || r.control.disabled || !containsAccessibilityAction(allowed, action.Kind) {
			return false
		}
		if action.Kind == "focus" {
			if b == nil {
				return false
			}
			accessibilityFocus(b, n.ID)
			return true
		}
		return applyAccessibilityAction(r, action)
	}
}
