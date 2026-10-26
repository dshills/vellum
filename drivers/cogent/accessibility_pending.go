package cogent

import (
	"encoding/json"
	"math"
)

// This is a private transport optimization. The full canonical publication is
// still sent: native admission can fall back if its last commit was refused.
type accessibilityPendingScope struct {
	ID          string `json:"id"`
	Incarnation uint64 `json:"incarnation"`
	Completed   uint64 `json:"completed"`
	Target      uint64 `json:"target"`
}

type accessibilityPendingPublication struct {
	Previous   uint64                      `json:"previous"`
	Generation uint64                      `json:"generation"`
	Scopes     []accessibilityPendingScope `json:"scopes"`
}

func accessibilityImmutableScalar(value any) bool {
	switch value.(type) {
	case nil, string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		return true
	default:
		return false
	}
}

func accessibilityPendingBaseline(nodes []accessibilityNode) []accessibilityNode {
	out := cloneAccessibilityNodes(nodes)
	for i := range out {
		if nodes[i].Actions == nil {
			out[i].Actions = nil // Preserve null versus [] in the exact wire baseline.
		}
		out[i].activationGuard = nil
		out[i].pressFocusID = ""
	}
	return out
}

// A ready baseline references the already detached, immutable per-grid cache.
// Nil-action wire spelling is kept separately because cache clones preserve the
// older normalization to a nonnil empty slice. Referenced records never mutate.
type accessibilityNodeRef struct {
	node       *accessibilityNode
	actionsNil bool
}

type accessibilityWireBaseline struct {
	nodes      []accessibilityNode
	references []accessibilityNodeRef
}

func (b *accessibilityBridge) wireBaseline() accessibilityWireBaseline {
	return accessibilityWireBaseline{nodes: b.pendingBaseline, references: b.readyBaseline}
}
func (b accessibilityWireBaseline) len() int {
	if b.references != nil {
		return len(b.references)
	}
	return len(b.nodes)
}
func (b accessibilityWireBaseline) at(i int) accessibilityNode {
	if b.references == nil {
		return b.nodes[i]
	}
	reference := b.references[i]
	node := *reference.node
	if reference.actionsNil {
		node.Actions = nil
	}
	node.activationGuard = nil
	node.pressFocusID = ""
	return node
}

// Prove exact membership, order and wire semantics against the last emitted
// detached snapshot. A pending subtree may only disable actions/focus and text
// editing while retaining its completed identity, values and frames. Host label
// changes, new incarnations, removals and any ordinary change use full admission.
func accessibilityPendingPatch(previous, next []accessibilityNode, generation uint64) string {
	return accessibilityPendingPatchBaseline(accessibilityWireBaseline{nodes: previous}, next, generation)
}

func accessibilityPendingPatchBaseline(previous accessibilityWireBaseline, next []accessibilityNode, generation uint64) string {
	if generation == 0 || generation == math.MaxUint64 || previous.len() == 0 || previous.len() != len(next) {
		return ""
	}
	patch := accessibilityPendingPublication{Previous: generation, Generation: generation + 1}
	scopes := make(map[string]accessibilityPendingScope)
	for _, n := range next {
		if n.GridScope != "" && n.ID == n.GridScope && n.GridPending {
			if n.GridIncarnation == 0 || n.GridGeneration == 0 || n.GridTargetGeneration < n.GridGeneration || n.GridFrame == nil {
				return ""
			}
			if _, exists := scopes[n.ID]; exists {
				return ""
			}
			scope := accessibilityPendingScope{n.ID, n.GridIncarnation, n.GridGeneration, n.GridTargetGeneration}
			scopes[n.ID] = scope
			patch.Scopes = append(patch.Scopes, scope)
		}
	}
	if len(scopes) == 0 {
		return ""
	}
	for i, n := range next {
		old := previous.at(i)
		if scope, pending := scopes[n.GridScope]; pending {
			if !n.GridPending || n.GridIncarnation != scope.Incarnation || n.GridGeneration != scope.Completed || n.GridTargetGeneration != scope.Target {
				return ""
			}
			old.GridPending, old.GridTargetGeneration = true, scope.Target
			old.Disabled, old.Focused, old.Actions = true, false, []string{}
			old.TextClient, old.TextSelection, old.TextFocus = 0, nil, 0
		}
		if !accessibilitySameWireNode(old, n) {
			return ""
		}
	}
	encoded, err := json.Marshal(patch)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func accessibilitySameWireNode(a, b accessibilityNode) bool {
	if a.ID != b.ID || a.DomainID != b.DomainID || a.Parent != b.Parent || a.Role != b.Role || a.Secure != b.Secure || a.Name != b.Name || a.Help != b.Help ||
		!accessibilitySameScalar(a.Value, b.Value) || a.Disabled != b.Disabled || a.Focusable != b.Focusable || a.Focused != b.Focused || a.Selected != b.Selected ||
		a.Orientation != b.Orientation || a.RowIndex != b.RowIndex || a.RowCount != b.RowCount || a.ColumnIndex != b.ColumnIndex || a.ColumnCount != b.ColumnCount ||
		a.GridGeneration != b.GridGeneration || a.GridScope != b.GridScope || a.GridIncarnation != b.GridIncarnation || a.GridPending != b.GridPending || a.GridTargetGeneration != b.GridTargetGeneration ||
		!accessibilitySamePointer(a.NativeContent, b.NativeContent) || a.Level != b.Level || a.TextClient != b.TextClient || a.TextFocus != b.TextFocus ||
		!accessibilitySamePointer(a.Expanded, b.Expanded) || !accessibilitySamePointer(a.TextSelection, b.TextSelection) ||
		!accessibilitySameFloatPointer(a.Min, b.Min) || !accessibilitySameFloatPointer(a.Max, b.Max) || (a.Actions == nil) != (b.Actions == nil) || len(a.Actions) != len(b.Actions) ||
		(a.GridFrame == nil) != (b.GridFrame == nil) {
		return false
	}
	for i := range a.Actions {
		if a.Actions[i] != b.Actions[i] {
			return false
		}
	}
	if a.GridFrame != nil {
		for i, v := range *a.GridFrame {
			if math.Float32bits(v) != math.Float32bits(b.GridFrame[i]) {
				return false
			}
		}
	}
	return true
}

func accessibilitySamePointer[T comparable](a, b *T) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func accessibilitySameFloatPointer(a, b *float32) bool {
	return a == nil && b == nil || a != nil && b != nil && math.Float32bits(*a) == math.Float32bits(*b)
}

func accessibilitySameScalar(a, b any) bool {
	switch v := a.(type) {
	case float32:
		w, ok := b.(float32)
		return ok && math.Float32bits(v) == math.Float32bits(w)
	case float64:
		w, ok := b.(float64)
		return ok && math.Float64bits(v) == math.Float64bits(w)
	case nil, string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		// Type assertion above deliberately excludes custom and mutable values.
		// Interface comparison against a different (even noncomparable) dynamic
		// type returns false without comparing that value.
		return a == b
	default:
		return false
	}
}
