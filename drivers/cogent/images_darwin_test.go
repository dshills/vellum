//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
)

func TestManagedImageNativeRoleFromActualDescriptor(t *testing.T) {
	a, b := newAXTest(t, 8)
	r, err := reconcile(core.NewFrame(core.NewBody("image role")), nil, ui.Node{ID: "photo", Kind: ui.Image, Image: managedImageSource(t), Text: "Preview"}, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, _ := accessibilitySnapshot(r, b)
	payload, _ := json.Marshal(struct {
		Generation int                 `json:"generation"`
		Nodes      []accessibilityNode `json:"nodes"`
	}{1, nodes})
	result := nativeAccessibilityGridFixture(string(payload), nodes[0].ID, 0, 0)
	var observed struct {
		Role string `json:"role"`
	}
	if err = json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != "AXImage" {
		t.Fatal("actual managed image native role", result, err)
	}
}
