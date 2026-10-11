package virtual

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type gridCalls struct{ RowKey, IndexOf, Cell int }
type indexedGridSource struct {
	id        string
	revision  uint64
	rows      int
	reverse   bool
	calls     gridCalls
	rowHook   func(int) (string, error)
	indexHook func(string) (int, bool, error)
	cellHook  func(string, string) (string, error)
}

func gridSource(rows int) *indexedGridSource {
	return &indexedGridSource{id: "parts:seed7", revision: 1, rows: rows}
}
func (s *indexedGridSource) Identity() string { return s.id }
func (s *indexedGridSource) Revision() uint64 { return s.revision }
func (s *indexedGridSource) RowCount() int    { return s.rows }
func (s *indexedGridSource) RowKey(i int) (string, error) {
	s.calls.RowKey++
	if s.rowHook != nil {
		return s.rowHook(i)
	}
	if i < 0 || i >= s.rows {
		return "", errors.New("row bounds")
	}
	if s.reverse {
		i = s.rows - 1 - i
	}
	return "r:" + strconv.Itoa(i), nil
}
func (s *indexedGridSource) IndexOf(key string) (int, bool, error) {
	s.calls.IndexOf++
	if s.indexHook != nil {
		return s.indexHook(key)
	}
	i, err := strconv.Atoi(strings.TrimPrefix(key, "r:"))
	if err != nil || key != "r:"+strconv.Itoa(i) || i < 0 || i >= s.rows {
		return 0, false, nil
	}
	if s.reverse {
		i = s.rows - 1 - i
	}
	return i, true, nil
}
func (s *indexedGridSource) Cell(key, col string) (string, error) {
	s.calls.Cell++
	if s.cellHook != nil {
		return s.cellHook(key, col)
	}
	return key + "/" + col, nil
}
func gridColumns(n int, width int) []GridColumn {
	cols := make([]GridColumn, n)
	for i := range cols {
		cols[i] = GridColumn{fmt.Sprintf("c:%d", i), fmt.Sprintf("Column %d", i), width}
	}
	return cols
}
func mustGrid(t *testing.T, s GridSource, c []GridColumn, cfg GridConfig, v GridViewport) *Grid {
	t.Helper()
	g, err := NewGrid(s, c, cfg, v)
	if err != nil {
		t.Fatal(err)
	}
	return g
}
func gridState(t *testing.T, g *Grid) GridState {
	t.Helper()
	s, err := g.State()
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func gridSnapshot(t *testing.T, g *Grid) GridSnapshot {
	t.Helper()
	s, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func visibleGridCells(s GridSnapshot) int {
	n := 0
	for _, c := range s.Cells {
		if !c.Clip.Empty() {
			n++
		}
	}
	return n
}

func TestGridTwoAxisGeometryAndFrozenRegions(t *testing.T) {
	s := gridSource(100)
	g := mustGrid(t, s, gridColumns(20, 100), GridConfig{RowHeight: 24, HeaderHeight: 24, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1}, GridViewport{Width: 800, Height: 264, ScrollX: 350, ScrollY: 127})
	one := gridSnapshot(t, g)
	if one.State.Columns != 20 || len(one.Headers) >= one.State.Columns {
		t.Fatal("total column count follows virtualized headers", one.State.Columns, len(one.Headers))
	}
	if len(one.Cells) != 13*11 || len(one.Headers) != 11 || visibleGridCells(one) != 11*9 {
		t.Fatalf("unexpected pool rows/cells/headers: %d/%d/%d", len(one.Cells), len(one.Headers), visibleGridCells(one))
	}
	for _, c := range one.Cells {
		if c.Rect.Y != 24+c.RowIndex*24-127 {
			t.Fatal("row geometry", c)
		}
		if c.Frozen && c.Rect.X != c.ColumnIndex*100 {
			t.Fatal("frozen column moved", c)
		}
		if !c.Frozen && c.Rect.X != c.ColumnIndex*100-350 {
			t.Fatal("scroll column", c)
		}
		if !c.Clip.Empty() && (c.Clip.X < 0 || c.Clip.Y < 24 || c.Clip.X+c.Clip.Width > 800 || c.Clip.Y+c.Clip.Height > 264) {
			t.Fatal("body clipped outside viewport", c)
		}
		if !c.Frozen && !c.Clip.Empty() && c.Clip.X < 200 {
			t.Fatal("scrolling cell overlays frozen region", c)
		}
	}
	for _, h := range one.Headers {
		if h.Rect.Y != 0 || (!h.Clip.Empty() && h.Clip.Height != 24) {
			t.Fatal("header vertical geometry", h)
		}
	}
	if err := g.Scroll(100, 24); err != nil {
		t.Fatal(err)
	}
	two := gridSnapshot(t, g)
	for _, c := range two.Cells {
		if c.Frozen && c.Rect.X != c.ColumnIndex*100 {
			t.Fatal("frozen after pan", c)
		}
	}
	if two.Headers[0].Rect != one.Headers[0].Rect || two.Headers[1].Rect != one.Headers[1].Rect {
		t.Fatal("frozen header changed during pan")
	}
}

func TestGridCapacitiesRefuseBeforeMaterializationAndPreserveRefs(t *testing.T) {
	s := gridSource(1_000_000)
	g := mustGrid(t, s, gridColumns(20, 100), GridConfig{RowHeight: 20, HeaderHeight: 20, FrozenColumns: 1, MaxRows: 4, MaxColumns: 4, MaxCells: 12}, GridViewport{Width: 300, Height: 80})
	snap := gridSnapshot(t, g)
	before := gridState(t, g)
	calls := s.calls
	for _, v := range []GridViewport{{Width: 300, Height: 120}, {Width: 500, Height: 80}, {Width: 400, Height: 100}} {
		if err := g.SetViewport(v); !errors.Is(err, ErrGridCapacity) {
			t.Fatal("missing capacity refusal", v, err)
		}
		if gridState(t, g) != before || s.calls != calls {
			t.Fatal("refusal mutated state or called source")
		}
	}
	if err := g.Activate(snap.Cells[0].Ref); err != nil {
		t.Fatal("refusal retired ref", err)
	}
	if _, err := NewGrid(s, gridColumns(20, 100), GridConfig{RowHeight: 1, MaxRows: 1}, GridViewport{Width: 100, Height: 2}); !errors.Is(err, ErrGridCapacity) {
		t.Fatal(err)
	}
}

func TestGridRefsRejectPoolingSourceAndForeignIdentity(t *testing.T) {
	s := gridSource(100)
	g := mustGrid(t, s, gridColumns(4, 100), GridConfig{RowHeight: 20, HeaderHeight: 20, OverscanRows: 1}, GridViewport{Width: 200, Height: 60, ScrollY: 40})
	snap := gridSnapshot(t, g)
	var visible, overscan GridCellRef
	for _, c := range snap.Cells {
		if c.Clip.Empty() {
			overscan = c.Ref
		} else {
			visible = c.Ref
		}
	}
	if err := g.Activate(overscan); !errors.Is(err, ErrGridStale) {
		t.Fatal("overscan accepted", err)
	}
	other := mustGrid(t, s, gridColumns(4, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 60})
	if err := other.Activate(visible); !errors.Is(err, ErrGridStale) {
		t.Fatal("foreign grid accepted", err)
	}
	if err := g.Scroll(100, 20); err != nil {
		t.Fatal(err)
	}
	if err := g.Activate(visible); !errors.Is(err, ErrGridStale) {
		t.Fatal("rebound slot accepted", err)
	}
	ref := gridSnapshot(t, g).Cells[2].Ref
	s.revision++
	s.reverse = true
	if err := g.Activate(ref); !errors.Is(err, ErrGridStale) {
		t.Fatal("old source accepted", err)
	}
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	if err := g.Activate(ref); !errors.Is(err, ErrGridStale) {
		t.Fatal("ref survived revision", err)
	}
	ref = gridSnapshot(t, g).Cells[2].Ref
	replacement := gridSource(100)
	replacement.id = "new-incarnation"
	if err := g.ReplaceSource(replacement); err != nil {
		t.Fatal(err)
	}
	if err := g.Activate(ref); !errors.Is(err, ErrGridStale) {
		t.Fatal("ref survived replacement", err)
	}
	if err := g.Activate(GridCellRef{}); !errors.Is(err, ErrGridStale) {
		t.Fatal("forged empty ref", err)
	}
}

func TestGridNavigationAndIndexedRefresh(t *testing.T) {
	s := gridSource(1_000_000)
	g := mustGrid(t, s, gridColumns(20, 100), GridConfig{RowHeight: 24, HeaderHeight: 24, FrozenColumns: 2}, GridViewport{Width: 800, Height: 1224})
	s.calls = gridCalls{}
	if err := g.JumpToKey("r:999999", "c:19"); err != nil {
		t.Fatal(err)
	}
	state := gridState(t, g)
	if state.Active != (GridActive{"r:999999", "c:19"}) || state.Viewport.ScrollY != 999950*24 || state.Viewport.ScrollX != 1200 {
		t.Fatal(state)
	}
	if s.calls.Cell != 0 || s.calls.IndexOf > 3 || s.calls.RowKey > 3 {
		t.Fatal("navigation scanned", s.calls)
	}
	if err := g.Page(-1); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Active.RowKey != "r:999949" {
		t.Fatal("page rows", gridState(t, g))
	}
	if err := g.Move(math.MinInt, math.MinInt); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Active != (GridActive{"r:0", "c:0"}) {
		t.Fatal("move underflow")
	}
	if err := g.Move(math.MaxInt, math.MaxInt); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Active != (GridActive{"r:999999", "c:19"}) {
		t.Fatal("move overflow")
	}
	if err := g.Scroll(0, -1); err != nil {
		t.Fatal(err)
	}
	before := gridState(t, g)
	oldAnchor := g.anchorKey
	offset := g.anchorOffset
	s.revision++
	s.reverse = true
	s.calls = gridCalls{}
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	after := gridState(t, g)
	if after.Active != before.Active {
		t.Fatal("stable active lost", after)
	}
	index, found, _ := s.IndexOf(oldAnchor)
	if !found || after.Viewport.ScrollY != index*24+offset {
		t.Fatal("anchor lost", after, index, offset)
	}
	if s.calls.Cell != 0 || s.calls.RowKey > 4 || s.calls.IndexOf > 5 {
		t.Fatal("refresh scanned", s.calls)
	}
	s.rows = 2
	s.revision++
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Active != (GridActive{}) {
		t.Fatal("deleted active retained")
	}
	newsource := gridSource(20)
	newsource.id = "replacement"
	if err := g.ReplaceSource(newsource); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Viewport.ScrollX != 0 || gridState(t, g).Viewport.ScrollY != 0 {
		t.Fatal("new incarnation retained scroll")
	}
}

func TestGridSourceFailuresAreBoundedAndAtomic(t *testing.T) {
	s := gridSource(1_000_000)
	g := mustGrid(t, s, gridColumns(4, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 40})
	before := gridState(t, g)
	s.rowHook = func(i int) (string, error) {
		if i == 500_000 {
			return "", nil
		}
		return "r:" + strconv.Itoa(i), nil
	}
	if _, err := g.Snapshot(); err != nil {
		t.Fatal("validated hidden rows", err)
	}
	if err := g.JumpToKey("r:500000", "c:0"); !errors.Is(err, ErrGridSource) {
		t.Fatal("invalid accessed key", err)
	}
	if gridState(t, g) != before {
		t.Fatal("failed jump changed state")
	}
	s.rowHook = nil
	s.cellHook = func(string, string) (string, error) { s.revision++; return "changed", nil }
	if snap, err := g.Snapshot(); !errors.Is(err, ErrGridStale) || len(snap.Cells) != 0 {
		t.Fatal("mixed revision returned", snap, err)
	}
	s.cellHook = nil
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	before = gridState(t, g)
	s.indexHook = func(string) (int, bool, error) { return math.MaxInt, true, nil }
	if err := g.JumpToKey("r:20", "c:0"); !errors.Is(err, ErrGridSource) {
		t.Fatal("bad index", err)
	}
	if gridState(t, g) != before {
		t.Fatal("invalid index mutation")
	}
	s.indexHook = nil
	s.revision--
	if err := g.Refresh(); !errors.Is(err, ErrGridStale) {
		t.Fatal("revision rollback", err)
	}
}

func TestGridTextBudgetAndDetachedRecords(t *testing.T) {
	s := gridSource(20)
	cols := gridColumns(2, 100)
	g := mustGrid(t, s, cols, GridConfig{RowHeight: 20, HeaderHeight: 20, MaxTextBytes: 80}, GridViewport{Width: 200, Height: 60})
	cols[0] = GridColumn{ID: "mutated", Width: 1}
	one := gridSnapshot(t, g)
	one.Headers[0].ColumnID = "bad"
	one.Cells[0].Text = "bad"
	one.Cells[0].Ref = GridCellRef{}
	two := gridSnapshot(t, g)
	if two.Headers[0].ColumnID != "c:0" || two.Cells[0].Text == "bad" {
		t.Fatal("snapshot or columns alias")
	}
	before := gridState(t, g)
	s.cellHook = func(string, string) (string, error) { return strings.Repeat("x", 81), nil }
	if out, err := g.Snapshot(); !errors.Is(err, ErrGridCapacity) || len(out.Cells) != 0 {
		t.Fatal("large payload", out, err)
	}
	if gridState(t, g) != before {
		t.Fatal("payload refusal mutates model")
	}
}

func TestGridInvalidExhaustedEmptyAndHiddenPanes(t *testing.T) {
	cols := gridColumns(2, 100)
	s := gridSource(1)
	var typednil *indexedGridSource
	for _, source := range []GridSource{nil, typednil, &indexedGridSource{id: "x", revision: 0, rows: 1}, &indexedGridSource{id: "x", revision: 1, rows: -1}} {
		if _, err := NewGrid(source, cols, GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 60}); !errors.Is(err, ErrGridSource) {
			t.Fatal("invalid source", err)
		}
	}
	for _, cfg := range []GridConfig{{RowHeight: 0}, {RowHeight: 20, FrozenColumns: 3}, {RowHeight: 20, MaxRows: -1}, {RowHeight: 20, OverscanColumns: -1}} {
		if _, err := NewGrid(s, cols, cfg, GridViewport{Width: 200, Height: 60}); !errors.Is(err, ErrGridInvalid) {
			t.Fatal("invalid config", cfg, err)
		}
	}
	for _, badcols := range [][]GridColumn{{{ID: "x", Width: 1}, {ID: "x", Width: 1}}, {{ID: "x", Width: math.MaxInt}, {ID: "y", Width: 1}}, {{ID: "x", Width: 0}}} {
		if _, err := NewGrid(s, badcols, GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 60}); !errors.Is(err, ErrGridInvalid) {
			t.Fatal("invalid columns", err)
		}
	}
	s.rows = math.MaxInt
	if _, err := NewGrid(s, cols, GridConfig{RowHeight: 2}, GridViewport{Width: 200, Height: 60}); !errors.Is(err, ErrGridInvalid) {
		t.Fatal("extent overflow", err)
	}
	g := mustGrid(t, gridSource(0), cols, GridConfig{RowHeight: 20, HeaderHeight: 20}, GridViewport{Width: 200, Height: 60})
	if len(gridSnapshot(t, g).Cells) != 0 || g.Move(1, 1) != ErrGridRange {
		t.Fatal("empty model")
	}
	g = mustGrid(t, gridSource(1), cols, GridConfig{RowHeight: 20, HeaderHeight: 20, FrozenColumns: 1}, GridViewport{Width: 50, Height: 10})
	if len(gridSnapshot(t, g).Cells) != 0 {
		t.Fatal("header consumes body")
	}
	if err := g.JumpToKey("r:0", "c:1"); err != nil {
		t.Fatal(err)
	} // active may be clipped when panes have zero area
	copyGrid := *g
	if _, err := copyGrid.State(); !errors.Is(err, ErrGridInvalid) {
		t.Fatal("copied model")
	}
	if _, err := (*Grid)(nil).State(); !errors.Is(err, ErrGridInvalid) {
		t.Fatal("nil model")
	}
	g.generation = math.MaxUint64
	before := gridState(t, g)
	if err := g.SetViewport(GridViewport{Width: 60, Height: 10}); !errors.Is(err, ErrGridExhausted) {
		t.Fatal("wrap", err)
	}
	if gridState(t, g) != before {
		t.Fatal("exhaustion changed state")
	}
}

// A deliberately slow brute oracle checks small arbitrary geometry against the
// indexed layout without using its column search/range logic.
func TestGridGeometryPropertyOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(70019))
	for trial := 0; trial < 800; trial++ {
		rows, n := rng.Intn(50), 1+rng.Intn(20)
		cols := gridColumns(n, 1)
		for i := range cols {
			cols[i].Width = 1 + rng.Intn(80)
		}
		cfg := GridConfig{RowHeight: 1 + rng.Intn(30), HeaderHeight: rng.Intn(30), FrozenColumns: rng.Intn(n + 1), MaxRows: 256, MaxColumns: 64, MaxCells: 4096}
		g := mustGrid(t, gridSource(rows), cols, cfg, GridViewport{Width: 1 + rng.Intn(700), Height: 1 + rng.Intn(300), ScrollX: rng.Intn(1000), ScrollY: rng.Intn(2000)})
		snap := gridSnapshot(t, g)
		v := snap.State.Viewport
		want := map[[2]int]GridRect{}
		prefix := 0
		frozenWidth := 0
		for i, c := range cols {
			if i < cfg.FrozenColumns {
				frozenWidth += c.Width
			}
		}
		for col, c := range cols {
			x := prefix
			prefix += c.Width
			left, right := 0, min(v.Width, frozenWidth)
			if col >= cfg.FrozenColumns {
				x -= v.ScrollX
				left, right = min(v.Width, frozenWidth), v.Width
			}
			for row := 0; row < rows; row++ {
				y := cfg.HeaderHeight + row*cfg.RowHeight - v.ScrollY
				x0, x1 := max(x, left), min(x+c.Width, right)
				y0, y1 := max(y, min(v.Height, cfg.HeaderHeight)), min(y+cfg.RowHeight, v.Height)
				if x1 > x0 && y1 > y0 {
					want[[2]int{row, col}] = GridRect{x0, y0, x1 - x0, y1 - y0}
				}
			}
		}
		got := map[[2]int]GridRect{}
		for _, c := range snap.Cells {
			if !c.Clip.Empty() {
				got[[2]int{c.RowIndex, c.ColumnIndex}] = c.Clip
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("trial %d cfg=%+v viewport=%+v\ngot=%v\nwant=%v", trial, cfg, v, got, want)
		}
	}
}

func TestGridWorkloadCallbacks(t *testing.T) {
	var results []struct {
		Rows, Columns, VisibleCells, MaterializedCells, Headers int
		Calls                                                   gridCalls
		NavigationCalls                                         gridCalls
	}
	for _, rows := range []int{10_000, 1_000_000} {
		s := gridSource(rows)
		g := mustGrid(t, s, gridColumns(20, 100), GridConfig{RowHeight: 24, HeaderHeight: 24, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1, MaxRows: 64, MaxColumns: 12, MaxCells: 768}, GridViewport{Width: 800, Height: 1224, ScrollX: 400, ScrollY: (rows / 2) * 24})
		if s.calls != (gridCalls{RowKey: 1, IndexOf: 1}) {
			t.Fatal("construction enumerated", s.calls)
		}
		s.calls = gridCalls{}
		snap := gridSnapshot(t, g)
		calls := s.calls
		if visibleGridCells(snap) != 400 || len(snap.Cells) != 520 || calls != (gridCalls{52, 52, 520}) {
			t.Fatal("viewport budget", rows, len(snap.Cells), calls)
		}
		s.calls = gridCalls{}
		if err := g.JumpToKey("r:"+strconv.Itoa(rows-1), "c:19"); err != nil {
			t.Fatal(err)
		}
		navigation := s.calls
		if navigation != (gridCalls{2, 2, 0}) {
			t.Fatal("jump enumerated", navigation)
		}
		results = append(results, struct {
			Rows, Columns, VisibleCells, MaterializedCells, Headers int
			Calls                                                   gridCalls
			NavigationCalls                                         gridCalls
		}{rows, 20, visibleGridCells(snap), len(snap.Cells), len(snap.Headers), calls, navigation})
	}
	data, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GRID_WORKLOAD " + string(data))
}
