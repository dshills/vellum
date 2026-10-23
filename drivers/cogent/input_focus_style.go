package cogent

import (
	"image"
	"sync/atomic"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
)

type inputOutlineState struct{ shared, focused atomic.Bool }

func mountInputOutline(r *record) {
	if r.node.Kind != ui.Row {
		return
	}
	r.inputOutline = &inputOutlineState{}
	r.inputOutline.shared.Store(sharesInputOutline(r.node, r.theme))
}

func syncInputOutline(r *record) {
	if r.inputOutline == nil {
		return
	}
	next := sharesInputOutline(r.node, r.theme)
	focused := next && !controlDisabled(r) && r.widget.AsWidget().ContainsFocus()
	if r.inputOutline.shared.Swap(next) != next || r.inputOutline.focused.Load() != focused {
		r.widget.AsWidget().Update()
	}
}

// A bordered row containing an explicitly borderless, transparent editor owns
// the focus outline for the whole composite (including its accessory buttons).
func sharesInputOutline(n ui.Node, theme ui.Theme) bool {
	if n.Kind != ui.Row {
		return false
	}
	style, err := ui.ResolveStyle(n, theme)
	if err != nil || !style.Border.Set || style.Border.Value.Width <= 0 {
		return false
	}
	for _, child := range n.Children {
		if child.Kind != ui.TextField {
			continue
		}
		style, err := ui.ResolveStyle(child, theme)
		if err == nil && style.Border.Set && style.Border.Value.Width == 0 && style.Background.Set && style.Background.Value.A == 0 {
			return true
		}
	}
	return false
}

func styleManagedInputFocus(r *record, s *styles.Style) {
	if r.node.Kind == ui.Row {
		if r.inputOutline == nil {
			return
		}
		if !r.inputOutline.shared.Load() {
			r.inputOutline.focused.Store(false)
			return
		}
		// Reserve the same outline allowance in both states, avoiding layout
		// changes while the native text client owns a selection or composition.
		s.MaxBorder = s.Border
		s.MaxBorder.Width.Set(units.Dp(max(s.Border.Width.Top.Value, 2)))
		s.MaxBorder.Color.Set(colors.Scheme.Primary.Base)
		focused := !controlDisabled(r) && r.widget.AsWidget().ContainsFocus()
		r.inputOutline.focused.Store(focused)
		if focused {
			s.Border = s.MaxBorder
		}
		return
	}
	if r.node.Kind != ui.TextField {
		return
	}
	style, err := ui.ResolveStyle(r.node, r.theme)
	if err != nil {
		return
	}
	if style.Background.Set {
		s.StateLayer = 0
	}
	if !style.Border.Set {
		return
	}
	border := style.Border.Value
	// Native filled-field MaxBorder is a bottom underline. An explicit managed
	// border also owns its focus shape; zero means no private editor outline.
	s.Border.Width.Set(units.Dp(border.Width))
	s.Border.Color.Set(image.NewUniform(border.Color))
	s.Border.Style.Set(styles.BorderSolid)
	s.MaxBorder = s.Border
	if border.Width > 0 {
		s.MaxBorder.Width.Set(units.Dp(max(border.Width, 2)))
		s.MaxBorder.Color.Set(colors.Scheme.Primary.Base)
		if !controlDisabled(r) && s.Is(states.Focused) {
			s.Border = s.MaxBorder
		}
	} else if style.Background.Set && style.Background.Value.A == 0 {
		s.RenderBox = false
	}
}
