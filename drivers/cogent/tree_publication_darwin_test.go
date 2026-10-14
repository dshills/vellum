//go:build darwin && cgo

package cogent

import (
	"testing"
)

func treeEnvelopeNodes(incarnation, generation uint64, pending bool) []map[string]any {
	nodes := []map[string]any{
		{"id": "outline", "parent": "", "role": "outline", "value": "tree", "actions": []string{"focus"}},
		{"id": "row", "parent": "outline", "role": "tree-row", "value": "row value", "actions": []string{"press", "expand"}, "expanded": true},
		{"id": "page", "parent": "outline", "role": "button", "actions": []string{"press"}},
	}
	for i, n := range nodes {
		n["grid_scope"] = "outline"
		n["grid_incarnation"] = incarnation
		n["grid_target_generation"] = generation
		n["grid_generation"] = generation
		n["grid_pending"] = pending
		n["grid_frame"] = [4]float64{float64(i * 10), 20, 90, 40}
		if pending {
			n["actions"] = []string{}
			n["disabled"] = true
			n["focused"] = false
		}
	}
	return nodes
}

func TestNativeTreePublicationRetainsPendingObjectsAndRetiresScopes(t *testing.T) {
	ready := treeEnvelopeNodes(1, 1, false)
	pending := treeEnvelopeNodes(1, 1, true)
	restored := treeEnvelopeNodes(1, 1, false)
	restored[1]["grid_frame"] = [4]float64{77, 88, 90, 40}
	replacement := treeEnvelopeNodes(2, 1, false)
	kindChanged := treeEnvelopeNodes(2, 1, false)
	kindChanged[0]["role"] = "table"
	results := axEnvelopeRun(t,
		axEnvelopeUpdate(1, ready), axEnvelopeUpdate(2, pending),
		map[string]any{"kind": "geometry", "bounds": map[string]any{"row": [4]float64{999, 999, 1, 1}}},
		axEnvelopeUpdate(3, restored), axEnvelopeUpdate(4, replacement),
		axEnvelopeUpdate(5, []map[string]any{}), axEnvelopeUpdate(6, replacement),
		axEnvelopeUpdate(7, kindChanged), map[string]any{"kind": "close"},
	)
	for _, index := range []int{0, 1, 3, 4, 5, 6, 7} {
		if !results[index].Accepted {
			t.Fatal("valid scoped publication refused", index)
		}
	}
	for _, id := range []string{"outline", "row", "page"} {
		first := axEnvelopeFind(t, results[0], id)
		old := axEnvelopeFind(t, results[1], id)
		ignored := axEnvelopeFind(t, results[2], id)
		newFrame := axEnvelopeFind(t, results[3], id)
		if first.ObjectIdentity == 0 || old.ObjectIdentity != first.ObjectIdentity || ignored.ObjectIdentity != first.ObjectIdentity || newFrame.ObjectIdentity != first.ObjectIdentity {
			t.Fatal("pending/completed transaction changed identity", id)
		}
		if !old.Live || !old.GridPending || old.Enabled || old.Focused || old.PressAllowed || old.Frame != first.Frame || ignored.Frame != first.Frame {
			t.Fatal("pending tree lost readable frame or remained interactive", id)
		}
		if newFrame.GridPending || !newFrame.Enabled {
			t.Fatal("completed tree remained pending", id)
		}
		swapped := axEnvelopeFind(t, results[4], id)
		readded := axEnvelopeFind(t, results[6], id)
		changed := axEnvelopeFind(t, results[7], id)
		if swapped.ObjectIdentity == first.ObjectIdentity || readded.ObjectIdentity == swapped.ObjectIdentity || changed.ObjectIdentity == readded.ObjectIdentity {
			t.Fatal("replacement/removal/kind transition reused retired scope", id)
		}
	}
	if len(results[1].PendingGuards) != 3 {
		t.Fatal("native pending defenses not exercised")
	}
	for _, guard := range results[1].PendingGuards {
		if guard.Admitted || guard.PressAllowed || guard.FocusAllowed || guard.SetAllowed || guard.Enabled || guard.Focused {
			t.Fatal("native pending guard bypassed presentation flags")
		}
	}
	if axEnvelopeFind(t, results[3], "row").Frame != [4]float64{77, 88, 90, 40} {
		t.Fatal("same-generation completed geometry was not atomic")
	}
	for _, retired := range results[8].Retired {
		if retired.Live || retired.Enabled || retired.Admitted {
			t.Fatal("closed tree scope remains active")
		}
	}
}
