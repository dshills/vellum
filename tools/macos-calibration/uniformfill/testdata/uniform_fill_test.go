// Copyright (c) 2026 Vellum contributors. SPDX-License-Identifier: MIT.
package pimage

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"testing"
)

func fillSeed() (*image.RGBA, *image.RGBA) {
	parent := image.NewRGBA(image.Rect(-13, -17, 117, 105))
	for i := range parent.Pix {
		parent.Pix[i] = byte(i*37 + 11)
	}
	return parent, parent.SubImage(image.Rect(-5, -9, 103, 91)).(*image.RGBA)
}
func TestOpaqueUniformFastPathOracle(t *testing.T) {
	colors := []color.Color{color.RGBA{17, 29, 71, 255}, color.NRGBA{3, 5, 7, 255}, color.RGBA64{1, 257, 65534, 65535}, color.NRGBA64{599, 50000, 12345, 65535}, color.Alpha{255}, color.Alpha16{65535}, color.Gray{71}, color.Gray16{40123}}
	for _, op := range []draw.Op{draw.Src, draw.Over} {
		for _, c := range colors {
			for _, rect := range []image.Rectangle{image.Rect(-5, -9, 103, 91), image.Rect(0, 0, 1, 1), image.Rect(-2, -3, 1, -1), image.Rect(3, 7, 98, 81)} {
				a, ad := fillSeed()
				b, bd := fillSeed()
				u := image.NewUniform(c)
				sp := image.Pt(31, -47)
				draw.Draw(bd, rect, u, sp, op)
				if !fillOpaqueUniform(ad, rect, u, sp, op) {
					t.Fatal("eligible path fell back")
				}
				if !bytes.Equal(a.Pix, b.Pix) {
					t.Fatalf("pixel/padding parity %T %v %v", c, op, rect)
				}
			}
		}
	}
}

type countedFillColor struct{ calls *int }

func (c countedFillColor) RGBA() (uint32, uint32, uint32, uint32) {
	*c.calls++
	return 771, 1285, 1799, 65535
}
func TestOpaqueUniformFallbackAndCalls(t *testing.T) {
	for _, tc := range []struct {
		c    color.Color
		rect image.Rectangle
		sp   image.Point
		op   draw.Op
	}{
		{color.NRGBA{11, 31, 71, 0}, image.Rect(0, 0, 20, 20), image.Point{}, draw.Src},
		{color.NRGBA{11, 31, 71, 99}, image.Rect(0, 0, 20, 20), image.Point{}, draw.Over},
		{color.RGBA{1, 2, 3, 255}, image.Rect(-7, -12, 30, 30), image.Point{}, draw.Src},
		{color.RGBA{1, 2, 3, 255}, image.Rect(0, 0, 20, 20), image.Pt(image.White.Bounds().Max.X-2, 0), draw.Src},
		{color.RGBA{1, 2, 3, 255}, image.Rect(0, 0, 20, 20), image.Pt(image.White.Bounds().Min.X-2, 0), draw.Over},
		{color.RGBA{1, 2, 3, 255}, image.Rect(0, 0, 20, 20), image.Point{}, draw.Op(99)},
	} {
		a, ad := fillSeed()
		b, bd := fillSeed()
		u := image.NewUniform(tc.c)
		before := append([]byte(nil), a.Pix...)
		if fillOpaqueUniform(ad, tc.rect, u, tc.sp, tc.op) || !bytes.Equal(before, a.Pix) {
			t.Fatal("fallback wrote pixels")
		}
		draw.Draw(bd, tc.rect, u, tc.sp, tc.op)
		p := Params{Cmd: Draw, Rect: tc.rect, Source: u, SourcePos: tc.sp, Op: tc.op}
		p.Render(ad)
		if !bytes.Equal(a.Pix, b.Pix) {
			t.Fatal("fallback integration differs")
		}
	}
	a, ad := fillSeed()
	b, bd := fillSeed()
	ca, cb := 0, 0
	p := Params{Cmd: Draw, Rect: image.Rect(1, 2, 31, 29), Source: image.NewUniform(countedFillColor{&ca}), Op: draw.Over}
	p.Render(ad)
	draw.Draw(bd, p.Rect, image.NewUniform(countedFillColor{&cb}), p.SourcePos, p.Op)
	if ca != cb || !bytes.Equal(a.Pix, b.Pix) {
		t.Fatalf("custom color calls/pixels %d/%d", ca, cb)
	}
}
func TestOpaqueUniformParamsRandomizedOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(41027))
	for i := 0; i < 1200; i++ {
		a, ad := fillSeed()
		b, bd := fillSeed()
		c := color.NRGBA{byte(rng.Intn(256)), byte(rng.Intn(256)), byte(rng.Intn(256)), 255}
		if i%4 == 0 {
			c.A = byte(rng.Intn(255))
		}
		rect := image.Rect(rng.Intn(160)-35, rng.Intn(150)-35, rng.Intn(160)-35, rng.Intn(150)-35)
		if i%19 == 0 {
			rect = image.Rectangle{}
		}
		sp := image.Pt(rng.Intn(70)-35, rng.Intn(70)-35)
		op := draw.Op(i % 2)
		u := image.NewUniform(c)
		var mask image.Image
		if i%7 == 0 {
			mask = image.NewUniform(color.Alpha{123})
		}
		p := Params{Cmd: Draw, Rect: rect, Source: u, SourcePos: sp, Op: op, Mask: mask, MaskPos: image.Pt(-1, 3)}
		wantRect := rect
		if rect == (image.Rectangle{}) {
			wantRect = bd.Bounds()
		}
		draw.DrawMask(bd, wantRect, u, sp, mask, p.MaskPos, op)
		p.Render(ad)
		if p.Rect != wantRect || !bytes.Equal(a.Pix, b.Pix) {
			t.Fatalf("seed41027 case%d params/pixels differ", i)
		}
	}
}
func BenchmarkOpaqueUniform(b *testing.B) {
	for _, fast := range []bool{false, true} {
		name := "stdlib"
		if fast {
			name = "rowcopy"
		}
		b.Run(name, func(b *testing.B) {
			dst := image.NewRGBA(image.Rect(0, 0, 1519, 1266))
			rect := dst.Bounds().Inset(8)
			src := []*image.Uniform{image.NewUniform(color.NRGBA{255, 0, 0, 255}), image.NewUniform(color.NRGBA{0, 255, 0, 255})}
			b.SetBytes(int64(rect.Dx() * rect.Dy() * 4))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if fast {
					if !fillOpaqueUniform(dst, rect, src[i&1], rect.Min, draw.Over) {
						b.Fatal("unexpected fallback")
					}
				} else {
					draw.Draw(dst, rect, src[i&1], rect.Min, draw.Over)
				}
			}
			b.StopTimer()
			if dst.RGBAAt(9, 9).A != 255 {
				b.Fatal("no output")
			}
		})
	}
}
