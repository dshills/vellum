package main

import (
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestIndexedSourceProjectionBijectionAndStableCellValues(t *testing.T) {
	for _, total := range []int{1, 3, 10_000, 1_000_000} {
		s, err := newSource("source", total)
		if err != nil {
			t.Fatal(err)
		}
		for _, reverse := range []bool{false, true} {
			for _, filtered := range []bool{false, true} {
				p, err := s.projected(reverse, filtered)
				if err != nil {
					t.Fatal(err)
				}
				for _, index := range []int{0, p.RowCount() / 2, p.RowCount() - 1} {
					key, err := p.RowKey(index)
					if err != nil {
						t.Fatal(err)
					}
					actual, found, err := p.IndexOf(key)
					if err != nil || !found || actual != index {
						t.Fatal("inverse", total, reverse, filtered, index, key, actual, found, err)
					}
					if filtered {
						row, _ := strconv.Atoi(key[4:])
						if row%2 != 0 {
							t.Fatal("odd row survived filter")
						}
					}
					value, err := p.Cell(key, columnID(19))
					if err != nil {
						t.Fatal(err)
					}
					original, err := s.Cell(key, columnID(19))
					if err != nil || value != original {
						t.Fatal("projection changed data", value, original, err)
					}
				}
			}
		}
	}
}
func TestSourceRefusalAndImmutableRevision(t *testing.T) {
	for _, rows := range []int{0, -1, 1_000_001} {
		if _, err := newSource("x", rows); err == nil {
			t.Fatal("invalid rows", rows)
		}
	}
	s, err := newSource("x", 10_000)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "row:1", "row:000010000", "row:-00000001", "bad:000000000"} {
		if _, found, err := s.IndexOf(key); err != nil || found {
			t.Fatal("invalid key", key, found, err)
		}
	}
	if _, err := s.Cell(rowKey(0), "bad"); !errors.Is(err, virtual.ErrGridRange) {
		t.Fatal(err)
	}
	p, err := s.projected(true, true)
	if err != nil {
		t.Fatal(err)
	}
	if s.Revision() != 1 || s.reverse || s.filtered || p.Revision() != 2 || p.RowCount() != 5_000 {
		t.Fatal("mutated source", s, p)
	}
	s.revision = math.MaxUint64
	if _, err := s.projected(true, false); !errors.Is(err, virtual.ErrGridExhausted) {
		t.Fatal(err)
	}
}
func TestCommandProjectionRetainsKeyAndRejectsStalePress(t *testing.T) {
	p, err := newProgram()
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	if err = p.owner.Do(func() {
		w := p.windows[1]
		if err := w.grid.JumpToKey(rowKey(999_998), columnID(19)); err != nil {
			t.Fatal(err)
		}
		snapshot, err := w.grid.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		var old virtual.GridCellRef
		for _, cell := range snapshot.Cells {
			if cell.Active {
				old = cell.Ref
			}
		}
		calls := *w.source.calls
		if err := w.projection(true, true); err != nil {
			t.Fatal(err)
		}
		state, err := w.grid.State()
		if err != nil {
			t.Fatal(err)
		}
		if state.Active.RowKey != rowKey(999_998) || state.Stamp.Rows != 500_000 || state.Stamp.Revision != 2 {
			t.Fatal(state)
		}
		if w.source.calls.Cell != calls.Cell || w.source.calls.RowKey-calls.RowKey > 4 || w.source.calls.IndexOf-calls.IndexOf > 4 {
			t.Fatal("projection scanned/materialized", *w.source.calls, calls)
		}
		if err := w.grid.Activate(old); !errors.Is(err, virtual.ErrGridStale) {
			t.Fatal("stale press accepted", err)
		}
		if err := w.projection(false, false); err != nil {
			t.Fatal(err)
		}
		if err := w.grid.JumpToKey(rowKey(999_999), columnID(19)); err != nil {
			t.Fatal(err)
		}
		if err := w.projection(false, true); err != nil {
			t.Fatal(err)
		}
		state, err = w.grid.State()
		if err != nil || state.Active != (virtual.GridActive{}) {
			t.Fatal("deleted selection retained", state, err)
		}
	}); err != nil {
		t.Fatal(err)
	}
}
func TestMillionAndTenThousandSameViewportCommands(t *testing.T) {
	p, err := newProgram()
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	var first sourceCalls
	if err = p.owner.Do(func() {
		for i, w := range p.windows {
			if err := w.grid.SetViewport(virtual.GridViewport{Width: 960, Height: 1228, ScrollX: 480, ScrollY: (w.source.total / 2) * 24}); err != nil {
				t.Fatal(err)
			}
			*w.source.calls = sourceCalls{}
			snap, err := w.grid.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			visible := 0
			for _, cell := range snap.Cells {
				if !cell.Clip.Empty() {
					visible++
				}
			}
			if visible != 400 || len(snap.Cells) != 520 || *w.source.calls != (sourceCalls{52, 52, 520}) {
				t.Fatal("workload mismatch", visible, len(snap.Cells), *w.source.calls)
			}
			if i == 0 {
				first = *w.source.calls
			} else if first != *w.source.calls {
				t.Fatal("calls depend on total rows")
			}
			if err := w.last(); err != nil {
				t.Fatal(err)
			}
			if err := w.first(); err != nil {
				t.Fatal(err)
			}
			if err := w.pan(); err != nil {
				t.Fatal(err)
			}
			if err := w.grid.Page(1); err != nil {
				t.Fatal(err)
			}
			if err := w.saturateScroll(); err != nil {
				t.Fatal(err)
			}
			if err := uiValidate(w); err != nil {
				t.Fatal(err)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
}
func uiValidate(w *gridWindow) error { return ui.Validate(w.view()) }

func TestProgramIndependentWindowLifetimes(t *testing.T) {
	p, err := newProgram()
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	first, second := p.windows[0].windowRef.Context(), p.windows[1].windowRef.Context()
	if err = p.owner.Do(func() {
		if err := p.windows[0].last(); err != nil {
			t.Fatal(err)
		}
		state, err := p.windows[1].grid.State()
		if err != nil || state.Active != (virtual.GridActive{}) {
			t.Fatal("models shared", state, err)
		}
		if err := p.owner.CloseWindow(p.windows[0].windowRef); err != nil {
			t.Fatal(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if first.Err() == nil || second.Err() != nil || p.owner.Context().Err() != nil {
		t.Fatal("window lifetime leaked")
	}
	if err = p.owner.CloseWindow(p.windows[1].windowRef); err != nil {
		t.Fatal(err)
	}
	if second.Err() == nil || p.owner.Context().Err() == nil {
		t.Fatal("last close did not terminate")
	}
	if err = p.owner.Invalidate("large"); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
}

// Stable active identity survives a revision independently of the top-row
// anchor. Removing an odd anchor and reversing the source leaves the retained
// active row outside the bounded pool until an explicit reveal command.
func TestProjectionPreservesHiddenActiveUntilExplicitReveal(t *testing.T) {
	p, err := newProgram()
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	if err = withModelOwner(p.owner, func() error {
		w := p.windows[1]
		const oldTopRow = 999_979 // removed by the even-row filter
		if err := w.grid.SetViewport(virtual.GridViewport{Width: 800, Height: 508, ScrollX: 1600, ScrollY: oldTopRow * 24}); err != nil {
			return err
		}
		want := virtual.GridActive{RowKey: rowKey(999_998), ColumnID: columnID(19)}
		if err := w.grid.JumpToKey(want.RowKey, want.ColumnID); err != nil {
			return err
		}
		before, err := w.grid.State()
		if err != nil {
			return err
		}
		if before.Viewport.ScrollY != oldTopRow*24 {
			t.Fatalf("fixture anchor moved: %+v", before.Viewport)
		}
		if err := w.projection(true, true); err != nil {
			return err
		}
		retained, err := w.grid.State()
		if err != nil {
			return err
		}
		if retained.Active != want || retained.Stamp.Rows != 500_000 || retained.Stamp.Revision != 2 || retained.Viewport.ScrollY == 0 {
			t.Fatalf("active/anchor retention: %+v", retained)
		}
		projected, err := w.grid.Snapshot()
		if err != nil {
			return err
		}
		for _, cell := range projected.Cells {
			if cell.Active || cell.RowKey == want.RowKey {
				t.Fatal("retained hidden active unexpectedly materialized", cell)
			}
		}
		if err := w.grid.JumpToKey(want.RowKey, want.ColumnID); err != nil {
			return err
		}
		revealed, err := w.grid.State()
		if err != nil {
			return err
		}
		if revealed.Active != want || revealed.Stamp != retained.Stamp || revealed.Viewport.ScrollY != 0 {
			t.Fatalf("explicit reveal changed identity or missed row zero: %+v", revealed)
		}
		snapshot, err := w.grid.Snapshot()
		if err != nil {
			return err
		}
		visible := false
		for _, cell := range snapshot.Cells {
			visible = visible || cell.Active && cell.RowKey == want.RowKey && cell.ColumnID == want.ColumnID && !cell.Clip.Empty()
		}
		if !visible {
			t.Fatal("explicit reveal did not materialize visible active")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func withModelOwner(a *app.Application, fn func() error) error {
	var result error
	if err := a.Do(func() { result = fn() }); err != nil {
		return err
	}
	return result
}
