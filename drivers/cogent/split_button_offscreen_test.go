//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestSplitButtonNativeKeyboardIsolation(t *testing.T) {
	primary, secondary := 0, 0
	joined := true
	b := &ui.SplitButton{ID: "split", Label: "Save", Options: []ui.Option{{ID: "as", Label: "Save As"}}, Primary: func() { primary++ }, Change: func(ui.ChangeEvent) { secondary++ }}
	view := func() ui.Node {
		n, err := b.View()
		if err != nil {
			t.Fatal(err)
		}
		n.SplitButtonLayout = joined
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.SplitPrimaryID("split"), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 819), view)
	unitSettle(t, m)
	r := findRecord(m.implementation.find("one").state.root, "split")
	primaryWidget := findRecord(m.implementation.find("one").state.root, ui.SplitPrimaryID("split")).widget.AsWidget()
	menuWidget := findRecord(m.implementation.find("one").state.root, ui.SplitMenuID("split")).widget.AsWidget()
	if primaryWidget.Styles.Margin.Left.Value != 0 || menuWidget.Styles.Margin.Left.Value != 0 || r.widget.AsWidget().Styles.Padding.Left.Value != 0 || menuWidget.Styles.Border.Width.Left.Value != 1 {
		t.Fatal("native margins separated split segments")
	}
	snap, _ := m.Snapshot(testContext)
	w := snap.Windows[0]
	for _, code := range []key.Codes{key.CodeReturnEnter, key.CodeTab, key.CodeReturnEnter, key.CodeReturnEnter} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if !menuWidget.ContainsFocus() || menuWidget.Styles.Border.Width.Top.Value < 2 || menuWidget.Styles.Border.Width.Left.Value != menuWidget.Styles.Border.Width.Top.Value {
		t.Fatal("menu segment lost full keyboard focus ring")
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeTab}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if menuWidget.ContainsFocus() {
		t.Fatal("test did not leave menu segment focus")
	}
	if err := a.Do(func() { joined = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if menuWidget.Styles.Border.Width.Left.Value != 0 {
		t.Fatal("removed layout kept stale divider", menuWidget.Styles.Border.Width.Left.Value)
	}
	if err := a.Do(func() { joined = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if menuWidget.Styles.Border.Width.Left.Value != 1 {
		t.Fatal("retained layout did not restore divider")
	}
	if primary != 1 || secondary != 1 {
		t.Fatal("primary/disclosure crossed paths", primary, secondary)
	}
}
