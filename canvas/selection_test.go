package canvas

import "testing"

func TestExactObjectSelectionAndPureHitTest(t *testing.T) {
	s, err := New([]Rect{{ID: "rear", Width: 40, Height: 40}, {ID: "front", Width: 40, Height: 40}})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Select("rear"); err != nil {
		t.Fatal(err)
	}
	point := s.View.Screen(Point{20, 20})
	if got := s.ObjectAt(point); got != "front" || s.Selected != "rear" {
		t.Fatal("hit test mutated selection or lost paint order", got, s.Selected)
	}
	if err := s.Select("missing"); err == nil || s.Selected != "rear" {
		t.Fatal("unknown ID changed selection")
	}
	if s.Pick(point) != "front" {
		t.Fatal("pointer picking changed")
	}
	if err := s.Select(""); err != nil || s.Selected != "" {
		t.Fatal("clear selection", err)
	}
}
