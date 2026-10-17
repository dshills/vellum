//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/ppath"
	"cogentcore.org/core/paint/renderers/rasterx"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func gridPaintPixels(scene *core.Scene, reference bool) (*image.RGBA, int) {
	scene.SetProperty(gridReferenceBoxPaintProperty, reference)
	scene.RenderWidget()
	items := scene.Painter.RenderDone()
	renderer := rasterx.New(math32.FromPoint(scene.SceneGeom.Size)).(*rasterx.Renderer)
	return renderer.Render(items).Image().(*image.RGBA), len(items)
}

func TestManagedGridPanePaintMatchesOriginal(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 162, 192} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("dpi%g/dark%v", dpi, dark), func(t *testing.T) {
				model, _ := testNativeGrid(t, 10000)
				config := manualConfig(t, 7)
				config.LogicalDPI = dpi
				if dark {
					config.Theme.Default.Foreground = ui.Some(color.NRGBA{R: 230, G: 230, B: 230, A: 255})
					config.Theme.Default.Background = ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255})
				}
				disabled := false
				m, a := gridManual(t, config, func() ui.Node {
					return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Disabled: disabled}
				})
				reduced := 0
				for frame, delta := range []image.Point{{}, {0, 12}, {0, 12}, {90, 12}, {-30, -5}, {}, {}, {0, 24}} {
					if frame != 0 {
						if err := a.Do(func() {
							if frame == 5 {
								_ = model.SelectToKey("row-4", "c3", virtual.GridSelectReplace)
							}
							if frame == 6 {
								disabled = true
							}
							if frame == 7 {
								disabled = false
							}
							if err := model.Scroll(delta.X, delta.Y); err != nil {
								t.Error(err)
							}
							_ = a.Invalidate("one")
						}); err != nil {
							t.Fatal(err)
						}
					}
					gridFrame(t, m)
					if err := a.Do(func() {
						window := m.implementation.find("one").state
						scene, g := window.body.Scene, window.root.control.grid
						original, before := gridPaintPixels(scene, true)
						actual, after := gridPaintPixels(scene, false)
						if !bytes.Equal(original.Pix, actual.Pix) {
							t.Fatalf("frame%d pixels differ", frame)
						}
						if after > before {
							t.Fatal("elision added commands", before, after)
						}
						reduced += before - after
						for _, tx := range append(append([]*core.Text{}, g.cells...), g.headers...) {
							if !tx.Styles.RenderBox {
								t.Fatal("temporary box flag escaped pane render")
							}
						}
						// Packet replay exercises the actual capture path at every DPI,
						// including the162-DPI bitmap-mask route.
						scene.RenderWidget()
						items := scene.Painter.RenderDone()
						capture := rasterx.New(math32.FromPoint(scene.SceneGeom.Size)).(*rasterx.Renderer)
						packet := capture.ScenePacket(items)
						if packet == nil {
							t.Fatal("optimized full scene refused capture")
						}
						replay := image.NewRGBA(actual.Bounds())
						if !packet.Replay(replay) || !bytes.Equal(replay.Pix, original.Pix) {
							t.Fatal("captured packet differs from original painter")
						}
					}); err != nil {
						t.Fatal(err)
					}
				}
				if reduced == 0 {
					t.Fatal("no actual pane coverage at requested DPI")
				}
				t.Log("removed render commands", reduced)
			})
		}
	}
}

func TestManagedGridPanePaintDecorationsAndForeignContent(t *testing.T) {
	model, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
	gridFrame(t, m)
	if err := a.Do(func() {
		window := m.implementation.find("one").state
		scene, g := window.body.Scene, window.root.control.grid
		// Decorations and an opaque different cell background must remain.
		tx := g.cells[len(g.cells)/2]
		tx.Styles.Background = image.NewUniform(color.NRGBA{R: 180, G: 40, B: 90, A: 255})
		tx.Styles.Border.Width.Set(units.Dot(1))
		original, _ := gridPaintPixels(scene, true)
		actual, _ := gridPaintPixels(scene, false)
		if !bytes.Equal(original.Pix, actual.Pix) {
			t.Fatal("decorated cell changed")
		}
		// Even a shadow/border extending geometrically beyond this cell is
		// clipped by its ordinary Text.StartRender TotalBBox. Neighbor boxes
		// may elide without erasing or accumulating that preceding foreground.
		tx.Styles.Border.Width.Set(units.Dot(6))
		tx.Styles.Border.Offset.Set(units.Dot(4))
		tx.Styles.BoxShadow = []styles.Shadow{{OffsetX: units.Dot(12), OffsetY: units.Dot(5), Blur: units.Dot(6), Color: image.NewUniform(color.NRGBA{R: 10, G: 20, B: 30, A: 180})}}
		actual, _ = gridPaintPixels(scene, false)
		original, _ = gridPaintPixels(scene, true)
		if !bytes.Equal(original.Pix, actual.Pix) {
			t.Fatal("preceding shadow/border escaped disjoint cell clip")
		}
		scene.SetProperty(gridReferenceBoxPaintProperty, false)
		pane := tx.Parent.(*gridPane)
		restore := pane.elideCoveredCellBoxes()
		elided := 0
		for _, child := range pane.Children {
			if label, ok := child.(*core.Text); ok && !label.Styles.RenderBox {
				elided++
			}
		}
		if elided == 0 || !tx.Styles.RenderBox {
			restore()
			t.Fatal("decorated-cell proof did not retain foreground while eliding neighbours")
		}
		restore()
		// An extra child invalidates that pane's proof rather than guessing
		// whether the cell/foreign paint overlaps.
		foreign := core.NewText(g.panes[3]).SetText("foreign")
		foreign.Style()
		original, _ = gridPaintPixels(scene, true)
		actual, _ = gridPaintPixels(scene, false)
		if !bytes.Equal(original.Pix, actual.Pix) {
			t.Fatal("foreign child changed")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

// Make proof failures observable independently of pixel coincidence.
func TestManagedGridPanePaintProofRefusesUnsafeState(t *testing.T) {
	for name, mutate := range map[string]func(*gridPane){
		"translucent": func(p *gridPane) { p.Styles.Opacity = .5 },
		"state layer": func(p *gridPane) { p.Styles.StateLayer = .2 },
		"rounded":     func(p *gridPane) { p.Styles.Border.Radius.Set(units.Dot(4)) },
		"border":      func(p *gridPane) { p.Styles.Border.Width.Set(units.Dot(1)) },
		"fractional":  func(p *gridPane) { p.Geom.Pos.Total.X += .25 },
		"no reset":    func(p *gridPane) { p.Styles.RenderBox = false },
	} {
		t.Run(name, func(t *testing.T) {
			model, _ := testNativeGrid(t, 1000)
			m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			if err := a.Do(func() {
				g := m.implementation.find("one").state.root.control.grid
				pane := g.panes[3]
				mutate(pane)
				restore := pane.elideCoveredCellBoxes()
				defer restore()
				for _, tx := range g.cells {
					if !tx.Styles.RenderBox {
						t.Fatal("unsafe pane elided a cell box")
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagedGridPanePaintFreshBackgroundAndPartialRepaint(t *testing.T) {
	model, _ := testNativeGrid(t, 10000)
	config := manualConfig(t, 7)
	config.LogicalDPI = 162
	m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
	gridFrame(t, m)
	if err := a.Do(func() {
		window := m.implementation.find("one").state
		scene, g := window.body.Scene, window.root.control.grid
		_, _ = gridPaintPixels(scene, false)
		// The pane retains the earlier ActualBackground until StartRender.
		// Run optimized first so a preceding reference render cannot refresh it.
		g.host.Styles.Background = image.NewUniform(color.RGBA{R: 31, G: 51, B: 81, A: 255})
		g.cells[0].SetState(true, states.Selected)
		actual, _ := gridPaintPixels(scene, false)
		expected, _ := gridPaintPixels(scene, true)
		if !bytes.Equal(actual.Pix, expected.Pix) {
			t.Fatal("theme change or preceding selected cell changed pixels")
		}
		var tx *core.Text
		for _, cell := range g.cells {
			if cell.IsVisible() && integerGridBox(&cell.WidgetBase) && cell.Text != "" {
				tx = cell
				break
			}
		}
		if tx == nil {
			t.Fatal("missing visible text")
		}
		// A long-to-short isolated text repaint must clear the old glyphs after
		// the optimized parent traversal has restored its original box flag.
		tx.SetText("x")
		tx.UpdateWidget()
		if !tx.Styles.RenderBox {
			t.Fatal("individual repaint lost clearing")
		}
		tx.RenderWidget()
		partial := scene.Painter.RenderDone()
		renderer := rasterx.New(math32.FromPoint(scene.SceneGeom.Size)).(*rasterx.Renderer)
		copy(renderer.Image().(*image.RGBA).Pix, actual.Pix)
		got := renderer.Render(partial).Image().(*image.RGBA)
		expected, _ = gridPaintPixels(scene, true)
		if !bytes.Equal(got.Pix, expected.Pix) {
			t.Fatal("individual repaint left stale pixels")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestManagedGridPanePaintRefusesOverlapAndAncestorEffects(t *testing.T) {
	for _, kind := range []string{"overlap", "mask", "clip path", "bounds path", "transform"} {
		t.Run(kind, func(t *testing.T) {
			model, _ := testNativeGrid(t, 10000)
			config := manualConfig(t, 7)
			config.LogicalDPI = 162
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			if err := a.Do(func() {
				g := m.implementation.find("one").state.root.control.grid
				p := g.panes[3]
				count := func() int {
					n := 0
					for _, tx := range g.cells {
						if !tx.Styles.RenderBox {
							n++
						}
					}
					return n
				}
				restore := p.elideCoveredCellBoxes()
				if count() == 0 {
					restore()
					t.Fatal("unmodified pane did not exercise elision")
				}
				restore()
				if count() != 0 {
					t.Fatal("baseline flags not restored")
				}
				ctx := p.Scene.Painter.Context()
				originalContext := *ctx
				defer func() { *ctx = originalContext }()
				switch kind {
				case "overlap":
					var visible []*core.Text
					for _, child := range p.Children {
						tx := child.(*core.Text)
						if tx.IsVisible() && !tx.Geom.TotalBBox.Empty() {
							visible = append(visible, tx)
						}
					}
					if len(visible) < 2 {
						t.Fatal("missing overlap fixture")
					}
					original := visible[1].Geom.TotalBBox
					defer func() { visible[1].Geom.TotalBBox = original }()
					visible[1].Geom.TotalBBox = visible[0].Geom.TotalBBox
				case "mask":
					ctx.Mask = image.NewAlpha(image.Rect(0, 0, 8, 8))
				case "clip path":
					ctx.ClipPath = *ppath.New().Rectangle(0, 0, 8, 8)
				case "bounds path":
					ctx.Bounds.Path = *ppath.New().Rectangle(0, 0, 8, 8)
				case "transform":
					ctx.Cumulative = math32.Scale2D(2, 2)
				}
				restore = p.elideCoveredCellBoxes()
				defer restore()
				if count() != 0 {
					t.Fatal("unsafe ancestor/overlap retained elision")
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
