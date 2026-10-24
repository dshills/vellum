//go:build offscreen

package cogent

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/system/composer"
	"github.com/dshills/vellum/ui"
)

func TestMacManagedInputBordersFollowWindowActivation(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, kind := range []string{"combo", "accessory", "secure", "plain", "borderless", "number", "button"} {
			t.Run(kind+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				cfg := manualConfig(t, 965)
				bg := cfg.Theme.Default.Background
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Background = bg
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				style := cfg.Theme.Kinds[ui.TextField]
				style.Padding = ui.Some(ui.Insets{})
				calls := 0
				target, outerID := "plain", "plain"
				var render func() (ui.Node, error)
				switch kind {
				case "combo", "accessory":
					p := &ui.ComboBox{ID: "combo", Label: "Units", Text: "Millimeters", Options: []ui.Option{{ID: "mm", Label: "Millimeters"}}, Style: style, Change: func(ui.ComboEvent) { calls++ }}
					render, target, outerID = p.View, ui.ComboInputID(p.ID), p.ID
					if kind == "accessory" {
						target = ui.ComboMenuID(p.ID)
					}
				case "secure":
					p := &ui.SecureField{ID: "secure", Text: "synthetic", AllowReveal: true, Style: style, Change: func(ui.ChangeEvent) { calls++ }}
					render, target, outerID = p.View, ui.SecureInputID(p.ID), p.ID
				default:
					target, outerID = kind, kind
					render = func() (ui.Node, error) {
						n := ui.Node{ID: kind, Kind: ui.TextField, Text: "Draft", Change: func(ui.ChangeEvent) { calls++ }}
						if kind == "borderless" {
							n.Style = ui.Style{Border: ui.Some(ui.Border{}), Background: cfg.Theme.Kinds[ui.TextField].Background}
						}
						if kind == "number" {
							n = ui.Node{ID: kind, Kind: ui.NumberField, NumberButtons: ui.NumberButtonsStepper, Number: 1, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}, Change: func(ui.ChangeEvent) { calls++ }}
						}
						if kind == "button" {
							n = ui.Node{ID: kind, Kind: ui.Button, Text: "Action", Activate: func() { calls++ }}
						}
						return n, nil
					}
				}
				view := func() ui.Node {
					n, err := render()
					if err != nil {
						panic(err)
					}
					return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: target, Sequence: 1}, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12))}, Children: []ui.Node{n}}
				}
				m, a := simpleManual(t, cfg, view)
				unitSettle(t, m)
				managed := m.implementation.find("one")
				w := managed.state
				outer := findRecord(w.root, outerID)
				resolved, err := ui.ResolveStyle(outer.node, cfg.Theme)
				if err != nil || !resolved.Border.Set {
					t.Fatal("fixture has no explicit resting border", err)
				}
				box := outer.widget.AsWidget().Geom.TotalBBox
				widths := func() [4]float32 {
					b := outer.widget.AsWidget().Styles.Border.Width
					return [4]float32{b.Top.Value, b.Right.Value, b.Bottom.Value, b.Left.Value}
				}
				originalWidth := widths()
				var edit *core.TextField
				switch field := findRecord(w.root, target).widget.(type) {
				case *core.TextField:
					edit = field
				case *core.Spinner:
					edit = &field.TextField
				}
				if edit != nil && !edit.SetEditingSelection(0, 1) {
					t.Fatal("fixture selection")
				}
				unitSettle(t, m) // Render the synthetic selection before measuring activation frames.
				bindWindowOriginCustom(w, nil, func(*nativeWindow, events.Event) {})
				cd := w.body.Scene.RenderWindow().SystemWindow.Composer().(*composer.ComposerDrawer)
				old := cd.Drawer
				draw := &activationDrawer{Drawer: old}
				cd.Drawer = draw
				t.Cleanup(func() { cd.Drawer = old })
				imager := old.(interface{ GetImage() *image.RGBA })
				// Compare painted border/margin pixels, including rounded/offset borders.
				// Exclude the editable content and its independently blinking caret.
				pixels := func() []byte {
					frame := imager.GetImage()
					var data []byte
					for y := box.Min.Y; y < box.Max.Y; y++ {
						for x := box.Min.X; x < box.Max.X; x++ {
							if image.Pt(x, y).In(outer.widget.AsWidget().Geom.ContentBBox) {
								continue
							}
							c := frame.RGBAAt(x, y)
							data = append(data, c.R, c.G, c.B, c.A)
						}
					}
					return data
				}
				beforePixels := pixels()
				v := &windowActivation{}
				t.Cleanup(v.close)
				if err := bindWindowActivationState(w, v); err != nil {
					t.Fatal(err)
				}
				pump := func() {
					t.Helper()
					if _, err := managed.core.Step(testContext, 128); err != nil {
						t.Fatal(err)
					}
				}
				for range 4 {
					if err := a.Post(func() {}); err != nil {
						t.Fatal(err)
					}
				}
				beforePaints := draw.calls
				v.observe(false)
				if widths() != originalWidth {
					t.Fatal("observation styled before originating delivery")
				}
				pump()
				s := &outer.widget.AsWidget().Styles
				if s.Border.Width.Top.Value != resolved.Border.Value.Width || s.Border.Width.Bottom.Value != resolved.Border.Value.Width {
					t.Fatal("inactive input retained focus width", s.Border.Width, resolved.Border.Value)
				}
				got := color.NRGBAModel.Convert(s.Border.Color.Top.At(0, 0)).(color.NRGBA)
				if got != resolved.Border.Value.Color {
					t.Fatal("inactive input retained active border color", got, resolved.Border.Value.Color)
				}
				if !findRecord(w.root, target).widget.AsWidget().ContainsFocus() || outer.widget.AsWidget().Geom.TotalBBox != box || calls != 0 {
					t.Fatal("appearance changed focus, geometry or accepted value", calls)
				}
				if edit != nil {
					state := edit.EditingState()
					if state.Anchor != 0 || state.Focus != 1 {
						t.Fatal("activation lost editing selection", state)
					}
				}
				if a.Pending().Ordinary != 4 || draw.calls != beforePaints+1 {
					t.Fatal("activation drained model work or composed extra frames", a.Pending(), beforePaints, draw.calls)
				}
				if bytes.Equal(pixels(), beforePixels) {
					t.Fatal("inactive border produced no pixel change", kind, box)
				}
				v.observe(true)
				pump()
				if widths() != originalWidth || !bytes.Equal(pixels(), beforePixels) {
					t.Fatal("reactivation did not restore original cue")
				}
				beforePaints = draw.calls
				v.observe(true)
				pump()
				if draw.calls != beforePaints {
					t.Fatal("duplicate activation repainted")
				}
				// Keep all geometry assertions in native scene pixels at the fixed test DPI.
				raw := math32.RectFromPosSizeMax(outer.widget.AsWidget().Geom.Pos.Total, outer.widget.AsWidget().Geom.Size.Actual.Total)
				if raw.Empty() || !raw.In(imager.GetImage().Bounds()) {
					t.Fatal("fixture outline outside captured frame", raw)
				}
			})
		}
	}
}
