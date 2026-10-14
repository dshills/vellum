package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
	"testing"
)

func closeTabsFixture(handler func(ui.TabCloseEvent)) ui.Node {
	return ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "First", Closable: true}, {ID: "b", Label: "Second", Closable: true}}, TabClose: handler, Children: []ui.Node{{ID: "a", Kind: ui.Column, Children: []ui.Node{{ID: "draft", Kind: ui.TextField, Text: "accepted"}}}, {ID: "b", Kind: ui.Column}}}
}
func TestTabCloseControlledRejectLatestAndPrivateBinding(t *testing.T) {
	a, _ := newAXTest(t, 32)
	host := core.NewFrame(core.NewBody("tabs"))
	first, latest, selections := 0, 0, 0
	n := closeTabsFixture(func(e ui.TabCloseEvent) {
		first++
		if e.ID != "tabs" || e.PageID != "b" {
			t.Error("wrong close event", e)
		}
	})
	n.Change = func(ui.ChangeEvent) { selections++ }
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	oldPage := r.control.tabs.pages["b"]
	binding := r.control.tabs.bindings["b"]
	close := tabCloseButton(binding)
	if close == nil {
		t.Fatal("private close button absent")
	}
	close.Send(events.Click)
	if r.widget.(*core.Tabs).NumTabs() != 2 || r.node.SelectedID != "a" || !r.control.tabs.buttons["a"].StateIs(states.Selected) {
		t.Fatal("native close selected/deleted before owner admission")
	}
	n.TabClose = func(ui.TabCloseEvent) { latest++ }
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if first != 0 || latest != 1 || selections != 0 || r.control.tabs.pages["b"] != oldPage {
		t.Fatal("rejection/latest/selection isolation", first, latest, selections)
	}
	// Hide/recreate only the close artwork; held old private buttons cannot act on
	// a current capability of the same still-retained tab.
	n.Options[1].Closable = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	n.Options[1].Closable = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	current := tabCloseButton(binding)
	if current == close || current == nil {
		t.Fatal("fixture did not replace private close button")
	}
	close.Send(events.Click)
	binding.button.CloseRequest(close) // a held private callback can run after its artwork was retired
	a.Drain()
	if latest != 1 {
		t.Fatal("retired private close button acted on replacement")
	}
	// Retire and readd a page with the same domain ID. Both queued close and tab
	// selection must retain the originating binding, never an ID-only lookup.
	current.Send(events.Click)
	binding.button.Send(events.Click)
	saved := n.Children[1]
	n.Options = n.Options[:1]
	n.Children = n.Children[:1]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	n.Options = append(n.Options, ui.Option{ID: "b", Label: "Replacement", Closable: true})
	n.Children = append(n.Children, saved)
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if latest != 1 || selections != 0 || r.control.tabs.bindings["b"] == binding {
		t.Fatal("stale tab request reached new binding")
	}
}
func TestTabCloseDisabledHiddenAndAcceptedReparent(t *testing.T) {
	a, _ := newAXTest(t, 32)
	host := core.NewFrame(core.NewBody("tabs"))
	calls := 0
	n := closeTabsFixture(func(ui.TabCloseEvent) { calls++ })
	root := ui.Node{ID: "root", Kind: ui.Disclosure, Text: "Documents", Checked: true, Children: []ui.Node{n}}
	r, err := reconcile(host, nil, root, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	tabs := findRecord(r, "tabs")
	binding := tabs.control.tabs.bindings["a"]
	close := tabCloseButton(binding)
	draft := findRecord(r, "draft").widget.(*core.TextField)
	if !draft.ApplyNativeEdit("pending draft", 13, 13, -1, -1, "insert") {
		t.Fatal("native draft setup failed")
	}
	close.Send(events.Click)
	root.Checked = false
	r, err = reconcile(host, r, root, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if calls != 0 {
		t.Fatal("hidden close delivered")
	}
	root.Checked = true
	root.Children[0].Options[0].Disabled = true
	r, err = reconcile(host, r, root, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	close.Send(events.Click)
	a.Drain()
	if calls != 0 {
		t.Fatal("disabled option close delivered")
	}
	root.Children[0].Options[0].Disabled = false
	r, err = reconcile(host, r, root, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	// Accepted removal moves the original draft out before deleting the old page.
	next := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "draft", Kind: ui.TextField, Text: "accepted"}, {ID: "tabs", Kind: ui.Tabs, SelectedID: "b", Options: []ui.Option{{ID: "b", Label: "Second", Closable: true}}, TabClose: n.TabClose, Children: []ui.Node{n.Children[1]}}}}
	close.Send(events.Click)
	a.Drain()
	if calls != 1 {
		t.Fatal("enabled close missing")
	}
	r, err = reconcile(host, r, next, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "draft").widget != draft || draft.EditingState().Text != "pending draft" || findRecord(r, "tabs").widget.(*core.Tabs).NumTabs() != 1 {
		t.Fatal("accepted close destroyed retained reparented draft")
	}
	close.Send(events.Click)
	a.Drain()
	if calls != 1 {
		t.Fatal("late detached close reenabled unmanaged behavior")
	}
}
func TestTabCloseQueueRefusalConsumesWithoutDeleting(t *testing.T) {
	a, _ := newAXTest(t, 1)
	host := core.NewFrame(core.NewBody("tabs"))
	calls := 0
	n := closeTabsFixture(func(ui.TabCloseEvent) { calls++ })
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	if err = a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	binding := r.control.tabs.bindings["a"]
	close := tabCloseButton(binding)
	e := events.NewMouse(events.Click, events.Left, close.Geom.TotalBBox.Min, 0)
	close.HandleEvent(e)
	if !e.IsHandled() || r.widget.(*core.Tabs).NumTabs() != 2 {
		t.Fatal("saturated native close was destructive/unhandled")
	}
	a.Drain()
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 || r.widget.(*core.Tabs).NumTabs() != 2 || r.widget.AsTree().Property("vellum.control-error") == nil {
		t.Fatal("queue refusal lost control state/error")
	}
}
func TestTabCloseCapabilityRevokedAfterDisableShowAndSemanticReplacement(t *testing.T) {
	a, _ := newAXTest(t, 16)
	host := core.NewFrame(core.NewBody("tabs"))
	calls := 0
	n := closeTabsFixture(func(ui.TabCloseEvent) { calls++ })
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	_, handlers := accessibilitySnapshot(r, nil)
	oldHandler := handlers[tabCloseID("tabs", "a")]
	if oldHandler == nil {
		t.Fatal("missing handler")
	}
	binding := r.control.tabs.bindings["a"]
	close := tabCloseButton(binding)
	close.Send(events.Click)
	n.Disabled = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, current := accessibilitySnapshot(r, nil)
	for _, x := range nodes {
		if x.ID == tabCloseID("tabs", "a") && (!x.Disabled || len(x.Actions) != 0) {
			t.Fatal("disabled close exposed actions", x)
		}
	}
	if current[tabCloseID("tabs", "a")] != nil {
		t.Fatal("disabled close retained handler")
	}
	n.Disabled = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if calls != 0 || oldHandler(accessibilityAction{Kind: "press"}) {
		t.Fatal("disabled-then-restored capability admitted old queue/semantic action")
	}
	_, current = accessibilitySnapshot(r, nil)
	if !current[tabCloseID("tabs", "a")](accessibilityAction{Kind: "press"}) {
		t.Fatal("fresh semantic action refused")
	}
	a.Drain()
	if calls != 1 {
		t.Fatal("fresh close not delivered")
	}
}
func TestTabCloseDialogNativeDismissalRevokesBeforeParking(t *testing.T) {
	a, _ := newAXTest(t, 32)
	host := core.NewFrame(core.NewBody("tabs"))
	calls := 0
	n := ui.Node{ID: "dialog", Kind: ui.Dialog, Text: "Documents", Open: true, Dismiss: func() {}, Children: []ui.Node{closeTabsFixture(func(ui.TabCloseEvent) { calls++ })}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	tabs := findRecord(r, "tabs")
	binding := tabs.control.tabs.bindings["a"]
	if tabCloseEpochLive(binding.closeEpoch.Load()) {
		t.Fatal("unopened dialog admitted close")
	}
	body := prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	if err = syncRecord(r, n); err != nil {
		t.Fatal(err)
	}
	oldEpoch := binding.closeEpoch.Load()
	if !tabCloseEpochLive(oldEpoch) {
		t.Fatal("opened dialog close missing")
	}
	_, handlers := accessibilitySnapshot(r, nil)
	oldHandler := handlers[tabCloseID("tabs", "a")]
	queueTabClose(tabs, "a", binding)
	body.Send(events.Close)
	if tabCloseEpochLive(oldEpoch) {
		t.Fatal("native dismissal left old close capability alive before owner drain")
	}
	a.Drain()
	if calls != 0 {
		t.Fatal("queued close delivered after dismissed/parked dialog")
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if tabs.control.tabs.visible || tabCloseEpochLive(binding.closeEpoch.Load()) {
		t.Fatal("rejected Dismiss with Open true revived close")
	}
	n.Open = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	n.Open = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	body = prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	if err = syncRecord(r, n); err != nil {
		t.Fatal(err)
	}
	if oldHandler == nil || oldHandler(accessibilityAction{Kind: "press"}) || binding.closeEpoch.Load() == oldEpoch || !tabCloseEpochLive(binding.closeEpoch.Load()) {
		t.Fatal("old dialog epoch/handler resurrected on genuine reopen")
	}
	queueTabClose(tabs, "a", binding)
	a.Drain()
	if calls != 1 {
		t.Fatal("new dialog capability failed")
	}
}

func TestTabCloseDialogWorkbenchKeepsEnclosingLifetime(t *testing.T) {
	a, _ := newAXTest(t, 32)
	host := core.NewFrame(core.NewBody("nested scope"))
	calls := 0
	workspace, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "group", Tabs: []string{"panel"}, Active: "panel"}}, []string{"panel"})
	if err != nil {
		t.Fatal(err)
	}
	w := &Workbench{Owner: a, Window: "ax", Workspace: workspace, Panels: []Panel{{ID: "panel", Title: "Documents", View: func() ui.Node { return closeTabsFixture(func(ui.TabCloseEvent) { calls++ }) }}}}
	n := ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true, Dismiss: func() {}, Children: []ui.Node{{ID: "workspace", Kind: ui.Extension, Custom: w}}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	body := prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	if err = syncRecord(r, n); err != nil {
		t.Fatal(err)
	}
	tabs := w.records["panel"]
	binding := tabs.control.tabs.bindings["a"]
	epoch := binding.closeEpoch.Load()
	if !tabCloseEpochLive(epoch) || len(epoch.scopes) != 1 {
		t.Fatal("panel dropped enclosing scope")
	}
	queueTabClose(tabs, "a", binding)
	body.Send(events.Close)
	if tabCloseEpochLive(epoch) {
		t.Fatal("native dismissal left workbench close live")
	}
	a.Drain()
	if calls != 0 {
		t.Fatal("nested workbench close delivered after dismissal")
	}
	if err = w.Sync(findRecord(r, "workspace").widget); err != nil {
		t.Fatal(err)
	}
	if tabCloseEpochLive(binding.closeEpoch.Load()) || tabs.control.tabs.visible {
		t.Fatal("independent panel Sync revived closed scope")
	}
}
