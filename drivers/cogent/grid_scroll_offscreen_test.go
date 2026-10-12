//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"image"
	"testing"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/render"
	"cogentcore.org/core/paint/renderers/rasterx"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestManagedGridScrollSurfaceFullReplayPixels(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			model, _ := testNativeGrid(t, 10_000)
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			config.ExperimentalGridScrollReuse = true
			disabled := false
			m, a := gridManual(t, config, func() ui.Node {
				return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Disabled: disabled}
			})
			reuse := 0
			for frame, step := range []image.Point{{}, {0, 12}, {0, 12}, {96, 12}, {-24, -5}, {0, 0}, {3, 1}, {90, 24}, {}, {}, {}, {}, {0, 12}, {3, 2}} {
				if frame > 0 {
					if err := a.Do(func() {
						switch frame {
						case 8:
							if err := model.ReplaceSource(&scrollGridPaintSource{nativeGridSource: nativeGridSource{rows: 10_000, revision: 2}, prefix: "changed אבג é "}); err != nil {
								t.Error(err)
							}
						case 9:
							if err := model.SelectToKey("row-4", "c3", virtual.GridSelectReplace); err != nil {
								t.Error(err)
							}
						case 10:
							disabled = true
						case 11:
							disabled = false
						}
						if err := model.Scroll(step.X, step.Y); err != nil {
							t.Error(err)
						}
						m.implementation.find("one").state.body.Scene.NeedsRender()
						_ = a.Invalidate("one")
					}); err != nil {
						t.Fatal(err)
					}
				}
				gridFrame(t, m)
				actual, err := m.Capture(testContext, "one")
				if err != nil {
					t.Fatal(err)
				}
				var expected *image.RGBA
				var stats rasterx.ScrollStats
				if err := a.Do(func() {
					scene := m.implementation.find("one").state.body.Scene
					renderer, ok := scene.Renderer.(*rasterx.Renderer)
					if !ok {
						t.Fatal("unexpected native raster renderer")
					}
					stats = renderer.ScrollStats()
					scene.RenderWidget()
					items := scene.Painter.RenderDone()
					hints := 0
					for i, item := range items {
						if push, ok := item.(*render.ContextPush); ok && push.Scroll != nil {
							if !push.Scroll.VerifyPaint {
								t.Fatal("native pane uses unchecked epoch promise")
							}
							copy := *push
							copy.Scroll = nil
							items[i] = &copy
							hints++
						}
					}
					if hints != 4 {
						t.Fatalf("got %d native surface groups", hints)
					}
					oracle := rasterx.New(math32.FromPoint(scene.SceneGeom.Size)).(*rasterx.Renderer)
					expected = oracle.Render(items).Image().(*image.RGBA)
				}); err != nil {
					t.Fatal(err)
				}
				if actual.Bounds() != expected.Bounds() || !bytes.Equal(actual.Pix, expected.Pix) {
					count := 0
					var first image.Point
					for y := actual.Rect.Min.Y; y < actual.Rect.Max.Y; y++ {
						for x := actual.Rect.Min.X; x < actual.Rect.Max.X; x++ {
							if actual.RGBAAt(x, y) != expected.RGBAAt(x, y) {
								if count == 0 {
									first = image.Pt(x, y)
								}
								count++
							}
						}
					}
					t.Fatalf("frame %d different pixels=%d first=%v actual=%v full=%v stats=%+v", frame, count, first, actual.RGBAAt(first.X, first.Y), expected.RGBAAt(first.X, first.Y), stats)
				}
				if stats.CapturedPixels != stats.RedrawPixels {
					t.Fatal("native cache rewrote more than redrawn/exposed area", stats)
				}
				reuse += stats.ReusedGroups
				t.Logf("frame=%d dpi=%g scroll=%v stats=%+v", frame, dpi, step, stats)
				if stats.CachedBytes > 32<<20 || stats.DescriptorBytes > 8<<20 {
					t.Fatal("cache exceeded bounded renderer budgets", stats)
				}
			}
			if reuse == 0 {
				t.Fatal("native pixel oracle never exercised actual reuse")
			}
		})
	}
}

// Stable identities and changed native text make unchanged lifetime/epoch
// insufficient to admit the previous pixels.
type scrollGridPaintSource struct {
	nativeGridSource
	prefix string
}

func (s *scrollGridPaintSource) Cell(key, column string) (string, error) {
	value, err := s.nativeGridSource.Cell(key, column)
	return s.prefix + value, err
}

func TestManagedGridScrollReuseDisabledByDefault(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			model, _ := testNativeGrid(t, 10_000)
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			if config.ExperimentalGridScrollReuse {
				t.Fatal("manual default enables experimental cache")
			}
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			for frame := 0; frame < 3; frame++ {
				if err := a.Do(func() {
					if frame > 0 {
						if err := model.Scroll(3, 12); err != nil {
							t.Error(err)
						}
						_ = a.Invalidate("one")
					}
				}); err != nil {
					t.Fatal(err)
				}
				gridFrame(t, m)
				if err := a.Do(func() {
					scene := m.implementation.find("one").state.body.Scene
					stats := scene.Renderer.(*rasterx.Renderer).ScrollStats()
					if stats.Groups != 0 || stats.CachedBytes != 0 || stats.DescriptorBytes != 0 {
						t.Fatal("default emitted/retained cache", stats)
					}
					scene.RenderWidget()
					items := scene.Painter.RenderDone()
					for _, item := range items {
						if push, ok := item.(*render.ContextPush); ok && push.Scroll != nil {
							t.Fatal("default native hint emitted")
						}
					}
				}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
