package demo

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"github.com/dshills/vellum/canvas"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func blend(a, b color.NRGBA, t float64) color.NRGBA {
	t = math.Max(0, math.Min(1, t))
	return rgb(uint8(float64(a.R)*(1-t)+float64(b.R)*t), uint8(float64(a.G)*(1-t)+float64(b.G)*t), uint8(float64(a.B)*(1-t)+float64(b.B)*t))
}
func tiled(p *image.NRGBA) *canvas.TiledImage {
	out, e := canvas.NewTiledImage(p)
	if e != nil {
		panic(e)
	}
	return out
}
func text(p *image.NRGBA, s string, x, y int, size float64, bold bool, c color.NRGBA) {
	data := goregular.TTF
	if bold {
		data = gobold.TTF
	}
	f, e := opentype.Parse(data)
	if e != nil {
		panic(e)
	}
	face, e := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if e != nil {
		panic(e)
	}
	defer face.Close()
	d := font.Drawer{Dst: p, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
	d.DrawString(s)
}

var posterAssets = sync.OnceValue(func() [2]*canvas.TiledImage {
	var out [2]*canvas.TiledImage
	for v := range out {
		p := image.NewNRGBA(image.Rect(0, 0, 800, 560))
		bg := rgb(247, 243, 232)
		if v == 1 {
			bg = rgb(230, 238, 246)
		}
		draw.Draw(p, p.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
		text(p, "FORMA / STUDIES IN SPACE", 40, 48, 13, true, ink)
		text(p, "Make", 38, 166, 78, true, ink)
		text(p, "some", 38, 250, 78, true, ink)
		text(p, "space.", 38, 334, 78, true, ink)
		text(p, "A little room for a new perspective.", 42, 395, 16, false, rgb(89, 93, 99))
		text(p, "VISUAL EXPERIMENT 001", 42, 510, 11, true, ink)
		text(p, "2026", 705, 510, 12, false, ink)
		out[v] = tiled(p)
	}
	return out
})

// Original mathematical landscapes. All pictures are generated locally from
// bounded deterministic functions, with no network, personal files or AI assets.
var landscapeAssets = sync.OnceValue(func() [6][3]*canvas.TiledImage {
	var out [6][3]*canvas.TiledImage
	for k := range out {
		original := landscape(k, 960, 600)
		for look := range 3 {
			p := image.NewNRGBA(original.Bounds())
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					c := original.NRGBAAt(x, y)
					if look == 1 {
						c = blend(c, rgb(234, 158, 82), .17)
					}
					if look == 2 {
						g := uint8(.299*float64(c.R) + .587*float64(c.G) + .114*float64(c.B))
						c = rgb(g, g, g)
					}
					p.SetNRGBA(x, y, c)
				}
			}
			out[k][look] = tiled(p)
		}
	}
	return out
})

func landscape(k, w, h int) *image.NRGBA {
	skies := [6][2]color.NRGBA{{rgb(96, 113, 143), rgb(244, 190, 143)}, {rgb(54, 107, 146), rgb(193, 218, 220)}, {rgb(57, 130, 157), rgb(185, 224, 215)}, {rgb(121, 74, 94), rgb(250, 190, 139)}, {rgb(76, 120, 112), rgb(224, 223, 167)}, {rgb(35, 43, 81), rgb(144, 122, 165)}}
	mountains := [6][3]color.NRGBA{{rgb(176, 119, 103), rgb(190, 114, 78), rgb(118, 69, 63)}, {rgb(145, 173, 183), rgb(81, 129, 146), rgb(38, 76, 105)}, {rgb(121, 186, 181), rgb(58, 142, 153), rgb(35, 94, 119)}, {rgb(194, 111, 98), rgb(161, 74, 72), rgb(95, 47, 65)}, {rgb(150, 167, 113), rgb(96, 135, 102), rgb(42, 89, 81)}, {rgb(124, 120, 159), rgb(76, 78, 127), rgb(41, 46, 87)}}
	p := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			u, v := float64(x)/float64(w), float64(y)/float64(h)
			c := blend(skies[k][0], skies[k][1], v/.8)
			sunX := .70
			sunY := .29
			radius := .075
			if math.Hypot((u-sunX)*float64(w)/float64(h), v-sunY) < radius {
				c = rgb(255, 228, 176)
			}
			for layer := range 3 {
				phase := float64(k)*1.4 + float64(layer)*1.7
				ridge := .51 + float64(layer)*.17 + .065*math.Sin(u*7+phase) + .035*math.Sin(u*15+phase)
				if k == 1 || k == 5 {
					ridge -= .13 * math.Abs(math.Sin(u*10+phase))
				}
				if v > ridge {
					c = blend(mountains[k][layer], blend(mountains[k][layer], rgb(32, 35, 57), .28), (v-ridge)*1.3)
				}
			}
			grain := float64((x*17+y*31+k*13)%13-6) * .25
			for _, ptr := range []*uint8{&c.R, &c.G, &c.B} {
				*ptr = uint8(math.Max(0, math.Min(255, float64(*ptr)+grain)))
			}
			p.SetNRGBA(x, y, c)
		}
	}
	return p
}

type meshBuilder struct {
	v  []canvas.Vertex
	ix []uint32
}

func (b *meshBuilder) quad(x, y, w, h float64, c color.NRGBA) {
	i := uint32(len(b.v))
	b.v = append(b.v, canvas.Vertex{Position: canvas.Point{X: x, Y: y}, Color: c}, canvas.Vertex{Position: canvas.Point{X: x + w, Y: y}, Color: c}, canvas.Vertex{Position: canvas.Point{X: x + w, Y: y + h}, Color: c}, canvas.Vertex{Position: canvas.Point{X: x, Y: y + h}, Color: c})
	b.ix = append(b.ix, i, i+1, i+2, i, i+2, i+3)
}
func (b *meshBuilder) circle(x, y, r float64, c color.NRGBA) {
	i := uint32(len(b.v))
	b.v = append(b.v, canvas.Vertex{Position: canvas.Point{X: x, Y: y}, Color: blend(c, white, .1)})
	for n := 0; n <= 96; n++ {
		a := float64(n) / 96 * 2 * math.Pi
		b.v = append(b.v, canvas.Vertex{Position: canvas.Point{X: x + r*math.Cos(a), Y: y + r*math.Sin(a)}, Color: c})
		if n > 0 {
			b.ix = append(b.ix, i, i+uint32(n), i+uint32(n)+1)
		}
	}
}
func (b *meshBuilder) append(m *canvas.Mesh) {
	i := uint32(len(b.v))
	b.v = append(b.v, m.Vertices()...)
	for _, n := range m.Indices() {
		b.ix = append(b.ix, n+i)
	}
}
func (b *meshBuilder) stroke(p []canvas.Point, width float64, c color.NRGBA) {
	m, e := canvas.StrokePolyline(p, width, c)
	if e != nil {
		panic(e)
	}
	b.append(m)
}
func (b *meshBuilder) mesh() (*canvas.Mesh, error) { return canvas.NewMesh(b.v, b.ix) }
