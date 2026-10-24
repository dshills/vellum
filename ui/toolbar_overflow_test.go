package ui

import "testing"

func TestToolbarOverflowPublicComposition(t *testing.T) {
	calls := 0
	source := Node{ID: "tools", Kind: Toolbar, Children: []Node{
		{ID: "a", Kind: Button, Icon: SymbolAdd, Accessibility: Accessibility{Name: "New"}, Activate: func() { calls++ }},
		{ID: "sep", Kind: Separator},
		{ID: "b", Kind: Button, Text: "Delete", Disabled: true, Activate: func() { calls += 10 }},
	}}
	n := WithToolbarOverflow(source)
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	if len(source.Children) != 3 || n.Children[0].ID != "a" {
		t.Fatal("mutated input or identity")
	}
	menu := n.Children[3]
	if menu.ID != ToolbarOverflowID("tools") || menu.Options[0].Label != "New" || menu.Options[0].Icon != SymbolAdd {
		t.Fatal("lost artwork, label or identity")
	}
	menu.Change(ChangeEvent{SelectedID: "a"})
	menu.Change(ChangeEvent{SelectedID: "b"})
	menu.Change(ChangeEvent{SelectedID: "unknown"})
	if calls != 1 {
		t.Fatal("duplicate/disabled action", calls)
	}
	for _, change := range []func(*Node){
		func(n *Node) { n.Kind = Row }, func(n *Node) { n.Orientation = Vertical },
		func(n *Node) { n.Children[0].Kind = TextField }, func(n *Node) { n.Children[0].Accessibility.Name = "" },
		func(n *Node) { n.Children[3].ID = "wrong" },
	} {
		bad := WithToolbarOverflow(source)
		change(&bad)
		if Validate(bad) == nil {
			t.Fatal("unsupported toolbar accepted", bad)
		}
	}
}
func TestMenuOptionArtworkValidation(t *testing.T) {
	base := Node{ID: "menu", Kind: MenuButton, Text: "Actions", Options: []Option{{ID: "one", Label: "Open", Icon: SymbolFolder}}}
	if e := Validate(base); e != nil {
		t.Fatal(e)
	}
	base.Options[0].Icon = "unknown"
	if Validate(base) == nil {
		t.Fatal("unknown symbol")
	}
	base.Options[0] = Option{ID: "sep", Separator: true, Icon: SymbolFolder}
	if Validate(base) == nil {
		t.Fatal("decorated separator")
	}
}
