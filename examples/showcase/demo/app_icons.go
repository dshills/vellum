package demo

import (
	"fmt"
	"image"
	"math"
)

// AppIcon creates original 1024px artwork for this example's bundle packager.
func AppIcon(kind string) (*image.NRGBA, error) {
	if kind != "forma" && kind != "luma" && kind != "pulse" {
		return nil, fmt.Errorf("unknown demo %q", kind)
	}
	p := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	if kind == "luma" {
		p = landscape(0, 1024, 1024)
	}
	for y := range 1024 {
		for x := range 1024 {
			u, v := float64(x)/1024, float64(y)/1024
			if kind == "forma" {
				c := rgb(247, 243, 232)
				if math.Hypot(u-.43, v-.42) < .23 {
					c = blue
				}
				if u > .54 && u < .83 && v > .54 && v < .83 {
					c = rgb(34, 48, 78)
				}
				if math.Hypot(u-.63, v-.65) < .19 {
					c = rgb(246, 115, 83)
				}
				p.SetNRGBA(x, y, c)
			}
			if kind == "pulse" {
				c := rgb(29, 33, 47)
				curve := .59 - .10*math.Sin(u*10) - .28*math.Exp(-math.Pow((u-.57)/.06, 2))
				if u > .13 && u < .87 && math.Abs(v-curve) < .011 {
					c = rgb(80, 210, 218)
				}
				p.SetNRGBA(x, y, c)
			}
			dx := math.Max(math.Abs(float64(x)-511.5)-302, 0)
			dy := math.Max(math.Abs(float64(y)-511.5)-302, 0)
			alpha := math.Max(0, math.Min(1, 175-math.Hypot(dx, dy)))
			c := p.NRGBAAt(x, y)
			c.A = uint8(alpha * 255)
			p.SetNRGBA(x, y, c)
		}
	}
	return p, nil
}
