//go:build offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"testing"
)

type readyMutatingJSONValue struct{ mutate func() }

func (v readyMutatingJSONValue) MarshalJSON() ([]byte, error) {
	v.mutate()
	return []byte(`"encoded before mutation"`), nil
}

func TestAccessibilityReadyDuplicateCustomEncodingCannotCertifyMutatedInput(t *testing.T) {
	for _, owned := range []bool{false, true} {
		t.Run(fmt.Sprintf("owned-%t", owned), func(t *testing.T) {
			_, bridge := newAXTest(t, 16)
			input := gridAXCompleted("g", "root", 1, 7)
			input[2].Value = readyMutatingJSONValue{mutate: func() {
				input[2].Value = "fresh builtin value"
			}}
			if owned {
				bridge.updateOwned(input, nil)
			} else {
				bridge.update(input, nil)
			}
			initial, generation := bridge.lastSnapshot, bridge.generation.Load()
			var published []accessibilityNode
			if err := json.Unmarshal([]byte(initial), &published); err != nil || len(published) != len(input) || published[2].Value != "encoded before mutation" || generation != 1 {
				t.Fatal("ordinary custom publication was dropped or changed", err)
			}
			if bridge.completedSnapshot != "" || bridge.pendingBaseline != nil {
				t.Fatal("custom encoder certified a ready cache or wire baseline")
			}
			// This independent fresh description must publish its actual value.
			// A post-encoding type scan alone would certify the mutated first
			// input while the native wire still contained the marshaler's output.
			next := gridAXCompleted("g", "root", 1, 7)
			next[2].Value = "fresh builtin value"
			want, err := json.Marshal(next)
			if err != nil {
				t.Fatal(err)
			}
			bridge.updateOwned(next, nil)
			if bridge.lastSnapshot != string(want) || bridge.lastSnapshot == initial || bridge.generation.Load() != generation+1 {
				t.Fatal("custom encoder certified stale wire bytes for a fresh builtin description")
			}
		})
	}
}
