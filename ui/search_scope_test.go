package ui

import (
	"strings"
	"testing"
)

func TestSearchScopeEditsMergeAndStaleChoicesRefuse(t *testing.T) {
	s := &SearchScopeBar{ID: "search", Label: "Search", Value: SearchScopeValue{ScopeID: "all"}, Items: []SearchScopeItem{{ID: "all", Label: "All", Count: 2}, {ID: "project", Label: "Project", Count: 1}, {ID: "blocked", Label: "Blocked", Disabled: true, Count: -1}}}
	calls := 0
	s.Change = func(e SearchScopeEvent) { calls++; s.Value = e.Apply(s.Value) }
	n, e := s.View()
	if e != nil {
		t.Fatal(e)
	}
	project := pickerNode(n, SearchScopeItemID(s.ID, "project"))
	input := pickerNode(n, SearchInputID(s.ID+"/scope/query"))
	project.Activate()
	input.Change(ChangeEvent{Text: "drawing"})
	if s.Value != (SearchScopeValue{ScopeID: "project", Query: "drawing"}) {
		t.Fatal(s.Value)
	}
	pickerNode(n, SearchScopeItemID(s.ID, "blocked")).Activate()
	if calls != 2 {
		t.Fatal("disabled scope")
	}
	s.Items = s.Items[:1]
	s.Value.ScopeID = "all"
	project.Activate()
	input.Change(ChangeEvent{Text: strings.Repeat("x", 4097)})
	if calls != 2 {
		t.Fatal("removed/overflow scope")
	}
	s.Disabled = true
	input.Change(ChangeEvent{Text: "x"})
	if calls != 2 {
		t.Fatal("disabled query")
	}
}
func TestSearchScopeBoundsAndCompact(t *testing.T) {
	s := &SearchScopeBar{ID: "search", Label: "Search", Small: true, Value: SearchScopeValue{ScopeID: "all"}, Items: []SearchScopeItem{{ID: "all", Label: "All", Count: 0}}}
	n, e := s.View()
	if e != nil || pickerNode(n, SearchScopeItemID(s.ID, "all")).Style.MinSize.Value.Height != 24 {
		t.Fatal(e)
	}
	s.Items = append(s.Items, s.Items...)
	if _, e = s.View(); e == nil {
		t.Fatal("duplicate")
	}
}

func TestSearchScopeQueuedReturnUsesCurrentAcceptedState(t *testing.T) {
	s := &SearchScopeBar{ID: "scope", Label: "Scope", Value: SearchScopeValue{ScopeID: "all"}, Items: []SearchScopeItem{{ID: "all", Label: "All"}, {ID: "project", Label: "Project"}}}
	calls := 0
	s.Change = func(e SearchScopeEvent) { calls++; s.Value = e.Apply(s.Value) }
	n, _ := s.View()
	pickerNode(n, SearchScopeItemID(s.ID, "project")).Activate()
	pickerNode(n, SearchScopeItemID(s.ID, "all")).Activate()
	pickerNode(n, SearchScopeItemID(s.ID, "all")).Activate()
	if s.Value.ScopeID != "all" || calls != 2 {
		t.Fatal(s.Value, calls)
	}
}
