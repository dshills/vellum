package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func (m *macComponentsGallery) pdfViewerView() ui.Node {
	p := ui.PDFViewer{ID: "pdf/document", Label: "PDF document", Path: m.pdfPath, ShowThumbnails: true, Dark: m.dark, IconImage: m.artwork[ui.SymbolDocument]}
	viewer, e := p.View()
	if e != nil {
		return ui.Node{ID: "pdf/error", Kind: ui.Label, Text: e.Error()}
	}
	choose := ui.Node{ID: "pdf/open", Kind: ui.Button, Text: "Open PDF…", IconImage: m.artwork[ui.SymbolFolder], Disabled: m.pdfTicket != nil && m.pdfTicket.Err() == nil, Activate: func() {
		if m.pdfTicket != nil && m.pdfTicket.Err() == nil || m.essentials == nil || m.essentials.application == nil {
			return
		}
		a, w := m.essentials.application, m.essentials.window
		t, e := a.ReserveWindowCompletion(w, nil)
		if e != nil {
			m.status = e.Error()
			return
		}
		m.pdfTicket = t
		go func() {
			s, e := a.Paths()
			var path string
			if e == nil {
				path, e = s.ChoosePath(w.Context(), w, app.PathOptions{Kind: ui.FilePickerFile, Title: "Open PDF", Extensions: []string{"pdf"}})
			}
			_ = t.Publish(func() {
				if m.pdfTicket != t {
					return
				}
				m.pdfTicket = nil
				if e == nil {
					m.pdfPath = path
				} else if e != app.ErrCanceled {
					m.status = e.Error()
				}
				a.Invalidate(w.ID())
			})
		}()
	}}
	return ui.Node{ID: "pdf/page", Kind: ui.Column, Grow: true, Children: []ui.Node{choose, viewer}}
}
