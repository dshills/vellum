package main

import (
	"context"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

func editModel(t *testing.T) *model {
	t.Helper()
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.palette.Close)
	for i, kind := range []string{"block", "connector", "note"} {
		if err := m.palette.Source(kind).PlaceKeyboard(context.Background(), m.target, canvas.Point{X: float64(i * 200)}); err != nil {
			t.Fatal(err)
		}
	}
	return m
}
func TestMoveAndDeleteHistoryPreservesIdentityOrderAndBranch(t *testing.T) {
	m := editModel(t)
	m.scene.Selected = "connector-2"
	original := slices.Clone(m.scene.Objects)
	view, next := m.scene.View, m.next
	if err := m.commands.Invoke(context.Background(), "move-right"); err != nil {
		t.Fatal(err)
	}
	if m.scene.Objects[1].X != original[1].X+20 || m.journal.Snapshot().UndoCount != 4 || m.scene.View != view || m.next != next {
		t.Fatal("move did not use one history change")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, original) {
		t.Fatal("move undo", err)
	}
	if err := m.journal.Redo(); err != nil || m.scene.Objects[1].ID != "connector-2" {
		t.Fatal("move redo", err)
	}
	moved := slices.Clone(m.scene.Objects)
	if err := m.commands.Invoke(context.Background(), "delete"); err != nil {
		t.Fatal(err)
	}
	if len(m.scene.Objects) != 2 || m.scene.Selected != "" || m.next != next || m.scene.Objects[1].ID != "note-3" {
		t.Fatal("delete changed IDs/order/counter")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, moved) || m.scene.Selected != "connector-2" {
		t.Fatal("delete undo lost original paint order or selection", err)
	}
	if err := m.journal.Redo(); err != nil || len(m.scene.Objects) != 2 {
		t.Fatal("delete redo", err)
	}
	if err := m.journal.Undo(); err != nil {
		t.Fatal(err)
	}
	m.snap = false
	if err := m.nudge(0, -1); err != nil || m.scene.Objects[1].Y != moved[1].Y-1 || m.journal.Snapshot().RedoCount != 0 {
		t.Fatal("fine move did not branch history", err)
	}
	m.scene.Selected = ""
	if m.commands.Enabled("delete") == nil || m.commands.Enabled("move-left") == nil {
		t.Fatal("editing enabled without selection")
	}
}

func TestTransformedDragSingleCommitCancelAndNoop(t *testing.T) {
	m := editModel(t)
	m.snap = false
	g := placement.Geometry{Window: "placement", Origin: canvas.Point{X: 10, Y: 20}, Clip: canvas.Bounds{Min: canvas.Point{X: 10, Y: 20}, Max: canvas.Point{X: 1610, Y: 1220}}, Scale: 2, View: canvas.Viewport{Width: 800, Height: 600, Zoom: 2, Center: canvas.Point{X: 30, Y: -20}}}
	local := g.View.Screen(canvas.Point{X: 10, Y: 10})
	start := canvas.Point{X: g.Origin.X + local.X*g.Scale, Y: g.Origin.Y + local.Y*g.Scale}
	original := slices.Clone(m.scene.Objects)
	base := m.journal.Snapshot().UndoCount
	s := m.beginMove(g, start)
	if s == nil || s.id != "block-1" {
		t.Fatal("transformed hit test failed")
	}
	if _, err := drawingCodec(m).Capture(); err == nil {
		t.Fatal("save captured a transient move")
	}
	for n := 1; n <= 20; n++ {
		m.previewMove(s, canvas.Point{X: start.X + float64(n*8), Y: start.Y + float64(n*4)})
	}
	if m.journal.Snapshot().UndoCount != base || m.scene.Objects[0].X != 40 || m.scene.Objects[0].Y != 20 {
		t.Fatal("preview added history or used wrong coordinates")
	}
	end := canvas.Point{X: start.X + 160, Y: start.Y + 80}
	if err := m.finishMove(s, end); err != nil {
		t.Fatal(err)
	}
	if m.move != nil || m.journal.Snapshot().UndoCount != base+1 {
		t.Fatal("drag did not commit exactly once")
	}
	if err := m.finishMove(s, end); err != nil || m.journal.Snapshot().UndoCount != base+1 {
		t.Fatal("duplicate release committed")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, original) {
		t.Fatal("drag undo", err)
	}
	for _, mode := range []string{"click", "escape", "outside", "snap-noop"} {
		t.Run(mode, func(t *testing.T) {
			m.snap = mode == "snap-noop"
			s := m.beginMove(g, start)
			if s == nil {
				t.Fatal("begin")
			}
			switch mode {
			case "click":
				_ = m.finishMove(s, canvas.Point{X: start.X + 4, Y: start.Y})
			case "escape":
				m.previewMove(s, end)
				m.cancelMove(s)
			case "outside":
				m.previewMove(s, end)
				_ = m.finishMove(s, canvas.Point{X: 0, Y: 0})
			case "snap-noop":
				_ = m.finishMove(s, canvas.Point{X: start.X + 16, Y: start.Y}) // four world units round back to origin.
			}
			if m.move != nil || !slices.Equal(m.scene.Objects, original) || m.journal.Snapshot().UndoCount != base {
				t.Fatal("canceled/no-op move changed document")
			}
		})
	}
}

func TestEditingBoundsAndOperationGuards(t *testing.T) {
	m := editModel(t)
	m.scene.Selected = "block-1"
	m.scene.Objects[0].X = 10000
	before := slices.Clone(m.scene.Objects)
	n := m.journal.Snapshot().UndoCount
	if m.nudge(1, 0) == nil || !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot().UndoCount != n {
		t.Fatal("invalid move changed drawing/history")
	}
	m.prompt = "open"
	if m.deleteSelected() == nil || m.nudge(-1, 0) == nil {
		t.Fatal("edit bypassed discard decision")
	}
	m.prompt = ""
	m.move = &objectMove{}
	if m.commands.Enabled("undo") == nil || m.commands.Enabled("redo") == nil || m.deleteSelected() == nil {
		t.Fatal("edit interleaved with drag")
	}
}

func TestSavedDrawingContainsMovesAndDeletion(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200, Y: 100})
	var saved drawingFile
	f.owner(t, func() {
		if err := f.m.nudge(-1, 0); err != nil {
			t.Fatal(err)
		}
		f.m.scene.Selected = "block-1"
		if err := f.m.deleteSelected(); err != nil {
			t.Fatal(err)
		}
		var err error
		saved, err = drawingCodec(f.m).Capture()
		if err != nil {
			t.Fatal(err)
		}
		if err = f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := drawingCodec(f.m).Decode(data)
	if err != nil || !reflect.DeepEqual(decoded, saved) || len(decoded.Objects) != 1 || decoded.Objects[0].ID != "note-2" || decoded.Objects[0].X != 180 || decoded.NextID != 2 {
		t.Fatal("edited drawing did not persist", err, decoded)
	}
}

func TestDragLayoutSettlementAndCameraChange(t *testing.T) {
	m := editModel(t)
	m.snap = false
	g := placement.Geometry{Window: "placement", Origin: canvas.Point{}, Clip: canvas.Bounds{Min: canvas.Point{}, Max: canvas.Point{X: 800, Y: 600}}, Scale: 1, View: canvas.Viewport{Width: 800, Height: 600, Zoom: 1}}
	start := canvas.Point{X: 410, Y: 310}
	move := m.beginMove(g, start)
	if move == nil {
		t.Fatal("begin")
	}
	g.Origin.Y = 18
	g.View.Height = 582
	g.Clip.Min.Y = 18
	if !m.updateMoveGeometry(move, g) {
		t.Fatal("layout settlement canceled")
	}
	if err := m.finishMove(move, canvas.Point{X: 450, Y: 330}); err != nil || m.scene.Objects[0].X != 40 || m.scene.Objects[0].Y != 11 {
		t.Fatal("layout settlement changed grab offset", err)
	}
	start = canvas.Point{X: 450, Y: 330}
	move = m.beginMove(g, start)
	if move == nil {
		t.Fatal("begin again")
	}
	m.previewMove(move, canvas.Point{X: 480, Y: 350})
	g.View.Zoom = 2
	if m.updateMoveGeometry(move, g) || m.move != nil || m.scene.Objects[0].X != 40 {
		t.Fatal("camera change retained transient move")
	}
}
