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
	b := &ui.SplitButton{ID: "split", Label: "Save", Options: []ui.Option{{ID: "as", Label: "Save As"}}, Primary: func() { primary++ }, Change: func(ui.ChangeEvent) { secondary++ }}
	view := func() ui.Node {
		n, err := b.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.SplitPrimaryID("split"), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 819), view)
	unitSettle(t, m)
	snap, _ := m.Snapshot(testContext)
	w := snap.Windows[0]
	for _, code := range []key.Codes{key.CodeReturnEnter, key.CodeTab, key.CodeReturnEnter, key.CodeReturnEnter} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if primary != 1 || secondary != 1 {
		t.Fatal("primary/disclosure crossed paths", primary, secondary)
	}
}
