package main

import (
	"github.com/dshills/vellum/ui"
	"reflect"
	"strings"
	"testing"
)

func TestRenameHistoryAndPersistence(t *testing.T) {
	m := newWorkbench()
	before := clone(m.state)
	id := m.state.Selected
	if e := m.renameBody(id, "Precision carriage"); e != nil {
		t.Fatal(e)
	}
	if m.selected().Name != "Precision carriage" || m.selected().ID != id {
		t.Fatal("rename")
	}
	b, e := encodeStudy(m.captureStudy())
	if e != nil {
		t.Fatal(e)
	}
	s, e := decodeStudy(b)
	if e != nil {
		t.Fatal(e)
	}
	n := newWorkbench()
	_ = n.installStudy(s)
	if n.selected().Name != "Precision carriage" {
		t.Fatal("persistence")
	}
	if e = m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, m.state) {
		t.Fatal("undo")
	}
	if e = m.journal.Redo(); e != nil {
		t.Fatal(e)
	}
	if m.selected().Name != "Precision carriage" {
		t.Fatal("redo")
	}
}
func TestRenameValidationAndStaleDialog(t *testing.T) {
	m := newWorkbench()
	for _, name := range []string{" ", "bad\nname", strings.Repeat("a", 129)} {
		if m.renameBody(m.state.Selected, name) == nil {
			t.Fatal("invalid name")
		}
	}
	m.beginRename("shaft")
	old := m.renamePrompt()
	m.renameID = ""
	m.beginRename("cap")
	var apply func()
	for _, n := range old.Children {
		if n.ID == "rename-body-apply" {
			apply = n.Activate
		}
	}
	m.renameText = "New cap"
	apply()
	if m.selected().Name == "New cap" {
		t.Fatal("stale rename")
	}
	if e := ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
}
