package ui

import (
	"image/color"
	"testing"
)

func TestScopedThemeValidationAndPaletteSnapshots(t *testing.T) {
	purple := color.NRGBA{R: 126, G: 65, B: 194, A: 255}
	for _, dark := range []bool{false, true} {
		theme := MacThemeWithAccent(dark, purple)
		if err := theme.Validate(); err != nil {
			t.Fatal(err)
		}
		if theme.Palette.Value.Accent != purple {
			t.Fatal("accent lost")
		}
		clone := theme.Clone()
		theme.Kinds[Button] = Style{}
		if !clone.Kinds[Button].Border.Set {
			t.Fatal("theme map was shared")
		}
		n := Node{ID: "scope", Kind: Column, Theme: &clone, Children: []Node{{ID: "field", Kind: TextField, Text: "value"}}}
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
		bad := clone.Clone()
		bad.Palette.Value.Accent.A = 0
		n.Theme = &bad
		if Validate(n) == nil {
			t.Fatal("invalid scoped palette accepted")
		}
	}
	restricted := Theme{Default: Style{MaxSize: Some(Size{Width: 30})}}
	if Validate(Node{ID: "scope", Kind: Column, Theme: &restricted, Children: []Node{{ID: "wide", Kind: Button, MinWidth: 60}}}) == nil {
		t.Fatal("inherited style conflict not rejected")
	}
}
