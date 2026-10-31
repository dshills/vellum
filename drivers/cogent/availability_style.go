package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
)

// Disabled appearance must not withdraw layout allowance reserved by the enabled
// control. Preserve only maximum decoration space: no focusability, handlers or
// painted disabled borders/shadows are restored here.
func styleManagedAvailability(r *record, s *styles.Style) {
	if !s.Is(states.Disabled) {
		return
	}
	value, err := ui.ResolveStyle(r.node, r.theme)
	if err != nil {
		return
	}
	switch r.node.Kind {
	case ui.Button, ui.MenuButton:
		// Managed Mac borders intentionally suppress Material shadow space.
		if r.theme.Palette.Set && value.Border.Set {
			return
		}
		switch r.widget.(*core.Button).Type {
		case core.ButtonAction, core.ButtonMenu:
			s.MaxBoxShadow = styles.BoxShadow0()
		case core.ButtonElevated:
			s.MaxBoxShadow = styles.BoxShadow2()
		default:
			s.MaxBoxShadow = styles.BoxShadow1()
		}
	case ui.Checkbox:
		if value.Border.Set {
			return
		}
		s.MaxBorder.Width.Set(units.Dp(1))
	case ui.TextField:
		if value.Border.Set {
			return
		}
		s.MaxBorder = s.Border
		if r.widget.(*core.TextField).Type == core.TextFieldFilled {
			s.MaxBorder.Width.Bottom = units.Dp(2)
		} else {
			s.MaxBorder.Width.Set(units.Dp(2))
		}
	}
}
