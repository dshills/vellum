//go:build offscreen

package cogent

import (
	"image/color"
	"testing"

	"cogentcore.org/core/text/textcore"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

func TestMacSharedPresentationAllocatedControlsAndProse(t *testing.T) {
	for _, dark := range []bool{false, true} {
		name := "light"
		if dark {
			name = "dark"
		}
		t.Run(name, func(t *testing.T) {
			cfg := manualConfig(t, 7)
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = ui.Some(color.NRGBA{})
			model, err := editor.New("A proportional writing surface 😀", editor.Config{})
			if err != nil {
				t.Fatal(err)
			}
			m, _ := simpleManual(t, cfg, func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
					{ID: "field", Kind: ui.TextField, Text: "Search", Accessibility: ui.Accessibility{Name: "Search"}},
					{ID: "number", Kind: ui.NumberField, Number: 3, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}},
					{ID: "body", Kind: ui.MultilineText, Editor: model, Prose: true, SoftWrap: true, Style: ui.MacProseStyle(dark)},
				}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			for _, id := range []string{"field", "number"} {
				r := findRecord(w.state.root, id)
				bounds := r.widget.AsWidget().Geom.TotalBBox
				if bounds.Dy() < 26 || bounds.Dy() > 34 {
					t.Fatal("shared control height", id, bounds)
				}
			}
			r := findRecord(w.state.root, "body")
			ed := r.widget.(*textcore.Editor)
			if ed.LineNumberPixels() != 0 || ed.EditingState().Text != model.Snapshot().Text {
				t.Fatal("prose lost shared presentation/document")
			}
			if ed.Styles.Font.CustomFont != ui.MacFontFamily {
				t.Fatal("font fallback not shared", ed.Styles.Font)
			}
			if _, err := m.Capture(testContext, "one"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
