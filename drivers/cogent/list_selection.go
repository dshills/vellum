package cogent

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
	"image/color"
)

func queueListSelection(r *record, id string, b *core.Switch, mod key.Modifiers, anchorHint string) {
	guard := listBoxActionGuard(r, id, b)
	catalog := r.control.choice.listCatalog
	postControl(r, func() {
		if !guard() || r.control.choice.listCatalog != catalog || !r.node.Multiple || r.node.SelectionChange == nil {
			return
		}
		c := r.control.choice
		if c.listAnchor == "" {
			c.listAnchor = anchorHint
		}
		ids := []string{id}
		if mod.HasFlag(key.Shift) {
			first, last := -1, -1
			for i, o := range r.node.Options {
				if o.ID == c.listAnchor {
					first = i
				}
				if o.ID == id {
					last = i
				}
			}
			if first < 0 {
				first = last
			}
			if first > last {
				first, last = last, first
			}
			ids = nil
			for i := first; i >= 0 && i <= last; i++ {
				if !r.node.Options[i].Disabled {
					ids = append(ids, r.node.Options[i].ID)
				}
			}
		} else {
			c.listAnchor = id
		}
		r.node.SelectionChange(ui.ListSelectionEvent{ID: r.node.ID, TargetID: id, IDs: ids, Toggle: mod.HasFlag(key.Meta) && !mod.HasFlag(key.Shift), Add: mod.HasFlag(key.Meta) && mod.HasFlag(key.Shift)})
		requestControlRestore(r)
	})
}
func styleSourceRow(r *record, id string, s *styles.Style) {
	if r.node.Kind != ui.ListBox || !r.node.Multiple {
		return
	}
	applyStyle(s, ui.Node{Kind: ui.Label, Style: ui.Style{Typography: r.theme.Default.Typography, Foreground: r.theme.Default.Foreground}}, ui.Theme{})
	s.Grow.Set(1, 0)
	s.Border.Radius.Set(units.Dp(4))
	s.Padding.Set(units.Dp(4), units.Dp(8))
	s.Min.Y = units.Dp(24)
	if ui.ListOptionSelected(r.node, id) {
		if r.widget.AsWidget().ContainsFocus() {
			s.Background = colors.Uniform(color.NRGBA{R: 0, G: 112, B: 235, A: 255})
			s.Color = colors.Uniform(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		} else {
			s.Background = colors.Uniform(color.NRGBA{R: 160, G: 160, B: 165, A: 100})
		}
	} else {
		s.Background = colors.Uniform(color.NRGBA{})
	}
}

// Select-all uses the same bounded owner queue and rendered eligibility lease as
// range selection, without moving focus or treating Command as a toggle.
func queueListSelectAll(r *record, id string, button *core.Switch) {
	guard := listBoxActionGuard(r, id, button)
	catalog := r.control.choice.listCatalog
	postControl(r, func() {
		if !guard() || r.control.choice.listCatalog != catalog || !r.node.Multiple || r.node.SelectionChange == nil {
			return
		}
		ids := make([]string, 0, len(r.node.Options))
		for _, o := range r.node.Options {
			if !o.Disabled {
				ids = append(ids, o.ID)
			}
		}
		if len(ids) == 0 {
			return
		}
		r.node.SelectionChange(ui.ListSelectionEvent{ID: r.node.ID, TargetID: id, IDs: ids})
		requestControlRestore(r)
	})
}
