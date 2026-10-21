package main

import (
	"github.com/dshills/vellum/canvas"
	"testing"
)

func TestHistoryCaptionsFollowReplayBranchAndNew(t *testing.T) {
	f := fileModel(t, false)
	check := func(undo, redo string) {
		t.Helper()
		f.owner(t, func() {
			root := view(f.a, f.ref, f.m)
			if geometryDialogControl(t, root, "undo").Text != undo || geometryDialogControl(t, root, "redo").Text != redo {
				t.Fatal("wrong history captions", f.m.historyCaption("Undo"), f.m.historyCaption("Redo"))
			}
		})
	}
	check("Undo", "Redo")
	f.place(t, "block", canvas.Point{})
	check("Undo Place block", "Redo")
	f.owner(t, func() {
		if err := f.m.nudge(1, 0); err != nil {
			t.Fatal(err)
		}
	})
	check("Undo Move selection", "Redo")
	f.owner(t, func() { _ = f.m.journal.Undo() })
	check("Undo Place block", "Redo Move selection")
	f.owner(t, func() { _ = f.m.journal.Redo() })
	check("Undo Move selection", "Redo")
	f.owner(t, func() {
		_ = f.m.journal.Undo()
		if err := f.m.deleteSelected(); err != nil {
			t.Fatal(err)
		}
	})
	check("Undo Delete selection", "Redo")
	f.owner(t, func() {
		if err := f.m.newDrawing(f.a, f.ref, true); err != nil {
			t.Fatal(err)
		}
	})
	check("Undo", "Redo")
}
