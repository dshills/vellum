//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestTabsActualControlledKeyboardNavigation(t *testing.T) {
	selected := "a"
	calls := 0
	reject, readonly, disabled := false, false, false
	view := func() ui.Node {
		n := ui.Node{ID: "tabs", Kind: ui.Tabs, Grow: true, SelectedID: selected, Disabled: disabled, Focus: ui.FocusRequest{ID: "tabs", Sequence: 1}, Options: []ui.Option{{ID: "a", Label: "Same"}, {ID: "b", Label: "Blocked", Disabled: true}, {ID: "c", Label: "Same"}, {ID: "d", Label: "Last"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column, Children: []ui.Node{{ID: "draft", Kind: ui.TextField, Text: "Retained"}}}, {ID: "b", Kind: ui.Column}, {ID: "c", Kind: ui.Column}, {ID: "d", Kind: ui.Column}}}
		if !readonly {
			n.Change = func(e ui.ChangeEvent) {
				calls++
				if !reject {
					selected = e.SelectedID
				}
			}
		}
		return n
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := w.state.root
	draft := findRecord(r, "draft").widget
	send := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	focused := func(id string) {
		t.Helper()
		for current, b := range r.control.tabs.buttons {
			if b.ContainsFocus() != (current == id) {
				t.Fatal("actual tab focus", current, id)
			}
		}
	}
	send(key.CodeRightArrow)
	if selected != "c" || calls != 1 {
		t.Fatal("disabled skip", selected, calls)
	}
	focused("c")
	send(key.CodeRightArrow)
	send(key.CodeRightArrow)
	if selected != "a" || calls != 3 {
		t.Fatal("wrap", selected, calls)
	}
	focused("a")
	send(key.CodeEnd)
	if selected != "d" || calls != 4 {
		t.Fatal("End", selected, calls)
	}
	send(key.CodeHome)
	if selected != "a" || calls != 5 {
		t.Fatal("Home", selected, calls)
	}
	reject = true
	send(key.CodeRightArrow)
	if selected != "a" || calls != 6 {
		t.Fatal("rejection", selected, calls)
	}
	focused("c")
	if !r.control.tabs.buttons["a"].StateIs(states.Selected) || r.control.tabs.buttons["c"].StateIs(states.Selected) {
		t.Fatal("rejection left optimistic selection")
	}
	reject = false
	send(key.CodeSpacebar)
	if selected != "c" || calls != 7 {
		t.Fatal("focused activation", selected, calls)
	}
	send(key.CodeReturnEnter)
	if calls != 7 {
		t.Fatal("selected Enter duplicated callback")
	}
	if err := a.Do(func() { readonly = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	send(key.CodeHome)
	focused("a")
	if selected != "c" || calls != 7 {
		t.Fatal("readonly mutated selection", selected, calls)
	}
	if err := a.Do(func() { readonly = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	send(key.CodeReturnEnter)
	if selected != "a" || calls != 8 || findRecord(r, "draft").widget != draft {
		t.Fatal("retained page or controlled activation lost", selected, calls)
	}
	// A handled or modified chord belongs to the earlier command/native route.
	for _, modified := range []bool{false, true} {
		e := events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
		if modified {
			e.Mods = 1 << key.Meta
		} else {
			e.SetHandled()
		}
		tabNavigationKey(r, "a", r.control.tabs.bindings["a"], e)
		a.Drain()
		if calls != 8 {
			t.Fatal("command/modified chord was intercepted")
		}
	}
	if err := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if r.control.tabs.bindings["a"].selectEpoch.Load() != nil {
		t.Fatal("disabled navigation retained authority")
	}
}

func TestTabActivationAfterPendingNavigation(t *testing.T) {
	for _, activation := range []key.Codes{key.CodeSpacebar, key.CodeReturnEnter} {
		t.Run(activation.String(), func(t *testing.T) {
			selected := "a"
			var proposals []string
			view := func() ui.Node {
				return ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: selected, Focus: ui.FocusRequest{ID: "tabs", Sequence: 1}, Options: []ui.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}, Change: func(e ui.ChangeEvent) { proposals = append(proposals, e.SelectedID); selected = e.SelectedID }}
			}
			m, a := simpleManual(t, manualConfig(t, 31), view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := w.state.root
			// Use the real Core window queue and manager focus route, withholding only owner delivery.
			route := func(code key.Codes, mods key.Modifiers) {
				t.Helper()
				if err := w.core.Input(testContext, core.ManualInput{Type: events.KeyChord, Code: code, Modifiers: mods}); err != nil {
					t.Fatal(err)
				}
				for range 32 {
					step, err := w.core.Step(testContext, 1)
					if err != nil {
						t.Fatal(err)
					}
					if step.LastType == events.KeyChord {
						return
					}
				}
				t.Fatal("key not routed within bound")
			}
			route(key.CodeRightArrow, 0)
			if !r.control.tabs.buttons["b"].ContainsFocus() {
				t.Fatal("Right did not focus B")
			}
			route(key.CodeTab, 1<<key.Shift)
			if !r.control.tabs.buttons["a"].ContainsFocus() || len(proposals) != 0 {
				t.Fatal("manager focus/owner boundary", proposals)
			}
			route(activation, 0)
			a.Drain()
			if len(proposals) != 2 || proposals[0] != "b" || proposals[1] != "a" || selected != "a" {
				t.Fatal("explicit activation lost behind pending selection", proposals, selected)
			}
			unitSettle(t, m)
			route(activation, 0)
			a.Drain()
			if len(proposals) != 2 {
				t.Fatal("accepted activation must remain a no-op", proposals)
			}
		})
	}
}

func TestTabFocusedOptionDisableRepairsNavigation(t *testing.T) {
	selected := "a"
	disableA, disableAll := false, false
	calls := 0
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "tabs", Sequence: 1}, Children: []ui.Node{{ID: "tabs", Kind: ui.Tabs, Disabled: disableAll, SelectedID: selected, Options: []ui.Option{{ID: "a", Label: "A", Disabled: disableA || disableAll}, {ID: "b", Label: "B", Disabled: disableAll}, {ID: "c", Label: "C", Disabled: disableAll}}, Children: []ui.Node{{ID: "a", Kind: ui.Column, Children: []ui.Node{{ID: "draft-a", Kind: ui.TextField, Text: "A"}}}, {ID: "b", Kind: ui.Column, Children: []ui.Node{{ID: "draft-b", Kind: ui.TextField, Text: "B"}}}, {ID: "c", Kind: ui.Column, Children: []ui.Node{{ID: "draft-c", Kind: ui.TextField, Text: "C"}}}}, Change: func(e ui.ChangeEvent) { calls++; selected = e.SelectedID }}, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
	}
	m, a := simpleManual(t, manualConfig(t, 29), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(w.state.root, "tabs")
	if !r.control.tabs.buttons["a"].ContainsFocus() {
		t.Fatal("initial focus")
	}
	if err := a.Do(func() { disableA = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if r.control.tabs.buttons["a"].ContainsFocus() || !r.control.tabs.buttons["b"].ContainsFocus() || selected != "a" || calls != 0 {
		t.Fatal("disabled focused origin was not repaired without selection", selected, calls)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if selected != "c" || calls != 1 || !r.control.tabs.buttons["c"].ContainsFocus() {
		t.Fatal("navigation from repaired focus", selected, calls)
	}
	if err := a.Do(func() { disableAll = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !findRecord(w.state.root, "after").widget.AsWidget().ContainsFocus() || calls != 1 {
		t.Fatal("all-disabled strip did not yield to enabled outside focus", calls)
	}
}
