package ui

import (
	"context"
	"testing"

	"github.com/dshills/vellum/virtual"
)

func outlineForDescriptor(t *testing.T) *virtual.Outline {
	t.Helper()
	data, e := virtual.NewTreeData(context.Background(), "ui-tree", 1, []virtual.TreeNode{{ID: "root", Label: "Root"}}, virtual.TreeLimits{})
	if e != nil {
		t.Fatal(e)
	}
	index, e := virtual.BuildTreeIndex(context.Background(), data, virtual.TreeIndexOptions{})
	if e != nil {
		t.Fatal(e)
	}
	model, e := virtual.NewOutline(index, virtual.OutlineConfig{}, virtual.OutlineViewport{Width: 200, Height: 150})
	if e != nil {
		t.Fatal(e)
	}
	return model
}
func TestDataTreeDescriptor(t *testing.T) {
	model := outlineForDescriptor(t)
	n := Node{ID: "outline", Kind: DataTree, Tree: model}
	if e := Validate(n); e != nil {
		t.Fatal(e)
	}
	if !IsFocusable(DataTree) {
		t.Fatal("tree is not a focus target")
	}
	for name, mutate := range map[string]func(*Node){"nil": func(n *Node) { n.Tree = nil }, "copy": func(n *Node) { copy := *model; n.Tree = &copy }, "foreign": func(n *Node) { n.Kind = Label }, "children": func(n *Node) { n.Children = []Node{{ID: "child", Kind: Label}} }} {
		t.Run(name, func(t *testing.T) {
			copy := n
			mutate(&copy)
			if e := Validate(copy); e == nil {
				t.Fatal("invalid descriptor accepted")
			}
		})
	}
	if e := Validate(Node{ID: "root", Kind: Column, Children: []Node{n, {ID: "other", Kind: DataTree, Tree: model}}}); e == nil {
		t.Fatal("duplicate model accepted")
	}
	for _, id := range []string{"tree:x", "tree:eA:row:eQ", "tree:eA:disclosure:eQ", "tree:eA:page:next"} {
		if e := Validate(Node{ID: id, Kind: Button}); e == nil {
			t.Fatal("reserved identity accepted", id)
		}
	}
	if e := Validate(Node{ID: "tree", Kind: Button}); e != nil {
		t.Fatal(e)
	}
	for _, callback := range []Node{{ID: "wrong", Kind: Button, TreeChange: func(TreeEvent) {}}, {ID: "wrong", Kind: Button, TreeDisclosure: func(TreeDisclosureEvent) {}}} {
		if e := Validate(callback); e == nil {
			t.Fatal("foreign callback accepted")
		}
	}
}
