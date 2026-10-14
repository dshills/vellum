package ui

import "testing"

func TestTabCloseMetadataRequiresTabsHandlerAndExistingPages(t *testing.T) {
	valid := Node{ID: "tabs", Kind: Tabs, SelectedID: "one", Options: []Option{{ID: "one", Label: "One", Closable: true}}, Children: []Node{{ID: "one", Kind: Column}}, TabClose: func(TabCloseEvent) {}}
	if err := Validate(valid); err != nil {
		t.Fatal(err)
	}
	for _, n := range []Node{{ID: "not-tabs", Kind: Selector, Options: []Option{{ID: "one", Closable: true}}, TabClose: func(TabCloseEvent) {}}, {ID: "button", Kind: Button, TabClose: func(TabCloseEvent) {}}, {ID: "tabs", Kind: Tabs, SelectedID: "one", Options: valid.Options, Children: valid.Children}, {ID: "tabs", Kind: Tabs, SelectedID: "one", TabClose: valid.TabClose}} {
		if Validate(n) == nil {
			t.Fatal("invalid close contract accepted", n.Kind)
		}
	}
	if order := FocusOrder(valid); len(order) != 1 || order[0] != "tabs" {
		t.Fatal("close buttons changed managed focus order", order)
	}
}
func TestTabCloseGeneratedTargetNamespaceCannotAliasCallerNode(t *testing.T) {
	root := Node{ID: "root", Kind: Column, Children: []Node{{ID: "tab-close:dGFicw:YQ", Kind: Button, Text: "Collision"}}}
	if Validate(root) == nil {
		t.Fatal("caller could alias generated private close target")
	}
}
