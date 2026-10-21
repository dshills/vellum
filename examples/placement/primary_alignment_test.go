package main

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func TestPrimaryAlignmentKeepsReferenceExactAndReplays(t *testing.T) {
	for _, o := range primaryAlignOptions() {
		t.Run(o.ID, func(t *testing.T) {
			m := dimensionModel(t)
			m.snap = true
			m.gridSpacing = 100
			m.scene.View.Center = canvas.Point{X: 400, Y: -200}
			m.scene.View.Zoom = 2.5
			before := slices.Clone(m.scene.Objects)
			groups := maps.Clone(m.groups)
			ids := m.selectedIDs()
			state := m.journal.Snapshot()
			if err := m.commands.Invoke(context.Background(), o.ID); err != nil {
				t.Fatal(err)
			}
			got := m.scene.Objects[2]
			x, y := 300.0, 40.0
			switch o.ID {
			case "align-primary-left":
				x = 0
			case "align-primary-hcenter":
				x = 40
			case "align-primary-right":
				x = 80
			case "align-primary-top":
				y = 0
			case "align-primary-vcenter":
				y = 10
			case "align-primary-bottom":
				y = 20
			}
			if got.X != x || got.Y != y || !slices.Equal(before[:2], m.scene.Objects[:2]) || got.Width != before[2].Width || got.Height != before[2].Height {
				t.Fatal("wrong alignment or moved primary", o.ID, m.scene.Objects)
			}
			if !maps.Equal(groups, m.groups) || !slices.Equal(ids, m.selectedIDs()) || m.scene.Selected != "block-1" || m.journal.Snapshot().UndoCount != state.UndoCount+1 {
				t.Fatal("metadata/selection/history changed")
			}
			after := slices.Clone(m.scene.Objects)
			if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
				t.Fatal("Undo", err)
			}
			if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) {
				t.Fatal("Redo", err)
			}
			state = m.journal.Snapshot()
			if err := m.alignToPrimary(o.ID); err != nil || state != m.journal.Snapshot() {
				t.Fatal("no-op alignment added history", err)
			}
		})
	}
}
func TestPrimaryAlignmentMovesGroupsRigidlyAndRejectsAtomically(t *testing.T) {
	m := dimensionModel(t)
	m.setSelection(m.selectedIDs(), "connector-3")
	before := slices.Clone(m.scene.Objects)
	if err := m.alignToPrimary("align-primary-right"); err != nil {
		t.Fatal(err)
	}
	if m.scene.Objects[0].X != 220 || m.scene.Objects[1].X != 320 || m.scene.Objects[0].Y != 0 || m.scene.Objects[1].Y != 0 || m.scene.Objects[2] != before[2] {
		t.Fatal("group not rigid or primary moved", m.scene.Objects)
	}
	_ = m.journal.Undo()
	m.scene.Objects[2].X = 10000
	before = slices.Clone(m.scene.Objects)
	state := m.journal.Snapshot()
	if m.alignToPrimary("align-primary-left") == nil || !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("member overflow not atomic")
	}
	if m.alignToPrimary("not-an-alignment") == nil {
		t.Fatal("unknown action")
	}
	m.prompt = "close"
	if m.alignToPrimary("align-primary-left") == nil {
		t.Fatal("busy action")
	}
	m.prompt = ""
	m.setSelection([]string{"connector-3"}, "connector-3")
	if m.primaryAlignmentAvailable() == nil {
		t.Fatal("single unit alignment")
	}
}
