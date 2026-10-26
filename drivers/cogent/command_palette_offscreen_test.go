//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestCommandPaletteLiveSearchExactlyOnceAndFocusReturn(t *testing.T) {
	calls := 0
	p := &ui.CommandPalette{ID: "palette", Open: true, ReturnFocusID: "open", Items: []ui.CommandPaletteItem{{ID: "alpha", Label: "Alpha"}, {ID: "beta", Label: "Beta"}}, Invoke: func(e ui.CommandPaletteEvent) error {
		if e.CommandID != "beta" {
			t.Fatal(e)
		}
		calls++
		return nil
	}}
	m, a := simpleManual(t, manualConfig(t, 1702), func() ui.Node {
		n, e := p.View()
		if e != nil {
			t.Fatal(e)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "open", Kind: ui.Button, Text: "Search Commands", Activate: func() { p.Open = true }}, n}}
	})
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.CommandPaletteInputID(p.ID))
	send := func(code key.Codes, char rune) {
		t.Helper()
		if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Rune: char}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
	}
	send(key.CodeB, 'b')
	if p.Query != "b" {
		t.Fatal("search not live", p.Query)
	}
	send(key.CodeReturnEnter, 0)
	if calls != 1 || p.Open {
		t.Fatal("Return did not execute exactly once", calls, p.Open)
	}
	_, opener := imageManualElement(t, m, "open")
	if !opener.Focused {
		t.Fatal("parent button focus not restored")
	}
	if e := a.Do(func() { p.Open = true; p.Query = ""; a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	send(key.CodeDownArrow, 0)
	if p.SelectedID != "beta" {
		t.Fatal("Down did not move selection", p.SelectedID)
	}
	send(key.CodeEscape, 0)
	if p.Open || calls != 1 {
		t.Fatal("Escape executed an action", p.Open, calls)
	}
}
