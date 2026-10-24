package ui

import "testing"

func TestMenuBranchSnapshotAndValidation(t *testing.T) {
	items := []Option{{ID: "leaf", Label: "Save", Shortcut: "⌘S"}}
	branch, err := NewMenuBranch(items)
	if err != nil {
		t.Fatal(err)
	}
	items[0].Label = "changed"
	copy := branch.Items()
	copy[0].ID = "changed"
	n := Node{ID: "m", Kind: MenuButton, Text: "Actions", Options: []Option{{ID: "branch", Label: "Export", Branch: branch}}, Change: func(ChangeEvent) {}}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	if o, ok := MenuAction(n.Options, "leaf"); !ok || o.Label != "Save" {
		t.Fatal(o, ok)
	}
	n.Options[0].Disabled = true
	if _, ok := MenuAction(n.Options, "leaf"); ok || MenuEnabled(n) {
		t.Fatal("disabled ancestor admitted")
	}
	n.Options[0].Disabled = false
	n.Options = append(n.Options, Option{ID: "leaf", Label: "Duplicate"})
	if Validate(n) == nil {
		t.Fatal("duplicate descendant admitted")
	}
	for i := 0; i < MaxMenuDepth; i++ {
		branch, _ = NewMenuBranch([]Option{{ID: string(rune('a' + i)), Label: "Parent", Branch: branch}})
	}
	if branch != nil {
		t.Fatal("excessive depth admitted")
	}
	if _, err := NewMenuBranch([]Option{{ID: "x", Label: "X", Shortcut: "a\nb"}}); err == nil {
		t.Fatal("control caption admitted")
	}
}
