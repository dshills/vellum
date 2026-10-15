package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"slices"
	"testing"
)

func TestTypedPlacementHistoryAndValidation(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer m.palette.Close()
	for _, id := range []string{"block", "connector", "note"} {
		if err := m.palette.Source(id).PlaceKeyboard(context.Background(), m.target, canvas.Point{X: 33, Y: -27}); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.scene.Objects) != 3 || m.scene.Objects[0].X != 40 || m.scene.Objects[0].Y != -20 || m.journal.Snapshot().UndoCount != 3 {
		t.Fatal("typed placements did not use snap/history", m.scene)
	}
	original := slices.Clone(m.scene.Objects)
	for range 3 {
		if err := m.commands.Invoke(context.Background(), "undo"); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.scene.Objects) != 0 {
		t.Fatal("Undo did not remove exactly one object per placement")
	}
	for range 3 {
		if err := m.commands.Invoke(context.Background(), "redo"); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(original, m.scene.Objects) || m.next != 3 || m.pending != nil {
		t.Fatal("Redo changed object identities or retained payload")
	}
	m.rejectNotes = true
	if err := m.palette.Source("note").PlaceKeyboard(context.Background(), m.target, canvas.Point{}); !errors.Is(err, placement.ErrRejected) {
		t.Fatal(err)
	}
	m.locked = true
	if err := m.palette.Source("block").PlaceKeyboard(context.Background(), m.target, canvas.Point{}); !errors.Is(err, placement.ErrRejected) {
		t.Fatal(err)
	}
	if len(m.scene.Objects) != 3 || m.journal.Snapshot().UndoCount != 3 {
		t.Fatal("rejection changed history")
	}
	m.locked = false
	m.snap = false
	if err := m.palette.Source("connector").PlaceKeyboard(context.Background(), m.target, canvas.Point{X: 33, Y: 27}); err != nil {
		t.Fatal(err)
	}
	if m.scene.Objects[3].X != 33 || m.scene.Objects[3].Y != 27 {
		t.Fatal("optional snapping ignored")
	}
}
