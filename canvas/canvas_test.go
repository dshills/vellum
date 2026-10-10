package canvas

import (
	"math"
	"testing"
)

func TestPointerAnchoredZoomAndPicking(t *testing.T) {
	s, err := New([]Rect{{ID: "back", X: 0, Y: 0, Width: 100, Height: 100}, {ID: "front", X: 20, Y: 20, Width: 10, Height: 10}})
	if err != nil {
		t.Fatal(err)
	}
	s.View.SetSize(800, 600)
	s.View.Center = Point{50, 50}
	p := s.View.Screen(Point{25, 25})
	before := s.View.World(p)
	if err := s.View.ZoomAt(p, 4); err != nil {
		t.Fatal(err)
	}
	after := s.View.World(p)
	if math.Abs(before.X-after.X) > 1e-9 || math.Abs(before.Y-after.Y) > 1e-9 {
		t.Fatal("zoom anchor moved")
	}
	if s.Pick(p) != "front" {
		t.Fatal(s.Selected)
	}
	s.View.Pan(100, 0)
	if s.Pick(p) == "front" {
		t.Fatal("pan not reflected in picking")
	}
	if s.View.ZoomAt(p, math.NaN()) == nil {
		t.Fatal("nonfinite zoom")
	}
}
func TestGeometryOwnedAndBounded(t *testing.T) {
	objects := []Rect{{ID: "a", Width: 10, Height: 10}}
	s, err := New(objects)
	if err != nil {
		t.Fatal(err)
	}
	objects[0].ID = "mutated"
	if s.Objects[0].ID != "a" {
		t.Fatal("unowned geometry")
	}
	if _, err := New([]Rect{{ID: "x", Width: 0, Height: 10}}); err == nil {
		t.Fatal("empty extent")
	}
	s.Fit()
	if s.View.Zoom <= 0 {
		t.Fatal("fit")
	}
}

func TestInvalidNavigationDoesNotMutateViewport(t *testing.T) {
	v := Viewport{Width: 800, Height: 600, Zoom: 1}
	old := v
	if v.ZoomAt(Point{X: math.MaxFloat64}, 100) == nil || v != old {
		t.Fatal("overflowing zoom changed viewport")
	}
	if v.Pan(math.Inf(1), 0) == nil || v != old {
		t.Fatal("invalid pan changed viewport")
	}
	v.Zoom = math.NaN()
	if v.Pan(1, 1) == nil || v.ZoomAt(Point{}, 2) == nil {
		t.Fatal("accepted nonfinite zoom")
	}
}
