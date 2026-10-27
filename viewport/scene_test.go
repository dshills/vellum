package viewport

import (
	"image/color"
	"testing"
)

func TestSceneNearestVisibleAndCubeScaling(t *testing.T) {
	c := DefaultCamera()
	c.SetView(Front)
	s := &Scene{Camera: c, Objects: []Object{{ID: "far", Name: "Far", Bounds: Box{Vec3{-1, -1, -4}, Vec3{1, 1, -2}}, Color: color.NRGBA{A: 255}, Visible: true}, {ID: "near", Name: "Near", Bounds: Box{Vec3{-1, -1, -1}, Vec3{1, 1, 1}}, Color: color.NRGBA{A: 255}, Visible: true}}}
	if id, e := s.Pick(400, 300, 800, 600); e != nil || id != "near" {
		t.Fatal(id, e)
	}
	s.Objects[1].Visible = false
	if id, _ := s.Pick(400, 300, 800, 600); id != "far" {
		t.Fatal(id)
	}
	for _, scale := range []float64{1, 1.5, 2} {
		rect := CubeRect(int(800*scale), int(600*scale), scale)
		pt := rect.Min.Add(rect.Size().Div(2))
		if v, ok := CubeViewAt(c, pt.X, pt.Y, int(800*scale), int(600*scale), scale); !ok || v != Front {
			t.Fatal(scale, v, ok)
		}
	}
	if !CubeRect(40, 40, 2).Empty() {
		t.Fatal("tiny overlay")
	}
	s.SelectedID = "gone"
	if s.Validate() == nil {
		t.Fatal("missing selection")
	}
}
