package main

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func TestPrimarySelectionKeepsWholeSetGeometryAndHistory(t *testing.T) {
	m := dimensionModel(t)
	before := slices.Clone(m.scene.Objects)
	groups := maps.Clone(m.groups)
	ids := m.selectedIDs()
	state := m.journal.Snapshot()
	if err := m.beginPrimarySelection(); err != nil {
		t.Fatal(err)
	}
	d := m.primaryDraft
	d.choice = "connector-3"
	m.prompt = ""
	if err := m.applyPrimarySelection(d); err != nil {
		t.Fatal(err)
	}
	if m.scene.Selected != "connector-3" || !slices.Equal(ids, m.selectedIDs()) || !slices.Equal(before, m.scene.Objects) || !maps.Equal(groups, m.groups) || state != m.journal.Snapshot() {
		t.Fatal("primary changed document/selection/history")
	}
	m.primaryDraft = nil
	if err := m.beginPrimarySelection(); err != nil {
		t.Fatal(err)
	}
	d = m.primaryDraft
	d.choice = "group-1"
	m.prompt = ""
	if err := m.applyPrimarySelection(d); err != nil {
		t.Fatal(err)
	}
	if m.scene.Selected != "note-2" || m.primaryLayerSuffix("group-1") != " [primary group]" || m.primaryLayerSuffix("note-2") != " [primary]" || !slices.Equal(ids, m.selectedIDs()) || state != m.journal.Snapshot() {
		t.Fatal("group primary or labels")
	}
}
func TestPrimarySelectionRejectsStaleAndUnselectedUnits(t *testing.T) {
	for _, change := range []string{"scene", "selection", "primary", "group", "revision"} {
		t.Run(change, func(t *testing.T) {
			m := dimensionModel(t)
			_ = m.beginPrimarySelection()
			d := m.primaryDraft
			m.prompt = ""
			switch change {
			case "scene":
				m.scene.Objects[0].X++
			case "selection":
				m.setSelection([]string{"connector-3"}, "connector-3")
			case "primary":
				m.setSelection(m.selectedIDs(), "connector-3")
			case "group":
				delete(m.groups, "block-1")
			case "revision":
				if err := m.journal.Reset(false); err != nil {
					t.Fatal(err)
				}
			}
			if !errors.Is(m.applyPrimarySelection(d), history.ErrStale) {
				t.Fatal("stale draft accepted")
			}
		})
	}
	m := dimensionModel(t)
	_ = m.beginPrimarySelection()
	d := m.primaryDraft
	m.prompt = ""
	d.choice = "block-999"
	state := m.journal.Snapshot()
	if m.applyPrimarySelection(d) == nil || state != m.journal.Snapshot() {
		t.Fatal("unselected primary")
	}
	m.primaryDraft = nil
	m.prompt = "close"
	if m.beginPrimarySelection() == nil {
		t.Fatal("busy selection")
	}
}
func TestPrimarySelectionDialogCancelAndRetiredCallbacks(t *testing.T) {
	f := fileModel(t, false)
	f.m.scene.Objects = dimensionModel(t).scene.Objects
	f.m.groups = map[string]string{"block-1": "group-1", "note-2": "group-1"}
	f.m.setSelection([]string{"block-1", "connector-3"}, "block-1")
	f.owner(t, func() {
		if err := f.m.commands.Invoke(f.ref.Context(), "primary-selection"); err != nil {
			t.Fatal(err)
		}
		dialog := f.m.primarySelectionDialog(f.a, f.ref)
		before := f.m.selectedIDs()
		state := f.m.journal.Snapshot()
		geometryDialogControl(t, dialog, "primary-choice").Change(ui.ChangeEvent{SelectedID: "connector-3"})
		geometryDialogControl(t, dialog, "primary-cancel").Activate()
		f.m.prompt = "close"
		dialog.Dismiss()
		geometryDialogControl(t, dialog, "primary-apply").Activate()
		if f.m.scene.Selected != "block-1" || !slices.Equal(before, f.m.selectedIDs()) || f.m.prompt != "close" || state != f.m.journal.Snapshot() {
			t.Fatal("retired callbacks changed selection/modal")
		}
	})
}
