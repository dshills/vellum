//go:build offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

// Measure the duplicate publication phase with fresh transferred containers on
// every owned call. No native window or changed native publication is created.
func BenchmarkAccessibilityDuplicatePublication(b *testing.B) {
	for _, cells := range []int{0, 40, 400} {
		for _, owned := range []bool{false, true} {
			b.Run(fmt.Sprintf("cells-%d/owned-%t", cells, owned), func(b *testing.B) {
				fixture := gridAXPayloadFixture(cells)
				semantic, err := accessibilitySemantic(fixture)
				if err != nil {
					b.Fatal(err)
				}
				bridge := &accessibilityBridge{lastSnapshot: semantic, pendingBaseline: accessibilityPendingBaseline(fixture)}
				bridge.prepareGridPublication(fixture, nil)
				publish := func() {
					nodes := cloneAccessibilityNodes(fixture)
					handlers := gridAXHandlers(nodes, nil)
					if owned {
						bridge.updateOwned(nodes, handlers)
					} else {
						bridge.update(nodes, handlers)
					}
				}
				publish()
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					publish()
				}
				b.StopTimer()
				if bridge.generation.Load() != 0 {
					b.Fatal("benchmark published native changes")
				}
				accessibilityPayloadBenchmarkSink = bridge.lastSnapshot
			})
		}
	}
}

// Seed a duplicate-only bridge so these allocation/cache tests never enqueue a
// changed native payload. The owner/action lifecycle oracle is tested separately.
func duplicateAXBridge(t testing.TB, nodes []accessibilityNode) *accessibilityBridge {
	t.Helper()
	semantic, err := accessibilitySemantic(nodes)
	if err != nil {
		t.Fatal(err)
	}
	bridge := &accessibilityBridge{lastSnapshot: semantic, pendingBaseline: accessibilityPendingBaseline(nodes), done: make(chan struct{})}
	bridge.updateOwned(cloneAccessibilityNodes(nodes), gridAXHandlers(nodes, nil))
	if bridge.generation.Load() != 0 {
		t.Fatal("seed changed native publication")
	}
	return bridge
}

func TestAccessibilityDuplicateRetainsDetachedCompletedView(t *testing.T) {
	one, two := gridAXCompleted("one", "root", 1, 4), gridAXCompleted("two", "root", 2, 7)
	var input []accessibilityNode
	for i := range one {
		input = append(input, one[i], two[i])
	}
	bridge := duplicateAXBridge(t, input)
	oldOne, oldTwo := &bridge.gridPublications["one"][0], &bridge.gridPublications["two"][0]
	fresh := cloneAccessibilityNodes(input)
	freshFrame := fresh[4].GridFrame
	calls := 0
	bridge.updateOwned(fresh, gridAXHandlers(fresh, &calls))
	if &bridge.gridPublications["one"][0] != oldOne || &bridge.gridPublications["two"][0] != oldTwo {
		t.Fatal("identical owned refresh replaced detached completed storage")
	}
	if bridge.gridPublications["one"][2].GridFrame == freshFrame {
		t.Fatal("reused completed view aliases transferred pointer fields")
	}
	if bridge.handlers["two:cell"] == nil || !bridge.handlers["two:cell"](accessibilityAction{Kind: "press"}) || calls != 1 {
		t.Fatal("identical refresh retained old callbacks")
	}
	bridge.close()
	bridge.updateOwned(cloneAccessibilityNodes(input), nil)
	if bridge.completedSnapshot != "" || bridge.gridPublications != nil || bridge.handlers != nil {
		t.Fatal("retirement retained or resurrected ready reuse state")
	}
}

func TestAccessibilityDuplicatePreparationInvalidatesReuse(t *testing.T) {
	nodes := gridAXCompleted("g", "root", 1, 7)
	bridge := duplicateAXBridge(t, nodes)
	bridge.prepareGridPublication(nil, nil)
	if bridge.completedSnapshot != "" || len(bridge.gridPublications) != 0 {
		t.Fatal("direct removal kept ready identity")
	}
	bridge.updateOwned(cloneAccessibilityNodes(nodes), nil)
	if len(bridge.gridPublications["g"]) != len(nodes) {
		t.Fatal("identical refresh failed to restore removed completed cache")
	}
	previous := &bridge.gridPublications["g"][0]
	bridge.prepareGridPublication(gridAXPending("g", "root", 1, 8), nil)
	bridge.updateOwned(cloneAccessibilityNodes(nodes), nil)
	if &bridge.gridPublications["g"][0] == previous {
		t.Fatal("direct pending preparation reused an uncertified cache")
	}
}

type duplicateAXCustomValue struct{ Hidden int }

func (duplicateAXCustomValue) MarshalJSON() ([]byte, error) { return []byte(`"same"`), nil }

func TestAccessibilityDuplicateRebuildsNonEquivalentCacheMetadata(t *testing.T) {
	for _, kind := range []string{"focus", "scalar type", "custom", "mutable", "borrowed"} {
		t.Run(kind, func(t *testing.T) {
			nodes := gridAXCompleted("g", "root", 1, 7)
			switch kind {
			case "scalar type":
				nodes[2].Value = int(1)
			case "custom":
				nodes[2].Value = duplicateAXCustomValue{Hidden: 1}
			case "mutable":
				nodes[2].Value = map[string]int{"answer": 1}
			}
			bridge := duplicateAXBridge(t, nodes)
			previous := &bridge.gridPublications["g"][0]
			next := cloneAccessibilityNodes(nodes)
			switch kind {
			case "focus":
				next[2].pressFocusID = "new-focus-host"
			case "scalar type":
				next[2].Value = float64(1)
			case "custom":
				next[2].Value = duplicateAXCustomValue{Hidden: 2}
			case "mutable":
				next[2].Value = map[string]int{"answer": 1}
			}
			want := cloneAccessibilityNode(next[2])
			if kind == "borrowed" {
				bridge.update(next, nil)
			} else {
				bridge.updateOwned(next, nil)
			}
			if &bridge.gridPublications["g"][0] == previous {
				t.Fatal("non-reusable update preserved previous cache storage")
			}
			if got := bridge.gridPublications["g"][2]; got.pressFocusID != want.pressFocusID || !reflect.DeepEqual(got.Value, want.Value) {
				t.Fatal("semantic equality lost owner metadata or scalar type")
			}
			if bridge.generation.Load() != 0 {
				t.Fatal("equal JSON retired the semantic generation")
			}
		})
	}
}

func TestAccessibilityDuplicateScalarAfterCustomCannotReuseOldValue(t *testing.T) {
	nodes := gridAXCompleted("g", "root", 1, 7)
	nodes[2].Value = duplicateAXCustomValue{Hidden: 1}
	bridge := duplicateAXBridge(t, nodes)
	if bridge.completedSnapshot != "" {
		t.Fatal("custom value certified a reusable completed cache")
	}
	next := cloneAccessibilityNodes(nodes)
	next[2].Value = "same"
	bridge.updateOwned(next, nil)
	if bridge.gridPublications["g"][2].Value != "same" {
		t.Fatal("builtin scalar reused older custom value with identical JSON")
	}
}

// Refused new input must leave the last validated view and callbacks intact,
// even when most of its nodes match the ready baseline used by the shortcut.
func TestAccessibilityReadyDuplicateInvalidEncodingIsAtomic(t *testing.T) {
	for name, mutate := range map[string]func(*accessibilityNode){
		"float64 NaN":      func(n *accessibilityNode) { n.Value = math.NaN() },
		"float32 infinity": func(n *accessibilityNode) { n.Value = float32(math.Inf(1)) },
		"frame NaN":        func(n *accessibilityNode) { n.GridFrame[0] = float32(math.NaN()) },
		"minimum infinity": func(n *accessibilityNode) { v := float32(math.Inf(-1)); n.Min = &v },
		"maximum NaN":      func(n *accessibilityNode) { v := float32(math.NaN()); n.Max = &v },
		"function":         func(n *accessibilityNode) { n.Value = func() {} },
		"invalid JSON":     func(n *accessibilityNode) { n.Value = json.RawMessage("{") },
	} {
		t.Run(name, func(t *testing.T) {
			nodes := gridAXCompleted("g", "root", 1, 7)
			bridge := duplicateAXBridge(t, nodes)
			defer bridge.close()
			original, refused := 0, 0
			bridge.handlers = gridAXHandlers(nodes, &original)
			before, storage := bridge.lastSnapshot, &bridge.gridPublications["g"][0]
			next := cloneAccessibilityNodes(nodes)
			mutate(&next[2])
			bridge.updateOwned(next, gridAXHandlers(next, &refused))
			if bridge.lastSnapshot != before || bridge.generation.Load() != 0 || &bridge.gridPublications["g"][0] != storage {
				t.Fatal("invalid encoding changed accepted publication or cache")
			}
			if !bridge.handlers["g:cell"](accessibilityAction{Kind: "press"}) || original != 1 || refused != 0 {
				t.Fatal("invalid encoding replaced the accepted action")
			}
		})
	}
}

func TestAccessibilityReadyDuplicateEmptyWireDistinction(t *testing.T) {
	_, bridge := newAXTest(t, 16)
	for _, nodes := range [][]accessibilityNode{nil, {}, nil} {
		bridge.updateOwned(nodes, nil)
		want, err := json.Marshal(nodes)
		if err != nil || bridge.lastSnapshot != string(want) {
			t.Fatal("empty and nil snapshots lost their wire distinction", bridge.lastSnapshot, err)
		}
	}
	if bridge.generation.Load() != 3 {
		t.Fatal("empty/nil changes did not retire old generations")
	}
}

func TestAccessibilityReadyDuplicateReplacesAndRemovesActivationGuards(t *testing.T) {
	a, bridge := newAXTest(t, 16)
	nodes := gridAXCompleted("g", "root", 1, 7)
	calls := 0
	publish := func(guard func() bool) {
		t.Helper()
		next := cloneAccessibilityNodes(nodes)
		next[2].activationGuard = guard
		bridge.updateOwned(next, gridAXHandlers(next, &calls))
	}
	admit := func() {
		t.Helper()
		if !accessibilityDeliver(bridge.token, bridge.generation.Load(), "g:cell", accessibilityAction{Kind: "press"}) {
			t.Fatal("live action admission refused")
		}
	}
	drain := func() {
		t.Helper()
		if err := a.Drain(); err != nil {
			t.Fatal(err)
		}
	}
	publish(func() bool { return false })
	generation, storage := bridge.generation.Load(), &bridge.gridPublications["g"][0]
	admit()
	publish(func() bool { return true })
	drain()
	if calls != 0 {
		t.Fatal("a duplicate refresh bypassed the already-admitted old guard")
	}
	admit()
	drain()
	if calls != 1 {
		t.Fatal("fresh admission did not use the replacement guard")
	}
	publish(func() bool { return false })
	admit()
	drain()
	if calls != 1 {
		t.Fatal("fresh false guard did not revoke activation")
	}
	publish(nil)
	admit()
	drain()
	if calls != 2 || len(bridge.activationGuards) != 0 {
		t.Fatal("removed guard still blocked new admission")
	}
	if bridge.generation.Load() != generation || &bridge.gridPublications["g"][0] != storage {
		t.Fatal("Go-only guard refresh retired unchanged semantics or completed cache")
	}
}
