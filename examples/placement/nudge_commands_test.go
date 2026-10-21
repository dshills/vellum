package main

import (
	"context"
	"maps"
	"slices"
	"testing"
)

func TestExactNudgesBypassSnapPreserveGroupsAndReplay(t *testing.T) {
	m := groupedModel(t)
	m.snap = true
	m.gridSpacing = 100
	before := slices.Clone(m.scene.Objects)
	groups := maps.Clone(m.groups)
	ids := m.selectedIDs()
	primary := m.scene.Selected
	count := m.journal.Snapshot().UndoCount
	for _, tt := range []struct {
		id     string
		dx, dy float64
	}{{"nudge-1-right", 1, 0}, {"nudge-10-down", 0, 10}, {"nudge-1-left", -1, 0}, {"nudge-10-up", 0, -10}} {
		if err := m.commands.Invoke(context.Background(), tt.id); err != nil {
			t.Fatal(err)
		}
		after := slices.Clone(m.scene.Objects)
		for i, o := range before {
			want := o
			if slices.Contains(ids, o.ID) {
				want.X += tt.dx
				want.Y += tt.dy
			}
			if after[i] != want {
				t.Fatal("grid affected exact nudge", tt.id, after[i], want)
			}
		}
		if m.journal.Snapshot().UndoCount != count+1 {
			t.Fatal("nudge not one edit")
		}
		if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
			t.Fatal("nudge Undo", err)
		}
		if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) {
			t.Fatal("nudge Redo", err)
		}
		_ = m.journal.Undo()
	}
	if !maps.Equal(groups, m.groups) || !slices.Equal(ids, m.selectedIDs()) || primary != m.scene.Selected {
		t.Fatal("group or selection changed")
	}
	m.scene.Objects[0].X = 10000
	current := slices.Clone(m.scene.Objects)
	state := m.journal.Snapshot()
	if m.commands.Invoke(context.Background(), "nudge-10-right") == nil || !slices.Equal(current, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("partial out-of-bounds nudge")
	}
	m.prompt = "geometry"
	if m.nudgeWorld(1, 0, "Nudge") == nil {
		t.Fatal("prompt allowed exact nudge")
	}
	m.prompt = ""
	m.setSelection(nil, "")
	if m.nudgeWorld(1, 0, "Nudge") == nil {
		t.Fatal("empty nudge")
	}
}
