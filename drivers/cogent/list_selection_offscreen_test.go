//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"slices"
	"testing"
)

func TestMultipleListRoutedModifiersKeyboardRejectionAndAX(t *testing.T) {
	selected := []string{}
	reject, disabled := false, false
	count := 0
	view := func() ui.Node {
		return ui.Node{ID: "list", Kind: ui.ListBox, Multiple: true, SelectedIDs: selected, Disabled: disabled, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "x", Label: "Unavailable", Disabled: true}, {ID: "c", Label: "Charlie"}}, SelectionChange: func(e ui.ListSelectionEvent) {
			count++
			if !reject {
				selected = e.Apply(selected)
			}
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 919), view)
	unitSettle(t, m)
	r := m.implementation.find("one").state.root
	widths := map[int]bool{}
	for _, row := range r.control.choice.buttons {
		widths[row.Geom.TotalBBox.Dx()] = true
	}
	if len(widths) != 1 {
		t.Fatal("source list rows must fill list width", widths)
	}
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snap.Windows[0]
	click := func(id string, mod key.Modifiers) {
		t.Helper()
		point := r.control.choice.buttons[id].Geom.TotalBBox
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Position: point.Min.Add(point.Size().Div(2)), Button: events.Left, Modifiers: mod}); err != nil {
				t.Fatal(err)
			}
		}
		unitSettle(t, m)
	}
	click("a", 0)
	click("c", 1<<key.Meta)
	if !slices.Equal(selected, []string{"a", "c"}) {
		t.Fatal("Command membership", selected)
	}
	click("a", 1<<key.Shift)
	if !slices.Equal(selected, []string{"a", "b", "c"}) {
		t.Fatal("Shift range/disabled skip", selected)
	}
	click("b", 1<<key.Meta)
	if !slices.Equal(selected, []string{"a", "c"}) {
		t.Fatal("toggle removal", selected)
	}
	reject = true
	click("c", 0)
	if !slices.Equal(selected, []string{"a", "c"}) || r.control.choice.buttons["b"].IsChecked() {
		t.Fatal("rejected selection", selected)
	}
	reject = false
	click("a", 0)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeDownArrow, Modifiers: 1 << key.Shift}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !slices.Equal(selected, []string{"a", "b"}) {
		t.Fatal("Shift arrow", selected)
	}
	nodes, handlers := accessibilitySnapshot(r, nil)
	selectedRows := 0
	for _, n := range nodes {
		if n.Role == "list-option" && n.Selected {
			selectedRows++
		}
	}
	if selectedRows != 2 {
		t.Fatal("AX membership", selectedRows)
	}
	old := handlers[accessibilityOptionID("list", "c")]
	if old == nil || !old(accessibilityAction{Kind: "press"}) {
		t.Fatal("AX press absent")
	}
	unitSettle(t, m)
	if !slices.Equal(selected, []string{"c"}) {
		t.Fatal("AX replacement", selected)
	}
	before := count
	if err := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if old(accessibilityAction{Kind: "press"}) || count != before {
		t.Fatal("disabled stale action")
	}
}
