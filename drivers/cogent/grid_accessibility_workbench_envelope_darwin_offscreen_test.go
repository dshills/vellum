//go:build darwin && cgo && offscreen

package cogent

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func TestNativeGridAccessibilityDisabledWorkbenchProductionEncoding(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true}
	})
	gridFrame(t, m)
	measured := gridInspection(t, a, g)
	if len(measured.Cells) == 0 || measured.Generation == 0 {
		t.Fatal("fixture has no actual completed Grid frame")
	}
	layout := workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "group", Tabs: []string{"panel"}, Active: "panel"}}
	workspace, err := workbench.New(layout, []string{"panel"})
	if err != nil {
		t.Fatal(err)
	}
	var nodes []accessibilityNode
	var handlers map[string]accessibilityHandler
	var payload []byte
	var marshalErr error
	if err := a.Do(func() {
		// Manual intentionally refuses Extension descriptors. Exercise the real
		// Workbench semantic adapter with an actually rendered managed Grid as
		// its active panel record, without claiming manual Workbench routing.
		panel := m.implementation.find("one").state.root
		w := &Workbench{Owner: a, Window: "one", Workspace: workspace, layoutSnapshot: workspace.Layout().Root,
			records: map[string]*record{"panel": panel}, Panels: []Panel{{ID: "panel", Title: "Grid panel"}}}
		root := &record{node: ui.Node{ID: "workbench", Kind: ui.Extension, Custom: w, Disabled: true}}
		bridge := &accessibilityBridge{owner: a}
		nodes, handlers = accessibilitySnapshot(root, bridge)
		nodes, handlers = bridge.prepareGridPublication(nodes, handlers)
		payload, marshalErr = json.Marshal(struct {
			Generation uint64              `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, nodes})
	}); err != nil {
		t.Fatal(err)
	}
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if len(handlers) != 0 {
		t.Fatal("disabled Workbench retained action handlers")
	}
	prefix := accessibilityWorkbenchNamespace("workbench", "panel")
	scope := prefix + accessibilityManagedID("g")
	cells := 0
	for _, n := range nodes {
		if !n.Disabled || n.Actions == nil || len(n.Actions) != 0 {
			t.Fatalf("disabled node did not encode an empty action array: %+v", n)
		}
		if n.GridScope == "" {
			continue
		}
		if n.GridScope != scope || !strings.HasPrefix(n.ID, prefix) || n.GridPending || n.GridGeneration != measured.Generation || n.GridTargetGeneration != measured.Generation || n.GridFrame == nil {
			t.Fatalf("completed namespaced Grid envelope changed: %+v", n)
		}
		if n.Role == "grid-cell" {
			cells++
		}
	}
	if cells == 0 {
		t.Fatal("actual completed Grid cells missing from Workbench publication")
	}
	if !nativeAccessibilitySnapshotValid(string(payload)) {
		t.Fatalf("production disabled Workbench encoding refused by native validator: %s", payload)
	}
}
