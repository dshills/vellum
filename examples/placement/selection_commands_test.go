package main

import (
	"context"
	"github.com/dshills/vellum/command"
	"slices"
	"testing"
)

func TestSelectionCommandsAtomicUnitsFiltersAndHistory(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"connector-2": {Locked: true}}
	before := slices.Clone(m.scene.Objects)
	state := m.journal.Snapshot()
	if err := m.commands.Invoke(context.Background(), "select-all"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) {
		t.Fatal("filter/groups", m.selectedIDs())
	}
	primary := m.scene.Selected
	if err := m.selectCommand("invert-selection"); err != nil {
		t.Fatal(err)
	}
	if len(m.selectedIDs()) != 0 {
		t.Fatal("invert locked selection")
	}
	m.layerInfo = nil
	if err := m.selectCommand("invert-selection"); err != nil {
		t.Fatal(err)
	}
	if len(m.selectedIDs()) != 3 {
		t.Fatal("invert empty")
	}
	m.setSelection([]string{"block-1"}, primary)
	if err := m.selectCommand("invert-selection"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) {
		t.Fatal("invert split group", m.selectedIDs())
	}
	if err := m.selectCommand("deselect-all"); err != nil {
		t.Fatal(err)
	}
	if len(m.selectedIDs()) != 0 || m.scene.Selected != "" {
		t.Fatal("deselect")
	}
	m.layerInfo = map[string]layerMetadata{"note-3": {Hidden: true}}
	_ = m.selectCommand("select-all")
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) {
		t.Fatal("imported member hide did not exclude whole group")
	}
	if state != m.journal.Snapshot() || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("selection dirtied geometry/history")
	}
	for _, c := range m.commands.Search("") {
		if c.ID == "select-all" && c.ShortcutScope != command.AfterFocusShortcuts {
			t.Fatal("text fields cannot select text first")
		}
	}
}
func TestSelectionCommandBusyAdmission(t *testing.T) {
	m := groupedModel(t)
	before := m.selectedIDs()
	for _, kind := range []string{"move", "prompt", "layers"} {
		switch kind {
		case "move":
			m.move = &objectMove{}
		case "prompt":
			m.prompt = "geometry"
		case "layers":
			m.layerDrag = &layerDrag{}
		}
		for _, id := range []string{"select-all", "deselect-all", "invert-selection"} {
			if m.selectCommand(id) == nil {
				t.Fatal("busy admission", kind, id)
			}
		}
		m.move = nil
		m.prompt = ""
		m.layerDrag = nil
	}
	if !slices.Equal(before, m.selectedIDs()) {
		t.Fatal("refused selection changed")
	}
}
