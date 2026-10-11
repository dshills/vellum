//go:build offscreen

package cogent

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestReviewGridNamespacedPressFocusesQualifiedHost(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	ins := gridInspection(t, a, g)
	visible := ""
	for _, c := range ins.Cells {
		if !c.Bounds.Empty() {
			visible = c.ID
			break
		}
	}
	if visible == "" {
		t.Fatal("no actual visible cell")
	}
	prefix := accessibilityWorkbenchNamespace("workbench", "panel")
	bridge := &accessibilityBridge{owner: a}
	if err := a.Do(func() {
		root := m.implementation.find("one").state.root
		nodes, handlers := accessibilitySnapshot(root, bridge)
		_, wrapped := namespaceAccessibility(nodes, handlers, prefix, bridge)
		press := wrapped[prefix+accessibilityManagedID(visible)]
		if press == nil || !press(accessibilityAction{Kind: "press"}) {
			t.Fatal("valid cell press refused")
		}
		pending := bridge.pendingFocus.Load()
		if pending == nil || *pending != prefix+accessibilityManagedID("g") {
			t.Errorf("namespaced Grid cell press produced unreachable host focus %v; want %q", pending, prefix+accessibilityManagedID("g"))
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestManagedGridNamespaceDoesNotRewriteOrdinaryPressFocus(t *testing.T) {
	bridge := &accessibilityBridge{}
	nodes := []accessibilityNode{{ID: "ordinary", Role: "button", Actions: []string{"press"}}}
	handlers := map[string]accessibilityHandler{"ordinary": func(action accessibilityAction) bool { accessibilityFocus(bridge, "external-destination"); return true }}
	_, wrapped := namespaceAccessibility(nodes, handlers, "scope:", bridge)
	if !wrapped["scope:ordinary"](accessibilityAction{Kind: "press"}) {
		t.Fatal("ordinary press refused")
	}
	pending := bridge.pendingFocus.Load()
	if pending == nil || *pending != "external-destination" {
		t.Fatal("ordinary press focus was rewritten", pending)
	}
}
