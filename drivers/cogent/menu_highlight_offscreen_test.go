//go:build offscreen

package cogent

import (
	"image/color"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
)

func TestMacMenuHasOnePointerOrKeyboardHighlight(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			cfg := manualConfig(t, 811)
			background := cfg.Theme.Default.Background
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = background
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			chosen := ""
			m, _ := simpleManual(t, cfg, func() ui.Node {
				return ui.Node{ID: "menu", Kind: ui.MenuButton, Text: "Units", Focus: ui.FocusRequest{ID: "menu", Sequence: 1}, Options: []ui.Option{{ID: "one", Label: "Millimeters"}, {ID: "two", Label: "Centimeters"}, {ID: "three", Label: "Inches"}, {ID: "disabled", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) { chosen = e.SelectedID }}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			send := func(input ManualInput) {
				t.Helper()
				input.Window = "one"
				input.WindowIncarnation = w.incarnation
				if err := m.Input(testContext, input); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			send(ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter})
			p := w.state.root.control.menu.popup
			if p == nil {
				t.Fatal("menu not open")
			}
			highlight := func(want string) {
				t.Helper()
				count := 0
				for id, item := range p.items {
					got := color.NRGBAModel.Convert(item.Styles.Background.At(0, 0)).(color.NRGBA)
					if got == cfg.Theme.Palette.Value.Accent {
						count++
						if id != want {
							t.Fatal("wrong highlighted row", id, want)
						}
					}
				}
				if count != 1 || !p.items[want].ContainsFocus() {
					t.Fatal("menu needs one focused highlight", want, count)
				}
			}
			// AppKit's regular popup rows are about 24 points. The minimum is
			// content size; adding 24 plus padding made these oversized at 32.
			for id, item := range p.items {
				height := float32(item.Geom.TotalBBox.Dy()) / (item.Styles.UnitContext.DPI / units.DpPerInch)
				if height < 20 || height > 28 {
					t.Fatal("menu row height differs from compact Mac rows", id, height)
				}
			}
			highlight("one")
			move := func(id string) {
				t.Helper()
				box := p.items[id].Geom.TotalBBox
				send(ManualInput{Type: events.MouseMove, Position: box.Min.Add(box.Size().Div(2)).Add(p.scene.SceneGeom.Pos)})
			}
			move("three")
			highlight("three")
			send(ManualInput{Type: events.KeyChord, Code: key.CodeUpArrow})
			highlight("two") // Pointer is still over three; it must not keep a second highlight.
			move("disabled")
			highlight("two")
			send(ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter})
			if chosen != "two" || menuOpen(w.state.root) {
				t.Fatal("keyboard chose the wrong row or failed to dismiss", chosen)
			}
		})
	}
}
