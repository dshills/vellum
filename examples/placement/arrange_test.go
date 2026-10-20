package main

import (
	"math"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func TestAllAlignmentActionsHistoryAndSelection(t *testing.T) {
	for _, action := range arrangementActions[:6] {
		t.Run(action.id, func(t *testing.T) {
			m := editModel(t)
			m.scene.Objects[0].Y = -100
			m.scene.Objects[2].Y = 150
			m.setSelection([]string{"block-1", "note-3"}, "block-1")
			before := slices.Clone(m.scene.Objects)
			view, next := m.scene.View, m.next
			b, _ := selectionBounds(before, m.selectedIDs())
			if err := m.arrangeSelected(action.id); err != nil {
				t.Fatal(err)
			}
			after := slices.Clone(m.scene.Objects)
			for _, i := range []int{0, 2} {
				o := after[i]
				want := before[i]
				switch action.id {
				case "align-left":
					want.X = b.X
				case "align-hcenter":
					want.X = b.X + (b.Width-want.Width)/2
				case "align-right":
					want.X = b.X + b.Width - want.Width
				case "align-top":
					want.Y = b.Y
				case "align-vcenter":
					want.Y = b.Y + (b.Height-want.Height)/2
				case "align-bottom":
					want.Y = b.Y + b.Height - want.Height
				}
				if o != want {
					t.Fatal("alignment geometry", o, want)
				}
			}
			if after[1] != before[1] || m.scene.View != view || m.next != next || m.journal.Snapshot().UndoCount != 4 || m.scene.Selected != "block-1" || len(m.selectedIDs()) != 2 {
				t.Fatal("unrelated state changed")
			}
			_ = m.journal.Undo()
			if !slices.Equal(m.scene.Objects, before) || len(m.selectedIDs()) != 2 {
				t.Fatal("alignment Undo")
			}
			_ = m.journal.Redo()
			if !slices.Equal(m.scene.Objects, after) {
				t.Fatal("alignment Redo")
			}
			if err := m.arrangeSelected(action.id); err != nil || m.journal.Snapshot().UndoCount != 4 {
				t.Fatal("no-op alignment added history", err)
			}
		})
	}
}
func TestDistributionGapsOrderCameraAndHistory(t *testing.T) {
	for _, axis := range []string{"h", "v"} {
		m := editModel(t)
		m.scene.Objects[0].X, m.scene.Objects[0].Y = 500, 500
		m.scene.Objects[1].X, m.scene.Objects[1].Y = 0, 0
		m.scene.Objects[2].X, m.scene.Objects[2].Y = 180, 180
		m.setSelection([]string{"block-1", "connector-2", "note-3"}, "note-3")
		m.scene.View.Zoom = 2.25
		m.scene.View.Center = canvas.Point{X: 321, Y: -19}
		before := slices.Clone(m.scene.Objects)
		view := m.scene.View
		if err := m.arrangeSelected("distribute-" + axis); err != nil {
			t.Fatal(err)
		}
		after := slices.Clone(m.scene.Objects)
		var gap1, gap2 float64
		if axis == "h" {
			gap1 = after[2].X - (after[1].X + after[1].Width)
			gap2 = after[0].X - (after[2].X + after[2].Width)
		} else {
			gap1 = after[2].Y - (after[1].Y + after[1].Height)
			gap2 = after[0].Y - (after[2].Y + after[2].Height)
		}
		if math.Abs(gap1-gap2) > 1e-10 || gap1 < 0 || m.scene.View != view || m.journal.Snapshot().UndoCount != 4 {
			t.Fatal("unequal gaps or camera changed", gap1, gap2)
		}
		for i, o := range after {
			if o.ID != before[i].ID || o.Width != before[i].Width || o.Height != before[i].Height || (axis == "h" && o.Y != before[i].Y) || (axis == "v" && o.X != before[i].X) {
				t.Fatal("paint order/dimensions/cross-axis changed")
			}
		}
		_ = m.journal.Undo()
		if !slices.Equal(m.scene.Objects, before) {
			t.Fatal("distribution Undo")
		}
		_ = m.journal.Redo()
		if !slices.Equal(m.scene.Objects, after) {
			t.Fatal("distribution Redo")
		}
		_ = m.arrangeSelected("distribute-" + axis)
		if m.journal.Snapshot().UndoCount != 4 {
			t.Fatal("no-op distribution history")
		}
	}
}
func TestArrangementRefusalsAtomicAndAvailability(t *testing.T) {
	m := editModel(t)
	if m.commands.Enabled("align-left") == nil || m.commands.Enabled("distribute-h") == nil {
		t.Fatal("too-few selection enabled")
	}
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	if m.commands.Enabled("align-left") != nil || m.commands.Enabled("distribute-h") == nil {
		t.Fatal("wrong minimum")
	}
	m.scene.Objects[0].X = 10000
	m.scene.Objects[0].Width = 2000
	before := slices.Clone(m.scene.Objects)
	if m.arrangeSelected("align-right") == nil || !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != 3 {
		t.Fatal("partially applied invalid origin")
	}
	m = editModel(t)
	m.scene.Objects[1].X = 10
	m.scene.Objects[2].X = 20
	m.setSelection([]string{"block-1", "connector-2", "note-3"}, "note-3")
	before = slices.Clone(m.scene.Objects)
	if m.arrangeSelected("distribute-h") == nil || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("crowded distribution accepted")
	}
	m.move = &objectMove{}
	if m.commands.Enabled("align-top") == nil {
		t.Fatal("gesture not gated")
	}
}

func TestDistributionTiesAndFractionalTightFit(t *testing.T) {
	objects := []canvas.Rect{{ID: "a", Width: 10, Height: 10}, {ID: "b", Width: 10, Height: 10}, {ID: "c", X: 100, Width: 10, Height: 10}}
	after, err := arrangedObjects(objects, []string{"c", "b", "a"}, "distribute-h")
	if err != nil || after[0].X != 0 || after[1].X != 50 || after[2].X != 100 {
		t.Fatal("paint-order ties not stable", after, err)
	}
	start := 1000.1
	objects = []canvas.Rect{{ID: "a", X: start, Width: 9.1, Height: 10}, {ID: "b", X: start + 9.1, Width: 13.4, Height: 10}, {ID: "c", X: start + 9.1 + 13.4, Width: 8.2, Height: 10}}
	after, err = arrangedObjects(objects, []string{"a", "b", "c"}, "distribute-h")
	if err != nil {
		t.Fatal("fractional tightly packed set rejected", err)
	}
	for i := 0; i < 2; i++ {
		if math.Abs(after[i+1].X-(after[i].X+after[i].Width)) > 1e-9 {
			t.Fatal("tight gaps", after)
		}
	}
}
