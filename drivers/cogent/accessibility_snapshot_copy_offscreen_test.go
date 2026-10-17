//go:build offscreen

package cogent

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

// Frozen single-record copy from 5ba6e86. Keep this independent of the batch
// implementation so its existing detachment and nil-action behavior are tested.
func cloneAccessibilityNode(n accessibilityNode) accessibilityNode {
	n.Actions = append([]string{}, n.Actions...)
	if n.GridFrame != nil {
		value := *n.GridFrame
		n.GridFrame = &value
	}
	if n.TextSelection != nil {
		value := *n.TextSelection
		n.TextSelection = &value
	}
	if n.Expanded != nil {
		value := *n.Expanded
		n.Expanded = &value
	}
	if n.Min != nil {
		value := *n.Min
		n.Min = &value
	}
	if n.Max != nil {
		value := *n.Max
		n.Max = &value
	}
	return n
}

func legacyAccessibilitySnapshotCopy(nodes []accessibilityNode) []accessibilityNode {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]accessibilityNode, len(nodes))
	for i := range nodes {
		out[i] = cloneAccessibilityNode(nodes[i])
	}
	return out
}

func accessibilityCopyFixture(count int) []accessibilityNode {
	frame, selection, expanded := [4]float32{1, 2, 3, 4}, [2]int{1, 3}, true
	minimum, maximum := float32(-1), float32(10)
	nodes := make([]accessibilityNode, count)
	for i := range nodes {
		nodes[i] = accessibilityNode{ID: fmt.Sprint(i), Name: "A😀עברית<>&", Value: i,
			Actions: []string{"press", "focus"}, GridFrame: &frame,
			TextSelection: &selection, Expanded: &expanded, Min: &minimum, Max: &maximum}
	}
	return nodes
}

func TestAccessibilitySnapshotCopyMatchesLegacy(t *testing.T) {
	nodes := accessibilityCopyFixture(17)
	for i := range nodes {
		if i%2 == 0 {
			nodes[i].GridFrame, nodes[i].Min, nodes[i].Expanded = nil, nil, nil
		}
		if i%3 == 0 {
			nodes[i].Actions, nodes[i].TextSelection, nodes[i].Max = nil, nil, nil
		} else if i%3 == 1 {
			nodes[i].Actions = []string{}
		}
	}
	for _, input := range [][]accessibilityNode{
		nil, {},
		// No action backing is needed in either case. Both nil and empty input
		// Actions must still become nonnil empty slices, as in the prior clone.
		{{ID: "nil-one"}, {ID: "nil-two"}},
		{{ID: "nil"}, {ID: "empty", Actions: []string{}}},
		nodes, gridAXPayloadFixture(400),
	} {
		got, want := cloneAccessibilityNodes(input), legacyAccessibilitySnapshotCopy(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatal("batch copy changed legacy records or nil/empty optional fields")
		}
		for i := range got {
			if !accessibilitySameWireNode(got[i], want[i]) {
				t.Fatalf("record %d changed exact wire fields", i)
			}
		}
	}
	// Pointer copies must preserve exact float bits, including signed zero and
	// nonfinite values rejected later by encoding. Copying does not validate.
	for _, bits := range []uint32{0, 0x80000000, 0x7fc01234, 0xff800000} {
		v := math.Float32frombits(bits)
		frame := [4]float32{v, v, v, v}
		input := []accessibilityNode{{GridFrame: &frame, Min: &v, Max: &v}}
		if !accessibilitySameWireNode(cloneAccessibilityNodes(input)[0], cloneAccessibilityNode(input[0])) {
			t.Fatalf("copy changed float bits %08x", bits)
		}
	}
}

func TestAccessibilitySnapshotCopyIsolation(t *testing.T) {
	input := accessibilityCopyFixture(3) // All pointer fields deliberately alias.
	want := legacyAccessibilitySnapshotCopy(input)
	out := cloneAccessibilityNodes(input)
	out[0].Actions[0] = "changed"
	out[0].GridFrame[0], out[0].TextSelection[0] = 99, 99
	*out[0].Expanded, *out[0].Min, *out[0].Max = false, 99, 99
	if !reflect.DeepEqual(input, want) || !reflect.DeepEqual(out[1:], want[1:]) {
		t.Fatal("output mutation escaped into input or another copied node")
	}
	for i := range out {
		if cap(out[i].Actions) != len(out[i].Actions) {
			t.Fatal("action append can overwrite a neighboring node")
		}
	}
	out[0].Actions = append(out[0].Actions, "new")
	if !reflect.DeepEqual(out[1].Actions, want[1].Actions) {
		t.Fatal("appended action overwrote another node")
	}
	before := legacyAccessibilitySnapshotCopy(out)
	input[1].Actions[0], input[1].GridFrame[0] = "source mutation", 500
	next := cloneAccessibilityNodes(input)
	next[1].GridFrame[0] = 600
	if !reflect.DeepEqual(out, before) {
		t.Fatal("source mutation or later copy overwrote retained snapshot storage")
	}
	// Batch copy retains the prior shallow Value contract; mutable/custom values
	// still take the publication's non-reuse path rather than gaining a new cache.
	value := []byte{1, 2}
	custom := cloneAccessibilityNodes([]accessibilityNode{{Value: value}})
	if &custom[0].Value.([]byte)[0] != &value[0] {
		t.Fatal("copy changed the established Value contract")
	}
}

func TestAccessibilitySnapshotCopyAllocationBound(t *testing.T) {
	for _, count := range []int{1, 32, 512} {
		nodes := accessibilityCopyFixture(count)
		allocations := testing.AllocsPerRun(20, func() {
			accessibilityPublicationBenchmarkSink = cloneAccessibilityNodes(nodes)
		})
		if allocations > 7 {
			t.Fatalf("%d nodes: %.0f allocations exceed one node slice and six field slices", count, allocations)
		}
	}
}

func BenchmarkAccessibilitySnapshotCopy(b *testing.B) {
	nodes := accessibilityCopyFixture(400)
	for _, tc := range []struct {
		name string
		copy func([]accessibilityNode) []accessibilityNode
	}{{"legacy", legacyAccessibilitySnapshotCopy}, {"batch", cloneAccessibilityNodes}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				accessibilityPublicationBenchmarkSink = tc.copy(nodes)
			}
		})
	}
}
