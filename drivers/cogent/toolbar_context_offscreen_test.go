//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestToolbarContextRoutesChildSpaceKeyboardAndRetirement(t *testing.T) {
	calls, clicks := 0, 0
	disabled, own := false, false
	view := func() ui.Node {
		child := ui.Node{ID: "action", Kind: ui.Button, Text: "Action", Activate: func() { clicks++ }}
		if own {
			child.ContextMenu = &ui.ContextMenu{Options: []ui.Option{{ID: "own", Label: "Own action"}}, Change: func(ui.ChangeEvent) {}}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "action", Sequence: 1}, Children: []ui.Node{{ID: "tools", Kind: ui.Toolbar, Disabled: disabled, ContextMenu: &ui.ContextMenu{Options: []ui.Option{{ID: "icons", Label: "Icons Only", Checkable: true, Checked: true}, {ID: "customize", Label: "Customize Toolbar…"}}, Change: func(ui.ChangeEvent) { calls++ }}, Children: []ui.Node{child, {ID: "space", Kind: ui.ToolbarFlexibleSpace}}}}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	root := w.state.root
	bar := findRecord(root, "tools")
	child := findRecord(root, "action")
	space := findRecord(root, "space")
	child.widget.AsWidget().Send(events.ContextMenu, events.NewMouse(events.MouseUp, events.Right, child.widget.AsWidget().Geom.TotalBBox.Min, 0))
	unitSettle(t, m)
	if !menuOpen(bar) || menuOpen(child) || clicks != 0 {
		t.Fatal("child did not route to toolbar context")
	}
	_, target := imageManualElement(t, m, menuActionID("tools", "customize"))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Target: target.ID, TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}); e != nil {
			t.Fatal(e)
		}
	}
	unitSettle(t, m)
	if !child.widget.AsWidget().ContainsFocus() {
		t.Fatal("context menu lost focused toolbar child")
	}
	if calls != 1 || clicks != 0 || menuOpen(bar) {
		t.Fatal("context command/click count", calls, clicks)
	}
	child.widget.AsWidget().Send(events.KeyChord, events.NewKey(events.KeyChord, 0, key.CodeF10, 1<<key.Shift|1<<key.NumLock))
	unitSettle(t, m)
	if !menuOpen(bar) {
		t.Fatal("Shift F10 with lock flags did not open toolbar context")
	}
	_, handlers := accessibilitySnapshot(root, nil)
	old := handlers[menuActionID("tools", "icons")]
	disabled = true
	a.Invalidate("one")
	unitSettle(t, m)
	if menuOpen(bar) || old == nil || old(accessibilityAction{Kind: "press"}) {
		t.Fatal("disabled toolbar retained menu action")
	}
	disabled = false
	a.Invalidate("one")
	unitSettle(t, m)
	space.widget.AsWidget().Send(events.ContextMenu, events.NewMouse(events.MouseUp, events.Right, space.widget.AsWidget().Geom.TotalBBox.Min, 0))
	unitSettle(t, m)
	if !menuOpen(bar) {
		t.Fatal("space did not route to toolbar context")
	}
	requestMenuClose(bar)
	a.Invalidate("one")
	unitSettle(t, m)
	own = true
	a.Invalidate("one")
	unitSettle(t, m)
	child.widget.AsWidget().Send(events.ContextMenu, events.NewMouse(events.MouseUp, events.Right, child.widget.AsWidget().Geom.TotalBBox.Min, 0))
	unitSettle(t, m)
	if !menuOpen(child) || menuOpen(bar) {
		t.Fatal("toolbar stole explicit child context")
	}
	if e := m.Close(testContext); e != nil {
		t.Fatal(e)
	}
	if menuOpen(child) || menuOpen(bar) {
		t.Fatal("shutdown retained context popup")
	}
}
