//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestIconPickerNativePopoverKeyboardAcceptance(t *testing.T) {
	source := fullColorButtonSource(t, false)
	p := &ui.IconPicker{ID: "picker", Label: "Artwork", Items: []ui.IconPickerItem{{ID: "color", Label: "Color icon", Image: source}}}
	calls := 0
	p.Change = func(e ui.IconPickerEvent) { calls++; p.SelectedID = e.ItemID }
	focus := ui.FocusRequest{ID: ui.IconPickerButtonID(p.ID), Sequence: 1}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 826), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	input := func() {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	input()
	if !p.Open {
		t.Fatal("button did not open picker")
	}
	if err := a.Do(func() {
		focus = ui.FocusRequest{ID: ui.IconPickerCollectionID(p.ID), Sequence: 2}
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	input()
	if calls != 1 || p.SelectedID != "color" || p.Open {
		t.Fatal("grid Return did not choose once", calls, p.SelectedID)
	}
}
