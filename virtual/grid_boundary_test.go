package virtual

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestGridUTF8AndWrappedSourceErrors(t *testing.T) {
	s := gridSource(10)
	for _, columns := range [][]GridColumn{{{ID: "\xff", Width: 20}}, {{ID: "x", Title: "\xff", Width: 20}}} {
		if _, err := NewGrid(s, columns, GridConfig{RowHeight: 20}, GridViewport{Width: 20, Height: 20}); !errors.Is(err, ErrGridInvalid) {
			t.Fatal("invalid UTF8 column", err)
		}
	}
	g := mustGrid(t, s, gridColumns(2, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 40})
	before := gridState(t, g)
	failure := errors.New("injected source failure")
	for _, kind := range []string{"row", "index", "cell"} {
		s.rowHook = nil
		s.indexHook = nil
		s.cellHook = nil
		switch kind {
		case "row":
			s.rowHook = func(int) (string, error) { return "", failure }
		case "index":
			s.indexHook = func(string) (int, bool, error) { return 0, false, failure }
		case "cell":
			s.cellHook = func(string, string) (string, error) { return "", failure }
		}
		if out, err := g.Snapshot(); !errors.Is(err, ErrGridSource) || !errors.Is(err, failure) || len(out.Cells) != 0 {
			t.Fatal(kind, out, err)
		}
		if gridState(t, g) != before {
			t.Fatal("failure changed grid", kind)
		}
	}
	s.rowHook = nil
	s.indexHook = nil
	s.cellHook = func(string, string) (string, error) { return "\xff", nil }
	if _, err := g.Snapshot(); !errors.Is(err, ErrGridSource) {
		t.Fatal("invalid cell UTF8", err)
	}
	s.cellHook = nil
	s.rowHook = func(int) (string, error) { return "\xff", nil }
	if _, err := g.Snapshot(); !errors.Is(err, ErrGridSource) {
		t.Fatal("invalid key UTF8", err)
	}
}

func TestGridMaximumCoordinateBoundary(t *testing.T) {
	s := gridSource(math.MaxInt)
	g := mustGrid(t, s, []GridColumn{{ID: "wide", Width: math.MaxInt}}, GridConfig{RowHeight: 1, MaxRows: 4, MaxCells: 4}, GridViewport{Width: 2, Height: 2, ScrollX: math.MaxInt, ScrollY: math.MaxInt})
	state := gridState(t, g)
	if state.Viewport.ScrollX != math.MaxInt-2 || state.Viewport.ScrollY != math.MaxInt-2 {
		t.Fatal("extent clamp", state)
	}
	snap := gridSnapshot(t, g)
	if len(snap.Cells) != 2 || visibleGridCells(snap) != 2 {
		t.Fatal("last cells", snap)
	}
	for i, c := range snap.Cells {
		if c.Clip != (GridRect{X: 0, Y: i, Width: 2, Height: 1}) {
			t.Fatal("overflow in clip", c)
		}
	}
	if err := g.Scroll(math.MinInt, math.MinInt); err != nil {
		t.Fatal(err)
	}
	if state := gridState(t, g); state.Viewport.ScrollX != 0 || state.Viewport.ScrollY != 0 {
		t.Fatal("negative extreme", state)
	}
}

func TestGridNoOpAndFailureDoNotRetireReferences(t *testing.T) {
	s := gridSource(10)
	g := mustGrid(t, s, gridColumns(2, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 40})
	snap := gridSnapshot(t, g)
	before := gridState(t, g)
	if err := g.Scroll(-10, -10); err != nil {
		t.Fatal(err)
	}
	if err := g.SetViewport(before.Viewport); err != nil {
		t.Fatal(err)
	}
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	if err := g.JumpToKey("missing", "c:0"); !errors.Is(err, ErrGridRange) {
		t.Fatal(err)
	}
	if err := g.JumpToKey("r:0", "missing"); !errors.Is(err, ErrGridRange) {
		t.Fatal(err)
	}
	if err := g.Page(0); !errors.Is(err, ErrGridInvalid) {
		t.Fatal(err)
	}
	if gridState(t, g) != before {
		t.Fatal("no-op or failure retired state")
	}
	if err := g.Activate(snap.Cells[0].Ref); err != nil {
		t.Fatal(err)
	}
	active := gridState(t, g)
	if err := g.Activate(gridSnapshot(t, g).Cells[0].Ref); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g) != active {
		t.Fatal("same active changed generation")
	}
}

func TestGridRefreshFailuresAndEmptyRevisionTransitions(t *testing.T) {
	s := gridSource(20)
	g := mustGrid(t, s, gridColumns(2, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 40, ScrollY: 100})
	if err := g.JumpToKey("r:10", "c:1"); err != nil {
		t.Fatal(err)
	}
	before := g.state()
	s.rows = 21 // changing membership without changing revision violates the source seam
	if err := g.Refresh(); !errors.Is(err, ErrGridSource) {
		t.Fatal("unversioned membership change", err)
	}
	if g.state() != before {
		t.Fatal("invalid refresh installed")
	}
	s.revision++
	s.rows = math.MaxInt
	if err := g.Refresh(); !errors.Is(err, ErrGridInvalid) {
		t.Fatal("extent refusal", err)
	}
	if g.state() != before {
		t.Fatal("extent refusal mutated")
	}
	s.rows = 0
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	if state := gridState(t, g); state.Active != (GridActive{}) || state.Viewport.ScrollY != 0 {
		t.Fatal("empty refresh", state)
	}
	s.rows = 20
	s.revision++
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	if len(gridSnapshot(t, g).Cells) != 4 {
		t.Fatal("empty-to-populated refresh")
	}
}

func TestGridMidRefreshSourceChangeNeverInstallsPartialState(t *testing.T) {
	s := gridSource(20)
	g := mustGrid(t, s, gridColumns(2, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 40})
	if err := g.JumpToKey("r:10", "c:1"); err != nil {
		t.Fatal(err)
	}
	before := g.state()
	s.revision++
	s.indexHook = func(key string) (int, bool, error) {
		s.revision++
		i, err := strconv.Atoi(strings.TrimPrefix(key, "r:"))
		return i, err == nil, err
	}
	if err := g.Refresh(); !errors.Is(err, ErrGridStale) {
		t.Fatal("mid-refresh revision drift", err)
	}
	if g.state() != before {
		t.Fatal("partial refresh installed")
	}
	s.indexHook = nil
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Active != before.Active {
		t.Fatal("failed retry lost active key")
	}
}

func TestGridAnchorRetentionClampsNearSourceEnd(t *testing.T) {
	s := gridSource(100)
	g := mustGrid(t, s, gridColumns(2, 100), GridConfig{RowHeight: 20, HeaderHeight: 20}, GridViewport{Width: 200, Height: 60})
	if err := g.JumpToKey("r:0", "c:0"); err != nil {
		t.Fatal(err)
	}
	s.reverse = true
	s.revision++
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	state := gridState(t, g)
	if state.Viewport.ScrollY != 1960 || state.Active.RowKey != "r:0" || g.anchorKey != "r:1" {
		t.Fatal("end clamp incorrectly preserves impossible top anchor", state, g.anchorKey)
	}
	snapshot := gridSnapshot(t, g)
	if snapshot.Cells[2].RowKey != "r:0" || snapshot.Cells[2].Clip.Y != 40 || !snapshot.Cells[2].Active {
		t.Fatal("stable active key missing at clamped end", snapshot.Cells)
	}
}

func TestGridOverscanAndCapacityProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(9217))
	for trial := 0; trial < 300; trial++ {
		rows, n := 1+rng.Intn(40), 1+rng.Intn(15)
		cols := gridColumns(n, 1)
		for i := range cols {
			cols[i].Width = 1 + rng.Intn(40)
		}
		cfg := GridConfig{RowHeight: 1 + rng.Intn(25), HeaderHeight: rng.Intn(15), FrozenColumns: rng.Intn(n + 1), OverscanRows: rng.Intn(4), OverscanColumns: rng.Intn(3), MaxRows: 256, MaxColumns: 64, MaxCells: 4096}
		g := mustGrid(t, gridSource(rows), cols, cfg, GridViewport{Width: 1 + rng.Intn(300), Height: 1 + rng.Intn(180), ScrollX: rng.Intn(500), ScrollY: rng.Intn(1000)})
		snap := gridSnapshot(t, g)
		// Exact clipped visible cells must agree with an otherwise identical grid
		// without overscan; extra records may not change or duplicate visible ones.
		plain := cfg
		plain.OverscanRows, plain.OverscanColumns = 0, 0
		other := mustGrid(t, gridSource(rows), cols, plain, g.viewport)
		visible := func(s GridSnapshot) map[[2]int]GridRect {
			out := map[[2]int]GridRect{}
			for _, c := range s.Cells {
				if !c.Clip.Empty() {
					key := [2]int{c.RowIndex, c.ColumnIndex}
					if _, ok := out[key]; ok {
						t.Fatal("duplicate visible cell")
					}
					out[key] = c.Clip
				}
			}
			return out
		}
		if !reflect.DeepEqual(visible(snap), visible(gridSnapshot(t, other))) {
			t.Fatal("overscan changed visibility", trial)
		}
		if len(snap.Cells) > 0 {
			limited := cfg
			limited.MaxCells = len(snap.Cells) - 1
			if limited.MaxCells == 0 {
				limited.MaxCells = -1
			}
			_, err := NewGrid(gridSource(rows), cols, limited, g.viewport)
			if !errors.Is(err, ErrGridCapacity) && !errors.Is(err, ErrGridInvalid) {
				t.Fatal("cell limit ignored", trial, err)
			}
		}
	}
}

func TestGridIndependentOwnersParallel(t *testing.T) {
	for i := 0; i < 8; i++ {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			t.Parallel()
			s := gridSource(1_000_000)
			g := mustGrid(t, s, gridColumns(20, 100), GridConfig{RowHeight: 20, HeaderHeight: 20, FrozenColumns: 2}, GridViewport{Width: 800, Height: 220})
			for step := 0; step < 100; step++ {
				if err := g.Scroll(33, 17); err != nil {
					t.Fatal(err)
				}
				if visibleGridCells(gridSnapshot(t, g)) == 0 {
					t.Fatal("empty independent viewport")
				}
			}
		})
	}
}
