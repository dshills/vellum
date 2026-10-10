package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"example.com/vellum/ui"
	"example.com/vellum/workbench"
	"reflect"
	"testing"
)

func TestAccessibilityWorkbenchNamespacesVisibilityActivationAndModal(t *testing.T) {
	a, b := newAXTest(t, 16)
	layout := workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "split", Axis: workbench.Horizontal, Ratio: .5, First: &workbench.Group{ID: "left", Tabs: []string{"a", "inactive"}, Active: "a"}, Second: &workbench.Group{ID: "right", Tabs: []string{"b"}, Active: "b"}}}
	workspace, err := workbench.New(layout, []string{"a", "b", "inactive"})
	if err != nil {
		t.Fatal(err)
	}
	modal := false
	panel := func(title string) func() ui.Node {
		return func() ui.Node {
			return ui.Node{ID: "panel-root", Kind: ui.Column, Children: []ui.Node{{ID: "name", Kind: ui.TextField, Text: title, Change: func(ui.ChangeEvent) {}}}}
		}
	}
	panelB := func() ui.Node {
		n := panel("B")()
		n.Children = append(n.Children, ui.Node{ID: "modal", Kind: ui.Dialog, Open: modal, Dismiss: func() { modal = false }, Children: []ui.Node{{ID: "modal-name", Kind: ui.TextField, Text: "Modal", Change: func(ui.ChangeEvent) {}}}})
		return n
	}
	w := &Workbench{Owner: a, Window: "ax", Workspace: workspace, Panels: []Panel{{ID: "a", Title: "A", View: panel("A")}, {ID: "b", Title: "B", View: panelB}, {ID: "inactive", Title: "Inactive", View: panel("Inactive")}}}
	host := core.NewFrame(core.NewBody("AX dock headless"))
	root, err := reconcile(host, nil, ui.Node{ID: "workbench", Kind: ui.Extension, Custom: w}, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if !accessibilityDialogVisible(w.records["a"].children[0]) || accessibilityDialogVisible(w.records["inactive"].children[0]) {
		t.Fatal("native active panel visibility mismatch")
	}
	activeRecords := activeWorkbenchRecords(root)
	if len(activeRecords) != 2 || activeRecords[0] != w.records["a"] || activeRecords[1] != w.records["b"] {
		t.Fatalf("active control traversal includes inactive panels: %v", activeRecords)
	}
	nodes, handlers := accessibilitySnapshot(root, b)
	seen := map[string]bool{}
	names := 0
	for _, n := range nodes {
		if seen[n.ID] {
			t.Fatal("duplicate dock identifier", n.ID)
		}
		seen[n.ID] = true
		if n.DomainID == "name" {
			names++
		}
		if n.Value == "Inactive" {
			t.Fatal("inactive panel controls published")
		}
	}
	if names != 2 {
		t.Fatalf("active panel controls=%d", names)
	}
	for _, id := range []string{accessibilityWorkbenchNamespace("workbench", "a") + accessibilityManagedID("name"), accessibilityWorkbenchNamespace("workbench", "b") + accessibilityManagedID("name")} {
		if !seen[id] {
			t.Fatal("panel namespace absent", id)
		}
	}
	projection := accessibilityFocusProjection(root, "")
	order := ui.FocusOrder(projection)
	if len(order) != 4 {
		t.Fatalf("dock keyboard order %v", order)
	}
	if !handlers[accessibilityDockOptionID("workbench", "inactive")](accessibilityAction{Kind: "press"}) {
		t.Fatal("dock header activation")
	}
	if workspace.Layout().Root.First.Active != "inactive" {
		t.Fatal("owner layout did not accept activation")
	}
	if err := w.Sync(root.widget); err != nil {
		t.Fatal(err)
	}
	modal = true
	if err := w.Sync(root.widget); err != nil {
		t.Fatal(err)
	}
	nodes, _ = accessibilitySnapshot(root, b)
	prefix := accessibilityWorkbenchNamespace("workbench", "b")
	if len(nodes) != 2 || nodes[0].ID != prefix+accessibilityManagedID("modal") || nodes[1].ID != prefix+accessibilityManagedID("modal-name") {
		t.Fatalf("dock modal did not isolate window: %+v", nodes)
	}
	if got := ui.FocusOrder(accessibilityFocusProjection(root, "")); !reflect.DeepEqual(got, []string{prefix + accessibilityManagedID("modal-name")}) {
		t.Fatalf("modal focus order %v", got)
	}
	accessibilityStampVirtualGeometry(root)
	accessibilityStampDialogNamespaces(root, "")
	if got, _ := w.records["b"].children[1].control.dialog.host.Property("vellum.accessibility-namespace").(string); got != prefix {
		t.Fatalf("dock modal geometry namespace %q", got)
	}
	bounds := accessibilityBounds(root)
	if _, ok := bounds[prefix+accessibilityManagedID("name")]; !ok {
		t.Fatal("panel bounds absent")
	}
	root.node.Disabled = true
	if len(activeWorkbenchRecords(root)) != 2 {
		t.Fatal("disabled dock omitted modal cleanup traversal")
	}
	root.widget.AsWidget().SetProperty(disabledProperty, true)
	dialog := w.records["b"].children[1]
	wakes := 0
	bindControlWake(root, func(events.Event) { wakes++ })
	body := prepareDialogBody(dialog)
	body.Scene.Stage = &core.Stage{}
	dialog.control.dialog.active = true
	event := &events.CustomEvent{}
	event.Data = "vellum.wake"
	event.Typ = events.Custom
	body.Scene.HandleEvent(event)
	if wakes != 1 {
		t.Fatal("panel modal Custom wake did not reach origin callback")
	}
	flushDialogs(root)
	if dialog.control.dialog.active || dialog.control.dialog.body != nil {
		t.Fatal("disabled active panel retained native modal")
	}
	nodes, handlers = accessibilitySnapshot(root, b)
	for _, node := range nodes {
		if node.Role == "dialog" || !node.Disabled || len(node.Actions) != 0 {
			t.Fatalf("disabled dock published interactive or hidden modal content: %+v", node)
		}
	}
	if len(handlers) != 0 {
		t.Fatal("disabled dock handlers retained")
	}
}
func TestAccessibilityFocusProjectionDoesNotMutateOptions(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("AX tabs"))
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "one", Options: []ui.Option{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}, Children: []ui.Node{{ID: "one", Kind: ui.Column}, {ID: "two", Kind: ui.Column}}}
	root, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	_ = accessibilityFocusProjection(root, "")
	if root.node.Options[0].ID != "one" || root.node.Options[1].ID != "two" || n.Options[0].ID != "one" {
		t.Fatal("projection mutated retained model options")
	}
}

func TestAccessibilityNamespacePreservesFocusPermissionsAndGeneration(t *testing.T) {
	_, b := newAXTest(t, 8)
	generation := uint64(1)
	captured := generation
	calls := 0
	nodes := []accessibilityNode{
		{ID: "row", Role: "row", Actions: []string{"press"}},
		{ID: "disabled", Role: "text-field", Focusable: true, Disabled: true, Actions: []string{"focus"}},
		{ID: "enabled", Role: "text-field", Focusable: true, Actions: []string{"focus"}},
	}
	handlers := map[string]accessibilityHandler{}
	for _, n := range nodes {
		id := n.ID
		handlers[id] = func(action accessibilityAction) bool {
			calls++
			if action.Kind != "focus" || generation != captured {
				return false
			}
			accessibilityFocus(b, id)
			return true
		}
	}
	_, wrapped := namespaceAccessibility(nodes, handlers, "panel:", b)
	for _, id := range []string{"row", "disabled"} {
		if wrapped["panel:"+id](accessibilityAction{Kind: "focus"}) {
			t.Fatal("namespace invented focus permission", id)
		}
	}
	if calls != 0 || b.pendingFocus.Load() != nil {
		t.Fatal("refused focus reached original callback")
	}
	if !wrapped["panel:enabled"](accessibilityAction{Kind: "focus"}) {
		t.Fatal("valid namespaced focus refused")
	}
	if focus := b.pendingFocus.Swap(nil); focus == nil || *focus != "panel:enabled" {
		t.Fatal("qualified focus missing", focus)
	}
	generation++
	if wrapped["panel:enabled"](accessibilityAction{Kind: "focus"}) || b.pendingFocus.Load() != nil {
		t.Fatal("namespace bypassed retired generation")
	}
}
