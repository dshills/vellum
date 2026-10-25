package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

type macPreviewGallery struct {
	parent  *macComponentsGallery
	list    *ui.DocumentList
	outline *ui.OutlineView
	busy    bool
}

func (m *macComponentsGallery) quickLookView() ui.Node {
	if m.preview == nil {
		m.preview = &macPreviewGallery{parent: m}
	}
	p := m.preview
	if p.list == nil {
		p.list = &ui.DocumentList{ID: "preview/files", Label: "Quick Look samples", Sort: ui.DocumentSort{ColumnID: "name"}, Items: []ui.DocumentItem{{ID: "note", Name: "Preview notes.txt", Kind: "Text", Image: m.artwork[ui.SymbolDocument]}, {ID: "art", Name: "Color artwork.png", Kind: "Image", Image: m.artwork[ui.SymbolSave]}}}
		p.list.Change = func(e ui.ListSelectionEvent) { p.list.SelectedIDs = e.Apply(p.list.SelectedIDs) }
		p.list.Preview = func(e ui.DocumentEvent) { p.open(e.ItemID) }
		p.list.Activate = p.list.Preview
		p.outline = &ui.OutlineView{ID: "preview/outline", Label: "Preview outline", ExpandedIDs: []string{"samples"}, Items: []ui.OutlineItem{{ID: "samples", Label: "Sample files", Image: m.artwork[ui.SymbolFolder]}, {ID: "note", Parent: "samples", Label: "Preview notes.txt", Image: m.artwork[ui.SymbolDocument]}, {ID: "art", Parent: "samples", Label: "Color artwork.png", Image: m.artwork[ui.SymbolSave]}}}
		p.outline.Change = func(e ui.ListSelectionEvent) { p.outline.SelectedIDs = e.Apply(p.outline.SelectedIDs) }
		p.outline.Disclosure = func(e ui.OutlineDisclosureEvent) { p.outline.ExpandedIDs = e.Apply(p.outline.ExpandedIDs) }
		p.outline.Preview = func(e ui.OutlineEvent) {
			if e.ItemID != "samples" {
				p.open(e.ItemID)
			}
		}
	}
	p.list.Disabled = p.busy
	p.outline.Disabled = p.busy
	list, e := p.list.View()
	if e != nil {
		panic(e)
	}
	outline, e := p.outline.View()
	if e != nil {
		panic(e)
	}
	return ui.Node{ID: "preview/page", Kind: ui.Column, Children: []ui.Node{{ID: "preview/help", Kind: ui.Label, Text: "Select a sample and press Space for actual macOS Quick Look. Escape/Space closes it. The samples are temporary text and full-color PNG files; no personal files are read."}, list, {ID: "preview/open", Kind: ui.Button, Text: "Quick Look sample", IconImage: m.artwork[ui.SymbolDocument], Disabled: p.busy, Activate: func() {
		id := "note"
		if len(p.list.SelectedIDs) > 0 {
			id = p.list.SelectedIDs[0]
		}
		p.open(id)
	}}, outline}}
}
func (p *macPreviewGallery) open(id string) {
	m := p.parent
	if p.busy || id != "note" && id != "art" {
		return
	}
	a, w := m.essentials.application, m.essentials.window
	if a == nil {
		m.status = "Quick Look unavailable offscreen"
		return
	}
	ticket, err := a.ReserveWindowCompletion(w, nil)
	if err != nil {
		m.status = err.Error()
		return
	}
	p.busy = true
	go func() {
		err := p.previewFiles(a, w, id)
		_ = ticket.Publish(func() {
			p.busy = false
			if err != nil {
				m.status = fmt.Sprint("Quick Look: ", err)
			} else {
				m.count++
				m.status = "Quick Look closed"
			}
			_ = a.Invalidate(w.ID())
		})
	}()
}
func (p *macPreviewGallery) previewFiles(a *app.Application, w app.WindowRef, id string) error {
	dir, err := os.MkdirTemp("", "vellum-quick-look-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err = os.WriteFile(filepath.Join(dir, "Preview notes.txt"), []byte("Vellum Quick Look\n\nNative macOS previews through public Go APIs.\n"), 0600); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "Color artwork.png"))
	if err != nil {
		return err
	}
	err = png.Encode(f, documentPreviewPixels())
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	service, err := a.QuickLook()
	if err != nil {
		return err
	}
	return service.PreviewFiles(w.Context(), w, app.QuickLookOptions{SelectedID: id, Items: []app.QuickLookItem{{ID: "note", Path: filepath.Join(dir, "Preview notes.txt"), Title: "Preview notes"}, {ID: "art", Path: filepath.Join(dir, "Color artwork.png"), Title: "Full-color artwork"}}})
}

func documentPreviewPixels() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 320, 180))
	for y := 0; y < 180; y++ {
		for x := 0; x < 320; x++ {
			im.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 255 / 319), G: uint8(y * 255 / 179), B: 180, A: 255})
		}
	}
	return im
}
