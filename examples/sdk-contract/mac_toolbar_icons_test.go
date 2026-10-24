package sdkcontract_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

func TestPublicFullColorToolbarAction(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	pixels.SetNRGBA(0, 0, color.NRGBA{R: 240, A: 255})
	pixels.SetNRGBA(1, 0, color.NRGBA{B: 240, A: 255})
	source, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	n := ui.Node{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{{ID: "open", Kind: ui.Button, IconImage: source, Text: "Open", IconPlacement: ui.IconAboveText, Appearance: ui.ButtonTool, Activate: func() { calls++ }, Accessibility: ui.Accessibility{Name: "Open"}}}}
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	if got := ui.FocusOrder(n); len(got) != 1 || got[0] != "open" {
		t.Fatal(got)
	}
	n.Children[0].Activate()
	if calls != 1 {
		t.Fatal(calls)
	}
}
