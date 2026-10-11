package virtual

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func TestReviewEditPreparedSourceCASAndViewportIndependence(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	if err := g.SetEditText("local"); err != nil {
		t.Fatal(err)
	}
	r, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Scroll(20, 20); err != nil {
		t.Fatal(err)
	}
	candidate := s.candidate(e.RowKey, e.ColumnID, "local", 2)
	if err := g.CompleteEdit(r, candidate); err != nil {
		t.Fatal("viewport-only transition retired proposal", err)
	}
	g, s = editableFixtureGrid(t, 100)
	e = beginFixtureEdit(t, g)
	if err := g.SetEditText("local"); err != nil {
		t.Fatal(err)
	}
	r, err = g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	updated := s.candidate("r:50", "c:3", "unrelated change", 2)
	if err := g.ReplaceSource(updated); err != nil {
		t.Fatal(err)
	}
	before := g.state()
	if err := g.CompleteEdit(r, s.candidate(e.RowKey, e.ColumnID, "local", 2)); !errors.Is(err, ErrGridStale) {
		t.Fatal("old prepared source overwrote newer domain", err)
	}
	if g.state() != before || g.source != updated {
		t.Fatal("stale candidate changed state")
	}
	fresh, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.CompleteEdit(fresh, updated.candidate(e.RowKey, e.ColumnID, "local", 2)); err != nil {
		t.Fatal(err)
	}
	live := g.source.(*editableGridFixture)
	if v, err := live.Cell("r:50", "c:3"); err != nil || v != "unrelated change" {
		t.Fatal("new proposal lost unrelated update", v, err)
	}
}

func TestReviewEditCandidateSourceCallbacksCannotPublishLiveDrift(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	if err := g.SetEditText("local"); err != nil {
		t.Fatal(err)
	}
	r, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	before := g.state()
	candidate := s.candidate(e.RowKey, e.ColumnID, "local", 2)
	first := true
	candidate.rowHook = func(i int) (string, error) {
		if first {
			s.revision++
			first = false
		}
		return "r:" + strings.TrimPrefix(candidate.indexedGridSourceKey(i), "r:"), nil
	}
	if err := g.CompleteEdit(r, candidate); !errors.Is(err, ErrGridStale) {
		t.Fatal("candidate reads changed live source but completion succeeded", err)
	}
	if g.state() != before || g.source != s {
		t.Fatal("failed completion published candidate or destroyed draft")
	}
	state, err := g.Edit()
	if err != nil || state.Text != "local" {
		t.Fatal("lost recoverable draft", state, err)
	}
	if err := g.CancelEdit(); err != nil {
		t.Fatal(err)
	}
}

func (s *editableGridFixture) indexedGridSourceKey(i int) string {
	if s.reverse {
		i = s.rows - 1 - i
	}
	// The isolated source is deliberately a 100-row fixture.
	const digits = "0123456789"
	if i < 10 {
		return "r:" + string(digits[i])
	}
	return "r:" + string(digits[i/10]) + string(digits[i%10])
}

func TestReviewEditVersionHistorySurvivesMissingAndIncarnation(t *testing.T) {
	for _, mode := range []string{"missing", "incarnation"} {
		t.Run(mode, func(t *testing.T) {
			g, s := editableFixtureGrid(t, 100)
			e := beginFixtureEdit(t, g)
			conflict := s.candidate(e.RowKey, e.ColumnID, "remote", 3)
			if err := g.ReplaceSource(conflict); err != nil {
				t.Fatal(err)
			}
			away := conflict.candidate("r:99", "c:3", "unrelated", 2)
			if mode == "missing" {
				away.rows = 0
			} else {
				away.id = "other-domain"
			}
			if err := g.ReplaceSource(away); err != nil {
				t.Fatal(err)
			}
			before := g.state()
			back := conflict.candidate(e.RowKey, e.ColumnID, "remote", 2)
			back.revision = away.revision + 1
			if err := g.ReplaceSource(back); !errors.Is(err, ErrGridSource) {
				t.Fatal("forgot cell version across unavailable/incarnation", err)
			}
			if g.state() != before || g.source != away {
				t.Fatal("invalid restoration admitted")
			}
		})
	}
}

func TestReviewEditRetainedPayloadDetachmentAndExhaustion(t *testing.T) {
	s := newEditableFixture(100)
	backing := strings.Repeat("x", 1<<20)
	baseline := backing[123:132]
	s.values[GridActive{"r:0", "c:0"}] = GridEditValue{baseline, 1, true}
	g := mustGrid(t, s, gridColumns(4, 100), GridConfig{RowHeight: 24, HeaderHeight: 28, FrozenColumns: 1, CellEditing: true}, GridViewport{Width: 300, Height: 124})
	e := beginFixtureEdit(t, g)
	draft := backing[2048:2053]
	if err := g.SetEditText(draft); err != nil {
		t.Fatal(err)
	}
	r, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	start := uintptr(unsafe.Pointer(unsafe.StringData(backing)))
	end := start + uintptr(len(backing))
	for name, text := range map[string]string{"original": e.Original, "current": g.edit.state.CurrentText, "draft": g.edit.state.Text, "requestOriginal": r.State().Original, "requestText": r.State().Text, "lastValue": g.edit.lastValue.Text} {
		p := uintptr(unsafe.Pointer(unsafe.StringData(text)))
		if len(text) > 0 && p >= start && p < end {
			t.Fatal("retained caller backing", name)
		}
	}
	g.generation = math.MaxUint64
	before := g.state()
	sourceBefore := g.source
	if err := g.CompleteEdit(r, s.candidate(e.RowKey, e.ColumnID, draft, 2)); !errors.Is(err, ErrGridExhausted) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g.state(), before) || g.source != sourceBefore {
		t.Fatal("exhausted completion changed state")
	}
}
