package ui

import "testing"

func TestSidebarControlledAndStale(t *testing.T) {
	s := &Sidebar{ID: "nav", SelectedID: "a", Sections: []SidebarSection{{ID: "files", Label: "Favorites", Items: []Option{{ID: "a", Label: "All", Icon: SymbolFolder, Badge: "12"}, {ID: "b", Label: "Recent", Icon: SymbolDocument}}}}}
	calls := 0
	s.Navigate = func(e SidebarEvent) { calls++; s.SelectedID = e.ItemID }
	s.Disclosure = func(e SidebarDisclosureEvent) { s.Sections[0].Collapsed = !e.Expanded }
	n, err := s.View()
	if err != nil {
		t.Fatal(err)
	}
	list := n.Children[0].Children[0]
	list.SelectionChange(ListSelectionEvent{TargetID: "b", IDs: []string{"b"}})
	if calls != 1 || s.SelectedID != "b" {
		t.Fatal(calls, s.SelectedID)
	}
	s.Sections[0].Items[0].Disabled = true
	list.SelectionChange(ListSelectionEvent{TargetID: "a", IDs: []string{"a"}})
	if calls != 1 {
		t.Fatal("stale item accepted")
	}
	n.Children[0].Change(ChangeEvent{Checked: false})
	if !s.Sections[0].Collapsed {
		t.Fatal("collapse")
	}
	list.SelectionChange(ListSelectionEvent{TargetID: "b", IDs: []string{"b"}})
	if calls != 1 {
		t.Fatal("hidden navigation")
	}
	s.Disabled = true
	n.Children[0].Change(ChangeEvent{Checked: true})
	if !s.Sections[0].Collapsed {
		t.Fatal("disabled disclosure")
	}
}
func TestSidebarValidation(t *testing.T) {
	s := &Sidebar{ID: "nav", Sections: []SidebarSection{{ID: "a", Label: "A", Items: []Option{{ID: "i", Label: "One", Icon: SymbolFolder}}}, {ID: "b", Label: "B", Items: []Option{{ID: "i", Label: "Two"}}}}}
	if _, err := s.View(); err == nil {
		t.Fatal("duplicate identity")
	}
	s.Sections = s.Sections[:1]
	s.SelectedID = "missing"
	if _, err := s.View(); err == nil {
		t.Fatal("missing selection")
	}
	s.SelectedID = ""
	s.Sections[0].Items[0].Icon = "unknown"
	if _, err := s.View(); err == nil {
		t.Fatal("unknown artwork")
	}
	if ValidateControl(Node{Kind: Selector, Options: []Option{{ID: "a", Icon: SymbolFolder}}}) == nil {
		t.Fatal("decoration on unsupported kind")
	}
}
