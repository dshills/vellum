package ui

import (
	"slices"
	"testing"
)

func TestOutlineViewProjectionHiddenSelectionAndStaleDisclosure(t *testing.T) {
	p := &OutlineView{ID: "outline", Label: "Objects", Items: []OutlineItem{{ID: "root", Label: "Root"}, {ID: "child", Parent: "root", Label: "Child"}, {ID: "leaf", Parent: "child", Label: "Leaf"}}, ExpandedIDs: []string{"root", "child"}, SelectedIDs: []string{"leaf"}}
	calls := 0
	p.Disclosure = func(e OutlineDisclosureEvent) { calls++; p.ExpandedIDs = e.Apply(p.ExpandedIDs) }
	p.Change = func(e ListSelectionEvent) { p.SelectedIDs = e.Apply(p.SelectedIDs) }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Options) != 3 || n.Outline.Rows[2].Depth != 2 {
		t.Fatal("bad hierarchy")
	}
	n.Outline.Disclosure(OutlineDisclosureEvent{ItemID: "root", Expanded: false})
	next, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Options) != 1 || len(next.SelectedIDs) != 0 || !slices.Equal(p.SelectedIDs, []string{"leaf"}) {
		t.Fatal("collapse discarded hidden selection")
	}
	n.Outline.Disclosure(OutlineDisclosureEvent{ItemID: "child", Expanded: false})
	if calls != 1 {
		t.Fatal("hidden stale disclosure acted")
	}
	p.Items[0].Disabled = true
	p.ExpandedIDs = []string{"root", "child"}
	next, err = p.View()
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range next.Options {
		if !option.Disabled {
			t.Fatal("disabled ancestry escaped")
		}
	}
}
func TestOutlineViewRejectsMalformedTrees(t *testing.T) {
	for _, items := range [][]OutlineItem{{{ID: "a", Parent: "missing", Label: "A"}}, {{ID: "a", Parent: "b", Label: "A"}, {ID: "b", Parent: "a", Label: "B"}}, {{ID: "a", Label: "A"}, {ID: "a", Label: "B"}}} {
		p := &OutlineView{ID: "tree", Label: "Tree", Items: items}
		if _, err := p.View(); err == nil {
			t.Fatal("malformed tree admitted")
		}
	}
}

func TestOutlineViewReparentAndDisableRevokeOldActions(t *testing.T) {
	p := &OutlineView{ID: "tree", Label: "Tree", Items: []OutlineItem{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "c", Parent: "a", Label: "C"}, {ID: "d", Parent: "c", Label: "D"}}, ExpandedIDs: []string{"a", "b", "c"}}
	// b starts without children, so do not mark it expanded until reparenting.
	p.ExpandedIDs = []string{"a", "c"}
	calls := 0
	p.Change = func(ListSelectionEvent) { calls++ }
	p.Disclosure = func(OutlineDisclosureEvent) { calls++ }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	p.Items[2].Parent = "b"
	p.ExpandedIDs = []string{"b", "c"}
	n.SelectionChange(ListSelectionEvent{TargetID: "c", IDs: []string{"c"}})
	n.Outline.Disclosure(OutlineDisclosureEvent{ItemID: "c", Expanded: false})
	if calls != 0 {
		t.Fatal("reparented stale action admitted")
	}
	n, err = p.View()
	if err != nil {
		t.Fatal(err)
	}
	p.Items[2].Disabled = true
	n.SelectionChange(ListSelectionEvent{TargetID: "c", IDs: []string{"c"}})
	n.Outline.Disclosure(OutlineDisclosureEvent{ItemID: "c", Expanded: false})
	if calls != 0 {
		t.Fatal("disabled stale action admitted")
	}
}
