//go:build offscreen

package cogent

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
)

func TestAccessibilityTextPublicationEmptyAndRetiredBridge(t *testing.T) {
	sends := 0
	send := func(uint64, string) { sends++ }
	bridge := &accessibilityBridge{token: 17}
	bridge.publishTextGeometryTo(nil, send)
	bridge.publishTextGeometryTo(map[string]any{}, send)
	(*accessibilityBridge)(nil).publishTextGeometryTo(map[string]any{"field": "value"}, send)
	bridge.closed.Store(true)
	bridge.publishTextGeometryTo(map[string]any{"field": "value"}, send)
	if sends != 0 {
		t.Fatal("empty or retired bridge enqueued native work", sends)
	}
}

func TestAccessibilityTextPublicationEncodingRefusalIsAtomic(t *testing.T) {
	bridge := &accessibilityBridge{token: 17}
	bridge.publishTextGeometryTo(map[string]any{"good": "value", "invalid": math.NaN()}, func(uint64, string) { t.Fatal("refused text geometry was published") })
}

func TestAccessibilityTextPublicationActualClientsAndNamespaces(t *testing.T) {
	source := &nativeGridSource{rows: 10000, revision: 1}
	grid, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c0", Width: 121}}, virtual.GridConfig{RowHeight: 23, HeaderHeight: 27}, virtual.GridViewport{Width: 300, Height: 180})
	if err != nil {
		t.Fatal(err)
	}
	showField := false
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		children := []ui.Node{{ID: "label", Kind: ui.Label, Text: "Numeric table"}, {ID: "g", Kind: ui.DataGrid, Grid: grid, Grow: true}}
		if showField {
			children = append(children, ui.Node{ID: "field", Kind: ui.TextField, Text: "A😀é"})
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: children}
	})
	gridFrame(t, m)
	bridge := &accessibilityBridge{token: 17}
	sends := 0
	var published string
	send := func(token uint64, payload string) {
		sends++
		if token != 17 {
			t.Fatal("wrong bridge", token)
		}
		published = payload
	}
	check := func(want int) {
		t.Helper()
		if err := a.Do(func() {
			geometry := accessibilityTextGeometry(m.implementation.find("one").state.root)
			if len(geometry) != want {
				t.Fatalf("actual client count=%d want=%d", len(geometry), want)
			}
			before := sends
			bridge.publishTextGeometryTo(geometry, send)
			if want == 0 {
				if sends != before {
					t.Fatal("text-free frame sent native geometry")
				}
			} else {
				// Compare complete entry values, including client identity, Unicode whole
				// text and measured UTF-16 geometry; every nonempty update still publishes.
				expected, err := json.Marshal(geometry)
				if err != nil || sends != before+1 || published != string(expected) {
					t.Fatal("nonempty geometry changed or was withheld", err)
				}
			}
		}); err != nil {
			t.Fatal(err)
		}
	}
	check(0)
	if err := a.Do(func() { showField = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	check(1)
	check(1) // There is no new payload cache or suppression of live text clients.
	if err := a.Do(func() {
		root := m.implementation.find("one").state.root
		layout := workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "group", Tabs: []string{"panel"}, Active: "panel"}}
		workspace, err := workbench.New(layout, []string{"panel"})
		if err != nil {
			t.Fatal(err)
		}
		w := &Workbench{Workspace: workspace, layoutSnapshot: workspace.Layout().Root, records: map[string]*record{"panel": root}}
		dock := &record{node: ui.Node{ID: "dock", Kind: ui.Extension, Custom: w}}
		geometry := accessibilityTextGeometry(dock)
		id := accessibilityWorkbenchNamespace("dock", "panel") + accessibilityManagedID("field")
		if len(geometry) != 1 || geometry[id] == nil {
			t.Fatal("text-client namespace lost", geometry)
		}
		bridge.publishTextGeometryTo(geometry, send)
		expected, _ := json.Marshal(geometry)
		if published != string(expected) {
			t.Fatal("namespaced entry changed")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.Do(func() { showField = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	check(0)
}
