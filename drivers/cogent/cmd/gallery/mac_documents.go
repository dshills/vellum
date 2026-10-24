package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"math"
	"time"
)

type macDocumentGallery struct {
	parent       *macComponentsGallery
	list         *ui.DocumentList
	outline      *ui.OutlineView
	colorWell    *ui.ColorWell
	colorPreview color.NRGBA
	nativeBusy   bool
	nativeEpoch  uint64
	colorArtwork *canvas.TiledImage
}

func (m *macComponentsGallery) documentListView() ui.Node {
	if m.documents == nil {
		m.documents = &macDocumentGallery{parent: m}
	}
	e := m.documents
	if e.list == nil {
		e.list = &ui.DocumentList{ID: "documents/files", Label: "Drawing documents", Columns: ui.DefaultDocumentColumns(), Sort: ui.DocumentSort{ColumnID: "name"}, Items: []ui.DocumentItem{
			{ID: "projects", Name: "Projects", Kind: "Folder", Folder: true, Image: m.artwork[ui.SymbolFolder]},
			{ID: "floor", Name: "Floor plan.vellum", Kind: "Drawing", SizeBytes: 18432, Modified: time.Date(2026, 10, 24, 9, 0, 0, 0, time.Local), Image: m.artwork[ui.SymbolDocument]},
			{ID: "details", Name: "Details.vellum", Kind: "Drawing", SizeBytes: 8192, Modified: time.Date(2026, 10, 23, 12, 0, 0, 0, time.Local), Image: m.artwork[ui.SymbolSave]},
			{ID: "archive", Name: "Archive (unavailable)", Kind: "Folder", Folder: true, Disabled: true, Image: m.artwork[ui.SymbolFolder]},
		}}
		e.list.Change = func(event ui.ListSelectionEvent) {
			e.list.SelectedIDs = event.Apply(e.list.SelectedIDs)
			m.status = fmt.Sprint("Document selection: ", e.list.SelectedIDs)
		}
		e.list.SortChange = func(event ui.DocumentSortEvent) { e.list.Sort = event.Sort }
		e.list.Resize = func(event ui.DocumentResizeEvent) {
			for i := range e.list.Columns {
				if e.list.Columns[i].ID == event.ColumnID {
					e.list.Columns[i].Width = event.Width
				}
			}
		}
		e.list.Activate = func(event ui.DocumentEvent) { m.count++; m.status = "Open document: " + event.ItemID }
	}
	n, err := e.list.View()
	if err != nil {
		return ui.Node{ID: "documents/list-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "documents/list-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "documents/list-help", Kind: ui.Label, Text: "Finder-style document metadata. Click a header to sort. Drag its divider or focus the divider and use Left/Right. Command/Shift select; Return opens the sample only."}, n}}
}

func (m *macComponentsGallery) outlineView() ui.Node {
	if m.documents == nil {
		m.documents = &macDocumentGallery{parent: m}
	}
	e := m.documents
	if e.outline == nil {
		e.outline = &ui.OutlineView{ID: "documents/outline", Label: "Project outline", ExpandedIDs: []string{"project", "drawings"}, Items: []ui.OutlineItem{
			{ID: "project", Label: "My project", Badge: "4", Image: m.artwork[ui.SymbolFolder]},
			{ID: "drawings", Parent: "project", Label: "Drawings", Image: m.artwork[ui.SymbolFolder]},
			{ID: "floor", Parent: "drawings", Label: "Floor plan", Image: m.artwork[ui.SymbolDocument]},
			{ID: "detail", Parent: "drawings", Label: "Details", Image: m.artwork[ui.SymbolSave]},
			{ID: "components", Parent: "project", Label: "Components", Image: m.artwork[ui.SymbolFolder]},
			{ID: "block", Parent: "components", Label: "Block", Image: m.artwork[ui.SymbolDocument]},
			{ID: "archive", Parent: "project", Label: "Archive (unavailable)", Disabled: true, Image: m.artwork[ui.SymbolFolder]},
		}}
		e.outline.Change = func(event ui.ListSelectionEvent) {
			e.outline.SelectedIDs = event.Apply(e.outline.SelectedIDs)
			m.status = fmt.Sprint("Outline selection: ", e.outline.SelectedIDs)
		}
		e.outline.Disclosure = func(event ui.OutlineDisclosureEvent) { e.outline.ExpandedIDs = event.Apply(e.outline.ExpandedIDs) }
		e.outline.Activate = func(event ui.OutlineEvent) { m.count++; m.status = "Open outline item: " + event.ItemID }
	}
	n, err := e.outline.View()
	if err != nil {
		return ui.Node{ID: "documents/outline-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "documents/outline-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "documents/outline-help", Kind: ui.Label, Text: "Expand folders with their triangles or Right. Left collapses or moves to the parent. Command/Shift selects; Return opens a leaf. Hidden selections survive collapse."}, n}}
}

func documentColorArtwork() *canvas.TiledImage {
	pixels := image.NewNRGBA(image.Rect(0, 0, 48, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			dx, dy := float64(x)-23.5, float64(y)-23.5
			if math.Hypot(dx, dy) > 22 {
				continue
			}
			h := (math.Atan2(dy, dx) + math.Pi) / (math.Pi * 2) * 6
			sector := int(h)
			f := h - float64(sector)
			a, b := uint8(255*(1-f)), uint8(255*f)
			var c color.NRGBA
			switch sector % 6 {
			case 0:
				c = color.NRGBA{255, b, 0, 255}
			case 1:
				c = color.NRGBA{a, 255, 0, 255}
			case 2:
				c = color.NRGBA{0, 255, b, 255}
			case 3:
				c = color.NRGBA{0, a, 255, 255}
			case 4:
				c = color.NRGBA{b, 0, 255, 255}
			case 5:
				c = color.NRGBA{255, 0, a, 255}
			}
			pixels.SetNRGBA(x, y, c)
		}
	}
	art, err := canvas.NewTiledImage(pixels)
	if err != nil {
		panic(err)
	}
	return art
}
func (m *macComponentsGallery) nativeColorView() ui.Node {
	if m.documents == nil {
		m.documents = &macDocumentGallery{parent: m}
	}
	e := m.documents
	if e.colorWell == nil {
		e.colorWell = &ui.ColorWell{ID: "documents/color-well", Label: "Drawing color", Color: color.NRGBA{R: 220, G: 70, B: 50, A: 255}}
		e.colorWell.Change = func(event ui.ColorEdit) {
			e.colorWell.Color = event.Apply(e.colorWell.Color)
			e.colorPreview = e.colorWell.Color
		}
		e.colorPreview = e.colorWell.Color
		e.colorArtwork = documentColorArtwork()
	}
	e.colorWell.Disabled = e.nativeBusy
	e.colorWell.PopoverStyle = ui.Style{Background: ui.Some(m.theme().Palette.Value.Popup)}
	well, err := e.colorWell.View()
	if err != nil {
		return ui.Node{ID: "documents/color-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "documents/native-color-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "documents/native-color-help", Kind: ui.Label, Text: "Open the actual macOS Colors panel. Live changes preview below; Use Color accepts and Cancel/Escape restores the original. The managed color well remains independently available."}, well, {ID: "documents/color-preview", Kind: ui.FormGroup, Text: fmt.Sprintf("Live preview: #%02X%02X%02X · alpha %d", e.colorPreview.R, e.colorPreview.G, e.colorPreview.B, e.colorPreview.A), Style: ui.Style{Background: ui.Some(color.NRGBA{R: e.colorPreview.R, G: e.colorPreview.G, B: e.colorPreview.B, A: 255}), MinSize: ui.Some(ui.Size{Height: 70})}}, {ID: "documents/choose-native-color", Kind: ui.Button, Text: "Show macOS Colors…", IconImage: e.colorArtwork, Disabled: e.nativeBusy, Activate: e.chooseColor}}}
}
func (e *macDocumentGallery) chooseColor() {
	m := e.parent
	a, window := m.essentials.application, m.essentials.window
	if e.nativeBusy {
		return
	}
	if a == nil {
		m.status = "Native Colors unavailable offscreen"
		return
	}
	ticket, err := a.ReserveWindowCompletion(window, nil)
	if err != nil {
		m.status = err.Error()
		return
	}
	e.nativeBusy = true
	e.nativeEpoch++
	epoch := e.nativeEpoch
	accepted := e.colorWell.Color
	go func() {
		var value color.NRGBA
		service, failure := a.ColorPanels()
		if failure == nil {
			value, failure = service.ChooseColor(window.Context(), window, app.ColorPanelOptions{Title: "Drawing Colors", Color: accepted, ShowsAlpha: true, Preview: func(c color.NRGBA) {
				_ = a.PostWindow(window, func() {
					if e.nativeBusy && e.nativeEpoch == epoch {
						e.colorPreview = c
						_ = a.Invalidate(window.ID())
					}
				})
			}})
		}
		_ = ticket.Publish(func() {
			if e.nativeEpoch != epoch {
				return
			}
			e.nativeBusy = false
			if failure != nil {
				e.colorPreview = accepted
				m.status = fmt.Sprint("Native Colors: ", failure)
			} else {
				e.colorWell.Color = value
				e.colorPreview = value
				m.count++
				m.status = "Native color accepted"
			}
			_ = a.Invalidate(window.ID())
		})
	}()
}
