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
	overflow := ui.WithToolbarOverflow(n)
	if err := ui.Validate(overflow); err != nil {
		t.Fatal(err)
	}
	menu := overflow.Children[1]
	if menu.ID != ui.ToolbarOverflowID("tools") || menu.Options[0].IconImage != source || menu.Options[0].Label != "Open" {
		t.Fatal("public overflow lost artwork/label")
	}
	menu.Change(ui.ChangeEvent{SelectedID: "open"})

	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestPublicToolbarCustomizationAndOverflow(t *testing.T) {
	calls := 0
	catalogue := ui.Node{ID: "drawing-tools", Kind: ui.Toolbar, Children: []ui.Node{
		{ID: "open", Kind: ui.Button, Text: "Open", Icon: ui.SymbolFolder, Activate: func() { calls++ }},
		{ID: "save", Kind: ui.Button, Text: "Save", Icon: ui.SymbolSave, Activate: func() { calls++ }},
	}}
	layout := ui.ToolbarLayout{Order: []string{"save", "open"}, Hidden: []string{"open"}}
	applied, err := layout.Apply(catalogue)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied.Children) != 1 || applied.Children[0].ID != "save" {
		t.Fatal("layout lost")
	}
	overflow := ui.WithToolbarOverflow(applied)
	if err := ui.Validate(overflow); err != nil {
		t.Fatal(err)
	}
	overflow.Children[1].Change(ui.ChangeEvent{SelectedID: "save"})
	if calls != 1 {
		t.Fatal("action changed")
	}
	customizer := ui.ToolbarCustomizer{ID: "customize", Toolbar: catalogue, Value: layout, Open: true, Dismiss: func() {}, Change: func(next ui.ToolbarLayout) { layout = next }}
	panel, err := customizer.View()
	if err != nil || panel.Kind != ui.Dialog || !panel.Sheet {
		t.Fatal("public customization sheet", err)
	}
}
