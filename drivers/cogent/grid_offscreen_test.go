//go:build offscreen

package cogent

import (
	"errors"
	"fmt"
	"image"
	"math"
	"strconv"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/abilities"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type nativeGridSource struct {
	rows                 int
	revision             uint64
	cells, keys, indexes int
	failCell             bool
}

func (s *nativeGridSource) Identity() string { return "native-grid" }
func (s *nativeGridSource) Revision() uint64 { return s.revision }
func (s *nativeGridSource) RowCount() int    { return s.rows }
func (s *nativeGridSource) RowKey(i int) (string, error) {
	s.keys++
	return fmt.Sprintf("row-%d", i), nil
}
func (s *nativeGridSource) IndexOf(k string) (int, bool, error) {
	s.indexes++
	if len(k) < 4 {
		return 0, false, nil
	}
	i, e := strconv.Atoi(k[4:])
	return i, e == nil && i >= 0 && i < s.rows, nil
}
func (s *nativeGridSource) Cell(k, c string) (string, error) {
	s.cells++
	if s.failCell {
		return "", errors.New("fixture cell refused")
	}
	return k + " · " + c, nil
}
func testNativeGrid(t *testing.T, rows int) (*virtual.Grid, *nativeGridSource) {
	t.Helper()
	s := &nativeGridSource{rows: rows, revision: 1}
	columns := make([]virtual.GridColumn, 20)
	for i := range columns {
		columns[i] = virtual.GridColumn{ID: fmt.Sprintf("c%d", i), Title: fmt.Sprintf("Column %d", i), Width: 90}
	}
	g, e := virtual.NewGrid(s, columns, virtual.GridConfig{RowHeight: 24, HeaderHeight: 26, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
	if e != nil {
		t.Fatal(e)
	}
	return g, s
}
func gridInspection(t *testing.T, a *app.Application, g *virtual.Grid) GridInspection {
	t.Helper()
	var out GridInspection
	var e error
	if err := a.Do(func() { out, e = InspectGrid(a, g) }); err != nil {
		t.Fatal(err)
	}
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func gridManual(t *testing.T, config ManualConfig, view func() ui.Node) (*Manual, *app.Application) {
	t.Helper()
	a, e := app.New(app.Config{QueueCapacity: 32})
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Open(app.Window{ID: "one", Title: "Grid", Width: 860, Height: 620, View: view}); e != nil {
		t.Fatal(e)
	}
	m, e := NewManual(testContext, a, config)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := m.Close(testContext); e != nil {
			t.Error(e)
		}
	})
	return m, a
}
func gridFrame(t *testing.T, m *Manual) { t.Helper(); unitSettle(t, m) }
func TestManagedGridActualGeometryAndPools(t *testing.T) {
	for _, dpi := range []float32{96, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			g, s := testNativeGrid(t, 1_000_000)
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
			gridFrame(t, m)
			first := gridInspection(t, a, g)
			if first.Scale != dpi/96 || len(first.Cells) == 0 || first.Columns != 20 || first.Stamp.Rows != 1_000_000 {
				t.Fatalf("bad geometry %+v", first)
			}
			for _, c := range first.Cells {
				if !c.Bounds.Empty() && !c.Bounds.In(first.Bounds) {
					t.Fatal("cell escaped content", c)
				}
			}
			if s.cells > gridNativeMaxCells*8 {
				t.Fatal("whole source materialized", s.cells)
			}
			beforeFrozen := first.Cells[0].Bounds
			for _, c := range first.Cells {
				if c.Column == 0 && !c.Bounds.Empty() {
					beforeFrozen = c.Bounds
					break
				}
			}
			if e := a.Do(func() {
				if e := g.Scroll(110, 48); e != nil {
					t.Error(e)
				}
				_ = a.Invalidate("one")
			}); e != nil {
				t.Fatal(e)
			}
			var pending error
			if e := a.Do(func() { _, pending = InspectGrid(a, g) }); e != nil {
				t.Fatal(e)
			}
			if !errors.Is(pending, ErrGridLayout) {
				t.Fatal("old frame accepted", pending)
			}
			gridFrame(t, m)
			after := gridInspection(t, a, g)
			if len(after.Cells) > gridNativeMaxCells {
				t.Fatal("pool became source-sized")
			}
			afterFrozen := after.Cells[0].Bounds
			for _, c := range after.Cells {
				if c.Column == 0 && !c.Bounds.Empty() {
					afterFrozen = c.Bounds
					break
				}
			}
			if afterFrozen.Min.X != beforeFrozen.Min.X {
				t.Fatal("frozen pane panned")
			}
			if after.Headers[0].Bounds.Min.Y != first.Headers[0].Bounds.Min.Y {
				t.Fatal("header moved")
			}
			if err := a.Do(func() {
				binding := m.implementation.find("one").state.root.control.grid
				for i, c := range after.Cells {
					logical := binding.projection.Cells[i].Clip
					if logical.Empty() {
						if !c.Bounds.Empty() {
							t.Error("overscan measured as visible", c)
						}
						continue
					}
					edge := func(v int) int { return int(math.Round(float64(float32(v) * after.Scale))) }
					expected := image.Rect(after.Bounds.Min.X+edge(logical.X), after.Bounds.Min.Y+edge(logical.Y), after.Bounds.Min.X+edge(logical.X+logical.Width), after.Bounds.Min.Y+edge(logical.Y+logical.Height))
					if c.Bounds != expected {
						t.Errorf("measured clip mismatch cell=%s got=%v expected=%v logical=%+v", c.ID, c.Bounds, expected, logical)
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
			img, e := m.Capture(testContext, "one")
			if e != nil || img.Bounds().Empty() {
				t.Fatal("actual composition unavailable", e)
			}
			var state virtual.GridState
			var stateErr error
			if err := a.Do(func() { state, stateErr = m.implementation.find("one").state.root.control.grid.model.State() }); err != nil {
				t.Fatal(err)
			}
			if stateErr != nil {
				t.Fatal(stateErr)
			}
			if state.Viewport.ScrollX != 110 || state.Viewport.ScrollY != 48 {
				t.Fatal(state.Viewport)
			}
		})
	}
}
func TestManagedGridActualInputAndStalePress(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	changes := 0
	config := manualConfig(t, 7)
	m, a := gridManual(t, config, func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(ui.GridEvent) { changes++ }}
	})
	gridFrame(t, m)
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
	up := down
	up.Type = events.MouseUp
	up.Target = ""
	up.Position = target.Bounds.Min.Add(target.Bounds.Size().Div(2))
	if e = m.Input(testContext, up); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
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

func TestManagedGridSourceScaleAndDetachedInspection(t *testing.T) {
	type counts struct{ cells, keys, indexes, pool int }
	var baseline counts
	for _, rows := range []int{10000, 1000000} {
		g, s := testNativeGrid(t, rows)
		m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
		gridFrame(t, m)
		inspection := gridInspection(t, a, g)
		if len(inspection.Cells) == 0 {
			t.Fatal("missing measured cells")
		}
		before := counts{s.cells, s.keys, s.indexes, len(m.implementation.find("one").state.root.control.grid.cells)}
		for range 3 {
			detached := gridInspection(t, a, g)
			detached.Cells[0].Text = "tampered"
			detached.Cells[0].Bounds = image.Rectangle{}
			detached.Headers[0].Title = "tampered"
		}
		after := gridInspection(t, a, g)
		if after.Cells[0] != inspection.Cells[0] || after.Headers[0] != inspection.Headers[0] || before != (counts{s.cells, s.keys, s.indexes, len(m.implementation.find("one").state.root.control.grid.cells)}) {
			t.Fatal("observation mutates/reads source or aliases frame")
		}
		if rows == 10000 {
			baseline = before
		} else if before != baseline {
			t.Fatal("native projection grows with source size", baseline, before)
		}
		if err := m.Close(testContext); err != nil {
			t.Fatal(err)
		}
	}
}
func TestManagedGridActualFractionalWheelAndAX(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(ui.GridEvent) { changes++ }}
	})
	gridFrame(t, m)
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
	inspection = gridInspection(t, a, g)
	if inspection.Frames < 2 {
		t.Fatal("frame counter did not advance")
	}
}

func TestManagedGridSingleTabStop(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "g", Sequence: 1}, Children: []ui.Node{{ID: "before", Kind: ui.Button, Text: "Before"}, {ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true}, {ID: "after", Kind: ui.Button, Text: "After"}}}
	})
	gridFrame(t, m)
	snapshot, e := m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	window := snapshot.Windows[0]
	if window.FocusedID != "g" {
		t.Fatal("grid initial focus missing", window.FocusedID)
	}
	if e = a.Do(func() {
		binding := findRecord(m.implementation.find("one").state.root, "g").control.grid
		if binding.horizontal.AbilityIs(abilities.Focusable) || binding.vertical.AbilityIs(abilities.Focusable) {
			t.Error("scrollbar owns a Tab stop")
		}
	}); e != nil {
		t.Fatal(e)
	}
	tab := ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.KeyChord, Code: key.CodeTab}
	if e = m.Input(testContext, tab); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	snapshot, e = m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	if snapshot.Windows[0].FocusedID != "after" {
		t.Fatal("Tab entered internal grid widgets", snapshot.Windows[0].FocusedID)
	}
	tab.Modifiers.SetFlag(true, key.Shift)
	if e = m.Input(testContext, tab); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	snapshot, e = m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	if snapshot.Windows[0].FocusedID != "g" {
		t.Fatal("ShiftTab did not return to host", snapshot.Windows[0].FocusedID)
	}
}

func TestManagedGridFiniteExtremeWheelThenNormal(t *testing.T) {
	for _, dpi := range []float32{48, 96} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			g, _ := testNativeGrid(t, 10000)
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
			gridFrame(t, m)
			inspection := gridInspection(t, a, g)
			snapshot, e := m.Snapshot(testContext)
			if e != nil {
				t.Fatal(e)
			}
			wheel := ManualInput{Window: "one", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.Scroll, Position: inspection.Bounds.Min.Add(image.Pt(200, 80)), Delta: math32.Vector2{X: math.MaxFloat32, Y: math.MaxFloat32}}
			// A desktop producer may supply a larger finite delta than the manual
			// convenience admission bound. Exercise that real originating Source.
			m.implementation.find("one").state.source.Load().Scroll(wheel.Position, wheel.Delta)
			gridFrame(t, m)
			var before, after virtual.GridState
			var stateErr error
			if err := a.Do(func() { before, stateErr = g.State() }); err != nil {
				t.Fatal(err)
			}
			if stateErr != nil {
				t.Fatal(stateErr)
			}
			wheel.Delta = math32.Vector2{X: -1.25 * inspection.Scale, Y: -1.25 * inspection.Scale}
			if e = m.Input(testContext, wheel); e != nil {
				t.Fatal(e)
			}
			gridFrame(t, m)
			if err := a.Do(func() {
				after, stateErr = g.State()
				b := m.implementation.find("one").state.root.control.grid
				if math.IsNaN(b.wheelX) || math.IsNaN(b.wheelY) || math.IsInf(b.wheelX, 0) || math.IsInf(b.wheelY, 0) || math.Abs(b.wheelX) >= 1 || math.Abs(b.wheelY) >= 1 {
					t.Error("wheel remainder poisoned")
				}
			}); err != nil {
				t.Fatal(err)
			}
			if stateErr != nil {
				t.Fatal(stateErr)
			}
			if after.Viewport.ScrollY != before.Viewport.ScrollY-1 || after.Viewport.ScrollX != max(0, before.Viewport.ScrollX-1) {
				t.Fatal("normal wheel failed after finite extreme", before.Viewport, after.Viewport)
			}
		})
	}
}

func TestManagedGridPendingRequestRejectsRetainedHitAndAX(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	inspection := gridInspection(t, a, g)
	cell := inspection.Cells[0]
	root := m.implementation.find("one").state.root
	binding := root.control.grid
	binding.requestViewport(gridRequest{inspection.Viewport.Width - 1, inspection.Viewport.Height, inspection.Scale})
	// The old retained publication still exists, before the checked owner callback
	// changes generation. It must not grant hit or AX action eligibility.
	if _, ok := binding.refAt(cell.Bounds.Min.Add(cell.Bounds.Size().Div(2))); ok {
		t.Fatal("pending geometry hit old publication")
	}
	if e := a.Do(func() {
		extra, handlers := gridAccessibility(root, "parent", nil, false)
		if len(extra) != 0 || len(handlers) != 0 {
			t.Error("pending geometry exposed AX actions")
		}
		if _, e := InspectGrid(a, g); !errors.Is(e, ErrGridLayout) {
			t.Error(e)
		}
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
}
