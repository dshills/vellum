package ui

import (
	"reflect"
	"testing"
)

func TestDisclosureValidationAndVisibleFocusScope(t *testing.T) {
	n := Node{ID: "section", Kind: Disclosure, Text: "Details", Children: []Node{
		{ID: "field", Kind: TextField},
		{ID: "modal", Kind: Dialog, Open: true, Children: []Node{{ID: "confirm", Kind: Button}}},
	}}
	if err := Validate(n); err != nil || !IsContainer(Disclosure) || !IsFocusable(Disclosure) {
		t.Fatal("disclosure admission", err)
	}
	if !reflect.DeepEqual(FocusOrder(n), []string{"section"}) || FocusScope(n).ID != "section" {
		t.Fatal("collapsed modal/children own focus")
	}
	if _, ok := FocusTarget(n, "field"); ok {
		t.Fatal("collapsed child admits a focus request")
	}
	n.Checked = true
	if FocusScope(n).ID != "modal" || !reflect.DeepEqual(FocusOrder(n), []string{"confirm"}) {
		t.Fatal("expanded nested modal not selected")
	}
	n.Children = n.Children[:1]
	if !reflect.DeepEqual(FocusOrder(n), []string{"section", "field"}) {
		t.Fatal("expanded visual focus order")
	}
	n.Disabled = true
	if len(FocusOrder(n)) != 0 {
		t.Fatal("disabled group remains focusable")
	}
	for _, change := range []func(*Node){func(n *Node) { n.Text = "  " }, func(n *Node) { n.Activate = func() {} }, func(n *Node) { n.Orientation = Vertical }} {
		invalid := n
		change(&invalid)
		if Validate(invalid) == nil {
			t.Fatal("invalid disclosure admitted", invalid)
		}
	}
	if Validate(Node{ID: "disclosure:reserved", Kind: Label}) == nil {
		t.Fatal("generated heading namespace admitted as a domain ID")
	}
}
