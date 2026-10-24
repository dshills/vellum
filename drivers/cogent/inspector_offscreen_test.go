//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestInspectorNativeDisclosureRetainsEditor(t *testing.T) {
	p := &ui.Inspector{ID: "inspector", Title: "Properties", Sections: []ui.InspectorSection{{ID: "geometry", Title: "Geometry", Expanded: true, Properties: []ui.InspectorProperty{{ID: "name", Label: "Name", Mixed: true, Editor: ui.Node{ID: "name", Kind: ui.TextField, Text: "Preserved"}}}}}}
	p.Disclosure = func(e ui.InspectorDisclosureEvent) { p.Sections[0].Expanded = e.Expanded }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.InspectorSectionID(p.ID, "geometry"), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 820), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	field := findRecord(w.state.root, "name").widget
	for i, code := range []key.Codes{key.CodeLeftArrow, key.CodeRightArrow} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		if field.AsWidget().IsVisible() != (i == 1) {
			t.Fatal("disclosure visibility")
		}
	}
	if findRecord(w.state.root, "name").widget != field {
		t.Fatal("editor recreated")
	}
}
