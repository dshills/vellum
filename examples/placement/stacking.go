package main

import (
	"context"
	"slices"

	"github.com/dshills/vellum/command"
)

func (m *model) bindStacking() error {
	for _, action := range []struct {
		id, label string
		front     bool
	}{
		{"bring-front", "Bring to front", true},
		{"send-back", "Send to back", false},
	} {
		if err := m.commands.Register(command.Command{ID: action.id, Label: action.label,
			Available: m.editAvailable,
			Execute:   func(context.Context) error { return m.stackSelected(action.front) },
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
