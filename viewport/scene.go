package viewport

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// Shape is a bounded preview primitive; picking follows the primitive silhouette.
type Shape uint8

const (
	BoxShape       Shape = iota
	CylinderXShape       // Cylinder along X, inscribed in Bounds; Y/Z extents set radii.
)

// Object is an opaque preview fixture with primitive picking. A Scene is a small interactive
// preview, not a boundary-representation model or an assembly acceleration index.
type Object struct {
	ID, Name string
	Bounds   Box
	Color    color.NRGBA
	Visible  bool
	Shape    Shape
}
type Scene struct {
	Camera                Camera
	Objects               []Object
	SelectedID, HoveredID string
	Grid, ShowCube        bool
	Revision              uint64
}

func (s *Scene) Validate() error {
	if s == nil || s.Camera.Validate() != nil || len(s.Objects) > 256 {
		return fmt.Errorf("viewport: invalid preview scene")
	}
	seen := map[string]bool{}
	for _, o := range s.Objects {
		if o.ID == "" || o.Name == "" || len(o.ID) > 256 || len(o.Name) > 1024 || seen[o.ID] || o.Bounds.Validate() != nil || o.Color.A != 255 || o.Shape > CylinderXShape {
			return fmt.Errorf("viewport: invalid preview object")
		}
		seen[o.ID] = true
	}
	if s.SelectedID != "" && !seen[s.SelectedID] || s.HoveredID != "" && !seen[s.HoveredID] {
		return fmt.Errorf("viewport: missing selected/hovered object")
	}
	return nil
}
func (s *Scene) Touch() error {
	if s == nil || s.Revision == ^uint64(0) {
		return fmt.Errorf("viewport: exhausted revision")
	}
	s.Revision++
	return nil
}
func (s *Scene) Pick(x, y, width, height float64) (string, error) {
	if e := s.Validate(); e != nil {
		return "", e
	}
	ray, e := s.Camera.RayAt(x, y, width, height)
	if e != nil {
		return "", e
	}
	distance := math.Inf(1)
	id := ""
	for _, o := range s.Objects {
		if !o.Visible {
			continue
		}
		if d, ok := hitObject(ray, o); ok && d < distance {
			distance, id = d, o.ID
		}
	}
	return id, nil
}

// CubeRect reserves an in-canvas orientation control in full-content physical
// pixels. Small views omit it. Scale must match the renderer's dp scale.
func CubeRect(width, height int, scale float64) image.Rectangle {
	if width < 1 || height < 1 || !finite(scale) || scale <= 0 || scale > 16 {
		return image.Rectangle{}
	}
	size, pad := int(math.Round(88*scale)), int(math.Round(14*scale))
	if width < size+pad*3 || height < size+pad*3 {
		return image.Rectangle{}
	}
	return image.Rect(width-pad-size, pad, width-pad, pad+size)
}
func CubeCamera(c Camera) Camera { c.Target = Vec3{}; c.Distance = 4; c.Orthographic = false; return c }

// CubeViewAt resolves the nearest visible face of the orientation cube. The
// same camera/rectangle is used by the native renderer and CPU pointer picker.
func CubeViewAt(c Camera, x, y, width, height int, scale float64) (View, bool) {
	rect := CubeRect(width, height, scale)
	if rect.Empty() || !image.Pt(x, y).In(rect) {
		return "", false
	}
	ray, e := CubeCamera(c).RayAt(float64(x-rect.Min.X), float64(y-rect.Min.Y), float64(rect.Dx()), float64(rect.Dy()))
	if e != nil {
		return "", false
	}
	d, ok := ray.Hit(Box{Vec3{-1, -1, -1}, Vec3{1, 1, 1}})
	if !ok {
		return "", false
	}
	p := ray.Origin.Add(ray.Direction.Mul(d))
	if math.Abs(p.Y) >= math.Abs(p.X) && math.Abs(p.Y) >= math.Abs(p.Z) {
		if p.Y > 0 {
			return Top, true
		}
		return Bottom, true
	}
	if math.Abs(p.X) >= math.Abs(p.Z) {
		if p.X > 0 {
			return Right, true
		}
		return Left, true
	}
	if p.Z > 0 {
		return Front, true
	}
	return Back, true
}

// hitObject tests the actual preview primitive, including capped elliptical
// cylinders. Distances retain the caller's ray parameter for nearest-hit order.
func hitObject(r Ray, o Object) (float64, bool) {
	if o.Shape != CylinderXShape {
		return r.Hit(o.Bounds)
	}
	if _, ok := r.Hit(o.Bounds); !ok {
		return 0, false
	}
	b := o.Bounds
	cy := (b.Min.Y + b.Max.Y) / 2
	cz := (b.Min.Z + b.Max.Z) / 2
	ry := (b.Max.Y - b.Min.Y) / 2
	rz := (b.Max.Z - b.Min.Z) / 2
	y, z := (r.Origin.Y-cy)/ry, (r.Origin.Z-cz)/rz
	dy, dz := r.Direction.Y/ry, r.Direction.Z/rz
	a := dy*dy + dz*dz
	bb := 2 * (y*dy + z*dz)
	c := y*y + z*z - 1
	best := math.Inf(1)
	accept := func(t float64) {
		x := r.Origin.X + t*r.Direction.X
		if t >= 0 && x >= b.Min.X-1e-10 && x <= b.Max.X+1e-10 && t < best {
			best = t
		}
	}
	if a > 1e-24 {
		disc := bb*bb - 4*a*c
		if disc >= 0 {
			root := math.Sqrt(disc)
			accept((-bb - root) / (2 * a))
			accept((-bb + root) / (2 * a))
		}
	}
	if math.Abs(r.Direction.X) > 1e-12 {
		for _, x := range []float64{b.Min.X, b.Max.X} {
			t := (x - r.Origin.X) / r.Direction.X
			yy, zz := y+t*dy, z+t*dz
			if yy*yy+zz*zz <= 1+1e-12 {
				accept(t)
			}
		}
	}
	return best, !math.IsInf(best, 1)
}
