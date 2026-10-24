package ui

import "testing"

func TestPopoverValidationAndFocus(t *testing.T) {
	p := Popover{ID: "panel", AnchorID: "anchor", Open: true, Dismiss: func() {}, Children: []Node{{ID: "value", Kind: TextField}}}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	root := Node{ID: "root", Kind: Column, Children: []Node{{ID: "anchor", Kind: Button}, n}}
	order := FocusOrder(root)
	if len(order) != 1 || order[0] != "value" {
		t.Fatal(order)
	}
	p.Open = false
	n, _ = p.View()
	root.Children[1] = n
	if _, ok := FocusTarget(root, "value"); ok {
		t.Fatal("closed content exposed")
	}
	p.AnchorID = "value"
	if _, err := p.View(); err == nil {
		t.Fatal("descendant anchor admitted")
	}
	p.AnchorID = "anchor"
	p.Dismiss = nil
	if _, err := p.View(); err == nil {
		t.Fatal("missing dismiss admitted")
	}
}
