package ui

import (
	"image/color"
	"math"
	"testing"
)

func TestStyleResolutionOverridesAndExplicitZero(t *testing.T) {
	theme := Theme{Default: Style{Gap: Some(float32(12)), Grow: Some(float32(2)), MinSize: Some(Size{80, 20}), Foreground: Some(color.NRGBA{A: 255})}, Kinds: map[Kind]Style{Row: {Gap: Some(float32(6))}}}
	n := Node{ID: "r", Kind: Row, Grow: true, MinWidth: 100, Style: Style{Gap: Some(float32(0)), Grow: Some(float32(0)), MinSize: Some(Size{})}}
	got, err := ResolveStyle(n, theme)
	if err != nil {
		t.Fatal(err)
	}
	if got.Gap.Value != 0 || got.Grow.Value != 0 || got.MinSize.Value != (Size{}) || got.Foreground != theme.Default.Foreground {
		t.Fatalf("wrong precedence: %+v", got)
	}
	n.Style = Style{}
	got, err = ResolveStyle(n, theme)
	if err != nil {
		t.Fatal(err)
	}
	if got.Gap.Value != 6 || got.Grow.Value != 1 || got.MinSize.Value != (Size{100, 20}) {
		t.Fatalf("defaults/legacy not restored: %+v", got)
	}
	clone := theme.Clone()
	theme.Kinds[Row] = Style{Gap: Some(float32(99))}
	if clone.Kinds[Row].Gap.Value != 6 {
		t.Fatal("theme snapshot aliases original kind map")
	}
}
func TestStyleRejectsInvalidAndMergedConstraints(t *testing.T) {
	for _, s := range []Style{
		{Gap: Some(float32(math.NaN()))}, {Grow: Some(float32(math.Inf(1)))},
		{Padding: Some(Insets{Left: -1})}, {MinSize: Some(Size{Width: -1})},
		{MaxSize: Some(Size{Height: float32(math.Inf(1))})},
		{Align: Some(Alignment("space-between"))}, {Justify: Some(Alignment("unknown"))},
		{Typography: Some(Typography{Size: 0})}, {Typography: Some(Typography{Size: 14, Weight: 450})},
		{MinSize: Some(Size{30, 0}), MaxSize: Some(Size{20, 0})},
	} {
		if s.Validate() == nil {
			t.Errorf("accepted invalid style %+v", s)
		}
	}
	theme := Theme{Default: Style{MaxSize: Some(Size{20, 0})}}
	if _, err := ResolveStyle(Node{ID: "bad", Kind: Label, MinWidth: 30}, theme); err == nil {
		t.Fatal("theme maximum and legacy minimum conflict accepted")
	}
	if err := Validate(Node{ID: "bad", Kind: Label, MinWidth: float32(math.NaN())}); err == nil {
		t.Fatal("invalid legacy minimum accepted")
	}
	if err := theme.Validate(); err != nil {
		t.Fatal(err)
	}
	theme.Kinds = map[Kind]Style{Label: {MinSize: Some(Size{30, 0})}}
	if err := theme.Validate(); err == nil {
		t.Fatal("invalid merged theme accepted")
	}
}
