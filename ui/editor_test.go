package ui

import (
	"github.com/dshills/vellum/editor"
	"testing"
)

func TestManagedEditorDescriptorsAndFocus(t *testing.T) {
	model, _ := editor.New("a\r\nb\n", editor.Config{})
	node := Node{ID: "doc", Kind: MultilineText, Editor: model, TabInserts: true, SoftWrap: true}
	if err := Validate(node); err != nil {
		t.Fatal(err)
	}
	root := Node{ID: "root", Kind: Column, Children: []Node{node, {ID: "button", Kind: Button}}}
	if next, ok := NextFocus(root, "doc", false); !ok || next != "button" {
		t.Fatal("managed editor missing focus navigation", next, ok)
	}
	root.Children = append(root.Children, Node{ID: "peer", Kind: MultilineText, Editor: model})
	if err := Validate(root); err == nil {
		t.Fatal("same Model mounted more than once")
	}
	for _, invalid := range []Node{{ID: "doc", Kind: MultilineText}, {ID: "button", Kind: Button, Editor: model}, {ID: "field", Kind: TextField, SoftWrap: true}, {ID: "field", Kind: TextField, TabInserts: true}} {
		if err := Validate(invalid); err == nil {
			t.Fatal("invalid managed editor descriptor admitted", invalid)
		}
	}
}
