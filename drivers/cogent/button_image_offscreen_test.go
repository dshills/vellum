//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events/key"
	"fmt"
	"image"
	"image/color"
	"math"
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

func TestButtonIconAboveTextGeometryRoutingAndRestoration(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, fullColor := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark-%v-color-%v", dark, fullColor), func(t *testing.T) {
				source := fullColorButtonSource(t, false)
				placement := ui.IconAboveText
				font, height := float32(12), float32(52)
				gap := ui.Some(float32(4))
				calls := 0
				cfg := manualConfig(t, 7)
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				m, a := simpleManual(t, cfg, func() ui.Node {
					n := ui.Node{ID: "stacked", Kind: ui.Button, Text: "Save", IconPlacement: placement, Appearance: ui.ButtonTool, Activate: func() { calls++ }, Accessibility: ui.Accessibility{Name: "Save drawing"}, Style: ui.Style{Padding: ui.Some(ui.Insets{Top: 5, Bottom: 5, Left: 6, Right: 6}), Gap: gap, Typography: ui.Some(ui.Typography{Size: font}), MinSize: ui.Some(ui.Size{Width: height + 12, Height: height}), MaxSize: ui.Some(ui.Size{Width: height + 12, Height: height})}}
					if fullColor {
						n.IconImage = source
					} else {
						n.Icon = ui.SymbolSave
					}
					return ui.Node{ID: "root", Kind: ui.Toolbar, Focus: ui.FocusRequest{ID: "stacked", Sequence: 1}, Children: []ui.Node{n}}
				})
				unitSettle(t, m)
				w := m.implementation.find("one")
				r := findRecord(w.state.root, "stacked")
				button := r.widget.(*core.Button)
				icon := func() core.Widget {
					if fullColor {
						return button.ChildByName(buttonImageName).(*managedImage)
					}
					return button.ChildByName("icon").(*core.Icon)
				}
				label := button.ChildByName("text").(*core.Text)
				var lastIconWidth int
				for _, size := range []struct{ font, height float32 }{{12, 52}, {13, 60}, {16, 72}} {
					font, height = size.font, size.height
					if err := a.Invalidate("one"); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					artBox, textBox := icon().AsWidget().Geom.TotalBBox, label.Geom.TotalBBox
					content := button.Geom.ContentBBox
					if artBox.Empty() || textBox.Empty() || artBox.Max.Y > textBox.Min.Y || !artBox.In(content) || !textBox.In(content) {
						t.Fatal("stacked parts overlap or clip", size, artBox, textBox, content)
					}
					if math.Abs(float64((artBox.Min.X+artBox.Max.X)-(textBox.Min.X+textBox.Max.X))) > 3 {
						t.Fatal("artwork/caption are not centered", artBox, textBox)
					}
					if button.ChildByName("space") != nil {
						t.Fatal("inline spacer retained in stacked layout")
					}
					if r.widget != button || button.ChildByName("text") != label {
						t.Fatal("layout replaced action or caption")
					}
					lastIconWidth = artBox.Dx()
					frame, err := m.Capture(testContext, "one")
					if err != nil {
						t.Fatal(err)
					}
					if fullColor {
						box := icon().AsWidget().Geom.ContentBBox
						y := (box.Min.Y + box.Max.Y) / 2
						left := frame.RGBAAt(box.Min.X+box.Dx()/6, y)
						right := frame.RGBAAt(box.Max.X-1-box.Dx()/6, y)
						if left.R < 200 || right.B < 190 {
							t.Fatal("stacked artwork lost color", left, right)
						}
					}
					before := calls
					for _, box := range []image.Rectangle{artBox, textBox} {
						for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
							if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: box.Min.Add(box.Size().Div(2))}); err != nil {
								t.Fatal(err)
							}
							unitSettle(t, m)
						}
					}
					if calls != before+2 {
						t.Fatal("icon and caption must each activate once", calls, before)
					}
				}
				// Local explicit zero must override the adapter's default gap; omission
				// restores it. This also pins managed/native styler ordering.
				gap = ui.Some(float32(0))
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if button.Styles.Gap.Y.Dots != 0 || button.Styles.Padding.Top.Value != 5 {
					t.Fatal("local spacing overridden", button.Styles.Gap, button.Styles.Padding)
				}
				gap = ui.Optional[float32]{}
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if button.Styles.Gap.Y.Value != 4 {
					t.Fatal("default stacked gap not restored", button.Styles.Gap)
				}
				nodes, _ := accessibilitySnapshot(w.state.root, nil)
				count := 0
				for _, node := range nodes {
					if node.DomainID == "stacked" {
						count++
						if node.Name != "Save drawing" {
							t.Fatal("caption replaced action semantics")
						}
					} else if node.DomainID != "root" {
						t.Fatal("private content exposed", node)
					}
				}
				if count != 1 {
					t.Fatal("expected one accessible action", count)
				}
				before := calls
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter, Modifiers: 1 << key.NumLock}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if calls != before+1 {
					t.Fatal("keyboard activation", calls, before)
				}
				placement = ui.IconLeading
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				artBox, textBox := icon().AsWidget().Geom.TotalBBox, label.Geom.TotalBBox
				if artBox.Max.X > textBox.Min.X || artBox.Dx() >= lastIconWidth {
					t.Fatal("inline layout/icon size not restored", artBox, textBox, lastIconWidth)
				}
				if fullColor && textBox.Min.X-artBox.Max.X < 4 {
					t.Fatal("full-color inline artwork touches its label", artBox, textBox)
				}
				if !fullColor && button.ChildByName("space") == nil {
					t.Fatal("native inline spacer not restored")
				}
				placement = ""
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if r.widget != button || button.ChildByName("text") != label {
					t.Fatal("default placement replaced action")
				}
			})
		}
	}
}
