//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
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

// A preview often marks its source busy while a native worker is open. The
// accepted owner state must reach controls even if no later pointer event occurs.
func TestQuickLookSpacePublishesBusyStateWithoutAnotherInput(t *testing.T) {
	for _, input := range []struct {
		char rune
		code key.Codes
	}{{0, key.CodeSpacebar}, {' ', key.CodeSpacebar}, {' ', key.CodeUnknown}} {
		t.Run(fmt.Sprint(input.char, "/", input.code), func(t *testing.T) {
			calls := 0
			d := &ui.DocumentList{ID: "files", Label: "Files", Items: []ui.DocumentItem{{ID: "a", Name: "A"}}, SelectedIDs: []string{"a"}}
			d.Preview = func(ui.DocumentEvent) { calls++; d.Disabled = true }
			view := func() ui.Node {
				n, err := d.View()
				if err != nil {
					t.Fatal(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.DocumentListID(d.ID), Sequence: 1}, Children: []ui.Node{n}}
			}
			m, _ := simpleManual(t, manualConfig(t, 943), view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			send := func() {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Rune: input.char, Code: input.code, Modifiers: 1 << key.NumLock}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			send()
			record := findRecord(w.state.root, ui.DocumentListID(d.ID))
			if calls != 1 || !record.node.Disabled || !controlDisabled(record) {
				t.Fatal("busy state was not published", calls, record.node.Disabled)
			}
			send()
			if calls != 1 {
				t.Fatal("disabled source opened again", calls)
			}
		})
	}
}
