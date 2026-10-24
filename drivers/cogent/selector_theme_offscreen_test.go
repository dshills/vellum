//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestScopedSelectorPopupPaletteDispatchAndRetirement(t *testing.T) {
	selected := "one"
	calls := 0
	dark := true
	options := []ui.Option{{ID: "one", Label: "One"}, {ID: "blocked", Label: "Unavailable", Disabled: true}, {ID: "two", Label: "Two"}}
	view := func() ui.Node {
		theme := ui.MacThemeWithAccent(dark, color.NRGBA{R: 126, G: 65, B: 194, A: 255})
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Theme: &theme, Focus: ui.FocusRequest{ID: "choose", Sequence: 1}, Children: []ui.Node{{ID: "choose", Kind: ui.Selector, Options: options, SelectedID: selected, Change: func(e ui.ChangeEvent) {
			if e.Kind != ui.Selector || e.Checked {
				t.Fatal("selector event changed kind", e)
			}
			calls++
			selected = e.SelectedID
		}}}}
	}
	m, a := simpleManual(t, manualConfig(t, 824), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	input := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	nodes, _ := accessibilitySnapshot(w.state.root, nil)
	declared := 0
	for _, n := range nodes {
		if n.Parent == accessibilityManagedID("choose") {
			declared++
		}
	}
	if declared != 3 {
		t.Fatal("closed selector lost direct choices", nodes)
	}
	input(key.CodeReturnEnter)
	r := findRecord(w.state.root, "choose")
	p := r.control.menu.popup
	nodes, _ = accessibilitySnapshot(w.state.root, nil)
	for _, n := range nodes {
		if n.ID == accessibilityOptionID("choose", "one") {
			t.Fatal("open popup duplicated closed choices")
		}
	}
	if p == nil {
		t.Fatal("themed popup not opened")
	}
	if p.scene.Styles.Background.At(0, 0) != r.theme.Palette.Value.Popup {
		t.Fatal("popup ignored palette")
	}
	if !p.items["blocked"].StateIs(states.Disabled) {
		t.Fatal("disabled item enabled")
	}
	_, handlers := menuAccessibility(r, "choose", nil)
	stale := handlers[menuActionID("choose", "two")]
	input(key.CodeDownArrow)
	input(key.CodeReturnEnter)
	if calls != 1 || selected != "two" || menuOpen(r) {
		t.Fatal(selected, calls)
	}
	input(key.CodeReturnEnter)
	if err := a.Do(func() { dark = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if menuOpen(r) || stale(accessibilityAction{Kind: "press"}) {
		t.Fatal("theme replacement retained popup action")
	}
	style, _ := ui.ResolveStyle(r.node, r.theme)
	if !style.Background.Set {
		t.Fatal("selector fill not explicit")
	}
}
