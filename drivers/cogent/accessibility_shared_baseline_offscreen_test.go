//go:build offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestAccessibilitySharedBaselineOrderedAndDetached(t *testing.T) {
	_, bridge := newAXTest(t, 16)
	one, two := gridAXCompleted("one", "root", 1, 4), gridAXCompleted("two", "root", 2, 7)
	frame := [4]float32{3, 4, 5, 6}
	actions := []string{"press"}
	minimum := float32(-5)
	ordinary := accessibilityNode{ID: "ordinary", Role: "button", Value: "ordinary", Actions: actions, GridFrame: &frame, Min: &minimum, activationGuard: func() bool { return true }, pressFocusID: "ordinary-focus"}
	input := []accessibilityNode{ordinary}
	for i := range one {
		input = append(input, one[i], two[i])
	}
	input[3].Actions = nil
	input[4].Actions = []string{}
	input[5].pressFocusID = "one-focus"
	want := accessibilityPendingBaseline(input)
	bridge.updateOwned(input, gridAXHandlers(input, nil))
	if len(bridge.pendingBaseline) != 0 || len(bridge.readyBaseline) != len(want) {
		t.Fatal("ready snapshot repeated full baseline or missed records")
	}
	baseline := bridge.wireBaseline()
	for i := range want {
		if !reflect.DeepEqual(baseline.at(i), want[i]) {
			t.Fatalf("baseline record %d differs from detached flat oracle", i)
		}
	}
	for i := range one {
		if bridge.readyBaseline[1+2*i].node != &bridge.gridPublications["one"][i] || bridge.readyBaseline[2+2*i].node != &bridge.gridPublications["two"][i] {
			t.Fatal("interleaved wire order or cache sharing changed")
		}
	}
	// Source-owned pointer fields and action backing may outlive a transfer;
	// they must not mutate the detached ordinary record or the proof baseline.
	frame[0] = 999
	actions[0] = "changed"
	minimum = 99
	if !reflect.DeepEqual(baseline.at(0), want[0]) || bridge.readyBaseline[0].node.activationGuard != nil {
		t.Fatal("ordinary snapshot retained mutable source or callback")
	}
	if bridge.activationGuards["ordinary"] == nil || !bridge.activationGuards["ordinary"]() {
		t.Fatal("live callback admission lost")
	}
	// The getter restores null without normalizing the retained cache in place.
	if baseline.at(3).Actions != nil || bridge.gridPublications["one"][1].Actions == nil || baseline.at(4).Actions == nil {
		t.Fatal("nil/empty wire spelling changed")
	}
	if baseline.at(5).pressFocusID != "" || bridge.gridPublications["one"][2].pressFocusID != "one-focus" {
		t.Fatal("wire proof lost distinction from Go-only cache metadata")
	}
}

func TestAccessibilitySharedBaselineLifecycleMatchesFlat(t *testing.T) {
	_, old := newAXTest(t, 16)
	_, current := newAXTest(t, 16)
	build := func(g uint64) []accessibilityNode {
		one, two := gridAXCompleted("one", "root", 1, g), gridAXCompleted("two", "root", 2, g+3)
		nodes := []accessibilityNode{{ID: "ordinary", Role: "button", Actions: nil}}
		for i := range one {
			nodes = append(nodes, one[i], two[i])
		}
		nodes[3].Actions = nil
		nodes[4].Actions = []string{}
		return nodes
	}
	sequence := [][]accessibilityNode{build(4), build(4), build(5), append(gridAXPending("one", "root", 1, 6), gridAXCompleted("two", "root", 2, 8)...), gridAXPending("two", "root", 2, 9), gridAXCompleted("two", "root", 2, 9), {{ID: "only", Role: "button", Actions: []string{}}}}
	for step, input := range sequence {
		a, b := cloneAccessibilityNodes(input), cloneAccessibilityNodes(input)
		// Ordinary clone normalizes nil Actions; restore source spelling for both.
		for i := range input {
			if input[i].Actions == nil {
				a[i].Actions = nil
				b[i].Actions = nil
			}
		}
		old.update(a, gridAXHandlers(a, nil))
		current.updateOwned(b, gridAXHandlers(b, nil))
		if old.lastSnapshot != current.lastSnapshot || old.generation.Load() != current.generation.Load() || !reflect.DeepEqual(old.gridPublications, current.gridPublications) || len(old.handlers) != len(current.handlers) {
			t.Fatalf("step %d differs from flat borrowed publication", step)
		}
		baseline := current.wireBaseline()
		var emitted []accessibilityNode
		// The Go cache includes only detached values, and every new baseline matches
		// the actual last wire publication even during pending replacement/removal.
		if err := json.Unmarshal([]byte(current.lastSnapshot), &emitted); err != nil {
			t.Fatal(err)
		}
		if baseline.len() != len(emitted) {
			t.Fatalf("step %d baseline membership", step)
		}
		semanticNodes := make([]accessibilityNode, baseline.len())
		for i := range semanticNodes {
			semanticNodes[i] = baseline.at(i)
		}
		semantic, err := accessibilitySemantic(semanticNodes)
		if err != nil || semantic != current.lastSnapshot {
			t.Fatalf("step %d baseline wire differs: %v", step, err)
		}
		if step == 4 {
			if len(current.readyBaseline) != 0 || len(current.gridPublications) != 1 || current.gridPublications["one"] != nil {
				t.Fatal("pending survivor retained removed sibling")
			}
			for i := range baseline.nodes {
				if baseline.nodes[i].GridScope == "one" {
					t.Fatal("wire baseline retained removed grid")
				}
			}
		}
	}
	current.close()
	if current.pendingBaseline != nil || current.readyBaseline != nil || current.gridPublications != nil {
		t.Fatal("close retained shared baseline")
	}
}

func TestAccessibilitySharedBaselineDirectPreparationAndTypedReplacement(t *testing.T) {
	_, bridge := newAXTest(t, 16)
	nodes := gridAXCompleted("g", "root", 1, 7)
	nodes[2].Value = int(1)
	bridge.updateOwned(nodes, nil)
	first := bridge.readyBaseline[2].node
	bridge.prepareGridPublication(nil, nil)
	if bridge.completedSnapshot != "" {
		t.Fatal("direct preparation kept ready identity")
	}
	next := gridAXCompleted("g", "root", 1, 7)
	next[2].Value = float64(1)
	bridge.updateOwned(next, nil)
	if bridge.readyBaseline[2].node == first || bridge.readyBaseline[2].node != &bridge.gridPublications["g"][2] || !reflect.DeepEqual(bridge.wireBaseline().at(2).Value, float64(1)) {
		t.Fatal("JSON-identical typed replacement retained obsolete cache")
	}
	previous := bridge.readyBaseline[2].node
	bridge.updateOwned(gridAXCompleted("g", "root", 1, 8), nil)
	if bridge.readyBaseline[2].node == previous {
		t.Fatal("changed ready baseline retained old group")
	}
}

var sharedBaselineSink accessibilityWireBaseline

func BenchmarkAccessibilityChangedPublicationOwnership(b *testing.B) {
	fixture := gridAXPayloadFixture(400)
	for _, mode := range []string{"flat", "references"} {
		b.Run(mode, func(b *testing.B) {
			bridge := &accessibilityBridge{}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				nodes := cloneAccessibilityNodes(fixture)
				nodes[0].Name = strconv.Itoa(i)
				handlers := gridAXHandlers(nodes, nil)
				semantic, e := accessibilitySemantic(nodes)
				if e != nil {
					b.Fatal(e)
				}
				if mode == "flat" {
					nodes, _ = bridge.legacySharedBaselinePrepare(nodes, handlers, true, semantic)
					sharedBaselineSink = accessibilityWireBaseline{nodes: accessibilityPendingBaseline(nodes)}
				} else {
					var refs []accessibilityNodeRef
					nodes, _, refs = bridge.prepareGridPublicationWithBaseline(nodes, handlers, true, semantic)
					sharedBaselineSink = accessibilityWireBaseline{references: refs}
				}
				if sharedBaselineSink.len() != 401 {
					b.Fatal(fmt.Sprint("unexpected record count ", sharedBaselineSink.len()))
				}
				accessibilityPayloadBenchmarkSink = accessibilityPayload(uint64(i+1), semantic)
			}
		})
	}
}

func TestAccessibilitySharedBaselinePendingProofParity(t *testing.T) {
	_, bridge := newAXTest(t, 16)
	one, two := gridAXCompleted("one", "root", 1, 4), gridAXCompleted("two", "root", 2, 7)
	var ready []accessibilityNode
	for i := range one {
		ready = append(ready, one[i], two[i])
	}
	ready[2].Actions = []string{}
	ready[3].Actions = nil
	transfer := cloneAccessibilityNodes(ready)
	transfer[3].Actions = nil
	flat := accessibilityPendingBaseline(ready)
	bridge.updateOwned(transfer, nil)
	pending := func() []accessibilityNode {
		next := cloneAccessibilityNodes(ready)
		next[3].Actions = nil
		for i := range next {
			n := &next[i]
			if n.GridScope == "one" {
				n.GridPending = true
				n.GridTargetGeneration = 5
				n.Disabled = true
				n.Focused = false
				n.Actions = []string{}
				n.TextClient = 0
				n.TextSelection = nil
				n.TextFocus = 0
			}
		}
		return next
	}
	for name, change := range map[string]func([]accessibilityNode) []accessibilityNode{
		"ready sibling": func(n []accessibilityNode) []accessibilityNode { return n },
		"both pending": func(n []accessibilityNode) []accessibilityNode {
			for i := range n {
				if n[i].GridScope == "two" {
					n[i].GridPending = true
					n[i].GridTargetGeneration = 8
					n[i].Disabled = true
					n[i].Focused = false
					n[i].Actions = []string{}
					n[i].TextClient = 0
					n[i].TextSelection = nil
					n[i].TextFocus = 0
				}
			}
			return n
		},
		"changed ready sibling":   func(n []accessibilityNode) []accessibilityNode { n[1].Name = "changed"; return n },
		"changed completed value": func(n []accessibilityNode) []accessibilityNode { n[4].Value = "changed"; return n },
		"frame":                   func(n []accessibilityNode) []accessibilityNode { n[4].GridFrame[0]++; return n },
		"null pending actions":    func(n []accessibilityNode) []accessibilityNode { n[4].Actions = nil; return n },
		"remove":                  func(n []accessibilityNode) []accessibilityNode { return n[:len(n)-1] },
		"reorder":                 func(n []accessibilityNode) []accessibilityNode { n[2], n[3] = n[3], n[2]; return n },
	} {
		t.Run(name, func(t *testing.T) {
			next := change(pending())
			want := legacySharedBaselinePendingPatch(flat, next, 17)
			got := accessibilityPendingPatchBaseline(bridge.wireBaseline(), next, 17)
			if got != want {
				t.Fatal("reference representation changed pending proof")
			}
			if (name == "ready sibling" || name == "both pending") != (got != "") {
				t.Fatal("proof acceptance/refusal fixture did not exercise expected path")
			}
		})
	}
	first := pending()
	bridge.updateOwned(first, nil)
	if bridge.readyBaseline != nil || bridge.pendingBaseline == nil {
		t.Fatal("pending replacement kept ready refs")
	}
	// Pending preparation groups the replaced scope at its host. A subsequent
	// proof uses the actual emitted order, rather than the earlier interleaving.
	next := cloneAccessibilityNodes(bridge.pendingBaseline)
	for i := range next {
		if bridge.pendingBaseline[i].Actions == nil {
			next[i].Actions = nil
		}
		if next[i].GridScope == "one" {
			next[i].GridTargetGeneration = 6
		}
	}
	want := legacySharedBaselinePendingPatch(bridge.pendingBaseline, next, 18)
	if want == "" || accessibilityPendingPatchBaseline(bridge.wireBaseline(), next, 18) != want {
		t.Fatal("repeated pending changed legacy proof")
	}
}
