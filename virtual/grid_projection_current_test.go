package virtual

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// These providers deliberately break the borrowed immutable-input contract
// after construction. The built view's own metadata still stays unchanged.
type projectionCurrentSource struct {
	*projectionTestSource
	identity string
}

func (s *projectionCurrentSource) Identity() string { return s.identity }

type projectionCurrentMetrics struct {
	*projectionCurrentSource
	heights *RowHeights
}

func (s *projectionCurrentMetrics) RowOffset(i int) (int, error) {
	return s.heights.RowOffset(i)
}
func (s *projectionCurrentMetrics) RowAt(i int) (int, error) {
	return s.heights.RowAt(i)
}

type projectionCurrentEdit struct{ *projectionCurrentSource }

func (s *projectionCurrentEdit) EditCell(key, column string) (GridEditValue, error) {
	text, err := s.Cell(key, column)
	return GridEditValue{Text: text, Version: 1, Editable: true}, err
}

type projectionCurrentMetricsEdit struct{ *projectionCurrentMetrics }

func (s *projectionCurrentMetricsEdit) EditCell(key, column string) (GridEditValue, error) {
	text, err := s.Cell(key, column)
	return GridEditValue{Text: text, Version: 1, Editable: true}, err
}

func projectionCurrentInput(t *testing.T, rows, variant int) (*projectionCurrentSource, GridSource) {
	t.Helper()
	base := &projectionCurrentSource{projectionTestSource: projectionSource(rows), identity: "current-domain"}
	heights := make([]int, rows)
	for i := range heights {
		heights[i] = 7 + i*2
	}
	index, err := NewRowHeights(heights)
	if err != nil {
		t.Fatal(err)
	}
	metrics := &projectionCurrentMetrics{projectionCurrentSource: base, heights: index}
	switch variant {
	case 0:
		return base, base
	case 1:
		return base, metrics
	case 2:
		return base, &projectionCurrentEdit{base}
	case 3:
		return base, &projectionCurrentMetricsEdit{metrics}
	default:
		t.Fatal("unknown capability variant")
		return nil, nil
	}
}

func projectionCurrentGrid(t *testing.T, source GridSource, query GridQuery) (*Grid, GridProjection) {
	t.Helper()
	view, err := BuildGridProjection(context.Background(), source, 91, query, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	_, editable := view.(GridEditSource)
	g, err := NewGrid(view, []GridColumn{{ID: "c:0", Width: 60}}, GridConfig{RowHeight: 9, HeaderHeight: 3, RowSelection: true, CellEditing: editable}, GridViewport{Width: 80, Height: 30})
	if err != nil {
		t.Fatal(err)
	}
	return g, view
}

func TestGridProjectionCurrentRejectsBorrowedInputDrift(t *testing.T) {
	for variant, name := range []string{"fixed", "metrics", "edit", "metrics-edit"} {
		for _, shape := range []string{"visible", "empty-input", "empty-filter"} {
			for _, drift := range []string{"identity", "revision", "rows"} {
				t.Run(fmt.Sprintf("%s/%s/%s", name, shape, drift), func(t *testing.T) {
					rows := 3
					if shape == "empty-input" {
						rows = 0
					}
					base, source := projectionCurrentInput(t, rows, variant)
					query := GridQuery{}
					if shape == "empty-filter" {
						query.Filters = []GridFilter{{ColumnID: "c:0", Kind: GridFilterEquals, Text: "not present"}}
					}
					g, view := projectionCurrentGrid(t, source, query)
					before, err := g.State()
					if err != nil {
						t.Fatal(err)
					}
					if shape != "visible" && before.Stamp.Rows != 0 {
						t.Fatal("fixture did not create an empty view", before)
					}
					request, err := g.PrepareProjection()
					if err != nil {
						t.Fatal(err)
					}
					oldIdentity, oldRevision, oldRows := base.identity, base.revision, base.rows
					switch drift {
					case "identity":
						base.identity = "different-domain"
					case "revision":
						base.revision++
					case "rows":
						base.rows = append(append([][]string(nil), base.rows...), []string{"new", "0", "keep"})
					}
					if got, err := sourceStamp(view); err != nil || got != before.Stamp {
						t.Fatal("test must retain the projection's accepted view stamp", got, before.Stamp, err)
					}
					operations := []struct {
						name string
						run  func() error
					}{
						{"State", func() error { _, err := g.State(); return err }},
						{"Snapshot", func() error { _, err := g.Snapshot(); return err }},
						{"Refresh", g.Refresh},
						{"SetViewportNoop", func() error { return g.SetViewport(before.Viewport) }},
						{"ScrollNoop", func() error { return g.Scroll(0, 0) }},
						{"Move", func() error { return g.Move(1, 0) }},
						{"Page", func() error { return g.Page(1) }},
						{"Activate", func() error { return g.Activate(GridCellRef{}) }},
						{"Selection", func() error { _, err := g.Selection(); return err }},
						{"SelectAll", g.SelectAll},
						{"BeginActiveEdit", g.BeginActiveEdit},
						{"PrepareEdit", func() error { _, err := g.PrepareEdit(); return err }},
						{"PrepareProjection", func() error { _, err := g.PrepareProjection(); return err }},
					}
					for _, op := range operations {
						t.Run(op.name, func(t *testing.T) {
							err := op.run()
							if !errors.Is(err, ErrGridSource) || !errors.Is(err, ErrGridStale) {
								t.Errorf("borrowed input drift admitted or error cause lost: %v", err)
							}
							if g.state() != before || g.pendingProjection != request.token {
								t.Fatal("refusal changed accepted view or pending request")
							}
						})
					}
					base.identity, base.revision, base.rows = oldIdentity, oldRevision, oldRows
					if got, err := g.State(); err != nil || got != before {
						t.Fatal("refusal poisoned the original accepted source", got, err)
					}
					if err := g.CancelProjection(request); err != nil {
						t.Fatal("refusal retired the original request", err)
					}
				})
			}
		}
	}
}

func TestGridProjectionCurrentMetadataOnlyAndCallbackDrift(t *testing.T) {
	base, source := projectionCurrentInput(t, 3, 0)
	g, _ := projectionCurrentGrid(t, source, GridQuery{})
	before, err := g.State()
	if err != nil {
		t.Fatal(err)
	}
	indexed := 0
	base.hook = func(string, int) error {
		indexed++
		return errors.New("current-state inspection must not materialize indexed data")
	}
	for i := 0; i < 5; i++ {
		if got, err := g.State(); err != nil || got != before {
			t.Fatal("valid state inspection changed state or read indexed data", got, err)
		}
	}
	if indexed != 0 {
		t.Fatal("state inspection enumerated the source", indexed)
	}
	// Revision returns its old value, then the metadata callback changes it.
	// A second input-stamp observation must catch that change before publication.
	once := false
	base.metadata = func(operation string) {
		if operation == "revision" && !once {
			once = true
			base.revision++
		}
	}
	if _, err := g.State(); !errors.Is(err, ErrGridSource) || !errors.Is(err, ErrGridStale) {
		t.Fatal("metadata callback drift escaped state publication", err)
	}
	if !once || indexed != 0 || g.state() != before {
		t.Fatal("metadata drift fixture did not remain bounded and atomic", once, indexed)
	}
}
