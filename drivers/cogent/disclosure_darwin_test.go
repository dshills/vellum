//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

func TestDisclosureCocoaRolesExpandedSelectorWithoutWindow(t *testing.T) {
	for _, expanded := range []bool{false, true} {
		payload, _ := json.Marshal(struct {
			Generation int                 `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, []accessibilityNode{{ID: "group", Role: "disclosure", Actions: []string{}},
			{ID: "heading", Parent: "group", Role: "disclosure-heading", Expanded: &expanded, Actions: []string{"press", "expand", "collapse"}}}})
		for _, tc := range []struct{ id, role string }{{"group", "AXGroup"}, {"heading", "AXDisclosureTriangle"}} {
			var observed struct {
				Role     string `json:"role"`
				Expanded bool   `json:"expanded"`
			}
			result := nativeAccessibilityGridFixture(string(payload), tc.id, 0, 0)
			if err := json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != tc.role || tc.id == "heading" && observed.Expanded != expanded {
				t.Fatal("native disclosure role/state", result, err)
			}
		}
	}
}
