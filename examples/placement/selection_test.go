package main

import (
	"cogentcore.org/core/events/key"
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/placement"
)

func selectionPoint(g placement.Geometry, p canvas.Point) canvas.Point {
	local := g.View.Screen(p)
	return canvas.Point{X: g.Origin.X + local.X*g.Scale, Y: g.Origin.Y + local.Y*g.Scale}
}
func toggleSelection(t *testing.T, m *model, g placement.Geometry, p canvas.Point) {
	t.Helper()
	p = selectionPoint(g, p)
	s := m.selectionToggle(g, p)
	if s == nil || !s.selectionOnly {
		t.Fatal("toggle capture")
	}
	if err := m.finishMove(s, p); err != nil {
		t.Fatal(err)
	}
}
func TestShiftToggleAndClickSelectionRules(t *testing.T) {
	m := editModel(t)
	g := cornerGeometry()
	g.View.Zoom = 1
	g.View.Width = 1200
	g.Clip.Max.X = 2600
	m.setSelection(nil, "")
	count := m.journal.Snapshot().UndoCount
	toggleSelection(t, m, g, canvas.Point{X: 40, Y: 25})
	toggleSelection(t, m, g, canvas.Point{X: 440, Y: 30})
	if !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) || m.scene.Selected != "note-3" {
		t.Fatal("add selection", m.selectedIDs())
	}
	toggleSelection(t, m, g, canvas.Point{X: 440, Y: 30})
	if !slices.Equal(m.selectedIDs(), []string{"block-1"}) {
		t.Fatal("toggle off primary")
	}
	toggleSelection(t, m, g, canvas.Point{X: 440, Y: 30})
	toggleSelection(t, m, g, canvas.Point{X: -150, Y: 100})
	if len(m.selectedIDs()) != 2 || m.journal.Snapshot().UndoCount != count {
		t.Fatal("empty Shift click changed selection/history")
	}
	p := selectionPoint(g, canvas.Point{X: 40, Y: 25})
	s := m.beginMove(g, p)
	if s == nil || len(s.selection) != 2 || s.resize {
		t.Fatal("group press did not preserve members")
	}
	if err := m.finishMove(s, p); err != nil || !slices.Equal(m.selectedIDs(), []string{"block-1"}) {
		t.Fatal("plain click did not replace selection", err)
	}
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	// Explicitly reselecting the current primary must collapse too.
	m.setSelection([]string{"block-1", "note-3"}, "block-1")
	m.managedSelection("block-1")
	if !slices.Equal(m.selectedIDs(), []string{"block-1"}) {
		t.Fatal("same-primary reselection kept group")
	}
	m.scene.Selected = "connector-2"
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) {
		t.Fatal("managed keyboard/AX selection retained stale group")
	}
	m.scene.Selected = ""
	if len(m.selectedIDs()) != 0 {
		t.Fatal("blank click selection not cleared")
	}
}
func TestGroupDragRigidTransformSingleHistoryCancelAndBounds(t *testing.T) {
	m := editModel(t)
	g := cornerGeometry()
	m.snap = true
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	before := slices.Clone(m.scene.Objects)
	next, count := m.next, m.journal.Snapshot().UndoCount
	p := selectionPoint(g, canvas.Point{X: 40, Y: 25})
	end := canvas.Point{X: p.X + 23*g.Scale*g.View.Zoom, Y: p.Y + 17*g.Scale*g.View.Zoom}
	s := m.beginMove(g, p)
	if s == nil {
		t.Fatal("begin group")
	}
	for range 10 {
		m.previewMove(s, end)
	}
	if m.scene.Objects[0].X != 20 || m.scene.Objects[0].Y != 20 || m.scene.Objects[2].X-before[2].X != 20 || m.scene.Objects[2].Y-before[2].Y != 20 || m.scene.Objects[1] != before[1] || m.journal.Snapshot().UndoCount != count {
		t.Fatal("group preview not rigid/atomic")
	}
	if err := m.finishMove(s, end); err != nil {
		t.Fatal(err)
	}
	after := slices.Clone(m.scene.Objects)
	if m.journal.Snapshot().UndoCount != count+1 || m.next != next || m.move != nil {
		t.Fatal("group move not exactly1")
	}
	_ = m.finishMove(s, end)
	if m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("duplicate group commit")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) {
		t.Fatal("group Undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
		t.Fatal("group Redo", err)
	}
	_ = m.journal.Undo()
	for _, cancel := range []string{"escape", "outside", "camera"} {
		t.Run(cancel, func(t *testing.T) {
			s := m.beginMove(g, p)
			m.previewMove(s, end)
			switch cancel {
			case "escape":
				m.cancelMove(s)
			case "outside":
				_ = m.finishMove(s, canvas.Point{X: -1, Y: -1})
			case "camera":
				changed := g
				changed.View.Zoom = 2
				m.updateMoveGeometry(s, changed)
			}
			if m.move != nil || !slices.Equal(m.scene.Objects, before) || len(m.selectedIDs()) != 2 || m.journal.Snapshot().UndoCount != count {
				t.Fatal("canceled group changed document")
			}
		})
	}
	m.scene.Objects[2].X = 9990
	before = slices.Clone(m.scene.Objects)
	s = m.beginMove(g, p)
	_ = m.finishMove(s, end)
	if !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != count {
		t.Fatal("one invalid member partially moved group")
	}
	if m.nudge(1, 0) == nil || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("group nudge bounds not atomic")
	}
}
func TestGroupDeleteUndoSelectionOrderPlacementAndBranch(t *testing.T) {
	m := editModel(t)
	ids := []string{"block-1", "note-3"}
	m.setSelection(ids, "note-3")
	before := slices.Clone(m.scene.Objects)
	if m.commands.Enabled("width-more") == nil || m.resizeSelected(1, 0) == nil {
		t.Fatal("group resize unexpectedly enabled")
	}
	if err := m.deleteSelected(); err != nil || len(m.scene.Objects) != 1 || m.scene.Objects[0].ID != "connector-2" || len(m.selectedIDs()) != 0 {
		t.Fatal("group delete", err)
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), ids) || m.scene.Selected != "note-3" {
		t.Fatal("delete Undo lost IDs/order/selection", err)
	}
	if err := m.journal.Redo(); err != nil || len(m.scene.Objects) != 1 {
		t.Fatal("group delete Redo", err)
	}
	_ = m.journal.Undo()
	if err := m.nudge(1, 0); err != nil || m.journal.Snapshot().RedoCount != 0 || m.scene.Objects[0].X != 20 || m.scene.Objects[2].X != 420 || m.scene.Objects[1] != before[1] {
		t.Fatal("group nudge/branch", err)
	}
	if err := m.palette.Source("block").PlaceKeyboard(context.Background(), m.target, canvas.Point{X: 700}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(m.selectedIDs(), []string{"block-4"}) {
		t.Fatal("placement did not replace group")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.selectedIDs(), ids) {
		t.Fatal("placement Undo did not restore prior group", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(m.selectedIDs(), []string{"block-4"}) {
		t.Fatal("placement Redo", err)
	}
}
func TestGroupEditedFileRoundTripResetsTransientSelection(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "connector", canvas.Point{X: 200})
	f.place(t, "note", canvas.Point{X: 400, Y: 100})
	var saved []canvas.Rect
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-3"}, "note-3")
		if err := f.m.nudge(1, 0); err != nil {
			t.Fatal(err)
		}
		saved = slices.Clone(f.m.scene.Objects)
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if err := f.m.deleteSelected(); err != nil {
			t.Fatal(err)
		}
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if !reflect.DeepEqual(f.m.scene.Objects, saved) || !slices.Equal(f.m.selectedIDs(), []string{"note-3"}) || f.m.journal.Snapshot().UndoCount != 0 {
			t.Fatal("group-edit file load/selection/history")
		}
	})
}

func TestNavigationChordSelectionIntent(t *testing.T) {
	for _, tc := range []struct {
		code    key.Codes
		mods    key.Modifiers
		handled bool
		dx, dy  float64
	}{
		{key.CodeRightArrow, 0, true, -30, 0},
		{key.CodeRightArrow, key.Modifiers(1 << key.Shift), true, -30, 0},
		{key.CodeUpArrow, key.Modifiers(1 << key.Alt), true, 0, 30},
		{key.CodeRightArrow, key.Modifiers(1 << key.Alt), false, 0, 0},
		{key.CodeHome, key.Modifiers(1 << key.Alt), false, 0, 0},
		{key.CodeRightArrow, key.Modifiers(1<<key.Meta | 1<<key.Alt), true, 0, 0},
	} {
		handled, dx, dy := canvasNavigationKey(tc.code, tc.mods)
		if handled != tc.handled || dx != tc.dx || dy != tc.dy {
			t.Fatal("navigation/selection routing", tc, handled, dx, dy)
		}
	}
}
func TestActiveGroupSelectionAndInstallCannotOverwriteGesture(t *testing.T) {
	m := editModel(t)
	g := cornerGeometry()
	m.snap = false
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	saved, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(m.scene.Objects)
	p := selectionPoint(g, canvas.Point{X: 40, Y: 25})
	s := m.beginMove(g, p)
	end := canvas.Point{X: p.X + 60, Y: p.Y + 30}
	m.previewMove(s, end)
	if err := drawingCodec(m).Install(saved); err == nil {
		t.Fatal("install accepted active group move")
	}
	m.managedSelection("connector-2")
	if m.move != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), []string{"connector-2"}) {
		t.Fatal("semantic reselection did not cancel group preview")
	}
	_ = m.finishMove(s, end)
	if !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != 3 {
		t.Fatal("stale release after reselection committed")
	}
}
