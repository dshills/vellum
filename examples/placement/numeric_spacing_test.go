package main

import (
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"testing"
)

func TestNumericSpacingRigidGroupsExactGapAndReplay(t *testing.T) {
	for _, axis := range []string{"horizontal", "vertical"} {
		m := dimensionModel(t)
		before := slices.Clone(m.scene.Objects)
		count := m.journal.Snapshot().UndoCount
		m.snap = true
		m.gridSpacing = 100
		if err := m.beginSpacing(); err != nil {
			t.Fatal(err)
		}
		d := m.spacingDraft
		d.axis = axis
		d.gap = "25.125"
		m.prompt = ""
		if err := m.applySpacing(d); err != nil {
			t.Fatal(err)
		}
		after := slices.Clone(m.scene.Objects)
		if !slices.Equal(before[:2], after[:2]) || (axis == "horizontal" && (after[2].X != 225.125 || after[2].Y != 40)) || (axis == "vertical" && (after[2].Y != 75.125 || after[2].X != 300)) {
			t.Fatal("gap/group/cross-axis", axis, after)
		}
		if m.journal.Snapshot().UndoCount != count+1 {
			t.Fatal("not one spacing edit")
		}
		if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
			t.Fatal("Undo", err)
		}
		if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) {
			t.Fatal("Redo", err)
		}
	}
}
func TestNumericSpacingValidationAtomicBoundsAndStaleDraft(t *testing.T) {
	m := dimensionModel(t)
	before := slices.Clone(m.scene.Objects)
	for _, gap := range []float64{-1, 10001, math.NaN(), math.Inf(1)} {
		if _, err := m.spacedSelection("horizontal", gap); err == nil {
			t.Fatal("invalid gap")
		}
	}
	if _, err := m.spacedSelection("diagonal", 10); err == nil {
		t.Fatal("bad axis")
	}
	m.scene.Objects[0].X = 9900
	m.scene.Objects[1].X = 10000
	m.scene.Objects[2].X = 10000
	state := m.journal.Snapshot()
	current := slices.Clone(m.scene.Objects)
	if _, err := m.spacedSelection("horizontal", 100); err == nil || !slices.Equal(current, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("partial bounds change")
	}
	m.scene.Objects = before
	_ = m.beginSpacing()
	d := m.spacingDraft
	m.prompt = ""
	m.scene.Objects[0].X = 1
	if m.applySpacing(d) == nil {
		t.Fatal("stale draft applied")
	}
}
func TestNumericSpacingDialogInvalidVisibleCancelAndOldCallbacks(t *testing.T) {
	f := fileModel(t, false)
	f.m.scene.Objects = dimensionModel(t).scene.Objects
	f.m.groups = map[string]string{"block-1": "group-1", "note-2": "group-1"}
	f.m.setSelection([]string{"block-1", "connector-3"}, "block-1")
	f.owner(t, func() {
		before := slices.Clone(f.m.scene.Objects)
		state := f.m.journal.Snapshot()
		if err := f.m.commands.Invoke(f.ref.Context(), "numeric-spacing"); err != nil {
			t.Fatal(err)
		}
		dialog := f.m.spacingDialog(f.a, f.ref)
		geometryDialogControl(t, dialog, "spacing-gap").Change(ui.ChangeEvent{Text: "NaN"})
		geometryDialogControl(t, dialog, "spacing-apply").Activate()
		if f.m.prompt != "spacing" || f.m.spacingDraft.message == "" || !slices.Equal(before, f.m.scene.Objects) || state != f.m.journal.Snapshot() {
			t.Fatal("invalid gap not visible/atomic")
		}
		geometryDialogControl(t, f.m.spacingDialog(f.a, f.ref), "spacing-cancel").Activate()
		f.m.prompt = "close"
		dialog.Dismiss()
		geometryDialogControl(t, dialog, "spacing-apply").Activate()
		if f.m.prompt != "close" || state != f.m.journal.Snapshot() {
			t.Fatal("stale dialog changed close")
		}
	})
}

func TestNumericSpacingPaintTiesAndNonFirstGroupTranslation(t *testing.T) {
	m := dimensionModel(t)
	m.scene.Objects[2].X = 0
	after, err := m.spacedSelection("horizontal", 25)
	if err != nil || after[2].X != 225 || after[0].X != 0 {
		t.Fatal("paint tie ordering", err, after)
	}
	m.scene.Objects[2].X = -200
	after, err = m.spacedSelection("horizontal", 25)
	if err != nil || after[0].X != -55 || after[1].X != 45 || after[2].X != -200 || after[0].Y != 0 || after[1].Y != 0 {
		t.Fatal("group was not translated rigidly", err, after)
	}
}
