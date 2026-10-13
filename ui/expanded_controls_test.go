package ui

import (
	"math"
	"reflect"
	"testing"
)

func TestExpandedControlsValidationAndFocus(t *testing.T) {
	root := Node{ID: "root", Kind: Column, Children: []Node{
		{ID: "switch", Kind: Switch, Checked: true}, {ID: "toggle", Kind: ToggleButton},
		{ID: "progress", Kind: ProgressBar, Number: 0.5}, {ID: "separator", Kind: Separator, Orientation: Vertical},
	}}
	if err := Validate(root); err != nil {
		t.Fatal(err)
	}
	if order := FocusOrder(root); !reflect.DeepEqual(order, []string{"switch", "toggle"}) {
		t.Fatal("passive controls entered focus order", order)
	}
	for _, kind := range []Kind{Switch, ToggleButton, ProgressBar, Separator} {
		n := Node{ID: "leaf", Kind: kind, Children: []Node{{ID: "child", Kind: Label}}}
		if err := Validate(n); err == nil {
			t.Fatal("leaf accepted children", kind)
		}
	}
	for _, value := range []float32{-0.01, 1.01, float32(math.NaN()), float32(math.Inf(1))} {
		if err := Validate(Node{ID: "bad", Kind: ProgressBar, Number: value}); err == nil {
			t.Fatal("invalid fraction accepted", value)
		}
	}
	for _, value := range []float32{0, 1} {
		if err := Validate(Node{ID: "edge", Kind: ProgressBar, Number: value}); err != nil {
			t.Fatal(err)
		}
	}
	for _, kind := range []Kind{ProgressBar, Separator} {
		if err := Validate(Node{ID: "passive", Kind: kind, Change: func(ChangeEvent) {}}); err == nil {
			t.Fatal("passive change callback accepted")
		}
		if err := Validate(Node{ID: "passive", Kind: kind, Activate: func() {}}); err == nil {
			t.Fatal("passive activation callback accepted")
		}
		if err := Validate(Node{ID: "bad", Kind: kind, Orientation: "diagonal"}); err == nil {
			t.Fatal("unknown orientation accepted")
		}
	}
	if err := Validate(Node{ID: "bad", Kind: Switch, Orientation: Horizontal}); err == nil {
		t.Fatal("switch accepted display orientation")
	}
	root.Children[0].Disabled = true
	if order := FocusOrder(root); !reflect.DeepEqual(order, []string{"toggle"}) {
		t.Fatal("disabled switch focused", order)
	}
}
