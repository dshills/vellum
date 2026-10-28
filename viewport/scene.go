package viewport

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// Shape is a bounded preview primitive; all shapes retain box picking proxies.
type Shape uint8

const (
	BoxShape       Shape = iota
	CylinderXShape       // Cylinder along X, inscribed in Bounds; Y/Z extents set radii.
)

// Object is an opaque preview fixture with a box picking proxy. A Scene is a small interactive
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
		if d, ok := ray.Hit(o.Bounds); ok && d < distance {
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
