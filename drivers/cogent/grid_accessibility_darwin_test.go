//go:build darwin && cgo && offscreen

package cogent

import (
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestManagedGridNativeAccessibilitySelectors(t *testing.T) {
	g, _ := testNativeGrid(t, 1000000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	inspection := gridInspection(t, a, g)
	var nodes []accessibilityNode
	if e := a.Do(func() { root := m.implementation.find("one").state.root; nodes, _ = accessibilitySnapshot(root, nil) }); e != nil {
		t.Fatal(e)
	}
	visible := inspection.Cells[0]
	for _, c := range inspection.Cells {
		if !c.Bounds.Empty() {
			visible = c
			break
		}
	}
	// Select through the owner model, then capture a completed projection.
	if e := a.Do(func() {
		binding := m.implementation.find("one").state.root.control.grid
		for _, c := range binding.projection.Cells {
			if c.RowKey == visible.RowKey && c.ColumnID == visible.ColumnID {
				if e := g.Activate(c.Ref); e != nil {
					t.Error(e)
				}
				break
			}
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if e := a.Do(func() { nodes, _ = accessibilitySnapshot(m.implementation.find("one").state.root, nil) }); e != nil {
		t.Fatal(e)
	}
	payload, e := json.Marshal(map[string]any{"generation": 1, "nodes": nodes})
	if e != nil {
		t.Fatal(e)
	}
	query := func(column, row int) map[string]any {
		t.Helper()
		var out map[string]any
		if e := json.Unmarshal([]byte(nativeAccessibilityGridFixture(string(payload), accessibilityManagedID("g"), column, row)), &out); e != nil {
			t.Fatal(e)
		}
		return out
	}
	actual := query(visible.Column, visible.Row)
	if actual["role"] != "AXTable" || actual["rows"] != float64(1000000) || actual["columns"] != float64(20) || actual["cell"] != accessibilityManagedID(visible.ID) || actual["cell_role"] != "AXCell" || actual["row"] != float64(visible.Row) || actual["column"] != float64(visible.Column) || actual["selected_cells"] != float64(1) {
		t.Fatal("native table selector mismatch", actual)
	}
	if actual["visible_cells"].(float64) > float64(len(inspection.Cells)) || actual["visible_columns"].(float64) > float64(len(inspection.Headers)) {
		t.Fatal("native AX materialized hidden cells")
	}
	hidden := query(19, 999999)
	if hidden["cell"] != "" {
		t.Fatal("hidden native cell lookup fabricated a cell", hidden)
	}
}
