package ui

import (
	"slices"
	"testing"
)

func TestSettingsPaneRetainedPagesAndNavigationGuards(t *testing.T) {
	p := &SettingsPane{ID: "settings", Title: "Settings", SelectedID: "general", Pages: []SettingsPage{{ID: "general", Label: "General", Children: []Node{{ID: "name", Kind: TextField, Text: "Drawing"}}}, {ID: "appearance", Label: "Appearance", Children: []Node{{ID: "visible", Kind: Checkbox, Text: "Visible"}}}}}
	calls := 0
	p.Navigate = func(e SettingsEvent) { calls++; p.SelectedID = e.PageID }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	stack := pickerNode(n, "settings/settings/pages")
	if !stack.TabsHideHeader || len(stack.Children) != 2 {
		t.Fatal("inactive pages removed")
	}
	order := FocusOrder(n)
	if slices.Contains(order, stack.ID) || slices.Contains(order, "visible") || !slices.Contains(order, "name") {
		t.Fatal("hidden page/header leaked focus", order)
	}
	nav := pickerNode(n, SettingsNavigationID(p.ID))
	nav.SelectionChange(ListSelectionEvent{TargetID: "appearance"})
	if calls != 1 || p.SelectedID != "appearance" {
		t.Fatal("navigation")
	}
	p.Pages[0].Disabled = true
	nav.SelectionChange(ListSelectionEvent{TargetID: "general"})
	if calls != 1 {
		t.Fatal("disabled page admitted")
	}
}
