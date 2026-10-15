//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

func TestListBoxCocoaSelectedRowsWritabilityLifetimes(t *testing.T) {
	nodes := func(writable, disabled bool) []map[string]any {
		actions := []string{}
		if writable {
			actions = []string{"set"}
		}
		return []map[string]any{
			{"id": "list", "parent": "", "role": "list-box", "actions": actions, "disabled": disabled},
			{"id": "row", "parent": "list", "role": "list-option", "actions": []string{"press"}},
		}
	}
	steps := []map[string]any{
		axEnvelopeUpdate(1, nodes(true, false)),
		axEnvelopeUpdate(2, nodes(false, false)),
		axEnvelopeUpdate(3, nodes(false, true)),
		axEnvelopeUpdate(4, nodes(true, false)),
		axEnvelopeUpdate(5, []map[string]any{}),
		axEnvelopeUpdate(6, nodes(true, false)),
		{"kind": "close"},
	}
	payload, _ := json.Marshal(steps)
	raw := nativeAccessibilityScopedFixture(string(payload))
	type element struct {
		ID       string
		Identity uint64 `json:"object_identity"`
		Allowed  bool   `json:"set_rows_allowed"`
		Reported bool   `json:"set_rows_reported"`
		Unknown  bool   `json:"unknown_settable"`
		Live     bool
	}
	var out []struct {
		Accepted bool
		Elements []element
		Retired  []element
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) != len(steps) {
		t.Fatal("metadata fixture", raw, err)
	}
	var identity uint64
	for i, result := range out {
		if i < 6 && !result.Accepted {
			t.Fatal("snapshot refused", i, raw)
		}
		if i == 4 && len(result.Elements) != 0 {
			t.Fatal("removed elements still published")
		}
		for _, e := range result.Elements {
			want := e.ID == "list" && (i == 0 || i == 3 || i == 5)
			if e.Reported != want || e.Allowed != want || e.Unknown {
				t.Fatalf("step %d: legacy metadata and selector policy differ: %+v; want %v", i, e, want)
			}
			if e.ID == "list" {
				if e.Identity == 0 || i > 0 && i < 4 && e.Identity != identity || i == 5 && e.Identity == identity {
					t.Fatal("unexpected retained/replaced identity", i, e, identity)
				}
				if i == 0 {
					identity = e.Identity
				}
			}
		}
		for _, e := range result.Retired {
			if e.Live || e.Reported || e.Allowed {
				t.Fatal("retired object still reports selectable rows", i, e)
			}
		}
	}
}

func TestListBoxCocoaMetadataClassTransitions(t *testing.T) {
	steps := []map[string]any{}
	for i, role := range []string{"list-box", "list-box", "button", "checkbox", "list-box"} {
		steps = append(steps, axEnvelopeUpdate(uint64(i+1), []map[string]any{
			{"id": "same", "parent": "", "role": role, "actions": []string{"set"}},
		}))
	}
	payload, _ := json.Marshal(steps)
	raw := nativeAccessibilityScopedFixture(string(payload))
	var out []struct {
		Accepted bool
		Elements []struct {
			Identity uint64 `json:"object_identity"`
			Reported *bool  `json:"set_rows_reported"`
		}
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) != len(steps) {
		t.Fatal(raw, err)
	}
	var previous uint64
	for i, step := range out {
		if !step.Accepted || len(step.Elements) != 1 {
			t.Fatal("native role transition refused", i, raw)
		}
		e := step.Elements[0]
		if e.Identity == 0 || i > 0 && (e.Identity == previous) != (i == 1 || i == 3) {
			t.Fatal("metadata capability crossed a class boundary or changed ordinary identity", i, e.Identity, previous)
		}
		list := i == 0 || i == 1 || i == 4
		if (e.Reported != nil) != list || list && !*e.Reported {
			t.Fatal("wrong role received list metadata implementation", i, raw)
		}
		previous = e.Identity
	}
}

func TestListBoxCocoaClassTransitionsKeepSuppressedGeometry(t *testing.T) {
	b := &accessibilityBridge{token: 17}
	bounds := map[string][4]float32{"same": {10, 20, 160, 80}}
	steps := []map[string]any{}
	sends := 0
	for i, role := range []string{"button", "list-box", "button", "list-box"} {
		b.mu.Lock()
		b.updateGeometryMembership([]accessibilityNode{{ID: "same", Role: role}})
		b.mu.Unlock()
		steps = append(steps, axEnvelopeUpdate(uint64(i+1), []map[string]any{
			{"id": "same", "parent": "", "role": role, "actions": []string{"set"}},
		}))
		b.publishGeometryTo(bounds, func(_ uint64, payload string) {
			sends++
			var geometry map[string][4]float64
			if err := json.Unmarshal([]byte(payload), &geometry); err != nil {
				t.Fatal(err)
			}
			steps = append(steps, map[string]any{"kind": "geometry", "bounds": geometry})
		})
	}
	if sends != 1 {
		t.Fatal("fixture did not exercise suppressed duplicate geometry", sends)
	}
	out := axEnvelopeRun(t, steps...)
	for i, result := range out[1:] {
		e := axEnvelopeFind(t, result, "same")
		if e.Frame != [4]float64{10, 20, 160, 80} {
			t.Fatal("Cocoa class transition lost the suppressed ordinary bounds", i+1, e)
		}
	}
}
