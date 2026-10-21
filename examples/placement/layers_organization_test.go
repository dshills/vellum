package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestLayerFilterCollapseProjectionKeepsDocumentSelectionHistory(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"group-1": {Name: "Assembly"}, "block-1": {Name: "Housing"}, "note-3": {Name: "Label"}, "connector-2": {Name: "Cable"}}
	before := slices.Clone(m.scene.Objects)
	groups := maps.Clone(m.groups)
	selection := m.selectedIDs()
	state := m.journal.Snapshot()
	if err := m.setGroupRows("collapse-all"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(layerKeys(m.layerRows()), []string{"group-1", "connector-2"}) {
		t.Fatal("collapsed projection", layerKeys(m.layerRows()))
	}
	_ = m.setLayerFilter("  hOuSiNg  ")
	if !slices.Equal(layerKeys(m.layerRows()), []string{"group-1", "block-1"}) {
		t.Fatal("member match context", layerKeys(m.layerRows()))
	}
	_ = m.setLayerFilter("assembly")
	if !slices.Equal(layerKeys(m.layerRows()), []string{"group-1", "note-3", "block-1"}) {
		t.Fatal("group match omitted members", layerKeys(m.layerRows()))
	}
	_ = m.setLayerFilter("CONNECTOR-2")
	if !slices.Equal(layerKeys(m.layerRows()), []string{"connector-2"}) {
		t.Fatal("stable ID query")
	}
	_ = m.setLayerFilter("nothing")
	if len(m.layerRows()) != 0 || m.emptyLayersLabel() != "No matching layers" {
		t.Fatal("empty match")
	}
	_ = m.setLayerFilter("")
	if len(m.layerRows()) != 2 || !m.collapsedGroups["group-1"] {
		t.Fatal("filter lost collapse state")
	}
	if !slices.Equal(before, m.scene.Objects) || !maps.Equal(groups, m.groups) || !slices.Equal(selection, m.selectedIDs()) || state != m.journal.Snapshot() {
		t.Fatal("projection changed document")
	}
	if err := m.setGroupRows("toggle-group-1"); err != nil || len(m.layerRows()) != 4 {
		t.Fatal("expand group", err)
	}
	m.collapsedGroups["stale-group"] = true
	_ = m.setGroupRows("collapse-all")
	if len(m.collapsedGroups) != 1 {
		t.Fatal("collapse state unbounded")
	}
	if m.setLayerFilter(strings.Repeat("界", 129)) == nil || m.layerFilter != "" {
		t.Fatal("unbounded filter")
	}
	if m.setGroupRows("toggle-missing") == nil {
		t.Fatal("stale group admitted")
	}
}
func TestFilteredLayerClickWorksDragRefusedAndControlsGuardCapture(t *testing.T) {
	m := groupedModel(t)
	_ = m.setLayerFilter("connector")
	g := testLayerLayout(m, 1)
	p := layerPoint(g, "connector-2", true)
	if m.beginLayerDrag("connector-2", p, g) != nil {
		t.Fatal("filtered drag admitted")
	}
	if err := m.selectLayer("connector-2"); err != nil || m.scene.Selected != "connector-2" {
		t.Fatal("filtered click refused", err)
	}
	_ = m.setLayerFilter("")
	g = testLayerLayout(m, 1)
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil {
		t.Fatal("unfiltered drag refused")
	}
	if m.setLayerFilter("block") == nil || m.setGroupRows("collapse-all") == nil {
		t.Fatal("projection changed during capture")
	}
	m.cancelLayerDrag(s)
}
func TestCollapsedGroupReorderUsesWholeGroupAndUndo(t *testing.T) {
	m := groupedModel(t)
	_ = m.setGroupRows("collapse-all")
	before := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	m.setSelection([]string{"connector-2"}, "connector-2")
	g := testLayerLayout(m, 1)
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	p := layerPoint(g, "group-1", false)
	if s == nil || !m.updateLayerDrag(s, p, g) || !s.valid {
		t.Fatal("collapsed group below target rejected", s)
	}
	if err := m.finishLayerDrag(s, p, g); err != nil {
		t.Fatal(err)
	}
	if m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("reorder not one edit")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("reorder undo", err)
	}
	if err := m.journal.Redo(); err != nil {
		t.Fatal("reorder redo", err)
	}
}
func TestLayerOrganizationControlsAndDocumentReplacement(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		root := view(f.a, f.ref, f.m)
		field := geometryDialogControl(t, root, "layers-filter")
		field.Change(ui.ChangeEvent{Text: "block"})
		if f.m.layerFilter != "block" {
			t.Fatal("filter callback")
		}
		geometryDialogControl(t, view(f.a, f.ref, f.m), "layers-filter-clear").Activate()
		if f.m.layerFilter != "" {
			t.Fatal("clear callback")
		}
		f.m.layerFilter = "old"
		f.m.collapsedGroups = map[string]bool{"group-old": true}
		saved, _ := drawingCodec(f.m).Capture()
		if err := drawingCodec(f.m).Install(saved); err != nil || f.m.layerFilter != "" || len(f.m.collapsedGroups) != 0 {
			t.Fatal("replacement retained projection", err)
		}
	})
}
