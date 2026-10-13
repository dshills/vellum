package cogent

import (
	"html"
	"strings"

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

// Private pane hosts preserve scroll/layout state and isolate native Splits'
// child styling from the managed child descriptors. Parts owns one native handle.
type splitState struct {
	native     *core.Splits
	panes      [2]*core.Frame
	childIDs   [2]string
	plannedIDs [2]string // Preparation routes children without publishing accepted binding.
	handle     *core.Handle
}

// Native Splits hides any pane below 1%, even though a positive managed fraction
// can be smaller. Use its Frame renderer to retain content/Parts/scroll drawing
// without that implicit collapse policy. Layout still uses native Splits sizing.
type managedSplit struct{ core.Splits }

func (s *managedSplit) RenderWidget() { s.Frame.RenderWidget() }

// Keep positive accepted fractions drawable at subpixel sizes. Temporary
// display fractions are used only during layout; accepted/native state is
// restored before returning, and no owner callbacks run in these methods.
func (s *managedSplit) withDisplayFractions(layout func()) {
	if len(s.TileSplits) != 2 || s.Parts == nil || s.Parts.NumChildren() != 1 {
		layout()
		return
	}
	dim := s.Styles.Direction.Dim()
	space := s.Geom.Size.Alloc.Content.Dim(dim) - s.Parts.Child(0).(core.Widget).AsWidget().Geom.Size.Actual.Total.Dim(dim) - s.Styles.Gap.Dots().Floor().Dim(dim)
	if space < 2 {
		layout()
		return
	}
	first := math32.Round(s.TileSplits[0] * space)
	if first >= 1 && first <= space-1 {
		layout()
		return
	}
	accepted := s.TileSplits
	display := [2]float32{max(1, min(first, space-1)) / space, 0}
	display[1] = 1 - display[0]
	s.TileSplits = display[:]
	defer func() { s.TileSplits = accepted }()
	layout()
}

func (s *managedSplit) SizeDownSetAllocs(iter int) {
	s.withDisplayFractions(func() { s.Splits.SizeDownSetAllocs(iter) })
}

func (s *managedSplit) Position() { s.withDisplayFractions(s.Splits.Position) }

func mountSplit(parent tree.Node, r *record) core.Widget {
	outer := tree.New[managedSplit](parent)
	w := &outer.Splits
	c := &splitState{native: w}
	r.control.split = c
	for i := range c.panes {
		c.panes[i] = core.NewFrame(w)
	}
	w.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		if r.node.Orientation == ui.Vertical {
			s.Direction = styles.Column
		}
		s.Min.Set(units.Dp(120), units.Dp(80))
		s.Gap.Zero()
	})
	w.SetSplits(r.node.Number, 1-r.node.Number).Update()
	c.handle = w.Parts.Child(0).(*core.Handle)
	// Keep the native Maker's handle-0 key: renaming it causes Update to
	// replace the widget and retire our retained input/focus binding.
	c.handle.Styler(func(s *styles.Style) {
		if s.Is(states.Focused) {
			s.Background = colors.Scheme.Primary.Base
		}
	})
	// Default native digit shortcuts collapse/equalize panes without owner
	// acceptance. They cannot bypass this controlled two-pane descriptor.
	// Normal listeners run newest first, after the scene's command First phase.
	w.On(events.KeyChord, func(e events.Event) {
		chord := string(e.KeyChord())
		for _, prefix := range []string{"Meta+", "Control+"} {
			if digit := strings.TrimPrefix(chord, prefix); digit != chord && len(digit) == 1 && digit[0] >= '0' && digit[0] <= '9' {
				e.SetHandled()
				return
			}
		}
	})
	c.handle.OnFirst(events.KeyChord, func(e events.Event) { splitKey(r, e) })
	c.handle.OnFirst(events.SlideMove, func(e events.Event) {
		e.SetHandled() // Never invoke native optimistic collapse/normalization.
		if !splitWritable(r) {
			return
		}
		span := c.handle.Max - c.handle.Min
		if span <= 0 {
			return
		}
		position := w.AsWidget().PointToRelPos(e.Pos())
		axis := math32.X
		if r.node.Orientation == ui.Vertical {
			axis = math32.Y
		}
		value := (math32.FromPoint(position).Dim(axis) - c.handle.Min) / span
		queueSplit(r, value)
	})
	return outer
}

func splitWritable(r *record) bool {
	return r.control != nil && r.control.split != nil && !r.control.retired && !r.control.disabled && r.node.Change != nil
}

func prepareSplit(r *record, n ui.Node) error {
	if err := ui.ValidateControl(n); err != nil {
		return err
	}
	for i, child := range n.Children {
		r.control.split.plannedIDs[i] = child.ID
	}
	return nil
}

func queueSplit(r *record, value float32) {
	if !splitWritable(r) {
		return
	}
	value, ok := ui.ClampNumeric(value, ui.SplitRange(r.node.Numeric))
	if !ok {
		return
	}
	panes, orientation := r.control.split.childIDs, r.node.Orientation
	postControl(r, func() {
		if !splitWritable(r) || r.control.split.childIDs != panes || r.node.Orientation != orientation {
			requestControlRestore(r)
			return
		}
		current, ok := ui.ClampNumeric(value, ui.SplitRange(r.node.Numeric))
		if ok && current != r.node.Number {
			r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.SplitPane, Number: current})
			r.control.committed = true
		}
		requestControlRestore(r)
	})
}

func splitKey(r *record, e events.Event) {
	if e.IsHandled() || e.Modifiers() != 0 {
		return
	}
	bounds := ui.SplitRange(r.node.Numeric)
	value := r.node.Number
	switch e.KeyCode() {
	case key.CodeHome:
		value = bounds.Min
	case key.CodeEnd:
		value = bounds.Max
	case key.CodeLeftArrow:
		if r.node.Orientation == ui.Vertical {
			return
		}
		value -= bounds.Step
	case key.CodeRightArrow:
		if r.node.Orientation == ui.Vertical {
			return
		}
		value += bounds.Step
	case key.CodeUpArrow:
		if r.node.Orientation != ui.Vertical {
			return
		}
		value -= bounds.Step
	case key.CodeDownArrow:
		if r.node.Orientation != ui.Vertical {
			return
		}
		value += bounds.Step
	default:
		return
	}
	e.SetHandled()
	queueSplit(r, value)
}

func syncSplit(r *record, previous ui.Node) {
	c := r.control.split
	w := c.native
	dirty := w.Split(0) != r.node.Number || previous.Orientation != r.node.Orientation || previous.Numeric != r.node.Numeric || c.handle.StateIs(states.Disabled) != r.control.disabled || c.handle.StateIs(states.ReadOnly) != (r.node.Change == nil)
	w.SetSplits(r.node.Number, 1-r.node.Number)
	c.handle.SetState(r.control.disabled, states.Disabled)
	c.handle.SetState(r.node.Change == nil, states.ReadOnly)
	c.handle.Tooltip = html.EscapeString(r.node.Accessibility.Help)
	if dirty {
		w.Update()
		c.handle.Update()
	}
}

func accessibilitySplitHandleID(id string) string { return "split:" + accessibilityIDPart(id) }

func splitAccessibility(r *record, parent string, b *accessibilityBridge, disabled bool) (accessibilityNode, accessibilityHandler) {
	bounds := ui.SplitRange(r.node.Numeric)
	n := accessibilityNode{ID: accessibilitySplitHandleID(r.node.ID), Parent: parent, DomainID: r.node.ID, Role: "split-handle", Name: r.node.Accessibility.Name, Help: r.node.Accessibility.Help, Value: r.node.Number, Min: &bounds.Min, Max: &bounds.Max, Disabled: disabled, Focusable: !disabled, Focused: r.control.split.handle.ContainsFocus(), Actions: []string{}, Orientation: "vertical"}
	if n.Name == "" {
		n.Name = r.node.Text
	}
	if r.node.Orientation == ui.Vertical {
		n.Orientation = "horizontal"
	}
	if disabled {
		return n, nil
	}
	n.Actions = append(n.Actions, "focus")
	if r.node.Change != nil {
		n.Actions = append(n.Actions, "set", "increment", "decrement")
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
