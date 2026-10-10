package cogent

import (
	"cogentcore.org/core/core"
	"example.com/vellum/app"
	"example.com/vellum/virtual"
	"fmt"
	"testing"
)

type largeSource struct{}

func (largeSource) Len() int { return 1000000 }
func (largeSource) Row(i int) virtual.Row {
	return virtual.Row{ID: fmt.Sprint(i), Cells: []string{fmt.Sprint(i)}}
}
func TestNativePoolStaysBoundedAcrossScroll(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	model, err := virtual.NewTable(largeSource{}, 20)
	if err != nil {
		t.Fatal(err)
	}
	v := &VirtualTable{Owner: a, Model: model}
	host := core.NewBody("headless")
	native := v.Mount(host)
	if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
	first := v.rows[0]
	if err := model.SetViewport(999980, 20); err != nil {
		t.Fatal(err)
	}
	if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
	if len(v.rows) != 20 || first != v.rows[0] || first.Text != "999980" {
		t.Fatal("row pool grew or lost identity")
	}
	model.Selected = "999980"
	if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
}

func TestTrackpadDeltasAccumulate(t *testing.T) {
	var remaining float32
	total := 0
	for i := 0; i < 48; i++ {
		total += wheelRows(&remaining, .5)
	}
	if total != 1 || remaining != 0 {
		t.Fatal(total, remaining)
	}
}

func TestExternalTreeCollapseClampsViewport(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	nodes := []virtual.TreeNode{{ID: "root"}}
	for i := 0; i < 100; i++ {
		nodes = append(nodes, virtual.TreeNode{ID: fmt.Sprint(i), Parent: "root"})
	}
	model, err := virtual.NewTree(nodes)
	if err != nil {
		t.Fatal(err)
	}
	model.Toggle("root")
	v := &VirtualTree{Owner: a, Model: model, Start: 80, Count: 20}
	host := core.NewBody("headless")
	native := v.Mount(host)
	if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
	model.Toggle("root")
	if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
	if v.Start != 0 || len(v.visible) != 1 {
		t.Fatal("blank tree after collapse")
	}
}

func TestVirtualAdaptersRemountFreshPools(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	model, _ := virtual.NewTable(largeSource{}, 20)
	table := &VirtualTable{Owner: a, Model: model}
	table.Mount(core.NewBody("first"))
	old := table.rows[0]
	w := table.Mount(core.NewBody("second"))
	if len(table.rows) != 20 || table.rows[0] == old {
		t.Fatal("stale table pool")
	}
	if err := table.Sync(w); err != nil {
		t.Fatal(err)
	}
	modelTree, _ := virtual.NewTree([]virtual.TreeNode{{ID: "root"}})
	tree := &VirtualTree{Owner: a, Model: modelTree, Count: 20}
	tree.Mount(core.NewBody("first tree"))
	old = tree.rows[0]
	w = tree.Mount(core.NewBody("second tree"))
	if len(tree.rows) != 20 || tree.rows[0] == old {
		t.Fatal("stale tree pool")
	}
	if err := tree.Sync(w); err != nil {
		t.Fatal(err)
	}
}
