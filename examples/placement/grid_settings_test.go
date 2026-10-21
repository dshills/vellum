package main

import (
	"context"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"testing"
)

func TestGridVisibilityIndependentAndSessionOnly(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		root := view(f.a, f.ref, f.m)
		state := f.m.journal.Snapshot()
		if geometryDialogControl(t, root, "drawing").CanvasGrid != 20 {
			t.Fatal("default grid")
		}
		f.m.snap = false
		root = view(f.a, f.ref, f.m)
		if geometryDialogControl(t, root, "drawing").CanvasGrid != 20 {
			t.Fatal("snap hid grid")
		}
		geometryDialogControl(t, root, "grid-visible").Change(ui.ChangeEvent{Checked: false})
		if geometryDialogControl(t, view(f.a, f.ref, f.m), "drawing").CanvasGrid != 0 {
			t.Fatal("grid did not hide")
		}
		if err := f.m.setGridSpacing("grid-50"); err != nil {
			t.Fatal(err)
		}
		if f.m.snap || f.m.spacing() != 50 || state != f.m.journal.Snapshot() {
			t.Fatal("settings changed history/snap")
		}
		saved, _ := drawingCodec(f.m).Capture()
		n, err := newModel()
		if err != nil {
			t.Fatal(err)
		}
		defer n.palette.Close()
		if err = drawingCodec(n).Install(saved); err != nil || n.gridHidden || n.spacing() != 20 {
			t.Fatal("session settings leaked", err)
		}
		f.m.move = &objectMove{}
		if f.m.setGridSpacing("grid-5") == nil {
			t.Fatal("changed spacing during gesture")
		}
		f.m.move = nil
		if f.m.setGridSpacing("invalid") == nil || f.m.spacing() != 50 {
			t.Fatal("invalid spacing")
		}
	})
}
func TestConfiguredGridPlacementMoveResizeUndo(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer m.palette.Close()
	_ = m.setGridSpacing("grid-50")
	if err = m.palette.Source("block").PlaceKeyboard(context.Background(), m.target, canvas.Point{X: 33, Y: -27}); err != nil {
		t.Fatal(err)
	}
	if m.scene.Objects[0].X != 50 || m.scene.Objects[0].Y != -50 {
		t.Fatal("placement grid", m.scene.Objects)
	}
	before := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	if err = m.nudge(1, 0); err != nil || m.scene.Objects[0].X != 100 {
		t.Fatal("grid nudge", err)
	}
	if err = m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("nudge Undo", err)
	}
	if err = m.journal.Redo(); err != nil || m.scene.Objects[0].X != 100 {
		t.Fatal("nudge Redo", err)
	}
	_ = m.journal.Undo()
	g := placement.Geometry{Window: "placement", Scale: 2, Origin: canvas.Point{X: 20, Y: 40}, Clip: canvas.Bounds{Min: canvas.Point{}, Max: canvas.Point{X: 2000, Y: 1500}}, View: canvas.Viewport{Width: 800, Height: 600, Zoom: 2}}
	local := g.View.Screen(canvas.Point{X: 60, Y: -40})
	p := canvas.Point{X: g.Origin.X + local.X*g.Scale, Y: g.Origin.Y + local.Y*g.Scale}
	s := m.beginMove(g, p)
	if s == nil {
		t.Fatal("move begin")
	}
	end := canvas.Point{X: p.X + 120, Y: p.Y + 120}
	if err = m.finishMove(s, end); err != nil || m.scene.Objects[0].X != 100 || m.scene.Objects[0].Y != 0 {
		t.Fatal("transformed grid move", err, m.scene.Objects)
	}
	if m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("exactly one grid history edit")
	}
	_ = m.journal.Undo()
	if !slices.Equal(before, m.scene.Objects) {
		t.Fatal("grid move undo")
	}
	if err = m.resizeSelected(1, 0); err != nil || m.scene.Objects[0].Width != 150 || m.scene.Objects[0].Height != 50 {
		t.Fatal("grid resize", err, m.scene.Objects)
	}
}
func TestConfiguredGridCornerAndAspect(t *testing.T) {
	before := canvas.Rect{ID: "block-1", X: 10, Y: 20, Width: 80, Height: 50}
	unlocked := resizedCornerGrid(before, southEast, 31, 0, true, 50, false)
	if unlocked.Width != 100 || unlocked.Height != 50 || unlocked.X != 10 || unlocked.Y != 20 {
		t.Fatal("corner grid", unlocked)
	}
	after, err := resizedSelectionGrid([]canvas.Rect{before}, []string{before.ID}, southEast, 31, 0, true, 50, true)
	if err != nil || after[0].Width != 100 || math.Abs(after[0].Height-62.5) > 1e-9 {
		t.Fatal("aspect grid", after, err)
	}
	for _, n := range []float64{5, 10, 20, 50, 100} {
		if gridUnit(n) != n {
			t.Fatal("spacing")
		}
	}
	if gridUnit(math.NaN()) != 20 {
		t.Fatal("invalid fallback")
	}
}
