//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestCascadingMenuKeyboardAndRetirement(t *testing.T) {
	branch, _ := ui.NewMenuBranch([]ui.Option{{ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "save", Label: "Save", Shortcut: "⌘S"}})
	calls := 0
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "menu", Sequence: 1}, Children: []ui.Node{{ID: "menu", Kind: ui.MenuButton, Text: "Actions", Options: []ui.Option{{ID: "export", Label: "Export", Branch: branch}}, Change: func(ui.ChangeEvent) { calls++ }}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 817), view)
	unitSettle(t, m)
	snap, _ := m.Snapshot(testContext)
	w := snap.Windows[0]
	input := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	input(key.CodeReturnEnter)
	r := findRecord(m.implementation.windows[0].state.root, "menu")
	p := r.control.menu.popup
	input(key.CodeRightArrow)
	if p.child == nil || p.child.dead.Load() || calls != 0 {
		t.Fatal("branch did not open without action")
	}
	_, handlers := menuAccessibility(r, "menu", nil)
	stale := handlers[menuActionID("menu", "save")]
	if stale == nil || handlers[menuActionID("menu", "blocked")] != nil {
		t.Fatal("submenu AX actions")
	}
	input(key.CodeLeftArrow)
	if !menuOpen(r) || !p.child.dead.Load() {
		t.Fatal("Left should close only submenu")
	}
	if stale(accessibilityAction{Kind: "press"}) {
		t.Fatal("closed submenu handler admitted")
	}
	input(key.CodeRightArrow)
	input(key.CodeDownArrow)
	input(key.CodeReturnEnter)
	if calls != 1 || menuOpen(r) {
		t.Fatal("leaf must commit once and close chain", calls)
	}
	input(key.CodeReturnEnter)
	input(key.CodeRightArrow)
	input(key.CodeEscape)
	if menuOpen(r) || r.control.menu.popup != nil || r.control.menu.closing != nil || calls != 1 {
		t.Fatal("Escape committed or leaked")
	}
}

func TestCascadingMenuMouseLeafExactlyOnce(t *testing.T) {
	branch, _ := ui.NewMenuBranch([]ui.Option{{ID: "leaf", Label: "Save image"}})
	calls := 0
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "menu", Sequence: 1}, Children: []ui.Node{{ID: "menu", Kind: ui.MenuButton, Text: "Actions", Options: []ui.Option{{ID: "branch", Label: "Export", Branch: branch}}, Change: func(ui.ChangeEvent) { calls++ }}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 818), view)
	unitSettle(t, m)
	snap, _ := m.Snapshot(testContext)
	w := snap.Windows[0]
	for _, code := range []key.Codes{key.CodeReturnEnter, key.CodeRightArrow} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	_, target := imageManualElement(t, m, menuActionID("menu", "leaf"))
	for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: kind, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	r := findRecord(m.implementation.windows[0].state.root, "menu")
	if calls != 1 || r.control.menu.popup != nil || r.control.menu.closing != nil {
		t.Fatal("mouse action lost, duplicated, or retained", calls)
	}
}
