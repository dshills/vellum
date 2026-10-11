package ui

import "testing"

func TestGridCommitDescriptorKindGuard(t *testing.T) {
	n := Node{ID: "field", Kind: TextField, GridCommit: func(GridEditEvent) {}}
	if err := Validate(n); err == nil {
		t.Fatal("grid commit callback admitted on text field")
	}
}
