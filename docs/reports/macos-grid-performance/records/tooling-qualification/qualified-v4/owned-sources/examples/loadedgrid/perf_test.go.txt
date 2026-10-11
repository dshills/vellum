package main

import (
	"image"
	"testing"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func TestPerfConfigAndClosedWarmTrajectory(t *testing.T) {
	cfg := perfConfig{Rows: 1_000_000, Samples: 120, Warmup: 128}
	if e := cfg.validate(); e != nil {
		t.Fatal(e)
	}
	for _, invalid := range []perfConfig{{Rows: 9_999, Samples: 120, Warmup: 128}, {Rows: 10_000, Samples: 119, Warmup: 128}, {Rows: 10_000, Samples: 120, Warmup: 65}} {
		if e := invalid.validate(); e == nil {
			t.Fatalf("accepted%+v", invalid)
		}
	}
	x, y := 0, 0
	for i := 0; i < perfPeriod; i++ {
		dx, dy := perfDelta(i)
		x += dx
		y += dy
		if x < 0 || y < 0 || x > 384 || y > 384 {
			t.Fatalf("unbounded trajectory%d,%d", x, y)
		}
	}
	if x != 0 || y != 0 {
		t.Fatalf("warm trajectory does not return%d,%d", x, y)
	}
}
func TestPerfVariableRowsUseStoredSourceAndExactIndex(t *testing.T) {
	s := &perfVariableSource{&fullSource{rows: 10_000}}
	for i := 0; i < 10_000; i++ {
		start, e := s.RowOffset(i)
		if e != nil {
			t.Fatal(e)
		}
		end, e := s.RowOffset(i + 1)
		if e != nil {
			t.Fatal(e)
		}
		for _, p := range []int{start, end - 1} {
			got, e := s.RowAt(p)
			if e != nil || got != i {
				t.Fatalf("row%d pixel%d: %d %v", i, p, got, e)
			}
		}
	}
	if _, e := s.RowAt(perfOffset(s.rows)); e == nil {
		t.Fatal("accepted end pixel")
	}
	if _, e := s.RowOffset(s.rows + 1); e == nil {
		t.Fatal("accepted outside boundary")
	}
}
func perfFakeGeometry(rows, columns int) cogent.GridInspection {
	g := cogent.GridInspection{Frames: 1, Generation: 1, Bounds: image.Rect(9, 30, 809, 630), Scale: 1, Viewport: virtual.GridViewport{Width: 800, Height: 624}}
	for row := 0; row < rows; row++ {
		for col := 0; col < columns; col++ {
			g.Cells = append(g.Cells, cogent.GridCellGeometry{ID: cogent.GridCellID("data", key(row), column(col)), Row: row, Column: col, Bounds: image.Rect(9+col*100, 30+row*12, 9+(col+1)*100, 30+(row+1)*12)})
		}
	}
	return g
}
func TestPerfGeometryCountsRenderedClipsAndIgnoresDomainOrdinal(t *testing.T) {
	g := perfFakeGeometry(50, 8)
	g.Cells = append(g.Cells, cogent.GridCellGeometry{ID: "overscan", Bounds: image.Rectangle{}})
	a, e := checkPerfGeometry(g)
	if e != nil {
		t.Fatal(e)
	}
	if a.Cells != 400 || a.Rows != 50 || a.Columns != 8 {
		t.Fatalf("wrong visible workload%+v", a)
	}
	for i := range g.Cells {
		g.Cells[i].Row += 1_000_000
	}
	b, e := checkPerfGeometry(g)
	if e != nil || a.LayoutSHA256 != b.LayoutSHA256 {
		t.Fatalf("ordinal changed measured clip%v", e)
	}
	g.Cells[0].Bounds = image.Rectangle{}
	if _, e := checkPerfGeometry(g); e == nil {
		t.Fatal("399 actual clips accepted as400")
	}
	g = perfFakeGeometry(50, 8)
	g.Cells[1].ID = g.Cells[0].ID
	if _, e := checkPerfGeometry(g); e == nil {
		t.Fatal("duplicate measured identity accepted")
	}
}
func TestPerfSourceCounterOverflowRefuses(t *testing.T) {
	a := countSnapshot{Keys: 2, Inverse: 3, Cells: 4}
	got, e := perfCountDelta(a, countSnapshot{Keys: 5, Inverse: 5, Cells: 404})
	if e != nil || got.Cells != 400 || got.Keys != 3 || got.Inverse != 2 {
		t.Fatalf("%+v %v", got, e)
	}
	if _, e = perfCountDelta(a, countSnapshot{Keys: 1, Inverse: 3, Cells: 4}); e == nil {
		t.Fatal("counter wrap accepted")
	}
}
