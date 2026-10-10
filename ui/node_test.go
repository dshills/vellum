package ui

import (
	"fmt"
	"testing"
)

func TestIdentityValidation(t *testing.T) {
	n := Node{ID: "root", Kind: Column, Children: []Node{{ID: "same", Kind: Label}, {ID: "same", Kind: Button}}}
	if Validate(n) == nil {
		t.Fatal("duplicate identity")
	}
	n.Children[1].ID = "button"
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
}

func TestValidationBoundaries(t *testing.T) {
	for _, n := range []Node{{Kind: Label}, {ID: "x", Kind: "unknown"}, {ID: "x", Kind: Button, Children: []Node{{ID: "y", Kind: Label}}}} {
		if Validate(n) == nil {
			t.Fatal("invalid node accepted")
		}
	}
	n := Node{ID: "127", Kind: Column}
	for i := 126; i >= 0; i-- {
		n = Node{ID: fmt.Sprint(i), Kind: Column, Children: []Node{n}}
	}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	n = Node{ID: "extra", Kind: Column, Children: []Node{n}}
	if Validate(n) == nil {
		t.Fatal("129 levels accepted")
	}
}
