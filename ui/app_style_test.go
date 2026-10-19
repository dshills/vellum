package ui

import (
	"image/color"
	"math"
	"testing"
)

func TestAppSurfaceStyleResolution(t *testing.T) {
	base := Style{Grow: Some(float32(1)), GrowX: Some(float32(3)), GrowY: Some(float32(2)), TextAlign: Some(AlignEnd), CornerRadius: Some(float32(12)), Border: Some(Border{Width: 2, Color: color.NRGBA{R: 42, A: 255}})}
	clear := Style{GrowX: Some(float32(0)), GrowY: Some(float32(0)), TextAlign: Some(AlignStart), CornerRadius: Some(float32(0)), Border: Some(Border{})}
	got, err := ResolveStyle(Node{ID: "card", Kind: Column, Style: clear}, Theme{Default: base})
	if err != nil {
		t.Fatal(err)
	}
	if got.Grow.Value != 1 || got.GrowX.Value != 0 || got.GrowY.Value != 0 || got.TextAlign.Value != AlignStart || got.CornerRadius.Value != 0 || got.Border.Value != (Border{}) {
		t.Fatalf("explicit zeros lost: %+v", got)
	}
	if base.Overlay(Style{}) != base {
		t.Fatal("omitted properties changed")
	}
}
func TestAppSurfaceStyleRejectsInvalid(t *testing.T) {
	for _, v := range []float32{-1, float32(math.NaN()), float32(math.Inf(1))} {
		for _, s := range []Style{{GrowX: Some(v)}, {GrowY: Some(v)}, {CornerRadius: Some(v)}, {Border: Some(Border{Width: v})}} {
			if s.Validate() == nil {
				t.Fatal("accepted invalid style", s)
			}
		}
	}
	if (Style{TextAlign: Some(Alignment("justify"))}).Validate() == nil {
		t.Fatal("unsupported alignment")
	}
}
