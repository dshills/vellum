//go:build offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
)

// Derive the omitted IDs from the production semantic projection rather than
// from grid ID spelling or a second implementation of the bounds traversal.
func assertNativeGeometryProjection(t *testing.T, root *record) {
	t.Helper()
	nodes, _ := accessibilitySnapshot(root, nil)
	full := accessibilityBounds(root)
	want := make(map[string][4]float32, len(full))
	for id, rect := range full {
		want[id] = rect
	}
	omitted := 0
	for _, node := range nodes {
		if node.GridScope != "" {
			if _, ok := full[node.ID]; !ok {
				t.Errorf("inspection lost scoped bounds for %s", node.ID)
			}
			delete(want, node.ID)
			omitted++
		}
	}
	if omitted == 0 {
		t.Fatal("fixture has no scoped grid semantics")
	}
	got := nativeAccessibilityBounds(root)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("native geometry differs from full projection minus scoped IDs: got %v want %v", got, want)
	}
	before, _ := json.Marshal(full)
	after, _ := json.Marshal(accessibilityBounds(root))
	if string(before) != string(after) {
		t.Fatal("native projection mutated inspection geometry")
	}
	filtered, _ := json.Marshal(got)
	t.Logf("native geometry: full=%d bytes/%d IDs, unscoped=%d bytes/%d IDs", len(before), len(full), len(filtered), len(got))
}

func TestNativeAccessibilityGeometryOmitsScopedGridOnly(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			source := &inlineSource{nativeGridSource: nativeGridSource{rows: 10000, revision: 1}, text: "A😀é", version: 1}
			g, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c0", Width: 121}, {ID: "c1", Width: 121}, {ID: "c2", Width: 121}}, virtual.GridConfig{RowHeight: 23, HeaderHeight: 27, FrozenColumns: 1, CellEditing: true, MaxEditBytes: 32}, virtual.GridViewport{Width: 300, Height: 180})
			if err != nil {
				t.Fatal(err)
			}
			disabled := false
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
					{ID: "toolbar", Kind: ui.Row, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Tools"}, {ID: "button", Kind: ui.Button, Text: "Apply", Activate: func() {}}, {ID: "field", Kind: ui.TextField, Text: "Filter"}}},
					{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, Disabled: disabled},
				}}
			})
			gridFrame(t, m)
			root := m.implementation.find("one").state.root
			check := func(phase string) {
				t.Helper()
				if err := a.Do(func() {
					t.Log(phase)
					assertNativeGeometryProjection(t, root)
					bounds := nativeAccessibilityBounds(root)
					for _, id := range []string{"root", "toolbar", "label", "button", "field"} {
						if _, ok := bounds[accessibilityManagedID(id)]; !ok {
							t.Error("ordinary control geometry missing", id)
						}
					}
				}); err != nil {
					t.Fatal(err)
				}
			}
			check("ready")
			if err := a.Do(func() {
				snap, e := g.Snapshot()
				if e != nil {
					t.Error(e)
					return
				}
				for _, cell := range snap.Cells {
					if cell.RowKey == "row-0" && cell.ColumnID == "c1" {
						if e := g.BeginEdit(cell.Ref); e != nil {
							t.Error(e)
						}
						break
					}
				}
				_ = a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			check("editing")
			if err := a.Do(func() {
				edit := gridEditingRecord(root.children[1])
				if edit == nil {
					t.Error("private editor not mounted")
					return
				}
				if _, ok := accessibilityTextGeometry(root)[accessibilityManagedID(edit.node.ID)]; !ok {
					t.Error("separate native text geometry lost private editor")
				}
				layout := workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "group", Tabs: []string{"panel"}, Active: "panel"}}
				workspace, e := workbench.New(layout, []string{"panel"})
				if e != nil {
					t.Error(e)
					return
				}
				w := &Workbench{Workspace: workspace, layoutSnapshot: workspace.Layout().Root, records: map[string]*record{"panel": root}}
				dock := &record{node: ui.Node{ID: "dock", Kind: ui.Extension, Custom: w}}
				assertNativeGeometryProjection(t, dock)
				prefix := accessibilityWorkbenchNamespace("dock", "panel")
				if _, ok := nativeAccessibilityBounds(dock)[prefix+accessibilityManagedID("button")]; !ok {
					t.Error("namespaced ordinary control geometry missing")
				}
				if e := g.Scroll(80, 10); e != nil {
					t.Error(e)
				}
				if e := syncGrid(root.children[1]); e != nil {
					t.Error(e)
				}
				assertNativeGeometryProjection(t, root) // owner accepted, frame pending
				_ = a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			check("scrolled")
			if err := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			check("disabled")
		})
	}
}
