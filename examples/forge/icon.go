package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// writeAppIcon renders original procedural artwork at the packager's 1024px
// source size. It contains no Autodesk symbols or third-party image/font bytes.
func writeAppIcon(path string) error {
	im := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	for y := 70; y < 954; y++ {
		for x := 70; x < 954; x++ {
			dx, dy := 0, 0
			if x < 220 {
				dx = 220 - x
			} else if x > 804 {
				dx = x - 804
			}
			if y < 220 {
				dy = 220 - y
			} else if y > 804 {
				dy = y - 804
			}
			if dx*dx+dy*dy > 150*150 {
				continue
			}
			im.SetNRGBA(x, y, color.NRGBA{R: uint8(22 + y/45), G: uint8(54 + y/35), B: uint8(94 + y/20), A: 255})
		}
	}
	poly := func(points []image.Point, c color.NRGBA) {
		for y := 170; y < 850; y++ {
			for x := 170; x < 850; x++ {
				inside := false
				for i, j := 0, len(points)-1; i < len(points); j, i = i, i+1 {
					a, b := points[i], points[j]
					if (a.Y > y) != (b.Y > y) && float64(x) < float64(b.X-a.X)*float64(y-a.Y)/float64(b.Y-a.Y)+float64(a.X) {
						inside = !inside
					}
				}
				if inside {
					im.SetNRGBA(x, y, c)
				}
			}
		}
	}
	poly([]image.Point{{512, 205}, {802, 365}, {512, 532}, {222, 365}}, rgb(108, 179, 255))
	poly([]image.Point{{222, 365}, {512, 532}, {512, 835}, {222, 666}}, rgb(55, 120, 226))
	poly([]image.Point{{512, 532}, {802, 365}, {802, 666}, {512, 835}}, rgb(240, 157, 74))
	poly([]image.Point{{512, 280}, {665, 365}, {512, 454}, {359, 365}}, rgb(212, 234, 255))
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	e = png.Encode(f, im)
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
