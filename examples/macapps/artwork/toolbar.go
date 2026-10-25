package artwork

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

// Original Vellum sample artwork, generated once at 96px with 4x sampling.
// No OS icon extraction, network resources or third-party image licensing.
var toolbarColorIcons = sync.OnceValues(func() (map[ui.Symbol]*canvas.TiledImage, error) {
	result := make(map[ui.Symbol]*canvas.TiledImage, 7)
	for _, symbol := range []ui.Symbol{ui.SymbolAdd, ui.SymbolFolder, ui.SymbolSave, ui.SymbolUndo, ui.SymbolRedo, ui.SymbolDelete, ui.SymbolDocument} {
		pixels := image.NewNRGBA(image.Rect(0, 0, 96, 96))
		for y := range 96 {
			for x := range 96 {
				var red, green, blue, alpha float64
				for sy := range 4 {
					for sx := range 4 {
						c := toolbarArtworkPixel(symbol, float64(x)+(float64(sx)+0.5)/4, float64(y)+(float64(sy)+0.5)/4)
						a := float64(c.A) / 255
						red += float64(c.R) * a
						green += float64(c.G) * a
						blue += float64(c.B) * a
						alpha += a
					}
				}
				if alpha > 0 {
					pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(math.Round(red / alpha)), G: uint8(math.Round(green / alpha)), B: uint8(math.Round(blue / alpha)), A: uint8(math.Round(alpha / 16 * 255))})
				}
			}
		}
		source, err := canvas.NewTiledImage(pixels)
		if err != nil {
			return nil, err
		}
		result[symbol] = source
	}
	return result, nil
})

func toolbarArtworkPixel(symbol ui.Symbol, x, y float64) color.NRGBA {
	var c color.NRGBA
	paint := func(hit bool, v color.NRGBA) {
		if hit {
			c = v
		}
	}
	rgb := func(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 255} }
	rect := func(l, t, r, b float64) bool { return x >= l && x < r && y >= t && y < b }
	round := func(l, t, r, b, radius float64) bool {
		cx := math.Max(l+radius, math.Min(x, r-radius))
		cy := math.Max(t+radius, math.Min(y, b-radius))
		return rect(l, t, r, b) && math.Hypot(x-cx, y-cy) <= radius
	}
	polygon := func(points ...[2]float64) bool {
		inside := false
		j := len(points) - 1
		for i, p := range points {
			q := points[j]
			if (p[1] > y) != (q[1] > y) && x < (q[0]-p[0])*(y-p[1])/(q[1]-p[1])+p[0] {
				inside = !inside
			}
			j = i
		}
		return inside
	}
	switch symbol {
	case ui.SymbolDocument:
		paint(round(19, 8, 77, 88, 7), rgb(54, 120, 190))
		paint(round(23, 12, 73, 84, 4), rgb(245, 250, 255))
		paint(rect(31, 25, 65, 31), rgb(63, 145, 218))
		paint(rect(31, 42, 65, 46) || rect(31, 54, 65, 58) || rect(31, 66, 56, 70), rgb(130, 169, 204))
	case ui.SymbolAdd:
		paint(round(19, 8, 73, 86, 7), rgb(38, 123, 210))
		paint(round(23, 12, 69, 82, 4), rgb(240, 248, 255))
		paint(rect(30, 26, 60, 30) || rect(30, 38, 57, 42) || rect(30, 50, 54, 54), rgb(115, 164, 211))
		paint(math.Hypot(x-69, y-69) < 19, rgb(27, 165, 90))
		paint(rect(66, 58, 72, 80) || rect(58, 66, 80, 72), rgb(255, 255, 255))
	case ui.SymbolFolder:
		paint(round(8, 19, 52, 44, 6), rgb(27, 127, 193))
		paint(round(8, 29, 88, 82, 7), rgb(30, 143, 206))
		paint(round(12, 35, 84, 78, 5), rgb(90, 191, 240))
		paint(rect(16, 40, 80, 47), rgb(141, 218, 249))
	case ui.SymbolSave:
		paint(round(12, 10, 84, 86, 8), rgb(30, 103, 196))
		paint(round(17, 15, 79, 81, 5), rgb(57, 147, 234))
		paint(round(27, 13, 69, 40, 3), rgb(237, 246, 255))
		paint(rect(58, 17, 64, 35), rgb(68, 89, 115))
		paint(round(26, 51, 70, 84, 3), rgb(39, 61, 90))
		paint(rect(33, 58, 63, 61) || rect(33, 68, 59, 71), rgb(145, 188, 226))
	case ui.SymbolUndo, ui.SymbolRedo:
		paint(round(9, 10, 87, 87, 19), rgb(123, 77, 207))
		if symbol == ui.SymbolRedo {
			paint(round(9, 10, 87, 87, 19), rgb(15, 147, 160))
			x = 96 - x
		}
		paint(polygon([2]float64{20, 44}, [2]float64{44, 22}, [2]float64{44, 36}, [2]float64{57, 36}, [2]float64{71, 43}, [2]float64{77, 56}, [2]float64{77, 73}, [2]float64{64, 73}, [2]float64{64, 59}, [2]float64{58, 51}, [2]float64{44, 51}, [2]float64{44, 65}), rgb(255, 255, 255))
	case ui.SymbolDelete:
		paint(round(23, 27, 73, 88, 7), rgb(204, 63, 71))
		paint(rect(29, 31, 67, 80), rgb(238, 91, 91))
		paint(round(15, 17, 81, 29, 4), rgb(73, 90, 110))
		paint(round(35, 8, 61, 22, 4), rgb(73, 90, 110))
		paint(rect(40, 12, 56, 17), rgb(220, 229, 237))
		paint(rect(37, 39, 42, 72) || rect(54, 39, 59, 72), rgb(255, 208, 206))
	}
	return c
}

// ToolbarIcon returns immutable original-color sample artwork. This example
// package imports only public Vellum APIs, never gallery/driver internals.
func ToolbarIcon(symbol ui.Symbol) (*canvas.TiledImage, error) {
	icons, err := toolbarColorIcons()
	if err != nil {
		return nil, err
	}
	icon := icons[symbol]
	if icon == nil {
		return nil, fmt.Errorf("artwork: unknown sample action %q", symbol)
	}
	return icon, nil
}
