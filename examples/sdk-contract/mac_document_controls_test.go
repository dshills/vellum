package sdkcontract_test

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"testing"
)

func TestPublicMacDocumentControls(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	pixels.SetNRGBA(0, 0, color.NRGBA{R: 220, G: 40, B: 180, A: 255})
	art, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	list := &ui.DocumentList{ID: "sdk-documents", Label: "Documents", Columns: ui.DefaultDocumentColumns(), Items: []ui.DocumentItem{{ID: "drawing", Name: "Drawing", Kind: "Document", Image: art}}}
	outline := &ui.OutlineView{ID: "sdk-outline", Label: "Outline", Items: []ui.OutlineItem{{ID: "project", Label: "Project", Image: art}, {ID: "drawing", Parent: "project", Label: "Drawing", Image: art}}, ExpandedIDs: []string{"project"}}
	dial := &ui.AngleDial{ID: "sdk-angle", Label: "Rotation", Degrees: 30, Step: 15, IconImage: art, Change: func(ui.AngleEvent) {}}
	for _, view := range []func() (ui.Node, error){list.View, outline.View, dial.View} {
		n, err := view()
		if err != nil {
			t.Fatal(err)
		}
		if err = ui.Validate(n); err != nil {
			t.Fatal(err)
		}
	}
	if err := (app.ColorPanelOptions{Title: "Colors", Color: color.NRGBA{R: 255, A: 255}, ShowsAlpha: true}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (app.FontPanelOptions{Title: "Fonts", Name: "HelveticaNeue", Size: 18}).Validate(); err != nil {
		t.Fatal(err)
	}
}
