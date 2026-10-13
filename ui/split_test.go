package ui

import (
	"math"
	"reflect"
	"testing"
)

func TestSplitValidationAndFocusOrder(t *testing.T) {
	n := Node{ID: "split", Kind: SplitPane, Number: 0.5, Children: []Node{{ID: "left", Kind: TextField}, {ID: "right", Kind: Button}}}
	if err := Validate(n); err != nil || !IsContainer(SplitPane) || !IsFocusable(SplitPane) {
		t.Fatal("valid split pane refused", err)
	}
	if order := FocusOrder(n); !reflect.DeepEqual(order, []string{"left", "split", "right"}) {
		t.Fatal("divider order does not follow pane order", order)
	}
	for _, edit := range []func(*Node){
		func(n *Node) { n.Children = n.Children[:1] },
		func(n *Node) { n.Children = append(n.Children, Node{ID: "third", Kind: Label}) },
		func(n *Node) { n.Number = float32(math.NaN()) },
		func(n *Node) { n.Number = 0 },
		func(n *Node) { n.Number = 1 },
		func(n *Node) { n.Numeric = Numeric{Min: 0, Max: 0.9, Step: 0.01} },
		func(n *Node) { n.Numeric = Numeric{Min: 0.1, Max: 1, Step: 0.01} },
		func(n *Node) { n.Numeric = Numeric{Min: 0.1, Max: 0.9} },
		func(n *Node) { n.Numeric = Numeric{Min: 0.1, Max: 0.9, Step: 1} },
		func(n *Node) { n.Activate = func() {} },
		func(n *Node) { n.Orientation = "diagonal" },
		func(n *Node) { n.ID = "split:reserved" },
	} {
		copy := n
		copy.Children = append([]Node(nil), n.Children...)
		edit(&copy)
		if err := Validate(copy); err == nil {
			t.Fatal("invalid split descriptor admitted", copy)
		}
	}
	n.Disabled = true
	if len(FocusOrder(n)) != 0 {
		t.Fatal("disabled split exposes focus targets")
	}
	n.Disabled = false
	n.Children[1] = Node{ID: "modal", Kind: Dialog, Open: true, Children: []Node{{ID: "modal-field", Kind: TextField}}}
	if order := FocusOrder(n); !reflect.DeepEqual(order, []string{"modal-field"}) {
		t.Fatal("split escaped modal focus scope", order)
	}
}
