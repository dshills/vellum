package canvas

import (
	"fmt"
	"image/color"
	"math/rand"
	"reflect"
	"testing"
)

func indexedFixture(n int) []Rect {
	objects := make([]Rect, n)
	for i := range objects {
		objects[i] = Rect{ID: fmt.Sprint(i), X: float64(i%400) * 20, Y: float64(i/400) * 20, Width: 10, Height: 10, Color: color.RGBA{67, 197, 166, 255}}
	}
	return objects
}
func TestHundredThousandRectCulling(t *testing.T) {
	objects := indexedFixture(100_000)
	r, err := NewRectIndex(objects)
	if err != nil {
		t.Fatal(err)
	}
	objects[0].X = 999999
	mesh, stats, err := r.Mesh(Viewport{Width: 400, Height: 300, Zoom: 1, Center: Point{200, 150}})
	if err != nil {
		t.Fatal(err)
	}
	v, ix := mesh.Counts()
	if stats.TotalObjects != 100000 || stats.VisibleObjects != 336 || stats.TestedObjects > 1000 || v != stats.VisibleObjects*4 || ix != stats.VisibleObjects*6 {
		t.Fatal(stats, v, ix)
	}
}
func TestIndexConservativeAndPaintOrdered(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	objects := make([]Rect, 3000)
	for i := range objects {
		objects[i] = Rect{ID: fmt.Sprint(i), X: rng.Float64()*2000 - 1000, Y: rng.Float64()*2000 - 1000, Width: 1 + rng.Float64()*1000, Height: 1 + rng.Float64()*1000}
	}
	objects[0] = Rect{ID: "spanning", X: -1000, Y: -1000, Width: 3000, Height: 3000}
	r, _ := NewRectIndex(objects)
	for range 30 {
		b := Bounds{Min: Point{rng.Float64()*1000 - 500, rng.Float64()*1000 - 500}}
		b.Max = Point{b.Min.X + 30, b.Min.Y + 60}
		want := []Rect{}
		for _, o := range objects {
			if intersects(rectBounds(o), b) {
				want = append(want, o)
			}
		}
		got, _, err := r.Query(b)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("query differs from paint-ordered brute force", err, len(got), len(want))
		}
	}
	got, _, _ := r.Query(Bounds{Point{5000, 5000}, Point{5100, 5100}})
	if len(got) != 0 {
		t.Fatal("offscreen objects")
	}
	if _, err := NewRectIndex([]Rect{{ID: "a", Width: 1, Height: 1}, {ID: "a", Width: 1, Height: 1}}); err == nil {
		t.Fatal("duplicate identity")
	}
}
func TestIndexOverlappingRefusesVisibleMeshOverflow(t *testing.T) {
	objects := make([]Rect, MaxMeshVertices/4+1)
	for i := range objects {
		objects[i] = Rect{ID: fmt.Sprint(i), Width: 1, Height: 1}
	}
	r, err := NewRectIndex(objects)
	if err != nil {
		t.Fatal(err)
	}
	_, stats, err := r.Mesh(Viewport{Width: 10, Height: 10, Zoom: 1})
	if err == nil || stats.VisibleObjects != len(objects) {
		t.Fatal("visible budget not enforced", err, stats)
	}
}
func BenchmarkIndexed100kVisibleMesh(b *testing.B) {
	r, _ := NewRectIndex(indexedFixture(100_000))
	view := Viewport{Width: 800, Height: 600, Zoom: 1, Center: Point{400, 300}}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.Mesh(view)
	}
}

func TestIndexedPremultipliedColorConversion(t *testing.T) {
	for alpha := 0; alpha <= 255; alpha++ {
		for red := 0; red <= alpha; red++ {
			c := color.RGBA{R: uint8(red), G: uint8(alpha - red), B: uint8(red / 2), A: uint8(alpha)}
			if got, want := straightRectColor(c), color.NRGBAModel.Convert(c).(color.NRGBA); got != want {
				t.Fatal(c, got, want)
			}
		}
	}
	if _, err := NewRectIndex([]Rect{{ID: "bad", Width: 1, Height: 1, Color: color.RGBA{R: 200, A: 10}}}); err == nil {
		t.Fatal("invalid premultiplied color accepted")
	}
}
