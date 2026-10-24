//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestRuleEditorNativeAddRemoveAndRetention(t *testing.T) {
	e := &ui.RuleEditor{ID: "filter", Fields: []ui.RuleField{{ID: "name", Label: "Name", Kind: ui.RuleText}}, Root: ui.Rule{ID: "root", Group: &ui.RuleGroup{Match: ui.RuleAll, Rules: []ui.Rule{{ID: "existing", FieldID: "name", Operator: ui.RuleContains}}}}}
	calls := 0
	e.Change = func(c ui.RuleEditEvent) { calls++; e.Root = c.Root }
	view := func() ui.Node {
		n, err := e.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.RuleControlID(e.ID, "root", "add-rule"), Sequence: 1}, Children: []ui.Node{n}}
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	if err = a.Open(app.Window{ID: "one", Title: "Rule editor", Width: 800, Height: 600, View: view}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 822))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	original := findRecord(w.state.root, ui.RuleControlID(e.ID, "existing", "value")).widget
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 1 || len(e.Root.Group.Rules) != 2 {
		t.Fatal("keyboard add failed", calls)
	}
	_, target := imageManualElement(t, m, ui.RuleControlID(e.ID, "rule-1", "remove"))
	if target.Bounds.Empty() {
		t.Fatalf("remove bounds empty: %+v", target)
	}
	for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: kind, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if calls != 2 || len(e.Root.Group.Rules) != 1 || findRecord(w.state.root, ui.RuleControlID(e.ID, "existing", "value")).widget != original {
		t.Fatal("remove lost retained sibling", calls)
	}
}
