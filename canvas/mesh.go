package canvas

import (
	"fmt"
	"image/color"
	"math"
	"slices"
)

// Vertex supplies a world position and straight-alpha sRGB color. The renderer
// converts colors to premultiplied linear RGB before interpolation and blending.
type Vertex struct {
	Position Point
	Color    color.NRGBA
}

// Mesh is immutable indexed triangle geometry. Index order is paint order;
// neither winding imposes back-face culling. Constructors and accessors copy.
// A Mesh can be shared between applications and worker/UI goroutines.
type Mesh struct {
	vertices []Vertex
	indices  []uint32
}

const MaxMeshVertices = 1_000_000
const MaxMeshIndices = 3_000_000

func NewMesh(vertices []Vertex, indices []uint32) (*Mesh, error) {
	if len(vertices) > MaxMeshVertices || len(indices) > MaxMeshIndices || len(indices)%3 != 0 {
		return nil, fmt.Errorf("canvas: triangle mesh limits exceeded or incomplete triangle")
	}
	for _, v := range vertices {
		if !finite(v.Position.X) || !finite(v.Position.Y) || math.Abs(v.Position.X) > 1e6 || math.Abs(v.Position.Y) > 1e6 {
			return nil, fmt.Errorf("canvas: mesh position outside finite coordinate range")
		}
	}
	for _, i := range indices {
		if uint64(i) >= uint64(len(vertices)) {
			return nil, fmt.Errorf("canvas: mesh index outside vertices")
		}
	}
	return &Mesh{slices.Clone(vertices), slices.Clone(indices)}, nil
}
func (m *Mesh) Vertices() []Vertex {
	if m == nil {
		return nil
	}
	return slices.Clone(m.vertices)
}
func (m *Mesh) Indices() []uint32 {
	if m == nil {
		return nil
	}
	return slices.Clone(m.indices)
}
func (m *Mesh) Counts() (vertices, indices int) {
	if m == nil {
		return
	}
	return len(m.vertices), len(m.indices)
}

// StrokePolyline tessellates an open stroke with butt caps and miter joins
// capped at four half-widths. Consecutive duplicate points are removed. Self
// intersections and reversals can overlap; this is not a polygon fill operation.
func StrokePolyline(points []Point, width float64, c color.NRGBA) (*Mesh, error) {
	if !finite(width) || width <= 0 || width > 1e6 || len(points) > MaxMeshVertices/2 {
		return nil, fmt.Errorf("canvas: invalid stroke size")
	}
	p := make([]Point, 0, len(points))
	for _, pt := range points {
		if !finite(pt.X) || !finite(pt.Y) || math.Abs(pt.X) > 1e6 || math.Abs(pt.Y) > 1e6 {
			return nil, fmt.Errorf("canvas: invalid stroke point")
		}
		if len(p) == 0 || pt != p[len(p)-1] {
			p = append(p, pt)
		}
	}
	if len(p) < 2 {
		return NewMesh(nil, nil)
	}
	normals := make([]Point, len(p)-1)
	for i := range normals {
		dx, dy := p[i+1].X-p[i].X, p[i+1].Y-p[i].Y
		length := math.Hypot(dx, dy)
		normals[i] = Point{-dy / length, dx / length}
	}
	v := make([]Vertex, 0, 2*len(p))
	ix := make([]uint32, 0, 6*(len(p)-1))
	for i, pt := range p {
		n := normals[min(i, len(normals)-1)]
		scale := width / 2
		if i > 0 && i < len(p)-1 {
			a := normals[i-1]
			x, y := a.X+n.X, a.Y+n.Y
			length := math.Hypot(x, y)
			if length > 1e-12 {
				n = Point{x / length, y / length}
				dot := n.X*a.X + n.Y*a.Y
				scale = min(scale/max(dot, .25), width*2)
			} else {
				n = a
			}
		}
		v = append(v, Vertex{Point{pt.X + n.X*scale, pt.Y + n.Y*scale}, c}, Vertex{Point{pt.X - n.X*scale, pt.Y - n.Y*scale}, c})
		if i > 0 {
			j := uint32(2 * (i - 1))
			ix = append(ix, j, j+1, j+2, j+1, j+3, j+2)
		}
	}
	return NewMesh(v, ix)
}

// Bounds is a world-space axis-aligned rectangle.
type Bounds struct{ Min, Max Point }

func (b Bounds) Validate() error {
	if !finite(b.Min.X) || !finite(b.Min.Y) || !finite(b.Max.X) || !finite(b.Max.Y) || b.Max.X <= b.Min.X || b.Max.Y <= b.Min.Y {
		return fmt.Errorf("canvas: invalid bounds")
	}
	return nil
}

// Bounds returns the full viewport in world coordinates.
func (v Viewport) Bounds() (Bounds, error) {
	if !finite(v.Width) || !finite(v.Height) || v.Width <= 0 || v.Height <= 0 || !finite(v.Zoom) || v.Zoom < .01 || v.Zoom > 100 || !finite(v.Center.X) || !finite(v.Center.Y) || math.Abs(v.Center.X) > 1e12 || math.Abs(v.Center.Y) > 1e12 {
		return Bounds{}, fmt.Errorf("canvas: invalid viewport")
	}
	b := Bounds{v.World(Point{}), v.World(Point{v.Width, v.Height})}
	if err := b.Validate(); err != nil {
		return Bounds{}, err
	}
	return b, nil
}

func appendQuad(v []Vertex, ix []uint32, b Bounds, c color.NRGBA) ([]Vertex, []uint32) {
	j := uint32(len(v))
	v = append(v, Vertex{b.Min, c}, Vertex{Point{b.Max.X, b.Min.Y}, c}, Vertex{b.Max, c}, Vertex{Point{b.Min.X, b.Max.Y}, c})
	return v, append(ix, j, j+1, j+2, j, j+2, j+3)
}
