package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func geometryDialogControl(t *testing.T, root ui.Node, id string) ui.Node {
	t.Helper()
	var find func(ui.Node) (ui.Node, bool)
	find = func(n ui.Node) (ui.Node, bool) {
		if n.ID == id {
			return n, true
		}
		for _, child := range n.Children {
			if found, ok := find(child); ok {
				return found, true
			}
		}
		return ui.Node{}, false
	}
	if node, ok := find(root); ok {
		return node
	}
	t.Fatalf("missing geometry dialog control %q", id)
	return ui.Node{}
}

func TestGeometryDialogStaleCallbacksKeepNewerPromptAndDraft(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		before, state := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
		if err := f.m.beginGeometry(); err != nil {
			t.Fatal(err)
		}
		oldDraft := f.m.geometry
		old := f.m.geometryDialog(f.a, f.ref)
		apply := geometryDialogControl(t, old, "geometry-apply").Activate
		cancel := geometryDialogControl(t, old, "geometry-cancel").Activate
		change := geometryDialogControl(t, old, "geometry-X").Change
		values := oldDraft.values
		f.m.prompt = "close"
		change(ui.ChangeEvent{Text: "999"})
		apply()
		cancel()
		old.Dismiss()
		if f.m.prompt != "close" || f.m.geometry != oldDraft || oldDraft.values != values ||
			!slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot() != state {
			t.Fatal("old geometry callbacks changed a newer close prompt")
		}
		f.m.prompt, f.m.geometry = "", nil
		if err := f.m.beginGeometry(); err != nil {
			t.Fatal(err)
		}
		newDraft := f.m.geometry
		newValues := newDraft.values
		if newDraft == oldDraft {
			t.Fatal("draft identity was reused")
		}
		change(ui.ChangeEvent{Text: "999"})
		apply()
		cancel()
		old.Dismiss()
		if f.m.prompt != "geometry" || f.m.geometry != newDraft || newDraft.values != newValues || oldDraft.values != values ||
			!slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot() != state {
			t.Fatal("old geometry callbacks changed or cleared a replacement draft")
		}
	})
}

func TestGeometryDialogRetiredWindowCallbacksAreInert(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	if _, err := f.a.OpenWindow(app.Window{ID: "geometry-keeper", Title: "Keeper", Width: 100, Height: 100,
		View: func() ui.Node { return ui.Node{ID: "keeper-root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if err := f.m.beginGeometry(); err != nil {
			t.Fatal(err)
		}
		draft := f.m.geometry
		old := f.m.geometryDialog(f.a, f.ref)
		before, state, values := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot(), draft.values
		if err := f.a.CloseWindow(f.ref); err != nil {
			t.Fatal(err)
		}
		if f.ref.Context().Err() == nil {
			t.Fatal("target window lifetime was not retired")
		}
		geometryDialogControl(t, old, "geometry-X").Change(ui.ChangeEvent{Text: "999"})
		geometryDialogControl(t, old, "geometry-apply").Activate()
		geometryDialogControl(t, old, "geometry-cancel").Activate()
		old.Dismiss()
		if f.m.geometry != draft || f.m.prompt != "geometry" || draft.values != values ||
			!slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot() != state {
			t.Fatal("retired-window callbacks touched retained owner state")
		}
	})
}

func TestGeometryDialogInvalidThenValidApplyHasVisibleErrorAndOneHistory(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.journal.MarkSaved(); err != nil {
			t.Fatal(err)
		}
		before, state := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
		if err := f.m.beginGeometry(); err != nil {
			t.Fatal(err)
		}
		draft := f.m.geometry
		dialog := f.m.geometryDialog(f.a, f.ref)
		geometryDialogControl(t, dialog, "geometry-Width").Change(ui.ChangeEvent{Text: "not-a-number"})
		geometryDialogControl(t, dialog, "geometry-apply").Activate()
		invalid := f.m.geometryDialog(f.a, f.ref)
		message := geometryDialogControl(t, invalid, "geometry-error")
		if !invalid.Open || f.m.geometry != draft || f.m.prompt != "geometry" || message.Kind != ui.Label ||
			!strings.Contains(message.Text, "Width") || message.Text != draft.message ||
			!slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot() != state {
			t.Fatal("invalid Apply hid its error or mutated history", message.Text)
		}
		for id, value := range map[string]string{"geometry-X": "123.25", "geometry-Y": "-17.5", "geometry-Width": "90.5", "geometry-Height": "60.25"} {
			geometryDialogControl(t, invalid, id).Change(ui.ChangeEvent{Text: value})
		}
		apply := geometryDialogControl(t, invalid, "geometry-apply").Activate
		apply()
		after := slices.Clone(f.m.scene.Objects)
		want := before[0]
		want.X, want.Y, want.Width, want.Height = 123.25, -17.5, 90.5, 60.25
		got := f.m.journal.Snapshot()
		if f.m.geometry != nil || f.m.prompt != "" || len(after) != 1 || after[0] != want || got.UndoCount != state.UndoCount+1 ||
			got.UndoLabel != "Edit selection geometry" || !got.Dirty || f.m.scene.Selected != before[0].ID {
			t.Fatal("valid Apply did not create one exact edit", after, got)
		}
		apply()
		if f.m.journal.Snapshot() != got {
			t.Fatal("old Apply replayed after successful admission")
		}
		if err := f.m.journal.Undo(); err != nil || !slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot().Dirty {
			t.Fatal("geometry dialog Undo", err)
		}
		if err := f.m.journal.Redo(); err != nil || !slices.Equal(after, f.m.scene.Objects) {
			t.Fatal("geometry dialog Redo", err)
		}
	})
}

func TestGeometryDialogStaleActiveDraftClosesWithoutOverwritingNewGeometry(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.beginGeometry(); err != nil {
			t.Fatal(err)
		}
		dialog := f.m.geometryDialog(f.a, f.ref)
		geometryDialogControl(t, dialog, "geometry-X").Change(ui.ChangeEvent{Text: "500"})
		// An external owner-held edit bypasses the normal prompt gate. The
		// stale draft must retire without replacing or reverting that edit.
		f.m.scene.Objects[0].Y = 41.5
		current, state := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
		geometryDialogControl(t, dialog, "geometry-apply").Activate()
		if f.m.geometry != nil || f.m.prompt != "" || !slices.Equal(current, f.m.scene.Objects) ||
			f.m.journal.Snapshot() != state || !strings.Contains(f.m.status, history.ErrStale.Error()) {
			t.Fatal("stale active draft remained open or overwrote current geometry", f.m.status)
		}
	})
}
