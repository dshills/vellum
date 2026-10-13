//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"

	"github.com/dshills/vellum/ui"
)

func TestExpandedNativeCheckableBooleanSet(t *testing.T) {
	for _, kind := range []ui.Kind{ui.Switch, ui.ToggleButton} {
		t.Run(string(kind), func(t *testing.T) {
			a, b := newAXTest(t, 8)
			var events []ui.ChangeEvent
			r := &record{node: ui.Node{ID: "boolean", Kind: kind, Change: func(e ui.ChangeEvent) { events = append(events, e) }}, control: &controlState{}}
			b.mu.Lock()
			b.handlers["boolean"] = func(action accessibilityAction) bool { return applyAccessibilityAction(r, action) }
			b.mu.Unlock()
			b.generation.Store(1)
			payload, _ := json.Marshal(accessibilityNode{ID: "boolean", Role: string(kind), Actions: []string{"set", "press"}})
			for _, value := range []string{"true", "false"} {
				if !nativeAccessibilitySetValueFixture(b.token, 1, string(payload), value) {
					t.Fatal("native NSNumber boolean refused", value)
				}
				a.Drain()
				if len(events) == 0 || events[len(events)-1].Kind != kind || events[len(events)-1].Checked != (value == "true") {
					t.Fatal("native boolean conversion", events)
				}
			}
		})
	}
}

func TestExpandedNativeRolesWithoutWindow(t *testing.T) {
	for _, test := range []struct {
		kind ui.Kind
		role string
	}{{ui.Switch, "AXCheckBox"}, {ui.ToggleButton, "AXCheckBox"}, {ui.ProgressBar, "AXProgressIndicator"}} {
		payload, _ := json.Marshal(struct {
			Generation int                 `json:"generation"`
			Nodes      []accessibilityNode `json:"nodes"`
		}{1, []accessibilityNode{{ID: "control", Role: string(test.kind), Actions: []string{}}}})
		// The inert selector fixture accepts any retained element ID; this exercises
		// the production role selector without an attachment or OS window.
		result := nativeAccessibilityGridFixture(string(payload), "control", 0, 0)
		var observed struct {
			Role string `json:"role"`
		}
		if err := json.Unmarshal([]byte(result), &observed); err != nil || observed.Role != test.role {
			t.Fatal("native role", test.kind, result, err)
		}
	}
}
