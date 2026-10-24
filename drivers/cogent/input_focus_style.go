package cogent

import (
	"image"
	"image/color"
	"sync/atomic"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

var uiColorAccent = color.NRGBA{R: 0, G: 112, B: 245, A: 255}

const inputOutlineProperty = "vellum.input-outline"

type inputOutlineState struct{ shared, focused atomic.Bool }

func mountInputOutline(r *record) {
	if r.node.Kind != ui.Row {
		return
	}
	r.inputOutline = &inputOutlineState{}
	r.inputOutline.shared.Store(sharesInputOutline(r.node, r.theme))
	// Install once, before publication. Stylers read this parent's stable state
	// under their native window owner; no model callback mutates the property.
	r.widget.AsTree().SetProperty(inputOutlineProperty, r.inputOutline)
}

func syncInputOutline(r *record) {
	if r.inputOutline == nil {
		return
	}
	next := sharesInputOutline(r.node, r.theme)
	focused := next && !controlDisabled(r) && r.widget.AsWidget().ContainsFocus() && nativeWindowActive(r)
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
		s.MaxBorder.Color.Set(colors.Uniform(managedAccent(r)))
		focused := !controlDisabled(r) && r.widget.AsWidget().ContainsFocus() && nativeWindowActive(r)
		r.inputOutline.focused.Store(focused)
		if focused {
			s.Border = s.MaxBorder
		}
		return
	}
	if r.node.Kind == ui.Button || r.node.Kind == ui.MenuButton || r.node.Kind == ui.Checkbox {
		style, err := ui.ResolveStyle(r.node, r.theme)
		if r.node.ToolbarReorder != nil {
			s.Cursor = cursors.Grab
		}
		if err == nil && style.Border.Set {
			border := style.Border.Value
			restoreManagedBorder(s, border)
			s.MaxBorder = s.Border
			s.MaxBorder.Width.Set(units.Dp(max(border.Width, 2)))
			s.MaxBorder.Color.Set(colors.Uniform(managedAccent(r)))
			focused := s.Is(states.Focused) && r.widget.AsWidget().ContainsFocus() && nativeWindowActive(r) && r.focusVisibility.Load() != 2 && !controlDisabled(r)
			if focused {
				s.Border = s.MaxBorder
			}
			if !nativeWindowActive(r) || r.focusVisibility.Load() == 2 {
				s.StateLayer = 0
			}
		}
		return
	}
	if r.node.Kind != ui.TextField && r.node.Kind != ui.NumberField {
		return
	}
	style, err := ui.ResolveStyle(r.node, r.theme)
	if err != nil {
		return
	}
	if style.Background.Set {
		s.StateLayer = 0
	}
	// Spinner keeps its native maximum outline allowance and arrow geometry.
	if r.node.Kind == ui.NumberField {
		if style.Border.Set && !nativeWindowActive(r) {
			restoreManagedBorder(s, style.Border.Value)
		}
		return
	}
	if !style.Border.Set {
		return
	}
	border := style.Border.Value
	// A zero resting border still permits the native focus cue unless a direct
	// shared-outline parent actually owns it. This also covers transparent
	// standalone fields and composites whose outer border has been removed.
	shared := false
	if parent := r.widget.AsTree().Parent; !tree.IsNil(parent) {
		outline, _ := parent.AsTree().Property(inputOutlineProperty).(*inputOutlineState)
		shared = outline != nil && outline.shared.Load()
	}
	if border.Width == 0 && (!shared || !style.Background.Set || style.Background.Value.A != 0) {
		if !nativeWindowActive(r) {
			restoreManagedBorder(s, border)
		}
		return
	}
	// Native filled-field MaxBorder is a bottom underline. An explicit managed
	// border also owns its focus shape; zero means no private editor outline.
	restoreManagedBorder(s, border)
	s.MaxBorder = s.Border
	if border.Width > 0 {
		s.MaxBorder.Width.Set(units.Dp(max(border.Width, 2)))
		s.MaxBorder.Color.Set(colors.Uniform(managedAccent(r)))
		if !controlDisabled(r) && s.Is(states.Focused) && nativeWindowActive(r) {
			s.Border = s.MaxBorder
		}
	} else if style.Background.Set && style.Background.Value.A == 0 {
		s.RenderBox = false
	}
}

// Change only the painted border. MaxBorder remains reserved by the native
// control so activation never changes its content allocation or hit geometry.
func restoreManagedBorder(s *styles.Style, border ui.Border) {
	s.Border.Width.Set(units.Dp(border.Width))
	s.Border.Color.Set(image.NewUniform(border.Color))
	s.Border.Style.Set(styles.BorderSolid)
}

func managedAccent(r *record) color.NRGBA {
	if r != nil && r.theme.Palette.Set {
		return r.theme.Palette.Value.Accent
	}
	return uiColorAccent
}
