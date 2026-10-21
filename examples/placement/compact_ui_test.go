package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestCompactMenusPreserveCommandsAndRejectUnknownDisabledActions(t *testing.T) {
	m := groupedModel(t)
	invoked := ""
	row := m.compactActions(func(id string) func() { return func() { invoked = id } })
	if err := ui.Validate(row); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, menu := range row.Children {
		for _, o := range menu.Options {
			if seen[o.ID] {
				t.Fatal("duplicate", o.ID)
			}
			seen[o.ID] = true
		}
		menu.Change(ui.ChangeEvent{SelectedID: "not-an-action"})
		if invoked != "" {
			t.Fatal("unknown command invoked")
		}
	}
	for _, id := range []string{"cut", "copy", "paste", "duplicate", "delete", "move-left", "move-right", "move-up", "move-down", "width-less", "width-more", "height-less", "height-more", "group", "ungroup", "bring-front", "send-back", "bring-forward", "send-backward"} {
		if !seen[id] {
			t.Fatal("missing command", id)
		}
	}
	for _, a := range arrangementActions {
		if !seen[a.id] {
			t.Fatal("missing arrangement", a.id)
		}
	}
	menu := m.actionMenu("actions", "Actions", []ui.Option{{ID: "delete", Label: "Delete"}, {ID: "width-more", Label: "Width"}}, func(id string) func() { return func() { invoked = id } })
	menu.Change(ui.ChangeEvent{SelectedID: "width-more"})
	if invoked != "" {
		t.Fatal("disabled multi-resize action invoked")
	}
	menu.Change(ui.ChangeEvent{SelectedID: "delete"})
	if invoked != "delete" {
		t.Fatal("enabled action unavailable")
	}
}
