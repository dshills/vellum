package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"slices"
)

func (m *model) selectionAvailable() error {
	if m.recovery.pending() {
		return document.ErrBusy
	}
	return m.clipboardAvailable()
}
func (m *model) bindSelectionCommands() error {
	for _, c := range []struct{ id, label, shortcut string }{
		{"select-all", "Select All", "Meta+A"},
		{"deselect-all", "Deselect All", "Meta+Shift+A"},
		{"invert-selection", "Invert Selection", ""},
	} {
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Shortcut: c.shortcut, ShortcutScope: command.AfterFocusShortcuts, Available: m.selectionAvailable, Execute: func(context.Context) error { return m.selectCommand(c.id) }}); err != nil {
			return err
		}
	}
	return nil
}

// Selection commands never change geometry or history. Canonical expansion keeps
// flat groups atomic and filters every hidden/locked unit, in paint order.
func (m *model) selectCommand(action string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	prior := m.selectedIDs()
	ids := make([]string, 0, len(m.scene.Objects))
	for _, o := range m.scene.Objects {
		if action == "select-all" || (action == "invert-selection" && !slices.Contains(prior, o.ID)) {
			ids = append(ids, o.ID)
		}
	}
	m.setSelection(ids, m.scene.Selected)
	m.status = fmt.Sprintf("Selection: %d components", len(m.selectedIDs()))
	return nil
}
