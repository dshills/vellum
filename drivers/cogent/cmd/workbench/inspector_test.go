package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"testing"
)

func TestInspectorPropertyValidationHistoryAndLatestSelection(t *testing.T) {
	parts := newParts(4)
	table, err := virtual.NewTable(parts, 2)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := history.New(8)
	if err != nil {
		t.Fatal(err)
	}
	board, err := canvas.New([]canvas.Rect{{ID: parts.Row(0).ID, Width: 50, Height: 25}})
	if err != nil {
		t.Fatal(err)
	}
	edit := &partEditor{source: parts, table: table, journal: journal, board: board}
	selected := 0
	open := true
	view := inspectorView(parts, edit, open, func(b bool) { open = b }, func() int { return selected }, func(string) {})
	if err := ui.Validate(view); err != nil {
		t.Fatal(err)
	}
	var field, rotation, pack, disclosure ui.Node
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		switch n.ID {
		case "inspector.value":
			field = n
		case "inspector.rotation":
			rotation = n
		case "inspector.package":
			pack = n
		case "inspector.properties":
			disclosure = n
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(view)
	before := parts.parts[0]
	revision := parts.Revision()
	for _, bad := range []string{"", "-1", "4294967296", "1.5"} {
		field.Change(ui.ChangeEvent{Text: bad})
	}
	if parts.parts[0] != before || parts.Revision() != revision || journal.Snapshot().UndoCount != 0 {
		t.Fatal("invalid property changed accepted document/history")
	}
	field.Change(ui.ChangeEvent{Text: "4294967295"})
	if parts.parts[0].Value != ^uint32(0) {
		t.Fatal("fulluint32 value lost")
	}
	if err := journal.Undo(); err != nil {
		t.Fatal(err)
	}
	if parts.parts[0] != before {
		t.Fatal("value undo mismatch")
	}
	if err := journal.Redo(); err != nil {
		t.Fatal(err)
	}
	// A descriptor callback always uses the owner’s current selection.
	selected = 1
	field.Change(ui.ChangeEvent{Text: "765"})
	if parts.parts[1].Value != 765 || parts.parts[0].Value != ^uint32(0) {
		t.Fatal("property request used stale selected index")
	}
	pack.Change(ui.ChangeEvent{SelectedID: "0"})
	if parts.parts[1].Package != 0 {
		t.Fatal("package request lost")
	}
	rotation.Change(ui.ChangeEvent{SelectedID: "180"})
	if parts.parts[1].Rotation != 180 {
		t.Fatal("rotation request lost")
	}
	rev := parts.Revision()
	rotation.Change(ui.ChangeEvent{SelectedID: "45"})
	pack.Change(ui.ChangeEvent{SelectedID: "2"})
	if parts.Revision() != rev {
		t.Fatal("unknown choice committed")
	}
	disclosure.Change(ui.ChangeEvent{Checked: false})
	if open {
		t.Fatal("expansion request lost")
	}
}
