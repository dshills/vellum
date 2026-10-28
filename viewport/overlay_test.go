package viewport

import (
	"image/color"
	"math"
	"testing"
)

func TestOverlayPickingAndDisplayScale(t *testing.T) {
	c := DefaultCamera()
	c.Yaw = 0
	c.Pitch = 0
	s := Scene{Camera: c, Objects: []Object{{ID: "body", Name: "body", Bounds: Box{Vec3{-2, -2, 1}, Vec3{2, 2, 2}}, Color: color.NRGBA{A: 255}, Visible: true}, {ID: "handle", Name: "handle", Bounds: Box{Vec3{-.1, -.1, -.1}, Vec3{.1, .1, .1}}, Color: color.NRGBA{A: 255}, Visible: true, Overlay: true, ScreenSize: 14}}}
	id, e := s.PickScaled(400, 300, 800, 600, 1)
	if e != nil || id != "handle" {
		t.Fatalf("overlay behind body: %s %v", id, e)
	}
	a, _ := s.DisplayObjects(800, 600, 1)
	b, _ := s.DisplayObjects(1600, 1200, 2)
	if math.Abs(a[1].Bounds.Max.X-b[1].Bounds.Max.X) > 1e-9 {
		t.Fatal("display scale mismatch")
	}
	s.Camera.Distance *= 2
	d, _ := s.DisplayObjects(800, 600, 1)
	if math.Abs(d[1].Bounds.Max.X/a[1].Bounds.Max.X-2) > 1e-9 {
		t.Fatal("zoom sizing")
	}
	if s.Objects[1].Bounds.Max.X != .1 {
		t.Fatal("mutated model")
	}
}
