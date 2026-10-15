//go:build darwin && cgo && offscreen

package cogent

import (
	"encoding/json"
	"testing"
)

// Cocoa elements/selectors only: no NSWindow, system event queue, or trust prompt.
func TestCanvasNativeRolesHitOrderAndAtomicFramesWithoutWindow(t *testing.T) {
	f := newCanvasAX(t, 7, 160)
	f.owner.Do(func() { f.metadata[0], f.metadata[1] = f.metadata[1], f.metadata[0] })
	f.refresh(t)
	nodes, _ := f.snapshot(t)
	payload, err := json.Marshal(struct {
		Generation uint64              `json:"generation"`
		Nodes      []accessibilityNode `json:"nodes"`
	}{1, nodes})
	if err != nil {
		t.Fatal(err)
	}
	if !nativeAccessibilitySnapshotValid(string(payload)) {
		t.Fatal("actual scoped canvas rejected")
	}
	host := accessibilityManagedID("drawing")
	front := accessibilityManagedID(CanvasObjectID("drawing", "front"))
	rear := accessibilityManagedID(CanvasObjectID("drawing", "rear:/β"))
	for _, tc := range []struct{ id, role string }{{host, "AXGroup"}, {front, "AXButton"}, {rear, "AXButton"}} {
		var observed struct {
			Role string `json:"role"`
		}
		if err := json.Unmarshal([]byte(nativeAccessibilityGridFixture(string(payload), tc.id, 0, 0)), &observed); err != nil || observed.Role != tc.role {
			t.Fatal("Cocoa role", tc, observed, err)
		}
	}
	// Production child traversal must hit the last painted overlapping primitive,
	// even when application metadata is intentionally reversed.
	var point [2]float64
	for _, n := range nodes {
		if n.ID == front {
			b := n.GridFrame
			point = [2]float64{float64(b[0] + b[2]*.75), float64(b[1] + b[3]*.5)}
		}
	}
	frames := map[string][4]float32{}
	for _, n := range nodes {
		if n.GridFrame != nil {
			frames[n.ID] = *n.GridFrame
		}
	}
	geometry, _ := json.Marshal(frames)
	if got := nativeAccessibilityHitFixture(string(payload), string(geometry), point[0], point[1]); got != front {
		t.Fatal("Cocoa z-order", got, front)
	}
	// The legacy scope protocol is narrowly extended; arbitrary groups cannot
	// masquerade as a geometry-qualified canvas host.
	for i := range nodes {
		if nodes[i].ID == host {
			nodes[i].Role = "group"
		}
	}
	rejected, _ := json.Marshal(struct {
		Generation uint64              `json:"generation"`
		Nodes      []accessibilityNode `json:"nodes"`
	}{1, nodes})
	if nativeAccessibilitySnapshotValid(string(rejected)) {
		t.Fatal("arbitrary scoped group admitted")
	}
	f.owner.Do(func() { f.model.Objects[0].X += 20 })
	f.refresh(t)
	ready, _ := f.snapshot(t)
	asMaps := func(ns []accessibilityNode) []map[string]any {
		raw, _ := json.Marshal(ns)
		var out []map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	initial := asMaps(ready)
	pending := asMaps(ready)
	for _, n := range pending {
		if n["grid_scope"] == host {
			n["grid_pending"] = true
			n["disabled"] = true
			n["focused"] = false
			n["actions"] = []string{}
		}
	}
	results := axEnvelopeRun(t, axEnvelopeUpdate(1, initial), axEnvelopeUpdate(2, pending), map[string]any{"kind": "geometry", "bounds": map[string]any{rear: [4]float64{999, 999, 1, 1}}}, axEnvelopeUpdate(3, initial), axEnvelopeUpdate(4, []map[string]any{}), axEnvelopeUpdate(5, initial), map[string]any{"kind": "close"})
	for _, i := range []int{0, 1, 3, 4, 5} {
		if !results[i].Accepted {
			t.Fatal("valid canvas transaction refused", i)
		}
	}
	a := axEnvelopeFind(t, results[0], rear)
	for _, i := range []int{1, 2} {
		b := axEnvelopeFind(t, results[i], rear)
		if a.ObjectIdentity != b.ObjectIdentity || b.Enabled || b.PressAllowed || b.Frame != a.Frame {
			t.Fatal("pending canvas changed geometry/identity or stayed actionable", i, a, b)
		}
	}
	b := axEnvelopeFind(t, results[3], rear)
	if b.ObjectIdentity != a.ObjectIdentity || !b.PressAllowed {
		t.Fatal("completed canvas did not restore")
	}
	if c := axEnvelopeFind(t, results[5], rear); c.ObjectIdentity == a.ObjectIdentity {
		t.Fatal("removed native object resurrected")
	}
	for _, element := range results[6].Elements {
		if element.Live || element.Enabled || element.PressAllowed {
			t.Fatal("closed retained reference remains live", element)
		}
	}
}
