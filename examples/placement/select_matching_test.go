package main

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func TestMatchingSelectionFiltersWholeGroupsAndPrimary(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"group-1": {Name: "Assembly"}, "block-1": {Name: "Housing"}, "connector-2": {Name: "Cable"}}
	state := m.journal.Snapshot()
	before := slices.Clone(m.scene.Objects)
	for _, test := range []struct {
		kind, query string
		want        []string
	}{
		{"block", "", []string{"block-1", "note-3"}},
		{"all", "houSING", []string{"block-1", "note-3"}},
		{"note", "ASSEMBLY", []string{"block-1", "note-3"}},
		{"all", "connector-2", []string{"connector-2"}},
		{"block", "Cable", nil},
	} {
		ids, err := m.matchingSelection(test.kind, test.query, false)
		if err != nil || !slices.Equal(ids, test.want) {
			t.Fatal(test.kind, test.query, ids, err)
		}
	}
	m.setSelection([]string{"connector-2"}, "connector-2")
	ids, err := m.matchingSelection("block", "", true)
	if err != nil {
		t.Fatal(err)
	}
	m.setSelection(ids, m.scene.Selected)
	if len(m.selectedIDs()) != 3 || m.scene.Selected != "connector-2" {
		t.Fatal("add lost selection or primary", m.selectedIDs(), m.scene.Selected)
	}
	m.layerInfo["note-3"] = layerMetadata{Hidden: true}
	ids, err = m.matchingSelection("block", "", false)
	if err != nil || len(ids) != 0 {
		t.Fatal("hidden member did not exclude whole group", ids, err)
	}
	m.layerInfo["note-3"] = layerMetadata{Locked: true}
	ids, err = m.matchingSelection("all", "Assembly", false)
	if err != nil || len(ids) != 0 {
		t.Fatal("locked member did not exclude whole group", ids, err)
	}
	m.layerInfo["note-3"] = layerMetadata{}
	m.layerInfo["group-1"] = layerMetadata{Name: "Assembly", Locked: true}
	ids, err = m.matchingSelection("note", "", false)
	if err != nil || len(ids) != 0 {
		t.Fatal("locked group selected", ids, err)
	}
	if state != m.journal.Snapshot() || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("search modified document/history")
	}
}
func TestMatchingDraftValidationNoMatchAndStale(t *testing.T) {
	for _, test := range []struct{ kind, query string }{{"all", ""}, {"invalid", "x"}, {"all", strings.Repeat("x", 129)}, {"all", string([]byte{0xff})}} {
		if validMatchingQuery(test.kind, test.query) == nil {
			t.Fatal("invalid filter admitted", test.kind)
		}
	}
	m := groupedModel(t)
	state := m.journal.Snapshot()
	if err := m.beginSelectMatching(); err != nil {
		t.Fatal(err)
	}
	d := m.matchingDraft
	d.query = "no-such-component"
	m.prompt = ""
	if err := m.applySelectMatching(d); err != nil || len(m.selectedIDs()) != 0 || m.scene.Selected != "" || state != m.journal.Snapshot() {
		t.Fatal("zero match replace/history", err)
	}
	m.matchingDraft = nil
	if err := m.beginSelectMatching(); err != nil {
		t.Fatal(err)
	}
	d = m.matchingDraft
	d.kind = "connector"
	m.prompt = ""
	m.setSelection([]string{"block-1"}, "block-1")
	if err := m.applySelectMatching(d); !errors.Is(err, history.ErrStale) {
		t.Fatal("changed selection not rejected", err)
	}
	m.matchingDraft = nil
	if err := m.beginSelectMatching(); err != nil {
		t.Fatal(err)
	}
	d = m.matchingDraft
	d.kind = "connector"
	m.prompt = ""
	m.layerInfo = map[string]layerMetadata{"connector-2": {Locked: true}}
	if err := m.applySelectMatching(d); !errors.Is(err, history.ErrStale) {
		t.Fatal("metadata change not rejected", err)
	}
}
func TestMatchingDialogCancelOldCallbacksAndInvalidVisible(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		state := f.m.journal.Snapshot()
		if err := f.m.beginSelectMatching(); err != nil {
			t.Fatal(err)
		}
		old := f.m.selectMatchingDialog(f.a, f.ref)
		geometryDialogControl(t, old, "matching-apply").Activate()
		if f.m.prompt != "select-matching" || f.m.matchingDraft.message == "" {
			t.Fatal("empty filter not shown")
		}
		geometryDialogControl(t, old, "matching-query").Change(ui.ChangeEvent{Text: strings.Repeat("x", 129)})
		geometryDialogControl(t, old, "matching-kind").Change(ui.ChangeEvent{SelectedID: "block"})
		geometryDialogControl(t, old, "matching-apply").Activate()
		if f.m.matchingDraft.message == "" || state != f.m.journal.Snapshot() {
			t.Fatal("oversized filter accepted")
		}
		old.Dismiss()
		if f.m.matchingDraft != nil || f.m.prompt != "" {
			t.Fatal("cancel retained draft")
		}
		if err := f.m.beginSelectMatching(); err != nil {
			t.Fatal(err)
		}
		current := f.m.matchingDraft
		geometryDialogControl(t, old, "matching-query").Change(ui.ChangeEvent{Text: "later"})
		geometryDialogControl(t, old, "matching-apply").Activate()
		old.Dismiss()
		if f.m.matchingDraft != current || current.query != "" || f.m.prompt != "select-matching" || state != f.m.journal.Snapshot() {
			t.Fatal("old callback changed new opening")
		}
		f.m.selectMatchingDialog(f.a, f.ref).Dismiss()
	})
}
