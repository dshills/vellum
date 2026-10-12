//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"reflect"
	"testing"
)

type axEnvelopeElement struct {
	ID                     string `json:"id"`
	ObjectIdentity         uint64 `json:"object_identity"`
	Live, Enabled, Focused bool
	PressAllowed           bool   `json:"press_allowed"`
	GridScope              string `json:"grid_scope"`
	GridIncarnation        uint64 `json:"grid_incarnation"`
	GridPending            bool   `json:"grid_pending"`
	GridTargetGeneration   uint64 `json:"grid_target_generation"`
	GridGeneration         uint64 `json:"grid_generation"`
	Frame                  [4]float64
	Value                  any
	Admitted               bool
	SetAllowed             bool `json:"set_allowed"`
	FocusAllowed           bool `json:"focus_allowed"`
}
type axEnvelopeResult struct {
	Accepted      bool
	Generation    uint64
	Elements      []axEnvelopeElement
	Retired       []axEnvelopeElement
	PendingGuards []axEnvelopeElement `json:"pending_guards"`
}

func axEnvelopeNodes(incarnation, target, completed uint64, pending bool, parent, value string) []map[string]any {
	nodes := []map[string]any{
		{"id": "grid", "parent": parent, "role": "table", "name": "Grid", "actions": []string{}},
		{"id": "row", "parent": "grid", "role": "grid-row", "actions": []string{}},
		{"id": "cell", "parent": "row", "role": "grid-cell", "actions": []string{"press", "focus"}, "focused": true, "value": value},
	}
	for i, n := range nodes {
		n["grid_scope"] = "grid"
		n["grid_incarnation"] = incarnation
		n["grid_pending"] = pending
		n["grid_target_generation"] = target
		n["grid_generation"] = completed
		n["grid_frame"] = [4]float64{float64(i * 10), 20, 90, 40}
		if pending {
			n["actions"] = []string{}
			n["disabled"] = true
			n["focused"] = false
		}
	}
	return nodes
}
func axEnvelopeSnapshot(generation uint64, nodes []map[string]any) map[string]any {
	return map[string]any{"generation": generation, "nodes": nodes}
}
func axEnvelopeUpdate(generation uint64, nodes []map[string]any) map[string]any {
	return map[string]any{"kind": "update", "snapshot": axEnvelopeSnapshot(generation, nodes)}
}
func axEnvelopeRun(t *testing.T, steps ...map[string]any) []axEnvelopeResult {
	t.Helper()
	payload, err := json.Marshal(steps)
	if err != nil {
		t.Fatal(err)
	}
	raw := nativeAccessibilityScopedFixture(string(payload))
	var out []axEnvelopeResult
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("native transaction fixture: %v: %q", err, raw)
	}
	if len(out) != len(steps) {
		t.Fatal("lost fixture operations", len(out), len(steps))
	}
	return out
}
func axEnvelopeFind(t *testing.T, result axEnvelopeResult, id string) axEnvelopeElement {
	t.Helper()
	for _, e := range result.Elements {
		if e.ID == id {
			return e
		}
	}
	t.Fatal("missing native element", id)
	return axEnvelopeElement{}
}
func TestNativeGridAccessibilityEnvelopeTransactions(t *testing.T) {
	button := func(value string) map[string]any {
		return map[string]any{"id": "button", "parent": "", "role": "button", "actions": []string{"press"}, "value": value}
	}
	startup := axEnvelopeNodes(1, 1, 0, true, "", "old")[:1]
	delete(startup[0], "grid_frame")
	ready := append(axEnvelopeNodes(1, 1, 1, false, "", "A日本"), button("old"))
	pending := append(axEnvelopeNodes(1, 2, 1, true, "", "A日本"), button("updated"))
	completed := append(axEnvelopeNodes(1, 2, 2, false, "", "B🙂"), button("updated"))
	completed[2]["grid_frame"] = [4]float64{120, 140, 90, 40}
	sameGeneration := append(axEnvelopeNodes(1, 2, 2, true, "", "B🙂"), button("updated"))
	sameGeneration[2]["grid_frame"] = [4]float64{120, 140, 90, 40}
	replacement := append(axEnvelopeNodes(2, 1, 1, false, "", "replacement"), button("updated"))
	moved := append(axEnvelopeNodes(2, 1, 1, false, "parent", "moved"), button("updated"), map[string]any{"id": "parent", "parent": "", "role": "column", "actions": []string{}})
	readded := append(axEnvelopeNodes(2, 1, 1, false, "", "readded"), button("updated"))
	ordinary := []map[string]any{{"id": "grid", "parent": "", "role": "button", "actions": []string{"press"}}, button("updated")}
	results := axEnvelopeRun(t,
		axEnvelopeUpdate(1, startup), axEnvelopeUpdate(2, ready), axEnvelopeUpdate(3, pending),
		map[string]any{"kind": "geometry", "bounds": map[string]any{"cell": [4]float64{999, 999, 4, 4}, "grid": [4]float64{999, 999, 4, 4}, "button": [4]float64{7, 8, 9, 10}}},
		axEnvelopeUpdate(4, completed), axEnvelopeUpdate(5, sameGeneration), axEnvelopeUpdate(6, replacement), axEnvelopeUpdate(7, moved),
		axEnvelopeUpdate(8, []map[string]any{button("updated")}), axEnvelopeUpdate(9, readded), axEnvelopeUpdate(9, ready), axEnvelopeUpdate(10, ordinary),
		map[string]any{"kind": "close"}, axEnvelopeUpdate(11, readded),
	)
	for _, i := range []int{0, 1, 2, 4, 5, 6, 7, 8, 9, 11} {
		if !results[i].Accepted {
			t.Fatal("valid transaction refused", i)
		}
	}
	if results[10].Accepted || results[13].Accepted {
		t.Fatal("stale/closed native transaction admitted")
	}
	start := axEnvelopeFind(t, results[0], "grid")
	host := axEnvelopeFind(t, results[1], "grid")
	if start.ObjectIdentity == 0 || start.ObjectIdentity != host.ObjectIdentity || start.Frame != ([4]float64{}) {
		t.Fatal("startup host lost identity or imported ambient frame", start, host)
	}
	cell := axEnvelopeFind(t, results[1], "cell")
	if cell.Value != "A日本" || !cell.Enabled || !cell.Focused || !cell.PressAllowed {
		t.Fatal("ready semantics not installed", cell)
	}
	for _, i := range []int{2, 3, 5} {
		e := axEnvelopeFind(t, results[i], "cell")
		if e.ObjectIdentity != cell.ObjectIdentity || e.Enabled || e.Focused || e.PressAllowed || !e.GridPending {
			t.Fatal("pending reused element remains actionable", i, e)
		}
		for _, guard := range results[i].PendingGuards {
			if guard.Enabled || guard.Focused || guard.PressAllowed || guard.SetAllowed || guard.FocusAllowed || guard.Admitted {
				t.Fatal("pending guard trusts presentation flags", guard)
			}
		}
	}
	if axEnvelopeFind(t, results[3], "cell").Frame != cell.Frame || axEnvelopeFind(t, results[3], "grid").Frame != host.Frame {
		t.Fatal("unversioned geometry overwrote grid envelope")
	}
	b := axEnvelopeFind(t, results[3], "button")
	if b.Value != "updated" || b.Frame != ([4]float64{7, 8, 9, 10}) {
		t.Fatal("unrelated control publication stalled", b)
	}
	e := axEnvelopeFind(t, results[4], "cell")
	if e.ObjectIdentity != cell.ObjectIdentity || e.Value != "B🙂" || e.Frame != ([4]float64{120, 140, 90, 40}) || e.GridGeneration != 2 || e.GridTargetGeneration != 2 || e.GridPending || !e.Enabled {
		t.Fatal("completed semantics/frame not installed together", e)
	}
	previous := e.ObjectIdentity
	for _, i := range []int{6, 7, 9} {
		e = axEnvelopeFind(t, results[i], "cell")
		if e.ObjectIdentity <= previous {
			t.Fatal("replacement/reparent/removal resurrected object", i, e)
		}
		previous = e.ObjectIdentity
	}
	if !reflect.DeepEqual(results[9].Elements, results[10].Elements) || results[10].Generation != 9 {
		t.Fatal("stale global generation mutated tree")
	}
	if axEnvelopeFind(t, results[11], "grid").ObjectIdentity <= host.ObjectIdentity {
		t.Fatal("removing scope resurrected its object")
	}
	for _, result := range results {
		for _, old := range result.Retired {
			if old.Live || old.Enabled || old.Focused || old.PressAllowed || old.Admitted {
				t.Fatal("retired reference remained actionable", old)
			}
		}
	}
}
func TestNativeGridAccessibilityEnvelopeValidationIsAtomic(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func([]map[string]any)
	}{
		{"missing host", func(n []map[string]any) { n[0]["id"] = "other" }},
		{"mixed scope", func(n []map[string]any) { n[2]["grid_scope"] = "other" }},
		{"mixed incarnation", func(n []map[string]any) { n[2]["grid_incarnation"] = uint64(2) }},
		{"mixed pending", func(n []map[string]any) { n[2]["grid_pending"] = true }},
		{"mixed target", func(n []map[string]any) { n[2]["grid_target_generation"] = uint64(2) }},
		{"mixed completed", func(n []map[string]any) { n[2]["grid_generation"] = uint64(0) }},
		{"ready no frame", func(n []map[string]any) { delete(n[2], "grid_frame") }},
		{"negative width", func(n []map[string]any) { n[2]["grid_frame"] = [4]float64{0, 0, -1, 2} }},
		{"negative height", func(n []map[string]any) { n[2]["grid_frame"] = [4]float64{0, 0, 1, -2} }},
		{"frame type", func(n []map[string]any) { n[2]["grid_frame"] = []any{0, 0, "1", 2} }},
		{"frame bool", func(n []map[string]any) { n[2]["grid_frame"] = []any{0, 0, true, 2} }},
		{"scope type", func(n []map[string]any) { n[2]["grid_scope"] = 3 }},
		{"zero incarnation", func(n []map[string]any) { n[2]["grid_incarnation"] = 0 }},
		{"fractional incarnation", func(n []map[string]any) { n[2]["grid_incarnation"] = 1.5 }},
		{"boolean incarnation", func(n []map[string]any) { n[2]["grid_incarnation"] = true }},
		{"negative generation", func(n []map[string]any) { n[2]["grid_generation"] = -1 }},
		{"fractional generation", func(n []map[string]any) { n[2]["grid_generation"] = 0.5 }},
		{"boolean generation", func(n []map[string]any) { n[2]["grid_generation"] = true }},
		{"null generation", func(n []map[string]any) { n[2]["grid_generation"] = nil }},
		{"numeric pending", func(n []map[string]any) { n[2]["grid_pending"] = 0 }},
		{"string pending", func(n []map[string]any) { n[2]["grid_pending"] = "false" }},
		{"null pending", func(n []map[string]any) { n[2]["grid_pending"] = nil }},
		{"unscoped descendant", func(n []map[string]any) {
			for _, k := range []string{"grid_scope", "grid_incarnation", "grid_pending", "grid_target_generation", "grid_frame"} {
				delete(n[2], k)
			}
		}},
		{"outside scoped ancestry", func(n []map[string]any) { n[2]["parent"] = "" }},
		{"pending actions", func(n []map[string]any) {
			for _, v := range n {
				v["grid_pending"] = true
				v["disabled"] = true
				v["focused"] = false
			}
		}},
		{"pending enabled", func(n []map[string]any) {
			for _, v := range n {
				v["grid_pending"] = true
				v["actions"] = []string{}
				v["disabled"] = false
				v["focused"] = false
			}
		}},
		{"pending focused", func(n []map[string]any) {
			for _, v := range n {
				v["grid_pending"] = true
				v["actions"] = []string{}
				v["disabled"] = true
				v["focused"] = true
			}
		}},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			valid := axEnvelopeNodes(1, 1, 1, false, "", "accepted")
			invalid := axEnvelopeNodes(1, 1, 1, false, "", "bad")
			test.mutate(invalid)
			r := axEnvelopeRun(t, axEnvelopeUpdate(1, valid), axEnvelopeUpdate(2, invalid))
			if !r[0].Accepted || r[1].Accepted || r[1].Generation != 1 || !reflect.DeepEqual(r[0].Elements, r[1].Elements) {
				t.Fatal("malformed envelope partly mutated native tree", r)
			}
		})
	}
	// Encoding failures cannot install a non-finite frame/global generation.
	for _, payload := range []string{
		`{"generation":-1,"nodes":[]}`, `{"generation":true,"nodes":[]}`, `{"generation":1.5,"nodes":[]}`,
		`{"generation":18446744073709551616,"nodes":[]}`,
	} {
		if nativeAccessibilitySnapshotValid(payload) {
			t.Fatal("invalid global generation admitted", payload)
		}
	}
	good := axEnvelopeNodes(^uint64(0), ^uint64(0), ^uint64(0), false, "", "uint64 exact")
	r := axEnvelopeRun(t, axEnvelopeUpdate(^uint64(0), good))
	if !r[0].Accepted || axEnvelopeFind(t, r[0], "cell").GridGeneration != ^uint64(0) {
		t.Fatal("valid uint64 lost precision", r)
	}
}
func TestNativeGridAccessibilityEnvelopeGeometryCompatibility(t *testing.T) {
	generic := []map[string]any{{"id": "button", "parent": "", "role": "button", "actions": []string{"press"}, "grid_generation": 7}}
	reparented := []map[string]any{
		{"id": "parent", "parent": "", "role": "column", "actions": []string{}},
		{"id": "button", "parent": "parent", "role": "button", "actions": []string{"press"}, "value": "reparented", "grid_generation": 7},
	}
	ready := axEnvelopeNodes(1, 1, 1, false, "", "ready")
	r := axEnvelopeRun(t, map[string]any{"kind": "geometry", "bounds": map[string]any{"grid": [4]float64{999, 999, 1, 1}, "button": [4]float64{4, 5, 6, 7}}}, axEnvelopeUpdate(1, generic), axEnvelopeUpdate(2, reparented), axEnvelopeUpdate(3, ready))
	if !r[1].Accepted || axEnvelopeFind(t, r[1], "button").Frame != ([4]float64{4, 5, 6, 7}) || !r[2].Accepted || !r[3].Accepted || axEnvelopeFind(t, r[3], "grid").Frame != ([4]float64{0, 20, 90, 40}) {
		t.Fatal("generic compatibility or scoped frame ownership changed", r)
	}
	before, after := axEnvelopeFind(t, r[1], "button"), axEnvelopeFind(t, r[2], "button")
	if before.ObjectIdentity == 0 || after.ObjectIdentity != before.ObjectIdentity || after.Frame != before.Frame || after.Value != "reparented" {
		t.Fatal("ordinary reparent discarded retained object or cached geometry", before, after)
	}
	for _, old := range r[2].Retired {
		if old.ObjectIdentity == before.ObjectIdentity || old.ID == before.ID {
			t.Fatal("ordinary reparent retired its retained object", old)
		}
	}
}

func TestNativeGridAccessibilityEnvelopeProductionEncoding(t *testing.T) {
	// Use the actual owner publication transformation and accessibilityNode
	// encoding, rather than explicit false/zero fields in hand-written JSON.
	b := &accessibilityBridge{}
	frame := [4]float32{3, 5, 91, 43}
	group := func(target, completed uint64, pending bool) []accessibilityNode {
		nodes := []accessibilityNode{
			{ID: "grid", Role: "table", Actions: []string{}},
			{ID: "row", Parent: "grid", Role: "grid-row", Actions: []string{}},
			{ID: "cell", Parent: "row", Role: "grid-cell", Value: "日本🙂", Focused: true, Actions: []string{"press"}},
		}
		for i := range nodes {
			n := &nodes[i]
			n.GridScope, n.GridIncarnation = "grid", 1
			n.GridTargetGeneration, n.GridGeneration, n.GridPending = target, completed, pending
			if !pending {
				n.GridFrame = &frame
			}
		}
		return nodes
	}
	inputs := [][]accessibilityNode{group(1, 0, true), group(1, 1, false), group(2, 0, true)}
	steps := make([]map[string]any, 0, len(inputs))
	for i, input := range inputs {
		nodes, _ := b.prepareGridPublicationSnapshot(input, nil, true)
		semantic, err := json.Marshal(nodes)
		if err != nil {
			t.Fatal(err)
		}
		payload := []byte(accessibilityPayload(uint64(i+1), string(semantic)))
		if !nativeAccessibilitySnapshotValid(string(payload)) {
			t.Fatalf("production encoding refused at stage %d: %s", i, payload)
		}
		var snapshot map[string]any
		if err := json.Unmarshal(payload, &snapshot); err != nil {
			t.Fatal(err)
		}
		encoded := snapshot["nodes"].([]any)
		for _, value := range encoded {
			node := value.(map[string]any)
			if i == 0 {
				if _, present := node["grid_generation"]; present {
					t.Fatal("fixture did not exercise omitted startup zero")
				}
			}
			if i == 1 {
				if _, present := node["grid_pending"]; present {
					t.Fatal("fixture did not exercise omitted ready false")
				}
			}
		}
		// Both explicit and omitted representations must compare consistently.
		if i == 1 {
			encoded[2].(map[string]any)["grid_pending"] = false
		}
		if i == 0 {
			encoded[0].(map[string]any)["grid_generation"] = 0
		}
		steps = append(steps, map[string]any{"kind": "update", "snapshot": snapshot})
	}
	r := axEnvelopeRun(t, steps...)
	if !r[0].Accepted || !r[1].Accepted || !r[2].Accepted || len(r[0].Elements) != 1 {
		t.Fatal("production startup/ready/pending envelope refused", r)
	}
	ready, pending := axEnvelopeFind(t, r[1], "cell"), axEnvelopeFind(t, r[2], "cell")
	if ready.GridPending || ready.GridGeneration != 1 || ready.Frame != ([4]float64{3, 5, 91, 43}) || !pending.GridPending || pending.GridGeneration != 1 || pending.GridTargetGeneration != 2 || pending.ObjectIdentity != ready.ObjectIdentity || pending.Frame != ready.Frame {
		t.Fatal("production cached publication metadata/frame drifted", ready, pending)
	}
}
