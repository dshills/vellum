//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
)

func TestProgressShapesCocoaRoleThroughManagedSnapshotWithoutWindow(t *testing.T) {
	for _, kind := range []ui.Kind{ui.ProgressRing, ui.ProgressGauge} {
		a, b := newAXTest(t, 8)
		host := core.NewFrame(core.NewBody("Progress role"))
		r, err := reconcile(host, nil, ui.Node{ID: "shape", Kind: kind, Number: 0.75, Text: "75%"}, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		nodes, _ := accessibilitySnapshot(r, b)
		payload, err := json.Marshal(struct {
			Generation int                 `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, nodes})
		if err != nil {
			t.Fatal(err)
		}
		var observed struct {
			Role string `json:"role"`
		}
		result := nativeAccessibilityGridFixture(string(payload), accessibilityManagedID("shape"), 0, 0)
		if err := json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != "AXProgressIndicator" {
			t.Fatal("managed shape did not reach Cocoa progress role", kind, result, err)
		}
		disposeControlTree(r)
	}
}
