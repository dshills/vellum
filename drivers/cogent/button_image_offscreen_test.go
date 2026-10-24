//go:build offscreen

package cogent

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

func fullColorButtonSource(t *testing.T, swapped bool) *canvas.TiledImage {
	t.Helper()
	pixels := image.NewNRGBA(image.Rect(0, 0, 36, 24))
	colors := []color.NRGBA{{R: 240, A: 255}, {G: 220, A: 255}, {B: 230, A: 255}}
	if swapped {
		colors[0], colors[2] = colors[2], colors[0]
	}
	for y := range 24 {
		for x := range 36 {
			pixels.SetNRGBA(x, y, colors[x/12])
		}
	}
	source, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	return source
}
func TestButtonImageFullColorPixelsReplacementAndLifetime(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, kind := range []ui.Kind{ui.Button, ui.MenuButton} {
			t.Run(fmt.Sprintf("%v-%s", dark, kind), func(t *testing.T) {
				source := fullColorButtonSource(t, false)
				replacement := fullColorButtonSource(t, true)
				symbol := ui.Symbol("")
				text := "Open"
				disabled := false
				calls := 0
				cfg := manualConfig(t, 7)
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				m, a := simpleManual(t, cfg, func() ui.Node {
					n := ui.Node{ID: "action", Kind: kind, Text: text, IconImage: source, Icon: symbol, Disabled: disabled, Appearance: ui.ButtonTool, Style: ui.Style{Foreground: ui.Some(color.NRGBA{R: 255, B: 255, A: 255}), Typography: ui.Some(ui.Typography{Size: 16})}, Accessibility: ui.Accessibility{Name: "Open"}}
					if kind == ui.Button {
						n.Activate = func() { calls++ }
					} else {
						n.Options = []ui.Option{{ID: "one", Label: "One"}}
						n.Change = func(ui.ChangeEvent) { calls++ }
					}
					return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "action", Sequence: 1}, Children: []ui.Node{n}}
				})
				unitSettle(t, m)
				w := m.implementation.find("one")
				r := findRecord(w.state.root, "action")
				button := r.widget.(*core.Button)
				picture := func() *managedImage {
					t.Helper()
					child, ok := button.ChildByName(buttonImageName).(*managedImage)
					if !ok {
						t.Fatal("missing image slot")
					}
					return child
				}
				original := picture()
				if button.Child(0) != original {
					t.Fatal("image is not the leading artwork")
				}
				assertColors := func(swapped bool) {
					t.Helper()
					frame, err := m.Capture(testContext, "one")
					if err != nil {
						t.Fatal(err)
					}
					box := picture().Geom.ContentBBox
					// 3:2 source is contained in a square icon slot: sample the middle line.
					y := (box.Min.Y + box.Max.Y) / 2
					left := frame.RGBAAt(box.Min.X+box.Dx()/6, y)
					right := frame.RGBAAt(box.Max.X-1-box.Dx()/6, y)
					if !swapped {
						if left.R < 200 || left.B > 30 || right.B < 190 || right.R > 30 {
							t.Fatal("image tinted/cropped", left, right, box)
						}
					} else if left.B < 190 || right.R < 200 {
						t.Fatal("replacement pixels stale", left, right)
					}
				}
				assertColors(false)
				if kind == ui.Button {
					// Image-only hit testing still reaches the same parent action exactly once.
					text = ""
					if err := a.Invalidate("one"); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					_, target := imageManualElement(t, m, "action")
					for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
						if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Target: "action", TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}); err != nil {
							t.Fatal(err)
						}
						unitSettle(t, m)
					}
					if calls != 1 {
						t.Fatal("artwork stole or duplicated click", calls)
					}
				}
				source = replacement
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if r.widget != button || picture() != original || picture().source != replacement {
					t.Fatal("replacement lost button/image identity")
				}
				assertColors(true)
				disabled = true
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if picture().Styles.Opacity >= 1 {
					t.Fatal("disabled color art not dimmed", picture().Styles.Opacity)
				}
				nodes, _ := accessibilitySnapshot(w.state.root, nil)
				actions := 0
				for _, n := range nodes {
					if n.DomainID == "action" {
						actions++
					} else if n.DomainID != "root" {
						t.Fatal("private artwork exposed semantics", n)
					}
				}
				if actions != 1 {
					t.Fatal("private artwork changed action semantics", actions)
				}
				source = nil
				symbol = ui.SymbolFolder
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if button.ChildByName(buttonImageName) != nil || original.source != nil {
					t.Fatal("removed artwork reference retained")
				}
				symbol = ""
				source = replacement
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				last := picture()
				disposeControl(r)
				last.UpdateWidget()
				button.UpdateWidget()
				if last.source != nil || r.node.IconImage != nil {
					t.Fatal("retired action retained pixels")
				}
			})
		}
	}
}
