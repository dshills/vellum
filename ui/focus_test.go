package ui

import (
	"reflect"
	"testing"
)

func TestFocusOrderModalTabsDisabledAndWrap(t *testing.T) {
	root := Node{ID: "root", Kind: Column, Children: []Node{
		{ID: "outside", Kind: Button},
		{ID: "tabs", Kind: Tabs, SelectedID: "one", Options: []Option{{ID: "one"}, {ID: "two"}}, Children: []Node{
			{ID: "one", Kind: Column, Children: []Node{{ID: "selected", Kind: TextField}}},
			{ID: "two", Kind: Column, Children: []Node{{ID: "hidden", Kind: TextField}}},
		}},
		{ID: "disabled", Kind: Column, Disabled: true, Children: []Node{{ID: "blocked", Kind: Button}}},
		{ID: "modal", Kind: Dialog, Children: []Node{{ID: "first", Kind: TextField}, {ID: "last", Kind: Button}}},
	}}
	if got := FocusOrder(root); !reflect.DeepEqual(got, []string{"outside", "tabs", "selected"}) {
		t.Fatalf("order %v", got)
	}
	root.Children[3].Open = true
	if got := FocusOrder(root); !reflect.DeepEqual(got, []string{"first", "last"}) {
		t.Fatalf("modal order %v", got)
	}
	for _, tc := range []struct {
		current string
		reverse bool
		want    string
	}{{"first", false, "last"}, {"last", false, "first"}, {"first", true, "last"}, {"missing", true, "last"}, {"missing", false, "first"}} {
		got, ok := NextFocus(root, tc.current, tc.reverse)
		if !ok || got != tc.want {
			t.Fatalf("%+v => %q %v", tc, got, ok)
		}
	}
	root.Children[3].Disabled = true
	if FocusScope(root).ID != "root" {
		t.Fatal("disabled modal owns focus")
	}
}
func TestFocusScopeNestedAndEmpty(t *testing.T) {
	root := Node{ID: "root", Kind: Column, Children: []Node{{ID: "outer", Kind: Dialog, Open: true, Children: []Node{{ID: "before", Kind: Button}, {ID: "inner", Kind: Dialog, Open: true, Children: []Node{{ID: "inside", Kind: Button}}}}}}}
	if got := FocusOrder(root); !reflect.DeepEqual(got, []string{"inside"}) {
		t.Fatal(got)
	}
	if _, ok := NextFocus(Node{ID: "empty", Kind: Column}, "", false); ok {
		t.Fatal("empty focus target")
	}
}
