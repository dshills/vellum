//go:build darwin && cgo

package cogent

import (
	"cogentcore.org/core/core"
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"reflect"
	"testing"
)

func TestListBoxCocoaSelectedRowsSetterOwnerBridge(t *testing.T) {
	for _, variant := range []string{"valid", "same", "zero", "multiple", "disabled", "readonly", "unknown", "foreign", "stale-admission", "stale-delivery", "row-aba"} {
		t.Run(variant, func(t *testing.T) {
			a, b := newAXTest(t, 16)
			calls := 0
			selected := ""
			n := listBoxDescriptor(func(e ui.ChangeEvent) { calls++; selected = e.SelectedID })
			if variant == "readonly" {
				n.Change = nil
			}
			host := core.NewFrame(core.NewBody("list setter"))
			r, err := reconcile(host, nil, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			nodes, handlers := accessibilitySnapshot(r, b)
			nodes = append(nodes, accessibilityNode{ID: "foreign", Role: "list-option", DomainID: "b", Actions: []string{}})
			b.mu.Lock()
			b.handlers = handlers
			b.activationGuards = accessibilityActivationGuards(nodes)
			b.mu.Unlock()
			b.generation.Store(1)
			rows := []string{accessibilityOptionID("list", "b")}
			admit, wantCalls := true, 1
			switch variant {
			case "same":
				rows = []string{accessibilityOptionID("list", "a")}
			case "zero":
				rows = []string{}
				admit = false
				wantCalls = 0
			case "multiple":
				rows = append(rows, accessibilityOptionID("list", "a"))
				admit = false
				wantCalls = 0
			case "disabled":
				rows = []string{accessibilityOptionID("list", "c")}
				admit = false
				wantCalls = 0
			case "readonly", "unknown", "foreign", "stale-admission":
				admit = false
				wantCalls = 0
				if variant == "unknown" {
					rows = []string{"missing"}
				}
				if variant == "foreign" {
					rows = []string{"foreign"}
				}
				if variant == "stale-admission" {
					b.generation.Store(2)
				}
			case "stale-delivery", "row-aba":
				wantCalls = 0
			}
			steps := []map[string]any{{"kind": "update", "snapshot": map[string]any{"generation": 1, "nodes": nodes}}, {"kind": "select_rows", "token": b.token, "list": accessibilityManagedID("list"), "rows": rows}}
			if variant == "valid" {
				for _, invalid := range [][]string{{}, {accessibilityOptionID("list", "a"), accessibilityOptionID("list", "b")}, {accessibilityOptionID("list", "c")}, {"foreign"}, {"missing"}} {
					steps = append(steps, map[string]any{"kind": "select_rows", "token": b.token, "list": accessibilityManagedID("list"), "rows": invalid})
				}
			}
			payload, _ := json.Marshal(steps)
			raw := nativeAccessibilityScopedFixture(string(payload))
			var out []axEnvelopeResult
			if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) != len(steps) || out[1].Accepted != admit {
				t.Fatal("Cocoa setter admission", raw, err)
			}
			for _, result := range out[2:] {
				if result.Accepted {
					t.Fatal("fixture retained a prior setter admission after refusal")
				}
			}
			if variant == "stale-delivery" {
				b.generation.Add(1)
			}
			if variant == "row-aba" {
				n.Options[1].Disabled = true
				if _, err := reconcile(host, r, n, a, "ax"); err != nil {
					t.Fatal(err)
				}
				n.Options[1].Disabled = false
				if _, err := reconcile(host, r, n, a, "ax"); err != nil {
					t.Fatal(err)
				}
			}
			a.Drain()
			if calls != wantCalls {
				t.Fatal("Cocoa setter bypassed owner/lifetime", variant, calls, selected)
			}
			if calls > 0 && selected != map[bool]string{true: "a", false: "b"}[variant == "same"] {
				t.Fatal("Cocoa setter changed stable target", selected)
			}
		})
	}
}

func TestListBoxCocoaRowsWithoutWindow(t *testing.T) {
	nodes := []map[string]any{{"id": "list", "role": "list-box", "name": "Items", "row_count": 3, "value": "Beta", "actions": []string{"set"}}, {"id": "a", "parent": "list", "role": "list-option", "row_index": 1, "actions": []string{}}, {"id": "b", "parent": "list", "role": "list-option", "row_index": 2, "selected": true, "actions": []string{"press"}}, {"id": "c", "parent": "list", "role": "list-option", "row_index": 3, "actions": []string{}}}
	steps := []map[string]any{axEnvelopeUpdate(1, nodes), {"kind": "geometry", "bounds": map[string][4]float64{"list": {10, 20, 200, 80}, "a": {10, 20, 200, 30}, "b": {10, 50, 200, 30}, "c": {0, 0, 0, 0}}}, {"kind": "geometry", "bounds": map[string][4]float64{"a": {0, 0, 0, 0}, "b": {0, 0, 0, 0}, "c": {10, 20, 200, 30}}}}
	payload, _ := json.Marshal(steps)
	raw := nativeAccessibilityScopedFixture(string(payload))
	var out []struct {
		Elements []struct {
			ID, Role, Value  string
			Index, Rows      int
			VisibleRows      []string `json:"visible_rows"`
			VisibleChildren  []string `json:"visible_children"`
			SelectedRows     []string `json:"selected_rows"`
			SelectedChildren []string `json:"selected_children"`
			SetRowsAllowed   bool     `json:"set_rows_allowed"`
		}
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) != 3 {
		t.Fatal(raw, err)
	}
	for i, ids := range [][]string{nil, {"a", "b"}, {"c"}} {
		for _, e := range out[i].Elements {
			if e.ID == "list" {
				if e.Role != "AXList" || e.Rows != 3 || e.Value != "Beta" || !reflect.DeepEqual(e.VisibleRows, ids) && !(len(ids) == 0 && len(e.VisibleRows) == 0) || !reflect.DeepEqual(e.VisibleRows, e.VisibleChildren) || !reflect.DeepEqual(e.SelectedRows, []string{"b"}) || !reflect.DeepEqual(e.SelectedRows, e.SelectedChildren) || !e.SetRowsAllowed {
					t.Fatal(i, e)
				}
			} else if e.Role != "AXRow" {
				t.Fatal("list option is not an AX row", e)
			}
		}
	}
}
