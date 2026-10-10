package canvas

import (
	"fmt"
	"image/color"
	"math"
	"sort"
)

// RectIndex is an immutable bounding-volume hierarchy. Query output preserves
// source paint order; offscreen objects are culled before constructing a Mesh.
// The index may be constructed on a worker and shared across readers.
type RectIndex struct {
	objects []Rect
	order   []int
	nodes   []rectNode
}
type rectNode struct {
	bounds                  Bounds
	start, end, left, right int
}
type QueryStats struct{ TotalObjects, VisitedNodes, TestedObjects, VisibleObjects int }

const MaxIndexedRects = 1_000_000

func NewRectIndex(objects []Rect) (*RectIndex, error) {
	if len(objects) > MaxIndexedRects {
		return nil, fmt.Errorf("canvas: indexed object limit exceeded")
	}
	seen := make(map[string]bool, len(objects))
	for _, o := range objects {
		if o.Color.R > o.Color.A || o.Color.G > o.Color.A || o.Color.B > o.Color.A {
			return nil, fmt.Errorf("canvas: rectangle color must be premultiplied RGBA")
		}
		if o.ID == "" || seen[o.ID] || !finite(o.X) || !finite(o.Y) || !finite(o.Width) || !finite(o.Height) || o.Width <= 0 || o.Height <= 0 || math.Abs(o.X) > 1e6 || math.Abs(o.Y) > 1e6 || o.Width > 1e6 || o.Height > 1e6 || math.Abs(o.X+o.Width) > 1e6 || math.Abs(o.Y+o.Height) > 1e6 {
			return nil, fmt.Errorf("canvas: invalid indexed rectangle %q", o.ID)
		}
		seen[o.ID] = true
	}
	r := &RectIndex{objects: append([]Rect(nil), objects...), order: make([]int, len(objects))}
	for i := range r.order {
		r.order[i] = i
	}
	if len(objects) > 0 {
		r.build(0, len(objects))
	}
	return r, nil
}
func rectBounds(o Rect) Bounds { return Bounds{Point{o.X, o.Y}, Point{o.X + o.Width, o.Y + o.Height}} }
func intersects(a, b Bounds) bool {
	return a.Min.X <= b.Max.X && a.Max.X >= b.Min.X && a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y
}
func (r *RectIndex) build(start, end int) int {
	b := rectBounds(r.objects[r.order[start]])
	for _, i := range r.order[start+1 : end] {
		o := rectBounds(r.objects[i])
		b.Min.X = min(b.Min.X, o.Min.X)
		b.Min.Y = min(b.Min.Y, o.Min.Y)
		b.Max.X = max(b.Max.X, o.Max.X)
		b.Max.Y = max(b.Max.Y, o.Max.Y)
	}
	n := len(r.nodes)
	r.nodes = append(r.nodes, rectNode{b, start, end, -1, -1})
	if end-start > 16 {
		byX := b.Max.X-b.Min.X >= b.Max.Y-b.Min.Y
		sort.Slice(r.order[start:end], func(i, j int) bool {
			a, b := r.objects[r.order[start+i]], r.objects[r.order[start+j]]
			x, y := a.X+a.Width*.5, b.X+b.Width*.5
			if !byX {
				x, y = a.Y+a.Height*.5, b.Y+b.Height*.5
			}
			if x == y {
				return r.order[start+i] < r.order[start+j]
			}
			return x < y
		})
		mid := start + (end-start)/2
		left := r.build(start, mid)
		right := r.build(mid, end)
		r.nodes[n].left = left
		r.nodes[n].right = right
	}
	return n
}
func (r *RectIndex) Query(b Bounds) ([]Rect, QueryStats, error) {
	if r == nil {
		return nil, QueryStats{}, fmt.Errorf("canvas: rectangle index required")
	}
	if err := b.Validate(); err != nil {
		return nil, QueryStats{}, err
	}
	stats := QueryStats{TotalObjects: len(r.objects)}
	found := make([]int, 0)
	var visit func(int)
	visit = func(n int) {
		stats.VisitedNodes++
		node := r.nodes[n]
		if !intersects(node.bounds, b) {
			return
		}
		if node.left >= 0 {
			visit(node.left)
			visit(node.right)
			return
		}
		for _, i := range r.order[node.start:node.end] {
			stats.TestedObjects++
			if intersects(rectBounds(r.objects[i]), b) {
				found = append(found, i)
			}
		}
	}
	if len(r.nodes) > 0 {
		visit(0)
	}
	sort.Ints(found)
	out := make([]Rect, len(found))
	for i, j := range found {
		out[i] = r.objects[j]
	}
	stats.VisibleObjects = len(out)
	return out, stats, nil
}
func (r *RectIndex) Mesh(view Viewport) (*Mesh, QueryStats, error) {
	b, err := view.Bounds()
	if err != nil {
		return nil, QueryStats{}, err
	}
	objects, stats, err := r.Query(b)
	if err != nil {
		return nil, stats, err
	}
	if len(objects) > MaxMeshVertices/4 || len(objects) > MaxMeshIndices/6 {
		return nil, stats, fmt.Errorf("canvas: visible rectangle mesh exceeds limits")
	}
	v := make([]Vertex, 0, 4*len(objects))
	ix := make([]uint32, 0, 6*len(objects))
	for _, o := range objects {
		v, ix = appendQuad(v, ix, rectBounds(o), straightRectColor(o.Color))
	}
	m, err := NewMesh(v, ix)
	return m, stats, err
}

func straightRectColor(c color.RGBA) color.NRGBA {
	if c.A == 0 {
		return color.NRGBA{}
	}
	if c.A == 255 {
		return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
	}
	alpha := uint32(c.A)
	return color.NRGBA{R: uint8((uint32(c.R) * 65535 / alpha) >> 8), G: uint8((uint32(c.G) * 65535 / alpha) >> 8), B: uint8((uint32(c.B) * 65535 / alpha) >> 8), A: c.A}
}
