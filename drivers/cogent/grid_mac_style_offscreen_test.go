//go:build offscreen

package cogent

import (
	"fmt"
	"testing"

	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestGridMacPaletteSelectionAndSurfaceParity(t *testing.T) {
	for _, dpi := range []float32{96, 160, 192} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("dpi%g/dark%v", dpi, dark), func(t *testing.T) {
				source := &selectionNativeSource{selectionNativeIndex: selectionNativeIndex{total: 24, revision: 1}}
				model, e := virtual.NewGrid(source, []virtual.GridColumn{{ID: "a", Title: "Service", Width: 150}, {ID: "b", Title: "Status", Width: 110}}, virtual.GridConfig{RowHeight: 22, HeaderHeight: 22, RowSelection: true}, virtual.GridViewport{Width: 500, Height: 280})
				if e != nil {
					t.Fatal(e)
				}
				if e = model.JumpToKey("selection-0", "a"); e != nil {
					t.Fatal(e)
				}
				scoped := ui.MacTheme(dark)
				cfg := manualConfig(t, 7)
				cfg.LogicalDPI = dpi
				m, a := gridManual(t, cfg, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Theme: &scoped} })
				gridFrame(t, m)
				g := surfaceBinding(t, m, a)
				check := func() {
					t.Helper()
					if e := a.Do(func() {
						found := 0
						for _, w := range g.cells {
							if !w.StateIs(states.Selected) {
								continue
							}
							found++
							if w.Styles.Background.At(0, 0) != scoped.Palette.Value.Accent || w.Styles.Color.At(0, 0) != scoped.Palette.Value.AccentInk {
								t.Fatal("selected row ignored scoped palette")
							}
							if w.Styles.Border.Width.Top.Dots != 0 {
								t.Fatal("read-only selected row retained a separate cell box")
							}
							if w.Styles.Padding.Top.Dots <= 0 || w.Styles.Padding.Left.Dots <= 0 {
								t.Fatal("cell text lost its inset")
							}
						}
						if found < 2 {
							t.Fatal("row selection did not cover both columns")
						}
					}); e != nil {
						t.Fatal(e)
					}
				}
				check()
				surfaceEnable(t, m, a, g)
				surfaceCompareWithOriginal(t, m, a, g)
				if e := a.Do(func() { scoped = ui.MacTheme(!dark); _ = a.Invalidate("one") }); e != nil {
					t.Fatal(e)
				}
				gridFrame(t, m)
				check()
				surfaceEnable(t, m, a, g)
				surfaceCompareWithOriginal(t, m, a, g)
			})
		}
	}
}

func TestGridMacEditableAndUnselectedCursorRetained(t *testing.T) {
	g := &gridBinding{r: &record{theme: ui.MacTheme(true)}}
	g.projection.State.Config.RowSelection = true
	for _, test := range []struct{ selected, editing bool }{{true, true}, {false, false}} {
		g.projection.State.Config.CellEditing = test.editing
		var s styles.Style
		s.Defaults()
		s.Border.Width.Set(units.Px(1))
		g.applyScopedCellStyle(&s, test.selected, true)
		if s.Border.Width.Top.Value != 1 {
			t.Fatal("editable or unselected cursor disappeared")
		}
	}
}
