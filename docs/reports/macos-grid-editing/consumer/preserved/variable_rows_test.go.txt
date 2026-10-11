package main

import (
	"errors"
	"testing"

	"github.com/dshills/vellum/virtual"
)

func bruteProjectedHeights(source *gridSource) []int {
	out := make([]int, source.RowCount())
	for i := range out {
		row := i
		if source.reverse {
			row = len(out) - 1 - row
		}
		if source.filtered {
			row *= 2
		}
		out[i] = [...]int{23, 35, 47, 29}[row%4]
	}
	return out
}

func TestVariableMetricsBrutePrefixAndPixelInverse(t *testing.T) {
	for _, total := range []int{1, 2, 3, 4, 5, 9, 37, 1000} {
		for _, reverse := range []bool{false, true} {
			for _, filtered := range []bool{false, true} {
				base, err := newSource("metrics", total)
				if err != nil {
					t.Fatal(err)
				}
				base, err = base.projected(reverse, filtered)
				if err != nil {
					t.Fatal(err)
				}
				metrics := &variableGridSource{gridSource: base, metrics: &rowMetricCalls{}}
				heights := bruteProjectedHeights(base)
				pixel := 0
				for row, height := range heights {
					offset, err := metrics.RowOffset(row)
					if err != nil || offset != pixel {
						t.Fatal("prefix", total, reverse, filtered, row, offset, pixel, err)
					}
					for p := pixel; p < pixel+height; p++ {
						actual, err := metrics.RowAt(p)
						if err != nil || actual != row {
							t.Fatal("pixel inverse", total, reverse, filtered, p, actual, row, err)
						}
					}
					pixel += height
				}
				if extent, err := metrics.RowOffset(len(heights)); err != nil || extent != pixel {
					t.Fatal("terminal offset", extent, pixel, err)
				}
				for _, boundary := range []int{-1, len(heights) + 1} {
					if _, err := metrics.RowOffset(boundary); !errors.Is(err, virtual.ErrGridRange) {
						t.Fatal("boundary accepted", boundary, err)
					}
				}
				for _, p := range []int{-1, pixel, pixel + 1} {
					if _, err := metrics.RowAt(p); !errors.Is(err, virtual.ErrGridRange) {
						t.Fatal("pixel accepted", p, err)
					}
				}
			}
		}
	}
	for _, metrics := range []*variableGridSource{nil, {}, {gridSource: &gridSource{}}} {
		if _, err := metrics.RowOffset(0); !errors.Is(err, errSourceConfig) {
			t.Fatal("invalid metrics offset", err)
		}
		if _, err := metrics.RowAt(0); !errors.Is(err, errSourceConfig) {
			t.Fatal("invalid metrics pixel", err)
		}
	}
}

func TestVariableMetricsMillionRowQueriesAvoidDomainEnumeration(t *testing.T) {
	for _, total := range []int{10_000, 1_000_000} {
		for _, reverse := range []bool{false, true} {
			for _, filtered := range []bool{false, true} {
				base, err := newSource("indexed", total)
				if err != nil {
					t.Fatal(err)
				}
				base, err = base.projected(reverse, filtered)
				if err != nil {
					t.Fatal(err)
				}
				metrics := &variableGridSource{gridSource: base, metrics: &rowMetricCalls{}}
				for _, row := range []int{0, 1, base.RowCount() / 2, base.RowCount() - 1} {
					start, err := metrics.RowOffset(row)
					if err != nil {
						t.Fatal(err)
					}
					end, err := metrics.RowOffset(row + 1)
					if err != nil {
						t.Fatal(err)
					}
					for _, pixel := range []int{start, end - 1} {
						if actual, err := metrics.RowAt(pixel); err != nil || actual != row {
							t.Fatal("large indexed query", total, row, pixel, actual, err)
						}
					}
				}
				if *base.calls != (sourceCalls{}) || *metrics.metrics != (rowMetricCalls{Offset: 8, At: 8}) {
					t.Fatal("metric callback enumerated domain data", total, *base.calls, *metrics.metrics)
				}
			}
		}
	}
}

func TestVariableMetricsHeightsFollowStableKeysAcrossProjection(t *testing.T) {
	base, err := newSource("stable-height", 10_000)
	if err != nil {
		t.Fatal(err)
	}
	original := &variableGridSource{gridSource: base}
	originalExtent, err := original.RowOffset(base.RowCount())
	if err != nil {
		t.Fatal(err)
	}
	next, err := base.projected(true, true)
	if err != nil {
		t.Fatal(err)
	}
	projected := &variableGridSource{gridSource: next}
	for _, row := range []int{0, 2, 4, 500, 9998} {
		index, found := next.rowIndex(rowKey(row))
		if !found {
			t.Fatal("retained key missing", row)
		}
		start, err := projected.RowOffset(index)
		if err != nil {
			t.Fatal(err)
		}
		end, err := projected.RowOffset(index + 1)
		if err != nil || end-start != domainRowHeight(row) {
			t.Fatal("height followed index rather than stable key", row, index, end-start, err)
		}
	}
	if extent, err := original.RowOffset(base.RowCount()); err != nil || extent != originalExtent || base.Revision() != 1 || next.Revision() != 2 {
		t.Fatal("projection mutated prior height revision", extent, originalExtent, err)
	}
}

func TestVariableModeHeightChangesAreIndependentImmutableRevisions(t *testing.T) {
	p, err := newProgramMode(1024, 760, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	if err := withModelOwner(p.owner, func() error {
		w := p.windows[1]
		original := rowSource(w.source, w.metricCalls, true).(*variableGridSource)
		originalExtent, err := original.RowOffset(original.RowCount())
		if err != nil {
			return err
		}
		if err := w.grid.JumpToKey(rowKey(4), columnID(19)); err != nil {
			return err
		}
		before, err := w.grid.Snapshot()
		if err != nil {
			return err
		}
		var retired virtual.GridCellRef
		for _, cell := range before.Cells {
			if cell.Active {
				retired = cell.Ref
			}
		}
		if err := w.setVariableRows(false); err != nil {
			return err
		}
		fixed, err := w.grid.State()
		if err != nil {
			return err
		}
		if fixed.VariableHeights || fixed.Stamp.Revision != 2 || fixed.Active.RowKey != rowKey(4) {
			t.Fatal("height toggle lost state", fixed)
		}
		if _, optional := rowSource(w.source, w.metricCalls, false).(virtual.GridRowMetrics); optional {
			t.Fatal("default fixed source unexpectedly opted into metrics")
		}
		if err := w.grid.Activate(retired); !errors.Is(err, virtual.ErrGridStale) {
			t.Fatal("height-change stale press accepted", err)
		}
		if err := w.setVariableRows(true); err != nil {
			return err
		}
		variable, err := w.grid.State()
		if err != nil {
			return err
		}
		if !variable.VariableHeights || variable.Stamp.Revision != 3 || variable.Active != fixed.Active {
			t.Fatal("variable height re-enable lost identity", variable)
		}
		if err := w.setVariableRows(true); err != nil {
			return err
		}
		unchanged, err := w.grid.State()
		if err != nil {
			return err
		}
		if unchanged != variable || p.windows[0].source.Revision() != 1 {
			t.Fatal("no-op toggle or independent window changed", unchanged, variable)
		}
		if extent, err := original.RowOffset(original.RowCount()); err != nil || extent != originalExtent || original.Revision() != 1 {
			t.Fatal("height toggle mutated old metrics", extent, originalExtent, err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

var _ virtual.GridRowMetrics = (*variableGridSource)(nil)
