package main

import (
	"context"
	"github.com/dshills/vellum/canvas"
	"maps"
	"math"
	"slices"
	"testing"
)

func dimensionModel(t *testing.T) *model {
	t.Helper()
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.palette.Close)
	m.scene.Objects = []canvas.Rect{{ID: "block-1", Width: 80, Height: 40}, {ID: "note-2", X: 100, Width: 100, Height: 50}, {ID: "connector-3", X: 300, Y: 40, Width: 120, Height: 30}}
	m.groups = map[string]string{"block-1": "group-1", "note-2": "group-1"}
	m.next = 3
	m.nextGroup = 1
	m.setSelection([]string{"block-1", "connector-3"}, "block-1")
	return m
}
func TestMatchDimensionsUsesPrimaryUnitAndOneHistory(t *testing.T) {
	for _, action := range []string{"match-width", "match-height", "match-size"} {
		m := dimensionModel(t)
		before := slices.Clone(m.scene.Objects)
		groups := maps.Clone(m.groups)
		ids := m.selectedIDs()
		state := m.journal.Snapshot()
		if err := m.commands.Invoke(context.Background(), action); err != nil {
			t.Fatal(err)
		}
		got := m.scene.Objects[2]
		if (action != "match-height" && got.Width != 200) || (action != "match-width" && got.Height != 50) || got.X != 300 || got.Y != 40 || !slices.Equal(before[:2], m.scene.Objects[:2]) {
			t.Fatal("primary/group changed or wrong match", action, m.scene.Objects)
		}
		after := slices.Clone(m.scene.Objects)
		if m.journal.Snapshot().UndoCount != state.UndoCount+1 || !maps.Equal(groups, m.groups) || !slices.Equal(ids, m.selectedIDs()) {
			t.Fatal("metadata/history")
		}
		if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
			t.Fatal("Undo", err)
		}
		if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) {
			t.Fatal("Redo", err)
		}
	}
}
func TestMatchDimensionsUniformGroupsAndRejectsIncompatibleOrOutOfBounds(t *testing.T) {
	m := dimensionModel(t)
	m.setSelection(m.selectedIDs(), "connector-3")
	before := slices.Clone(m.scene.Objects)
	if err := m.matchDimensions("match-size"); err != nil {
		t.Fatal(err)
	}
	bounds, _ := selectionBounds(m.scene.Objects, []string{"block-1", "note-2"})
	if bounds.Width != 120 || bounds.Height != 30 || math.Abs(m.scene.Objects[0].Width/m.scene.Objects[0].Height-2) > 1e-12 || m.scene.Objects[1].X != 60 {
		t.Fatal("group proportions/spacing", m.scene.Objects)
	}
	_ = m.journal.Undo()
	m.scene.Objects[2].Height = 60
	state := m.journal.Snapshot()
	current := slices.Clone(m.scene.Objects)
	if m.matchDimensions("match-size") == nil || !slices.Equal(current, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("incompatible group distorted")
	}
	m.scene.Objects[2].Height = 30
	m.scene.Objects[2].Width = 8
	current = slices.Clone(m.scene.Objects)
	if m.matchDimensions("match-width") == nil || !slices.Equal(current, m.scene.Objects) {
		t.Fatal("member limits allowed partial edit")
	}
	m.scene.Objects = slices.Clone(before)
	m.prompt = "geometry"
	if m.matchDimensions("match-width") == nil {
		t.Fatal("busy match")
	}
	m.prompt = ""
	m.setSelection([]string{"block-1"}, "block-1")
	if m.matchDimensions("match-width") == nil {
		t.Fatal("single unit match")
	}
}
