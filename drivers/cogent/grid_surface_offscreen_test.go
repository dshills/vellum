//go:build offscreen

package cogent

import (
	"bytes"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/paint/render"
	"cogentcore.org/core/styles/units"
	"errors"
	"fmt"
	"image"
	"image/color"
	"reflect"
	"strconv"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/renderers/rasterx"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestGridCellSurfaceOriginalPixelGeometryAndActiveCell(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 162, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			model, _ := testNativeGrid(t, 10000)
			cfg := manualConfig(t, 7)
			cfg.LogicalDPI = dpi
			m, a := gridManual(t, cfg, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			before := gridInspection(t, a, model)
			var original *image.RGBA
			var g *gridBinding
			var oldWidgets int
			if err := a.Do(func() {
				window := m.implementation.find("one").state
				g = window.root.control.grid
				original, _ = gridPaintPixels(window.body.Scene, false)
				window.body.WidgetWalkDown(func(_ core.Widget, _ *core.WidgetBase) bool { oldWidgets++; return true })
				snap, err := model.Snapshot()
				if err != nil {
					t.Fatal(err)
				}
				window.body.Scene.SetProperty(gridCellSurfaceProperty, false)
				if !g.wantsCellSurface(snap) {
					t.Fatalf("surface not eligible layout=%+v projection=%v global=%v", g.layout, g.projectionReady, g.host.Scene.GlobalLayoutPending())
				}
				if err := g.syncCellSurface(snap, true); err != nil {
					t.Fatal(err)
				}
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			after := gridInspection(t, a, model)
			before.Frames, after.Frames = 0, 0
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("surface geometry differs\nbefore=%+v\nafter=%+v", before, after)
			}
			if err := a.Do(func() {
				window := m.implementation.find("one").state
				if g.cellSurface == nil || len(g.cells) != 0 || len(g.cellSurface.cells) != len(g.projection.Cells) {
					t.Fatal("body widgets were not replaced with projection values")
				}
				widgets := 0
				window.body.WidgetWalkDown(func(_ core.Widget, _ *core.WidgetBase) bool { widgets++; return true })
				if oldWidgets-widgets != len(g.cellSurface.cells) {
					t.Fatalf("removed widgets %d, want %d", oldWidgets-widgets, len(g.cellSurface.cells))
				}
				actual, _ := gridPaintPixels(window.body.Scene, false)
				if !bytes.Equal(original.Pix, actual.Pix) {
					for i := range actual.Pix {
						if original.Pix[i] != actual.Pix[i] {
							t.Fatalf("surface pixels differ offset=%d xy=%d,%d old=%d new=%d", i, (i%actual.Stride)/4, (i / actual.Stride), original.Pix[i], actual.Pix[i])
						}
					}
				}
				window.body.Scene.RenderWidget()
				commands := window.body.Scene.Painter.RenderDone()
				r := rasterx.New(math32.FromPoint(window.body.Scene.SceneGeom.Size)).(*rasterx.Renderer)
				packet := r.ScenePacket(commands)
				pixels := image.NewRGBA(actual.Bounds())
				if packet == nil || !packet.Replay(pixels) || !bytes.Equal(pixels.Pix, original.Pix) {
					t.Fatal("packet replay differs from original widget pixels")
				}
				t.Log("native body widgets removed", oldWidgets-widgets, "retained value layouts", len(g.cellSurface.cells))
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func surfaceBinding(t *testing.T, m *Manual, a *app.Application) *gridBinding {
	t.Helper()
	var g *gridBinding
	if err := a.Do(func() { g = m.implementation.find("one").state.root.control.grid }); err != nil {
		t.Fatal(err)
	}
	return g
}

func surfaceEnable(t *testing.T, m *Manual, a *app.Application, g *gridBinding) {
	t.Helper()
	if err := a.Do(func() {
		g.host.Scene.SetProperty(gridCellSurfaceProperty, false)
		snap, e := g.model.Snapshot()
		if e != nil {
			t.Fatal(e)
		}
		if !g.wantsCellSurface(snap) {
			t.Fatal("surface not eligible")
		}
		if e = g.syncCellSurface(snap, true); e != nil {
			t.Fatal(e)
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
}

func surfaceCompareWithOriginal(t *testing.T, m *Manual, a *app.Application, g *gridBinding) {
	t.Helper()
	actualInspection := gridInspection(t, a, g.model)
	var actual *image.RGBA
	if err := a.Do(func() {
		if g.cellSurface == nil || len(g.cells) != 0 {
			t.Fatal("value route missing")
		}
		actual, _ = gridPaintPixels(g.host.Scene, false)
		g.host.Scene.SetProperty(gridCellSurfaceProperty, true)
		if e := g.owner.Invalidate(g.window); e != nil {
			t.Fatal(e)
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	reference := gridInspection(t, a, g.model)
	actualInspection.Frames, reference.Frames = 0, 0
	if !reflect.DeepEqual(actualInspection, reference) {
		t.Fatal("value/original geometry differs after transition")
	}
	if err := a.Do(func() {
		if g.cellSurface != nil || len(g.cells) != len(g.projection.Cells) {
			t.Fatal("ordinary fallback did not restore native cells")
		}
		pixels, _ := gridPaintPixels(g.host.Scene, false)
		if !bytes.Equal(pixels.Pix, actual.Pix) {
			for i := range actual.Pix {
				if actual.Pix[i] != pixels.Pix[i] {
					t.Fatalf("value/original pixels differ offset=%d xy=%d,%d value=%d original=%d", i, (i%actual.Stride)/4, i/actual.Stride, actual.Pix[i], pixels.Pix[i])
				}
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGridCellSurfaceScrollRetentionAndDisableFallback(t *testing.T) {
	model, _ := testNativeGrid(t, 10000)
	disabled := false
	cfg := manualConfig(t, 7)
	cfg.LogicalDPI = 162
	m, a := gridManual(t, cfg, func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Disabled: disabled}
	})
	gridFrame(t, m)
	g := surfaceBinding(t, m, a)
	for _, delta := range []image.Point{{0, 12}, {90, 12}, {-30, -5}, {0, 24}} {
		surfaceEnable(t, m, a, g)
		old := map[gridCellKey]*shaped.Lines{}
		if e := a.Do(func() {
			for _, entry := range g.cellSurface.cells {
				old[entry.key] = entry.shape
			}
			if e := model.Scroll(delta.X, delta.Y); e != nil {
				t.Fatal(e)
			}
			_ = a.Invalidate("one")
		}); e != nil {
			t.Fatal(e)
		}
		gridFrame(t, m)
		if e := a.Do(func() {
			if g.cellSurface == nil || len(g.cells) != 0 {
				t.Fatal("ordinary scroll fell back")
			}
			hits := 0
			for _, entry := range g.cellSurface.cells {
				if old[entry.key] != nil && old[entry.key] == entry.shape {
					hits++
				}
			}
			if hits == 0 {
				t.Fatal("unchanged current-projection shapes were not retained")
			}
		}); e != nil {
			t.Fatal(e)
		}
		surfaceCompareWithOriginal(t, m, a, g)
	}
	surfaceEnable(t, m, a, g)
	if e := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if g.cellSurface != nil || len(g.cells) == 0 {
		t.Fatal("disabled transition failed ordinary fallback")
	}
	if e := a.Do(func() { disabled = false; _ = a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	surfaceEnable(t, m, a, g)
	owned := g.cellSurface
	if e := m.Close(testContext); e != nil {
		t.Fatal(e)
	}
	if g.cellSurface != nil || len(owned.cells) != 0 {
		t.Fatal("retired layout/font/text references retained")
	}
}

type surfaceEditableSource struct {
	selectionNativeSource
	value string
}

func (s *surfaceEditableSource) Cell(row, col string) (string, error) {
	if row == "selection-0" && col == "c1" {
		return s.value, nil
	}
	return s.selectionNativeSource.Cell(row, col)
}
func (s *surfaceEditableSource) EditCell(row, col string) (virtual.GridEditValue, error) {
	value, err := s.Cell(row, col)
	return virtual.GridEditValue{Text: value, Version: 1, Editable: true}, err
}

func TestGridCellSurfaceRowSelectionCursorAndEditFallback(t *testing.T) {
	source := &surfaceEditableSource{selectionNativeSource: selectionNativeSource{selectionNativeIndex: selectionNativeIndex{total: 10000, revision: 1}}, value: "A😀é <b>\tאבג\nsecond"}
	model, e := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c0", Width: 121}, {ID: "c1", Width: 121}, {ID: "c2", Width: 121}}, virtual.GridConfig{RowHeight: 23, HeaderHeight: 27, FrozenColumns: 1, RowSelection: true, CellEditing: true, MaxEditBytes: 128}, virtual.GridViewport{Width: 300, Height: 180})
	if e != nil {
		t.Fatal(e)
	}
	cfg := manualConfig(t, 7)
	cfg.LogicalDPI = 162
	m, a := gridManual(t, cfg, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
	gridFrame(t, m)
	g := surfaceBinding(t, m, a)
	if e = a.Do(func() {
		if e := model.SelectToKey("selection-0", "c1", virtual.GridSelectReplace); e != nil {
			t.Fatal(e)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	surfaceEnable(t, m, a, g)
	// Rendering backgrounds must not mutate style templates and invalidate shapes.
	if err := a.Do(func() {
		before := make([]*shaped.Lines, len(g.cellSurface.cells))
		for i, c := range g.cellSurface.cells {
			before[i] = c.shape
		}
		g.host.Scene.RenderWidget()
		snap, err := model.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		if err = g.syncCellSurface(snap, true); err != nil {
			t.Fatal(err)
		}
		g.host.ApplyScenePos()
		for i, c := range g.cellSurface.cells {
			if c.shape != before[i] {
				t.Fatal("selected/cursor shape invalidated by rendering", i)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	surfaceCompareWithOriginal(t, m, a, g)
	surfaceEnable(t, m, a, g)
	if e = a.Do(func() {
		if e := model.SelectToKey("selection-0", "c2", virtual.GridSelectReplace); e != nil {
			t.Fatal(e)
		}
		if e := model.BeginActiveEdit(); e != nil {
			t.Fatal(e)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if g.cellSurface != nil || g.editing == nil || len(g.cells) == 0 {
		t.Fatal("editor did not take original native route")
	}
	if e = a.Do(func() {
		if e := model.CancelEdit(); e != nil {
			t.Fatal(e)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	surfaceEnable(t, m, a, g)
	surfaceCompareWithOriginal(t, m, a, g)
}

func TestGridCellSurfaceRoutedInputAndStalePress(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	changes := 0
	config := manualConfig(t, 7)
	m, a := gridManual(t, config, func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(ui.GridEvent) { changes++ }}
	})
	gridFrame(t, m)
	binding := surfaceBinding(t, m, a)
	surfaceEnable(t, m, a, binding)
	inspection := gridInspection(t, a, g)
	cell := inspection.Cells[0]
	s, e := m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	window := s.Windows[0]
	var target ManualElement
	for _, x := range window.Elements {
		if x.ID == cell.ID {
			target = x
		}
	}
	if target.ID == "" {
		t.Fatal("cell missing from actual measured capture")
	}
	down := ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Target: target.ID, TargetIncarnation: target.Incarnation, Type: events.MouseDown, Button: events.Left}
	if e = m.Input(testContext, down); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Step(testContext, 1); e != nil {
		t.Fatal(e)
	}
	before := changes
	if e = a.Do(func() {
		if err := g.Scroll(0, 24); err != nil {
			t.Error(err)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if binding.cellSurface == nil || len(binding.cells) != 0 {
		t.Fatal("value mode fell back during authority test")
	}
	up := down
	up.Type = events.MouseUp
	up.Target = ""
	up.Position = target.Bounds.Min.Add(target.Bounds.Size().Div(2))
	if e = m.Input(testContext, up); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if binding.cellSurface == nil || len(binding.cells) != 0 {
		t.Fatal("value mode fell back during authority test")
	}
	if changes != before {
		t.Fatal("recycled press changed selection")
	}
	s, e = m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	window = s.Windows[0]
	inspection = gridInspection(t, a, g)
	for _, c := range inspection.Cells {
		if !c.Bounds.Empty() {
			cell = c
			break
		}
	}
	for _, x := range window.Elements {
		if x.ID == cell.ID {
			target = x
		}
	}
	down.WindowIncarnation = window.Incarnation
	down.Target = target.ID
	down.TargetIncarnation = target.Incarnation
	if e = m.Input(testContext, down); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Step(testContext, 1); e != nil {
		t.Fatal(e)
	}
	up = down
	up.Type = events.MouseUp
	if e = m.Input(testContext, up); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if binding.cellSurface == nil || len(binding.cells) != 0 {
		t.Fatal("value mode fell back during authority test")
	}
	if changes != before+1 {
		t.Fatalf("actual click did not select: %d/%d", changes, before)
	}
	arrow := ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.KeyChord, Code: key.CodeDownArrow}
	if e = m.Input(testContext, arrow); e != nil {
		t.Fatal(e)
	}
	if e = m.Input(testContext, arrow); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if binding.cellSurface == nil || len(binding.cells) != 0 {
		t.Fatal("value mode fell back during authority test")
	}
	var state virtual.GridState
	var stateErr error
	if err := a.Do(func() { state, stateErr = g.State() }); err != nil {
		t.Fatal(err)
	}
	if stateErr != nil {
		t.Fatal(stateErr)
	}
	wantRow, _ := strconv.Atoi(cell.RowKey[4:])
	want := fmt.Sprintf("row-%d", wantRow+2)
	if state.Active.RowKey != want {
		t.Fatal("queued keys did not route through grid focus")
	}

}

func TestGridCellSurfaceFractionalWheelAndAX(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(ui.GridEvent) { changes++ }}
	})
	gridFrame(t, m)
	binding := surfaceBinding(t, m, a)
	surfaceEnable(t, m, a, binding)
	inspection := gridInspection(t, a, g)
	snapshot, e := m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	window := snapshot.Windows[0]
	pos := inspection.Bounds.Min.Add(image.Pt(200, 80))
	for range 5 {
		if e = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.Scroll, Position: pos, Delta: math32.Vector2{X: .25 * inspection.Scale, Y: .25 * inspection.Scale}}); e != nil {
			t.Fatal(e)
		}
		gridFrame(t, m)
		if binding.cellSurface == nil || len(binding.cells) != 0 {
			t.Fatal("value mode fell back during authority test")
		}
	}
	var state virtual.GridState
	var stateErr error
	if err := a.Do(func() { state, stateErr = g.State() }); err != nil {
		t.Fatal(err)
	}
	if stateErr != nil {
		t.Fatal(stateErr)
	}
	if state.Viewport.ScrollX != 1 || state.Viewport.ScrollY != 1 {
		t.Fatal("fractional wheel discarded", state.Viewport)
	}
	inspection = gridInspection(t, a, g)
	root := m.implementation.find("one").state.root
	var oldAction accessibilityHandler
	var bounds map[string][4]float32
	if e = a.Do(func() {
		nodes, handlers := accessibilitySnapshot(root, nil)
		bounds = accessibilityBounds(root)
		table := nodes[0]
		if table.Role != "table" || table.RowCount != 10000 || table.ColumnCount != 20 {
			t.Error("wrong full table metadata", table)
		}
		cells, headers := 0, 0
		for _, n := range nodes {
			if n.Role == "grid-cell" {
				cells++
				if n.RowIndex < 1 || n.ColumnIndex < 1 || n.Value == nil || bounds[n.ID][2] <= 0 {
					t.Error("invalid actual cell metadata", n)
				}
				if oldAction == nil {
					oldAction = handlers[n.ID]
				}
			}
			if n.Role == "grid-column" {
				headers++
			}
		}
		if cells == 0 || cells > len(inspection.Cells) || headers == 0 {
			t.Error("unbounded or missing visible table", cells, headers)
		}
	}); e != nil {
		t.Fatal(e)
	}
	beforeChanges := changes
	staleErrors := 0
	if e = a.Do(func() {
		root.control.onError = func(err error) {
			if !errors.Is(err, virtual.ErrGridStale) {
				t.Error(err)
			}
			staleErrors++
		}
		if e := g.Scroll(0, 24); e != nil {
			t.Error(e)
		}
		if oldAction == nil || oldAction(accessibilityAction{Kind: "press"}) {
			t.Error("stale AX pooled action admitted")
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if binding.cellSurface == nil || len(binding.cells) != 0 {
		t.Fatal("value mode fell back during authority test")
	}
	if changes != beforeChanges || staleErrors != 1 {
		t.Fatal("stale AX action notified selection or lost refusal")
	}
	var accepted bool
	if e = a.Do(func() {
		nodes, handlers := accessibilitySnapshot(root, nil)
		for _, node := range nodes {
			if node.Role != "grid-cell" {
				continue
			}
			if handler := handlers[node.ID]; handler != nil {
				accepted = handler(accessibilityAction{Kind: "press"})
				break
			}
		}
	}); e != nil {
		t.Fatal(e)
	}
	if !accepted || changes != beforeChanges+1 {
		t.Fatal("current AX cell action refused")
	}
	gridFrame(t, m)
	if binding.cellSurface == nil || len(binding.cells) != 0 {
		t.Fatal("value mode fell back during authority test")
	}
	inspection = gridInspection(t, a, g)
	if inspection.Frames < 2 {
		t.Fatal("frame counter did not advance")
	}
}

// Hides the optional batching capability while preserving ordinary rendering.
type surfaceNoBatchRenderer struct{ render.Renderer }

func TestGridCellSurfaceLateRefusalRecoversOrdinaryCells(t *testing.T) {
	for _, kind := range []string{"color", "indent", "renderer", "widget-init"} {
		t.Run(kind, func(t *testing.T) {
			model, _ := testNativeGrid(t, 10000)
			m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			g := surfaceBinding(t, m, a)
			surfaceEnable(t, m, a, g)
			var version uint64
			if err := a.Do(func() {
				version = g.projection.State.Generation
				p := g.panes[3]
				switch kind {
				case "color":
					p.OverrideStyle = true
					p.Styles.Color = image.NewRGBA(image.Rect(0, 0, 1, 1))
				case "indent":
					p.OverrideStyle = true
					p.Styles.Text.Indent = units.Pw(10)
				case "renderer":
					g.host.Scene.Renderer = surfaceNoBatchRenderer{g.host.Scene.Renderer}
				case "widget-init":
					g.host.Scene.WidgetInit = func(core.Widget) {}
				}
				// Force a late layout check before any ordinary owner sync.
				g.host.ApplyScenePos()
				if !g.cellSurface.failed || g.cellSurface.completed || g.measurement.Load() != nil {
					t.Fatal("late refusal did not revoke completed values")
				}
				g.host.Scene.RenderWidget()
				if g.cellSurface.completed {
					t.Fatal("refused values painted as complete")
				}
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			if err := a.Do(func() {
				if g.cellSurface != nil || len(g.cells) != len(g.projection.Cells) || len(g.cells) == 0 {
					t.Fatal("same-generation fallback did not reconstruct ordinary cells")
				}
				if g.projection.State.Generation != version {
					t.Fatal("fallback changed model generation")
				}
			}); err != nil {
				t.Fatal(err)
			}
			_ = gridInspection(t, a, model)
		})
	}
}

func TestGridCellSurfaceUnitAdmissionDoesNotRunCustom(t *testing.T) {
	calls := 0
	for _, v := range []units.Value{units.Ew(1), units.Eh(1), units.Pw(1), units.Ph(1), units.Em(1), units.Value{Custom: func(*units.Context) float32 { calls++; return 1 }}} {
		if gridSurfaceAbsolute(v) {
			t.Fatal("unsupported unit admitted", v)
		}
	}
	if calls != 0 {
		t.Fatal("custom conversion ran during refusal")
	}
}

type surfaceVariableSource struct {
	*nativeGridSource
	heights *virtual.RowHeights
}

func (s *surfaceVariableSource) RowOffset(i int) (int, error) { return s.heights.RowOffset(i) }
func (s *surfaceVariableSource) RowAt(i int) (int, error)     { return s.heights.RowAt(i) }

func TestGridCellSurfaceVariableFrozenEmptyAndTypography(t *testing.T) {
	for _, tc := range []struct {
		name         string
		rows, frozen int
		variable     bool
	}{
		{"no-frozen", 10000, 0, false}, {"variable", 10000, 2, true}, {"empty", 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &nativeGridSource{rows: tc.rows, revision: 1}
			var input virtual.GridSource = source
			if tc.variable {
				heights := make([]int, tc.rows)
				for i := range heights {
					heights[i] = 19 + i%7
				}
				h, e := virtual.NewRowHeights(heights)
				if e != nil {
					t.Fatal(e)
				}
				input = &surfaceVariableSource{nativeGridSource: source, heights: h}
			}
			cols := make([]virtual.GridColumn, 20)
			for i := range cols {
				cols[i] = virtual.GridColumn{ID: fmt.Sprint("c", i), Title: fmt.Sprint("column", i), Width: 77 + i%4}
			}
			model, e := virtual.NewGrid(input, cols, virtual.GridConfig{RowHeight: 23, HeaderHeight: 27, FrozenColumns: tc.frozen, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
			if e != nil {
				t.Fatal(e)
			}
			cfg := manualConfig(t, 7)
			cfg.LogicalDPI = 162
			m, a := gridManual(t, cfg, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			g := surfaceBinding(t, m, a)
			surfaceEnable(t, m, a, g)
			if e := a.Do(func() {
				for _, pane := range g.panes[2:] {
					pane.OverrideStyle = true
					pane.Styles.Font.Size = units.Dp(19)
					pane.Styles.Text.Indent = units.Px(3)
					pane.Styles.Text.Hanging = units.Px(2)
				}
				g.host.Update()
			}); e != nil {
				t.Fatal(e)
			}
			gridFrame(t, m)
			if g.cellSurface == nil {
				t.Fatal("supported typography lost value route")
			}
			surfaceCompareWithOriginal(t, m, a, g)
			if tc.rows > 0 {
				surfaceEnable(t, m, a, g)
				if e := a.Do(func() {
					if e := model.Scroll(103, 11); e != nil {
						t.Fatal(e)
					}
					_ = a.Invalidate("one")
				}); e != nil {
					t.Fatal(e)
				}
				gridFrame(t, m)
				surfaceCompareWithOriginal(t, m, a, g)
			}
		})
	}
}

func TestGridCellSurfaceRestyleOnlyRetiresCachedStyles(t *testing.T) {
	model, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
	gridFrame(t, m)
	g := surfaceBinding(t, m, a)
	surfaceEnable(t, m, a, g)
	var generation uint64
	if e := a.Do(func() {
		generation = g.projection.State.Generation
		pane := g.panes[3]
		pane.OverrideStyle = true
		pane.Styles.Color = colors.Uniform(color.RGBA{R: 211, G: 31, B: 42, A: 255})
		// Restyle requests rendering without layout; old templates cannot be painted.
		pane.Restyle()
		if !g.cellSurface.dirty[1] {
			t.Fatal("restyle did not dirty inherited styles")
		}
		g.host.Scene.RenderWidget()
		if !g.cellSurface.failed || g.measurement.Load() != nil || !g.layoutPending.Load() {
			t.Fatal("paint-only restyle accepted stale templates")
		}
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if e := a.Do(func() {
		if g.cellSurface != nil || len(g.cells) == 0 || g.projection.State.Generation != generation {
			t.Fatal("restyle-only fallback did not recover ordinary cells at same generation")
		}
	}); e != nil {
		t.Fatal(e)
	}
	_ = gridInspection(t, a, model)
	surfaceEnable(t, m, a, g)
	surfaceCompareWithOriginal(t, m, a, g)
}

func TestGridCellSurfaceDefaultManualRoute(t *testing.T) {
	model, _ := testNativeGrid(t, 10000)
	a, e := app.New(app.Config{QueueCapacity: 32})
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Open(app.Window{ID: "one", Title: "Grid", Width: 860, Height: 620, View: func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} }}); e != nil {
		t.Fatal(e)
	}
	m, e := NewManual(testContext, a, manualConfig(t, 7))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := m.Close(testContext); e != nil {
			t.Error(e)
		}
	})
	gridFrame(t, m)
	g := surfaceBinding(t, m, a)
	if e = a.Do(func() {
		if ref, _ := g.host.Scene.Property(gridCellSurfaceProperty).(bool); ref {
			t.Fatal("normal Manual silently forced legacy widgets")
		}
		if e := model.Scroll(0, 12); e != nil {
			t.Fatal(e)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if g.cellSurface == nil || !g.cellSurface.completed || len(g.cells) != 0 {
		t.Fatal("normal Manual scroll did not promote to completed value route")
	}
	surfaceCompareWithOriginal(t, m, a, g)
}

func TestGridCellSurfaceLateForeignChildrenRemainOwned(t *testing.T) {
	for _, part := range []bool{false, true} {
		t.Run(fmt.Sprint("parts=", part), func(t *testing.T) {
			model, _ := testNativeGrid(t, 10000)
			m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			g := surfaceBinding(t, m, a)
			surfaceEnable(t, m, a, g)
			var foreign core.Widget
			var generation uint64
			if e := a.Do(func() {
				pane := g.panes[3]
				generation = g.projection.State.Generation
				if part {
					p := core.NewFrame()
					tree.SetParent(p, pane) // Core parts have a parent without Children membership.
					pane.Parts = p
					foreign = p
					if len(pane.Children) != 0 {
						t.Fatal("Parts refusal is masked by native Children")
					}
				} else {
					foreign = core.NewText(pane).SetText("foreign native child")
				}
				// A foreign child added after admission cannot be silently omitted by value painting.
				g.host.ApplyScenePos()
				g.host.Scene.RenderWidget()
				if !g.cellSurface.failed || g.measurement.Load() != nil {
					t.Fatal("late foreign native ownership was admitted")
				}
			}); e != nil {
				t.Fatal(e)
			}
			gridFrame(t, m)
			if e := a.Do(func() {
				if g.cellSurface != nil || len(g.cells) != len(g.projection.Cells) || g.projection.State.Generation != generation {
					t.Fatal("foreign-child fallback did not restore original cells")
				}
				if foreign.AsWidget().This == nil || foreign.AsWidget().Parent != g.panes[3] {
					t.Fatal("fallback deleted or moved unowned native child")
				}
				snap, e := model.Snapshot()
				if e != nil {
					t.Fatal(e)
				}
				if g.wantsCellSurface(snap) {
					t.Fatal("existing foreign child readmitted")
				}
			}); e != nil {
				t.Fatal(e)
			}
			_ = gridInspection(t, a, model)
		})
	}
}

// Direct resolved-context changes bypass Style and its dirty flag. These are
// offscreen invalidation checks, not physical display-transition evidence.
func TestGridCellSurfaceResolvedContextInvalidatesWithoutDirtyFlag(t *testing.T) {
	for _, name := range []string{"dpi", "font-size", "font-family"} {
		t.Run(name, func(t *testing.T) {
			model, _ := testNativeGrid(t, 10000)
			m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			g := surfaceBinding(t, m, a)
			surfaceEnable(t, m, a, g)
			if e := a.Do(func() {
				p, surface := g.panes[3], g.cellSurface
				original := p.Styles
				generation, epoch := g.projection.State.Generation, surface.epoch[1]
				before := make(map[gridCellKey]*shaped.Lines)
				for _, cell := range surface.cells {
					if cell.pane == 3 && cell.text != "" {
						before[cell.key] = cell.shape
					}
				}
				if surface.dirty[1] || len(before) == 0 {
					t.Fatal("test requires settled nonempty value cells")
				}
				switch name {
				case "dpi":
					p.Styles.UnitContext.DPI += 32
				case "font-size":
					p.Styles.Font.Size.Value += 2
				case "font-family":
					if p.Styles.Font.Family == rich.Monospace {
						p.Styles.Font.Family = rich.Serif
					} else {
						p.Styles.Font.Family = rich.Monospace
					}
				}
				p.ApplyScenePos()
				if surface.failed || surface.dirty[1] || surface.epoch[1] <= epoch || g.projection.State.Generation != generation {
					t.Fatal("resolved context did not invalidate without Style or model change")
				}
				for _, cell := range surface.cells {
					if old, ok := before[cell.key]; ok && (cell.shape == nil || cell.shape == old || cell.shapeEpoch != surface.epoch[1]) {
						t.Fatal("resolved context retained stale shaped text")
					}
				}
				// A simulated parent DPI is not the scene renderer's physical DPI. Restore
				// it before checking original-route pixels and published geometry.
				p.Styles = original
				g.host.ApplyScenePos()
			}); e != nil {
				t.Fatal(e)
			}
			surfaceCompareWithOriginal(t, m, a, g)
		})
	}
}
