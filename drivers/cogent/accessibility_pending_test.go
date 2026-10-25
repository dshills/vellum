//go:build offscreen

package cogent

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func pendingAXNodes(nodes []accessibilityNode, target uint64) []accessibilityNode {
	out := cloneAccessibilityNodes(nodes)
	for i := range out {
		if out[i].GridScope == "" {
			continue
		}
		n := &out[i]
		n.GridPending, n.GridTargetGeneration = true, target
		n.Disabled, n.Focused, n.Actions = true, false, []string{}
		n.TextClient, n.TextSelection, n.TextFocus = 0, nil, 0
	}
	return out
}

func TestAccessibilityPendingProof(t *testing.T) {
	previous := append([]accessibilityNode{{ID: "root", Role: "column", Actions: []string{}}}, gridAXCompleted("g", "root", 7, 9)...)
	next := pendingAXNodes(previous, 10)
	encoded := accessibilityPendingPatch(previous, next, 17)
	var patch accessibilityPendingPublication
	if err := json.Unmarshal([]byte(encoded), &patch); err != nil || patch.Previous != 17 || patch.Generation != 18 || !reflect.DeepEqual(patch.Scopes, []accessibilityPendingScope{{"g", 7, 9, 10}}) {
		t.Fatalf("pending proof: %s %v", encoded, err)
	}
	for name, mutate := range map[string]func([]accessibilityNode) []accessibilityNode{
		"ordinary control": func(n []accessibilityNode) []accessibilityNode { n[0].Name = "new"; return n },
		"host label":       func(n []accessibilityNode) []accessibilityNode { n[1].Name = "new"; return n },
		"host help":        func(n []accessibilityNode) []accessibilityNode { n[1].Help = "new"; return n },
		"reparent":         func(n []accessibilityNode) []accessibilityNode { n[1].Parent = "new"; return n },
		"incarnation": func(n []accessibilityNode) []accessibilityNode {
			for i := 1; i < len(n); i++ {
				n[i].GridIncarnation++
			}
			return n
		},
		"value":               func(n []accessibilityNode) []accessibilityNode { n[3].Value = "new"; return n },
		"frame":               func(n []accessibilityNode) []accessibilityNode { n[3].GridFrame[0]++; return n },
		"remove":              func(n []accessibilityNode) []accessibilityNode { return n[:len(n)-1] },
		"reorder":             func(n []accessibilityNode) []accessibilityNode { n[2], n[3] = n[3], n[2]; return n },
		"nil actions":         func(n []accessibilityNode) []accessibilityNode { n[3].Actions = nil; return n },
		"active actions":      func(n []accessibilityNode) []accessibilityNode { n[3].Actions = []string{"press"}; return n },
		"enabled":             func(n []accessibilityNode) []accessibilityNode { n[3].Disabled = false; return n },
		"focused":             func(n []accessibilityNode) []accessibilityNode { n[3].Focused = true; return n },
		"text client":         func(n []accessibilityNode) []accessibilityNode { n[5].TextClient = 9; return n },
		"text selection":      func(n []accessibilityNode) []accessibilityNode { n[5].TextSelection = &[2]int{1, 2}; return n },
		"text focus":          func(n []accessibilityNode) []accessibilityNode { n[5].TextFocus = 1; return n },
		"incomplete pending":  func(n []accessibilityNode) []accessibilityNode { n[3].GridPending = false; return n },
		"inconsistent target": func(n []accessibilityNode) []accessibilityNode { n[3].GridTargetGeneration++; return n },
		"startup":             func(n []accessibilityNode) []accessibilityNode { n[1].GridGeneration = 0; return n },
	} {
		t.Run(name, func(t *testing.T) {
			if got := accessibilityPendingPatch(previous, mutate(cloneAccessibilityNodes(next)), 17); got != "" {
				t.Fatalf("unsafe patch: %s", got)
			}
		})
	}
	for _, generation := range []uint64{0, math.MaxUint64} {
		if accessibilityPendingPatch(previous, next, generation) != "" {
			t.Fatal("unsafe generation")
		}
	}
	if accessibilityPendingPatch(nil, next, 17) != "" || accessibilityPendingPatch(previous, previous, 17) != "" {
		t.Fatal("unanchored or ready patch")
	}
	// Multiple scopes, including one that stays ready, retain exact semantics.
	two := gridAXCompleted("two", "root", 8, 9)
	if accessibilityPendingPatch(append(cloneAccessibilityNodes(previous), two...), append(cloneAccessibilityNodes(next), two...), 17) == "" {
		t.Fatal("ready sibling rejected")
	}
	if accessibilityPendingPatch(append(cloneAccessibilityNodes(previous), two...), append(cloneAccessibilityNodes(next), pendingAXNodes(two, 11)...), 17) == "" {
		t.Fatal("second pending scope rejected")
	}
	if accessibilityPendingPatch(next, pendingAXNodes(previous, 12), 18) == "" {
		t.Fatal("repeated pending target rejected")
	}
}

func TestAccessibilityPendingWireFieldAudit(t *testing.T) {
	// This list intentionally requires a decision when any wire field is added.
	fields := strings.Fields("ID DomainID Parent Role Secure Name Help Value Disabled Focusable Focused Selected Actions Orientation RowIndex RowCount ColumnIndex ColumnCount GridGeneration GridScope GridIncarnation GridPending GridTargetGeneration GridFrame Level Expanded Min TextClient TextSelection TextFocus Max")
	want := map[string]bool{}
	for _, name := range fields {
		want[name] = true
	}
	base := gridAXCompleted("g", "", 7, 9)[4]
	typ := reflect.TypeOf(base)
	seen := 0
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("json") == "" {
			continue
		}
		seen++
		if !want[field.Name] {
			t.Fatalf("unreviewed wire field %s", field.Name)
		}
		t.Run(field.Name, func(t *testing.T) {
			changed := cloneAccessibilityNode(base)
			v := reflect.ValueOf(&changed).Elem().Field(i)
			switch v.Kind() {
			case reflect.String:
				v.SetString(v.String() + "different")
			case reflect.Bool:
				v.SetBool(!v.Bool())
			case reflect.Int:
				v.SetInt(v.Int() + 1)
			case reflect.Uint64:
				v.SetUint(v.Uint() + 1)
			case reflect.Interface:
				v.Set(reflect.ValueOf("different"))
			case reflect.Slice:
				v.Set(reflect.ValueOf([]string{"different"}))
			case reflect.Pointer:
				p := reflect.New(v.Type().Elem())
				if !v.IsNil() {
					p.Elem().Set(v.Elem())
				}
				item := p.Elem()
				if item.Kind() == reflect.Array {
					item = item.Index(0)
				}
				switch item.Kind() {
				case reflect.Bool:
					item.SetBool(!item.Bool())
				case reflect.Int:
					item.SetInt(item.Int() + 1)
				case reflect.Float32:
					item.SetFloat(item.Float() + 1)
				default:
					t.Fatal("unreviewed pointer field")
				}
				v.Set(p)
			default:
				t.Fatal("unreviewed field type")
			}
			if accessibilitySameWireNode(base, changed) {
				t.Fatal("changed wire field ignored")
			}
			if !accessibilitySameWireNode(base, cloneAccessibilityNode(base)) {
				t.Fatal("equivalent detached node differs")
			}
		})
	}
	if seen != len(want) {
		t.Fatal("stale wire audit")
	}
	for _, values := range [][2]any{{int(1), float64(1)}, {float32(0), float32(math.Copysign(0, -1))}, {float64(0), math.Copysign(0, -1)}, {[]int{1}, []int{1}}, {map[string]int{}, map[string]int{}}} {
		if accessibilitySameScalar(values[0], values[1]) {
			t.Fatal("nonexact or mutable scalar admitted")
		}
	}
}

func TestAccessibilityPendingBaselineDetachedAndBounded(t *testing.T) {
	nodes := gridAXCompleted("g", "root", 7, 9)
	nodes[0].activationGuard = func() bool { return true }
	nodes[0].pressFocusID = "owner"
	nodes[1].Actions = nil
	baseline := accessibilityPendingBaseline(nodes)
	if baseline[0].activationGuard != nil || baseline[0].pressFocusID != "" || baseline[1].Actions != nil {
		t.Fatal("callback retention or wire normalization")
	}
	nodes[0].Actions[0] = "changed"
	nodes[0].GridFrame[0]++
	*nodes[4].Expanded = false
	nodes[4].TextSelection[0]++
	*nodes[4].Min = 20
	if baseline[0].Actions[0] != "focus" || baseline[0].GridFrame[0] != 10 || !*baseline[4].Expanded || baseline[4].TextSelection[0] != 1 || *baseline[4].Min != 0 {
		t.Fatal("baseline aliases input")
	}
	b := &accessibilityBridge{done: make(chan struct{})}
	b.updateOwned(cloneAccessibilityNodes(baseline), nil)
	if b.wireBaseline().len() != len(baseline) {
		t.Fatal("missing owned baseline")
	}
	b.update(cloneAccessibilityNodes(baseline), nil)
	if b.pendingBaseline != nil || b.readyBaseline != nil {
		t.Fatal("borrowed duplicate retained proof baseline")
	}
	b.updateOwned(pendingAXNodes(baseline, 10), nil)
	b.close()
	if b.pendingBaseline != nil || b.readyBaseline != nil {
		t.Fatal("close retained proof baseline")
	}
}

func TestAccessibilityPendingRestoredCustomValueRevokesBaseline(t *testing.T) {
	for _, value := range []any{duplicateAXCustomValue{Hidden: 1}, []string{"mutable"}} {
		b := &accessibilityBridge{done: make(chan struct{})}
		nodes := gridAXCompleted("g", "root", 7, 9)
		nodes[2].Value = value
		b.updateOwned(nodes, nil)
		if (b.pendingBaseline != nil || b.readyBaseline != nil) || len(b.gridPublications["g"]) == 0 {
			t.Fatal("nonbuiltin completion entered proof baseline or missed cache setup")
		}
		b.updateOwned(gridAXPending("g", "root", 7, 10), nil)
		if b.pendingBaseline != nil || b.readyBaseline != nil {
			t.Fatal("raw builtin host certified restored nonbuiltin completed values")
		}
		b.close()
	}
}

func TestAccessibilityPendingEpochExhaustionPermanentlyRevokesProof(t *testing.T) {
	b := &accessibilityBridge{done: make(chan struct{})}
	defer b.close()
	ready := gridAXCompleted("g", "", 7, 9)
	b.updateOwned(cloneAccessibilityNodes(ready), nil)
	if b.pendingExhausted || b.wireBaseline().len() == 0 {
		t.Fatal("ordinary publication did not establish eligible baseline")
	}
	// Seed the formal boundary rather than generating 2^64 publications. The
	// legacy full route continues; this optimization may never use a reused epoch.
	b.generation.Store(math.MaxUint64)
	b.updateOwned(pendingAXNodes(ready, 10), nil)
	if b.generation.Load() != 0 || !b.pendingExhausted || (b.pendingBaseline != nil || b.readyBaseline != nil) {
		t.Fatal("counter wrap preserved fast-path eligibility")
	}
	for generation := uint64(11); generation < 14; generation++ {
		nodes := gridAXCompleted("g", "", 7, generation)
		b.updateOwned(cloneAccessibilityNodes(nodes), nil)
		b.updateOwned(pendingAXNodes(nodes, generation+1), nil)
		if !b.pendingExhausted || (b.pendingBaseline != nil || b.readyBaseline != nil) {
			t.Fatal("a later owned ready/pending publication revived an exhausted proof")
		}
	}
}
