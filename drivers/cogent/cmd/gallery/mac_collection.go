package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

type macCollectionGallery struct {
	systemTheme ui.Theme

	collection       *ui.CollectionView
	reject, disabled bool
	requests, opens  int
	lastOpen         string
	dark             bool
}

func newMacCollectionGallery() *macCollectionGallery {
	m := &macCollectionGallery{}
	m.collection = &ui.CollectionView{ID: "mac-collection", Label: "Sample documents", SelectedIDs: []string{"notes"}, Items: []ui.CollectionItem{{ID: "projects", Label: "Projects", Icon: ui.SymbolFolder}, {ID: "notes", Label: "Design notes", Icon: ui.SymbolDocument}, {ID: "thumbnail", Label: "Color study", Thumbnail: collectionDemoThumbnail()}, {ID: "drawing", Label: "Floor plan", Icon: ui.SymbolDocument}, {ID: "archive", Label: "Archive", Icon: ui.SymbolFolder}, {ID: "locked", Label: "Unavailable", Icon: ui.SymbolDocument, Disabled: true}, {ID: "reference", Label: "Reference", Icon: ui.SymbolFolder}, {ID: "proposal", Label: "Proposal", Icon: ui.SymbolDocument}}}
	m.collection.Change = func(e ui.ListSelectionEvent) {
		m.requests++
		if !m.reject {
			m.collection.SelectedIDs = e.Apply(m.collection.SelectedIDs)
		}
	}
	m.collection.Activate = func(e ui.CollectionEvent) { m.opens++; m.lastOpen = e.ItemID }
	return m
}
func collectionDemoThumbnail() *canvas.TiledImage {
	im := image.NewNRGBA(image.Rect(0, 0, 64, 48))
	for y := range 48 {
		for x := range 64 {
			v := color.NRGBA{R: 245, G: 195, B: 92, A: 255}
			if x >= 32 {
				v = color.NRGBA{R: 67, G: 145, B: 196, A: 255}
			}
			if y > 30 {
				v = color.NRGBA{R: 91, G: 161, B: 105, A: 255}
			}
			im.SetNRGBA(x, y, v)
		}
	}
	tiled, _ := canvas.NewTiledImage(im)
	return tiled
}
func (m *macCollectionGallery) view() ui.Node {
	m.collection.Disabled = m.disabled
	grid, err := m.collection.View()
	if err != nil {
		return ui.Node{ID: "collection-error", Kind: ui.Label, Text: err.Error()}
	}
	theme := ui.MacTheme(m.dark)
	if m.systemTheme.Palette.Set {
		theme = m.systemTheme
	}
	grid.Style = theme.Kinds[ui.ListBox].Overlay(ui.Style{GrowX: ui.Some(float32(1)), MaxSize: ui.Some(ui.Size{Height: 360})})
	return ui.Node{ID: "mac-collection-page", Kind: ui.Column, Grow: true, Style: ui.Style{Gap: ui.Some(float32(12)), Padding: ui.Some(ui.Insets{Top: 16, Left: 16, Right: 16, Bottom: 16})}, Children: []ui.Node{
		{ID: "collection-title", Kind: ui.Label, Text: "Documents", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}},
		{ID: "collection-help", Kind: ui.Label, Text: "Resize the window to change columns. Command-click toggles, Shift-click extends, arrows navigate, Return or double-click opens."},
		grid,
		{ID: "collection-status", Kind: ui.Label, Text: fmt.Sprintf("Selected: %s · selection requests: %d · opened: %s (%d)", strings.Join(m.collection.SelectedIDs, ", "), m.requests, m.lastOpen, m.opens)},
		{ID: "collection-reject", Kind: ui.Checkbox, Text: "Reject selection requests", Checked: m.reject, Change: func(e ui.ChangeEvent) { m.reject = e.Checked }},
		{ID: "collection-disabled", Kind: ui.Checkbox, Text: "Disable collection", Checked: m.disabled, Change: func(e ui.ChangeEvent) { m.disabled = e.Checked }},
	}}
}
