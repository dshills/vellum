package main

import (
	"reflect"
	"testing"
)

func TestBatchSelectionMoveDeleteHistory(t *testing.T) {
	m := newWorkbench()
	m.selectObject("shaft")
	m.selectModified("cap", true)
	if len(m.selectionSet()) != 2 {
		t.Fatal("selection")
	}
	before := clone(m.state)
	if e := m.nudge(); e != nil {
		t.Fatal(e)
	}
	if m.journal.Snapshot().UndoCount != 1 {
		t.Fatal("batch history")
	}
	if e := m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, m.state) {
		t.Fatal("move undo")
	}
	if e := m.deleteSelected(); e != nil {
		t.Fatal(e)
	}
	if len(m.state.Objects) != 5 {
		t.Fatal("delete batch")
	}
	if e := m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, m.state) {
		t.Fatal("delete undo")
	}
	if e := m.journal.Redo(); e != nil {
		t.Fatal(e)
	}
	if len(m.state.Objects) != 5 {
		t.Fatal("redo")
	}
}
func TestShiftToggleAndPlainSelection(t *testing.T) {
	m := newWorkbench()
	m.selectObject("shaft")
	m.selectModified("cap", true)
	m.selectModified("shaft", true)
	if len(m.selectionSet()) != 1 || m.state.Selected != "cap" {
		t.Fatal("toggle")
	}
	m.selectObject("base")
	if len(m.selectionSet()) != 1 || m.state.Selected != "base" {
		t.Fatal("plain replace")
	}
}
