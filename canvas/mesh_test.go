package canvas

import (
	"image/color"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestMeshImmutableAndValidated(t *testing.T) {
	v := []Vertex{{Point{}, color.NRGBA{255, 0, 0, 128}}, {Point{10, 0}, color.NRGBA{}}, {Point{0, 10}, color.NRGBA{}}}
	ix := []uint32{0, 1, 2}
	m, err := NewMesh(v, ix)
	if err != nil {
		t.Fatal(err)
	}
	v[0].Position.X = 42
	ix[0] = 99
	got := m.Vertices()
	got[0].Position.X = 9
	m.Indices()[0] = 90
	if m.Vertices()[0].Position.X != 0 || m.Indices()[0] != 0 {
		t.Fatal("mutable mesh storage escaped")
	}
	for _, bad := range []struct {
		v  []Vertex
		ix []uint32
	}{{v, []uint32{0, 1}}, {v, []uint32{0, 1, 9}}, {[]Vertex{{Position: Point{math.NaN(), 0}}}, nil}, {[]Vertex{{Position: Point{1e7, 0}}}, nil}} {
		if _, err := NewMesh(bad.v, bad.ix); err == nil {
			t.Fatal("accepted invalid geometry", bad)
		}
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				if n, ix := m.Counts(); n != 3 || ix != 3 {
					t.Error("immutable counts")
				}
				m.Vertices()
				m.Indices()
			}
		}()
	}
	wg.Wait()
}
func TestStrokeCapsJoinsAndDuplicates(t *testing.T) {
	c := color.NRGBA{20, 70, 120, 128}
	m, err := StrokePolyline([]Point{{0, 0}, {0, 0}, {10, 0}}, 4, c)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.Vertices(), []Vertex{{Point{0, 2}, c}, {Point{0, -2}, c}, {Point{10, 2}, c}, {Point{10, -2}, c}}) {
		t.Fatal(m.Vertices())
	}
	for _, p := range [][]Point{{{0, 0}}, {{0, 0}, {10, 0}, {0, 0}}, {{0, 0}, {10, 0}, {.001, .001}}, {{0, 0}, {10, 0}, {10, 10}}} {
		m, err := StrokePolyline(p, 4, c)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range m.Vertices() {
			anchor := p[i/2]
			if math.Hypot(v.Position.X-anchor.X, v.Position.Y-anchor.Y) > 8.00001 {
				t.Fatal("miter exceeds bound")
			}
		}
	}
	if _, err := StrokePolyline([]Point{{0, 0}, {1, 1}}, math.Inf(1), c); err == nil {
		t.Fatal("invalid width")
	}
}
func TestViewportWorldBounds(t *testing.T) {
	b, err := (Viewport{Width: 100, Height: 80, Zoom: 2, Center: Point{70, 10}}).Bounds()
	if err != nil || b != (Bounds{Point{45, -10}, Point{95, 30}}) {
		t.Fatal(b, err)
	}
	if _, err := (Viewport{Width: 100, Height: 80, Zoom: 1, Center: Point{1e15, 0}}).Bounds(); err == nil {
		t.Fatal("unbounded viewport")
	}
}

func TestViewportBoundsRejectsComputedOverflowAndCollapse(t *testing.T) {
	for _, view := range []Viewport{{Width: math.MaxFloat64, Height: 10, Zoom: .01}, {Width: math.SmallestNonzeroFloat64, Height: 10, Zoom: 1, Center: Point{1e12, 0}}} {
		if _, err := view.Bounds(); err == nil {
			t.Fatal("accepted invalid computed bounds", view)
		}
	}
}
