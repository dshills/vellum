package main

import "github.com/dshills/vellum/ui"

// A palette capture temporarily blocks execution, but must not reflow the canvas
// underneath its pointer. Keep only availability presentation from the last idle
// view; callbacks and the command registry continue to check live admission.
type controlPresentation struct {
	disabled bool
	options  map[string]bool
}

func (m *model) retainPalettePresentation(root *ui.Node) {
	active := m.palette.State().Active
	if !active {
		m.palettePresentation = make(map[string]controlPresentation)
	}
	var visit func(*ui.Node)
	visit = func(n *ui.Node) {
		if active {
			if before, ok := m.palettePresentation[n.ID]; ok {
				n.Disabled = before.disabled
				for i := range n.Options {
					if disabled, ok := before.options[n.Options[i].ID]; ok {
						n.Options[i].Disabled = disabled
					}
				}
			}
		} else {
			before := controlPresentation{disabled: n.Disabled}
			if len(n.Options) > 0 {
				before.options = make(map[string]bool, len(n.Options))
				for _, option := range n.Options {
					before.options[option.ID] = option.Disabled
				}
			}
			m.palettePresentation[n.ID] = before
		}
		for i := range n.Children {
			visit(&n.Children[i])
		}
	}
	visit(root)
}
