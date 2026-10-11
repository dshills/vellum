package virtual

import (
	"errors"
	"fmt"
	"testing"
)

type reviewRowsSource struct {
	prefix   []int
	revision uint64
}

func (s *reviewRowsSource) Identity() string { return "review" }
func (s *reviewRowsSource) Revision() uint64 { return s.revision }
func (s *reviewRowsSource) RowCount() int    { return len(s.prefix) - 1 }
func (s *reviewRowsSource) RowKey(i int) (string, error) {
	if i < 0 || i >= s.RowCount() {
		return "", ErrGridRange
	}
	return fmt.Sprint(i), nil
}
func (s *reviewRowsSource) IndexOf(k string) (int, bool, error) {
	for i := 0; i < s.RowCount(); i++ {
		if fmt.Sprint(i) == k {
			return i, true, nil
		}
	}
	return 0, false, nil
}
func (s *reviewRowsSource) Cell(k, c string) (string, error) { return k, nil }
func (s *reviewRowsSource) RowOffset(i int) (int, error) {
	if i < 0 || i >= len(s.prefix) {
		return 0, ErrGridRange
	}
	return s.prefix[i], nil
}
func (s *reviewRowsSource) RowAt(p int) (int, error) {
	for i := 0; i < s.RowCount(); i++ {
		if p >= s.prefix[i] && p < s.prefix[i+1] {
			return i, nil
		}
	}
	return 0, ErrGridRange
}

func TestReviewVariableRowsDirectBoundsRequireInverse(t *testing.T) {
	s := &reviewRowsSource{prefix: []int{0, 100, 50, 300}, revision: 1}
	g, err := NewGrid(s, []GridColumn{{ID: "c", Width: 20}}, GridConfig{RowHeight: 1}, GridViewport{Width: 20, Height: 5})
	if err != nil {
		t.Fatal(err)
	}
	before := g.state()
	err = g.JumpToKey("2", "c")
	if !errors.Is(err, ErrGridSource) {
		t.Fatalf("indexed row2 has RowOffset=50 but RowAt(50)=row0; Jump must refuse contradictory accessed inverse, got err=%v state=%+v", err, g.state())
	}
	if g.state() != before {
		t.Fatalf("refused inverse changes state")
	}
}

func TestReviewVariableRowsPageCannotReverseDirection(t *testing.T) {
	s := &reviewRowsSource{prefix: []int{0, 100, 50, 300}, revision: 1}
	g, err := NewGrid(s, []GridColumn{{ID: "c", Width: 20}}, GridConfig{RowHeight: 1}, GridViewport{Width: 20, Height: 5})
	if err != nil {
		t.Fatal(err)
	}
	// Establish the row2 origin without accepting any malformed mutation. This
	// isolates Page's comparison of the two accessed, individually valid spans.
	g.active = GridActive{RowKey: "2", ColumnID: "c"}
	before := g.state()
	err = g.Page(1)
	if !errors.Is(err, ErrGridSource) {
		t.Fatalf("Page(+1) from row2 maps pixel55 to row0 and moves backward; expected source refusal, got %v active=%+v", err, g.active)
	}
	if g.state() != before {
		t.Fatalf("refused Page changes state")
	}
}

func TestReviewVariableRowsProspectiveInteriorGeometry(t *testing.T) {
	s := &reviewRowsSource{prefix: []int{0, 10, 20, 30, 40, 50}, revision: 1}
	g, err := NewGrid(s, []GridColumn{{ID: "c", Width: 20}}, GridConfig{RowHeight: 1}, GridViewport{Width: 20, Height: 5})
	if err != nil {
		t.Fatal(err)
	}
	// A malformed prospective source has valid endpoints and valid first/last
	// spans, but an invalid interior span in the newly visible bounded range.
	bad := &reviewRowsSource{prefix: []int{0, 10, 20, 15, 40, 50}, revision: 2}
	before := g.state()
	err = g.ReplaceSource(bad)
	if err != nil {
		t.Fatal("initial small viewport should not inspect hidden row2:", err)
	}
	stable := g.state()
	v := stable.Viewport
	v.Height = 45
	err = g.SetViewport(v)
	if !errors.Is(err, ErrGridSource) {
		t.Fatalf("newly materialized row2 has end15<=start20; geometry transition should refuse atomically, got err=%v state=%+v", err, g.state())
	}
	if g.state() != stable {
		t.Fatal("invalid interior geometry committed viewport")
	}
	_ = before
}
