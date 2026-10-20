package main

import (
	"context"
	"slices"

	"github.com/dshills/vellum/command"
)

func (m *model) bindStacking() error {
	for _, action := range []struct {
		id, label   string
		front, step bool
	}{
		{"bring-front", "Bring to front", true, false},
		{"send-back", "Send to back", false, false},
		{"bring-forward", "Bring forward", true, true},
		{"send-backward", "Send backward", false, true},
	} {
		if err := m.commands.Register(command.Command{ID: action.id, Label: action.label,
			Available: m.editAvailable,
			Execute: func(context.Context) error {
				if action.step {
					return m.stepSelected(action.front)
				}
				return m.stackSelected(action.front)
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *model) stackSelected(front bool) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	ids := m.selectedIDs() // Includes every member of each selected flat group.
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	// Stable partition of the existing back-to-front paint order. Neither the
	// selected set nor the remaining objects change their own relative order.
	after := slices.Clone(m.scene.Objects)
	i := 0
	for _, wanted := range []bool{!front, front} {
		for _, o := range m.scene.Objects {
			if selected[o.ID] == wanted {
				after[i] = o
				i++
			}
		}
	}
	label := "Send to back"
	if front {
		label = "Bring to front"
	}
	if err := m.applyDocument(label, after, m.groups, ids, m.scene.Selected); err != nil {
		return err
	}
	m.status = label
	return nil
}

// Flat groups expand selection, but do not create synthetic paint layers. Each
// selected run crosses one adjacent unselected component. Traversing against the
// movement direction ensures no selected component can advance twice per call.
func (m *model) stepSelected(forward bool) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	ids := m.selectedIDs()
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	after := slices.Clone(m.scene.Objects)
	label := "Send backward"
	if forward {
		label = "Bring forward"
		for i := len(after) - 2; i >= 0; i-- {
			if selected[after[i].ID] && !selected[after[i+1].ID] {
				after[i], after[i+1] = after[i+1], after[i]
			}
		}
	} else {
		for i := 1; i < len(after); i++ {
			if selected[after[i].ID] && !selected[after[i-1].ID] {
				after[i], after[i-1] = after[i-1], after[i]
			}
		}
	}
	if err := m.applyDocument(label, after, m.groups, ids, m.scene.Selected); err != nil {
		return err
	}
	m.status = label
	return nil
}
