//go:build darwin && cgo && offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMultipleListNativePayloadAdmitted(t *testing.T) {
	a, _ := newAXTest(t, 921)
	host := core.NewFrame(core.NewBody("selection payload"))
	for _, selected := range [][]string{nil, {"a"}, {"a", "b"}} {
		r, err := reconcile(host, nil, ui.Node{ID: "list", Kind: ui.ListBox, Multiple: true, SelectedIDs: selected, Options: []ui.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, SelectionChange: func(ui.ListSelectionEvent) {}}, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		nodes, _ := accessibilitySnapshot(r, nil)
		payload, err := json.Marshal(struct {
			Generation uint64              `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, nodes})
		if err != nil {
			t.Fatal(err)
		}
		if !nativeAccessibilitySnapshotValid(string(payload)) {
			t.Fatal("native bridge refused multi-list payload", string(payload))
		}
		rows := 0
		for _, n := range nodes {
			if n.Role == "list-option" && n.Selected {
				rows++
			}
		}
		if rows != len(selected) {
			t.Fatal("row membership lost", rows)
		}
		disposeControlTree(r)
		r.widget.AsTree().Delete()
	}
}
