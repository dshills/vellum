//go:build offscreen

package cogent

import (
	"bytes"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
)

// Exercise the focused native event path, rather than SetValue/SendChange.
// Accepted model values must also win when the application rejects an edit.
func TestOrdinarySelectorAndNumberKeyboardControlledRoundTrip(t *testing.T) {
	for _, kind := range []ui.Kind{ui.Selector, ui.NumberField} {
		t.Run(string(kind), func(t *testing.T) {
			selected, number := "a", float32(1)
			accept, disabled, calls := true, false, 0
			var proposed ui.ChangeEvent
			view := func() ui.Node {
				control := ui.Node{ID: "value", Kind: kind, Disabled: disabled, Change: func(e ui.ChangeEvent) {
					calls++
					proposed = e
					if accept {
						selected, number = e.SelectedID, e.Number
					}
				}}
				if kind == ui.Selector {
					control.SelectedID = selected
					control.Options = []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "blocked", Label: "Unavailable", Disabled: true}, {ID: "c", Label: "Charlie"}}
				} else {
					control.Number, control.Numeric = number, ui.Numeric{Min: 1, Max: 1.5, Step: .25}
				}
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "value", Sequence: 1}, Children: []ui.Node{control}}
			}
			m, a := simpleManual(t, manualConfig(t, 7), view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			before, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			code := key.CodeDownArrow
			if kind == ui.NumberField {
				code = key.CodeUpArrow
			}
			send := func() {
				t.Helper()
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			send()
			if calls != 1 || proposed.ID != "value" || proposed.Kind != kind {
				t.Fatal("focused keyboard change did not deliver exactly once", calls, proposed)
			}
			if kind == ui.Selector && selected != "c" || kind == ui.NumberField && number != 1.25 {
				t.Fatal("disabled option skip or numeric step failed", selected, number)
			}
			after, err := m.Capture(testContext, "one")
			if err != nil || bytes.Equal(before.Pix, after.Pix) {
				t.Fatal("accepted edit did not reach completed pixels", err)
			}
			accept = false
			send()
			if calls != 2 || kind == ui.Selector && proposed.SelectedID != "a" || kind == ui.NumberField && proposed.Number != 1.5 {
				t.Fatal("rejected proposal missing or duplicated", calls, proposed)
			}
			r := findRecord(w.state.root, "value")
			if kind == ui.Selector && r.widget.(*core.Chooser).CurrentItem.Value != "c" || kind == ui.NumberField && r.widget.(*core.Spinner).Value != 1.25 {
				t.Fatal("rejected edit remained in native widget")
			}
			if err := a.Post(func() {
				disabled = true
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
			}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if !r.widget.AsWidget().StateIs(states.Disabled) {
				t.Fatal("disabled model was not applied to native widget")
			}
			send()
			if calls != 2 {
				t.Fatal("disabled control admitted keyboard edit", calls)
			}
		})
	}
}
