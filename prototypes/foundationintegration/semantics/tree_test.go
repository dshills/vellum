package semantics

import "testing"

func TestRejectUnrepresentableNativeStrings(t *testing.T) {
	for _, value := range []string{"a\x00b", string([]byte{0xff})} {
		for _, node := range []Node{{ID: 1, Label: value}, {ID: 1, Value: value}} {
			if (Snapshot{Focus: 1, Nodes: []Node{node}}).Validate() == nil {
				t.Fatal("native string would truncate or be invalid UTF-8")
			}
		}
	}
}

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

func TestNativeTextSelectionValidation(t *testing.T) {
	valid := Snapshot{Focus: 3, Nodes: []Node{{ID: 1, Children: []uint64{3}}, {ID: 3, Role: "text", Children: []uint64{80}, TextSelection: &TextSelection{Anchor: TextPosition{Node: 80, Character: 0}, Focus: TextPosition{Node: 80, Character: 1}}}, {ID: 80, Role: "textrun", Value: "e\u0301", CharacterLengths: []uint8{3}}}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	copy := valid.Clone()
	valid.Nodes[1].TextSelection.Focus.Character = 99
	valid.Nodes[2].CharacterLengths[0] = 1
	if copy.Nodes[1].TextSelection.Focus.Character != 1 || copy.Nodes[2].CharacterLengths[0] != 3 {
		t.Fatal("text state not owned")
	}
	if valid.Validate() == nil {
		t.Fatal("bad text range accepted")
	}
	copy.Nodes[1].TextSelection.Focus.Node = 3
	if copy.Validate() == nil {
		t.Fatal("non-text-run position accepted")
	}
}

func TestRejectSplitScalar(t *testing.T) {
	s := Snapshot{Focus: 1, Nodes: []Node{{ID: 1, Role: "window", Children: []uint64{2}}, {ID: 2, Role: "textrun", Value: "é", CharacterLengths: []uint8{1, 1}}}}
	if s.Validate() == nil {
		t.Fatal("split UTF8 scalar accepted")
	}
}
