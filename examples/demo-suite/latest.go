package main

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
)

type latestFeatures struct {
	owner                          *demo
	path                           *ui.Breadcrumb
	ink                            *ui.ColorPicker
	picture                        *canvas.TiledImage
	fit                            ui.ImageFit
	split                          float32
	details, showGrid, rejectClose bool
	documents                      []string
	selected                       string
	drafts                         map[string]string
	message                        string
}

func newLatestFeatures(d *demo) *latestFeatures {
	f := &latestFeatures{owner: d, fit: ui.ImageContain, split: .43, details: true, documents: []string{"sketch", "notes"}, selected: "sketch", drafts: map[string]string{"sketch": "A new sketch — edit this draft, then close and reopen its tab.", "notes": "Drafts remain in Go state while tabs are closed."}, message: "Ready — try the controls"}
	f.path = &ui.Breadcrumb{ID: "new-path", Label: "Project location", Items: latestPath()}
	f.path.Navigate = func(e ui.BreadcrumbEvent) {
		for i, p := range f.path.Items {
			if p.ID == e.ItemID {
				f.path.Items = append([]ui.BreadcrumbItem(nil), f.path.Items[:i+1]...)
				f.message = "Navigated to " + p.Label
				d.refresh()
				return
			}
		}
	}
	f.ink = &ui.ColorPicker{ID: "new-ink", Label: "Accent color", Color: color.NRGBA{R: 35, G: 145, B: 210, A: 255}}
	f.ink.Change = func(e ui.ColorEdit) {
		f.ink.Color = e.Apply(f.ink.Color)
		f.ink.ValidationMessage = ""
		f.message = "Accent color updated"
		d.refresh()
	}
	f.ink.Invalid = func(e error) { f.ink.ValidationMessage = e.Error(); d.refresh() }
	pixels := image.NewNRGBA(image.Rect(0, 0, 160, 80))
	for y := range 80 {
		for x := range 160 {
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 255 / 159), G: uint8(y * 255 / 79), B: 180, A: 255})
		}
	}
	var err error
	f.picture, err = canvas.NewTiledImage(pixels)
	if err != nil {
		panic(err)
	}
	return f
}
func latestPath() []ui.BreadcrumbItem {
	return []ui.BreadcrumbItem{{ID: "projects", Label: "Projects"}, {ID: "vellum", Label: "Vellum"}, {ID: "demo", Label: "Demo suite"}}
}
func (f *latestFeatures) closeDocument(e ui.TabCloseEvent) {
	if f.rejectClose {
		f.message = "Close refused — the draft stays open"
		f.owner.refresh()
		return
	}
	for i, id := range f.documents {
		if id == e.PageID {
			f.documents = append(f.documents[:i], f.documents[i+1:]...)
			break
		}
	}
	if f.selected == e.PageID && len(f.documents) > 0 {
		f.selected = f.documents[0]
	}
	f.message = "Closed " + e.PageID + "; its draft is retained"
	f.owner.refresh()
}
func (f *latestFeatures) view() ui.Node {
	d := f.owner
	path, err := f.path.View()
	if err != nil {
		panic(err)
	}
	left := ui.Node{ID: "new-left", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		path,
		{ID: "new-reset-path", Kind: ui.Button, Text: "Reset breadcrumb path", Icon: ui.SymbolRight, Appearance: ui.ButtonQuiet, Activate: func() { f.path.Items = latestPath(); f.message = "Location reset"; d.refresh() }},
		{ID: "new-progress-group", Kind: ui.FormGroup, Text: "Shared progress", Children: []ui.Node{
			{ID: "new-progress-slider", Kind: ui.Slider, Number: d.level, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: func(e ui.ChangeEvent) { d.level = e.Number; d.refresh() }},
			{ID: "new-progress-shapes", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(18))}, Children: []ui.Node{
				{ID: "new-ring", Kind: ui.ProgressRing, Number: d.level / 100, Text: fmt.Sprintf("%.0f%%", d.level), Accessibility: ui.Accessibility{Name: "Circular progress"}},
				{ID: "new-gauge", Kind: ui.ProgressGauge, Number: d.level / 100, Text: fmt.Sprintf("%.0f%%", d.level), Accessibility: ui.Accessibility{Name: "Semicircular progress"}},
			}},
		}},
		{ID: "new-color-section", Kind: ui.Disclosure, Text: "Inline color picker", Checked: f.details, Change: func(e ui.ChangeEvent) { f.details = e.Checked; d.refresh() }, Children: []ui.Node{f.ink.View()}},
	}}
	docs := ui.Node{ID: "new-document-group", Kind: ui.FormGroup, Text: "Closable documents", Children: []ui.Node{
		{ID: "new-doc-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "new-reopen", Kind: ui.Button, Text: "Reopen tabs", Icon: ui.SymbolCheck, Appearance: ui.ButtonSecondary, Activate: func() {
				f.documents = []string{"sketch", "notes"}
				f.selected = "sketch"
				f.message = "Tabs reopened with retained drafts"
				d.refresh()
			}},
			{ID: "new-reject-close", Kind: ui.Checkbox, Text: "Reject close", Checked: f.rejectClose, Change: func(e ui.ChangeEvent) { f.rejectClose = e.Checked; d.refresh() }},
		}},
	}}
	if len(f.documents) > 0 {
		tabs := ui.Node{ID: "new-document-tabs", Kind: ui.Tabs, SelectedID: f.selected, MinHeight: 150, Change: func(e ui.ChangeEvent) { f.selected = e.SelectedID; d.refresh() }, TabClose: f.closeDocument}
		for _, id := range f.documents {
			id := id
			tabs.Options = append(tabs.Options, ui.Option{ID: "new-page-" + id, Label: id, Closable: true})
			tabs.Children = append(tabs.Children, ui.Node{ID: "new-page-" + id, Kind: ui.Column, Children: []ui.Node{{ID: "new-draft-" + id, Kind: ui.TextField, Text: f.drafts[id], Accessibility: ui.Accessibility{Name: id + " draft"}, Change: func(e ui.ChangeEvent) { f.drafts[id] = e.Text; d.refresh() }}}})
		}
		// Tab identities remain independent of their displayed labels.
		tabs.SelectedID = "new-page-" + f.selected
		tabs.Change = func(e ui.ChangeEvent) {
			for _, id := range f.documents {
				if "new-page-"+id == e.SelectedID {
					f.selected = id
				}
			}
			d.refresh()
		}
		tabs.TabClose = func(e ui.TabCloseEvent) {
			for _, id := range f.documents {
				if "new-page-"+id == e.PageID {
					e.PageID = id
					f.closeDocument(e)
					return
				}
			}
		}
		docs.Children = append(docs.Children, tabs)
	} else {
		docs.Children = append(docs.Children, label("new-doc-empty", "All tabs closed. Reopen to recover your drafts."))
	}
	right := ui.Node{ID: "new-right", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "new-actions", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "new-menu", Kind: ui.MenuButton, Text: "View options", Icon: ui.SymbolSettings, Options: []ui.Option{{ID: "grid", Label: "Show grid", Checkable: true, Checked: f.showGrid}, {ID: "divider", Separator: true}, {ID: "reset", Label: "Reset progress"}, {ID: "unavailable", Label: "Unavailable action", Disabled: true}}, Change: func(e ui.ChangeEvent) {
				if e.SelectedID == "grid" {
					f.showGrid = e.Checked
					f.message = fmt.Sprintf("Show grid: %t", f.showGrid)
				} else if e.SelectedID == "reset" {
					d.level = 0
					f.message = "Progress reset"
				}
				d.refresh()
			}},
			{ID: "new-primary", Kind: ui.Button, Text: "Try a primary action", Appearance: ui.ButtonPrimary, Icon: ui.SymbolCheck, Activate: func() { d.level = 100; f.message = "Action completed"; d.refresh() }},
		}},
		docs,
		{ID: "new-image-group", Kind: ui.FormGroup, Text: "Managed image preview", Children: []ui.Node{
			{ID: "new-image-fit", Kind: ui.SegmentedControl, SelectedID: string(f.fit), Options: []ui.Option{{ID: string(ui.ImageContain), Label: "Contain"}, {ID: string(ui.ImageCover), Label: "Cover"}, {ID: string(ui.ImageStretch), Label: "Stretch"}}, Change: func(e ui.ChangeEvent) { f.fit = ui.ImageFit(e.SelectedID); d.refresh() }},
			{ID: "new-image", Kind: ui.Image, Image: f.picture, Fit: f.fit, Text: "Opaque RGB gradient preview", MinWidth: 280, MinHeight: 160},
		}},
		label("new-help", "Drag the divider to resize. Breadcrumb ancestors navigate; document X buttons request a close. The full controls gallery has more examples."),
	}}
	return ui.Node{ID: "latest", Kind: ui.Column, Grow: true, Children: []ui.Node{
		label("release-summary", demoVersion()+": Forge, document workflows, native macOS services and refreshed controls. Explore All feature demos for the full suite."),
		{ID: "new-forge-launch", Kind: ui.Button, Text: "Open Forge design workbench", Disabled: d.busy, Activate: func() { d.launchFeature(featureByID("forge"), false) }},
		{ID: "new-documents-launch", Kind: ui.Button, Text: "Open two-window Documents", Disabled: d.busy, Activate: func() { d.launchFeature(featureByID("documents"), false) }},
		{ID: "new-placement-launch", Kind: ui.Button, Text: "Open palette-to-canvas drag and drop", Icon: ui.SymbolRight, Appearance: ui.ButtonPrimary, Disabled: d.busy, Activate: func() { d.launchFeature(featureByID("placement"), false) }},
		label("new-placement-help", "Drag a tool onto the canvas; try snapping, pan/zoom, rejection, Escape and Undo/Redo. Same-window placement is supported."),
		{ID: "new-split", Kind: ui.SplitPane, Number: f.split, Grow: true, Accessibility: ui.Accessibility{Name: "Latest features split"}, Change: func(e ui.ChangeEvent) { f.split = e.Number; d.refresh() }, Children: []ui.Node{left, right}},
		{ID: "new-status", Kind: ui.StatusBar, Text: f.message},
	}}
}
