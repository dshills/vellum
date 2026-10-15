package ui

import (
	"fmt"
	"testing"
)

func TestListBoxValidationAndFocus(t *testing.T) {
	base := Node{ID: "items", Kind: ListBox, Options: []Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta", Disabled: true}}, SelectedID: "a"}
	for _, test := range []struct {
		name string
		edit func(*Node)
	}{
		{"duplicate", func(n *Node) { n.Options[1].ID = "a" }},
		{"blank-id", func(n *Node) { n.Options[0].ID = "  " }},
		{"nul-id", func(n *Node) { n.Options[0].ID = "a\x00" }},
		{"utf8-id", func(n *Node) { n.Options[0].ID = "\xff" }},
		{"blank-label", func(n *Node) { n.Options[0].Label = " " }},
		{"utf8-label", func(n *Node) { n.Options[0].Label = "\xff" }},
		{"missing-selection", func(n *Node) { n.SelectedID = "missing" }},
		{"activate", func(n *Node) { n.Activate = func() {} }},
		{"orientation", func(n *Node) { n.Orientation = Horizontal }},
		{"children", func(n *Node) { n.Children = []Node{{ID: "child", Kind: Label}} }},
		{"menu-metadata", func(n *Node) { n.Options[0].Checked = true }},
		{"tab-metadata", func(n *Node) { n.Options[0].Closable = true }},
		{"oversized", func(n *Node) {
			for len(n.Options) <= MaxListBoxOptions {
				n.Options = append(n.Options, Option{ID: fmt.Sprint(len(n.Options)), Label: "Item"})
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := base
			n.Options = append([]Option(nil), base.Options...)
			test.edit(&n)
			if Validate(n) == nil {
				t.Fatal("invalid list accepted", n)
			}
		})
	}
	for _, selected := range []string{"", "a", "b"} {
		n := base
		n.SelectedID = selected
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
		if _, ok := FocusTarget(n, "items"); !ok {
			t.Fatal("enabled list is not focusable")
		}
	}
	for _, options := range [][]Option{nil, {{ID: "a", Label: "A", Disabled: true}}} {
		n := base
		n.Options = options
		n.SelectedID = ""
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
		if _, ok := FocusTarget(n, "items"); ok {
			t.Fatal("empty/all-disabled list focusable")
		}
	}
	base.Disabled = true
	if _, ok := FocusTarget(base, "items"); ok {
		t.Fatal("disabled list focusable")
	}
}
func TestListBoxClampedNavigation(t *testing.T) {
	n := Node{ID: "list", Kind: ListBox, Options: []Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B", Disabled: true}, {ID: "c", Label: "C"}, {ID: "d", Label: "D"}}}
	for _, test := range []struct {
		from  string
		delta int
		want  string
	}{{"", 1, "a"}, {"", -1, "d"}, {"a", 1, "c"}, {"c", -1, "a"}, {"a", -1, "a"}, {"d", 1, "d"}, {"a", 200, "d"}, {"d", -200, "a"}, {"b", 1, "c"}, {"b", -1, "a"}} {
		got, ok := MoveListBoxOption(n, test.from, test.delta)
		if !ok || got != test.want {
			t.Fatal(test, got, ok)
		}
	}
	for _, delta := range []int{0, 257, -257} {
		if _, ok := MoveListBoxOption(n, "a", delta); ok {
			t.Fatal("invalid movement")
		}
	}
	n.Disabled = true
	if _, ok := MoveListBoxOption(n, "a", 1); ok {
		t.Fatal("disabled movement")
	}
}
