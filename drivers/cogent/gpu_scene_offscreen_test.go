//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"image"
	"testing"

	"cogentcore.org/core/paint/renderers/gpuscene"
	"github.com/dshills/vellum/ui"
)

func TestManagedGPUSceneSoftwareFallbackMatchesOrdinary(t *testing.T) {
	for _, dpi := range []float32{96, 120, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			var baseline []*image.RGBA
			for _, candidate := range []bool{false, true} {
				model, _ := testNativeGrid(t, 10000)
				config := manualConfig(t, 7)
				config.LogicalDPI = dpi
				config.GPUSceneRendering = candidate
				m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
				for frame, step := range []image.Point{{0, 0}, {90, 24}, {-3, 2}, {3, -2}} {
					if frame != 0 {
						if err := a.Do(func() {
							if err := model.Scroll(step.X, step.Y); err != nil {
								t.Error(err)
							}
							_ = a.Invalidate("one")
						}); err != nil {
							t.Fatal(err)
						}
					}
					gridFrame(t, m)
					pixels, err := m.Capture(testContext, "one")
					if err != nil {
						t.Fatal(err)
					}
					if candidate {
						if pixels.Bounds() != baseline[frame].Bounds() || !bytes.Equal(pixels.Pix, baseline[frame].Pix) {
							t.Fatal("GPU adapter software fallback differs", frame)
						}
						if err := a.Do(func() {
							renderer, ok := m.implementation.find("one").state.body.Scene.Renderer.(*gpuscene.Renderer)
							if !ok || renderer.Stats().CPUFallbacks == 0 || renderer.Stats().Submitted != 0 {
								t.Fatal("software path did not exercise the GPU scene adapter's CPU fallback")
							}
						}); err != nil {
							t.Fatal(err)
						}
					} else {
						baseline = append(baseline, pixels)
					}
				}
				if err := m.Close(testContext); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
