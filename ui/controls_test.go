package ui

import (
	"math"
	"testing"
)

func TestControlValidationAndFocus(t *testing.T) {
	root := Node{ID: "root", Kind: Column, Focus: FocusRequest{ID: "edit", Sequence: 1}, Children: []Node{{ID: "edit", Kind: TextField}, {ID: "slider", Kind: Slider, Numeric: Numeric{0, 100, 1}, Number: 50}}}
	if err := Validate(root); err != nil {
		t.Fatal(err)
	}
	root.Children[0].Disabled = true
	if _, ok := FocusTarget(root, "edit"); ok {
		t.Fatal("disabled focus accepted")
	}
	root.Children[0].Disabled = false
	root.Children[0].Focus = FocusRequest{ID: "slider"}
	if err := Validate(root); err == nil {
		t.Fatal("nested focus request accepted")
	}
	for _, n := range []Node{
		{ID: "bad", Kind: Slider, Numeric: Numeric{0, 10, 1}, Number: float32(math.NaN())},
		{ID: "bad", Kind: NumberField, Numeric: Numeric{0, 10, 0}, Number: 1},
		{ID: "bad", Kind: Slider, Numeric: Numeric{-math.MaxFloat32, math.MaxFloat32, 1}},
		{ID: "bad", Kind: Selector, Options: []Option{{ID: "same"}, {ID: "same"}}},
		{ID: "bad", Kind: Selector, Options: []Option{{ID: "yes"}}, SelectedID: "absent"},
		{ID: "bad", Kind: Tabs, Options: []Option{{ID: "page"}}, SelectedID: "page", Children: []Node{{ID: "other", Kind: Column}}},
	} {
		if err := Validate(n); err == nil {
			t.Errorf("accepted invalid node %+v", n)
		}
	}
}
func TestFocusVisibilityAndFiniteClamp(t *testing.T) {
	root := Node{ID: "tabs", Kind: Tabs, SelectedID: "a", Options: []Option{{ID: "a"}, {ID: "b"}}, Children: []Node{{ID: "a", Kind: Column, Children: []Node{{ID: "first", Kind: TextField}}}, {ID: "b", Kind: Column, Children: []Node{{ID: "hidden", Kind: TextField}}}}}
	if _, ok := FocusTarget(root, "first"); !ok {
		t.Fatal("selected tab field not focusable")
	}
	if _, ok := FocusTarget(root, "hidden"); ok {
		t.Fatal("hidden tab field focusable")
	}
	root = Node{ID: "dialog", Kind: Dialog, Children: []Node{{ID: "field", Kind: TextField}}}
	if _, ok := FocusTarget(root, "field"); ok {
		t.Fatal("closed dialog focusable")
	}
	root.Open = true
	if _, ok := FocusTarget(root, "field"); !ok {
		t.Fatal("open dialog field unavailable")
	}
	if value, ok := ClampNumeric(20, Numeric{0, 10, 1}); !ok || value != 10 {
		t.Fatal("finite clamp failed")
	}
	if _, ok := ClampNumeric(float32(math.Inf(1)), Numeric{0, 10, 1}); ok {
		t.Fatal("infinite native input accepted")
	}
}

func TestTabsFocusRequiresEnabledOption(t *testing.T) {
	n := Node{ID: "tabs", Kind: Tabs, SelectedID: "a", Options: []Option{{ID: "a", Disabled: true}, {ID: "b"}}, Children: []Node{{ID: "a", Kind: Column}, {ID: "b", Kind: Column}}}
	if _, ok := FocusTarget(n, "tabs"); !ok {
		t.Fatal("enabled fallback tab unavailable")
	}
	n.Options[1].Disabled = true
	if _, ok := FocusTarget(n, "tabs"); ok {
		t.Fatal("all-disabled tabs focusable")
	}
}
