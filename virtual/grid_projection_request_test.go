package virtual

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
)

func requestProjection(t *testing.T, source GridSource, r GridProjectionRequest, query GridQuery) GridProjection {
	t.Helper()
	p, err := BuildGridProjection(context.Background(), source, r.OutputRevision(), query, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestGridProjectionRequestLatestWinsWithoutRetiringView(t *testing.T) {
	s := gridSource(30)
	g := mustGrid(t, s, gridColumns(2, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 200, Height: 80})
	snap := gridSnapshot(t, g)
	a, err := g.PrepareProjection()
	if err != nil {
		t.Fatal(err)
	}
	b, err := g.PrepareProjection()
	if err != nil {
		t.Fatal(err)
	}
	if gridState(t, g) != snap.State || a.Stamp() != snap.State.Stamp || a.InputStamp() != snap.State.Stamp || a.OutputRevision() != 2 {
		t.Fatal("request changed the accepted view", a, b)
	}
	p := requestProjection(t, s, b, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0", Descending: true}}})
	if err := g.CompleteProjection(a, p); !errors.Is(err, ErrGridStale) {
		t.Fatal("superseded result accepted", err)
	}
	if gridState(t, g) != snap.State || g.pendingProjection != b.token {
		t.Fatal("refusal changed state")
	}
	// Prepare/cancel/refusal alone leave original references valid.
	if err := g.Activate(snap.Cells[0].Ref); err != nil {
		t.Fatal(err)
	}
	before := gridState(t, g)
	if err := g.CompleteProjection(b, p); err != nil {
		t.Fatal(err)
	}
	if g.generation != before.Generation+1 || g.pendingProjection != nil || g.source != p {
		t.Fatal("result not installed once")
	}
	if err := g.CompleteProjection(b, p); !errors.Is(err, ErrGridStale) {
		t.Fatal("replay accepted", err)
	}
	next, err := g.PrepareProjection()
	if err != nil || next.InputStamp() != b.InputStamp() || next.OutputRevision() != 3 || next.Stamp().Revision != 2 {
		t.Fatal("query lost the full domain stamp", next, err)
	}
}

// Embedding inherits the private interface marker; CompleteProjection must
// refuse this wrapper before invoking any of its caller-controlled methods.
type projectionWrapper struct{ GridProjection }

func (*projectionWrapper) InputStamp() GridStamp { panic("untrusted projection method") }
func (*projectionWrapper) Identity() string      { panic("untrusted projection method") }

func TestGridProjectionRequestExactResultAndAtomicRefusals(t *testing.T) {
	s := gridSource(30)
	g := mustGrid(t, s, gridColumns(1, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 100, Height: 80})
	r, _ := g.PrepareProjection()
	p := requestProjection(t, s, r, GridQuery{})
	before := gridState(t, g)
	for _, bad := range []GridProjection{nil, (*gridProjectionBase)(nil), (*gridProjectionMetrics)(nil), (*gridProjectionEdit)(nil), (*gridProjectionMetricsEdit)(nil), &projectionWrapper{p}} {
		if err := g.CompleteProjection(r, bad); !errors.Is(err, ErrGridSource) {
			t.Fatal(err)
		}
	}
	other := gridSource(30)
	other.revision = 2
	wrongInput := requestProjection(t, other, r, GridQuery{})
	wrongRevision, err := BuildGridProjection(context.Background(), s, 3, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []GridProjection{wrongInput, wrongRevision} {
		if err := g.CompleteProjection(r, bad); !errors.Is(err, ErrGridStale) {
			t.Fatal(err)
		}
	}
	if gridState(t, g) != before || g.pendingProjection != r.token || g.source != s {
		t.Fatal("bad result changed state")
	}
	otherGrid := mustGrid(t, s, gridColumns(1, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 100, Height: 80})
	foreign, _ := otherGrid.PrepareProjection()
	if err := g.CompleteProjection(foreign, p); !errors.Is(err, ErrGridStale) || g.pendingProjection != r.token {
		t.Fatal("foreign admission changed the request", err)
	}
	// Generation exhaustion preserves the request and accepted view.
	g.generation = math.MaxUint64
	if err := g.CompleteProjection(r, p); !errors.Is(err, ErrGridExhausted) || g.pendingProjection != r.token || g.source != s {
		t.Fatal("exhaustion changed the admission", err)
	}
	g.generation = before.Generation
	if err := g.CompleteProjection(r, p); err != nil {
		t.Fatal(err)
	}
}

func TestGridProjectionRequestCancellationAndSourceReplacement(t *testing.T) {
	s := gridSource(30)
	g := mustGrid(t, s, gridColumns(1, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 100, Height: 80})
	r, _ := g.PrepareProjection()
	bad := gridSource(30)
	bad.id = ""
	before := gridState(t, g)
	if err := g.ReplaceSource(bad); !errors.Is(err, ErrGridSource) || g.pendingProjection != r.token || gridState(t, g) != before {
		t.Fatal("failed replacement canceled request", err)
	}
	if err := g.Refresh(); err != nil || g.pendingProjection != r.token {
		t.Fatal("healthy no-op refresh retired request", err)
	}
	if err := g.ReplaceSource(s); err != nil || g.pendingProjection != nil {
		t.Fatal("same-stamp source install failed to retire request", err)
	}
	if err := g.CancelProjection(r); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
	r, _ = g.PrepareProjection()
	s.revision++ // recovery cancellation must not read an unacknowledged source
	if err := g.CancelProjection(r); err != nil || g.pendingProjection != nil || g.generation != before.Generation+1 {
		t.Fatal(err)
	}
	if _, err := g.PrepareProjection(); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	r, _ = g.PrepareProjection()
	s.revision = math.MaxUint64
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	// A prepare refusal never clears an existing token. An artificial pending
	// token here isolates the revision-overflow boundary from source retirement.
	g.pendingProjection = r.token
	if _, err := g.PrepareProjection(); !errors.Is(err, ErrGridExhausted) || g.pendingProjection != r.token {
		t.Fatal(err)
	}
}

func TestGridProjectionCompletionPreservesCurrentSelectionAndDraft(t *testing.T) {
	s := newSelectionView(30, 1, false, false)
	g := selectionGrid(t, s, GridConfig{})
	r, _ := g.PrepareProjection()
	p := requestProjection(t, s, r, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0", Descending: true}}})
	if err := g.SelectToKey("r:3", "c:0", GridSelectReplace); err != nil {
		t.Fatal(err)
	}
	if err := g.SelectToKey("r:7", "c:0", GridSelectExtend); err != nil {
		t.Fatal(err)
	}
	keys, active := selectedKeys(t, captured(t, g)), g.active
	if err := g.Scroll(120, 20); err != nil {
		t.Fatal(err)
	}
	if err := g.CompleteProjection(r, p); err != nil || g.active != active || !reflect.DeepEqual(selectedKeys(t, captured(t, g)), keys) {
		t.Fatal("selection/cursor changed", err)
	}
	e, domain := editableFixtureGrid(t, 30)
	draft := beginFixtureEdit(t, e)
	r, _ = e.PrepareProjection()
	p = requestProjection(t, domain, r, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0", Descending: true}}})
	if err := e.SetEditText("draft arrived during worker"); err != nil {
		t.Fatal(err)
	}
	if err := e.CompleteProjection(r, p); err != nil {
		t.Fatal(err)
	}
	edit, err := e.Edit()
	if err != nil || edit.RowKey != draft.RowKey || edit.Text != "draft arrived during worker" || edit.Status != GridEditReady {
		t.Fatal(edit, err)
	}
	r, _ = e.PrepareProjection()
	er, err := e.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	next := domain.candidate(edit.RowKey, edit.ColumnID, edit.Text, 2)
	if err := e.CompleteEdit(er, next); !errors.Is(err, ErrGridStale) {
		// The accepted view is already revision2; domain revision2 alone is not
		// a newer view. Explicitly build the edited full snapshot at view3.
		t.Fatal("domain revision confused with view revision", err)
	}
	edited, err := BuildGridProjection(context.Background(), next, er.Stamp().Revision+1, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.CompleteEdit(er, edited); err != nil || e.pendingProjection != nil {
		t.Fatal("accepted edit did not retire pending query", err)
	}
	if err := e.CancelProjection(r); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
}

func TestGridProjectionEditCanFilterCommittedRowOut(t *testing.T) {
	g, domain := editableFixtureGrid(t, 10)
	draft := beginFixtureEdit(t, g)
	if err := g.SetEditText("changed filter membership"); err != nil {
		t.Fatal(err)
	}
	r, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	next := domain.candidate(draft.RowKey, draft.ColumnID, r.State().Text, 2)
	unchangedDomain := domain.candidate(draft.RowKey, draft.ColumnID, r.State().Text, 2)
	unchangedDomain.revision = domain.revision
	unrevisioned, err := BuildGridProjection(context.Background(), unchangedDomain, 2, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	unchanged := gridState(t, g)
	if err := g.CompleteEdit(r, unrevisioned); !errors.Is(err, ErrGridStale) || gridState(t, g) != unchanged || g.source != domain {
		t.Fatal("view revision hid an unchanged domain revision", err)
	}
	p, err := BuildGridProjection(context.Background(), next, 2, GridQuery{Filters: []GridFilter{{ColumnID: draft.ColumnID, Kind: GridFilterEquals, Text: draft.Original}}}, GridProjectionLimits{})
	if err != nil || p.RowCount() != 0 {
		t.Fatal("candidate did not exclude edited row", err)
	}
	before := gridState(t, g)
	if err := g.CompleteEdit(r, p); err != nil {
		t.Fatal("filtered edit refused", err)
	}
	if after := gridState(t, g); after.Edit.Active || after.Stamp.Rows != 0 || after.Generation != before.Generation+1 {
		t.Fatal("edit/source publication not atomic", after)
	}
	if v, _, err := g.checkedEditCell(next, p.InputStamp(), draft.RowKey, draft.ColumnID); err != nil || v.Text != r.State().Text || v.Version != 2 {
		t.Fatal(v, err)
	}
	if err := g.CompleteEdit(r, p); !errors.Is(err, ErrGridStale) {
		t.Fatal("edit replay accepted", err)
	}
}
