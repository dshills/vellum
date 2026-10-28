package main

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"golang.org/x/image/vector"
	"image"
	"image/color"
	"math"
	"sync"
)

var iconCache sync.Map

// Tool artwork is original vector geometry rendered with antialiasing at96px.
// Transparent backgrounds keep the native toolbar's own hover/focus treatment;
// each action has its own silhouette rather than a tinted tile or repeated glyph.
type iconPoint struct{ x, y float32 }
type iconPainter struct{ image *image.NRGBA }

func (p iconPainter) path(c color.NRGBA, f func(*vector.Rasterizer)) {
	r := vector.NewRasterizer(96, 96)
	f(r)
	r.Draw(p.image, p.image.Bounds(), &image.Uniform{C: c}, image.Point{})
}
func (p iconPainter) polygon(c color.NRGBA, points ...iconPoint) {
	p.path(c, func(r *vector.Rasterizer) {
		for j, v := range points {
			if j == 0 {
				r.MoveTo(v.x*2, v.y*2)
			} else {
				r.LineTo(v.x*2, v.y*2)
			}
		}
		r.ClosePath()
	})
}
func (p iconPainter) line(c color.NRGBA, width float32, a, b iconPoint) {
	dx, dy := b.x-a.x, b.y-a.y
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length == 0 {
		return
	}
	x, y := -dy/length*width/2, dx/length*width/2
	p.polygon(c, iconPoint{a.x + x, a.y + y}, iconPoint{b.x + x, b.y + y}, iconPoint{b.x - x, b.y - y}, iconPoint{a.x - x, a.y - y})
}
func (p iconPainter) circle(c color.NRGBA, x, y, radius float32) {
	p.path(c, func(r *vector.Rasterizer) {
		const k = float32(.55228475)
		x *= 2
		y *= 2
		radius *= 2
		q := radius * k
		r.MoveTo(x+radius, y)
		r.CubeTo(x+radius, y+q, x+q, y+radius, x, y+radius)
		r.CubeTo(x-q, y+radius, x-radius, y+q, x-radius, y)
		r.CubeTo(x-radius, y-q, x-q, y-radius, x, y-radius)
		r.CubeTo(x+q, y-radius, x+radius, y-q, x+radius, y)
		r.ClosePath()
	})
}
func lighter(c color.NRGBA) color.NRGBA {
	return rgb(uint8((int(c.R)+255)/2), uint8((int(c.G)+255)/2), uint8((int(c.B)+255)/2))
}
func darker(c color.NRGBA) color.NRGBA {
	return rgb(uint8(int(c.R)*3/4), uint8(int(c.G)*3/4), uint8(int(c.B)*3/4))
}
func (p iconPainter) cube(c color.NRGBA, face string) {
	top, left, right := lighter(c), c, darker(c)
	if face != "" {
		top, left, right = rgb(210, 217, 227), rgb(169, 181, 198), rgb(132, 149, 171)
		switch face {
		case "top":
			top = c
		case "front":
			left = c
		case "right":
			right = c
		}
	}
	p.polygon(top, iconPoint{24, 5}, iconPoint{43, 16}, iconPoint{24, 27}, iconPoint{5, 16})
	p.polygon(left, iconPoint{5, 16}, iconPoint{24, 27}, iconPoint{24, 45}, iconPoint{5, 34})
	p.polygon(right, iconPoint{24, 27}, iconPoint{43, 16}, iconPoint{43, 34}, iconPoint{24, 45})
}
func artwork(name string, c color.NRGBA) *canvas.TiledImage {
	key := fmt.Sprint(name, c)
	if v, ok := iconCache.Load(key); ok {
		return v.(*canvas.TiledImage)
	}
	p := iconPainter{image.NewNRGBA(image.Rect(0, 0, 96, 96))}
	ink := rgb(65, 77, 96)
	blue := rgb(54, 124, 225)
	teal := rgb(40, 155, 143)
	switch name {
	case "open":
		p.polygon(rgb(215, 149, 35), iconPoint{4, 12}, iconPoint{19, 12}, iconPoint{23, 17}, iconPoint{44, 17}, iconPoint{44, 40}, iconPoint{4, 40})
		p.polygon(rgb(250, 196, 75), iconPoint{3, 21}, iconPoint{45, 21}, iconPoint{40, 42}, iconPoint{7, 42})
	case "save":
		p.polygon(blue, iconPoint{7, 5}, iconPoint{35, 5}, iconPoint{42, 12}, iconPoint{42, 43}, iconPoint{7, 43})
		p.polygon(rgb(221, 237, 253), iconPoint{15, 5}, iconPoint{32, 5}, iconPoint{32, 19}, iconPoint{15, 19})
		p.polygon(rgb(248, 250, 253), iconPoint{14, 28}, iconPoint{35, 28}, iconPoint{35, 43}, iconPoint{14, 43})
		p.line(ink, 2, iconPoint{27, 8}, iconPoint{27, 15})
	case "add":
		p.cube(c, "")
		p.circle(rgb(255, 255, 255), 36, 35, 11)
		p.circle(rgb(46, 160, 91), 36, 35, 9)
		p.line(rgb(255, 255, 255), 2.5, iconPoint{31, 35}, iconPoint{41, 35})
		p.line(rgb(255, 255, 255), 2.5, iconPoint{36, 30}, iconPoint{36, 40})
	case "delete":
		p.polygon(rgb(219, 85, 87), iconPoint{12, 16}, iconPoint{36, 16}, iconPoint{33, 43}, iconPoint{15, 43})
		p.polygon(rgb(243, 117, 112), iconPoint{9, 10}, iconPoint{39, 10}, iconPoint{39, 14}, iconPoint{9, 14})
		p.polygon(rgb(181, 64, 71), iconPoint{19, 5}, iconPoint{29, 5}, iconPoint{31, 10}, iconPoint{17, 10})
		for _, x := range []float32{20, 28} {
			p.line(rgb(255, 224, 218), 2.5, iconPoint{x, 21}, iconPoint{x, 37})
		}
	case "move":
		p.polygon(teal, iconPoint{24, 3}, iconPoint{32, 12}, iconPoint{27, 12}, iconPoint{27, 21}, iconPoint{36, 21}, iconPoint{36, 16}, iconPoint{45, 24}, iconPoint{36, 32}, iconPoint{36, 27}, iconPoint{27, 27}, iconPoint{27, 36}, iconPoint{32, 36}, iconPoint{24, 45}, iconPoint{16, 36}, iconPoint{21, 36}, iconPoint{21, 27}, iconPoint{12, 27}, iconPoint{12, 32}, iconPoint{3, 24}, iconPoint{12, 16}, iconPoint{12, 21}, iconPoint{21, 21}, iconPoint{21, 12}, iconPoint{16, 12})
	case "grid":
		corners := []iconPoint{{4, 29}, {23, 6}, {45, 19}, {26, 42}}
		p.polygon(rgb(237, 243, 249), corners...)
		for i := 0; i <= 4; i++ {
			f := float32(i) / 4
			p.line(rgb(116, 141, 168), 1.4, iconPoint{4 + 19*f, 29 - 23*f}, iconPoint{26 + 19*f, 42 - 23*f})
			p.line(rgb(116, 141, 168), 1.4, iconPoint{4 + 22*f, 29 + 13*f}, iconPoint{23 + 22*f, 6 + 13*f})
		}
		p.line(rgb(229, 154, 44), 2.3, iconPoint{13.5, 17.5}, iconPoint{35.5, 30.5})
	case "projection":
		p.cube(blue, "top")
		p.line(rgb(244, 154, 62), 2.5, iconPoint{5, 34}, iconPoint{43, 16})
		p.line(rgb(244, 154, 62), 2.5, iconPoint{5, 16}, iconPoint{43, 34})
	case "orbit":
		p.circle(rgb(241, 161, 58), 24, 24, 9)
		p.circle(rgb(255, 205, 116), 21, 21, 4)
		for i := 0; i < 110; i++ {
			a := float64(i) * math.Pi * 1.8 / 110
			b := float64(i+1) * math.Pi * 1.8 / 110
			p.line(blue, 2.8, iconPoint{24 + float32(20*math.Cos(a)), 24 + float32(13*math.Sin(a))}, iconPoint{24 + float32(20*math.Cos(b)), 24 + float32(13*math.Sin(b))})
		}
		p.polygon(blue, iconPoint{42, 11}, iconPoint{43, 23}, iconPoint{33, 17})
	case "pan":
		p.path(teal, func(r *vector.Rasterizer) {
			r.MoveTo(34, 88)
			r.LineTo(27, 70)
			r.LineTo(12, 52)
			r.QuadTo(6, 43, 13, 40)
			r.QuadTo(18, 38, 29, 51)
			r.LineTo(29, 25)
			r.QuadTo(29, 17, 34.5, 17)
			r.QuadTo(40, 17, 40, 25)
			r.LineTo(40, 43)
			r.QuadTo(42, 47, 45, 43)
			r.LineTo(45, 13)
			r.QuadTo(45, 5, 50.5, 5)
			r.QuadTo(56, 5, 56, 13)
			r.LineTo(56, 43)
			r.QuadTo(58, 47, 61, 43)
			r.LineTo(61, 20)
			r.QuadTo(61, 12, 66.5, 12)
			r.QuadTo(72, 12, 72, 20)
			r.LineTo(72, 47)
			r.QuadTo(74, 50, 77, 47)
			r.LineTo(77, 32)
			r.QuadTo(77, 25, 82.5, 25)
			r.QuadTo(88, 25, 88, 32)
			r.LineTo(88, 58)
			r.QuadTo(88, 72, 76, 88)
			r.ClosePath()
		})
	case "select":
		p.polygon(rgb(239, 245, 255), iconPoint{7, 2}, iconPoint{43, 28}, iconPoint{29, 33}, iconPoint{23, 47})
		p.polygon(ink, iconPoint{10, 7}, iconPoint{38, 27}, iconPoint{26, 30}, iconPoint{22, 40})
		p.line(blue, 3, iconPoint{26, 30}, iconPoint{34, 42})
	case "fit":
		p.cube(rgb(77, 137, 221), "")
		for _, q := range [][]iconPoint{{{4, 15}, {4, 4}, {15, 4}}, {{33, 4}, {44, 4}, {44, 15}}, {{44, 33}, {44, 44}, {33, 44}}, {{15, 44}, {4, 44}, {4, 33}}} {
			for j := 1; j < len(q); j++ {
				p.line(ink, 2.5, q[j-1], q[j])
			}
		}
	case "undo", "redo":
		p.path(c, func(r *vector.Rasterizer) {
			x := func(v float32) float32 {
				if name == "redo" {
					return 96 - v
				}
				return v
			}
			r.MoveTo(x(31), 20)
			r.LineTo(x(7), 43)
			r.LineTo(x(31), 66)
			r.LineTo(x(31), 51)
			r.CubeTo(x(57), 43, x(81), 55, x(69), 80)
			r.CubeTo(x(93), 56, x(75), 21, x(31), 35)
			r.ClosePath()
		})
	case "front", "top", "right":
		p.cube(c, name)
	case "wheel":
		p.circle(lighter(c), 24, 24, 8)
		for j := 0; j < 8; j++ {
			a := float64(j) * math.Pi / 4
			p.circle(c, 24+float32(17*math.Cos(a)), 24+float32(17*math.Sin(a)), 3)
		}
	default:
		p.cube(c, "")
	}
	out, e := canvas.NewTiledImage(p.image)
	if e != nil {
		panic(e)
	}
	actual, _ := iconCache.LoadOrStore(key, out)
	return actual.(*canvas.TiledImage)
}
