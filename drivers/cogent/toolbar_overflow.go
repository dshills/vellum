package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
	"slices"
)

const toolbarOverflowHiddenProperty = "vellum.toolbar-overflow-hidden"
const toolbarOverflowOwnerProperty = "vellum.toolbar-overflow-owner"

func toolbarOverflowHidden(r *record) bool {
	if r == nil || tree.IsNil(r.widget) {
		return false
	}
	hidden, _ := r.widget.AsWidget().Property(toolbarOverflowHiddenProperty).(bool)
	return hidden
}

// The retained controls remain children of the same native frame. Invisible
// actions are excluded from routing, focus and accessibility; the managed menu
// owns only a short-lived presentation of their application action identities.
type overflowToolbarFrame struct {
	core.Frame
	owner        *record
	widths       []float32
	hidden       []bool
	notify       bool
	focusPending bool
}

func (f *overflowToolbarFrame) SizeUp() {
	if !f.owner.node.ToolbarOverflow {
		f.widths = nil
		f.hidden = nil
		f.notify = false
		f.Frame.SizeUp()
		return
	}
	for _, c := range f.owner.children {
		c.widget.AsWidget().SetState(false, states.Invisible)
	}
	f.Frame.SizeUp()
	f.widths = f.widths[:0]
	for _, c := range f.owner.children {
		f.widths = append(f.widths, c.widget.AsWidget().Geom.Size.Actual.Total.X)
	}

}
func (f *overflowToolbarFrame) SizeFromChildren(iter int, pass core.LayoutPasses) math32.Vector2 {
	size := f.Frame.SizeFromChildren(iter, pass)
	if f.owner.node.ToolbarOverflow && pass == core.SizeUpPass {
		size.X = f.Styles.UnitContext.FontEm * 2
	}
	return size
}
func (f *overflowToolbarFrame) SizeDown(iter int) bool {
	if !f.owner.node.ToolbarOverflow {
		return f.Frame.SizeDown(iter)
	}
	if len(f.widths) != len(f.owner.children) || len(f.widths) < 2 {
		return f.Frame.SizeDown(iter)
	}
	if iter == 0 {
		available := f.Geom.Size.Alloc.Content.X
		if f.Geom.Size.Max.X > 0 {
			available = min(available, f.Geom.Size.Max.X)
		}
		gap := f.Styles.Gap.X.Dots
		count := len(f.widths) - 1
		total := gap * float32(max(0, count-1))
		for _, w := range f.widths[:count] {
			total += w
		}
		fit := count
		if total > available {
			fit = 0
			used := f.widths[count]
			for i, w := range f.widths[:count] {
				if used+gap+w > available {
					break
				}
				used += gap + w
				fit = i + 1
			}
			for fit > 0 && (f.owner.children[fit-1].node.Kind == ui.Separator || ui.IsToolbarSpace(f.owner.children[fit-1].node.Kind)) {
				fit--
			}
		}
		hidden := make([]bool, count+1)
		for i := range count {
			hidden[i] = i >= fit
		}
		hidden[count] = fit == count
		if !slices.Equal(hidden, f.hidden) {
			f.notify = true
			f.hidden = hidden
		}
		focusHidden := false
		for i, c := range f.owner.children {
			wb := c.widget.AsWidget()
			focusHidden = focusHidden || hidden[i] && wb.ContainsFocus()
			wb.SetProperty(toolbarOverflowHiddenProperty, hidden[i])
			wb.SetState(hidden[i], states.Invisible)
		}
		// Rebuild native flex cells for the visible prefix without deleting or
		// reparenting controls. SizeUp measures all items again on the next layout.
		f.Frame.SizeUp()
		if focusHidden && !f.focusPending {
			f.focusPending = true
			// Native Defer runs after layout in the originating Scene context.
			// Revalidate current visibility so a newer layout cannot steal focus.
			f.Defer(func() {
				f.focusPending = false
				if f.owner.control.retired {
					return
				}
				needsFocus := false
				for _, child := range f.owner.children {
					needsFocus = needsFocus || toolbarOverflowHidden(child) && child.widget.AsWidget().ContainsFocus()
				}
				if !needsFocus {
					return
				}
				for i := len(f.owner.children) - 1; i >= 0; i-- {
					child := f.owner.children[i]
					if child.node.Kind == ui.MenuButton && !toolbarOverflowHidden(child) && menuAvailable(child) {
						child.widget.AsWidget().SetFocus()
						return
					}
				}
				for _, child := range f.owner.children {
					if child.node.Kind == ui.Button && !toolbarOverflowHidden(child) && !child.control.disabled && child.node.Activate != nil {
						child.widget.AsWidget().SetFocus()
						return
					}
				}
			})
		}
	}
	return f.Frame.SizeDown(iter)
}
func (f *overflowToolbarFrame) Render() {
	f.Frame.Render()
	if !f.notify {
		return
	}
	f.notify = false
	r := f.owner
	postControl(r, func() {
		if len(r.children) > 0 {
			more := r.children[len(r.children)-1]
			if more.control != nil && more.control.menu != nil {
				requestMenuClose(more)
			}
		}
		_ = r.control.a.Invalidate(r.control.window)
	})
}

func toolbarOverflowOptions(owner *record, options []ui.Option) []ui.Option {
	out := make([]ui.Option, 0, len(options))
	if len(options) != len(owner.children)-1 {
		return out
	}
	hidden := make(map[string]bool, len(owner.children))
	for _, child := range owner.children {
		if !ui.IsToolbarSpace(child.node.Kind) && toolbarOverflowHidden(child) {
			hidden[child.node.ID] = true
		}
	}
	for _, o := range options {
		if !hidden[o.ID] {
			continue
		}
		if o.Separator && (len(out) == 0 || out[len(out)-1].Separator) {
			continue
		}
		out = append(out, o)
	}
	if len(out) > 0 && out[len(out)-1].Separator {
		out = out[:len(out)-1]
	}
	return out
}

func installMenuArtwork(b *core.Button, option ui.Option) {
	if option.IconImage == nil {
		return
	}
	b.Styler(func(s *styles.Style) { s.Gap.Set(units.Dp(6)) })
	b.FirstMaker(func(p *tree.Plan) {
		tree.AddAt(p, buttonImageName, func(w *managedImage) {
			w.source = option.IconImage
			w.fit = ui.ImageContain
			w.Styler(func(s *styles.Style) {
				s.Grow.Set(0, 0)
				s.Padding.Zero()
				s.Margin.Zero()
				s.Min.Set(units.Dp(18))
				s.Max = s.Min
			})
		})
	})
}
