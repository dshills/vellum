package main

import (
	"context"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func TestAspectAdapterShiftHandleHasPriorityWithoutChangingPreference(t *testing.T) {
	m := editModel(t)
	m.snap = false
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	g := groupGeometry()
	bounds, _ := selectionBounds(before, m.selectedIDs())
	start := cornerHandles(bounds, g)[southEast-1].center
	s := m.beginCanvasGesture(g, start, true)
	if s == nil || !s.resize || s.corner != southEast || !s.aspectLocked || s.selectionOnly || s.marquee || m.aspectLocked ||
		!slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) {
		t.Fatal("Shift toggled selection instead of capturing its handle", s)
	}
	end := canvas.Point{X: start.X + bounds.Width*.25*g.Scale*g.View.Zoom, Y: start.Y}
	m.previewMove(s, end)
	afterBounds, _ := selectionBounds(m.scene.Objects, s.selection)
	if !aspectClose(afterBounds.Width/bounds.Width, afterBounds.Height/bounds.Height) || m.aspectLocked || m.journal.Snapshot() != state {
		t.Fatal("temporary Shift lock changed preference/history or failed to preserve ratio")
	}
	m.cancelMove(s)
	if !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot() != state || m.aspectLocked {
		t.Fatal("temporary aspect gesture did not cancel atomically")
	}
	// Shift away from a handle still toggles membership instead of resizing.
	s = m.beginCanvasGesture(g, selectionPoint(g, canvas.Point{X: 40, Y: 25}), true)
	if s == nil || s.resize || !s.selectionOnly || m.aspectLocked || !slices.Equal(m.selectedIDs(), []string{"note-3"}) {
		t.Fatal("Shift interior lost existing selection-toggle behavior", s)
	}
	m.cancelMove(s)
}

func TestAspectAdapterCapturesPreferenceForHistoryUndoRedoAndCancel(t *testing.T) {
	m := editModel(t)
	m.snap, m.aspectLocked = false, true
	m.setSelection([]string{"block-1"}, "block-1")
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	g := cornerGeometry()
	start := cornerHandles(before[0], g)[southEast-1].center
	end := canvas.Point{X: start.X + 40*g.Scale*g.View.Zoom, Y: start.Y}
	s := m.beginCanvasGesture(g, start, false)
	if s == nil || !s.resize || !s.aspectLocked {
		t.Fatal("persistent preference was not captured")
	}
	m.aspectLocked = false // Simulate a preference change after admission.
	m.previewMove(s, end)
	if m.scene.Objects[0].Width != 120 || m.scene.Objects[0].Height != 75 || m.scene.Objects[1] != before[1] ||
		m.scene.Objects[2] != before[2] || m.journal.Snapshot() != state {
		t.Fatal("preview reread preference or committed history")
	}
	if err := m.finishMove(s, end); err != nil {
		t.Fatal(err)
	}
	after := slices.Clone(m.scene.Objects)
	if m.move != nil || m.aspectLocked || m.next != 3 || m.journal.Snapshot().UndoCount != state.UndoCount+1 ||
		!slices.Equal(m.selectedIDs(), []string{"block-1"}) {
		t.Fatal("aspect release was not exactly one edit")
	}
	_ = m.finishMove(s, end)
	if m.journal.Snapshot().UndoCount != state.UndoCount+1 {
		t.Fatal("duplicate release committed twice")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("Undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
		t.Fatal("Redo", err)
	}
	if err := m.journal.Undo(); err != nil {
		t.Fatal(err)
	}
	state = m.journal.Snapshot()
	// The reverse preference change must not turn an unlocked drag into a lock.
	s = m.beginMove(g, start)
	if s == nil || s.aspectLocked {
		t.Fatal("unlocked preference not captured")
	}
	m.aspectLocked = true
	m.previewMove(s, end)
	if m.scene.Objects[0].Width != 120 || m.scene.Objects[0].Height != 50 {
		t.Fatal("gesture reread new preference")
	}
	m.cancelMove(s)
	if !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot() != state || !m.aspectLocked {
		t.Fatal("cancel changed history or preference")
	}
}

func TestAspectAdapterNumericDraftCapturesLockAndRejectsInconsistentLimits(t *testing.T) {
	m := editModel(t)
	m.setSelection([]string{"block-1"}, "block-1")
	m.aspectLocked = true
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	if err := m.beginGeometry(); err != nil {
		t.Fatal(err)
	}
	d := m.geometry
	m.aspectLocked = false
	d.values = [4]string{"1.25", "-2.5", "160", "50"} // Only width changes.
	m.prompt = ""
	if err := m.applyGeometry(d); err != nil {
		t.Fatal(err)
	}
	got := m.scene.Objects[0]
	if got.X != 1.25 || got.Y != -2.5 || got.Width != 160 || got.Height != 100 ||
		m.journal.Snapshot().UndoCount != state.UndoCount+1 || m.aspectLocked {
		t.Fatal("numeric draft did not retain its captured lock", got)
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("numeric Undo", err)
	}
	m.aspectLocked = true
	if err := m.beginGeometry(); err != nil {
		t.Fatal(err)
	}
	d = m.geometry
	d.values[2], d.values[3] = "160", "110"
	m.prompt = ""
	state = m.journal.Snapshot()
	if err := m.applyGeometry(d); err == nil || !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot() != state {
		t.Fatal("inconsistent sizes changed geometry/history", err)
	}
	// A derived dimension may not exceed a component limit even though the
	// explicitly edited dimension remains individually valid.
	m.scene.Objects[0].Width, m.scene.Objects[0].Height = 8, 2000
	before = slices.Clone(m.scene.Objects)
	if err := m.beginGeometry(); err != nil {
		t.Fatal(err)
	}
	d = m.geometry
	d.values[2] = "16"
	m.prompt = ""
	state = m.journal.Snapshot()
	if err := m.applyGeometry(d); err == nil || !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot() != state {
		t.Fatal("derived height limit was bypassed", err)
	}
}

func TestAspectAdapterResizeButtonsRespectPreferenceAndSingleSelection(t *testing.T) {
	m := editModel(t)
	m.setSelection([]string{"block-1"}, "block-1")
	m.aspectLocked = true
	before, count := slices.Clone(m.scene.Objects), m.journal.Snapshot().UndoCount
	if err := m.commands.Invoke(context.Background(), "width-more"); err != nil {
		t.Fatal(err)
	}
	if m.scene.Objects[0].Width != 100 || m.scene.Objects[0].Height != 62.5 || m.scene.Objects[0].X != before[0].X ||
		m.scene.Objects[0].Y != before[0].Y || m.scene.Objects[1] != before[1] || m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("width button did not derive the other dimension")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("button Undo", err)
	}
	m.aspectLocked = false
	m.snap = false
	if err := m.commands.Invoke(context.Background(), "height-more"); err != nil {
		t.Fatal(err)
	}
	if m.scene.Objects[0].Width != 80 || m.scene.Objects[0].Height != 51 {
		t.Fatal("unlocked button changed both dimensions")
	}
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	current, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	if err := m.commands.Invoke(context.Background(), "width-more"); err == nil || !slices.Equal(current, m.scene.Objects) || m.journal.Snapshot() != state {
		t.Fatal("button bypassed existing single-selection gate", err)
	}
}

func TestAspectAdapterLockedNumericDerivedDimensionAcceptsExactLimits(t *testing.T) {
	for _, test := range []struct {
		name, width  string
		before, want float64
	}{
		{name: "minimum", width: "8", before: 49, want: 8},
		{name: "maximum", width: "2000", before: 15, want: 2000},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := editModel(t)
			m.aspectLocked = true
			m.setSelection([]string{"block-1"}, "block-1")
			m.scene.Objects[0].Width, m.scene.Objects[0].Height = test.before, test.before
			before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
			if err := m.beginGeometry(); err != nil {
				t.Fatal(err)
			}
			draft := m.geometry
			unchangedHeight := draft.values[3]
			draft.values[2] = test.width
			m.prompt = ""
			if err := m.applyGeometry(draft); err != nil {
				t.Fatal("valid locked boundary refused", err)
			}
			got := m.scene.Objects[0]
			if got.Width != test.want || got.Height != test.want || draft.values[3] != unchangedHeight ||
				got.ID != before[0].ID || got.X != before[0].X || got.Y != before[0].Y ||
				m.scene.Objects[1] != before[1] || m.scene.Objects[2] != before[2] ||
				m.journal.Snapshot().UndoCount != state.UndoCount+1 {
				t.Fatal("derived boundary was not exact or changed unrelated state", got)
			}
			if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
				t.Fatal("boundary Undo", err)
			}
		})
	}
}
