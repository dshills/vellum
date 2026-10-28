//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestToggleGlyphResolvesExplicitScopedAccentInk(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			cfg := manualConfig(t, 7)
			cfg.Theme = ui.MacThemeWithAccent(dark, color.NRGBA{235, 180, 30, 255})
			cfg.Theme.Default.Background = ui.Some(cfg.Theme.Palette.Value.Window)
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			role := "accent-ink"
			checked := true
			m, a := simpleManual(t, cfg, func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "toggle", Kind: ui.ToggleButton, Icon: ui.SymbolVisibility, Checked: checked, Style: ui.Style{ForegroundRole: role}, Accessibility: ui.Accessibility{Name: "Visibility"}, Change: func(e ui.ChangeEvent) { checked = e.Checked }}}}
			})
			unitSettle(t, m)
			r := findRecord(m.implementation.find("one").state.root, "toggle")
			toggle := r.widget.(*core.Switch)
			assert := func(want color.NRGBA) {
				t.Helper()
				stack := toggle.ChildByName("stack", 0).(*core.Frame)
				for _, name := range []string{"icon-on", "icon-off"} {
					icon := stack.ChildByName(name, 0).(*core.Icon)
					got := color.NRGBAModel.Convert(icon.Styles.Color.At(0, 0)).(color.NRGBA)
					if got != want {
						t.Fatal(name, got, want)
					}
				}
			}
			assert(cfg.Theme.Palette.Value.AccentInk)
			if e := a.Post(func() { checked = false; role = "secondary"; _ = a.Invalidate("one") }); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			assert(cfg.Theme.Palette.Value.Secondary)
		})
	}
}
