//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestQuickLookSpaceDoesNotActivateOrToggleSelection(t *testing.T) {
	calls, opened := 0, 0
	d := &ui.DocumentList{ID: "files", Label: "Files", Items: []ui.DocumentItem{{ID: "a", Name: "A"}}, SelectedIDs: []string{"a"}, Preview: func(e ui.DocumentEvent) {
		if e.ItemID != "a" {
			t.Error(e)
		}
		calls++
	}, Activate: func(ui.DocumentEvent) { opened++ }}
	d.Change = func(e ui.ListSelectionEvent) { d.SelectedIDs = e.Apply(d.SelectedIDs) }
	view := func() ui.Node {
		n, e := d.View()
		if e != nil {
			t.Fatal(e)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.DocumentListID(d.ID), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 942), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeSpacebar, Modifiers: 1 << key.NumLock}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if calls != 1 || opened != 0 || len(d.SelectedIDs) != 1 {
		t.Fatal(calls, opened, d.SelectedIDs)
	}
}
