//go:build offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestAccessibilityPayloadMatchesValidatedEnvelope(t *testing.T) {
	selection := [2]int{1, 3}
	cases := [][]accessibilityNode{
		nil,
		{},
		{{ID: "control", Value: false, Actions: []string{}}},
		{{ID: "text", Name: "<script>&\"\\\n", Value: "A😀עברית\u2028\u2029\xff", TextSelection: &selection}},
		{{ID: "numeric", Value: float32(-0.125)}, {ID: "large", Value: uint64(math.MaxUint64)}},
		gridAXCompleted("g", "root", 17, 9),
		gridAXPending("g", "root", 17, 10),
		gridAXPayloadFixture(400),
	}
	for c, nodes := range cases {
		semantic, err := json.Marshal(nodes)
		if err != nil {
			t.Fatal(err)
		}
		for _, generation := range []uint64{0, 1, 12345, 1<<53 - 1, 1 << 53, math.MaxUint64} {
			// The prior stdlib envelope is the independent wire-format oracle.
			want, err := json.Marshal(struct {
				Generation uint64          `json:"generation"`
				Nodes      json.RawMessage `json:"nodes"`
			}{generation, semantic})
			if err != nil {
				t.Fatal(err)
			}
			if got := accessibilityPayload(generation, string(semantic)); got != string(want) {
				t.Fatalf("case %d generation %d changed the validated wire envelope", c, generation)
			}
		}
	}
}

func TestAccessibilityGridPublicationInterleavedScopeOwnership(t *testing.T) {
	b := &accessibilityBridge{}
	one := gridAXCompleted("one", "root", 1, 4)
	two := gridAXCompleted("two", "root", 2, 7)
	input := []accessibilityNode{{ID: "button", Role: "button", Actions: []string{"press"}}}
	for i := range one {
		input = append(input, one[i], two[i])
	}
	original, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	inputActions := gridAXHandlers(input, nil)
	out, actions := b.prepareGridPublication(input, inputActions)
	if !reflect.DeepEqual(out, input) || len(actions) != len(inputActions) || len(b.gridPublications) != 2 {
		t.Fatal("ready interleaved scopes changed output, handlers, or membership")
	}
	encoded, _ := json.Marshal(input)
	if string(encoded) != string(original) {
		t.Fatal("grouping changed caller records")
	}
	delete(actions, "one:cell")
	if inputActions["one:cell"] == nil {
		t.Fatal("returned action map aliases caller map")
	}
	// Both ready output and input remain independent of the retained cache.
	out[1].Name = "output mutation"
	input[5].Value = "input mutation"
	input[5].Actions[0] = "mutated action"
	input[5].GridFrame[0] = 999
	pending := append(gridAXPending("one", "root", 1, 5), gridAXPending("two", "root", 2, 8)...)
	retained, pendingActions := b.prepareGridPublication(pending, gridAXHandlers(pending, nil))
	gridAXAssertPending(t, retained, "one", 4, 5, pendingActions)
	gridAXAssertPending(t, retained, "two", 7, 8, pendingActions)
	byID := gridAXByID(t, retained)
	if byID["one:cell"].Value != "old cell" || byID["one:cell"].GridFrame[0] != 10 || len(retained) != 10 {
		t.Fatal("completed interleaved records alias caller storage or lost members")
	}
}

func TestAccessibilityOwnedPublicationMatchesBorrowedLifecycle(t *testing.T) {
	_, borrowed := newAXTest(t, 16)
	a, owned := newAXTest(t, 16)
	control := func(name string) accessibilityNode {
		return accessibilityNode{ID: "button", Role: "button", Name: name, Actions: []string{"press"}}
	}
	inputs := [][]accessibilityNode{
		append(gridAXCompleted("g", "root", 1, 1), control("initial")),
		append(gridAXCompleted("g", "root", 1, 1), control("initial")),
		append(gridAXPending("g", "root", 1, 2), control("pending")),
		append(gridAXPending("g", "root", 1, 2), control("pending")),
		append(append(gridAXPending("g", "root", 1, 3), gridAXCompleted("other", "root", 2, 7)...), control("mixed")),
		append(gridAXCompleted("g", "root", 1, 3), control("completed")),
		append(gridAXCompleted("g", "root", 2, 4), accessibilityNode{ID: "bad", Value: math.NaN()}),
		append(gridAXPending("g", "root", 1, 4), control("rejected predecessor")),
		{control("removed")},
		append(gridAXPending("g", "root", 1, 5), control("new pending host")),
	}
	callsByStep := make([]int, len(inputs))
	var admittedGeneration uint64
	for i, input := range inputs {
		// Separate input ownership for the two entries. The owned containers
		// are never touched after transfer, including rejected/closed updates.
		borrowedInput, ownedInput := cloneAccessibilityNodes(input), cloneAccessibilityNodes(input)
		borrowed.update(borrowedInput, gridAXHandlers(borrowedInput, nil))
		owned.updateOwned(ownedInput, gridAXHandlers(ownedInput, &callsByStep[i]))
		if borrowed.lastSnapshot != owned.lastSnapshot || borrowed.generation.Load() != owned.generation.Load() || !reflect.DeepEqual(borrowed.gridPublications, owned.gridPublications) {
			t.Fatalf("step %d changed publication, generation, or retained cache", i)
		}
		if len(borrowed.handlers) != len(owned.handlers) {
			t.Fatalf("step %d changed callback membership", i)
		}
		for id := range borrowed.handlers {
			if owned.handlers[id] == nil {
				t.Fatalf("step %d dropped callback %s", i, id)
			}
		}
		if i == 0 {
			admittedGeneration = owned.generation.Load()
			if !accessibilityDeliver(owned.token, admittedGeneration, "g:cell", accessibilityAction{Kind: "press"}) {
				t.Fatal("completed owned action not queued")
			}
		}
		if i == 1 {
			if owned.generation.Load() != admittedGeneration {
				t.Fatal("identical owned refresh invalidated queued action")
			}
			if err := a.Drain(); err != nil {
				t.Fatal(err)
			}
			if callsByStep[0] != 0 || callsByStep[1] != 1 {
				t.Fatal("identical owned refresh did not resolve the current callback")
			}
			if !accessibilityDeliver(owned.token, admittedGeneration, "g:cell", accessibilityAction{Kind: "press"}) {
				t.Fatal("owned completed action not queued before pending transition")
			}
		}
		if i == 2 {
			if accessibilityDeliver(owned.token, owned.generation.Load(), "g:cell", accessibilityAction{Kind: "press"}) {
				t.Fatal("owned pending records retained action authority")
			}
			if err := a.Drain(); err != nil {
				t.Fatal(err)
			}
			if callsByStep[0] != 0 || callsByStep[1] != 1 || callsByStep[2] != 0 {
				t.Fatal("completed action survived the owned pending transition")
			}
		}
	}
	borrowed.close()
	owned.close()
	borrowed.update(gridAXCompleted("g", "root", 3, 9), nil)
	owned.updateOwned(gridAXCompleted("g", "root", 3, 9), nil)
	if borrowed.lastSnapshot != owned.lastSnapshot || len(owned.gridPublications) != 0 || owned.handlers != nil {
		t.Fatal("closed owned bridge resurrected retained state")
	}
}

func gridAXPayloadFixture(cells int) []accessibilityNode {
	nodes := gridAXCompleted("g", "root", 1, 7)[:1]
	for i := 0; i < cells; i++ {
		frame := [4]float32{float32(i % 20 * 90), float32(i / 20 * 24), 90, 24}
		nodes = append(nodes, accessibilityNode{
			ID: fmt.Sprintf("g:cell:%d", i), Parent: "g", Role: "grid-cell", Value: "cell value",
			Actions: []string{"press"}, GridScope: "g", GridIncarnation: 1,
			GridGeneration: 7, GridTargetGeneration: 7, GridFrame: &frame,
		})
	}
	return nodes
}

var accessibilityPayloadBenchmarkSink string
var accessibilityPublicationBenchmarkSink []accessibilityNode

func BenchmarkAccessibilityValidatedEnvelope(b *testing.B) {
	semantic, err := json.Marshal(gridAXPayloadFixture(400))
	if err != nil {
		b.Fatal(err)
	}
	snapshot := string(semantic)
	b.Run("stdlib-RawMessage", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			payload, err := json.Marshal(struct {
				Generation uint64          `json:"generation"`
				Nodes      json.RawMessage `json:"nodes"`
			}{uint64(i + 12345), semantic})
			if err != nil {
				b.Fatal(err)
			}
			accessibilityPayloadBenchmarkSink = string(payload)
		}
	})
	b.Run("validated-envelope", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			accessibilityPayloadBenchmarkSink = accessibilityPayload(uint64(i+12345), snapshot)
		}
	})
}

func BenchmarkAccessibilityGridPublicationPrepare(b *testing.B) {
	nodes := gridAXPayloadFixture(400)
	handlers := gridAXHandlers(nodes, nil)
	for _, owned := range []bool{false, true} {
		b.Run(fmt.Sprintf("owned-%t", owned), func(b *testing.B) {
			bridge := &accessibilityBridge{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				accessibilityPublicationBenchmarkSink, _ = bridge.prepareGridPublicationSnapshot(nodes, handlers, owned)
			}
		})
	}
}
