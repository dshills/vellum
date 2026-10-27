// Package viewport provides small, renderer-independent 3D navigation and picking
// helpers. Models and scene ownership remain with the application.
package viewport

import (
	"fmt"
	"math"
)

type Vec3 struct{ X, Y, Z float64 }

func (a Vec3) Add(b Vec3) Vec3    { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3    { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Mul(s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Dot(b Vec3) float64 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}
func (a Vec3) Length() float64 { return math.Sqrt(a.Dot(a)) }
func (a Vec3) Unit() Vec3 {
	n := a.Length()
	if n == 0 || !finite(n) {
		return Vec3{}
	}
	return a.Mul(1 / n)
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (a Vec3) valid() bool {
	return finite(a.X) && finite(a.Y) && finite(a.Z) && math.Abs(a.X) <= 1e9 && math.Abs(a.Y) <= 1e9 && math.Abs(a.Z) <= 1e9
}

type View string

const (
	Front     View = "front"
	Back      View = "back"
	Left      View = "left"
	Right     View = "right"
	Top       View = "top"
	Bottom    View = "bottom"
	Isometric View = "isometric"
)

// Camera uses a right-handed Y-up world and radians. Distance controls both
// perspective dolly and orthographic height. Zero Camera requires initialization.
type Camera struct {
	Target               Vec3
	Yaw, Pitch, Distance float64
	Orthographic         bool
}

func DefaultCamera() Camera { return Camera{Yaw: .65, Pitch: .42, Distance: 8} }
func (c Camera) Validate() error {
	if !c.Target.valid() || !finite(c.Yaw) || math.Abs(c.Yaw) > 1e6 || !finite(c.Pitch) || math.Abs(c.Pitch) > math.Pi/2 || !finite(c.Distance) || c.Distance < .05 || c.Distance > 1e6 {
		return fmt.Errorf("viewport: invalid camera")
	}
	return nil
}
func (c Camera) Basis() (eye, right, up, forward Vec3) {
	sy, cy := math.Sincos(c.Yaw)
	sp, cp := math.Sincos(c.Pitch)
	eye = c.Target.Add(Vec3{sy * cp, sp, cy * cp}.Mul(c.Distance))
	forward = c.Target.Sub(eye).Unit()
	// Analytic right remains defined at the top and bottom views.
	right = Vec3{cy, 0, -sy}
	up = right.Cross(forward).Unit()
	return
}
func (c *Camera) Orbit(dx, dy float64) error {
	if c == nil || c.Validate() != nil || !finite(dx) || !finite(dy) {
		return fmt.Errorf("viewport: invalid orbit")
	}
	n := *c
	n.Yaw = math.Remainder(n.Yaw+dx, 2*math.Pi)
	n.Pitch = math.Max(-math.Pi/2+.001, math.Min(math.Pi/2-.001, n.Pitch+dy))
	if n.Validate() != nil {
		return fmt.Errorf("viewport: invalid orbit")
	}
	*c = n
	return nil
}
func (c *Camera) Zoom(factor float64) error {
	if c == nil || c.Validate() != nil || !finite(factor) || factor <= 0 {
		return fmt.Errorf("viewport: invalid zoom")
	}
	n := c.Distance * factor
	if !finite(n) {
		return fmt.Errorf("viewport: invalid zoom")
	}
	c.Distance = math.Max(.05, math.Min(1e6, n))
	return nil
}
func dimensions(width, height float64) bool {
	return finite(width) && finite(height) && width >= 1 && height >= 1 && width <= 1e9 && height <= 1e9
}
func (c *Camera) Pan(dx, dy, width, height float64) error {
	if c == nil || c.Validate() != nil || !finite(dx) || !finite(dy) || !dimensions(width, height) {
		return fmt.Errorf("viewport: invalid pan")
	}
	_, r, u, _ := c.Basis()
	scale := 2 * c.Distance * math.Tan(math.Pi/8) / height
	n := c.Target.Add(r.Mul(-dx * scale)).Add(u.Mul(dy * scale))
	if !n.valid() {
		return fmt.Errorf("viewport: invalid pan")
	}
	c.Target = n
	return nil
}
func (c *Camera) SetView(v View) error {
	if c == nil || c.Validate() != nil {
		return fmt.Errorf("viewport: invalid camera")
	}
	n := *c
	switch v {
	case Front:
		n.Yaw, n.Pitch = 0, 0
	case Back:
		n.Yaw, n.Pitch = math.Pi, 0
	case Left:
		n.Yaw, n.Pitch = -math.Pi/2, 0
	case Right:
		n.Yaw, n.Pitch = math.Pi/2, 0
	case Top:
		n.Yaw, n.Pitch = 0, math.Pi/2
	case Bottom:
		n.Yaw, n.Pitch = 0, -math.Pi/2
	case Isometric:
		n.Yaw, n.Pitch = .65, .42
	default:
		return fmt.Errorf("viewport: unknown view")
	}
	*c = n
	return nil
}

// Matrix is a column-major WebGPU 0..1 depth view-projection matrix. It is
// computed in float64 and converted at the render boundary; no CAD kernel is implied.
func (c Camera) Matrix(width, height float64) ([16]float32, error) {
	var out [16]float32
	if c.Validate() != nil || !dimensions(width, height) {
		return out, fmt.Errorf("viewport: invalid projection")
	}
	eye, r, u, f := c.Basis()
	v := [16]float64{r.X, u.X, -f.X, 0, r.Y, u.Y, -f.Y, 0, r.Z, u.Z, -f.Z, 0, -r.Dot(eye), -u.Dot(eye), f.Dot(eye), 1}
	near, far := .001, math.Max(100., c.Distance*100)
	s := 1 / math.Tan(math.Pi/8)
	aspect := width / height
	p := [16]float64{s / aspect, 0, 0, 0, 0, s, 0, 0, 0, 0, far / (near - far), -1, 0, 0, far * near / (near - far), 0}
	if c.Orthographic {
		h := c.Distance * math.Tan(math.Pi/8)
		p = [16]float64{1 / (h * aspect), 0, 0, 0, 0, 1 / h, 0, 0, 0, 0, 1 / (near - far), 0, 0, 0, near / (near - far), 1}
	}
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			var n float64
			for k := 0; k < 4; k++ {
				n += p[k*4+row] * v[col*4+k]
			}
			out[col*4+row] = float32(n)
		}
	}
	return out, nil
}

type Ray struct{ Origin, Direction Vec3 }

// RayAt takes content-local coordinates and matching content dimensions. Use
// physical pixels for both (or logical units for both), never mix display scales.
func (c Camera) RayAt(x, y, width, height float64) (Ray, error) {
	if c.Validate() != nil || !dimensions(width, height) || !finite(x) || !finite(y) || math.Abs(x) > 1e9 || math.Abs(y) > 1e9 {
		return Ray{}, fmt.Errorf("viewport: invalid ray")
	}
	eye, r, u, f := c.Basis()
	nx, ny := 2*x/width-1, 1-2*y/height
	h := math.Tan(math.Pi / 8)
	offset := r.Mul(nx * h * width / height).Add(u.Mul(ny * h))
	if c.Orthographic {
		return Ray{eye.Add(offset.Mul(c.Distance)), f}, nil
	}
	return Ray{eye, f.Add(offset).Unit()}, nil
}

// Box is an axis-aligned picking proxy; applications supply their own precise
// geometry tests or acceleration structures when needed.
type Box struct{ Min, Max Vec3 }

func (b Box) Validate() error {
	if !b.Min.valid() || !b.Max.valid() || b.Min.X >= b.Max.X || b.Min.Y >= b.Max.Y || b.Min.Z >= b.Max.Z {
		return fmt.Errorf("viewport: invalid box")
	}
	return nil
}
func (r Ray) Hit(b Box) (float64, bool) {
	if !r.Origin.valid() || !r.Direction.valid() || r.Direction.Length() < 1e-12 || b.Validate() != nil {
		return 0, false
	}
	lo, hi := 0., math.Inf(1)
	o := [3]float64{r.Origin.X, r.Origin.Y, r.Origin.Z}
	d := [3]float64{r.Direction.X, r.Direction.Y, r.Direction.Z}
	mn := [3]float64{b.Min.X, b.Min.Y, b.Min.Z}
	mx := [3]float64{b.Max.X, b.Max.Y, b.Max.Z}
	for i := 0; i < 3; i++ {
		if math.Abs(d[i]) < 1e-12 {
			if o[i] < mn[i] || o[i] > mx[i] {
				return 0, false
			}
			continue
		}
		a, z := (mn[i]-o[i])/d[i], (mx[i]-o[i])/d[i]
		if a > z {
			a, z = z, a
		}
		lo = math.Max(lo, a)
		hi = math.Min(hi, z)
		if lo > hi {
			return 0, false
		}
	}
	return lo, true
}
