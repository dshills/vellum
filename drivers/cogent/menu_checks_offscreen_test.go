//go:build offscreen

package cogent

import (
	"bytes"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestMenuCheckActionsRealRoutingAndControlledAcceptance(t *testing.T) {
	checked, reject := false, false
	var proposals []ui.ChangeEvent
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "view", Sequence: 1}, Children: []ui.Node{
			{ID: "view", Kind: ui.MenuButton, Text: "View", Options: []ui.Option{
				{ID: "toggle", Label: "Show <grid>", Checkable: true, Checked: checked},
				{ID: "divider", Separator: true},
				{ID: "blocked", Label: "Blocked", Checkable: true, Checked: true, Disabled: true},
				{ID: "reset", Label: "Reset"},
			}, Change: func(e ui.ChangeEvent) {
				proposals = append(proposals, e)
				if !reject && e.SelectedID == "toggle" {
					checked = e.Checked
				}
			}},
			{ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}},
		}}
	}
	m, _ := simpleManual(t, manualConfig(t, 911), view)
	unitSettle(t, m)
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	window := snapshot.Windows[0]
	input := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	root := m.implementation.windows[0].state.root
	r := findRecord(root, "view")
	open := func() *menuPopup {
		t.Helper()
		input(key.CodeReturnEnter)
		if !menuOpen(r) {
			t.Fatal("menu did not open")
		}
		return r.control.menu.popup
	}
	p := open()
	if len(p.items) != 3 {
		t.Fatal("separator became an action", len(p.items))
	}
	if _, ok := p.scene.AsTree().ChildByName("divider").(*core.Separator); !ok {
		t.Fatal("missing real menu separator")
	}
	nodes, handlers := menuAccessibility(r, "", nil)
	if len(nodes) != 3 || nodes[0].Value != false || nodes[1].Value != true || handlers[menuActionID("view", "blocked")] != nil || handlers[menuActionID("view", "divider")] != nil {
		t.Fatal("check/disabled/separator semantics", nodes)
	}
	snapshot, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range snapshot.Windows[0].Elements {
		if e.ID == menuActionID("view", "divider") {
			t.Fatal("separator exposed as input target")
		}
		if e.ID == menuActionID("view", "toggle") && e.Value != "false" {
			t.Fatal("manual accepted check value", e)
		}
	}
	unchecked, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	input(key.CodeDownArrow)
	if !p.items["reset"].ContainsFocus() {
		t.Fatal("keyboard did not skip separator/disabled check")
	}
	input(key.CodeReturnEnter)
	if len(proposals) != 1 || proposals[0].SelectedID != "reset" || proposals[0].Checked || checked {
		t.Fatal("ordinary action changed check state", proposals)
	}
	p = open()
	input(key.CodeReturnEnter)
	if len(proposals) != 2 || proposals[1].ID != "view" || proposals[1].Kind != ui.MenuButton || proposals[1].SelectedID != "toggle" || !proposals[1].Checked || !checked || menuOpen(r) {
		t.Fatal("check proposal not accepted once", proposals, checked)
	}
	p = open()
	if p.items["toggle"].Icon != nativeSymbol(ui.SymbolCheck) {
		t.Fatal("accepted check marker absent")
	}
	checkedPixels, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(unchecked.Pix, checkedPixels.Pix) {
		t.Fatal("check marker did not change composition")
	}
	reject = true
	input(key.CodeReturnEnter)
	if len(proposals) != 3 || proposals[2].Checked || !checked {
		t.Fatal("rejected uncheck changed accepted state", proposals, checked)
	}
	p = open()
	if p.items["toggle"].Icon != nativeSymbol(ui.SymbolCheck) {
		t.Fatal("rejection lost check marker")
	}
	nodes, _ = menuAccessibility(r, "", nil)
	if nodes[0].Value != true {
		t.Fatal("rejection lost accepted AX value")
	}
	// Even direct requests cannot turn a passive divider into a callback.
	queueMenuAction(r, p, "divider")
	unitSettle(t, m)
	if len(proposals) != 3 || !menuOpen(r) {
		t.Fatal("separator admitted a request")
	}
	// Pointer delivery uses the same current controlled proposal as keyboard.
	reject = false
	snapshot, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	var target ManualElement
	for _, e := range snapshot.Windows[0].Elements {
		if e.ID == menuActionID("view", "toggle") {
			target = e
		}
	}
	if target.Incarnation == 0 {
		t.Fatal("missing pointer check target")
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if len(proposals) != 4 || proposals[3].Checked || checked || menuOpen(r) {
		t.Fatal("pointer uncheck did not commit once", proposals, checked)
	}
	p = open()
	if p.items["toggle"].Icon != "" {
		t.Fatal("unchecked state retained marker")
	}
	// Descriptor change revokes an already queued check proposal.
	queueMenuAction(r, p, "toggle")
	checked = true
	if err := m.implementation.owner.Do(func() {
		var err error
		root, err = reconcile(m.implementation.windows[0].state.host, root, view(), m.implementation.owner, "one")
		if err != nil {
			t.Fatal(err)
		}
		m.implementation.windows[0].state.root = root
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if len(proposals) != 4 || menuOpen(r) {
		t.Fatal("changed accepted value did not revoke queued check")
	}
}
