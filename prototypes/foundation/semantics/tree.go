// Package semantics holds owned snapshots; native callbacks never inspect widgets.
package semantics

import "fmt"

type Node struct {
	ID                          uint64
	Role, Label, Value          string
	Bounds                      [4]float64
	Children                    []uint64
	Actions                     []string
	Checked, Expanded, Selected *bool
	RowIndex, RowCount          int
	Invalid                     bool
}
type Snapshot struct {
	Focus uint64
	Nodes []Node
}
type Action struct {
	Target      uint64
	Kind, Value string
}

func (s Snapshot) Validate() error {
	ids := make(map[uint64]bool)
	for _, n := range s.Nodes {
		if n.ID == 0 || ids[n.ID] {
			return fmt.Errorf("invalid or duplicate node %d", n.ID)
		}
		ids[n.ID] = true
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
	seen := make(map[uint64]bool)
	var visit func(uint64) error
	visit = func(id uint64) error {
		if seen[id] {
			return fmt.Errorf("cycle at %d", id)
		}
		seen[id] = true
		for _, n := range s.Nodes {
			if n.ID == id {
				for _, c := range n.Children {
					if err := visit(c); err != nil {
						return err
					}
				}
				break
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
