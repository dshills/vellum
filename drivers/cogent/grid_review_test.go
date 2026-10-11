//go:build offscreen

package cogent

import (
	"errors"
	"fmt"
	"image"
	"math"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestReviewGridEmptyHugeViewportRefusedBeforeMount(t *testing.T) {
	for _, rows := range []int{0, 1} {
		a, host := gridContractOwner(t, 8)
		source := &nativeGridSource{rows: rows, revision: 1}
		g, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: 80}}, virtual.GridConfig{RowHeight: 24}, virtual.GridViewport{Width: math.MaxInt, Height: math.MaxInt})
		if err != nil {
			t.Fatal("portable fixture invalid", err)
		}
		if err = validateNativeGrid(ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}); !errors.Is(err, ErrGridBudget) {
			t.Errorf("rows=%d huge local viewport admitted by native preallocation guard: %v", rows, err)
		}
		var r *record
		if err = a.Do(func() { r, err = reconcile(host, nil, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
			t.Fatal(err)
		}
		if r != nil {
			disposeControl(r)
			r.widget.AsTree().Delete()
		}
		if len(host.Children) != 0 {
			t.Fatal("mount cleanup incomplete")
		}
	}
}

func TestReviewGridMovedHostRetiresArmedSceneRect(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	changes := 0
	spacerHeight := float32(20)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
			{ID: "space", Kind: ui.Label, Text: "Space", Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 400, Height: spacerHeight}), MaxSize: ui.Some(ui.Size{Width: 400, Height: spacerHeight})}},
			{ID: "g", Kind: ui.DataGrid, Grid: g, GridChange: func(ui.GridEvent) { changes++ }, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 240}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 240})}},
		}}
	})
	gridFrame(t, m)
	first := gridInspection(t, a, g)
	beforeChanges := changes
	var cell GridCellGeometry
	for _, c := range first.Cells {
		if !c.Bounds.Empty() {
			cell = c
			break
		}
	}
	if cell.ID == "" {
		t.Fatal("no visible initial cell")
	}
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snap.Windows[0]
	pos := cell.Bounds.Min.Add(cell.Bounds.Size().Div(2))
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.MouseDown, Button: events.Left, Position: pos}); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Step(testContext, 1); err != nil {
		t.Fatal(err)
	}
	if err = a.Do(func() { spacerHeight = 60; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	after := gridInspection(t, a, g)
	if first.Generation != after.Generation || first.Viewport != after.Viewport {
		t.Fatal("fixture changed logical grid generation/viewport", first.Generation, after.Generation, first.Viewport, after.Viewport)
	}
	var moved GridCellGeometry
	for _, c := range after.Cells {
		if c.ID == cell.ID {
			moved = c
			break
		}
	}
	if moved.Bounds == cell.Bounds || moved.Bounds.Empty() {
		t.Fatal("fixture did not move cell", cell.Bounds, moved.Bounds)
	}
	pos = moved.Bounds.Min.Add(moved.Bounds.Size().Div(2))
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.MouseUp, Button: events.Left, Position: pos}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	if changes != beforeChanges {
		t.Fatalf("moved scene rectangle reinterpreted old press: changes=%d before=%d old=%v new=%v", changes, beforeChanges, cell.Bounds, moved.Bounds)
	}
}

func TestReviewGridFractionalGlobalEdges(t *testing.T) {
	for _, dpi := range []float32{120, 144} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			source := &nativeGridSource{rows: 10000, revision: 1}
			columns := []virtual.GridColumn{{ID: "c0", Width: 91}, {ID: "c1", Width: 91}, {ID: "c2", Width: 91}, {ID: "c3", Width: 91}, {ID: "c4", Width: 91}, {ID: "c5", Width: 91}, {ID: "c6", Width: 91}, {ID: "c7", Width: 91}}
			g, err := virtual.NewGrid(source, columns, virtual.GridConfig{RowHeight: 25, HeaderHeight: 27, FrozenColumns: 1, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
			if err != nil {
				t.Fatal(err)
			}
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
			gridFrame(t, m)
			if err = a.Do(func() {
				if err := g.Scroll(1, 1); err != nil {
					t.Error(err)
				}
				_ = a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			measured := gridInspection(t, a, g)
			if err = a.Do(func() {
				binding := m.implementation.find("one").state.root.control.grid
				edge := func(n int) int { return int(math.Round(float64(n) * float64(measured.Scale))) }
				for i, c := range measured.Cells {
					clip := binding.projection.Cells[i].Clip
					if clip.Empty() {
						continue
					}
					want := image.Rect(measured.Bounds.Min.X+edge(clip.X), measured.Bounds.Min.Y+edge(clip.Y), measured.Bounds.Min.X+edge(clip.X+clip.Width), measured.Bounds.Min.Y+edge(clip.Y+clip.Height))
					if c.Bounds != want {
						t.Errorf("shared global edge mismatch scale=%v row=%d col=%d got=%v want=%v clip=%+v", measured.Scale, c.Row, c.Column, c.Bounds, want, clip)
						return
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReviewGridFiniteExtremeWheelDoesNotPoisonRemainder(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	config := manualConfig(t, 7)
	config.LogicalDPI = 48
	m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	measured := gridInspection(t, a, g)
	pos := measured.Bounds.Min.Add(measured.Bounds.Size().Div(2))
	// Producer path mirrors a native backend event, avoiding only the public
	// manual fixture's deliberately narrower +-1e6 input convenience guard.
	m.implementation.find("one").state.source.Load().Scroll(pos, math32.Vector2{X: math.MaxFloat32})
	gridFrame(t, m)
	if err := a.Do(func() {
		b := m.implementation.find("one").state.root.control.grid
		if math.IsNaN(b.wheelX) || math.IsInf(b.wheelX, 0) {
			t.Errorf("finite native wheel poisoned remainder: %v", b.wheelX)
		}
	}); err != nil {
		t.Fatal(err)
	}
}
