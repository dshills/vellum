package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/dshills/vellum/ui"
)

func (g *gallery) imagesView() ui.Node {
	return ui.Node{ID: "images", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "image-instructions", Kind: ui.Label, Text: "Immutable shared pixels. Contain shows the full image; cover crops centrally; stretch fills both axes."},
		{ID: "image-fit", Kind: ui.SegmentedControl, SelectedID: string(g.imageFit), Accessibility: ui.Accessibility{Name: "Image fit"}, Options: []ui.Option{{ID: string(ui.ImageContain), Label: "Contain"}, {ID: string(ui.ImageCover), Label: "Cover"}, {ID: string(ui.ImageStretch), Label: "Stretch"}}, Change: func(e ui.ChangeEvent) { g.imageFitRequests++; g.imageFit = ui.ImageFit(e.SelectedID) }},
		{ID: "image-preview", Kind: ui.Image, Image: g.picture, Fit: g.imageFit, Text: "Three color strips, with transparency in the blue strip", Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 320, Height: 160})}},
		{ID: "image-replace", Kind: ui.Button, Text: "Replace first strip", Activate: func() {
			patch := image.NewNRGBA(image.Rect(0, 0, 32, 32))
			for y := range 32 {
				for x := range 32 {
					patch.SetNRGBA(x, y, color.NRGBA{R: 240, G: 192, A: 255})
				}
			}
			source, err := g.picture.WithPatch(patch.Bounds(), patch)
			if err != nil {
				panic(err)
			}
			g.picture = source
			g.imageReplacements++
		}},
	}}
}
func (g *gallery) colorsView() ui.Node {
	return ui.Node{ID: "colors", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		g.colorPicker.View(),
		{ID: "color-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "color-reset", Kind: ui.Button, Text: "Reset color", Activate: func() {
				g.colorPicker.Color = color.NRGBA{R: 32, G: 96, B: 224, A: 255}
				g.colorPicker.ValidationMessage = ""
			}},
			{ID: "color-disabled", Kind: ui.Checkbox, Text: "Disable color editing", Checked: g.colorPicker.Disabled, Change: func(e ui.ChangeEvent) { g.colorPicker.Disabled = e.Checked }},
			{ID: "color-focus", Kind: ui.Button, Text: "Focus hex", Activate: func() { g.focus.ID = ui.ColorFieldID("ink", ui.ColorHex); g.focus.Sequence++ }},
		}},
	}}
}
func (g *gallery) chromeView() ui.Node {
	tools := make([]ui.Node, 8)
	for i := range tools {
		i := i
		tools[i] = ui.Node{ID: fmt.Sprintf("media-tool-%d", i), Kind: ui.Button, Text: fmt.Sprintf("Tool %d", i), MinWidth: 100, Disabled: i == 1, Activate: func() { g.mediaTools++; g.mediaLast = i }}
	}
	return ui.Node{ID: "chrome", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "media-tools", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Document tools"}, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 260})}, Children: tools},
		{ID: "media-form", Kind: ui.FormGroup, Text: "<Project>", Children: []ui.Node{{ID: "media-name", Kind: ui.TextField, Text: g.text, Change: func(e ui.ChangeEvent) { g.text = e.Text }}}},
		{ID: "media-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Tool %d / actions %d / path %d", g.mediaLast, g.mediaTools, g.breadcrumbRequests)},
	}}
}
func (g *gallery) mediaCatalogueView() ui.Node {
	path, err := g.breadcrumb.View()
	if err != nil {
		panic(err)
	}
	chrome := g.chromeView()
	chrome.Children = append(chrome.Children, path, ui.Node{ID: "media-reset-path", Kind: ui.Button, Text: "Reset location", Activate: func() { g.breadcrumb.Items = galleryBreadcrumbItems() }})
	focus := g.focus
	if focus.Sequence == 0 {
		focus = ui.FocusRequest{ID: "media-tool-0", Sequence: 1}
	}
	return ui.Node{ID: "media-catalogue", Kind: ui.Column, Grow: true, Focus: focus, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16))}, Children: []ui.Node{
		{ID: "media-pages", Kind: ui.Tabs, Grow: true, SelectedID: g.mediaPage, Options: []ui.Option{{ID: "chrome", Label: "Chrome"}, {ID: "images", Label: "Images"}, {ID: "colors", Label: "Colors"}}, Change: func(e ui.ChangeEvent) { g.mediaPage = e.SelectedID }, Children: []ui.Node{chrome, g.imagesView(), g.colorsView()}},
	}}
}
