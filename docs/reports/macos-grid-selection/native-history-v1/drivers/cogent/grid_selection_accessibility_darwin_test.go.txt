//go:build darwin && cgo && offscreen

package cogent

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestManagedGridRowSelectionCocoaSelectedCells(t *testing.T) {
	g, _ := selectionNativeGrid(t, 1000000, true, 0)
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(e ui.GridEvent) {
			if e.Reason == ui.GridSelection {
				changes++
			}
		}}
	})
	gridFrame(t, m)
	var modelErr error
	if e := a.Do(func() {
		modelErr = g.SelectAll()
		if modelErr == nil {
			modelErr = a.Invalidate("one")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if modelErr != nil {
		t.Fatal(modelErr)
	}
	gridFrame(t, m)
	inspection := gridInspection(t, a, g)
	var visible GridCellGeometry
	visibleCount := 0
	for _, c := range inspection.Cells {
		if !c.Bounds.Empty() {
			visibleCount++
			if visible.ID == "" {
				visible = c
			}
			if !c.Selected {
				t.Fatal("All omitted visible member", c)
			}
		}
	}
	var nodes []accessibilityNode
	var handlers map[string]accessibilityHandler
	snapshot := func() {
		t.Helper()
		if e := a.Do(func() { nodes, handlers = accessibilitySnapshot(m.implementation.find("one").state.root, nil) }); e != nil {
			t.Fatal(e)
		}
	}
	query := func() map[string]any {
		t.Helper()
		payload, e := json.Marshal(map[string]any{"generation": 1, "nodes": nodes})
		if e != nil {
			t.Fatal(e)
		}
		var out map[string]any
		if e = json.Unmarshal([]byte(nativeAccessibilityGridFixture(string(payload), accessibilityManagedID("g"), visible.Column, visible.Row)), &out); e != nil {
			t.Fatal(e)
		}
		return out
	}
	snapshot()
	actual := query()
	if actual["rows"] != float64(1000000) || actual["columns"] != float64(20) || actual["selected_cells"] != float64(visibleCount) || actual["visible_cells"] != float64(visibleCount) {
		t.Fatal("native selected-cell selector materialized hidden rows or omitted selected members", actual, visibleCount)
	}
	for _, node := range nodes {
		if node.Role == "grid-row" && !node.Selected {
			t.Fatal("AX row membership omitted", node)
		}
	}
	press := handlers[accessibilityManagedID(visible.ID)]
	if press == nil {
		t.Fatal("visible AX press missing")
	}
	var accepted bool
	if e := a.Do(func() { accepted = press(accessibilityAction{Kind: "press"}) }); e != nil {
		t.Fatal(e)
	}
	if !accepted || changes != 1 {
		t.Fatal("AX plain press did not replace once", accepted, changes)
	}
	gridFrame(t, m)
	inspection = gridInspection(t, a, g)
	selectedCount := 0
	for _, c := range inspection.Cells {
		if !c.Bounds.Empty() && c.Selected {
			selectedCount++
			if c.RowKey != visible.RowKey {
				t.Fatal("AX replace retained another row", c)
			}
		}
	}
	snapshot()
	actual = query()
	if selectedCount < 2 || actual["selected_cells"] != float64(selectedCount) {
		t.Fatal("AX row selection collapsed to active-cell selection", actual, selectedCount)
	}
	// An action from the preceding accepted projection must remain stale even
	// though the target's stable row key is still present in the native pool.
	var refused error
	if e := a.Do(func() {
		m.implementation.find("one").state.root.control.onError = func(e error) { refused = e }
		accepted = press(accessibilityAction{Kind: "press"})
	}); e != nil {
		t.Fatal(e)
	}
	if !errors.Is(refused, virtual.ErrGridStale) {
		t.Fatal("stale AX action error not reported", refused)
	}
	if accepted || changes != 1 {
		t.Fatal("retired AX press selected a reused row", accepted, changes)
	}
}
