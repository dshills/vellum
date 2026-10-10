// Package semantics holds owned snapshots; native callbacks never inspect widgets.
package semantics

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

type Node struct {
	ID                          uint64
	Role, Label, Value          string
	Bounds                      [4]float64
	Children                    []uint64
	Actions                     []string
	Checked, Expanded, Selected *bool
	RowIndex, RowCount          int
	Invalid                     bool
	CharacterLengths            []uint8
	TextSelection               *TextSelection
}
type Snapshot struct {
	Focus uint64
	Nodes []Node
}
type TextPosition struct {
	Node      uint64
	Character int
}
type TextSelection struct{ Anchor, Focus TextPosition }
type Action struct {
	Target      uint64
	Kind, Value string
	Selection   *TextSelection
}

func (s Snapshot) Validate() error {
	ids := make(map[uint64]bool)
	nodes := make(map[uint64]Node)
	for _, n := range s.Nodes {
		for _, value := range []string{n.Label, n.Value} {
			if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
				return fmt.Errorf("unrepresentable native string for %d", n.ID)
			}
		}
		if n.ID == 0 || ids[n.ID] {
			return fmt.Errorf("invalid or duplicate node %d", n.ID)
		}
		ids[n.ID] = true
		nodes[n.ID] = n
		if n.Role == "textrun" {
			length := 0
			for _, v := range n.CharacterLengths {
				if v == 0 {
					return fmt.Errorf("zero character length")
				}
				length += int(v)
				if length > len(n.Value) || !utf8.ValidString(n.Value[length-int(v):length]) {
					return fmt.Errorf("invalid scalar boundary in text run %d", n.ID)
				}
			}
			if !utf8.ValidString(n.Value) || length != len(n.Value) {
				return fmt.Errorf("invalid text run %d", n.ID)
			}
		}
		for _, v := range n.Bounds {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("non-finite bounds for %d", n.ID)
			}
		}
		if n.Bounds[2] < n.Bounds[0] || n.Bounds[3] < n.Bounds[1] {
			return fmt.Errorf("invalid bounds for %d", n.ID)
		}
	}
	if !ids[1] || !ids[s.Focus] {
		return fmt.Errorf("root or focus missing")
	}
	parent := make(map[uint64]uint64)
	for _, n := range s.Nodes {
		for _, c := range n.Children {
			if !ids[c] || c == 1 || parent[c] != 0 {
				return fmt.Errorf("invalid child %d", c)
			}
			parent[c] = n.ID
		}
	}
	for _, n := range s.Nodes {
		if n.TextSelection != nil {
			for _, pos := range []TextPosition{n.TextSelection.Anchor, n.TextSelection.Focus} {
				run, ok := nodes[pos.Node]
				if !ok || run.Role != "textrun" || pos.Character < 0 || pos.Character > len(run.CharacterLengths) {
					return fmt.Errorf("invalid text selection")
				}
				ancestor := parent[pos.Node]
				for count := 0; ancestor != 0 && ancestor != n.ID && count < len(nodes); count++ {
					ancestor = parent[ancestor]
				}
				if ancestor != n.ID {
					return fmt.Errorf("selection outside text input")
				}
			}
		}
	}
	seen := make(map[uint64]bool)
	var visit func(uint64) error
	visit = func(id uint64) error {
		if seen[id] {
			return fmt.Errorf("cycle at %d", id)
		}
		seen[id] = true
		for _, c := range nodes[id].Children {
			if err := visit(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(1); err != nil {
		return err
	}
	if len(seen) != len(ids) {
		return fmt.Errorf("unreachable node")
	}
	return nil
}
func (s Snapshot) Allows(a Action) bool {
	for _, n := range s.Nodes {
		if n.ID == a.Target {
			for _, kind := range n.Actions {
				if kind == a.Kind {
					return true
				}
			}
		}
	}
	return false
}
func (s Snapshot) Clone() Snapshot {
	r := Snapshot{Focus: s.Focus, Nodes: append([]Node(nil), s.Nodes...)}
	for i := range r.Nodes {
		r.Nodes[i].Children = append([]uint64(nil), r.Nodes[i].Children...)
		r.Nodes[i].Actions = append([]string(nil), r.Nodes[i].Actions...)
		r.Nodes[i].CharacterLengths = append([]uint8(nil), r.Nodes[i].CharacterLengths...)
		if r.Nodes[i].TextSelection != nil {
			value := *r.Nodes[i].TextSelection
			r.Nodes[i].TextSelection = &value
		}
		for _, pair := range []struct {
			src *bool
			dst **bool
		}{{s.Nodes[i].Checked, &r.Nodes[i].Checked}, {s.Nodes[i].Expanded, &r.Nodes[i].Expanded}, {s.Nodes[i].Selected, &r.Nodes[i].Selected}} {
			if pair.src != nil {
				v := *pair.src
				*pair.dst = &v
			}
		}
	}
	return r
}
