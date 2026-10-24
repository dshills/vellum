package ui

import (
	"fmt"
	"image/color"
)

// Palette supplies semantic colors without tinting image artwork or changing OS
// appearance. Colors are opaque straight-alpha sRGB. An omitted palette retains
// backend defaults. Scoped Theme values can change on an owner View boundary.
type Palette struct {
	Window, Panel, Field, Popup color.NRGBA
	Ink, Secondary, Line        color.NRGBA
	Accent, AccentInk           color.NRGBA
}

func (p Palette) Validate() error {
	for _, c := range []color.NRGBA{p.Window, p.Panel, p.Field, p.Popup, p.Ink, p.Secondary, p.Line, p.Accent, p.AccentInk} {
		if c.A != 255 {
			return fmt.Errorf("ui: semantic palette colors must be opaque")
		}
	}
	return nil
}

// MacPalette returns compact application-local Mac appearance tokens. Zero
// accent selects blue. This is a preset, not a system accent/contrast observer.
func MacPalette(dark bool, accent color.NRGBA) Palette {
	if accent == (color.NRGBA{}) {
		accent = color.NRGBA{G: 112, B: 245, A: 255}
	}
	p := Palette{Window: color.NRGBA{246, 246, 248, 255}, Panel: color.NRGBA{249, 249, 250, 255}, Field: color.NRGBA{255, 255, 255, 255}, Popup: color.NRGBA{249, 249, 250, 255}, Ink: color.NRGBA{35, 35, 37, 255}, Secondary: color.NRGBA{100, 100, 105, 255}, Line: color.NRGBA{191, 191, 195, 255}, Accent: accent, AccentInk: color.NRGBA{255, 255, 255, 255}}
	if dark {
		p.Window = color.NRGBA{29, 29, 31, 255}
		p.Panel = color.NRGBA{65, 65, 68, 255}
		p.Field = color.NRGBA{35, 35, 37, 255}
		p.Popup = color.NRGBA{48, 48, 51, 255}
		p.Ink = color.NRGBA{241, 241, 243, 255}
		p.Secondary = color.NRGBA{175, 175, 181, 255}
		p.Line = color.NRGBA{98, 98, 102, 255}
	}
	// Bright custom accents use dark text; artwork remains full color.
	if int(accent.R)*299+int(accent.G)*587+int(accent.B)*114 > 160000 {
		p.AccentInk = color.NRGBA{20, 20, 22, 255}
	}
	return p
}
