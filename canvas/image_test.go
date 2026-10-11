package canvas

import (
	"image"
	"image/color"
	"math"
	"sync"
	"testing"
)

func TestTiledImageOriginStrideAndImmutablePixels(t *testing.T) {
	source := image.NewNRGBA(image.Rect(-20, 30, 300, 340))
	source.SetNRGBA(17, 91, color.NRGBA{220, 17, 99, 0})
	sub := source.SubImage(image.Rect(17, 91, 276, 222)).(*image.NRGBA)
	m, err := NewTiledImage(sub)
	if err != nil {
		t.Fatal(err)
	}
	if m.Size() != image.Pt(259, 131) || m.TileCount() != 6 || m.NRGBAAt(0, 0) != (color.NRGBA{220, 17, 99, 0}) {
		t.Fatal(m.Size(), m.TileCount(), m.NRGBAAt(0, 0))
	}
	source.SetNRGBA(17, 91, color.NRGBA{})
	if m.NRGBAAt(0, 0) != (color.NRGBA{220, 17, 99, 0}) {
		t.Fatal("source mutation escaped")
	}
	tiles := m.ChangedTiles(nil)
	if tiles[5].Bounds() != image.Rect(256, 128, 259, 131) || len(tiles[5].Pixels()) != 36 {
		t.Fatal("edge tile", tiles[5].Bounds())
	}
	pixels := tiles[0].Pixels()
	pixels[0] = 0
	if m.NRGBAAt(0, 0).R != 220 {
		t.Fatal("tile accessor mutation escaped")
	}
	if m.At(-1, 0) != (color.NRGBA{}) || len(m.ChangedTiles(m)) != 0 {
		t.Fatal("bounds/identity")
	}
}
func TestImagePatchCrossesTilesAndKeepsOtherIdentities(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 300, 300))
	old, _ := NewTiledImage(source)
	patch := image.NewNRGBA(image.Rect(-4, 8, -2, 10))
	for y := 8; y < 10; y++ {
		for x := -4; x < -2; x++ {
			patch.SetNRGBA(x, y, color.NRGBA{uint8(x + 5), uint8(y), 30, 128})
		}
	}
	changed, err := old.WithPatch(image.Rect(127, 127, 129, 129), patch)
	if err != nil {
		t.Fatal(err)
	}
	if len(changed.ChangedTiles(old)) != 4 {
		t.Fatal("expected four changed tiles")
	}
	for y := 0; y < 300; y++ {
		for x := 0; x < 300; x++ {
			want := color.NRGBA{}
			if x >= 127 && x < 129 && y >= 127 && y < 129 {
				want = patch.NRGBAAt(x-131, y-119)
			}
			if changed.NRGBAAt(x, y) != want || old.NRGBAAt(x, y) != (color.NRGBA{}) {
				t.Fatalf("pixel %d,%d", x, y)
			}
		}
	}
	same, err := changed.WithPatch(image.Rect(127, 127, 129, 129), patch)
	if err != nil || same != changed {
		t.Fatal("equal pixels changed version identity")
	}
	patch.Pix[0] = 99
	if changed.NRGBAAt(127, 127).R != 1 {
		t.Fatal("patch mutation escaped")
	}
	for _, index := range []int{2, 5, 6, 7, 8} {
		if changed.tiles[index] != old.tiles[index] {
			t.Fatal("unaffected tile copied", index)
		}
	}
	independent, _ := NewTiledImage(source)
	if len(changed.ChangedTiles(independent)) != changed.TileCount() {
		t.Fatal("unrelated identities reused")
	}
}
func TestImageRejectsMalformedLayoutAndPatch(t *testing.T) {
	for _, source := range []*image.NRGBA{nil, {Rect: image.Rectangle{Min: image.Pt(math.MaxInt, 0), Max: image.Pt(math.MinInt, 1)}, Stride: 4, Pix: make([]byte, 4)}, {Rect: image.Rect(0, 0, 8193, 1)}, {Rect: image.Rect(0, 0, 0, 1)}, {Rect: image.Rect(0, 0, 2, 2), Stride: 7, Pix: make([]byte, 16)}, {Rect: image.Rect(0, 0, 2, 2), Stride: 8, Pix: make([]byte, 15)}, {Rect: image.Rect(0, 0, 1, 2), Stride: math.MaxInt, Pix: make([]byte, 8)}} {
		if _, err := NewTiledImage(source); err == nil {
			t.Fatal("malformed source admitted", source)
		}
	}
	m, _ := NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 130, 130)))
	for _, r := range []image.Rectangle{image.Rect(-1, 0, 1, 2), image.Rect(129, 129, 131, 131), image.Rect(0, 0, 0, 2)} {
		if _, err := m.WithPatch(r, image.NewNRGBA(image.Rect(0, 0, 2, 2))); err == nil {
			t.Fatal("invalid patch admitted", r)
		}
	}
	if _, err := m.WithPatch(image.Rect(0, 0, 2, 2), image.NewNRGBA(image.Rect(0, 0, 3, 2))); err == nil {
		t.Fatal("mismatched patch")
	}
	valid := ImageLayer{Image: m, Bounds: Bounds{Max: Point{130, 130}}, Filter: ImageLinear}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, layer := range []ImageLayer{{}, {Image: m, Bounds: valid.Bounds, Filter: 9}, {Image: m, Bounds: Bounds{Max: Point{1e6 + 1, 10}}}, {Image: m, Bounds: Bounds{Max: Point{math.Inf(1), 10}}}} {
		if layer.Validate() == nil {
			t.Fatal("invalid layer admitted")
		}
	}
}
func TestImageConcurrentIndependentVersions(t *testing.T) {
	old, _ := NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 300, 300)))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			patch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
			patch.Pix[0] = uint8(i + 1)
			for n := 0; n < 100; n++ {
				v, err := old.WithPatch(image.Rect(i, i, i+1, i+1), patch)
				if err != nil || v.NRGBAAt(i, i).R != uint8(i+1) || old.NRGBAAt(i, i).R != 0 {
					t.Error("concurrent immutable version")
				}
			}
		}(i)
	}
	wg.Wait()
}
func BenchmarkImage4096SingleTilePatch(b *testing.B) {
	m, _ := NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 4096, 4096)))
	patch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch.Pix[0] = uint8(i + 1)
		var err error
		m, err = m.WithPatch(image.Rect(2048, 2048, 2049, 2049), patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}
