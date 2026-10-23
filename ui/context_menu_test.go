package ui

import "testing"

func TestContextMenuContract(t *testing.T) {
	menu := &ContextMenu{Options: []Option{{ID: "a", Label: "Action"}, {ID: "sep", Separator: true}, {ID: "check", Label: "Check", Checkable: true, Checked: true}}}
	if err := Validate(Node{ID: "host", Kind: Button, Text: "Host", ContextMenu: menu}); err != nil {
		t.Fatal(err)
	}
	for _, n := range []Node{{ID: "a", Kind: Label, ContextMenu: menu}, {ID: "a", Kind: Button, ContextMenu: &ContextMenu{}}, {ID: "a", Kind: Button, ContextMenu: &ContextMenu{Options: []Option{{ID: "a", Label: "A"}, {ID: "a", Label: "B"}}}}} {
		if Validate(n) == nil {
			t.Fatal("invalid menu accepted")
		}
	}
}
