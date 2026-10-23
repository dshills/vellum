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
