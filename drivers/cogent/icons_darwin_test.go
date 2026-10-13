//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

func TestIconsNativeImageRoleWithoutWindow(t *testing.T) {
	payload, _ := json.Marshal(struct {
		Generation int                 `json:"generation"`
		Nodes      []accessibilityNode `json:"nodes"`
	}{1, []accessibilityNode{{ID: "symbol", Role: "image", Name: "Warning", Actions: []string{}}}})
	result := nativeAccessibilityGridFixture(string(payload), "symbol", 0, 0)
	var observed struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != "AXImage" {
		t.Fatal("native image role", result, err)
	}
}
