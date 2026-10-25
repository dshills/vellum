package ui

import "testing"

func TestMacPresentationSharedDefaultsAndOverrides(t *testing.T) {
	for _, dark := range []bool{false, true} {
		theme := MacTheme(dark)
		if theme.Default.Typography.Value != MacTypography(13, 400) {
			t.Fatal("shared font stack lost")
		}
		body := MacProseStyle(dark)
		if err := body.Validate(); err != nil {
			t.Fatal(err)
		}
		if body.Border.Value.Width != 0 || body.Background.Value != theme.Palette.Value.Field || body.Foreground.Value != theme.Palette.Value.Ink {
			t.Fatal("prose palette/frame mismatch")
		}
		custom := body.Overlay(Style{Typography: Some(MacTypography(18, 600))})
		if custom.Typography.Value.Size != 18 || MacProseStyle(dark).Typography.Value.Size != 14 {
			t.Fatal("shared style mutated")
		}
	}
}
