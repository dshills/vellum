//go:build offscreen

package cogent

import (
	"bytes"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestIconsManualArtworkInputAndReplacement(t *testing.T) {
	selected, appearance := ui.SymbolSave, ui.ButtonPrimary
	count, checked := 0, false
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "action", Sequence: 1}, Style: ui.Style{Gap: ui.Some(float32(8))}, Children: []ui.Node{
			{ID: "symbol", Kind: ui.Icon, Icon: selected, Accessibility: ui.Accessibility{Name: "Selected artwork"}},
			{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{
				{ID: "action", Kind: ui.Button, Icon: selected, Appearance: appearance, Accessibility: ui.Accessibility{Name: "Run action"}, Activate: func() { count++ }},
				{ID: "toggle", Kind: ui.ToggleButton, Icon: ui.SymbolCheck, Checked: checked, Accessibility: ui.Accessibility{Name: "Guides"}, Change: func(e ui.ChangeEvent) { checked = e.Checked }},
			}},
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snapshot.Windows[0]
	if w.FocusedID != "action" {
		t.Fatal("icon action did not accept managed focus", w.FocusedID)
	}
	input := func(typ events.Types, code key.Codes, target string) {
		t.Helper()
		i := ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Code: code, Target: target}
		if target != "" {
			for _, e := range w.Elements {
				if e.ID == target {
					i.TargetIncarnation = e.Incarnation
				}
			}
			i.Button = events.Left
		}
		if err := m.Input(testContext, i); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	input(events.KeyChord, key.CodeReturnEnter, "")
	if count != 1 {
		t.Fatal("actual icon-button keyboard route", count)
	}
	input(events.KeyChord, key.CodeTab, "")
	input(events.KeyChord, key.CodeReturnEnter, "")
	if !checked {
		t.Fatal("actual icon-toggle keyboard route")
	}
	input(events.MouseDown, 0, "action")
	input(events.MouseUp, 0, "action")
	if count != 2 {
		t.Fatal("actual icon-button pointer route", count)
	}
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Post(func() { appearance = ui.ButtonSecondary; a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	after, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before.Pix, after.Pix) {
		t.Fatal("appearance replacement did not change completed pixels")
	}
	var incarnation uint64
	for _, symbol := range ui.Symbols() {
		if err := a.Post(func() { selected = symbol; a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		snapshot, err = m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		var target ManualElement
		for _, e := range snapshot.Windows[0].Elements {
			if e.ID == "symbol" {
				target = e
			}
		}
		if target.Bounds.Empty() || target.Kind != ui.Icon || !target.Visible {
			t.Fatal("symbol lacks measured geometry", symbol, target)
		}
		if incarnation == 0 {
			incarnation = target.Incarnation
		} else if incarnation != target.Incarnation {
			t.Fatal("artwork replacement retired the element", symbol)
		}
		frame, err := m.Capture(testContext, "one")
		if err != nil {
			t.Fatal(err)
		}
		first := frame.RGBAAt(target.Bounds.Min.X, target.Bounds.Min.Y)
		painted := false
		for y := target.Bounds.Min.Y; y < target.Bounds.Max.Y; y++ {
			for x := target.Bounds.Min.X; x < target.Bounds.Max.X; x++ {
				painted = painted || frame.RGBAAt(x, y) != first
			}
		}
		if !painted {
			t.Fatal("bundled symbol produced no painted artwork", symbol)
		}
	}
	if !checked || count != 2 {
		t.Fatal("presentation changes altered accepted action/toggle state")
	}
}
