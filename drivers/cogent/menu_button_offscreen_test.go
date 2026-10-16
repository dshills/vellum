//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMenuButtonRealPopupRoutingAndRetirement(t *testing.T) {
	var picked []string
	options := []ui.Option{{ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "one", Label: "Same"}, {ID: "two", Label: "Same"}}
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "actions", Sequence: 1}, Children: []ui.Node{{ID: "actions", Kind: ui.MenuButton, Text: "Actions <literal>", Options: options, Change: func(e ui.ChangeEvent) { picked = append(picked, e.SelectedID) }}, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 811), view)
	unitSettle(t, m)
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snap.Windows[0]
	input := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	input(key.CodeReturnEnter)
	root := m.implementation.windows[0].state.root
	r := findRecord(root, "actions")
	if !menuOpen(r) {
		t.Fatal("native key did not open popup")
	}
	p := r.control.menu.popup
	nodes, handlers := accessibilitySnapshot(root, nil)
	if len(nodes) != 4 || nodes[0].Role != "menu-button" {
		t.Fatal("popup did not isolate AX scope", nodes)
	}
	if handlers[menuActionID("actions", "blocked")] != nil {
		t.Fatal("disabled menu action exposed")
	}
	bounds := accessibilityBounds(root)
	b := bounds[menuActionID("actions", "two")]
	if b[2] <= 0 || b[3] <= 0 {
		t.Fatal("actual popup item lacks geometry", b)
	}
	snap, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	var target ManualElement
	for _, e := range snap.Windows[0].Elements {
		if e.ID == menuActionID("actions", "two") {
			target = e
		}
	}
	if target.Incarnation == 0 || target.Bounds.Empty() {
		t.Fatal("public snapshot lacks menu target", target)
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if len(picked) != 1 || picked[0] != "two" || menuOpen(r) {
		t.Fatal("real duplicate-label item did not choose stable action/close", picked, menuOpen(r))
	}
	oldHandler := handlers[menuActionID("actions", "one")]
	input(key.CodeReturnEnter)
	if !menuOpen(r) || r.control.menu.popup == p {
		t.Fatal("popup did not create new lifetime")
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.MouseDown, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); !errors.Is(err, ErrManualStale) {
		t.Fatal("previous popup target admitted", err)
	}
	if oldHandler(accessibilityAction{Kind: "press"}) {
		t.Fatal("retired menu handler admitted")
	}
	input(key.CodeTab)
	if menuOpen(r) || accessibilityFocusedID(root) != accessibilityManagedID("after") {
		t.Fatal("popup Tab did not advance", accessibilityFocusedID(root))
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeTab, Modifiers: key.Shift}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	input(key.CodeReturnEnter)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeTab, Modifiers: key.Shift}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if menuOpen(r) || accessibilityFocusedID(root) != accessibilityManagedID("after") {
		t.Fatal("popup Shift-Tab did not advance backward", accessibilityFocusedID(root))
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeTab, Modifiers: key.Shift}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	input(key.CodeReturnEnter)
	input(key.CodeEscape)
	if menuOpen(r) {
		t.Fatal("Escape did not retire popup")
	}
	input(key.CodeReturnEnter)
	p = r.control.menu.popup
	queueMenuAction(r, p, "one")
	options = append([]ui.Option(nil), options...)
	options[1].Disabled = true
	if err := m.implementation.owner.Do(func() {
		var e error
		root, e = reconcile(m.implementation.windows[0].state.host, root, view(), m.implementation.owner, "one")
		if e != nil {
			t.Fatal(e)
		}
		m.implementation.windows[0].state.root = root
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if len(picked) != 1 {
		t.Fatal("replaced option admitted queued old action", picked)
	}
}

func TestMenuButtonPeerDrainDefersNativeCloseAndModalWake(t *testing.T) {
	config := manualConfig(t, 811)
	disabled, calls := false, 0
	view := func() ui.Node {
		return ui.Node{ID: "actions", Kind: ui.MenuButton, Text: "Actions", Disabled: disabled, Options: []ui.Option{{ID: "run", Label: "Run"}}, Change: func(ui.ChangeEvent) { calls++ }}
	}
	m, a := simpleManual(t, config, view)
	if err := a.Open(app.Window{ID: "peer", Title: "Peer", Width: 200, Height: 120, View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label, Text: "Peer"} }}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	origin := m.implementation.find("one")
	peer := m.implementation.find("peer")
	r := origin.state.root
	if err := a.Do(func() { r.control.menu.request = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	p := r.control.menu.popup
	if !menuOpen(r) {
		t.Fatal("missing popup")
	}
	_, handlers := accessibilitySnapshot(r, nil)
	old := handlers[menuActionID("actions", "run")]
	queueMenuAction(r, p, "run")
	wake := &events.CustomEvent{}
	wake.Typ = events.Custom
	wake.Data = "vellum.wake"
	m.implementation.ownerBudget = 811
	m.implementation.custom(peer.state, wake)
	m.implementation.ownerBudget = 0
	if calls != 1 || menuOpen(r) || old(accessibilityAction{Kind: "press"}) {
		t.Fatal("logical retirement did not reject stale callback", calls)
	}
	if p.stage == nil || p.scene == nil || p.stage.Scene != p.scene || len(p.items) != 1 || p.context.Property("vellum.menu.active") != p {
		t.Fatal("peer drain destroyed originating native scene")
	}
	if err := m.implementation.refresh(origin); err != nil {
		t.Fatal(err)
	}
	if p.stage != nil || p.scene != nil || p.context != nil || r.control.menu.closing != nil {
		t.Fatal("origin refresh did not finish native retirement")
	}
	unitSettle(t, m)
	if err := a.Do(func() { r.control.menu.request = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	p = r.control.menu.popup
	if !menuOpen(r) {
		t.Fatal("missing second popup")
	}
	posted := false
	if err := a.PostWindow(fixtureWindowRef(a, "one"), func() { posted = true; disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	// Route through the actual modal scene, with no Manual.Step fallback drain.
	m.implementation.ownerBudget = 811
	modalWake := &events.CustomEvent{}
	modalWake.Typ = events.Custom
	modalWake.Data = "vellum.wake"
	p.scene.HandleEvent(modalWake)
	m.implementation.ownerBudget = 0
	if !posted {
		t.Fatal("modal custom wake stranded owner callback")
	}
	if err := m.implementation.refresh(origin); err != nil {
		t.Fatal(err)
	}
	if menuOpen(r) || p.stage != nil {
		t.Fatal("owner update did not close disabled popup")
	}
}
