package cogent

import (
	"reflect"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
)

func disclosureFixture() ui.Node {
	return ui.Node{ID: "details", Kind: ui.Disclosure, Text: "<Details>", Checked: true, Children: []ui.Node{
		{ID: "name", Kind: ui.TextField, Text: "Accepted"},
		{ID: "action", Kind: ui.Button, Text: "Action"},
	}}
}

func TestDisclosureRetainedDraftLatestCallbackAndRevokedRequests(t *testing.T) {
	a, _ := newAXTest(t, 16)
	host := core.NewFrame(core.NewBody("disclosure ownership"))
	n := disclosureFixture()
	n.Change = func(ui.ChangeEvent) { t.Fatal("old callback") }
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	c := r.control.disclosure
	if c.header.Text != "&lt;Details&gt;" {
		t.Fatal("heading entered native markup without escaping")
	}
	field := r.children[0].widget.(*core.TextField)
	if !field.ApplyNativeEdit("Draft", 5, 5, -1, -1, "insert") {
		t.Fatal("draft admission")
	}
	draft := field.EditingState()
	queueDisclosure(r, false)
	if c.host.StateIs(states.Invisible) {
		t.Fatal("request changed accepted expansion before owner")
	}
	calls := []ui.ChangeEvent{}
	n.Change = func(e ui.ChangeEvent) { calls = append(calls, e) }
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if !reflect.DeepEqual(calls, []ui.ChangeEvent{{ID: "details", Kind: ui.Disclosure, Checked: false}}) {
		t.Fatal("latest callback", calls)
	}
	// Refusal leaves the accepted expanded state without optimistic body changes.
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil || c.host.StateIs(states.Invisible) {
		t.Fatal("refusal", err)
	}
	n.Checked = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil || !c.host.StateIs(states.Invisible) || !r.children[0].control.disabled {
		t.Fatal("collapse did not revoke child admission", err)
	}
	n.Checked = true
	n.Children[0], n.Children[1] = n.Children[1], n.Children[0]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil || r.control.disclosure != c || r.children[1].widget != field || field.Parent != c.host || field.EditingState() != draft || c.host.StateIs(states.Invisible) {
		t.Fatal("restore/reorder lost identity or draft", err)
	}
	queueDisclosure(r, false)
	n.Disabled = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(calls) != 1 {
		t.Fatal("disabled queued request delivered")
	}
	n.Disabled = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	queueDisclosure(r, false)
	n.Checked = false // A later accepted descriptor revokes the older request.
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(calls) != 1 {
		t.Fatal("older request survived programmatic state change")
	}
	n.Checked = true
	n.Change = nil
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	queueDisclosure(r, false)
	a.Drain()
	if c.host.StateIs(states.Invisible) || len(calls) != 1 {
		t.Fatal("read-only section changed expansion")
	}
	n.Change = func(e ui.ChangeEvent) { calls = append(calls, e) }
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	queueDisclosure(r, false)
	disposeControl(r)
	a.Drain()
	if len(calls) != 1 {
		t.Fatal("retired request delivered")
	}
}

func TestDisclosureSemanticHierarchyAndHiddenActionRetirement(t *testing.T) {
	a, b := newAXTest(t, 16)
	n := disclosureFixture()
	calls := 0
	n.Change = func(e ui.ChangeEvent) { n.Checked = e.Checked; calls++ }
	n.Children[1].Activate = func() { calls += 100 }
	host := core.NewFrame(core.NewBody("disclosure semantics"))
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	id := accessibilityDisclosureID(n.ID)
	if len(nodes) != 4 || nodes[0].Focusable || nodes[1].ID != id || nodes[1].Role != "disclosure-heading" || nodes[1].Expanded == nil || !*nodes[1].Expanded || nodes[1].Parent != nodes[0].ID {
		t.Fatal("group/heading hierarchy", nodes)
	}
	if !handlers[id](accessibilityAction{Kind: "focus"}) || !accessibilityApplyFocus(r, id) {
		t.Fatal("heading focus proxy")
	}
	if !handlers[id](accessibilityAction{Kind: "collapse"}) || n.Checked || calls != 1 {
		t.Fatal("collapse request")
	}
	staleChild := handlers[accessibilityManagedID("action")]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers = accessibilitySnapshot(r, b)
	if len(nodes) != 2 || *nodes[1].Expanded || staleChild(accessibilityAction{Kind: "press"}) || calls != 1 {
		t.Fatal("collapsed child remains semantic/action target", nodes, calls)
	}
	if !handlers[id](accessibilityAction{Kind: "expand"}) || !n.Checked || calls != 2 {
		t.Fatal("expand request")
	}
	if handlers[id](accessibilityAction{Kind: "set", Value: "true"}) {
		t.Fatal("unsupported action admitted")
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	n.Children[1].Activate = func() { calls += 100 }
	action := r.children[1]
	action.widget.AsWidget().Send(events.Click)
	n.Checked = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if calls != 2 {
		t.Fatal("hidden child queued activation delivered", calls)
	}
}
