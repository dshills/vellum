// Copyright (c) 2026 Vellum contributors. SPDX-License-Identifier: MIT.
package pimage

import (
	"image"
	"image/color"
	"image/draw"
)

// fillOpaqueUniform changes only a fully contained, unmasked opaque builtin
// uniform draw. Unknown colors/images and clipped/translucent draws stay on the
// original path, including their observable RGBA/At calls and source clipping.
func fillOpaqueUniform(dst *image.RGBA, rect image.Rectangle, src image.Image, sp image.Point, op draw.Op) bool {
	if op != draw.Src && op != draw.Over {
		return false
	}
	u, ok := src.(*image.Uniform)
	if !ok || u == nil || rect.Empty() || !rect.In(dst.Bounds()) {
		return false
	}
	switch u.C.(type) {
	case color.RGBA, color.RGBA64, color.NRGBA, color.NRGBA64, color.Alpha, color.Alpha16, color.Gray, color.Gray16:
	default:
		return false
	}
	width, height := rect.Dx(), rect.Dy()
	sb := u.Bounds()
	if width <= 0 || height <= 0 || width > sb.Dx() || height > sb.Dy() ||
		sp.X < sb.Min.X || sp.Y < sb.Min.Y || sp.X > sb.Max.X-width || sp.Y > sb.Max.Y-height ||
		width > len(dst.Pix)/4 {
		return false
	}
	r, g, b, a := u.C.RGBA()
	if a != 0xffff {
		return false
	}
	n := width * 4
	first := dst.PixOffset(rect.Min.X, rect.Min.Y)
	last := dst.PixOffset(rect.Min.X, rect.Max.Y-1)
	if dst.Stride < n || first < 0 || last < first || first > len(dst.Pix)-n || last > len(dst.Pix)-n {
		return false
	}
	row := dst.Pix[first : first+n]
	row[0], row[1], row[2], row[3] = uint8(r>>8), uint8(g>>8), uint8(b>>8), 255
	for filled := 4; filled < n; {
		filled += copy(row[filled:], row[:filled])
	}
	for y := rect.Min.Y + 1; y < rect.Max.Y; y++ {
		offset := dst.PixOffset(rect.Min.X, y)
		copy(dst.Pix[offset:offset+n], row)
	}
	return true
}
