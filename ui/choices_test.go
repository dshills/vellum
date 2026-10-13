package ui

import (
	"reflect"
	"testing"
)

func TestChoiceDescriptorsAndNavigation(t *testing.T) {
	for _, kind := range []Kind{RadioGroup, SegmentedControl} {
		n := Node{ID: "group", Kind: kind, SelectedID: "a", Options: []Option{{ID: "a", Label: "same"}, {ID: "b", Label: "same", Disabled: true}, {ID: "c", Label: "C"}}}
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
		for _, test := range []struct {
			current string
			step    int
			want    string
		}{{"a", 1, "c"}, {"c", 1, "a"}, {"a", -1, "c"}, {"", 1, "a"}, {"", -1, "c"}, {"absent", 1, "a"}} {
			got, ok := NextChoiceOption(n, test.current, test.step)
			if !ok || got != test.want {
				t.Fatal(kind, test, got, ok)
			}
		}
		n.Orientation = Vertical
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
		n.SelectedID = "b"
		if err := Validate(n); err != nil {
			t.Fatal("disabled accepted selection refused", err)
		}
		root := Node{ID: "root", Kind: Row, Children: []Node{{ID: "before", Kind: Button}, n, {ID: "after", Kind: Button}}}
		if order := FocusOrder(root); !reflect.DeepEqual(order, []string{"before", "group", "after"}) {
			t.Fatal("group not one managed focus stop", order)
		}
		for i := range n.Options {
			n.Options[i].Disabled = true
		}
		if _, ok := FocusTarget(n, "group"); ok {
			t.Fatal("all-disabled group accepted focus")
		}
		if _, ok := NextChoiceOption(n, "a", 1); ok {
			t.Fatal("all-disabled group navigated")
		}
		n.Options[0].Disabled = false
		n.Disabled = true
		if _, ok := NextChoiceOption(n, "a", 1); ok {
			t.Fatal("disabled group navigated")
		}
	}
}
func TestChoiceValidationBoundsAndLeaf(t *testing.T) {
	n := Node{ID: "choice", Kind: RadioGroup, SelectedID: "a", Options: []Option{{ID: "a", Label: "A"}}}
	for _, bad := range []Node{
		{ID: "empty", Kind: RadioGroup},
		{ID: "unselected", Kind: SegmentedControl, Options: n.Options},
		{ID: "missing", Kind: RadioGroup, SelectedID: "b", Options: n.Options},
		{ID: "duplicate", Kind: RadioGroup, SelectedID: "a", Options: []Option{{ID: "a"}, {ID: "a"}}},
		{ID: "child", Kind: RadioGroup, SelectedID: "a", Options: n.Options, Children: []Node{{ID: "leaf", Kind: Label}}},
	} {
		if Validate(bad) == nil {
			t.Fatal("bad descriptor accepted", bad.ID)
		}
	}
	n.Options = make([]Option, MaxChoiceOptions+1)
	if Validate(n) == nil {
		t.Fatal("unbounded native options accepted")
	}
	if _, ok := NextChoiceOption(Node{Kind: Selector}, "", 1); ok {
		t.Fatal("wrong kind navigated")
	}
}
