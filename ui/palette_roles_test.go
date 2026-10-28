package ui

import (
	"image/color"
	"testing"
)

func TestPaletteRolesResolveScopedThemesAndExplicitOverrides(t *testing.T) {
	for _, dark := range []bool{false, true} {
		theme := MacThemeWithAccent(dark, color.NRGBA{220, 130, 20, 255})
		n := Node{ID: "selected", Kind: Button, Style: Style{BackgroundRole: "accent", ForegroundRole: "accent-ink"}}
		s, e := ResolveStyle(n, theme)
		if e != nil {
			t.Fatal(e)
		}
		if s.Background.Value != theme.Palette.Value.Accent || s.Foreground.Value != theme.Palette.Value.AccentInk {
			t.Fatal("palette not resolved")
		}
		n.Style = n.Style.Overlay(Style{Border: Some(Border{Width: 1}), BorderRole: "accent"})
		s, e = ResolveStyle(n, theme)
		if e != nil || s.Border.Value.Width != 1 || s.Border.Value.Color != theme.Palette.Value.Accent {
			t.Fatal("border role lost", e)
		}
		custom := color.NRGBA{20, 40, 60, 255}
		n.Style = n.Style.Overlay(Style{Background: Some(custom)})
		s, e = ResolveStyle(n, theme)
		if e != nil || s.Background.Value != custom {
			t.Fatal("explicit override lost", e)
		}
	}
	if (Style{BackgroundRole: "bogus"}).Validate() == nil {
		t.Fatal("unknown role accepted")
	}
}
