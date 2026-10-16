package cogent

import (
	"cogentcore.org/core/core"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func newAXTest(t *testing.T, capacity int) (*app.Application, *accessibilityBridge) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: capacity})
	if err != nil {
		t.Fatal(err)
	}
	err = a.Open(app.Window{ID: "ax", Title: "AX headless", Width: 200, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	if err != nil {
		t.Fatal(err)
	}
	b := newAccessibilityBridge(a)
	b.bind("ax")
	t.Cleanup(func() { b.close(); a.Close() })
	return a, b
}
func TestAccessibilityAdmissionBoundedStaleAndRetired(t *testing.T) {
	a, b := newAXTest(t, 1)
	calls := 0
	b.mu.Lock()
	b.handlers["button"] = func(accessibilityAction) bool { calls++; return true }
	b.mu.Unlock()
	b.generation.Store(1)
	if !accessibilityDeliver(b.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("valid action not admitted")
	}
	if accessibilityDeliver(b.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("queue overflow admitted")
	}
	b.generation.Store(2)
	_ = a.Drain()
	if calls != 0 {
		t.Fatal("stale admitted action executed")
	}
	if accessibilityDeliver(b.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("stale action admitted")
	}
	if !accessibilityDeliver(b.token, 2, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("current action not admitted")
	}
	_ = a.Drain()
	if calls != 1 {
		t.Fatal(calls)
	}
	if !accessibilityDeliver(b.token, 2, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("current action not admitted")
	}
	b.close()
	_ = a.Drain()
	if calls != 1 {
		t.Fatal("retired callback executed")
	}
	if accessibilityDeliver(b.token, 3, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("closed token admitted")
	}
}

func TestAccessibilityRetiredBridgeCannotRepublishWindow(t *testing.T) {
	_, b := newAXTest(t, 8)
	b.close()
	// A refresh can have passed its initial closed check before native teardown.
	b.bind("ax")
	accessibilityRegistry.Lock()
	_, registered := accessibilityRegistry.windows["ax"]
	_, tokenRegistered := accessibilityRegistry.bridges[b.token]
	accessibilityRegistry.Unlock()
	if registered || tokenRegistered {
		t.Fatal("retired bridge republished window registry entry", registered, tokenRegistered)
	}
}

func TestAccessibilityConcurrentBindAndCloseRetireRegistry(t *testing.T) {
	a, _ := newAXTest(t, 8)
	for range 256 {
		b := newAccessibilityBridge(a)
		b.bind("ax")
		start := make(chan struct{})
		var workers sync.WaitGroup
		workers.Add(2)
		go func() { defer workers.Done(); <-start; b.bind("ax") }()
		go func() { defer workers.Done(); <-start; b.closeFrom("concurrent-close") }()
		close(start)
		workers.Wait()
		accessibilityRegistry.Lock()
		registered := accessibilityRegistry.windows["ax"] == b || accessibilityRegistry.bridges[b.token] == b
		accessibilityRegistry.Unlock()
		if registered {
			t.Fatal("concurrent refresh retained closed bridge", b.token)
		}
	}
}

func TestAccessibilityLateRetiredBindPreservesReplacement(t *testing.T) {
	a, old := newAXTest(t, 8)
	old.close()
	next := newAccessibilityBridge(a)
	next.bind("ax")
	t.Cleanup(next.close)
	old.bind("ax")
	old.closeFrom("later-cleanup")
	accessibilityRegistry.Lock()
	current := accessibilityRegistry.windows["ax"]
	accessibilityRegistry.Unlock()
	if current != next {
		t.Fatal("retired refresh displaced current attachment")
	}
}

func TestAccessibilityRetirementErrorCopiesFirstClose(t *testing.T) {
	_, b := newAXTest(t, 8)
	b.generation.Store(17)
	b.closeFrom("native-attachment-close")
	b.closeFrom("driver-window-cleanup")
	var closed *AccessibilityClosedError
	err := b.closedError()
	if !errors.Is(err, ErrAccessibilityClosed) || !errors.As(err, &closed) {
		t.Fatal("unclassifiable attachment retirement", err)
	}
	want := AccessibilityRetirement{Window: "ax", Token: b.token, Generation: 17, Source: "native-attachment-close"}
	if closed.Retirement != want {
		t.Fatal("later cleanup overwrote first close", closed.Retirement)
	}
	closed.Retirement.Source = "mutated caller copy"
	var again *AccessibilityClosedError
	if !errors.As(b.closedError(), &again) || again.Retirement != want {
		t.Fatal("caller error mutated retained provenance", again)
	}
	select {
	case <-b.done:
	default:
		t.Fatal("retirement did not cancel requests")
	}
}
func TestAccessibilitySnapshotModalTabsDisabledAndIDNamespaces(t *testing.T) {
	a, b := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("AX headless"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
		{ID: "outside", Kind: ui.Button, Text: "Outside", Activate: func() {}},
		{ID: "blocked", Kind: ui.Column, Disabled: true, Children: []ui.Node{{ID: "disabled", Kind: ui.Button, Activate: func() { t.Fatal("disabled") }}}},
		{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, Change: func(ui.ChangeEvent) {}, Children: []ui.Node{{ID: "a", Kind: ui.Column, Children: []ui.Node{{ID: "visible", Kind: ui.Label, Text: "Visible"}}}, {ID: "b", Kind: ui.Column, Children: []ui.Node{{ID: "hidden", Kind: ui.Label, Text: "Hidden"}}}}},
		{ID: "dialog", Kind: ui.Dialog, Open: false, Children: []ui.Node{{ID: "modal-text", Kind: ui.TextField, Text: "Modal"}}},
	}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	seen := map[string]accessibilityNode{}
	for _, x := range nodes {
		if _, ok := seen[x.ID]; ok {
			t.Fatal("collision", x.ID)
		}
		seen[x.ID] = x
	}
	if _, ok := seen[accessibilityManagedID("hidden")]; ok {
		t.Fatal("inactive tab visible")
	}
	if _, ok := seen[accessibilityManagedID("modal-text")]; ok {
		t.Fatal("closed dialog visible")
	}
	if !seen[accessibilityManagedID("disabled")].Disabled || handlers[accessibilityManagedID("disabled")] != nil {
		t.Fatal("inherited disabled lost")
	}
	if !seen[accessibilityOptionID("tabs", "a")].Selected {
		t.Fatal("selected tab header missing")
	}
	n.Children[3].Open = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, _ = accessibilitySnapshot(r, b)
	if len(nodes) != 2 || nodes[0].DomainID != "dialog" || nodes[1].DomainID != "modal-text" {
		t.Fatalf("modal leaked outside content: %+v", nodes)
	}
	if accessibilityManagedID("row:x:y") == accessibilityRowID("x", "y") || accessibilityOptionID("a/b", "c") == accessibilityOptionID("a", "b/c") {
		t.Fatal("native namespaces collide")
	}
}
func TestAccessibilityControlledActionsValidation(t *testing.T) {
	changes := []ui.ChangeEvent{}
	r := &record{node: ui.Node{ID: "num", Kind: ui.NumberField, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 2}, Change: func(e ui.ChangeEvent) { changes = append(changes, e) }}, control: &controlState{}}
	if !applyAccessibilityAction(r, accessibilityAction{Kind: "increment"}) || changes[0].Number != 6 {
		t.Fatal(changes)
	}
	if !applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: "25"}) || changes[1].Number != 10 {
		t.Fatal(changes)
	}
	for _, value := range []string{"NaN", "1e999", "null", "[2]", "true"} {
		before := len(changes)
		if applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: value}) || len(changes) != before {
			t.Fatal("invalid numeric", value)
		}
	}
	r.control.retired = true
	if applyAccessibilityAction(r, accessibilityAction{Kind: "increment"}) {
		t.Fatal("retired edit")
	}
	r.control.retired = false
	r.control.disabled = true
	if applyAccessibilityAction(r, accessibilityAction{Kind: "increment"}) {
		t.Fatal("disabled edit")
	}
}

type axRows struct{}

func (axRows) Len() int { return 100000 }
func (axRows) Row(i int) virtual.Row {
	return virtual.Row{ID: string(rune('A' + i)), Cells: []string{"Row"}}
}
func TestAccessibilityVirtualRowsBoundedAndStale(t *testing.T) {
	a, b := newAXTest(t, 8)
	model, _ := virtual.NewTable(axRows{}, 4)
	v := &VirtualTable{Owner: a, Window: "ax", Model: model}
	widget := v.Mount(core.NewBody("virtual"))
	if widget == nil {
		t.Fatal("mount")
	}
	if err := v.Sync(widget); err != nil {
		t.Fatal(err)
	}
	r := &record{node: ui.Node{ID: "table", Kind: ui.Extension, Custom: v}, widget: widget}
	nodes, handlers := extensionAccessibility(r, b, false)
	if len(nodes) != 6 {
		t.Fatal("unbounded tree", len(nodes))
	}
	if nodes[2].RowCount != 100000 || nodes[2].RowIndex != 1 {
		t.Fatal("row count/index")
	}
	oldID := nodes[2].ID
	_ = model.SetViewport(10, 4)
	if handlers[oldID](accessibilityAction{Kind: "press"}) {
		t.Fatal("recycled slot selected different domain row")
	}
	accessibilityStampVirtualGeometry(r)
	bounds := accessibilityBounds(r)
	if _, ok := bounds[accessibilityRowID("table", string(rune('A'+10)))]; !ok {
		t.Fatal("domain geometry stamp missing")
	}
}

func TestAccessibilityWindowLifetimeDropsAdmittedWork(t *testing.T) {
	a, b := newAXTest(t, 8)
	if err := a.Open(app.Window{ID: "other", Title: "Other", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	b.mu.Lock()
	b.handlers["button"] = func(accessibilityAction) bool { calls++; return true }
	b.mu.Unlock()
	b.generation.Store(1)
	if !accessibilityDeliver(b.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("admission")
	}
	if err := a.CloseWindow(fixtureWindowRef(a, "ax")); err != nil {
		t.Fatal(err)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("window-scoped late action executed")
	}
	if accessibilityDeliver(b.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("closed-window action admitted")
	}
}
func TestAccessibilitySelectorLabelsAndDisabledIdentity(t *testing.T) {
	selected := ""
	r := &record{node: ui.Node{ID: "selector", Kind: ui.Selector, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "One"}, {ID: "b", Label: "Same"}, {ID: "c", Label: "Same"}, {ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "d", Label: "blocked"}}, Change: func(e ui.ChangeEvent) { selected = e.SelectedID }}, control: &controlState{}}
	if !applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: "One"}) || selected != "a" {
		t.Fatal(selected)
	}
	if !applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: "c"}) || selected != "c" {
		t.Fatal(selected)
	}
	for _, value := range []string{"Same", "blocked", "missing"} {
		if applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: value}) {
			t.Fatal("ambiguous/disabled/absent accepted", value)
		}
	}
}

func TestAccessibilityApplicationFocusPrefersLiteralDomainID(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("AX focus names"))
	encodedA := accessibilityManagedID("a")
	root, err := reconcile(host, nil, ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
		{ID: "a", Kind: ui.TextField},
		{ID: encodedA, Kind: ui.TextField},
	}}, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if got := accessibilityRequestedFocusID(root, encodedA); got != accessibilityManagedID(encodedA) {
		t.Fatalf("literal domain focus redirected to another native identifier: %q", got)
	}
}

func TestAccessibilityIdenticalRefreshKeepsAdmittedActionsValid(t *testing.T) {
	a, b := newAXTest(t, 8)
	calls, retiredCalls := 0, 0
	nodes := []accessibilityNode{{ID: "button", Role: "button", Actions: []string{"press"}}}
	handlers := map[string]accessibilityHandler{"button": func(accessibilityAction) bool { retiredCalls++; return true }}
	b.update(nodes, handlers)
	generation := b.generation.Load()
	if !accessibilityDeliver(b.token, generation, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("admission failed")
	}
	b.update(nodes, map[string]accessibilityHandler{"button": func(accessibilityAction) bool { calls++; return true }})
	if b.generation.Load() != generation {
		t.Fatal("identical refresh retired native action")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || retiredCalls != 0 {
		t.Fatalf("current callback dispatch incorrect: current=%d retired=%d", calls, retiredCalls)
	}
	nodes[0].Disabled = true
	b.update(nodes, nil)
	if b.generation.Load() == generation || accessibilityDeliver(b.token, generation, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("actual semantic change did not retire native action")
	}
}

func TestAccessibilityNumericDescriptionsRejectNonfiniteBeforeSnapshot(t *testing.T) {
	a, b := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("AX numeric validation"))
	for _, kind := range []ui.Kind{ui.NumberField, ui.Slider} {
		for _, bad := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
			for field := 0; field < 4; field++ {
				n := ui.Node{ID: "numeric", Kind: kind, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}}
				switch field {
				case 0:
					n.Number = bad
				case 1:
					n.Numeric.Min = bad
				case 2:
					n.Numeric.Max = bad
				case 3:
					n.Numeric.Step = bad
				}
				if ui.Validate(n) == nil {
					t.Fatalf("%s admitted nonfinite numeric field %d", kind, field)
				}
				if _, err := reconcile(host, nil, n, a, "ax"); err == nil {
					t.Fatal("control mount admitted nonfinite descriptor")
				}
			}
		}
	}
	n := ui.Node{ID: "numeric", Kind: ui.NumberField, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, _ := accessibilitySnapshot(r, b)
	if _, err = json.Marshal(nodes); err != nil {
		t.Fatal("validated descriptor is not serializable", err)
	}
}
func TestAccessibilityBoundsAllowsUnbuiltTabsAndMissingDockWidgets(t *testing.T) {
	r := &record{node: ui.Node{ID: "tabs", Kind: ui.Tabs, Options: []ui.Option{{ID: "page"}}}}
	_ = accessibilityBounds(r)
	r.control = &controlState{}
	_ = accessibilityBounds(r)
	w := &Workbench{layoutSnapshot: &workbench.Group{ID: "leaf", Tabs: []string{"panel"}, Active: "panel"}}
	_ = accessibilityBounds(&record{node: ui.Node{ID: "dock", Kind: ui.Extension, Custom: w}})
}
func TestAccessibilityVirtualTreePagingTargetsAndStableBounds(t *testing.T) {
	a, b := newAXTest(t, 8)
	model, err := virtual.NewTree([]virtual.TreeNode{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}, {ID: "three", Label: "Three"}, {ID: "four", Label: "Four"}})
	if err != nil {
		t.Fatal(err)
	}
	v := &VirtualTree{Owner: a, Window: "ax", Model: model, Count: 2}
	host := core.NewFrame(core.NewBody("AX outline pager"))
	r, err := reconcile(host, nil, ui.Node{ID: "outline", Kind: ui.Extension, Custom: v}, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	accessibilityStampVirtualGeometry(r)
	bounds := accessibilityBounds(r)
	if !reflect.DeepEqual(nativeAccessibilityBounds(r), bounds) {
		t.Fatal("ordinary legacy tree projection changed")
	}
	for _, page := range []string{"previous", "next"} {
		if _, ok := bounds[accessibilityPageID("outline", page)]; !ok {
			t.Fatal("actual outline pager target missing", page)
		}
	}
	_, handlers := accessibilitySnapshot(r, b)
	if !handlers[accessibilityPageID("outline", "next")](accessibilityAction{Kind: "press"}) || v.Start != 2 {
		t.Fatal("outline next page")
	}
	if err := v.Sync(r.widget); err != nil {
		t.Fatal(err)
	}
	accessibilityStampVirtualGeometry(r)
	bounds = accessibilityBounds(r)
	if !reflect.DeepEqual(nativeAccessibilityBounds(r), bounds) {
		t.Fatal("paged legacy tree projection changed")
	}
	if _, ok := bounds[accessibilityRowID("outline", "three")]; !ok {
		t.Fatal("recycled outline row has wrong domain bounds")
	}
	if !handlers[accessibilityPageID("outline", "previous")](accessibilityAction{Kind: "press"}) || v.Start != 0 {
		t.Fatal("outline previous page")
	}
}

func TestAccessibilityDialogKeyboardBindingIgnoresUnbuiltWidgets(t *testing.T) {
	body := core.NewBody("AX nil-safe modal")
	bindAccessibilityDialogKeyboard(nil, body)
	bindAccessibilityDialogKeyboard(&record{}, body)
	var absent *core.Frame
	bindAccessibilityDialogKeyboard(&record{widget: absent}, body)
	widget := core.NewFrame(body)
	widget.AsTree().Parent = absent
	bindAccessibilityDialogKeyboard(&record{widget: widget}, body)
	bindAccessibilityDialogKeyboard(&record{widget: widget}, nil)
	bindAccessibilityDialogKeyboard(&record{widget: widget}, &core.Body{})
}

func TestAccessibilityTreeNilHandlesConcreteAndRetiredNodes(t *testing.T) {
	var frame *core.Frame
	var tabs *core.Tabs
	if !accessibilityTreeNil(frame) || !accessibilityTreeNil(tabs) {
		t.Fatal("typed-nil widget survived guard")
	}
	widget := core.NewFrame(core.NewBody("AX retired widget"))
	if accessibilityTreeNil(widget) {
		t.Fatal("live widget considered nil")
	}
	widget.AsTree().This = nil
	if !accessibilityTreeNil(widget) {
		t.Fatal("retired widget survived guard")
	}
}

type axSmallRows int

func (r axSmallRows) Len() int { return int(r) }
func (r axSmallRows) Row(i int) virtual.Row {
	return virtual.Row{ID: fmt.Sprintf("small-%d", i), Cells: []string{"Small"}}
}
func TestAccessibilityTablePagingReachesTrailingRowsAndRejectsStaleCount(t *testing.T) {
	a, b := newAXTest(t, 8)
	model, err := virtual.NewTable(axSmallRows(6), 4)
	if err != nil {
		t.Fatal(err)
	}
	v := &VirtualTable{Owner: a, Window: "ax", Model: model}
	if native := v.Mount(core.NewBody("AX semantic fixture")); native == nil {
		t.Fatal("mount")
	} else if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
	r := &record{node: ui.Node{ID: "small", Kind: ui.Extension, Custom: v}}
	_, handlers := accessibilitySnapshot(r, b)
	next := handlers[accessibilityPageID("small", "next")]
	if !next(accessibilityAction{Kind: "press"}) || model.StartRow() != 2 {
		t.Fatal("trailing page not clamped")
	}
	view := model.View()
	if len(view) != 4 || view[len(view)-1].ID != "small-5" {
		t.Fatal("trailing rows unreachable", view)
	}
	if !next(accessibilityAction{Kind: "press"}) || model.StartRow() != 2 {
		t.Fatal("last-page clamp is not stable")
	}
	if !handlers[accessibilityPageID("small", "previous")](accessibilityAction{Kind: "press"}) || model.StartRow() != 0 {
		t.Fatal("previous clamp")
	}
	if err = model.SetViewport(0, 2); err != nil {
		t.Fatal(err)
	}
	if next(accessibilityAction{Kind: "press"}) || model.VisibleCount() != 2 {
		t.Fatal("stale viewport count overwritten")
	}
}
func TestAccessibilityTreeSelectAndDisclosureAreIndependentAndCurrent(t *testing.T) {
	a, b := newAXTest(t, 8)
	model, err := virtual.NewTree([]virtual.TreeNode{{ID: "parent", Label: "Parent"}, {ID: "child", Parent: "parent", Label: "Child"}, {ID: "sibling", Label: "Sibling"}})
	if err != nil {
		t.Fatal(err)
	}
	selected := 0
	v := &VirtualTree{Owner: a, Window: "ax", Model: model, Count: 3, Select: func(string) { selected++ }}
	if native := v.Mount(core.NewBody("AX semantic fixture")); native == nil {
		t.Fatal("mount")
	} else if err := v.Sync(native); err != nil {
		t.Fatal(err)
	}
	r := &record{node: ui.Node{ID: "tree", Kind: ui.Extension, Custom: v}}
	nodes, handlers := accessibilitySnapshot(r, b)
	var row accessibilityNode
	for _, n := range nodes {
		if n.DomainID == "parent" {
			row = n
		}
	}
	if !containsAccessibilityAction(row.Actions, "expand") || !containsAccessibilityAction(row.Actions, "collapse") {
		t.Fatal("disclosure permissions missing")
	}
	handler := handlers[accessibilityRowID("tree", "parent")]
	if !handler(accessibilityAction{Kind: "expand"}) || model.Selected != "" || selected != 0 {
		t.Fatal("disclosure changed selection")
	}
	if !handler(accessibilityAction{Kind: "press"}) || model.Selected != "parent" || selected != 1 {
		t.Fatal("selection failed")
	}
	current, _ := model.View(0, 3)
	if !current[0].Expanded {
		t.Fatal("selection collapsed expanded row")
	}
	if !handler(accessibilityAction{Kind: "collapse"}) || model.Selected != "parent" || selected != 1 {
		t.Fatal("collapse changed selection")
	}
	if !handler(accessibilityAction{Kind: "collapse"}) {
		t.Fatal("idempotent collapse")
	}
	v.Model, err = virtual.NewTree([]virtual.TreeNode{{ID: "parent", Label: "Now leaf"}})
	if err != nil {
		t.Fatal(err)
	}
	if handler(accessibilityAction{Kind: "expand"}) {
		t.Fatal("captured child status toggled replaced leaf")
	}
	v.Start = 1
	if handler(accessibilityAction{Kind: "press"}) {
		t.Fatal("offscreen domain row selected")
	}
}

func TestAccessibilityDialogInitialFocusBeforeDeferredShow(t *testing.T) {
	a, _ := newAXTest(t, 8)
	body := core.NewBody("AX modal focus origin")
	descriptor := ui.Node{ID: "modal", Kind: ui.Dialog, Open: true, Children: []ui.Node{
		{ID: "disabled", Kind: ui.TextField, Disabled: true},
		{ID: "inner", Kind: ui.Dialog, Open: true, Children: []ui.Node{{ID: "inner-field", Kind: ui.TextField}}},
		{ID: "first", Kind: ui.TextField},
		{ID: "later", Kind: ui.Button},
	}}
	record, err := reconcile(body, nil, descriptor, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nativeBody := prepareDialogBody(record)
	first := record.children[2].widget
	if got := accessibilityDialogInitialFocus(record, nativeBody); got != first {
		t.Fatalf("first explicit modal target=%v want %v", got, first)
	}
	if !record.children[1].node.Open {
		t.Fatal("focus projection mutated nested descriptor")
	}
	record.children[2].node.Disabled = true
	if got := accessibilityDialogInitialFocus(record, nativeBody); got != record.children[3].widget {
		t.Fatal("disabled target selected")
	}
	record.children[3].node.Disabled = true
	if got := accessibilityDialogInitialFocus(record, nativeBody); got != nativeBody {
		t.Fatal("empty managed scope lacks body Escape anchor")
	}
	record.node.Disabled = true
	if got := accessibilityDialogInitialFocus(record, nativeBody); got != nil {
		t.Fatal("disabled modal stole start focus")
	}
	record.node.Disabled = false
	record.node.Open = false
	if got := accessibilityDialogInitialFocus(record, nativeBody); got != nil {
		t.Fatal("closed modal stole start focus")
	}
}

func TestAccessibilityFieldSetRefusesNativeBudgetAndNULBeforeChange(t *testing.T) {
	a, _ := newAXTest(t, 8)
	body := core.NewBody("AX field admission")
	var field *record
	calls := 0
	n := ui.Node{ID: "field", Kind: ui.TextField, Text: "accepted", Change: func(e ui.ChangeEvent) {
		calls++
		field.node.Text = e.Text
	}}
	var err error
	field, err = reconcile(body, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { disposeControlTree(field) })
	before := field.widget.(*core.TextField).EditingState()
	if err := a.Do(func() {
		for _, value := range []string{strings.Repeat("x", textinput.MaxDocumentBytes+1), "prefix\x00suffix"} {
			if applyAccessibilityAction(field, accessibilityAction{Kind: "set", Value: value}) {
				t.Fatal("invalid native field value accepted")
			}
			if calls != 0 || field.node.Text != "accepted" || field.control.committed || field.widget.(*core.TextField).EditingState() != before {
				t.Fatal("refused AX value changed controlled or native state")
			}
		}
		if !applyAccessibilityAction(field, accessibilityAction{Kind: "set", Value: "valid🙂"}) || calls != 1 || field.node.Text != "valid🙂" {
			t.Fatal("valid AX value failed after refusal")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
