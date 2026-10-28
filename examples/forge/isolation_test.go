package main

import (
	"reflect"
	"testing"
)

func TestIsolationPreservesDocumentAndHistory(t *testing.T) {
	m := newWorkbench()
	_ = m.visibility("shaft", false)
	m.selectObject("carriage")
	before := m.captureStudy()
	history := m.journal.Snapshot()
	if e := m.isolateSelection(); e != nil {
		t.Fatal(e)
	}
	for _, o := range m.scene.Objects {
		if o.Visible && o.ID != "carriage" {
			t.Fatal("not isolated")
		}
	}
	if !reflect.DeepEqual(before, m.captureStudy()) || history != m.journal.Snapshot() {
		t.Fatal("isolation dirtied document")
	}
	m.restoreVisibility()
	for _, o := range m.scene.Objects {
		if o.ID == "shaft" && o.Visible {
			t.Fatal("lost prior hidden state")
		}
		if o.ID == "base" && !o.Visible {
			t.Fatal("failed restore")
		}
	}
}
func TestIsolationRemovalAndOpenCleanup(t *testing.T) {
	m := newWorkbench()
	_ = m.isolateSelection()
	if e := m.deleteSelected(); e != nil {
		t.Fatal(e)
	}
	if len(m.isolatedIDs) != 0 {
		t.Fatal("dead isolation references")
	}
	if e := m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	_ = m.isolateSelection()
	s := m.captureStudy()
	if e := m.installStudy(s); e != nil {
		t.Fatal(e)
	}
	if len(m.isolatedIDs) != 0 {
		t.Fatal("isolation survived open")
	}
}
