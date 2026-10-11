package virtual

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type selectionOrdinal struct {
	id            string
	revision      uint64
	rows          int
	reverse, even bool
	keys, lookups atomic.Int64
	keyHook       func(int) (string, error)
	lookupHook    func(string) (int, bool, error)
}

func (s *selectionOrdinal) Identity() string { return s.id }
func (s *selectionOrdinal) Revision() uint64 { return s.revision }
func (s *selectionOrdinal) RowCount() int    { return s.rows }
func (s *selectionOrdinal) RowKey(i int) (string, error) {
	s.keys.Add(1)
	if s.keyHook != nil {
		return s.keyHook(i)
	}
	if i < 0 || i >= s.rows {
		return "", ErrGridRange
	}
	if s.reverse {
		i = s.rows - 1 - i
	}
	if s.even {
		i *= 2
	}
	return "r:" + strconv.Itoa(i), nil
}
func (s *selectionOrdinal) IndexOf(key string) (int, bool, error) {
	s.lookups.Add(1)
	if s.lookupHook != nil {
		return s.lookupHook(key)
	}
	i, err := strconv.Atoi(strings.TrimPrefix(key, "r:"))
	if err != nil || i < 0 || key != "r:"+strconv.Itoa(i) {
		return 0, false, nil
	}
	if s.even {
		if i%2 != 0 {
			return 0, false, nil
		}
		i /= 2
	}
	if i >= s.rows {
		return 0, false, nil
	}
	if s.reverse {
		i = s.rows - 1 - i
	}
	return i, true, nil
}

type selectionView struct {
	*indexedGridSource
	index       *selectionOrdinal
	captures    int
	captureHook func() (GridRowIndex, error)
}

func newSelectionView(rows int, revision uint64, reverse, even bool) *selectionView {
	i := &selectionOrdinal{id: "parts:seed7", revision: revision, rows: rows, reverse: reverse, even: even}
	s := &selectionView{indexedGridSource: gridSource(rows), index: i}
	s.revision = revision
	s.rowHook = i.RowKey
	s.indexHook = i.IndexOf
	return s
}
func (s *selectionView) SelectionSnapshot() (GridRowIndex, error) {
	s.captures++
	if s.captureHook != nil {
		return s.captureHook()
	}
	return s.index, nil
}
func selectionGrid(t *testing.T, s GridSource, cfg GridConfig) *Grid {
	t.Helper()
	cfg.RowSelection = true
	if cfg.RowHeight == 0 {
		cfg.RowHeight = 20
	}
	return mustGrid(t, s, gridColumns(20, 120), cfg, GridViewport{Width: 800, Height: 140})
}
func captured(t *testing.T, g *Grid) GridSelection {
	t.Helper()
	s, e := g.Selection()
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func selectedKeys(t *testing.T, s GridSelection) []string {
	t.Helper()
	var keys []string
	if err := s.VisitContext(context.Background(), func(key string) error { keys = append(keys, key); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(keys) != s.State().CapturedRows {
		t.Fatal("count/visit mismatch", len(keys), s.State())
	}
	return keys
}
func TestGridSelectionRandomRangesAndToggles(t *testing.T) {
	rng := rand.New(rand.NewSource(91457))
	for trial := 0; trial < 40; trial++ {
		s := newSelectionView(100, 1, false, false)
		g := selectionGrid(t, s, GridConfig{})
		want := map[string]bool{}
		anchor := ""
		for step := 0; step < 100; step++ {
			row := rng.Intn(100)
			key := "r:" + strconv.Itoa(row)
			mode := GridSelectMode(rng.Intn(3))
			if mode == GridSelectToggle && anchor == "" {
				mode = GridSelectReplace
			}
			switch mode {
			case GridSelectReplace:
				want = map[string]bool{key: true}
				anchor = key
			case GridSelectExtend:
				if anchor == "" {
					anchor = key
				}
				a, _ := strconv.Atoi(strings.TrimPrefix(anchor, "r:"))
				want = map[string]bool{}
				for i := min(a, row); i <= max(a, row); i++ {
					want["r:"+strconv.Itoa(i)] = true
				}
			case GridSelectToggle:
				want[key] = !want[key]
			}
			if err := g.SelectToKey(key, "c:2", mode); err != nil {
				t.Fatal(trial, step, err)
			}
			selection := captured(t, g)
			count := 0
			for i := 0; i < 100; i++ {
				k := "r:" + strconv.Itoa(i)
				got, e := selection.Contains(k)
				if e != nil || got != want[k] {
					t.Fatal(trial, step, k, got, e)
				}
				if want[k] {
					count++
				}
			}
			if selection.State().CapturedRows != count || selection.State().AnchorKey != anchor {
				t.Fatal(selection.State(), count, anchor)
			}
			selectedKeys(t, selection)
			for _, cell := range gridSnapshot(t, g).Cells {
				if cell.Selected != want[cell.RowKey] {
					t.Fatal("row projection", cell)
				}
			}
		}
	}
}
func TestGridSelectionSnapshotRetentionAndNewKeys(t *testing.T) {
	g := selectionGrid(t, newSelectionView(10, 1, false, false), GridConfig{})
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	old := captured(t, g)
	if err := g.SelectToKey("r:2", "c:0", GridSelectToggle); err != nil {
		t.Fatal(err)
	}
	if old.State().CapturedRows != 10 || captured(t, g).State().CapturedRows != 9 {
		t.Fatal("snapshot mutated")
	}
	if err := g.ReplaceSource(newSelectionView(12, 2, true, false)); err != nil {
		t.Fatal(err)
	}
	if err := g.SelectToKey("r:11", "c:0", GridSelectToggle); err != nil {
		t.Fatal(err)
	}
	if got := selectedKeys(t, captured(t, g)); !reflect.DeepEqual(got, []string{"r:0", "r:1", "r:3", "r:4", "r:5", "r:6", "r:7", "r:8", "r:9", "r:11"}) {
		t.Fatal(got)
	}
	if err := g.ReplaceSource(newSelectionView(6, 3, false, true)); err != nil {
		t.Fatal(err)
	}
	if captured(t, g).State().CapturedRows != 10 {
		t.Fatal("filtered view changed captured domain")
	}
	for _, c := range gridSnapshot(t, g).Cells {
		if c.Selected != (c.RowKey != "r:2" && c.RowKey != "r:10") {
			t.Fatal(c)
		}
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if yes, e := old.Contains("r:2"); e != nil || !yes {
					t.Error(e, yes)
				}
			}
		}()
	}
	if err := g.ClearSelection(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if len(selectedKeys(t, old)) != 10 {
		t.Fatal("old immutable snapshot lost keys")
	}
	// Hidden Shift anchor deterministically starts a new range at the target.
	if err := g.SelectToKey("r:4", "c:0", GridSelectReplace); err != nil {
		t.Fatal(err)
	}
	if err := g.ReplaceSource(newSelectionView(2, 4, false, false)); err != nil {
		t.Fatal(err)
	}
	if err := g.SelectToKey("r:1", "c:0", GridSelectExtend); err != nil {
		t.Fatal(err)
	}
	if s := captured(t, g).State(); s.CapturedRows != 1 || s.AnchorKey != "r:1" {
		t.Fatal(s)
	}
}
func TestGridSelectionCapsFailuresAndRecovery(t *testing.T) {
	s := newSelectionView(20, 1, false, false)
	g := selectionGrid(t, s, GridConfig{MaxSelectionKeys: 1})
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	if err := g.SelectToKey("r:1", "c:0", GridSelectToggle); err != nil {
		t.Fatal(err)
	}
	before := gridState(t, g)
	ref := gridSnapshot(t, g).Cells[0].Ref
	if err := g.SelectToKey("r:2", "c:0", GridSelectToggle); !errors.Is(err, ErrGridCapacity) || gridState(t, g) != before {
		t.Fatal(err, g.state())
	}
	if err := g.SelectCell(ref, GridSelectMode(255)); !errors.Is(err, ErrGridInvalid) || gridState(t, g) != before {
		t.Fatal(err)
	}
	sentinel := errors.New("snapshot unavailable")
	s.captureHook = func() (GridRowIndex, error) { return nil, sentinel }
	if err := g.SelectToKey("r:2", "c:0", GridSelectReplace); !errors.Is(err, ErrGridSource) || !errors.Is(err, sentinel) || gridState(t, g) != before {
		t.Fatal(err)
	}
	s.captureHook = nil
	if err := g.SelectCell(ref, GridSelectExtend); err != nil {
		t.Fatal("failed operation retired ref", err)
	}
	// Invalid retained index cannot prevent explicit clear recovery.
	s.index.revision++
	if _, err := g.State(); !errors.Is(err, ErrGridSource) {
		t.Fatal(err)
	}
	if err := g.ClearSelection(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.State(); err != nil {
		t.Fatal(err)
	}
	// New incarnation similarly avoids touching a broken old captured index.
	s.index.revision = s.revision
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	s.index.rows++
	next := newSelectionView(5, 1, false, false)
	next.id = "new"
	next.index.id = "new"
	if err := g.ReplaceSource(next); err != nil || g.state().Selection.CapturedRows != 0 {
		t.Fatal(err, g.state())
	}
	if _, err := NewGrid(gridSource(20), gridColumns(1, 20), GridConfig{RowHeight: 20, RowSelection: true}, GridViewport{Width: 20, Height: 20}); !errors.Is(err, ErrGridSource) {
		t.Fatal(err)
	}
}
func TestGridSelectionVisitCancellationErrorsAndGeneration(t *testing.T) {
	s := newSelectionView(100, 1, false, false)
	g := selectionGrid(t, s, GridConfig{})
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	selection := captured(t, g)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	if err := selection.VisitContext(ctx, func(string) error { calls++; cancel(); return nil }); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal(err, calls)
	}
	sentinel := errors.New("callback failed")
	if err := selection.VisitContext(context.Background(), func(string) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	before := gridState(t, g)
	if err := g.SelectAll(); err != nil || gridState(t, g) != before {
		t.Fatal("all no-op changed generation", err)
	}
	g.generation = math.MaxUint64
	if err := g.ClearSelection(); !errors.Is(err, ErrGridExhausted) || g.selection.state.CapturedRows != 100 {
		t.Fatal(err)
	}
	g.generation = before.Generation
	g.selection.state.Revision = math.MaxUint64
	if err := g.ClearSelection(); !errors.Is(err, ErrGridExhausted) || g.selection.state.CapturedRows != 100 {
		t.Fatal(err)
	}
}
func TestGridSelectionMillionRowsUsesBoundedWork(t *testing.T) {
	type counts struct {
		calls         gridCalls
		keys, lookups int64
		captures      int
	}
	var first counts
	for _, rows := range []int{10_000, 1_000_000} {
		s := newSelectionView(rows, 1, false, false)
		g := selectionGrid(t, s, GridConfig{})
		s.calls = gridCalls{}
		s.index.keys.Store(0)
		s.index.lookups.Store(0)
		if err := g.SelectAll(); err != nil {
			t.Fatal(err)
		}
		all := counts{s.calls, s.index.keys.Load(), s.index.lookups.Load(), s.captures}
		if rows == 10_000 {
			first = all
		} else if all != first {
			t.Fatal("all work scales with source", first, all)
		}
		if all.calls.Cell != 0 || all.calls.RowKey > 16 || all.keys > 40 || all.lookups > 40 || g.state().Selection.CapturedRows != rows {
			t.Fatal(all, g.state())
		}
		if err := g.SelectToKey("r:9999", "c:19", GridSelectExtend); err != nil {
			t.Fatal(err)
		}
		if g.state().Selection.CapturedRows != 10_000 {
			t.Fatal(g.state())
		}
	}
	t.Logf("10k/1M SelectAll identical bounded work: %+v", first)
}

func TestGridSelectionCountOverflowAndAtomicDrift(t *testing.T) {
	s := newSelectionView(math.MaxInt, 1, false, false)
	g := mustGrid(t, s, []GridColumn{{ID: "c", Width: 1}}, GridConfig{RowHeight: 1, RowSelection: true}, GridViewport{Width: 1, Height: 1})
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	next := newSelectionView(1, 2, false, false)
	next.index.keyHook = func(int) (string, error) { return "new", nil }
	next.index.lookupHook = func(k string) (int, bool, error) { return 0, k == "new", nil }
	if err := g.ReplaceSource(next); err != nil {
		t.Fatal(err)
	}
	before := gridState(t, g)
	if err := g.SelectToKey("new", "c", GridSelectToggle); !errors.Is(err, ErrGridCapacity) || gridState(t, g) != before {
		t.Fatal("captured count overflow was not atomic", err, g.state())
	}
	small := newSelectionView(20, 1, false, false)
	g = selectionGrid(t, small, GridConfig{})
	if err := g.SelectToKey("r:1", "c:0", GridSelectReplace); err != nil {
		t.Fatal(err)
	}
	before = gridState(t, g)
	old := captured(t, g)
	ref := gridSnapshot(t, g).Cells[0].Ref
	small.captureHook = func() (GridRowIndex, error) { small.revision++; return small.index, nil }
	if err := g.SelectToKey("r:3", "c:0", GridSelectExtend); !errors.Is(err, ErrGridStale) || g.state() != before {
		t.Fatal("capture drift partially committed cursor/selection", err, g.state())
	}
	small.revision--
	small.captureHook = nil
	if yes, err := old.Contains("r:1"); err != nil || !yes {
		t.Fatal("failure mutated a retained selection", err)
	}
	if err := g.SelectCell(ref, GridSelectToggle); err != nil {
		t.Fatal("failure retired refs", err)
	}
}

type selectionBorrowedView struct {
	*borrowedGridSource
	prepared GridRowIndex
}

func (s *selectionBorrowedView) SelectionSnapshot() (GridRowIndex, error) { return s.prepared, nil }

func TestGridSelectionPayloadDetachmentAndBudget(t *testing.T) {
	backing := strings.Repeat("x", 1<<20)
	source := &borrowedGridSource{backing[10:19], 1, []string{backing[30:34], backing[50:56]}, ""}
	prepared := &borrowedGridSource{source.identity, source.revision, append([]string(nil), source.keys...), ""}
	g := selectionGrid(t, &selectionBorrowedView{source, prepared}, GridConfig{})
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	if err := g.SelectToKey(source.keys[1], "c:0", GridSelectToggle); err != nil {
		t.Fatal(err)
	}
	s := captured(t, g)
	assertGridDetached(t, "selection scope", s.state.Scope.Identity, backing)
	assertGridDetached(t, "selection anchor", s.state.AnchorKey, backing)
	for key := range s.overrides {
		assertGridDetached(t, "selection exception", key, backing)
	}
	small := &borrowedGridSource{"id", 1, []string{"aaaa", "bbbbbb"}, ""}
	view := &selectionBorrowedView{small, &borrowedGridSource{small.identity, 1, append([]string(nil), small.keys...), ""}}
	g = mustGrid(t, view, []GridColumn{{ID: "c", Width: 20}}, GridConfig{RowHeight: 20, RowSelection: true, MaxTextBytes: 7}, GridViewport{Width: 20, Height: 20})
	if err := g.SelectToKey("aaaa", "c", GridSelectReplace); err != nil {
		t.Fatal(err)
	}
	before := gridState(t, g)
	if err := g.SelectToKey("bbbbbb", "c", GridSelectToggle); !errors.Is(err, ErrGridCapacity) || gridState(t, g) != before {
		t.Fatal("selection payload budget was not atomic", err)
	}
}

func TestGridSelectionOptInAndEmptySource(t *testing.T) {
	s := newSelectionView(20, 1, false, false)
	g := mustGrid(t, s, gridColumns(1, 20), GridConfig{RowHeight: 20}, GridViewport{Width: 20, Height: 20})
	if err := g.JumpToKey("r:1", "c:0"); err != nil || s.captures != 0 {
		t.Fatal(err, s.captures)
	}
	if _, err := g.Selection(); !errors.Is(err, ErrGridSelection) {
		t.Fatal(err)
	}
	g = selectionGrid(t, newSelectionView(0, 1, false, false), GridConfig{})
	if err := g.SelectAll(); err != nil {
		t.Fatal(err)
	}
	if len(selectedKeys(t, captured(t, g))) != 0 {
		t.Fatal("empty selection captured rows")
	}
	before := gridState(t, g)
	if err := g.SelectAll(); err != nil || gridState(t, g) != before {
		t.Fatal("empty all not a no-op", err)
	}
	if err := g.ClearSelection(); err != nil || g.state().Selection.Scope != (GridStamp{}) {
		t.Fatal("empty captured scope not cleared", err)
	}
}
