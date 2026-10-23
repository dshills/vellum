package ui

import (
	"slices"
	"testing"
)

func TestListSelectionEditsAndValidation(t *testing.T) {
	current := []string{"a"}
	a := ListSelectionEvent{IDs: []string{"b"}, Toggle: true}
	b := ListSelectionEvent{IDs: []string{"c"}, Toggle: true}
	got := b.Apply(a.Apply(current))
	if !slices.Equal(got, []string{"a", "b", "c"}) || !slices.Equal(current, []string{"a"}) {
		t.Fatal(got, current)
	}
	got = a.Apply(got)
	if !slices.Equal(got, []string{"a", "c"}) {
		t.Fatal(got)
	}
	n := Node{ID: "list", Kind: ListBox, Multiple: true, Options: []Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, SelectedIDs: []string{"a"}}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	for _, ids := range [][]string{{"a", "a"}, {"missing"}} {
		bad := n
		bad.SelectedIDs = ids
		if Validate(bad) == nil {
			t.Fatal("bad IDs admitted")
		}
	}
	n.Change = func(ChangeEvent) {}
	if Validate(n) == nil {
		t.Fatal("mixed selection callbacks")
	}
	if Validate(Node{ID: "button", Kind: Button, Multiple: true}) == nil {
		t.Fatal("multiple selection outside list")
	}
}

func TestListSelectionRepeatedProposalIDOnlyTogglesOnce(t *testing.T) {
	e := ListSelectionEvent{IDs: []string{"a", "a"}, Toggle: true}
	if got := e.Apply(nil); !slices.Equal(got, []string{"a"}) {
		t.Fatal(got)
	}
}
