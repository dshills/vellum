package virtual

import (
	"errors"
	"reflect"
	"testing"
)

func TestGridVariableMetricsFailureIsAtomic(t *testing.T) {
	sentinel := errors.New("metrics unavailable")
	for _, operation := range []string{"viewport", "jump", "page", "move", "activate", "snapshot"} {
		t.Run(operation, func(t *testing.T) {
			s := variableSource(t, []int{20, 30, 40, 50})
			g := mustGrid(t, s, gridColumns(2, 20), GridConfig{RowHeight: 1}, GridViewport{Width: 40, Height: 15})
			ref := gridSnapshot(t, g).Cells[0].Ref
			before := gridState(t, g)
			s.atHook = func(int) (int, error) { return 0, sentinel }
			var err error
			switch operation {
			case "viewport":
				v := before.Viewport
				v.Height++
				err = g.SetViewport(v)
			case "jump":
				err = g.JumpToKey("r:2", "c:1")
			case "page":
				err = g.Page(1)
			case "move":
				err = g.Move(1, 0)
			case "activate":
				err = g.Activate(ref)
			case "snapshot":
				var out GridSnapshot
				out, err = g.Snapshot()
				if !reflect.DeepEqual(out, GridSnapshot{}) {
					t.Fatal("partially published snapshot")
				}
			}
			if !errors.Is(err, ErrGridSource) || !errors.Is(err, sentinel) || gridState(t, g) != before {
				t.Fatal("failure did not preserve cause/state", err, g.state(), before)
			}
			s.atHook = nil
			if err := g.Activate(ref); err != nil {
				t.Fatal("ref retired by failure", err)
			}
		})
	}
}

func TestGridVariableMetricsRevisionDriftIsAtomic(t *testing.T) {
	s := variableSource(t, []int{20, 30, 40, 50})
	g := mustGrid(t, s, gridColumns(1, 20), GridConfig{RowHeight: 1}, GridViewport{Width: 20, Height: 15})
	before := gridState(t, g)
	ref := gridSnapshot(t, g).Cells[0].Ref
	s.atHook = func(pixel int) (int, error) {
		s.revision++
		return s.index.RowAt(pixel)
	}
	v := before.Viewport
	v.ScrollY = 25
	if err := g.SetViewport(v); !errors.Is(err, ErrGridStale) || g.state() != before {
		t.Fatal("revision drift accepted", err, g.state())
	}
	s.atHook = nil
	s.revision = before.Stamp.Revision // Restore the test source; Grid never rolls it back.
	if err := g.Activate(ref); err != nil {
		t.Fatal(err)
	}
}

func TestGridVariableUnchangedRevisionRejectsGeometryChange(t *testing.T) {
	s := variableSource(t, []int{20, 20, 20})
	g := mustGrid(t, s, gridColumns(1, 20), GridConfig{RowHeight: 20}, GridViewport{Width: 20, Height: 5})
	before := gridState(t, g)
	for _, candidate := range []GridSource{gridSource(3), variableSource(t, []int{20, 30, 20})} {
		if err := g.ReplaceSource(candidate); !errors.Is(err, ErrGridSource) || gridState(t, g) != before {
			t.Fatal("same revision changed metrics mode/extent", err, g.state())
		}
	}
	oldIndex := s.index
	s.index, _ = NewRowHeights([]int{20, 30, 20})
	if err := g.Refresh(); !errors.Is(err, ErrGridSource) || g.state() != before {
		t.Fatal("unchanged-stamp refresh accepted changed extent", err)
	}
	s.index = oldIndex
	if gridState(t, g) != before {
		t.Fatal("refusal mutated grid")
	}
}

func TestGridVariableEmptyAndMalformedMetadata(t *testing.T) {
	empty := variableSource(t, nil)
	empty.atHook = func(int) (int, error) { t.Fatal("empty source pixel lookup"); return 0, nil }
	g := mustGrid(t, empty, gridColumns(1, 20), GridConfig{RowHeight: 1}, GridViewport{Width: 20, Height: 100})
	if out := gridSnapshot(t, g); len(out.Cells) != 0 || out.State.TotalHeight != 0 || !out.State.VariableHeights {
		t.Fatal(out)
	}
	for _, corrupt := range []func(*variableRowSource){
		func(s *variableRowSource) { s.offsetHook = func(int) (int, error) { return 1, nil } },
		func(s *variableRowSource) { s.offsetHook = func(int) (int, error) { return 0, nil } },
		func(s *variableRowSource) { s.atHook = func(int) (int, error) { return -1, nil } },
		func(s *variableRowSource) { s.atHook = func(int) (int, error) { return 2, nil } },
	} {
		s := variableSource(t, []int{20, 20})
		corrupt(s)
		if g, err := NewGrid(s, gridColumns(1, 20), GridConfig{RowHeight: 1}, GridViewport{Width: 20, Height: 5}); g != nil || !errors.Is(err, ErrGridSource) {
			t.Fatal("invalid accessed metrics accepted", g, err)
		}
	}
}

type proceduralVariableRows struct {
	*indexedGridSource
	offsets, pixels int
}

func (s *proceduralVariableRows) RowOffset(i int) (int, error) {
	s.offsets++
	if i < 0 || i > s.rows {
		return 0, ErrGridRange
	}
	return i/2*48 + i%2*11, nil
}
func (s *proceduralVariableRows) RowAt(pixel int) (int, error) {
	s.pixels++
	if pixel < 0 || pixel >= s.rows/2*48+s.rows%2*11 {
		return 0, ErrGridRange
	}
	row := pixel / 48 * 2
	if pixel%48 >= 11 {
		row++
	}
	return row, nil
}

func TestGridVariableMillionRowsHasSameBoundedWork(t *testing.T) {
	type work struct {
		gridCalls
		offsets, pixels, cells, headers int
	}
	var want work
	var first GridSnapshot
	for _, rows := range []int{10_000, 1_000_000} {
		s := &proceduralVariableRows{indexedGridSource: gridSource(rows)}
		g := mustGrid(t, s, gridColumns(20, 120), GridConfig{RowHeight: 7, HeaderHeight: 28, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1}, GridViewport{Width: 800, Height: 529, ScrollX: 333, ScrollY: 1037})
		s.calls, s.offsets, s.pixels = gridCalls{}, 0, 0
		out := gridSnapshot(t, g)
		got := work{s.calls, s.offsets, s.pixels, len(out.Cells), len(out.Headers)}
		materializedRows := len(out.Cells) / len(out.Headers)
		if got.RowKey != materializedRows || got.IndexOf != materializedRows || got.Cell != len(out.Cells) || got.offsets > 8*materializedRows+32 || got.pixels > 5*materializedRows+32 {
			t.Fatal("source-sized or unexpected work", rows, got)
		}
		if rows == 10_000 {
			want, first = got, out
		} else {
			if got != want {
				t.Fatal("work grew with dataset", got, want)
			}
			for i, cell := range out.Cells {
				other := first.Cells[i]
				if cell.Rect != other.Rect || cell.Clip != other.Clip || cell.RowKey != other.RowKey || cell.Text != other.Text {
					t.Fatal("common viewport differs", cell, other)
				}
			}
		}
		s.calls, s.offsets, s.pixels = gridCalls{}, 0, 0
		if err := g.JumpToKey("r:9999", "c:19"); err != nil || s.calls.Cell != 0 || s.calls.RowKey > 4 || s.calls.IndexOf > 4 || s.offsets > 8*256+32 || s.pixels > 5*256+32 {
			t.Fatal("far jump enumerated dataset or cells", err, s.calls, s.offsets, s.pixels)
		}
	}
	t.Logf("10k/1M identical snapshot work: %+v", want)
}
