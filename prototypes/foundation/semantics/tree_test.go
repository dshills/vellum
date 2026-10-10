package semantics

import "testing"

func TestRejectDanglingAndCycles(t *testing.T) {
	for _, s := range []Snapshot{{Focus: 1, Nodes: []Node{{ID: 1, Children: []uint64{2}}}}, {Focus: 1, Nodes: []Node{{ID: 1}, {ID: 2, Children: []uint64{2}}}}} {
		if s.Validate() == nil {
			t.Fatal("invalid snapshot accepted")
		}
	}
}
func TestActionOnRemovedVirtualRowRejected(t *testing.T) {
	s := Snapshot{Focus: 1, Nodes: []Node{{ID: 1, Children: []uint64{101}}, {ID: 101, Actions: []string{"click"}}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if !s.Allows(Action{Target: 101, Kind: "click"}) {
		t.Fatal("live row rejected")
	}
	s.Nodes = s.Nodes[:1]
	s.Nodes[0].Children = nil
	if s.Allows(Action{Target: 101, Kind: "click"}) {
		t.Fatal("removed row accepted")
	}
}
func TestSnapshotOwnsSlices(t *testing.T) {
	checked, expanded, selected := true, true, true
	s := Snapshot{Focus: 1, Nodes: []Node{{ID: 1, Children: []uint64{2}}, {ID: 2, Actions: []string{"click"}, Checked: &checked, Expanded: &expanded, Selected: &selected}}}
	c := s.Clone()
	s.Nodes[0].Children[0] = 99
	s.Nodes[1].Actions[0] = "set_value"
	checked, expanded, selected = false, false, false
	if c.Nodes[0].Children[0] != 2 || c.Nodes[1].Actions[0] != "click" || !*c.Nodes[1].Checked || !*c.Nodes[1].Expanded || !*c.Nodes[1].Selected {
		t.Fatal("snapshot shares mutable data")
	}
}
