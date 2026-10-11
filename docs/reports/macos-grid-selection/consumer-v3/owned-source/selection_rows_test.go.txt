package main

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/dshills/vellum/virtual"
)

func TestSelectionIndexImmutablePureAndConcurrent(t *testing.T) {
	for _, count := range []int{1, 5, 10_000, 1_000_000} {
		for _, reverse := range []bool{false, true} {
			for _, filtered := range []bool{false, true} {
				s, err := newSource("selection-index", count)
				if err != nil {
					t.Fatal(err)
				}
				s, err = s.projected(reverse, filtered)
				if err != nil {
					t.Fatal(err)
				}
				index, err := s.SelectionSnapshot()
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := index.(selectionRowIndex); !ok {
					t.Fatal("selection snapshot retained mutable owner source", reflect.TypeOf(index))
				}
				if index.Identity() != s.identity || index.Revision() != s.revision || index.RowCount() != s.RowCount() {
					t.Fatal("snapshot metadata mismatch")
				}
				before := *s.calls
				var wg sync.WaitGroup
				for range 8 {
					wg.Go(func() {
						for range 32 {
							for _, row := range []int{0, index.RowCount() / 2, index.RowCount() - 1} {
								key, e := index.RowKey(row)
								i, found, ie := index.IndexOf(key)
								if e != nil || ie != nil || !found || i != row {
									t.Errorf("pure snapshot inverse %d/%s/%d %t: %v/%v", row, key, i, found, e, ie)
								}
							}
						}
					})
				}
				wg.Wait()
				if *s.calls != before {
					t.Fatal("detached index mutates source diagnostics", before, *s.calls)
				}
				if _, e := index.RowKey(-1); !errors.Is(e, virtual.ErrGridRange) {
					t.Fatal(e)
				}
				if _, e := index.RowKey(index.RowCount()); !errors.Is(e, virtual.ErrGridRange) {
					t.Fatal(e)
				}
				if _, found, e := index.IndexOf("bad-key"); e != nil || found {
					t.Fatal("bad key admitted", found, e)
				}
				next, e := s.projected(!reverse, !filtered)
				if e != nil || index.Revision() != s.revision || index.RowCount() != s.RowCount() || next.revision == index.Revision() {
					t.Fatal("new view mutated old captured index", e)
				}
			}
		}
	}
	for _, s := range []*gridSource{nil, {}, {identity: "x", revision: 1, total: 0}, {identity: "x", revision: 0, total: 1}} {
		if _, err := s.SelectionSnapshot(); !errors.Is(err, errSourceConfig) {
			t.Fatal("malformed snapshot accepted", s, err)
		}
	}
}

func TestCapturedAllSelectionProjectionClearAndWorkerVisit(t *testing.T) {
	p, err := newProgramModes(1024, 760, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	w := p.windows[1]
	var captured virtual.GridSelection
	var stamp virtual.GridSelectionState
	if err := withModelOwner(p.owner, func() error {
		if err := w.grid.SelectAll(); err != nil {
			return err
		}
		for _, key := range []string{rowKey(1), rowKey(2)} {
			if err := w.grid.SelectToKey(key, columnID(19), virtual.GridSelectToggle); err != nil {
				return err
			}
		}
		var err error
		captured, err = w.grid.Selection()
		if err != nil {
			return err
		}
		stamp = captured.State()
		if stamp.CapturedRows != 999_998 || stamp.Scope.Rows != 1_000_000 || stamp.Overrides != 2 {
			t.Fatal(stamp)
		}
		if err := w.projection(true, true); err != nil {
			return err
		}
		if err := w.setVariableRows(true); err != nil {
			return err
		}
		s, err := w.grid.State()
		if err != nil {
			return err
		}
		if s.Selection != stamp || s.Stamp.Rows != 500_000 || !s.VariableHeights {
			t.Fatal("same-incarnation view changed captured universe", s)
		}
		return w.grid.ClearSelection()
	}); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]bool{rowKey(0): true, rowKey(1): false, rowKey(2): false, rowKey(3): true, rowKey(999_999): true} {
		got, err := captured.Contains(key)
		if err != nil || got != want {
			t.Fatal(key, got, want, err)
		}
	}
	// Deliberately visit only a bounded prefix outside the owner; cancellation
	// is the worker-side policy rather than an owner million-row scan.
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	err = captured.VisitContext(ctx, func(key string) error {
		selected, e := captured.Contains(key)
		if e != nil || !selected {
			t.Fatal("visit emitted unselected key", key, e)
		}
		count++
		if count == 17 {
			cancel()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) || count != 17 || captured.State() != stamp {
		t.Fatal("worker visit/capture changed", count, err, captured.State())
	}
	if err := withModelOwner(p.owner, func() error {
		for _, model := range p.windows {
			s, e := model.grid.State()
			if e != nil {
				return e
			}
			if s.Selection.CapturedRows != 0 {
				t.Fatal("Clear/other window selection", s)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCapturedRangeAddsKeyOutsideBaseAndIncarnationClears(t *testing.T) {
	p, err := newProgramModes(1024, 760, true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	w := p.windows[0]
	var captured virtual.GridSelection
	if err := withModelOwner(p.owner, func() error {
		if err := w.projection(false, true); err != nil {
			return err
		}
		if err := w.grid.SelectToKey(rowKey(2), columnID(0), virtual.GridSelectReplace); err != nil {
			return err
		}
		if err := w.grid.SelectToKey(rowKey(8), columnID(0), virtual.GridSelectExtend); err != nil {
			return err
		}
		s, e := w.grid.State()
		if e != nil {
			return e
		}
		if s.Selection.CapturedRows != 4 || s.Selection.Scope.Rows != 5000 {
			t.Fatal("range must use captured even-only view", s.Selection)
		}
		if err := w.projection(false, false); err != nil {
			return err
		}
		if err := w.grid.SelectToKey(rowKey(3), columnID(0), virtual.GridSelectToggle); err != nil {
			return err
		}
		captured, e = w.grid.Selection()
		if e != nil {
			return e
		}
		if captured.State().CapturedRows != 5 || captured.State().Overrides != 1 || captured.State().Scope.Rows != 5000 {
			t.Fatal("outside-base sparse addition", captured.State())
		}
		next, e := newSource("new-incarnation", 10000)
		if e != nil {
			return e
		}
		if e = w.grid.ReplaceSource(rowSource(next, w.metricCalls, true)); e != nil {
			return e
		}
		w.source = next
		s, e = w.grid.State()
		if e != nil {
			return e
		}
		if s.Selection.CapturedRows != 0 {
			t.Fatal("new incarnation retained selection", s)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := captured.VisitContext(context.Background(), func(key string) error { keys = append(keys, key); return nil }); err != nil {
		t.Fatal(err)
	}
	want := []string{rowKey(2), rowKey(4), rowKey(6), rowKey(8), rowKey(3)}
	if !reflect.DeepEqual(keys, want) {
		t.Fatal("captured order/membership changed after incarnation replacement", keys, want)
	}
}

func TestCapturedAllIndexedOwnerWorkDoesNotScaleWithRowCount(t *testing.T) {
	p, err := newProgramModes(1024, 760, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	var first sourceCalls
	if err := withModelOwner(p.owner, func() error {
		for i, w := range p.windows {
			*w.source.calls = sourceCalls{}
			if err := w.grid.SelectAll(); err != nil {
				return err
			}
			selected, err := w.grid.Selection()
			if err != nil {
				return err
			}
			if selected.State().CapturedRows != w.source.RowCount() {
				t.Fatal("All count mismatch", selected.State())
			}
			calls := *w.source.calls
			if calls.RowKey+calls.IndexOf > 2048 || calls.Cell != 0 {
				t.Fatal("All scanned/materialized owner source", calls)
			}
			if i == 0 {
				first = calls
			} else if first != calls {
				t.Fatal("same viewport All indexed owner work depends on 10k versus1M", first, calls)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
