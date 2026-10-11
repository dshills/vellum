//go:build offscreen

package cogent

import (
	"errors"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// The fixture owns immutable O(1) prefix geometry. Heights belong to stable row
// keys, so reversing the projection also reverses its metric index. RowAt uses
// a bounded binary search and never materializes an owner-side whole-row array.
type variableNativeSource struct {
	rows                              int
	revision                          uint64
	reverse                           bool
	giantIndex, giantHeight           int
	cells, keys, indices, offsets, at int
}

var variableNativePattern = [...]int{17, 31, 19, 43, 23}

func variableNativePrefix(boundary int) int {
	n := boundary / len(variableNativePattern)
	total := n * 133
	for _, h := range variableNativePattern[:boundary%len(variableNativePattern)] {
		total += h
	}
	return total
}
func (s *variableNativeSource) Identity() string { return "native-variable-grid" }
func (s *variableNativeSource) Revision() uint64 { return s.revision }
func (s *variableNativeSource) RowCount() int    { return s.rows }
func (s *variableNativeSource) original(row int) int {
	if s.reverse {
		return s.rows - 1 - row
	}
	return row
}
func (s *variableNativeSource) RowKey(row int) (string, error) {
	s.keys++
	if row < 0 || row >= s.rows {
		return "", virtual.ErrGridRange
	}
	return fmt.Sprintf("variable-%d", s.original(row)), nil
}
func (s *variableNativeSource) IndexOf(k string) (int, bool, error) {
	s.indices++
	row, err := strconv.Atoi(strings.TrimPrefix(k, "variable-"))
	if err != nil || row < 0 || row >= s.rows || !strings.HasPrefix(k, "variable-") {
		return 0, false, nil
	}
	if s.reverse {
		row = s.rows - 1 - row
	}
	return row, true, nil
}
func (s *variableNativeSource) Cell(row, column string) (string, error) {
	s.cells++
	return row + " · " + column, nil
}
func (s *variableNativeSource) prefix(boundary int) int {
	value := variableNativePrefix(boundary)
	if s.giantHeight > 0 && boundary > s.giantIndex {
		value += s.giantHeight - variableNativePattern[s.giantIndex%len(variableNativePattern)]
	}
	return value
}
func (s *variableNativeSource) offset(boundary int) int {
	if s.reverse {
		return s.prefix(s.rows) - s.prefix(s.rows-boundary)
	}
	return s.prefix(boundary)
}
func (s *variableNativeSource) rowAt(pixel int) int {
	lo, hi := 0, s.rows
	for lo < hi {
		mid := lo + (hi-lo)/2
		if s.offset(mid+1) <= pixel {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}
func variableNativeColumns() []virtual.GridColumn {
	columns := make([]virtual.GridColumn, 20)
	for i := range columns {
		width := 121
		if i%2 != 0 {
			width = 79
		}
		columns[i] = virtual.GridColumn{ID: fmt.Sprintf("c%d", i), Title: fmt.Sprintf("Column %d", i), Width: width}
	}
	return columns
}
func variableNativeState(t *testing.T, a interface{ Do(func()) error }, g *virtual.Grid) virtual.GridState {
	t.Helper()
	var state virtual.GridState
	var stateErr error
	if err := a.Do(func() { state, stateErr = g.State() }); err != nil {
		t.Fatal(err)
	}
	if stateErr != nil {
		t.Fatal(stateErr)
	}
	return state
}
func variableNativeFrameSnapshot(t *testing.T, m *Manual, g *virtual.Grid) (virtual.GridSnapshot, error) {
	t.Helper()
	var snapshot virtual.GridSnapshot
	var err error
	owner := m.implementation.owner
	if dispatchErr := owner.Do(func() {
		binding := m.implementation.find("one").state.root.control.grid
		if binding.model != g {
			err = ErrGridBinding
			return
		}
		snapshot = binding.projection
		snapshot.Cells = append([]virtual.GridCell(nil), snapshot.Cells...)
		snapshot.Headers = append([]virtual.GridHeader(nil), snapshot.Headers...)
	}); dispatchErr != nil {
		return snapshot, dispatchErr
	}
	return snapshot, err
}
func variableNativeInput(t *testing.T, m *Manual, in ManualInput) {
	t.Helper()
	if err := m.Input(testContext, in); err != nil {
		t.Fatal(err)
	}
}
func variableNativeTarget(t *testing.T, m *Manual, c GridCellGeometry) (ManualWindowSnapshot, ManualElement) {
	t.Helper()
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	window := snapshot.Windows[0]
	for _, element := range window.Elements {
		if element.ID == c.ID {
			return window, element
		}
	}
	t.Fatal("variable-height cell absent from actual router capture", c.ID)
	return window, ManualElement{}
}

var _ virtual.GridRowMetrics = (*variableNativeSource)(nil)

func (s *variableNativeSource) RowOffset(boundary int) (int, error) {
	s.offsets++
	if boundary < 0 || boundary > s.rows {
		return 0, virtual.ErrGridRange
	}
	return s.offset(boundary), nil
}
func (s *variableNativeSource) RowAt(pixel int) (int, error) {
	s.at++
	if pixel < 0 || pixel >= s.offset(s.rows) {
		return 0, virtual.ErrGridRange
	}
	return s.rowAt(pixel), nil
}

func TestManagedGridVariableRowGeometryPointerWheelAndPage(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			source := &variableNativeSource{rows: 10000, revision: 1}
			g, err := virtual.NewGrid(source, variableNativeColumns(), virtual.GridConfig{RowHeight: 1, HeaderHeight: 27, FrozenColumns: 1, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
			if err != nil {
				t.Fatal(err)
			}
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			changes := 0
			m, a := gridManual(t, config, func() ui.Node {
				return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(e ui.GridEvent) {
					if e.Reason == ui.GridSelection {
						changes++
					}
				}}
			})
			gridFrame(t, m)
			if err := a.Do(func() {
				if e := g.Scroll(37, source.offset(7)+3); e != nil {
					t.Error(e)
				}
				if e := a.Invalidate("one"); e != nil {
					t.Error(e)
				}
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			measured := gridInspection(t, a, g)
			state := variableNativeState(t, a, g)
			if !state.VariableHeights || state.TotalHeight != source.offset(source.rows) {
				t.Fatal("native model did not select indexed height metrics", state)
			}
			projection, err := variableNativeFrameSnapshot(t, m, g)
			if err != nil {
				t.Fatal(err)
			}
			if len(projection.Cells) != len(measured.Cells) {
				t.Fatal("measured variable projection cardinality mismatch")
			}
			edge := func(v int) int { return int(math.Round(float64(v) * float64(measured.Scale))) }
			var selected GridCellGeometry
			rowHeights := map[int]bool{}
			for i, cell := range measured.Cells {
				p := projection.Cells[i]
				expectedHeight := source.offset(cell.Row+1) - source.offset(cell.Row)
				if p.Rect.Y != 27+source.offset(cell.Row)-state.Viewport.ScrollY || p.Rect.Height != expectedHeight {
					t.Fatal("native projection flattened irregular row", cell, p.Rect, expectedHeight)
				}
				rowHeights[expectedHeight] = true
				if p.Clip.Empty() {
					if !cell.Bounds.Empty() {
						t.Fatal("overscan variable row became visible")
					}
					continue
				}
				expected := image.Rect(measured.Bounds.Min.X+edge(p.Clip.X), measured.Bounds.Min.Y+edge(p.Clip.Y), measured.Bounds.Min.X+edge(p.Clip.X+p.Clip.Width), measured.Bounds.Min.Y+edge(p.Clip.Y+p.Clip.Height))
				if cell.Bounds != expected {
					t.Fatal("variable row native global edge mismatch", cell.Bounds, expected, p.Clip, dpi)
				}
				if cell.Column == 2 && cell.Row >= 8 && cell.Bounds.Dy() > 10 && selected.ID == "" {
					selected = cell
				}
			}
			if len(rowHeights) < 3 || selected.ID == "" {
				t.Fatal("fixture did not render multiple irregular heights and target")
			}
			window, target := variableNativeTarget(t, m, selected)
			down := ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Target: target.ID, TargetIncarnation: target.Incarnation, Type: events.MouseDown, Button: events.Left}
			variableNativeInput(t, m, down)
			if _, err := m.Step(testContext, 1); err != nil {
				t.Fatal(err)
			}
			up := down
			up.Type = events.MouseUp
			// Focus from MouseDown can invalidate a captured dynamic target. The real
			// physical release has a scene position, not a reused target publication.
			up.Target = ""
			up.TargetIncarnation = 0
			up.Position = selected.Bounds.Min.Add(selected.Bounds.Size().Div(2))
			variableNativeInput(t, m, up)
			gridFrame(t, m)
			before := variableNativeState(t, a, g)
			if before.Active.RowKey != selected.RowKey || before.Active.ColumnID != selected.ColumnID || changes != 1 {
				t.Fatal("actual variable row pointer selected wrong key", before.Active, selected, changes)
			}
			inspection := gridInspection(t, a, g)
			variableNativeInput(t, m, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.Scroll, Position: inspection.Bounds.Min.Add(image.Pt(200, 80)), Delta: math32.Vector2{X: 1.25 * inspection.Scale, Y: 2.25 * inspection.Scale}})
			gridFrame(t, m)
			afterWheel := variableNativeState(t, a, g)
			if afterWheel.Viewport.ScrollY != before.Viewport.ScrollY+2 || afterWheel.Viewport.ScrollX != before.Viewport.ScrollX+1 || afterWheel.Active != before.Active {
				t.Fatal("variable-height wheel lost logical pixel units/selection", before, afterWheel)
			}
			body := max(1, afterWheel.Viewport.Height-afterWheel.Config.HeaderHeight)
			activeIndex, e := strconv.Atoi(strings.TrimPrefix(afterWheel.Active.RowKey, "variable-"))
			if e != nil {
				t.Fatal("selected source index absent", e)
			}
			// Pixel paging must use the row metric rather than RowHeight=1 fallback.
			wantIndex := source.rowAt(min(source.offset(source.rows)-1, source.offset(activeIndex)+body))
			variableNativeInput(t, m, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.KeyChord, Code: key.CodePageDown})
			gridFrame(t, m)
			paged := variableNativeState(t, a, g)
			wantKey := fmt.Sprintf("variable-%d", source.original(wantIndex))
			if paged.Active.RowKey != wantKey || paged.Active.ColumnID != selected.ColumnID || changes != 2 {
				t.Fatal("real PageDown ignored indexed row pixel geometry", paged.Active, wantKey, body, changes)
			}
			img, e := m.Capture(testContext, "one")
			if e != nil || img.Bounds().Empty() {
				t.Fatal("variable-height software frame absent", e)
			}
			if source.cells > gridNativeMaxCells*16 {
				t.Fatal("native variable projection materialized source", source.cells)
			}
		})
	}
}

func TestManagedGridVariableRowsReplaceAnchorAndRetireRefs(t *testing.T) {
	source := &variableNativeSource{rows: 10000, revision: 1}
	g, err := virtual.NewGrid(source, variableNativeColumns(), virtual.GridConfig{RowHeight: 1, HeaderHeight: 27, FrozenColumns: 1, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
	if err != nil {
		t.Fatal(err)
	}
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	if err := a.Do(func() {
		if e := g.JumpToKey("variable-222", "c2"); e != nil {
			t.Error(e)
		}
		state, e := g.State()
		if e != nil {
			t.Error(e)
			return
		}
		viewport := state.Viewport
		viewport.ScrollY = source.offset(230) + 5
		if e := g.SetViewport(viewport); e != nil {
			t.Error(e)
		}
		if e := a.Invalidate("one"); e != nil {
			t.Error(e)
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	before := gridInspection(t, a, g)
	oldProjection, err := variableNativeFrameSnapshot(t, m, g)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldProjection.Cells) == 0 {
		t.Fatal("variable source frame has no bounded cells")
	}
	retired := oldProjection.Cells[0].Ref
	replacement := &variableNativeSource{rows: 10000, revision: 2, reverse: true}
	var staleErr error
	if err := a.Do(func() {
		if e := g.ReplaceSource(replacement); e != nil {
			t.Error(e)
			return
		}
		staleErr = g.Activate(retired)
		if e := a.Invalidate("one"); e != nil {
			t.Error(e)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(staleErr, virtual.ErrGridStale) {
		t.Fatal("variable replacement admitted retired native reference", staleErr)
	}
	state := variableNativeState(t, a, g)
	wantOffset := replacement.offset(replacement.rows-1-230) + 5
	if state.Active != (virtual.GridActive{RowKey: "variable-222", ColumnID: "c2"}) || state.Viewport.ScrollY != wantOffset {
		t.Fatal("variable revision replacement lost stable anchor/key selection", state.Active, state.Viewport.ScrollY, wantOffset)
	}
	var pendingErr error
	if err := a.Do(func() { _, pendingErr = InspectGrid(a, g) }); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(pendingErr, ErrGridLayout) {
		t.Fatal("variable replacement inspected stale native frame", pendingErr)
	}
	gridFrame(t, m)
	after := gridInspection(t, a, g)
	if after.Stamp.Revision != 2 || after.Generation == before.Generation || after.CellIncarnation == before.CellIncarnation {
		t.Fatal("replacement did not retire measured source/projection", before, after)
	}
	foundAnchor := false
	for _, cell := range after.Cells {
		if cell.RowKey == "variable-230" && cell.Column == 0 && !cell.Bounds.Empty() {
			foundAnchor = true
			break
		}
	}
	if !foundAnchor {
		t.Fatal("stable anchored variable row disappeared from actual frozen pane")
	}
	counts := [5]int{replacement.cells, replacement.keys, replacement.indices, replacement.offsets, replacement.at}
	for range 3 {
		_ = gridInspection(t, a, g)
	}
	if counts[0] != replacement.cells || counts[1] != replacement.keys || counts[2] != replacement.indices || counts[4] != replacement.at || replacement.offsets-counts[3] > 3*8 {
		t.Fatal("InspectGrid enumerated variable source instead of bounded extent verification", counts, [5]int{replacement.cells, replacement.keys, replacement.indices, replacement.offsets, replacement.at})
	}
}

func TestManagedGridVariableRowsActualNativeBudget(t *testing.T) {
	t.Run("unused-fallback", func(t *testing.T) {
		source := &variableNativeSource{rows: 10000, revision: 1}
		g, err := virtual.NewGrid(source, variableNativeColumns(), virtual.GridConfig{RowHeight: math.MaxInt, HeaderHeight: 27, FrozenColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
		if err != nil {
			t.Fatal(err)
		}
		a, host := gridContractOwner(t, 8)
		r := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g})
		if r.control.grid == nil || r.control.grid.projection.State.TotalHeight != source.offset(source.rows) {
			t.Fatal("unused fixed fallback capped real indexed geometry")
		}
	})
	for _, giantIndex := range []int{0, 2} {
		t.Run(fmt.Sprintf("oversized-materialized-%d", giantIndex), func(t *testing.T) {
			source := &variableNativeSource{rows: 4, revision: 1, giantIndex: giantIndex, giantHeight: gridNativeMaxDimension + 1}
			g, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: 121}}, virtual.GridConfig{RowHeight: 1, HeaderHeight: 27, OverscanRows: 1}, virtual.GridViewport{Width: 200, Height: 60})
			if err != nil {
				t.Fatal("portable indexed metrics should remain valid", err)
			}
			a, host := gridContractOwner(t, 8)
			var r *record
			var mountErr error
			if err := a.Do(func() { r, mountErr = reconcile(host, nil, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
				t.Fatal(err)
			}
			gridRegistry.Lock()
			binding := gridRegistry.models[g]
			gridRegistry.Unlock()
			if !errors.Is(mountErr, ErrGridBudget) || r != nil || binding != nil || len(host.Children) != 0 {
				t.Fatal("oversized actual/overscan variable row allocated usable native pool", mountErr, r, binding, len(host.Children))
			}
		})
	}
	t.Run("row-record-capacity", func(t *testing.T) {
		source := &variableNativeSource{rows: 10000, revision: 1}
		g, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: 121}}, virtual.GridConfig{RowHeight: 1, HeaderHeight: 27, MaxRows: 257}, virtual.GridViewport{Width: 200, Height: 60})
		if err != nil {
			t.Fatal(err)
		}
		a, host := gridContractOwner(t, 8)
		var r *record
		var mountErr error
		if err := a.Do(func() { r, mountErr = reconcile(host, nil, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(mountErr, ErrGridBudget) || r != nil || len(host.Children) != 0 {
			t.Fatal("variable metrics bypassed native record budget", mountErr)
		}
	})
}
