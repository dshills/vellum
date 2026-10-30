//go:build offscreen

package cogent

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// This matrix changes only the deterministic software window and logical DPI.
// It does not change the desktop display settings or qualify native backing scale.
func TestMacPersonalWidgetsGeometryMatrix(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, width := range []int{480, 660, 960} {
			for _, scale := range []int{1, 2} {
				t.Run(fmt.Sprintf("dark=%v/width=%d/scale=%d", dark, width, scale), func(t *testing.T) {
					cfg := manualConfig(t, 947)
					cfg.LogicalDPI = float32(160 * scale)
					cfg.Theme = ui.MacTheme(dark)
					cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
					cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
					if dark {
						cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
					}
					// Embedded editors own their insets; the surrounding row owns only its outline.
					fieldStyle := cfg.Theme.Kinds[ui.TextField]
					fieldStyle.Padding = ui.Some(ui.Insets{})
					secure := &ui.SecureField{ID: "secure", Label: "Synthetic password", Text: "synthetic", AllowReveal: true, Style: fieldStyle, Change: func(ui.ChangeEvent) {}}
					path := &ui.PathControl{ID: "path", MaxVisible: 3, MaxLabelRunes: 12, Items: []ui.PathItem{{ID: "home", Label: "Home"}, {ID: "docs", Label: "Documents"}, {ID: "drawings", Label: "Drawings"}, {ID: "file", Label: "Untitled.vellum", Kind: ui.PathFile}}, Navigate: func(ui.PathEvent) {}}
					font := &ui.FontPicker{ID: "font", Label: "Drawing font", Value: ui.FontSelection{FamilyID: "noto", TypefaceID: "regular", Size: 16}, Families: []ui.FontFamily{{ID: "noto", Name: "Noto Sans", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}}, FieldStyle: cfg.Theme.Kinds[ui.TextField], Change: func(ui.FontSelectionEvent) {}}
					tags := &ui.TokenField{ID: "tags", Label: "Tags", Tokens: []ui.Token{{ID: "design", Text: "Design"}}, InputStyle: cfg.Theme.Kinds[ui.TextField], Change: func(ui.TokenEvent) {}}
					well := &ui.ColorWell{ID: "well", Label: "Drawing color", Color: color.NRGBA{R: 52, G: 120, B: 246, A: 255}, PopoverStyle: cfg.Theme.Kinds[ui.FormGroup], Change: func(ui.ColorEdit) {}}
					var viewErr error
					view := func() ui.Node {
						views := []struct {
							id     string
							render func() (ui.Node, error)
						}{{"secure", secure.View}, {"path", path.View}, {"font", font.View}, {"tags", tags.View}, {"well", well.View}}
						root := ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}}
						for _, v := range views {
							n, err := v.render()
							if err != nil {
								viewErr = err
								return ui.Node{ID: "fixture-error", Kind: ui.Label, Text: "Invalid test fixture"}
							}
							root.Children = append(root.Children, ui.Node{ID: v.id + "-group", Kind: ui.FormGroup, Text: v.id, Children: []ui.Node{n}})
						}
						return root
					}
					a, err := app.New(app.Config{QueueCapacity: 16})
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(a.Close)
					if err = a.Open(app.Window{ID: "one", Title: "Mac widget geometry", Width: float32(width * scale), Height: float32(900 * scale), View: view}); err != nil {
						t.Fatal(err)
					}
					m, err := NewManual(testContext, a, cfg)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() {
						if err := m.Close(testContext); err != nil {
							t.Error(err)
						}
					})
					unitSettle(t, m)
					if viewErr != nil {
						t.Fatal(viewErr)
					}
					frame, err := m.Capture(testContext, "one")
					if err != nil {
						t.Fatal(err)
					}
					if frame.Bounds() != image.Rect(0, 0, width*scale, 900*scale) {
						t.Fatal("unexpected frame size", frame.Bounds())
					}
					root := m.implementation.find("one").state.root
					for _, groupID := range []string{"secure", "path", "font", "tags", "well"} {
						group := findRecord(root, groupID+"-group")
						content := group.control.chrome.host.Geom.ContentBBox
						control := findRecord(root, groupID).widget.AsWidget().Geom.TotalBBox
						if control.Empty() || !control.In(content) {
							t.Fatal("composite escaped form", groupID, control, content)
						}
					}
					// The regular AppKit reference on macOS 27 uses 26-point fields.
					// Allow font/rounding differences while rejecting the previous
					// 38–40-point composites. DPI here is deterministic, not a display change.
					for _, id := range []string{"secure", ui.FontFamilyID("font"), ui.FontTypefaceID("font"), ui.FontSizeID("font"), ui.PathSegmentID("path", "home")} {
						widget := findRecord(root, id).widget.AsWidget()
						height := float32(widget.Geom.TotalBBox.Dy()) / float32(scale)
						if height < 18 || height > 32 {
							t.Fatal("regular control height diverged from compact Mac sizing", id, height)
						}
					}
					for _, pair := range [][2]string{{ui.SecureInputID("secure"), "secure"}, {ui.ComboInputID(ui.FontFamilyID("font")), ui.FontFamilyID("font")}, {ui.ComboMenuID(ui.FontFamilyID("font")), ui.FontFamilyID("font")}, {ui.ComboInputID(ui.FontTypefaceID("font")), ui.FontTypefaceID("font")}, {ui.FontSizeID("font"), "font/font/style-row"}, {ui.TokenInputID("tags"), "tags/tokens/entry"}, {ui.ColorWellButtonID("well"), "well"}} {
						child := findRecord(root, pair[0]).widget.AsWidget().Geom.TotalBBox
						parent := findRecord(root, pair[1]).widget.AsWidget().Geom.ContentBBox
						if child.Empty() || !child.In(parent) {
							t.Fatal("interactive control clipped", pair, child, parent)
						}
					}
					face := findRecord(root, ui.FontTypefaceID("font")).widget.AsWidget().Geom.TotalBBox
					size := findRecord(root, ui.FontSizeID("font")).widget.AsWidget().Geom.TotalBBox
					if face.Overlaps(size) {
						t.Fatal("typeface and stepper overlap", face, size)
					}
					w, button := imageManualElement(t, m, ui.ColorWellButtonID("well"))
					for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
						if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: button.ID, TargetIncarnation: button.Incarnation, Button: events.Left}); err != nil {
							t.Fatal(err)
						}
					}
					unitSettle(t, m)
					panel := findRecord(m.implementation.find("one").state.root, "well/color-well/popup")
					if panel == nil || panel.control == nil || panel.control.dialog == nil || !panel.node.Open || !panel.control.dialog.active || panel.control.dialog.body == nil {
						t.Fatal("scaled pointer route did not open color panel")
					}
					rect := panel.control.dialog.body.Scene.SceneGeom.Bounds()
					if rect.Empty() || !rect.In(frame.Bounds()) {
						t.Fatal("scaled color popup escaped viewport", rect, frame.Bounds())
					}
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeEscape}); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					panel = findRecord(m.implementation.find("one").state.root, "well/color-well/popup")
					if viewErr != nil {
						t.Fatal(viewErr)
					}
					if panel == nil || panel.control == nil || panel.control.dialog == nil {
						t.Fatal("popup record not retained on close")
					}
					if panel.control.dialog.active || panel.node.Open || !findRecord(m.implementation.find("one").state.root, button.ID).widget.AsWidget().ContainsFocus() {
						t.Fatal("scaled color panel did not clean up and restore focus")
					}
				})
			}
		}
	}
}
