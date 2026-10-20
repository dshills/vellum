package main

import (
	"context"
	"maps"
	"math"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
)

func groupedModel(t *testing.T) *model {
	t.Helper()
	m := editModel(t)
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	if err := m.changeGrouping(false); err != nil {
		t.Fatal(err)
	}
	return m
}
func TestGroupingMetadataHistorySelectionAndBranch(t *testing.T) {
	m := editModel(t)
	m.setSelection([]string{"block-1", "note-3"}, "block-1")
	_ = m.journal.MarkSaved()
	before := slices.Clone(m.scene.Objects)
	if err := m.changeGrouping(false); err != nil {
		t.Fatal(err)
	}
	groups := maps.Clone(m.groups)
	if !slices.Equal(m.scene.Objects, before) || m.groups["block-1"] != "group-1" || m.groups["note-3"] != "group-1" || m.nextGroup != 1 || m.journal.Snapshot().UndoCount != 4 || !m.journal.Snapshot().Dirty || m.scene.Selected != "block-1" {
		t.Fatal("group changed geometry or failed metadata/history")
	}
	if m.changeGrouping(false) == nil || m.journal.Snapshot().UndoCount != 4 {
		t.Fatal("grouped single unit regrouped")
	}
	_ = m.journal.Undo()
	if len(m.groups) != 0 || m.journal.Snapshot().Dirty || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("Group Undo")
	}
	_ = m.journal.Redo()
	if !maps.Equal(m.groups, groups) || m.nextGroup != 1 {
		t.Fatal("Group Redo identity")
	}
	m.managedSelection("note-3")
	if len(m.selectedIDs()) != 2 {
		t.Fatal("member click not whole group")
	}
	m.managedSelection("note-3")
	if len(m.selectedIDs()) != 2 {
		t.Fatal("same primary broke group")
	}
	if err := m.changeGrouping(true); err != nil || len(m.groups) != 0 || !slices.Equal(m.scene.Objects, before) || len(m.selectedIDs()) != 2 {
		t.Fatal("Ungroup geometry/selection", err)
	}
	_ = m.journal.Undo()
	if !maps.Equal(m.groups, groups) {
		t.Fatal("Ungroup Undo")
	}
	_ = m.journal.Redo()
	if len(m.groups) != 0 {
		t.Fatal("Ungroup Redo")
	}
	_ = m.journal.Undo()
	_ = m.journal.Undo()
	if err := m.changeGrouping(false); err != nil || m.groups["block-1"] != "group-2" || m.journal.Snapshot().RedoCount != 0 {
		t.Fatal("group branch reused ID", err)
	}
}

func TestGroupToggleMarqueeMoveResizeAndDelete(t *testing.T) {
	m := groupedModel(t)
	groups := maps.Clone(m.groups)
	g := groupGeometry()
	toggleSelection(t, m, g, canvas.Point{X: 40, Y: 25})
	if len(m.selectedIDs()) != 0 {
		t.Fatal("Shift toggle did not remove whole group")
	}
	toggleSelection(t, m, g, canvas.Point{X: 40, Y: 25})
	if len(m.selectedIDs()) != 2 {
		t.Fatal("Shift toggle did not add whole group")
	}
	m.setSelection(nil, "")
	p := selectionPoint(g, canvas.Point{X: -10, Y: -10})
	s := m.selectionToggle(g, p)
	_ = m.finishMove(s, selectionPoint(g, canvas.Point{X: 10, Y: 10}))
	if !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) {
		t.Fatal("partial marquee did not expand group")
	}
	before := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	m.snap = false
	p = selectionPoint(g, canvas.Point{X: 40, Y: 25})
	s = m.beginMove(g, p)
	_ = m.finishMove(s, canvas.Point{X: p.X + 30, Y: p.Y + 30})
	if !maps.Equal(groups, m.groups) || m.scene.Objects[2].X == before[2].X {
		t.Fatal("group move split membership")
	}
	_ = m.journal.Undo()
	if !slices.Equal(m.scene.Objects, before) {
		t.Fatal("group move Undo")
	}
	b, _ := selectionBounds(before, m.selectedIDs())
	p = cornerHandles(b, g)[2].center
	s = m.beginMove(g, p)
	_ = m.finishMove(s, canvas.Point{X: p.X + 60, Y: p.Y + 30})
	if !maps.Equal(groups, m.groups) || m.scene.Objects[2].Width == before[2].Width {
		t.Fatal("group resize failed")
	}
	_ = m.journal.Undo()
	if err := m.deleteSelected(); err != nil || len(m.groups) != 0 || len(m.scene.Objects) != 1 {
		t.Fatal("group delete prune", err)
	}
	_ = m.journal.Undo()
	if !maps.Equal(groups, m.groups) || !slices.Equal(m.scene.Objects, before) || len(m.selectedIDs()) != 2 || m.journal.Snapshot().UndoCount != count {
		t.Fatal("group delete Undo")
	}
	if err := m.palette.Source("block").PlaceKeyboard(context.Background(), m.target, canvas.Point{X: 700}); err != nil {
		t.Fatal(err)
	}
	_ = m.journal.Undo()
	if !maps.Equal(groups, m.groups) || len(m.selectedIDs()) != 2 {
		t.Fatal("placement Undo lost group")
	}
}

func TestGroupedCopyPasteAndMergeIndependentIdentities(t *testing.T) {
	m := groupedModel(t)
	text := copiedText(t, m)
	original := maps.Clone(m.groups)
	if err := m.duplicateSelected(); err != nil {
		t.Fatal(err)
	}
	if m.groups["block-4"] != "group-2" || m.groups["note-5"] != "group-2" || m.groups["block-1"] != "group-1" {
		t.Fatal("duplicate merged with source")
	}
	_ = m.journal.Undo()
	if !maps.Equal(m.groups, original) {
		t.Fatal("group duplicate Undo")
	}
	_ = m.journal.Redo()
	if m.groups["block-4"] != "group-2" {
		t.Fatal("group duplicate Redo")
	}
	if err := m.pasteSelection(text); err != nil || m.groups["block-6"] != "group-3" || m.groups["note-7"] != "group-3" {
		t.Fatal("Paste group remap", err)
	}
	m.setSelection([]string{"block-1", "block-4"}, "block-4")
	if err := m.changeGrouping(false); err != nil {
		t.Fatal(err)
	}
	if m.groups["block-1"] != "group-4" || m.groups["note-5"] != "group-4" || m.groups["block-6"] != "group-3" {
		t.Fatal("flat merge touched unrelated group")
	}
	_ = m.journal.Undo()
	if m.groups["block-1"] != "group-1" || m.groups["block-4"] != "group-2" {
		t.Fatal("merge Undo")
	}
	objects, _ := m.copiedSelection()
	objects[0].Group = 512
	objects[1].Group = 0
	before, next, ng := slices.Clone(m.scene.Objects), m.next, m.nextGroup
	if m.insertCopies(objects, 20, "Paste") == nil || m.next != next || m.nextGroup != ng || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("incomplete clipboard groups partially inserted")
	}
}

func TestGroupedArrangementTreatsUnitsRigidly(t *testing.T) {
	m := groupedModel(t)
	if m.commands.Enabled("align-left") == nil {
		t.Fatal("one group enabled alignment")
	}
	m.setSelection([]string{"block-1", "connector-2"}, "connector-2")
	before := slices.Clone(m.scene.Objects)
	if err := m.arrangeSelected("align-right"); err != nil {
		t.Fatal(err)
	}
	after := m.scene.Objects
	if after[2].X-after[0].X != before[2].X-before[0].X || after[0].Width != before[0].Width || after[1].X != 370 {
		t.Fatal("alignment split group", after)
	}
	_ = m.journal.Undo()
	if !slices.Equal(m.scene.Objects, before) {
		t.Fatal("arrangement group Undo")
	}
}

func TestGroupedDrawingRoundTripAndValidation(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200, Y: 100})
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		if err := f.m.changeGrouping(false); err != nil {
			t.Fatal(err)
		}
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if err := f.m.changeGrouping(true); err != nil {
			t.Fatal(err)
		}
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if f.m.groups["block-1"] != "group-1" || f.m.nextGroup != 1 || len(f.m.selectedIDs()) != 2 || f.m.journal.Snapshot().UndoCount != 0 {
			t.Fatal("group file roundtrip")
		}
	})
	// Persist the allocator even after the final group has been removed.
	f.owner(t, func() {
		if err := f.m.changeGrouping(true); err != nil {
			t.Fatal(err)
		}
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if err := f.m.changeGrouping(false); err != nil {
			t.Fatal(err)
		}
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if len(f.m.groups) != 0 || f.m.nextGroup != 1 {
			t.Fatal("ungrouped roundtrip lost allocator")
		}
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		if err := f.m.changeGrouping(false); err != nil || f.m.groups["block-1"] != "group-2" {
			t.Fatal("group ID reused after ungrouped Open", err)
		}
	})
	m := groupedModel(t)
	saved, err := drawingCodec(m).Capture()
	if err != nil || saved.Version != 3 {
		t.Fatal("group version", err)
	}
	for _, bad := range []string{"oldversion", "missing", "duplicate", "singleton", "counter", "groupID", "overlap"} {
		f := saved
		f.Groups = make([]drawingGroup, len(saved.Groups))
		for i, g := range saved.Groups {
			f.Groups[i] = drawingGroup{ID: g.ID, Members: slices.Clone(g.Members)}
		}
		switch bad {
		case "oldversion":
			f.Version = 2
		case "missing":
			f.Groups[0].Members[0] = "missing-1"
		case "duplicate":
			f.Groups[0].Members[1] = f.Groups[0].Members[0]
		case "singleton":
			f.Groups[0].Members = f.Groups[0].Members[:1]
		case "counter":
			f.NextGroup = 0
		case "groupID":
			f.Groups[0].ID = "group-01"
		case "overlap":
			f.NextGroup = 2
			f.Groups = append(f.Groups, drawingGroup{ID: "group-2", Members: []string{"block-1", "connector-2"}})
		}
		before := maps.Clone(m.groups)
		if drawingCodec(m).Install(f) == nil || !maps.Equal(m.groups, before) {
			t.Fatal("invalid grouping installed", bad)
		}
	}
	m.nextGroup = math.MaxUint64
	if _, err := drawingCodec(m).Capture(); err != nil {
		t.Fatal("exhausted counter not saveable", err)
	}
	m.setSelection([]string{"block-1", "connector-2"}, "connector-2")
	if m.changeGrouping(false) == nil {
		t.Fatal("exhausted group IDs allocated")
	}
}
