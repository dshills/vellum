package virtual

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type million struct{ calls *int }

func (m million) Len() int { return 1000000 }
func (m million) Row(i int) Row {
	*m.calls++
	return Row{ID: fmt.Sprint(i), Cells: []string{fmt.Sprint(i), "10k"}}
}
func TestMillionRowsMaterializesOnlyViewport(t *testing.T) {
	calls := 0
	table, err := NewTable(million{&calls}, 40)
	if err != nil {
		t.Fatal(err)
	}
	table.SetViewport(999960, 40)
	rows := table.View()
	if len(rows) != 40 || calls != 40 || rows[39].ID != "999999" {
		t.Fatal(len(rows), calls)
	}
	table.SelectVisible(0)
	if table.Selected != "999960" {
		t.Fatal(table.Selected)
	}
	if table.SetViewport(0, 257) == nil {
		t.Fatal("unbounded viewport")
	}
}
func TestTreeHundredThousandNodesAndCycles(t *testing.T) {
	nodes := make([]TreeNode, 100000)
	nodes[0] = TreeNode{ID: "root", Label: "root"}
	for i := 1; i < len(nodes); i++ {
		nodes[i] = TreeNode{ID: fmt.Sprint(i), Parent: "root"}
	}
	tree, err := NewTree(nodes)
	if err != nil {
		t.Fatal(err)
	}
	tree.Toggle("root")
	rows, err := tree.View(99960, 40)
	if err != nil || len(rows) != 40 || rows[39].ID != "99999" {
		t.Fatal(len(rows), err)
	}
	if _, err := NewTree([]TreeNode{{ID: "a", Parent: "b"}, {ID: "b", Parent: "a"}}); err == nil {
		t.Fatal("cycle")
	}
	if err := tree.Select("500"); err != nil {
		t.Fatal(err)
	}
	tree.Toggle("root")
	if tree.Selected != "500" {
		t.Fatal("selection lost when hidden")
	}
}
func BenchmarkMillionViewport(b *testing.B) {
	calls := 0
	t, _ := NewTable(million{&calls}, 40)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		t.SetViewport(i%999960, 40)
		t.View()
	}
}

type tiny []Row

func (s tiny) Len() int      { return len(s) }
func (s tiny) Row(i int) Row { return s[i] }
func TestProjectionKeepsSelectionAndRejectsMalformedIndices(t *testing.T) {
	source := tiny{{ID: "a", Cells: []string{"z"}}, {ID: "b", Cells: []string{"a"}}}
	table, err := NewTable(source, 2)
	if err != nil {
		t.Fatal(err)
	}
	table.Selected = "a"
	indices, err := Project(source, "", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := table.ApplyProjection(indices, table.ProjectionToken()); err != nil {
		t.Fatal(err)
	}
	if table.View()[0].ID != "b" || table.Selected != "a" {
		t.Fatal("selection/order")
	}
	if table.ApplyProjection(Projection{Indices: []int{0, 0}, length: 2}, table.ProjectionToken()) == nil {
		t.Fatal("duplicate index")
	}
	if table.View()[0].ID != "b" {
		t.Fatal("failed projection mutated order")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ProjectContext(ctx, source, "", 0, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := Project(source, "", 1, false); err == nil {
		t.Fatal("missing column")
	}
}

type mutableSource struct{ rows tiny }

func (s *mutableSource) Len() int      { return len(s.rows) }
func (s *mutableSource) Row(i int) Row { return s.rows[i] }
func TestSourceShrinkClearsStaleProjection(t *testing.T) {
	source := &mutableSource{rows: tiny{{ID: "one", Cells: []string{"z"}}, {ID: "two", Cells: []string{"a"}}}}
	table, err := NewTable(source, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := table.ApplyProjection(Projection{Indices: []int{1, 0}, length: 2}, table.ProjectionToken()); err != nil {
		t.Fatal(err)
	}
	source.rows = source.rows[:1]
	rows := table.View()
	if len(rows) != 1 || rows[0].ID != "one" {
		t.Fatal(rows)
	}
}

func TestSourceGrowthPreservesViewportBound(t *testing.T) {
	source := &mutableSource{rows: make(tiny, 100)}
	for i := range source.rows {
		source.rows[i] = Row{ID: fmt.Sprint(i)}
	}
	table, err := NewTable(source, 10)
	if err != nil {
		t.Fatal(err)
	}
	table.SetViewport(50, 10)
	source.rows = append(source.rows, make(tiny, 100)...)
	if rows := table.View(); len(rows) != 10 || table.StartRow() != 50 {
		t.Fatal("grew beyond viewport", len(rows))
	}
}

type revisionSource struct {
	rows     tiny
	revision uint64
}

func (s *revisionSource) Len() int         { return len(s.rows) }
func (s *revisionSource) Row(i int) Row    { return s.rows[i] }
func (s *revisionSource) Revision() uint64 { return s.revision }
func TestSameLengthRevisionInvalidatesOrder(t *testing.T) {
	source := &revisionSource{rows: tiny{{ID: "a"}, {ID: "b"}}, revision: 1}
	table, err := NewTable(source, 1)
	if err != nil {
		t.Fatal(err)
	}
	table.ApplyProjection(Projection{Indices: []int{1, 0}, length: 2, versioned: true, revision: 1}, table.ProjectionToken())
	table.SetViewport(1, 1)
	source.revision++
	if rows := table.View(); len(rows) != 1 || rows[0].ID != "b" || table.StartRow() != 1 || !table.ProjectionInvalidated() {
		t.Fatal("stale sort or viewport jump", rows, table.StartRow())
	}
}

func TestStaleWorkerProjectionRejected(t *testing.T) {
	s := &revisionSource{rows: tiny{{ID: "a", Cells: []string{"z"}}, {ID: "b", Cells: []string{"a"}}}, revision: 1}
	table, _ := NewTable(s, 2)
	token := table.ProjectionToken()
	result, err := Project(s, "", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	s.revision++
	if table.ApplyProjection(result, token) == nil {
		t.Fatal("accepted stale worker order")
	}
	if table.View()[0].ID != "a" {
		t.Fatal("stale order installed")
	}
}

func TestReplacedSourceRejectsProjectionToken(t *testing.T) {
	s := tiny{{ID: "a", Cells: []string{"z"}}, {ID: "b", Cells: []string{"a"}}}
	table, _ := NewTable(s, 2)
	token := table.ProjectionToken()
	result, _ := Project(s, "", 0, false)
	table.ReplaceSource(tiny{{ID: "c", Cells: []string{"a"}}, {ID: "d", Cells: []string{"z"}}})
	if table.ApplyProjection(result, token) == nil {
		t.Fatal("accepted prior-source order")
	}
}

type negativeSource struct{}

func (negativeSource) Len() int    { return -1 }
func (negativeSource) Row(int) Row { panic("must not materialize") }
func TestProjectRejectsNegativeSource(t *testing.T) {
	if _, err := Project(negativeSource{}, "", 0, false); err == nil {
		t.Fatal("negative source")
	}
}
