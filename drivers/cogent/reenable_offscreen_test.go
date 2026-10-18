//go:build offscreen

package cogent

import (
	"fmt"
	"strings"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/abilities"
	"github.com/dshills/vellum/ui"
)

// An ancestor update walks its descendants before reconciliation synchronizes
// each child's inherited disabled state. The child's descriptor can be unchanged.
func TestControlsRegainPointerFocusAfterDisabledTransition(t *testing.T) {
	for _, kind := range []ui.Kind{ui.NumberField, ui.TextField} {
		for _, inherited := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/inherited=%t", kind, inherited), func(t *testing.T) {
				disabled, calls, outside := false, 0, 0
				value, text := float32(25), "initial"
				m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
					return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
						{ID: "fields", Kind: ui.FormGroup, Text: "Fields", Disabled: inherited && disabled, Children: []ui.Node{
							{ID: "field", Kind: kind, Disabled: !inherited && disabled, Number: value, Text: text, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: func(e ui.ChangeEvent) { calls++; value = e.Number; text = e.Text }},
						}},
						{ID: "outside", Kind: ui.Button, Text: "Outside", Activate: func() { outside++ }},
					}}
				})
				unitSettle(t, m)
				w := m.implementation.find("one")
				widget := findRecord(w.state.root, "field").widget
				click := func(id string) {
					t.Helper()
					b := findRecord(w.state.root, id).widget.AsWidget().Geom.TotalBBox
					for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
						if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: b.Min.Add(b.Size().Div(2))}); err != nil {
							t.Fatal(err)
						}
						unitSettle(t, m)
					}
				}
				chord := func(code key.Codes, r rune) {
					t.Helper()
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Rune: r}); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
				}
				setDisabled := func(v bool) {
					t.Helper()
					if err := a.Do(func() { disabled = v; _ = a.Invalidate("one") }); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
				}
				// Repeat without changing the child value/descriptor or replacing widgets.
				for cycle := range 3 {
					setDisabled(true)
					before := calls
					click("field")
					chord(key.CodeUpArrow, 0)
					if calls != before {
						t.Fatal("disabled input changed model", calls, before)
					}
					click("outside")
					setDisabled(false)
					current := findRecord(w.state.root, "field").widget
					if current != widget {
						t.Fatal("re-enable replaced retained widget")
					}
					wb := current.AsWidget()
					if wb.IsReadOnly() || !wb.AbilityIs(abilities.Focusable) {
						t.Fatal("reenabled control kept disabled styles", cycle, wb.IsReadOnly(), wb.AbilityIs(abilities.Focusable))
					}
					click("field")
					snap, err := m.Snapshot(testContext)
					if err != nil {
						t.Fatal(err)
					}
					if snap.Windows[0].FocusedID != "field" || calls != before {
						t.Fatal("reenabled pointer focus", snap.Windows[0].FocusedID, calls, before)
					}
					if kind == ui.NumberField {
						old := value
						chord(key.CodeUpArrow, 0)
						if value != old+1 || calls != before+1 {
							t.Fatal("reenabled number step", value, old, calls, before)
						}
					} else {
						chord(key.CodeEnd, 0)
						chord(key.CodeUnknown, 'x')
						chord(key.CodeReturnEnter, 0)
						if text != "initial"+strings.Repeat("x", cycle+1) || calls != before+1 {
							t.Fatal("reenabled text commit", text, calls, before)
						}
					}
					if outside != cycle+1 {
						t.Fatal("field release activated unrelated button", outside, cycle)
					}
				}
			})
		}
	}
}
