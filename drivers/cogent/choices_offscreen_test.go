//go:build offscreen

package cogent

import (
	"bytes"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestChoiceManualKeyboardPointerAndFocus(t *testing.T) {
	for _, kind := range []ui.Kind{ui.RadioGroup, ui.SegmentedControl} {
		t.Run(string(kind), func(t *testing.T) {
			selected := "a"
			edits := 0
			view := func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "choice", Sequence: 1}, Children: []ui.Node{{ID: "choice", Kind: kind, SelectedID: selected, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Unavailable", Disabled: true}, {ID: "c", Label: "Charlie"}}, Change: func(e ui.ChangeEvent) { selected = e.SelectedID; edits++ }}, {ID: "after", Kind: ui.Button, Text: "After choices"}}}
			}
			m, _ := simpleManual(t, manualConfig(t, 7), view)
			unitSettle(t, m)
			snapshot, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			if snapshot.Windows[0].FocusedID != "choice" {
				t.Fatal("group initial focus missing", snapshot.Windows[0].FocusedID)
			}
			before, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			w := m.implementation.find("one")
			send := func(code key.Codes) {
				t.Helper()
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			send(key.CodeRightArrow)
			if selected != "c" || edits != 1 {
				t.Fatal("arrow did not skip disabled option", selected, edits)
			}
			send(key.CodeRightArrow)
			if selected != "a" || edits != 2 {
				t.Fatal("arrow did not wrap", selected, edits)
			}
			send(key.CodeEnd)
			if selected != "c" {
				t.Fatal("End did not choose last enabled")
			}
			send(key.CodeHome)
			if selected != "a" {
				t.Fatal("Home did not choose first enabled")
			}
			send(key.CodeSpacebar)
			if selected != "a" || edits != 4 {
				t.Fatal("Space cleared selected option or emitted redundant change", selected, edits)
			}
			r := findRecord(w.state.root, "choice")
			box := r.control.choice.buttons["c"].Geom.TotalBBox
			if box.Empty() {
				t.Fatal("option geometry missing")
			}
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: box.Min.Add(box.Size().Div(2))}); err != nil {
					t.Fatal(err)
				}
			}
			unitSettle(t, m)
			if selected != "c" || edits != 5 {
				t.Fatal("measured pointer selection failed", selected, edits)
			}
			after, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(before.Pix, after.Pix) {
				t.Fatal("selection produced no completed raster change")
			}
			send(key.CodeTab)
			snapshot, err = m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			if snapshot.Windows[0].FocusedID != "after" {
				t.Fatal("group added multiple managed Tab stops", snapshot.Windows[0].FocusedID)
			}
			targets := map[string]accessibilityFocusTarget{}
			accessibilityNativeFocusTargets(w.state.root, "", targets)
			optionID := accessibilityOptionID("choice", "a")
			if !accessibilityApplyFocus(w.state.root, optionID) {
				t.Fatal("visible option focus refused")
			}
			n := r.node
			n.Disabled = true
			if err := syncRecord(r, n); err != nil {
				t.Fatal(err)
			}
			if accessibilityApplyFocus(w.state.root, optionID) {
				t.Fatal("disabled group option focus accepted")
			}
		})
	}
}

// Option IDs are local to their group, so they may equal an ordinary node ID.
func TestChoiceManualCaptureDoesNotAliasOptionDomainIDs(t *testing.T) {
	for _, kind := range []ui.Kind{ui.RadioGroup, ui.SegmentedControl} {
		t.Run(string(kind), func(t *testing.T) {
			m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
					{ID: "choice", Kind: kind, SelectedID: "later", Options: []ui.Option{{ID: "later", Label: "Selected option"}}},
					{ID: "later", Kind: ui.Checkbox, Text: "Unchecked", Checked: false},
				}}
			})
			unitSettle(t, m)
			snapshot, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, element := range snapshot.Windows[0].Elements {
				if element.ID == "later" {
					found = true
					if element.Value != false || element.Selected {
						t.Fatalf("ordinary checkbox captured option value: %+v", element)
					}
				}
			}
			if !found {
				t.Fatal("checkbox missing")
			}
		})
	}
}

func TestChoiceManualQuietPointerFocusAdvancesOrdinaryControl(t *testing.T) {
	m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
			{ID: "check", Kind: ui.Checkbox, Text: "Check", Change: func(ui.ChangeEvent) {}},
			{ID: "after", Kind: ui.Button, Text: "After"},
		}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	box := findRecord(w.state.root, "check").widget.AsWidget().Geom.TotalBBox
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: box.Min.Add(box.Size().Div(2))}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Windows[0].FocusedID != "check" {
		t.Fatal("quiet pointer focus omitted", snapshot.Windows[0].FocusedID)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeTab}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	snapshot, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Windows[0].FocusedID != "after" {
		t.Fatal("Tab lost quiet pointer focus", snapshot.Windows[0].FocusedID)
	}
}

func TestChoicePendingSemanticFocusUsesCurrentVisibleBinding(t *testing.T) {
	options := []ui.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}
	hidden, disabled := false, false
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
			{ID: "tabs", Kind: ui.Tabs, SelectedID: map[bool]string{false: "visible", true: "hidden"}[hidden], Options: []ui.Option{{ID: "visible"}, {ID: "hidden"}}, Children: []ui.Node{
				{ID: "visible", Kind: ui.Column, Children: []ui.Node{{ID: "choice", Kind: ui.RadioGroup, SelectedID: "a", Options: options, Disabled: disabled}}},
				{ID: "hidden", Kind: ui.Label, Text: "Other page"},
			}},
		}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(w.state.root, "choice")
	bridge := &accessibilityBridge{}
	_, handlers := accessibilitySnapshot(w.state.root, bridge)
	id := accessibilityOptionID("choice", "b")
	old := r.control.choice.buttons["b"]
	if !handlers[id](accessibilityAction{Kind: "focus"}) {
		t.Fatal("live option focus admission refused")
	}
	refresh := func() {
		t.Helper()
		if err := a.Invalidate("one"); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	options = options[:1]
	refresh()
	pending := bridge.pendingFocus.Load()
	if pending == nil || *pending != id || accessibilityApplyFocus(w.state.root, *pending) {
		t.Fatal("removed option retained focus target")
	}
	options = append(options, ui.Option{ID: "b", Label: "Replacement B"})
	refresh()
	r = findRecord(w.state.root, "choice")
	if r.control.choice.buttons["b"] == old {
		t.Fatal("replacement reused retired binding")
	}
	if handlers[id](accessibilityAction{Kind: "focus"}) {
		t.Fatal("old handler admitted replacement")
	}
	if !accessibilityApplyFocus(w.state.root, *pending) || !r.control.choice.buttons["b"].ContainsFocus() {
		t.Fatal("pending ID did not focus current replacement")
	}
	disabled = true
	refresh()
	if accessibilityApplyFocus(w.state.root, *pending) {
		t.Fatal("disabled group admitted pending focus")
	}
	disabled = false
	hidden = true
	refresh()
	if accessibilityApplyFocus(w.state.root, *pending) {
		t.Fatal("hidden group admitted pending focus")
	}
}
