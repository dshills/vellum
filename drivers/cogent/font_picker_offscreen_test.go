//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestFontPickerManagedSizeAndFamily(t *testing.T) {
	p := &ui.FontPicker{ID: "font", Value: ui.FontSelection{FamilyID: "helv", TypefaceID: "regular", Size: 16}, Families: []ui.FontFamily{{ID: "helv", Name: "Helvetica Neue", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}, {ID: "times", Name: "Times New Roman", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}}}
	calls := 0
	p.Change = func(e ui.FontSelectionEvent) { calls++; p.Value = e.Apply(p.Value) }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.FontSizeID(p.ID), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 923), view)
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.FontSizeID(p.ID))
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeUpArrow}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 1 || p.Value.Size != 17 {
		t.Fatal(calls, p.Value)
	}
	field := findRecord(m.implementation.find("one").state.root, ui.ComboInputID(ui.FontFamilyID(p.ID))).widget.(*core.TextField)
	field.SetText("Times New Roman").SendChange()
	a.Drain()
	unitSettle(t, m)
	if calls != 2 || p.Value.FamilyID != "times" || p.Value.Size != 17 {
		t.Fatal(calls, p.Value)
	}
	preview := findRecord(m.implementation.find("one").state.root, ui.FontPreviewID(p.ID))
	if preview.node.Style.Typography.Value.Family != "Times New Roman" {
		t.Fatal("preview font")
	}
}
