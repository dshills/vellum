//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestSettingsPaneNativeNavigationRetainsEditors(t *testing.T) {
	p := &ui.SettingsPane{ID: "settings", Title: "Settings", SelectedID: "general", Pages: []ui.SettingsPage{{ID: "general", Label: "General", Children: []ui.Node{{ID: "name", Kind: ui.TextField, Text: "Retained"}}}, {ID: "appearance", Label: "Appearance", Children: []ui.Node{{ID: "grid", Kind: ui.Checkbox, Text: "Grid"}}}}}
	p.Navigate = func(e ui.SettingsEvent) { p.SelectedID = e.PageID }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.SettingsNavigationID(p.ID), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 827), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	field := findRecord(w.state.root, "name").widget
	tabs := findRecord(w.state.root, "settings/settings/pages")
	bar := tabs.control.tabs.buttons[ui.SettingsPageID(p.ID, "general")].AsTree().Parent.(core.Widget)
	if bar.AsWidget().IsDisplayable() {
		t.Fatal("duplicate tab header visible")
	}
	for i, code := range []key.Codes{key.CodeDownArrow, key.CodeUpArrow} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		if field.AsWidget().IsVisible() != (i == 1) {
			t.Fatal("inactive page scope")
		}
	}
	if findRecord(w.state.root, "name").widget != field {
		t.Fatal("editor recreated")
	}
}
