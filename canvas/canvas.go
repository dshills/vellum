// Package canvas provides toolkit-independent canvas navigation and object picking.
package canvas

import (
	"fmt"
	"image/color"
	"math"
)

type Point struct{ X, Y float64 }
type Rect struct {
	ID                  string
	X, Y, Width, Height float64
	Color               color.RGBA
}
type Viewport struct {
	Width, Height, Zoom float64
	Center              Point
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (v *Viewport) SetSize(width, height float64) error {
	if !finite(width) || !finite(height) || width <= 0 || height <= 0 {
		return fmt.Errorf("canvas: positive finite size required")
	}
	v.Width = width
	v.Height = height
	if v.Zoom == 0 {
		v.Zoom = 1
	}
	return nil
}
func (v Viewport) World(p Point) Point {
	return Point{v.Center.X + (p.X-v.Width/2)/v.Zoom, v.Center.Y + (p.Y-v.Height/2)/v.Zoom}
}
func (v Viewport) Screen(p Point) Point {
	return Point{(p.X-v.Center.X)*v.Zoom + v.Width/2, (p.Y-v.Center.Y)*v.Zoom + v.Height/2}
}
func (v *Viewport) Pan(dx, dy float64) error {
	if !finite(dx) || !finite(dy) || v.Zoom < .01 || v.Zoom > 100 || !finite(v.Zoom) {
		return fmt.Errorf("canvas: invalid pan")
	}
	x, y := v.Center.X-dx/v.Zoom, v.Center.Y-dy/v.Zoom
	if !finite(x) || !finite(y) || math.Abs(x) > 1e12 || math.Abs(y) > 1e12 {
		return fmt.Errorf("canvas: pan exceeds coordinate range")
	}
	v.Center = Point{x, y}
	return nil
}
func (v *Viewport) ZoomAt(p Point, factor float64) error {
	if !finite(p.X) || !finite(p.Y) || !finite(factor) || factor <= 0 || v.Zoom < .01 || v.Zoom > 100 || !finite(v.Zoom) {
		return fmt.Errorf("canvas: invalid zoom")
	}
	before := v.World(p)
	next := v.Zoom * factor
	if next < .01 {
		next = .01
	}
	if next > 100 {
		next = 100
	}
	candidate := *v
	candidate.Zoom = next
	after := candidate.World(p)
	candidate.Center.X += before.X - after.X
	candidate.Center.Y += before.Y - after.Y
	if !finite(candidate.Center.X) || !finite(candidate.Center.Y) || math.Abs(candidate.Center.X) > 1e12 || math.Abs(candidate.Center.Y) > 1e12 {
		return fmt.Errorf("canvas: zoom exceeds coordinate range")
	}
	*v = candidate
	return nil
}

type Scene struct {
	Objects  []Rect
	View     Viewport
	Selected string
}

func New(objects []Rect) (*Scene, error) {
	if len(objects) > 10000 {
		return nil, fmt.Errorf("canvas: 10,000 object limit")
	}
	seen := map[string]bool{}
	for _, o := range objects {
		if o.ID == "" || seen[o.ID] || !finite(o.X) || !finite(o.Y) || !finite(o.Width) || !finite(o.Height) || o.Width <= 0 || o.Height <= 0 || math.Abs(o.X) > 1e6 || math.Abs(o.Y) > 1e6 || o.Width > 1e6 || o.Height > 1e6 {
			return nil, fmt.Errorf("canvas: invalid object %q", o.ID)
		}
		seen[o.ID] = true
	}
	return &Scene{Objects: append([]Rect(nil), objects...), View: Viewport{Width: 800, Height: 600, Zoom: 1}}, nil
}

// ObjectAt uses reverse paint order without changing selection.
func (s *Scene) ObjectAt(screen Point) string {
	p := s.View.World(screen)
	for i := len(s.Objects) - 1; i >= 0; i-- {
		o := s.Objects[i]
		if p.X >= o.X && p.X <= o.X+o.Width && p.Y >= o.Y && p.Y <= o.Y+o.Height {
			return o.ID
		}
	}
	return ""
}

// Pick selects the topmost object at a screen point.
func (s *Scene) Pick(screen Point) string {
	s.Selected = s.ObjectAt(screen)
	return s.Selected
}

// Select selects an exact stable object ID, including an overlapped object. An
// empty ID clears selection; an unknown ID refuses without changing it.
func (s *Scene) Select(id string) error {
	if s == nil {
		return fmt.Errorf("canvas: scene required")
	}
	if id == "" {
		s.Selected = ""
		return nil
	}
	for _, o := range s.Objects {
		if o.ID == id {
			s.Selected = id
			return nil
		}
	}
	return fmt.Errorf("canvas: unknown object %q", id)
}
func (s *Scene) Fit() {
	if len(s.Objects) == 0 {
		s.View.Center = Point{}
		s.View.Zoom = 1
		return
	}
	minX, minY := s.Objects[0].X, s.Objects[0].Y
	maxX, maxY := minX+s.Objects[0].Width, minY+s.Objects[0].Height
	for _, o := range s.Objects {
		minX = math.Min(minX, o.X)
		minY = math.Min(minY, o.Y)
		maxX = math.Max(maxX, o.X+o.Width)
		maxY = math.Max(maxY, o.Y+o.Height)
	}
	s.View.Center = Point{(minX + maxX) / 2, (minY + maxY) / 2}
	s.View.Zoom = math.Min(s.View.Width/(maxX-minX+40), s.View.Height/(maxY-minY+40))
	if s.View.Zoom < .01 {
		s.View.Zoom = .01
	}
	if s.View.Zoom > 100 {
		s.View.Zoom = 100
	}
}

func (s *Scene) Validate() error {
	if s == nil || s.View.Width <= 0 || s.View.Height <= 0 || s.View.Zoom < .01 || s.View.Zoom > 100 || !finite(s.View.Width) || !finite(s.View.Height) || !finite(s.View.Zoom) || !finite(s.View.Center.X) || !finite(s.View.Center.Y) {
		return fmt.Errorf("canvas: invalid viewport")
	}
	if len(s.Objects) > 10000 {
		return fmt.Errorf("canvas: object limit exceeded")
	}
	seen := make(map[string]bool, len(s.Objects))
	for _, o := range s.Objects {
		if o.ID == "" || seen[o.ID] {
			return fmt.Errorf("canvas: unique object IDs required")
		}
		seen[o.ID] = true
		if !finite(o.X) || !finite(o.Y) || !finite(o.Width) || !finite(o.Height) || o.Width <= 0 || o.Height <= 0 || math.Abs(o.X) > 1e6 || math.Abs(o.Y) > 1e6 || o.Width > 1e6 || o.Height > 1e6 {
			return fmt.Errorf("canvas: invalid object geometry")
		}
	}
	return nil
}
