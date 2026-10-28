package main

import (
	"reflect"
	"testing"
)

func TestDuplicateAndPasteHistory(t *testing.T) {
	m := newWorkbench()
	m.selectObject("shaft")
	m.selectModified("cap", true)
	before := clone(m.state)
	objects, e := m.copiedBodies()
	if e != nil {
		t.Fatal(e)
	}
	text, e := encodeBodies(objects)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.duplicateBodies(); e != nil {
		t.Fatal(e)
	}
	if len(m.state.Objects) != 9 || len(m.selectionSet()) != 2 {
		t.Fatal("duplicate count")
	}
	for i, o := range m.state.Objects[7:] {
		if o.ID == objects[i].ID || o.Bounds.Min.Sub(objects[i].Bounds.Min) != vec(.25, 0, .25) {
			t.Fatal("identity or relative placement")
		}
	}
	after := clone(m.state)
	if e = m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(m.state, before) {
		t.Fatal("undo")
	}
	if e = m.journal.Redo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(m.state, after) {
		t.Fatal("redo identity")
	}
	if e = m.completePaste(clone(m.state), m.journal.Snapshot().Revision, text); e != nil {
		t.Fatal(e)
	}
	if len(m.state.Objects) != 11 {
		t.Fatal("paste")
	}
}
func TestPasteRejectsStaleMalformedAndLimits(t *testing.T) {
	m := newWorkbench()
	o, _ := m.copiedBodies()
	text, _ := encodeBodies(o)
	before := clone(m.state)
	rev := m.journal.Snapshot().Revision
	if e := m.add(); e != nil {
		t.Fatal(e)
	}
	snapshot := m.captureStudy()
	if m.completePaste(before, rev, text) == nil {
		t.Fatal("stale paste")
	}
	if _, e := decodeBodies("unrelated text"); e == nil {
		t.Fatal("malformed accepted")
	}
	if !reflect.DeepEqual(snapshot, m.captureStudy()) {
		t.Fatal("rejection mutated model")
	}
}
