package main

import (
	"image"
	"image/color"
	"testing"
)

func TestPixelAcceptanceRejectsClampedHDRAndLostAlpha(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	img.SetNRGBA(16, 48, referencePixel(1, false, false))
	img.SetNRGBA(48, 48, referencePixel(1, true, false))
	if err := checkPixels(img, 1); err != nil {
		t.Fatal(err)
	}
	// A premultiplied HDR blue of 1.5 clamped to 1 before presentation maps
	// to straight 2/3 (213 sRGB), instead of the correct 3/4 (225 sRGB).
	img.SetNRGBA(48, 48, color.NRGBA{0, 124, 213, 128})
	if checkPixels(img, 1) == nil {
		t.Fatal("clamped float attachment accepted")
	}
	img.SetNRGBA(48, 48, referencePixel(1, true, false))
	img.SetNRGBA(16, 48, color.NRGBA{213, 156, 0, 255})
	if checkPixels(img, 1) == nil {
		t.Fatal("lost partial alpha accepted")
	}
	img.SetNRGBA(16, 48, referencePixel(1, false, false))
	img.SetNRGBA(16, 8, color.NRGBA{0, 0, 0, 1})
	if checkPixels(img, 1) == nil {
		t.Fatal("nonzero transparent alpha accepted")
	}
}

func TestResizeRequestDiffersFromRestoredNativeSize(t *testing.T) {
	for _, current := range []image.Point{image.Pt(800, 640), image.Pt(900, 700), image.Pt(1000, 760), image.Pt(3072, 864)} {
		request := resizeRequest(current)
		if request == current || request.X < 800 || request.X > 1000 || request.Y < 640 || request.Y > 760 {
			t.Fatalf("resize %v from %v is unchanged or outside fixture bounds", request, current)
		}
	}
}
