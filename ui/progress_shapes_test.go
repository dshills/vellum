package ui

import (
	"math"
	"reflect"
	"testing"
)

func TestProgressShapesAreControlledPassiveLeaves(t *testing.T) {
	for _, kind := range []Kind{ProgressRing, ProgressGauge} {
		for _, fraction := range []float32{0, 0.25, 1} {
			n := Node{ID: "progress", Kind: kind, Number: fraction, Text: "25%"}
			if err := Validate(n); err != nil {
				t.Fatal(kind, err)
			}
			root := Node{ID: "root", Kind: Column, Children: []Node{n, {ID: "action", Kind: Button}}}
			if order := FocusOrder(root); !reflect.DeepEqual(order, []string{"action"}) {
				t.Fatal("progress shape entered focus order", kind, order)
			}
		}
		for _, fraction := range []float32{-0.01, 1.01, float32(math.NaN()), float32(math.Inf(1))} {
			if err := Validate(Node{ID: "bad", Kind: kind, Number: fraction}); err == nil {
				t.Fatal("invalid fraction accepted", kind, fraction)
			}
		}
		for _, invalid := range []Node{
			{ID: "bad", Kind: kind, Activate: func() {}},
			{ID: "bad", Kind: kind, Change: func(ChangeEvent) {}},
			{ID: "bad", Kind: kind, Orientation: Vertical},
			{ID: "bad", Kind: kind, Children: []Node{{ID: "child", Kind: Label}}},
		} {
			if err := Validate(invalid); err == nil {
				t.Fatal("invalid progress descriptor accepted", invalid)
			}
		}
	}
}
