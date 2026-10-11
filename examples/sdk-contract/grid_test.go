package sdkcontract_test

import (
	"errors"
	"testing"

	"github.com/dshills/vellum/virtual"
)

type consumerGridSource struct {
	keys     []string
	indices  map[string]int
	revision uint64
}

func (s *consumerGridSource) Identity() string             { return "consumer-grid" }
func (s *consumerGridSource) Revision() uint64             { return s.revision }
func (s *consumerGridSource) RowCount() int                { return len(s.keys) }
func (s *consumerGridSource) RowKey(i int) (string, error) { return s.keys[i], nil }
func (s *consumerGridSource) IndexOf(key string) (int, bool, error) {
	i, found := s.indices[key]
	return i, found, nil
}
func (s *consumerGridSource) Cell(key, column string) (string, error) {
	return key + ":" + column, nil
}
func (s *consumerGridSource) install(keys ...string) {
	s.keys = keys
	s.indices = make(map[string]int, len(keys))
	for i, key := range keys {
		s.indices[key] = i
	}
	s.revision++
}

// This uses only exported SDK types. A retained native cell action must preserve
// its domain identity through source reorder and refuse its retired press ref.
func TestGridConsumerRetainsIdentityAndRefusesRetiredPress(t *testing.T) {
	source := &consumerGridSource{}
	source.install("a", "b", "c", "d", "e", "f")
	grid, err := virtual.NewGrid(source,
		[]virtual.GridColumn{{ID: "name", Title: "Name", Width: 100}, {ID: "amount", Title: "Amount", Width: 100}, {ID: "notes", Title: "Notes", Width: 100}},
		virtual.GridConfig{RowHeight: 24, HeaderHeight: 24, FrozenColumns: 1, OverscanRows: 1, OverscanColumns: 1, MaxRows: 4, MaxColumns: 3, MaxCells: 12},
		virtual.GridViewport{Width: 200, Height: 72})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := grid.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Cells) != 9 || !snapshot.Headers[0].Frozen || snapshot.Headers[0].Rect.X != 0 {
		t.Fatalf("bounded frozen grid: %+v", snapshot)
	}
	ref := snapshot.Cells[4].Ref // row b, column amount
	if err := grid.Activate(ref); err != nil {
		t.Fatal(err)
	}
	source.install("c", "d", "e", "f", "a", "b")
	if _, err := grid.Snapshot(); !errors.Is(err, virtual.ErrGridStale) {
		t.Fatal(err)
	}
	if err := grid.Refresh(); err != nil {
		t.Fatal(err)
	}
	state, err := grid.State()
	if err != nil || state.Active != (virtual.GridActive{RowKey: "b", ColumnID: "amount"}) || state.Viewport.ScrollY != 96 {
		t.Fatalf("identity and top-row anchor after reorder: %+v %v", state, err)
	}
	if err := grid.Activate(ref); !errors.Is(err, virtual.ErrGridStale) {
		t.Fatal(err)
	}
	if err := grid.SetViewport(virtual.GridViewport{Width: 200, Height: 1000}); !errors.Is(err, virtual.ErrGridCapacity) {
		t.Fatal(err)
	}
	after, err := grid.State()
	if err != nil || after != state {
		t.Fatalf("failed resize changed retained state: %+v %v", after, err)
	}
}
