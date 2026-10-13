//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

func TestChromeCocoaRolesWithoutWindow(t *testing.T) {
	for _, test := range []struct{ role, want string }{{"toolbar", "AXToolbar"}, {"status-bar", "AXGroup"}, {"form-group", "AXGroup"}} {
		payload, _ := json.Marshal(struct {
			Generation int                 `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, []accessibilityNode{{ID: "control", Role: test.role, Actions: []string{}}}})
		var observed struct {
			Role string `json:"role"`
		}
		result := nativeAccessibilityGridFixture(string(payload), "control", 0, 0)
		if err := json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != test.want {
			t.Fatal(test, result, err)
		}
	}
}
