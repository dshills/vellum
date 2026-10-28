package viewport

import (
	"math"
	"testing"
)

func TestCylinderPickingSilhouetteCapsAndInside(t *testing.T) {
	o := Object{Shape: CylinderXShape, Bounds: Box{Vec3{-2, -1, -1}, Vec3{2, 1, 1}}}
	for _, c := range []struct {
		name string
		r    Ray
		hit  bool
		d    float64
	}{{"corner miss", Ray{Vec3{0, .9, 3}, Vec3{0, 0, -1}}, true, 3 - math.Sqrt(1-.81)}, {"box corner outside cap", Ray{Vec3{3, .9, .9}, Vec3{-1, 0, 0}}, false, 0}, {"cap", Ray{Vec3{3, 0, 0}, Vec3{-1, 0, 0}}, true, 1}, {"inside", Ray{Vec3{}, Vec3{0, 1, 0}}, true, 1}, {"tangent", Ray{Vec3{0, 1, 3}, Vec3{0, 0, -1}}, true, 3}} {
		t.Run(c.name, func(t *testing.T) {
			d, ok := hitObject(c.r, o)
			if ok != c.hit || ok && math.Abs(d-c.d) > 1e-8 {
				t.Fatalf("got %v %v", d, ok)
			}
		})
	}
}
