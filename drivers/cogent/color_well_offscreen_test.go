//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestColorWellManagedPanelAndAcceptedColor(t *testing.T) {
	p := &ui.ColorWell{ID: "well", Label: "Ink", Color: color.NRGBA{R: 25, A: 255}}
	calls := 0
	p.Change = func(e ui.ColorEdit) { calls++; p.Color = e.Apply(p.Color) }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.ColorWellButtonID(p.ID), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 921), view)
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.ColorWellButtonID(p.ID))
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	panel := findRecord(root, "well/color-well/popup")
	if !panel.node.Open {
		t.Fatal("panel did not open")
	}
	field := findRecord(root, ui.ColorFieldID(ui.ColorWellPickerID(p.ID), ui.ColorHex)).widget.(*core.TextField)
	field.SetText("#12345680").SendChange()
	a.Drain()
	unitSettle(t, m)
	if calls != 1 || p.Color != (color.NRGBA{R: 18, G: 52, B: 86, A: 128}) {
		t.Fatal(calls, p.Color)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeEscape}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	root = m.implementation.find("one").state.root
	if findRecord(root, "well/color-well/popup").node.Open {
		t.Fatal("Escape failed")
	}
}

func TestColorWellManagedCancellationAndReopenCycles(t *testing.T) {
	p := &ui.ColorWell{ID: "well", Label: "Ink", Color: color.NRGBA{R: 25, A: 255}, Change: func(ui.ColorEdit) {}}
	focus := ui.FocusRequest{ID: ui.ColorWellButtonID(p.ID), Sequence: 1}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: focus, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 932), view)
	unitSettle(t, m)
	send := func(code key.Codes) {
		t.Helper()
		w, _ := imageManualElement(t, m, ui.ColorWellButtonID(p.ID))
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	for _, mode := range []string{"identity", "disabled", "read-only"} {
		send(key.CodeReturnEnter)
		root := m.implementation.find("one").state.root
		panel := findRecord(root, p.ID+"/color-well/popup")
		if !panel.node.Open || !panel.control.dialog.active {
			t.Fatal("panel did not open", mode)
		}
		if err := a.Do(func() {
			switch mode {
			case "identity":
				p.ID = "replacement"
				focus = ui.FocusRequest{ID: ui.ColorWellButtonID(p.ID), Sequence: 2}
			case "disabled":
				p.Disabled = true
			case "read-only":
				p.Change = nil
			}
			_ = a.Invalidate("one")
		}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		root = m.implementation.find("one").state.root
		panel = findRecord(root, p.ID+"/color-well/popup")
		if panel.node.Open || panel.control.dialog.active {
			t.Fatal("cancellation did not close panel", mode)
		}
		if err := a.Do(func() {
			p.Disabled = false
			p.Change = func(ui.ColorEdit) {}
			focus.Sequence++
			_ = a.Invalidate("one")
		}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		if findRecord(m.implementation.find("one").state.root, p.ID+"/color-well/popup").node.Open {
			t.Fatal("panel reopened without activation", mode)
		}
	}
	// Repeated actual keyboard open/cancel must preserve the keyed editor while
	// dropping each transient popup focus scope and closing native dialog state.
	var editor core.Widget
	for cycle := 0; cycle < 12; cycle++ {
		send(key.CodeReturnEnter)
		root := m.implementation.find("one").state.root
		field := findRecord(root, ui.ColorFieldID(ui.ColorWellPickerID(p.ID), ui.ColorHex)).widget
		if editor != nil && field != editor {
			t.Fatal("retained editor replaced", cycle)
		}
		editor = field
		send(key.CodeEscape)
		root = m.implementation.find("one").state.root
		panel := findRecord(root, p.ID+"/color-well/popup")
		if panel.control.dialog.active || panel.node.Open || panel.control.dialog.body != nil {
			t.Fatal("transient dialog retained", cycle)
		}
		w, _ := imageManualElement(t, m, ui.ColorWellButtonID(p.ID))
		if w.FocusedID != ui.ColorWellButtonID(p.ID) {
			t.Fatal("focus not restored", cycle, w.FocusedID)
		}
	}
}
