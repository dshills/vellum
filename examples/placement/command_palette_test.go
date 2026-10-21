package main

import (
	"context"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"slices"
	"strings"
	"testing"
)

func TestCommandPaletteSearchDisabledReasonsReplayAndCancellation(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		before := slices.Clone(f.m.scene.Objects)
		state := f.m.journal.Snapshot()
		if err := f.m.beginCommandPalette(); err != nil {
			t.Fatal(err)
		}
		d := f.m.commandDraft
		f.m.updateCommandQuery(d, "nudge-1-right")
		if len(d.matches()) != 1 || d.selected != "nudge-1-right" {
			t.Fatal("filter")
		}
		dialog := f.m.commandPaletteDialog(f.a, f.ref)
		if err := ui.Validate(dialog); err != nil {
			t.Fatal(err)
		}
		if err := f.m.runCommandChoice(f.ref.Context(), d); err != nil {
			t.Fatal(err)
		}
		if f.m.scene.Objects[0].X != before[0].X+1 || f.m.journal.Snapshot().UndoCount != state.UndoCount+1 || f.m.prompt != "" {
			t.Fatal("command not routed exactly once")
		}
		if f.m.runCommandChoice(context.Background(), d) == nil {
			t.Fatal("stale Run replayed")
		}
		_ = f.m.journal.Undo()
		_ = f.m.beginCommandPalette()
		d = f.m.commandDraft
		old := f.m.commandPaletteDialog(f.a, f.ref)
		f.m.updateCommandQuery(d, "this does not match")
		if d.selected != "" || len(d.matches()) != 0 {
			t.Fatal("empty match")
		}
		f.m.updateCommandQuery(d, strings.Repeat("x", 129))
		if d.message == "" {
			t.Fatal("search unbounded")
		}
		f.m.cancelCommandPalette(d)
		if !slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot().UndoCount != state.UndoCount {
			t.Fatal("cancel mutated history")
		}
		f.m.prompt = "close"
		old.Dismiss()
		if f.m.prompt != "close" {
			t.Fatal("old cancel dismissed newer prompt")
		}
		f.m.prompt = ""
		f.m.setSelection(nil, "")
		_ = f.m.beginCommandPalette()
		d = f.m.commandDraft
		f.m.updateCommandQuery(d, "delete-forward")
		if len(d.matches()) != 1 || d.matches()[0].reason == "" {
			t.Fatal("disabled reason missing")
		}
		if f.m.runCommandChoice(f.ref.Context(), d) == nil || f.m.prompt != "commands" {
			t.Fatal("disabled ran")
		}
		f.m.cancelCommandPalette(d)
	})
}
func TestCommandPaletteStaleSelectionAndLiveAdmission(t *testing.T) {
	m := groupedModel(t)
	_ = m.beginCommandPalette()
	d := m.commandDraft
	m.updateCommandQuery(d, "delete-forward")
	m.scene.Objects[0].X += 1 // Revision is unchanged; direct external state is not a new UI selection.
	m.setSelection([]string{"connector-2"}, "connector-2")
	if m.runCommandChoice(context.Background(), d) == nil || len(m.scene.Objects) != 3 {
		t.Fatal("stale selection acted")
	}
	_ = m.beginCommandPalette()
	d = m.commandDraft
	m.updateCommandQuery(d, "delete-forward")
	m.move = &objectMove{}
	if m.runCommandChoice(context.Background(), d) == nil || len(m.scene.Objects) != 3 {
		t.Fatal("live busy gate bypassed")
	}
	m.move = nil
}
