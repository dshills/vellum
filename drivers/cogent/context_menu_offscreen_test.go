//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestContextMenuRoutedPointerKeyboardAndRetirement(t *testing.T) {
	calls, clicks := 0, 0
	show, disabled := true, false
	options := []ui.Option{{ID: "action", Label: "Inspect"}, {ID: "blocked", Label: "Blocked", Disabled: true}}
	view := func() ui.Node {
		n := ui.Node{ID: "host", Kind: ui.Button, Text: "Object", Disabled: disabled, Activate: func() { clicks++ }}
		if show {
			n.ContextMenu = &ui.ContextMenu{Options: options, Change: func(ui.ChangeEvent) { calls++ }}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "host", Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 915), view)
	unitSettle(t, m)
	w, host := imageManualElement(t, m, "host")
	click := func(target ManualElement, button events.Buttons) {
		t.Helper()
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: button}); err != nil {
				t.Fatal(err)
			}
		}
		unitSettle(t, m)
	}
	r := findRecord(m.implementation.find("one").state.root, "host")
	r.widget.AsWidget().Send(events.ContextMenu, events.NewMouse(events.MouseUp, events.Right, host.Bounds.Min.Add(host.Bounds.Size().Div(2)), 0))
	unitSettle(t, m)
	if !menuOpen(r) || clicks != 0 {
		t.Fatal("right-click activated host or no menu", clicks)
	}
	_, item := imageManualElement(t, m, menuActionID("host", "action"))
	click(item, events.Left)
	if calls != 1 || clicks != 0 || menuOpen(r) {
		t.Fatal("context action", calls, clicks)
	}
	// The public manual harness intentionally refuses right/Control-click and system
	// Menu chords. Open through actual native widget events; popup row selection
	// above is routed through the bounded manual harness. Reconciliation supplies a fresh ContextMenu value each time. That must not
	// reopen or invalidate an otherwise unchanged popup solely by pointer identity.
	r.widget.AsWidget().Send(events.KeyChord, events.NewKey(events.KeyChord, 0, key.CodeF10, 1<<key.Shift))
	unitSettle(t, m)
	if !menuOpen(r) {
		t.Fatal("keyboard menu did not open")
	}
	_, handlers := accessibilitySnapshot(m.implementation.find("one").state.root, nil)
	old := handlers[menuActionID("host", "action")]
	if old == nil {
		t.Fatal("menu AX absent")
	}
	if err := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if menuOpen(r) || old(accessibilityAction{Kind: "press"}) {
		t.Fatal("disabled menu retained action")
	}
	if err := a.Do(func() { disabled = false; show = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	_, host = imageManualElement(t, m, "host")
	click(host, events.Left)
	if clicks != 1 {
		t.Fatal("removing context menu disabled host")
	}
}

func TestContextMenuListRowConsumesSharedEventOnce(t *testing.T) {
	view := func() ui.Node {
		return ui.Node{ID: "list", Kind: ui.ListBox, Options: []ui.Option{{ID: "a", Label: "A"}}, ContextMenu: &ui.ContextMenu{Options: []ui.Option{{ID: "inspect", Label: "Inspect"}}, Change: func(ui.ChangeEvent) {}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 916), view)
	unitSettle(t, m)
	r := m.implementation.find("one").state.root
	e := events.NewMouse(events.ContextMenu, events.Right, r.widget.AsWidget().Geom.TotalBBox.Min, 0)
	row := r.control.choice.buttons["a"]
	row.HandleEvent(e)
	r.widget.AsWidget().HandleEvent(e)
	unitSettle(t, m)
	if !e.IsHandled() || !menuOpen(r) {
		t.Fatal("row context request missing")
	}
	popup := r.control.menu.popup
	unitSettle(t, m)
	if r.control.menu.popup != popup {
		t.Fatal("duplicate row/host request replaced popup")
	}
}
