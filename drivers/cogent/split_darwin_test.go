//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

func TestSplitCocoaRolesAndOrientationWithoutWindow(t *testing.T) {
	for _, test := range []struct {
		role, direction, wantRole string
		orientation               int
	}{{"split-pane", "", "AXSplitGroup", 0}, {"split-handle", "vertical", "AXSplitter", 1}, {"split-handle", "horizontal", "AXSplitter", 2}} {
		payload, _ := json.Marshal(struct {
			Generation int                 `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, []accessibilityNode{{ID: "split", Role: test.role, Orientation: test.direction, Actions: []string{}}}})
		var observed struct {
			Role        string `json:"role"`
			Orientation int    `json:"orientation"`
		}
		result := nativeAccessibilityGridFixture(string(payload), "split", 0, 0)
		if err := json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != test.wantRole || observed.Orientation != test.orientation {
			t.Fatal("native split role/orientation", test, result, err)
		}
	}
}
