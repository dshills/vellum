package main

import (
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func (m *macComponentsGallery) fileDropZoneView() ui.Node {
	if m.dropZone == nil {
		z := &ui.FileDropZone{ID: "drop-zone/files", Label: "Add reference files", Extensions: []string{"txt", "png"}, MaxFiles: 4, IconImage: m.artwork[ui.SymbolFolder], Change: func(e ui.FileDropZoneEvent) error {
			m.count++
			m.status = fmt.Sprintf("Accepted %d file reference(s); no content read", len(e.Paths))
			return nil
		}}
		z.Choose = func() {
			if m.essentials == nil || m.essentials.application == nil {
				m.status = "Native chooser unavailable offscreen"
				return
			}
			a, w := m.essentials.application, m.essentials.window
			ticket, err := a.ReserveWindowCompletion(w, nil)
			if err != nil {
				m.status = err.Error()
				return
			}
			m.dropTicket = ticket
			ext := append([]string(nil), z.Extensions...)
			go func() {
				s, err := a.Services()
				path := ""
				if err == nil {
					path, err = s.ChooseFile(w.Context(), app.FileOptions{Title: "Choose a reference file", Extensions: ext})
				}
				_ = ticket.Publish(func() {
					if m.dropTicket != ticket {
						return
					}
					m.dropTicket = nil
					m.applyFileChoice(z, path, err)
					a.Invalidate(w.ID())
				})
			}()
		}
		m.dropZone = z
	}
	m.dropZone.Dark = m.dark
	m.dropZone.ChooseDisabled = m.dropTicket != nil && m.dropTicket.Err() == nil
	n, e := m.dropZone.View()
	if e != nil {
		return ui.Node{ID: "drop-zone/error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "drop-zone/page", Kind: ui.Column, Children: []ui.Node{{ID: "drop-zone/help", Kind: ui.Label, Text: "Drop up to four .txt/.png files or use Choose File. The demo accepts references only and never reads their contents. Rejections show below the target. Finder's drag cursor is controlled by the existing backend and cannot preview extension rejection."}, n}}
}

func (m *macComponentsGallery) applyFileChoice(z *ui.FileDropZone, path string, err error) {
	if errors.Is(err, app.ErrCanceled) || (err == nil && path == "") {
		m.status = "File choice cancelled"
		return
	}
	if err != nil {
		m.status = err.Error()
		return
	}
	if e := z.ProposeFiles([]string{path}); e != nil {
		m.status = e.Error()
	}
}
