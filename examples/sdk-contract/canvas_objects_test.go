package sdkcontract_test

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestPublicCanvasObjectSemanticsAndSelection(t *testing.T) {
	scene, err := canvas.New([]canvas.Rect{{ID: "back", Width: 40, Height: 30}, {ID: "front", Width: 40, Height: 30}})
	if err != nil {
		t.Fatal(err)
	}
	node := ui.Node{ID: "canvas", Kind: ui.Canvas2D, Canvas: scene, CanvasObjects: []ui.CanvasObject{{ID: "back", Name: "Back component", Help: "Select component"}, {ID: "front", Name: "Front component"}}}
	if err := ui.Validate(node); err != nil {
		t.Fatal(err)
	}
	point := scene.View.Screen(canvas.Point{X: 20, Y: 15})
	if scene.ObjectAt(point) != "front" || scene.Selected != "" {
		t.Fatal("pure hit test changed selection or paint order")
	}
	if err := scene.Select("back"); err != nil || scene.Selected != "back" {
		t.Fatal("stable selection failed", err)
	}
	if err := scene.Select("missing"); err == nil || scene.Selected != "back" {
		t.Fatal("unknown object changed selection")
	}
	if ui.MaxCanvasObjects != 10000 || ui.MaxCanvasObjectMetadataBytes != 1<<20 {
		t.Fatal("unexpected metadata bounds")
	}
}
