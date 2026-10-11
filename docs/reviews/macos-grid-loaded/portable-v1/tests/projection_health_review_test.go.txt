package virtual

import (
 "context"
 "errors"
 "testing"
)

func reviewEmptyProjection(t *testing.T, input GridSource, revision uint64) GridProjection {
 t.Helper()
 p, err := BuildGridProjection(context.Background(), input, revision, GridQuery{}, GridProjectionLimits{})
 if err != nil { t.Fatal(err) }
 if p.RowCount() != 0 { t.Fatal("fixture is not empty") }
 return p
}

func TestReviewProjectionCompleteRejectsEmptyInputDrift(t *testing.T) {
 live, input := gridSource(0), gridSource(0)
 g := mustGrid(t, live, gridColumns(1, 100), GridConfig{RowHeight:20}, GridViewport{Width:100, Height:80})
 r, err := g.PrepareProjection(); if err != nil { t.Fatal(err) }
 p := reviewEmptyProjection(t, input, r.OutputRevision())
 before := g.state()
 input.revision++
 if err := g.CompleteProjection(r, p); !errors.Is(err, ErrGridSource) { t.Fatalf("empty drifted input accepted: %v", err) }
 if g.state() != before || g.source != live || g.pendingProjection != r.token { t.Fatal("refusal mutated accepted view or request") }
}

func TestReviewProjectionPrepareRejectsEmptyInputDrift(t *testing.T) {
 input := gridSource(0)
 p := reviewEmptyProjection(t, input, 2)
 g := mustGrid(t, p, gridColumns(1, 100), GridConfig{RowHeight:20}, GridViewport{Width:100, Height:80})
 prior, err := g.PrepareProjection(); if err != nil { t.Fatal(err) }
 before := g.state()
 input.revision++
 if _, err := g.PrepareProjection(); !errors.Is(err, ErrGridSource) { t.Fatalf("prepared query from unhealthy empty full input: %v", err) }
 if g.state() != before || g.pendingProjection != prior.token { t.Fatal("failed prepare superseded pending request") }
}

type reviewDriftingEmptySource struct {
 *indexedGridSource
 input *indexedGridSource
 trigger, calls, fired int
}
func (s *reviewDriftingEmptySource) RowCount() int {
 if s.trigger > 0 {
  s.calls++
  if s.calls == s.trigger { s.input.revision++; s.fired++ }
 }
 return s.rows
}

func reviewLateEmptyDrift(t *testing.T, trigger int) {
 t.Helper()
 input := gridSource(0)
 live := &reviewDriftingEmptySource{indexedGridSource:gridSource(0), input:input}
 g := mustGrid(t, live, gridColumns(1, 100), GridConfig{RowHeight:20}, GridViewport{Width:100, Height:80})
 r, err := g.PrepareProjection(); if err != nil { t.Fatal(err) }
 p := reviewEmptyProjection(t, input, r.OutputRevision())
 before := g.state()
 live.trigger = trigger
 err = g.CompleteProjection(r, p)
 if live.fired != 1 { t.Fatalf("did not reach intended live metadata boundary: calls=%d fired=%d", live.calls, live.fired) }
 if !errors.Is(err, ErrGridSource) { t.Fatalf("late empty input drift published: %v (calls=%d)", err, live.calls) }
 if g.state() != before || g.source != live || g.pendingProjection != r.token { t.Fatal("late refusal mutated accepted view or request") }
}

func TestReviewProjectionFinalRequestRejectsEmptyInputDrift(t *testing.T) {
 reviewLateEmptyDrift(t, 3)
}

func TestReviewProjectionFinalLiveStampRejectsEmptyInputDrift(t *testing.T) {
 reviewLateEmptyDrift(t, 5)
}
