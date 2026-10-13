package cogent

import (
	"errors"
	"reflect"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func TestChoiceRetainedDomainIdentityAndQueuedRequests(t *testing.T) {
	for _, kind := range []ui.Kind{ui.RadioGroup, ui.SegmentedControl} {
		t.Run(string(kind), func(t *testing.T) {
			a, _ := newAXTest(t, 8)
			host := core.NewFrame(core.NewBody("choices"))
			var edits []ui.ChangeEvent
			n := ui.Node{ID: "choice", Kind: kind, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "<A>"}, {ID: "b", Label: "B"}, {ID: "c", Label: "C", Disabled: true}}, Change: func(e ui.ChangeEvent) { edits = append(edits, e) }}
			r, err := reconcile(host, nil, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			c := r.control.choice
			aButton, bButton := c.buttons["a"], c.buttons["b"]
			if r.widget.AsTree().NumChildren() != 3 || aButton.Text != "&lt;A&gt;" || !aButton.IsChecked() || !c.buttons["c"].StateIs(states.Disabled) {
				t.Fatal("initial choice state")
			}
			bButton.SetChecked(true).SendChange()
			if len(edits) != 0 || !aButton.IsChecked() || bButton.IsChecked() {
				t.Fatal("native pending request bypassed owner or exclusive accepted state")
			}
			n.Options = []ui.Option{n.Options[1], n.Options[0], n.Options[2]}
			n.Change = func(e ui.ChangeEvent) { e.Text = "latest"; edits = append(edits, e) }
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if len(edits) != 1 || edits[0].SelectedID != "b" || edits[0].Kind != kind || edits[0].Text != "latest" || r.control.choice.buttons["a"] != aButton || r.control.choice.buttons["b"] != bButton || r.widget.AsTree().Child(0) != bButton {
				t.Fatal("index/stale callback/identity error", edits)
			}
			// Rejection retains A; an accepted descriptor selects exactly B.
			n.SelectedID = "b"
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			if aButton.IsChecked() || !bButton.IsChecked() {
				t.Fatal("accepted mutual selection failed")
			}
			bButton.SetChecked(false).SendChange()
			a.Drain()
			if len(edits) != 1 || !bButton.IsChecked() {
				t.Fatal("selected choice cleared or emitted redundant edit")
			}
			aButton.SetChecked(true).SendChange()
			n.Disabled = true
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if len(edits) != 1 {
				t.Fatal("disabled group delivered queued change")
			}
			n.Disabled = false
			n.SelectedID = "a"
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			bButton.SetChecked(true).SendChange()
			n.Options = []ui.Option{{ID: "a", Label: "A"}}
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			n.Options = append(n.Options, ui.Option{ID: "b", Label: "New B"})
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			if r.control.choice.buttons["b"] == bButton {
				t.Fatal("retired option reused native binding")
			}
			a.Drain()
			if len(edits) != 1 {
				t.Fatal("retired/readded option inherited queued request")
			}
		})
	}
}
func TestChoiceQueueRefusalRestoresAndReports(t *testing.T) {
	a, _ := newAXTest(t, 1)
	host := core.NewFrame(core.NewBody("refusal"))
	n := ui.Node{ID: "group", Kind: ui.RadioGroup, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Change: func(ui.ChangeEvent) { t.Fatal("refused callback delivered") }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	reports := 0
	r.control.onError = func(err error) {
		if !errors.Is(err, app.ErrQueueFull) {
			t.Fatal(err)
		}
		reports++
	}
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	r.control.choice.buttons["b"].SetChecked(true).SendChange()
	a.Drain()
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if reports != 1 || !r.control.choice.buttons["a"].IsChecked() || r.control.choice.buttons["b"].IsChecked() {
		t.Fatal("refusal restore/report", reports)
	}
}
func TestChoiceAccessibilityDomainActionsAndStaleBinding(t *testing.T) {
	a, b := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("AX choices"))
	var selected []string
	n := ui.Node{ID: "group", Kind: ui.SegmentedControl, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "duplicate"}, {ID: "b", Label: "duplicate"}, {ID: "c", Disabled: true}}, Change: func(e ui.ChangeEvent) { selected = append(selected, e.SelectedID) }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	if len(nodes) != 4 || nodes[0].Role != "radio-group" || !nodes[0].Focusable || nodes[1].Value != true || !reflect.DeepEqual(nodes[2].Actions, []string{"focus", "press"}) || len(nodes[3].Actions) != 0 {
		t.Fatal(nodes)
	}
	id := accessibilityOptionID("group", "b")
	old := handlers[id]
	if !old(accessibilityAction{Kind: "press"}) || len(selected) != 1 || selected[0] != "b" {
		t.Fatal("domain press failed", selected)
	}
	if applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: "duplicate"}) || applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: "c"}) {
		t.Fatal("choice accepted ambiguous label or disabled value")
	}
	n.Options = n.Options[:1]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	n.Options = append(n.Options, ui.Option{ID: "b"})
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if old(accessibilityAction{Kind: "press"}) || len(selected) != 1 {
		t.Fatal("stale semantic binding admitted")
	}
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(r, "", targets)
	if targets[id].scopeID != accessibilityManagedID("group") || targets[id].widget != r.control.choice.buttons["b"] {
		t.Fatal("option scope/focus mapping", targets[id])
	}
}

func TestChoiceWorkbenchModalFocusNamespaces(t *testing.T) {
	for _, kind := range []ui.Kind{ui.RadioGroup, ui.SegmentedControl} {
		t.Run(string(kind), func(t *testing.T) {
			a, b := newAXTest(t, 8)
			workspace, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "dock", Tabs: []string{"pane"}, Active: "pane"}}, []string{"pane"})
			if err != nil {
				t.Fatal(err)
			}
			open := true
			w := &Workbench{Owner: a, Window: "ax", Workspace: workspace, Panels: []Panel{{ID: "pane", Title: "Pane", View: func() ui.Node {
				return ui.Node{ID: "panel", Kind: ui.Column, Children: []ui.Node{{ID: "modal", Kind: ui.Dialog, Open: open, Dismiss: func() { open = false }, Children: []ui.Node{{ID: "choice", Kind: kind, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}}}}}}
			}}}}
			host := core.NewFrame(core.NewBody("choice namespace"))
			root, err := reconcile(host, nil, ui.Node{ID: "workbench", Kind: ui.Extension, Custom: w}, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			prefix := accessibilityWorkbenchNamespace("workbench", "pane")
			id := prefix + accessibilityOptionID("choice", "b")
			_, handlers := accessibilitySnapshot(root, b)
			if handlers[id] == nil || !handlers[id](accessibilityAction{Kind: "focus"}) {
				t.Fatal("namespaced modal option focus refused")
			}
			pending := b.pendingFocus.Swap(nil)
			if pending == nil || *pending != id {
				t.Fatal("pending focus namespace missing", pending)
			}
			targets := map[string]accessibilityFocusTarget{}
			accessibilityNativeFocusTargets(root, "", targets)
			if targets[id].scopeID != prefix+accessibilityManagedID("choice") || !accessibilityApplyFocus(root, id) {
				t.Fatal("qualified group scope refused", targets[id])
			}
			open = false
			if err := w.Sync(root.widget); err != nil {
				t.Fatal(err)
			}
			if accessibilityApplyFocus(root, id) {
				t.Fatal("closed modal admitted old option focus")
			}
		})
	}
}

func TestChoiceRetiredNativeChangeDoesNotRestoreWidgets(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("retired choice"))
	edits := 0
	n := ui.Node{ID: "choice", Kind: ui.RadioGroup, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Change: func(ui.ChangeEvent) { edits++ }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	late := r.control.choice.buttons["b"]
	disposeControl(r)
	late.SetChecked(true).SendChange()
	a.Drain()
	if edits != 0 {
		t.Fatal("retired callback delivered")
	}
	if !late.IsChecked() {
		t.Fatal("late event mutated retired choice during restoration")
	}
}
