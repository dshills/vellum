//go:build offscreen

package cogent

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// Counts the actual retained-text updater/style passes within one owner sync.
// Later native layout/paint work is separate; its detached RGBA/geometry is
// logged for comparison with the original updater path, even on assertion failure.
func TestManagedTreeRetainedTextsUpdateOncePerSync(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			model := nativeOutline(t, 100_000, true)
			disabled := false
			help := "initial"
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node {
				return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: model, Grow: true, Disabled: disabled, Accessibility: ui.Accessibility{Help: help}}
			})
			unitSettle(t, m)
			binding := m.implementation.find("one").state.root.control.outline
			counters := map[*core.Text]*gridUpdatePass{}
			if err := a.Do(func() {
				for _, row := range binding.rows {
					for _, text := range []*core.Text{row.arrow, row.label} {
						pass := &gridUpdatePass{}
						counters[text] = pass
						text.Updater(func() { pass.updates++ })
						text.FinalStyler(func(*styles.Style) { pass.styles++ })
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
			phases := []struct {
				name   string
				change func() error
			}{
				{"scroll-select", func() error {
					s, e := model.State()
					if e != nil {
						return e
					}
					v := s.Viewport
					v.ScrollY = 23
					return errorsJoinTree(model.SetViewport(v), model.Select("node-2"))
				}},
				{"selection", func() error { return model.Select("node-3") }},
				{"pending", func() error {
					_, e := model.Prepare(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "Node"})
					return e
				}},
			}
			for _, phase := range phases {
				if err := a.Do(func() {
					if e := phase.change(); e != nil {
						t.Error(e)
						return
					}
					for _, pass := range counters {
						*pass = gridUpdatePass{}
					}
					if e := syncTree(binding.r); e != nil {
						t.Error(e)
						return
					}
					for _, row := range binding.rows {
						for _, text := range []*core.Text{row.arrow, row.label} {
							pass := counters[text]
							if pass == nil {
								continue
							}
							if pass.updates != 1 || pass.styles != 1 {
								t.Errorf("%s text retained update/style passes=%d/%d want1/1", phase.name, pass.updates, pass.styles)
							}
						}
					}
				}); err != nil {
					t.Fatal(err)
				}
				if _, e := m.Frame(testContext, "one"); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
				measured := treeInspection(t, a, model)
				if len(measured.Rows) > 500 || measured.Generation != treeState(t, a, model).Generation {
					t.Fatal("completed bounded tree geometry missing")
				}
				if err := a.Do(func() {
					for i, row := range binding.rows {
						if row.label.StateIs(states.Selected) != binding.projection.Rows[i].Selected || row.label.Text != binding.projection.Rows[i].Label {
							t.Error("selection/label semantics changed")
						}
					}
				}); err != nil {
					t.Fatal(err)
				}
				pixels, e := m.Capture(testContext, "one")
				if e != nil || pixels == nil || pixels.Bounds().Empty() {
					t.Fatal("empty detached frame", e)
				}
				sum := sha256.Sum256(pixels.Pix)
				t.Logf("TREE_UPDATE_RGBA dpi=%g phase=%s bounds=%v stride=%d sha256=%x", dpi, phase.name, pixels.Bounds(), pixels.Stride, sum)
			}
		})
	}
}
func errorsJoinTree(a, b error) error {
	if a != nil {
		return a
	}
	return b
}
