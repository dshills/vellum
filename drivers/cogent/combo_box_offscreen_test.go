//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestComboBoxRoutedSuggestionsAndDraftRetention(t *testing.T) {
	c := &ui.ComboBox{ID: "units", Label: "Units", Text: "mm", Options: []ui.Option{{ID: "mm", Label: "mm"}, {ID: "in", Label: "inches"}}}
	count := 0
	reject := false
	c.Change = func(e ui.ComboEvent) {
		count++
		if !reject {
			c.Text = e.Text
		}
	}
	view := func() ui.Node {
		n, err := c.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.ComboMenuID(c.ID), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 914), view)
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.ComboMenuID(c.ID))
	input := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	r := findRecord(m.implementation.find("one").state.root, ui.ComboInputID(c.ID))
	field := r.widget.(*core.TextField)
	field.SetText("draft")
	input(key.CodeDownArrow)
	menu := findRecord(m.implementation.find("one").state.root, ui.ComboMenuID(c.ID))
	if menu.widget.(*core.Button).Indicator != "" {
		t.Fatal("icon-only combo contains duplicate disclosure indicator")
	}
	if !menuOpen(menu) {
		t.Fatal("suggestions did not open")
	}
	if field.Text() != "draft" {
		t.Fatal("opening menu destroyed draft")
	}
	// Route to the actual popup row, not a direct component callback.
	_, item := imageManualElement(t, m, menuActionID(ui.ComboMenuID(c.ID), "in"))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: item.ID, TargetIncarnation: item.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if c.Text != "inches" || count != 1 || field.Text() != "inches" || menuOpen(menu) {
		t.Fatal("selection", c.Text, count, field.Text())
	}
	// Native text commit reuses the managed field and current owner callback.
	reject = true
	field.SetText("custom").SendChange()
	a.Drain()
	unitSettle(t, m)
	if c.Text != "inches" || field.Text() != "inches" || count != 2 {
		t.Fatal("rejected typing", c.Text, field.Text(), count)
	}
	if findRecord(m.implementation.find("one").state.root, r.node.ID).widget != field {
		t.Fatal("input recreated")
	}
}
