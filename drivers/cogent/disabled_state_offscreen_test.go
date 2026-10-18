//go:build offscreen

package cogent

import (
	"testing"

	"github.com/dshills/vellum/ui"
)

func TestAncestorReenablePreservesOwnDisabledAndPassiveStyle(t *testing.T) {
	disabled := false
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Disabled: disabled, Children: []ui.Node{
			{ID: "passive", Kind: ui.Label, Text: "Status"},
			{ID: "locked", Kind: ui.NumberField, Disabled: true, Number: 25, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}},
		}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	passive, locked := findRecord(w.state.root, "passive").widget, findRecord(w.state.root, "locked").widget
	initial := passive.AsWidget().Styles.Opacity
	for cycle := range 3 {
		for _, v := range []bool{true, false} {
			if err := a.Do(func() { disabled = v; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			p, l := findRecord(w.state.root, "passive").widget, findRecord(w.state.root, "locked").widget
			if p != passive || l != locked {
				t.Fatal("ancestor transition replaced child identity")
			}
			if p.AsWidget().IsDisabled() != v {
				t.Fatal("passive effective state", cycle, v)
			}
			if !l.AsWidget().IsDisabled() {
				t.Fatal("ancestor re-enable cleared own Disabled")
			}
			opacity := p.AsWidget().Styles.Opacity
			if v && opacity >= initial || !v && opacity != initial {
				t.Fatal("passive disabled styling not refreshed", cycle, v, opacity, initial)
			}
		}
	}
}
