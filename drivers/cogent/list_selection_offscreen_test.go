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

func TestMultipleListCommandSelectAllAndShortcutIsolation(t *testing.T) {
	selected := []string{"b"}
	calls := 0
	reject := false
	view := func() ui.Node {
		return ui.Node{ID: "list", Kind: ui.ListBox, Multiple: true, SelectedIDs: selected, Focus: ui.FocusRequest{ID: "list", Sequence: 1}, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "disabled", Label: "Unavailable", Disabled: true}, {ID: "c", Label: "Charlie"}}, SelectionChange: func(e ui.ListSelectionEvent) {
			calls++
			if !reject {
				selected = e.Apply(selected)
			}
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 933), view)
	unitSettle(t, m)
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snap.Windows[0]
	send := func(code key.Codes, runeValue rune) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Rune: runeValue, Modifiers: 1 << key.Meta}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	// Command-C is deliberately refused by public Manual input because an
	// unhandled copy may touch OS clipboard services. Dispatch only to the actual
	// retained row for this shortcut-isolation check; do not weaken that guard.
	r := m.implementation.find("one").state.root
	r.control.choice.buttons["b"].HandleEvent(events.NewKey(events.KeyChord, 'c', key.CodeC, 1<<key.Meta))
	unitSettle(t, m)
	if calls != 0 || !slices.Equal(selected, []string{"b"}) {
		t.Fatal("Command-C acted as typeahead", calls, selected)
	}
	send(key.CodeA, 'a')
	if calls != 1 || !slices.Equal(selected, []string{"a", "b", "c"}) {
		t.Fatal("Command-A did not select all enabled rows", calls, selected)
	}
	if err := a.Do(func() { selected = []string{"b"}; reject = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	before := slices.Clone(selected)
	send(key.CodeA, 'a')
	if calls != 2 || !slices.Equal(selected, before) {
		t.Fatal("select-all rejection changed accepted state", calls, selected)
	}
	after, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if after.Windows[0].FocusedID != w.FocusedID {
		t.Fatal("select all moved input focus")
	}
	root := m.implementation.find("one").state.root
	if root.control.choice.buttons["a"].IsChecked() || !root.control.choice.buttons["b"].IsChecked() || root.control.choice.buttons["c"].IsChecked() {
		t.Fatal("rejected select-all left native rows selected")
	}
}

func TestMultipleListRejectedSelectAllPreservesRangeAnchor(t *testing.T) {
	selected := []string{"a"}
	reject := false
	view := func() ui.Node {
		return ui.Node{ID: "list", Kind: ui.ListBox, Multiple: true, SelectedIDs: selected, Focus: ui.FocusRequest{ID: "list", Sequence: 1}, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Charlie"}}, SelectionChange: func(e ui.ListSelectionEvent) {
			if !reject {
				selected = e.Apply(selected)
			}
		}}
	}
	m, _ := simpleManual(t, manualConfig(t, 934), view)
	unitSettle(t, m)
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snap.Windows[0]
	send := func(code key.Codes, mods key.Modifiers) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Modifiers: mods}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	send(key.CodeDownArrow, 1<<key.Shift)
	if !slices.Equal(selected, []string{"a", "b"}) {
		t.Fatal("initial range", selected)
	}
	reject = true
	send(key.CodeA, 1<<key.Meta)
	if !slices.Equal(selected, []string{"a", "b"}) {
		t.Fatal("rejected all changed selection", selected)
	}
	reject = false
	send(key.CodeDownArrow, 1<<key.Shift)
	if !slices.Equal(selected, []string{"a", "b", "c"}) {
		t.Fatal("rejected select-all moved original range anchor", selected)
	}
}

func TestMultipleListCommandSelectAllIgnoresLockFlags(t *testing.T) {
	for _, mode := range []string{"caps", "num", "both"} {
		t.Run(mode, func(t *testing.T) {
			selected := []string{"b"}
			view := func() ui.Node {
				return ui.Node{ID: "list", Kind: ui.ListBox, Multiple: true, SelectedIDs: selected, Focus: ui.FocusRequest{ID: "list", Sequence: 1}, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, SelectionChange: func(e ui.ListSelectionEvent) { selected = e.Apply(selected) }}
			}
			m, _ := simpleManual(t, manualConfig(t, 935), view)
			unitSettle(t, m)
			snapshot, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			mods := key.Modifiers(1 << key.Meta)
			if mode == "caps" || mode == "both" {
				mods |= 1 << key.CapsLock
			}
			if mode == "num" || mode == "both" {
				mods |= 1 << key.NumLock
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.KeyChord, Code: key.CodeA, Rune: 'A', Modifiers: mods}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if !slices.Equal(selected, []string{"a", "b"}) {
				t.Fatal("lock flag defeated select-all", mode, selected)
			}
		})
	}
}
