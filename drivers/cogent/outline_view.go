package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"fmt"
	"github.com/dshills/vellum/ui"
	"html"
	"slices"
)

func outlineViewRow(r *record, id string) ui.OutlineViewRow {
	if l := r.node.Outline; l != nil {
		for _, row := range l.Rows {
			if row.ID == id {
				return row
			}
		}
	}
	return ui.OutlineViewRow{}
}
func outlineViewChanged(a, b *ui.OutlineViewLayout) bool {
	if a == nil || b == nil {
		return a != b
	}
	return !a.SameSource(b) || !slices.Equal(a.Rows, b.Rows)
}
func queueOutlineDisclosure(r *record, id string, b *core.Switch, expanded bool) {
	guard := listBoxActionGuard(r, id, b)
	source := r.node.Outline
	postControl(r, func() {
		if guard() && source != nil && source.SameSource(r.node.Outline) && r.node.Outline.Disclosure != nil {
			r.node.Outline.Disclosure(ui.OutlineDisclosureEvent{ID: r.node.ID, ItemID: id, Expanded: expanded})
			requestControlRestore(r)
		}
	})
}
func queueOutlineOpen(r *record, id string, b *core.Switch) {
	guard := listBoxActionGuard(r, id, b)
	source := r.node.Outline
	row := outlineViewRow(r, id)
	if row.HasChildren {
		queueOutlineDisclosure(r, id, b, !row.Expanded)
		return
	}
	postControl(r, func() {
		if guard() && source != nil && source.SameSource(r.node.Outline) && r.node.Outline.Activate != nil {
			r.node.Outline.Activate(ui.OutlineEvent{ID: r.node.ID, ItemID: id})
			requestControlRestore(r)
		}
	})
}
func makeOutlineViewRow(r *record, id string, b *core.Switch, p *tree.Plan) bool {
	if r.node.Outline == nil {
		return false
	}
	p.Children = nil
	tree.AddAt(p, "outline-indent", func(w *core.Space) {
		w.Styler(func(s *styles.Style) {
			s.Min.X = units.Dp(float32(outlineViewRow(r, id).Depth) * 16)
			s.Max.X = s.Min.X
			s.Grow.Set(0, 0)
		})
	})
	tree.AddAt(p, "outline-disclosure-slot", func(slot *core.Frame) {
		slot.Styler(func(s *styles.Style) {
			s.Min.X = units.Dp(18)
			s.Max.X = units.Dp(18)
			s.Padding.Zero()
			s.Margin.Zero()
			s.Gap.Zero()
			s.Grow.Set(0, 0)
			s.RenderBox = false
		})
		slot.Maker(func(plan *tree.Plan) {
			if !outlineViewRow(r, id).HasChildren {
				return
			}
			tree.AddAt(plan, "disclosure", func(w *core.Button) {
				w.SetType(core.ButtonAction)
				w.Styler(func(s *styles.Style) {
					s.SetAbilities(false, abilities.Focusable)
					s.Padding.Zero()
					s.Margin.Zero()
					s.Border.Width.Zero()
					s.MaxBorder.Width.Zero()
					s.IconSize.Set(units.Dp(12))
					s.Min.Zero()
					s.Grow.Set(0, 0)
				})
				w.Updater(func() {
					row := outlineViewRow(r, id)
					symbol := ui.SymbolRight
					if row.Expanded {
						symbol = ui.SymbolDown
					}
					w.SetIcon(nativeSymbol(symbol))
					w.SetTooltip(fmt.Sprint("Toggle ", row.Label))
				})
				w.OnFirst(events.Click, func(e events.Event) {
					e.SetHandled()
					row := outlineViewRow(r, id)
					queueOutlineDisclosure(r, id, b, !row.Expanded)
				})
			})
		})
	})
	row := outlineViewRow(r, id)
	if row.Image != nil {
		tree.AddAt(p, "outline-image", func(w *managedImage) {
			w.Styler(func(s *styles.Style) {
				s.RenderBox = false
				s.Min.Set(units.Dp(18), units.Dp(18))
				s.Max.Set(units.Dp(18), units.Dp(18))
				s.Padding.Zero()
				s.Margin.Zero()
				s.Grow.Set(0, 0)
			})
			w.Updater(func() { w.source = outlineViewRow(r, id).Image; w.fit = ui.ImageContain })
		})
	} else if row.Icon != "" {
		tree.AddAt(p, "outline-icon", func(w *core.Icon) {
			w.Styler(func(s *styles.Style) { s.IconSize.Set(units.Dp(18)); s.RenderBox = false })
			w.Updater(func() { w.SetIcon(nativeSymbol(outlineViewRow(r, id).Icon)) })
		})
	}
	tree.AddAt(p, "outline-label", func(w *core.Text) {
		w.Styler(func(s *styles.Style) { s.RenderBox = false; s.SetNonSelectable(); s.Grow.X = 1; s.SetTextWrap(false) })
		w.Updater(func() { w.SetText(html.EscapeString(outlineViewRow(r, id).Label)) })
	})
	if row.Badge != "" {
		tree.AddAt(p, "outline-badge", func(w *core.Text) {
			w.Styler(func(s *styles.Style) { s.RenderBox = false; s.Font.Size.Dp(11) })
			w.Updater(func() { w.SetText(html.EscapeString(outlineViewRow(r, id).Badge)) })
		})
	}
	return true
}
func bindOutlineChoice(r *record, id string, b *core.Switch) {
	b.FinalStyler(func(s *styles.Style) {
		if r.node.Outline != nil {
			s.Gap.Set(units.Dp(5))
			s.SetAbilities(true, abilities.DoubleClickable)
		}
	})
	b.OnFirst(events.DoubleClick, func(e events.Event) {
		if r.node.Outline != nil && e.Modifiers() == 0 {
			e.SetHandled()
			queueOutlineOpen(r, id, b)
		}
	})
}
func outlineViewKey(r *record, id string, b *core.Switch, e events.Event) bool {
	if r.node.Outline == nil || !listBoxActionGuard(r, id, b)() || e.Modifiers() != 0 {
		return false
	}
	row := outlineViewRow(r, id)
	if e.KeyCode() == key.CodeReturnEnter {
		e.SetHandled()
		queueOutlineOpen(r, id, b)
		return true
	}
	if e.KeyCode() != key.CodeLeftArrow && e.KeyCode() != key.CodeRightArrow {
		return false
	}
	e.SetHandled()
	target := ""
	if e.KeyCode() == key.CodeRightArrow {
		if row.HasChildren && !row.Expanded {
			queueOutlineDisclosure(r, id, b, true)
			return true
		}
		for _, next := range r.node.Outline.Rows {
			if next.Parent == id && !next.Disabled {
				target = next.ID
				break
			}
		}
	} else {
		if row.HasChildren && row.Expanded {
			queueOutlineDisclosure(r, id, b, false)
			return true
		}
		target = row.Parent
	}
	if next := r.control.choice.buttons[target]; next != nil && !next.IsDisabled() {
		next.SetFocus()
		next.ScrollToThis()
		queueListSelection(r, target, next, 0, id)
	}
	return true
}
