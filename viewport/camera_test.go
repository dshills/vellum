package viewport

import (
	"math"
	"testing"
)

func TestCameraRaysAndProjection(t *testing.T) {
	for _, ortho := range []bool{false, true} {
		c := DefaultCamera()
		c.Orthographic = ortho
		_ = c.SetView(Front)
		for _, scale := range []float64{1, 1.5, 2} {
			r, e := c.RayAt(400*scale, 300*scale, 800*scale, 600*scale)
			if e != nil {
				t.Fatal(e)
			}
			hit, ok := r.Hit(Box{Vec3{-1, -1, -1}, Vec3{1, 1, 1}})
			if !ok || math.Abs(hit-7) > 1e-8 {
				t.Fatal(ortho, scale, r, hit)
			}
			m, e := c.Matrix(800*scale, 600*scale)
			if e != nil {
				t.Fatal(e)
			}
			w := m[15]
			if !ortho {
				w = m[15]
			}
			if w <= 0 {
				t.Fatal("origin behind camera", m)
			}
		}
		r, _ := c.RayAt(0, 0, 800, 600)
		if _, ok := r.Hit(Box{Vec3{-1, -1, -1}, Vec3{1, 1, 1}}); ok {
			t.Fatal("corner picked center box")
		}
	}
}
func TestCameraNavigationAndInvalidAtomicity(t *testing.T) {
	c := DefaultCamera()
	for _, v := range []View{Top, Bottom, Front, Back, Left, Right, Isometric} {
		if c.SetView(v) != nil {
			t.Fatal(v)
		}
		e, r, u, f := c.Basis()
		if !e.valid() || math.Abs(r.Dot(u)) > 1e-8 || math.Abs(r.Dot(f)) > 1e-8 || math.Abs(u.Dot(f)) > 1e-8 {
			t.Fatal("basis", v)
		}
	}
	before := c
	if c.Orbit(math.NaN(), 0) == nil || c != before {
		t.Fatal("invalid orbit changed camera")
	}
	if c.Zoom(-1) == nil || c != before {
		t.Fatal("invalid zoom")
	}
	if c.Pan(1, 2, 0, 300) == nil || c != before {
		t.Fatal("invalid pan")
	}
	if c.SetView("unknown") == nil || c != before {
		t.Fatal("invalid preset")
	}
	if c.Orbit(.1, .2) != nil || c == before {
		t.Fatal("orbit")
	}
	old := c.Target
	if c.Pan(40, 0, 800, 600) != nil || c.Target == old {
		t.Fatal("pan")
	}
	d := c.Distance
	if c.Zoom(.5) != nil || c.Distance != d/2 {
		t.Fatal("zoom")
	}
}
func TestBoxPickingDepthParallelAndInside(t *testing.T) {
	b := Box{Vec3{-1, -1, -1}, Vec3{1, 1, 1}}
	for _, r := range []Ray{{Vec3{2, 0, 5}, Vec3{0, 0, -1}}, {Vec3{0, 0, -5}, Vec3{0, 0, -1}}, {Vec3{}, Vec3{}}} {
		if _, ok := r.Hit(b); ok {
			t.Fatal("invalid/missing ray", r)
		}
	}
	if d, ok := (Ray{Vec3{}, Vec3{1, 0, 0}}).Hit(b); !ok || d != 0 {
		t.Fatal("inside", d, ok)
	}
	r := Ray{Vec3{0, 0, 8}, Vec3{0, 0, -1}}
	near, _ := r.Hit(b)
	far, _ := r.Hit(Box{Vec3{-1, -1, -4}, Vec3{1, 1, -2}})
	if near >= far {
		t.Fatal("depth ordering")
	}
}
