//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestNumberEnterAdvancesAndRoutedPointerRefocusAllowsStep(t *testing.T) {
	number, calls, outside := float32(25), 0, 0
	m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "number", Sequence: 1}, Children: []ui.Node{
			{ID: "number", Kind: ui.NumberField, Number: number, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: func(e ui.ChangeEvent) { number = e.Number; calls++ }},
			{ID: "outside", Kind: ui.Button, Text: "Outside", Activate: func() { outside++ }},
		}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	send := func(code key.Codes, r rune) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Rune: r}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	for range 4 {
		send(key.CodeRightArrow, 0)
	}
	for range 3 {
		send(key.CodeBackspace, 0)
	}
	send(key.CodeUnknown, '4')
	send(key.CodeUnknown, '2')
	send(key.CodeReturnEnter, 0)
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if number != 42 || calls != 1 || snap.Windows[0].FocusedID != "outside" {
		t.Fatal("Enter must commit once and advance actual focus", number, calls, snap.Windows[0].FocusedID)
	}
	bounds := findRecord(w.state.root, "number").widget.AsWidget().Geom.TotalBBox
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: bounds.Min.Add(bounds.Size().Div(2))}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	snap, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Windows[0].FocusedID != "number" || calls != 1 || outside != 0 {
		t.Fatal("routed refocus changed model or failed to focus", snap.Windows[0].FocusedID, calls, outside)
	}
	send(key.CodeUpArrow, 0)
	if number != 43 || calls != 2 || outside != 0 {
		t.Fatal("focused step after Enter/refocus failed", number, calls, outside)
	}
}
