package main

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestHDRReferenceIncludesPartialAlphaAndValuesAboveOne(t *testing.T) {
	left := referencePixel(1, false, false)
	if left != (color.NRGBA{213, 156, 0, 128}) {
		t.Fatal(left)
	}
	right := referencePixel(1, true, false)
	if right != (color.NRGBA{0, 124, 225, 128}) {
		t.Fatal(right)
	}
	if referencePixel(2, false, false).R <= left.R || referencePixel(2, true, false).B <= right.B {
		t.Fatal("exposure did not preserve HDR above one")
	}
	if referencePixel(2, true, true) != (color.NRGBA{}) {
		t.Fatal("transparent region colored")
	}
}
func TestExposureValidationAndExactUniformBytes(t *testing.T) {
	m, _ := newImagingProgram()
	for _, v := range []float32{0, -1, 17, float32(math.NaN()), float32(math.Inf(1))} {
		if m.setExposure(v) == nil {
			t.Fatalf("accepted %v", v)
		}
	}
	if m.exposureValue() != 1 {
		t.Fatal("invalid edit changed model")
	}
	if err := m.setExposure(2); err != nil {
		t.Fatal(err)
	}
	data := uniformData(m.exposureValue(), image.Pt(800, 600))
	if data != ([8]float32{2, 256, 256, 0, 800, 600, 0, 0}) || len(data)*4 != uniformBytes {
		t.Fatal(data)
	}
	if storageBytes != 1048576 || floatTextureBytes != 524288 {
		t.Fatal("resource accounting")
	}
}
func TestProgramCleanupToleratesPartialInitialization(t *testing.T) {
	m, _ := newImagingProgram()
	m.release()
	m.release()
	buffer, texture, releases := m.resourceIdentity()
	if buffer != nil || texture != nil || releases != 2 {
		t.Fatal("partial cleanup")
	}
}
