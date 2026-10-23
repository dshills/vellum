package cogent

import (
	"fmt"
	"html"
	"reflect"
	"slices"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

type choiceState struct {
	listAnchor  string
	buttons     map[string]*core.Switch
	order       []string
	listRows    map[string]*listBoxRowEpoch
	listScope   atomic.Pointer[workbenchInputScope]
	listCatalog *listBoxCatalogEpoch
	listContext tabCloseContext // originating reconciliation, before global modal projection
}

func choiceVertical(n ui.Node) bool {
	return n.Kind == ui.ListBox || n.Orientation == ui.Vertical || (n.Orientation == "" && n.Kind == ui.RadioGroup)
}
func choiceOptionEnabled(n ui.Node, id string) bool {
	for _, o := range n.Options {
		if o.ID == id {
			return !n.Disabled && !o.Disabled
		}
	}
	return false
}
func choiceLive(r *record, id string, button *core.Switch) bool {
	return r.control != nil && !r.control.retired && !r.control.disabled && r.control.choice != nil && r.control.choice.buttons[id] == button && choiceOptionEnabled(r.node, id)
}
func queueChoice(r *record, id string, button *core.Switch) {
	if r.node.Kind == ui.ListBox {
		queueListBox(r, id, button)
		return
	}
	postControl(r, func() {
		if !choiceLive(r, id, button) || r.node.Change == nil {
			requestControlRestore(r)
			return
		}
		if r.node.SelectedID != id {
			r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: r.node.Kind, SelectedID: id})
			r.control.committed = true
		}
		requestControlRestore(r) // Never leave a selected radio transiently unchecked.
	})
}
func choiceKey(r *record, id string, button *core.Switch, e events.Event) {
	if e.IsHandled() || !choiceLive(r, id, button) {
		return
	}
	if r.node.Kind == ui.ListBox {
		if (!r.node.Multiple && e.Modifiers() != 0) || e.Modifiers() & ^((1<<key.Shift)|(1<<key.Meta)) != 0 {
			return
		}
		listBoxKey(r, id, button, e)
		return
	}
	if e.Modifiers() != 0 {
		return
	}
	target := ""
	switch e.KeyCode() {
	case key.CodeLeftArrow, key.CodeUpArrow:
		target, _ = ui.NextChoiceOption(r.node, id, -1)
	case key.CodeRightArrow, key.CodeDownArrow:
		target, _ = ui.NextChoiceOption(r.node, id, 1)
	case key.CodeHome:
		target, _ = ui.NextChoiceOption(r.node, "", 1)
	case key.CodeEnd:
		target, _ = ui.NextChoiceOption(r.node, "", -1)
	case key.CodeSpacebar, key.CodeReturnEnter:
		target = id
	default:
		return
	}
	if target == "" {
		return
	}
	e.SetHandled()
	next := r.control.choice.buttons[target]
	next.SetFocus()
	queueChoice(r, target, next)
}

// Native clicks request a domain value; keep accepted selection exclusive until
// the owner's next view admits it. Programmatic setters do not emit Change.
func restoreChoiceSelection(r *record) {
	if r == nil || r.control == nil || r.control.retired || r.control.choice == nil {
		return
	}
	for id, button := range r.control.choice.buttons {
		checked := ui.ListOptionSelected(r.node, id)
		if button.IsChecked() != checked {
			button.SetChecked(checked).Update()
		}
	}
}
func prepareChoice(r *record, n ui.Node) error {
	if err := ui.ValidateControl(n); err != nil {
		return err
	}
	c := r.control.choice
	if n.Kind == ui.ListBox && c.listRows == nil {
		c.listRows = make(map[string]*listBoxRowEpoch)
		setListBoxScope(r, true, r.control.disabled, nil)
	}
	for _, o := range n.Options {
		if c.buttons[o.ID] != nil {
			continue
		}
		id := o.ID
		button := core.NewSwitch(r.widget)
		button.SetName("choice:" + id)
		if n.Kind == ui.ListBox {
			bindContextMenu(r, button)
		}
		c.buttons[id] = button
		button.OnChange(func(e events.Event) {
			if r.node.Kind == ui.ListBox && r.node.Multiple {
				queueListSelection(r, id, button, e.Modifiers(), id)
			} else {
				queueChoice(r, id, button)
			}
			restoreChoiceSelection(r)
		})
		button.OnFirst(events.KeyChord, func(e events.Event) { choiceKey(r, id, button, e) })
		button.FinalStyler(func(s *styles.Style) { styleSourceRow(r, id, s) })
		button.Styler(func(s *styles.Style) {
			if r.node.Kind == ui.ListBox {
				styleListBoxRow(s)
				return
			}
			if r.node.Kind != ui.SegmentedControl {
				return
			}
			// Joined edges retain the focused segment's full outline.
			i := button.IndexInParent()
			last := len(r.node.Options) - 1
			s.Border.Radius.Set(units.Dp(8))
			if last == 0 {
				return
			}
			if choiceVertical(r.node) {
				if i > 0 {
					s.Border.Radius.Top = units.Zero()
					s.Border.Radius.Right = units.Zero()
					if !s.Is(states.Focused) {
						s.Border.Width.SetTop(units.Zero())
					}
				}
				if i < last {
					s.Border.Radius.Bottom = units.Zero()
					s.Border.Radius.Left = units.Zero()
				}
			} else {
				if i > 0 {
					s.Border.Radius.Top = units.Zero()
					s.Border.Radius.Left = units.Zero()
					if !s.Is(states.Focused) {
						s.Border.Width.SetLeft(units.Zero())
					}
				}
				if i < last {
					s.Border.Radius.Right = units.Zero()
					s.Border.Radius.Bottom = units.Zero()
				}
			}
		})
	}
	return nil
}
func finishChoice(r *record, n ui.Node) {
	c := r.control.choice
	keep := make(map[string]bool, len(n.Options))
	children := make([]tree.Node, len(n.Options))
	order := make([]string, len(n.Options))
	for i, o := range n.Options {
		keep[o.ID] = true
		children[i] = c.buttons[o.ID]
		order[i] = o.ID
	}
	for id, button := range c.buttons {
		if !keep[id] {
			delete(c.buttons, id)
			delete(c.listRows, id)
			button.Delete()
		}
	}
	r.widget.AsTree().Children = children
	c.order = order
	if !choiceOptionEnabled(n, c.listAnchor) || !n.Multiple {
		c.listAnchor = ""
	}
}
func syncChoice(r *record, previous ui.Node) error {
	if err := prepareChoice(r, r.node); err != nil {
		return err
	}
	c := r.control.choice
	listCatalogChanged := false
	if r.node.Kind == ui.ListBox {
		listCatalogChanged = syncListBoxEpochs(r)
	}
	for _, o := range r.node.Options {
		b := c.buttons[o.ID]
		typ := core.SwitchRadioButton
		if r.node.Kind == ui.SegmentedControl || r.node.Kind == ui.ListBox {
			typ = core.SwitchSegmentedButton
		}
		text := html.EscapeString(o.Label)
		disabled := r.control.disabled || o.Disabled
		checked := ui.ListOptionSelected(r.node, o.ID)
		dirty := b.Text != text || b.Type != typ || b.IsChecked() != checked || b.StateIs(states.Disabled) != disabled
		b.SetType(typ).SetText(text).SetChecked(checked).SetState(disabled, states.Disabled)
		if dirty || previous.Orientation != r.node.Orientation || !reflect.DeepEqual(previous.Options, r.node.Options) {
			b.Update()
		}
	}
	syncSemantics(r)
	if listCatalogChanged || previous.Multiple != r.node.Multiple || !slices.Equal(previous.SelectedIDs, r.node.SelectedIDs) || previous.Orientation != r.node.Orientation || !reflect.DeepEqual(previous.Options, r.node.Options) || previous.SelectedID != r.node.SelectedID || previous.Disabled != r.node.Disabled || previous.Accessibility != r.node.Accessibility {
		r.widget.AsWidget().Update()
	}
	return nil
}
func choiceFocusWidget(r *record) (core.Widget, error) {
	var first *core.Switch
	for _, o := range r.node.Options {
		if o.Disabled {
			continue
		}
		b := r.control.choice.buttons[o.ID]
		if b == nil {
			continue
		}
		if ui.ListOptionSelected(r.node, o.ID) {
			return b, nil
		}
		if first == nil {
			first = b
		}
	}
	if first != nil {
		return first, nil
	}
	return nil, fmt.Errorf("cogent: choice %q has no enabled option", r.node.ID)
}
func choiceAccessibility(r *record, parent string, b *accessibilityBridge, disabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	nodes := make([]accessibilityNode, 0, len(r.node.Options))
	handlers := make(map[string]accessibilityHandler)
	for _, o := range r.node.Options {
		optionID := o.ID
		button := r.control.choice.buttons[optionID]
		id := accessibilityOptionID(r.node.ID, optionID)
		blocked := disabled || o.Disabled
		n := accessibilityNode{ID: id, Parent: parent, DomainID: optionID, Role: "option", Name: o.Label, Selected: ui.ListOptionSelected(r.node, optionID), Value: ui.ListOptionSelected(r.node, optionID), Disabled: blocked, Focusable: !blocked, Focused: button != nil && button.ContainsFocus(), Actions: []string{}}
		guard := listBoxActionGuard(r, optionID, button)
		if r.node.Kind == ui.ListBox {
			n.Role = "list-option"
			n.RowIndex = len(nodes) + 1
			n.activationGuard = guard
		}
		if !blocked {
			n.Actions = append(n.Actions, "focus")
			if r.node.Change != nil || r.node.Multiple && r.node.SelectionChange != nil {
				n.Actions = append(n.Actions, "press")
			}
			handlers[id] = func(action accessibilityAction) bool {
				if !guard() || !choiceLive(r, optionID, button) {
					return false
				}
				switch action.Kind {
				case "focus":
					if b == nil {
						return false
					}
					accessibilityFocus(b, id)
					return true
				case "press":
					return applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: optionID})
				default:
					return false
				}
			}
		}
		nodes = append(nodes, n)
	}
	return nodes, handlers
}
