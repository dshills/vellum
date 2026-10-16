//go:build darwin && cgo

package cogent

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"math/rand"
	"reflect"
	"testing"
)

type axGraphSpecNode struct{ id, parent, scope, role string }

func axGraphPayload(nodes []axGraphSpecNode) []map[string]any {
	out := make([]map[string]any, 0, len(nodes))
	for _, n := range nodes {
		x := map[string]any{"id": n.id, "parent": n.parent, "role": n.role, "actions": []string{}}
		if n.scope != "" {
			x["grid_scope"], x["grid_incarnation"] = n.scope, 1
			x["grid_target_generation"], x["grid_generation"] = 1, 1
			x["grid_frame"] = [4]float64{0, 0, 10, 20}
		}
		out = append(out, x)
	}
	return out
}

// Deliberately use separate ancestor sets and literal scope membership, rather
// than the implementation's shared graph traversal. Other fields are fixed valid
// values here; the existing envelope/integer suites cover their validation.
func axGraphOracle(nodes []axGraphSpecNode) bool {
	byID := map[string]axGraphSpecNode{}
	for _, n := range nodes {
		if n.id == "" {
			return false
		}
		if _, exists := byID[n.id]; exists {
			return false
		}
		byID[n.id] = n
	}
	for _, n := range nodes {
		ancestors := map[string]bool{}
		for cursor := n.id; cursor != ""; {
			current, exists := byID[cursor]
			if !exists || ancestors[cursor] {
				return false
			}
			ancestors[cursor] = true
			cursor = current.parent
		}
		parent := byID[n.parent]
		if n.scope == "" {
			if parent.scope != "" {
				return false
			}
			continue
		}
		host, exists := byID[n.scope]
		if !exists || host.scope != n.scope || (host.role != "table" && host.role != "outline" && host.role != "canvas") {
			return false
		}
		if n.id == n.scope {
			if parent.scope != "" {
				return false
			}
		} else if parent.scope != n.scope || !ancestors[n.scope] {
			return false
		}
	}
	return true
}

func axGraphCheck(t *testing.T, nodes []axGraphSpecNode, digest hash.Hash) bool {
	t.Helper()
	payload, err := json.Marshal(axEnvelopeSnapshot(1, axGraphPayload(nodes)))
	if err != nil {
		t.Fatal(err)
	}
	if digest != nil {
		_, _ = digest.Write(payload)
		_, _ = digest.Write([]byte{'\n'})
	}
	want := axGraphOracle(nodes)
	if got := nativeAccessibilitySnapshotValid(string(payload)); got != want {
		t.Fatalf("graph validation=%v, ancestor oracle=%v: %s", got, want, payload)
	}
	return want
}

func TestNativeAccessibilityGraphExhaustiveParents(t *testing.T) {
	digest := sha256.New()
	cases, accepted := 0, 0
	for count := 0; count <= 4; count++ {
		nodes := make([]axGraphSpecNode, count)
		choices := []string{"", "missing"}
		for i := range nodes {
			nodes[i] = axGraphSpecNode{id: fmt.Sprintf("n%d", i), role: "column"}
			choices = append(choices, nodes[i].id)
		}
		var visit func(int)
		visit = func(index int) {
			if index == count {
				cases++
				if axGraphCheck(t, nodes, digest) {
					accepted++
				}
				return
			}
			for _, parent := range choices {
				nodes[index].parent = parent
				visit(index + 1)
			}
		}
		visit(0)
	}
	t.Logf("parent graphs=%d accepted=%d refused=%d corpus_sha256=%x", cases, accepted, cases-accepted, digest.Sum(nil))
}

func TestNativeAccessibilityGraphExhaustiveScopes(t *testing.T) {
	choices := []string{"", "a", "b", "c", "missing"}
	nodes := []axGraphSpecNode{{id: "a"}, {id: "b"}, {id: "c"}}
	digest := sha256.New()
	cases, accepted := 0, 0
	var visit func(int)
	visit = func(index int) {
		if index == len(nodes) {
			cases++
			if axGraphCheck(t, nodes, digest) {
				accepted++
			}
			return
		}
		for _, parent := range choices {
			for _, scope := range choices {
				nodes[index].parent, nodes[index].scope = parent, scope
				nodes[index].role = "column"
				if scope == nodes[index].id {
					nodes[index].role = "table"
				}
				visit(index + 1)
			}
		}
	}
	visit(0)
	t.Logf("parent/scope graphs=%d accepted=%d refused=%d corpus_sha256=%x", cases, accepted, cases-accepted, digest.Sum(nil))
}

func TestNativeAccessibilityGraphDeepAndUnordered(t *testing.T) {
	const seed = 8317
	rng := rand.New(rand.NewSource(seed))
	for _, deep := range []bool{false, true} {
		nodes := make([]axGraphSpecNode, 1024)
		for i := range nodes {
			nodes[i] = axGraphSpecNode{id: fmt.Sprintf("n%d", i), role: "column"}
			if i > 0 {
				parent := i - 1
				if !deep {
					parent = rng.Intn(i)
				}
				nodes[i].parent = nodes[parent].id
			}
		}
		for range 4 {
			rng.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })
			if !axGraphCheck(t, nodes, nil) {
				t.Fatal("valid shuffled tree refused")
			}
		}
		for _, parent := range []string{"missing", "n1023"} {
			broken := append([]axGraphSpecNode(nil), nodes...)
			for i := range broken {
				if broken[i].id == "n0" {
					broken[i].parent = parent
				}
			}
			if axGraphCheck(t, broken, nil) {
				t.Fatal("deep missing parent or cycle admitted")
			}
		}
	}
	// A scoped table role alone does not designate a new scope host.
	nodes := []axGraphSpecNode{{"host", "", "host", "canvas"}, {"child", "host", "host", "table"}, {"leaf", "child", "host", "outline"}}
	if !axGraphCheck(t, nodes, nil) {
		t.Fatal("non-host role confused with a designated host")
	}
}

func TestNativeAccessibilityGraphRefusalKeepsCommittedTree(t *testing.T) {
	valid := []axGraphSpecNode{{"root", "", "", "column"}, {"host", "root", "host", "table"}, {"child", "host", "host", "grid-cell"}}
	cases := [][]axGraphSpecNode{
		{{"a", "b", "", "column"}, {"b", "a", "", "column"}},
		{{"a", "missing", "", "column"}},
		{{"host", "", "host", "table"}, {"child", "", "host", "grid-cell"}},
		{{"host", "", "host", "table"}, {"child", "host", "", "button"}},
		{{"host", "", "host", "table"}, {"other", "", "other", "outline"}, {"child", "other", "host", "grid-cell"}},
	}
	for i, nodes := range cases {
		if axGraphOracle(nodes) {
			t.Fatal("invalid refusal fixture")
		}
		results := axEnvelopeRun(t, axEnvelopeUpdate(1, axGraphPayload(valid)), axEnvelopeUpdate(2, axGraphPayload(nodes)))
		if !results[0].Accepted || results[1].Accepted || results[1].Generation != 1 || !reflect.DeepEqual(results[0].Elements, results[1].Elements) {
			t.Fatalf("refusal %d changed native publication: %+v", i, results)
		}
	}
}

func BenchmarkNativeAccessibilitySnapshotGraph(b *testing.B) {
	for _, shape := range []struct {
		name  string
		count int
		deep  bool
	}{{"deep64", 64, true}, {"deep256", 256, true}, {"deep1024", 1024, true}, {"wide1024", 1024, false}} {
		b.Run(shape.name, func(b *testing.B) {
			nodes := make([]axGraphSpecNode, shape.count)
			for i := range nodes {
				nodes[i] = axGraphSpecNode{id: fmt.Sprintf("n%d", i), role: "column"}
				if i > 0 {
					parent := "n0"
					if shape.deep {
						parent = nodes[i-1].id
					}
					nodes[i].parent = parent
				}
			}
			payload, err := json.Marshal(axEnvelopeSnapshot(1, axGraphPayload(nodes)))
			if err != nil {
				b.Fatal(err)
			}
			text := string(payload)
			if !nativeAccessibilitySnapshotValid(text) {
				b.Fatal("valid graph refused")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if !nativeAccessibilitySnapshotValid(text) {
					b.Fatal("valid graph refused")
				}
			}
		})
	}
}
