//go:build offscreen

package cogent

import (
	"fmt"
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
				bridge := &accessibilityBridge{lastSnapshot: semantic}
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
	bridge := &accessibilityBridge{lastSnapshot: semantic, done: make(chan struct{})}
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
