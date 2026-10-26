package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"os"
)

func (m *macComponentsGallery) filePickerView() ui.Node {
	if m.filePicker == nil {
		p := &ui.FilePickerField{ID: "file-picker/path", Label: "Reference file", Kind: ui.FilePickerFile, Extensions: []string{"txt", "png", "pdf"}, IconImage: m.artwork[ui.SymbolFolder]}
		p.Request = func(e ui.FilePickerEvent) error {
			if m.filePickerActive() {
				return app.ErrServiceBusy
			}
			if e.Action == ui.FilePickerClear {
				p.Path = ""
				return nil
			}
			if m.essentials == nil || m.essentials.application == nil {
				return fmt.Errorf("Native path services unavailable offscreen")
			}
			a, w := m.essentials.application, m.essentials.window
			ticket, err := a.ReserveWindowCompletion(w, nil)
			if err != nil {
				return err
			}
			m.filePickerTicket = ticket
			ext := append([]string(nil), p.Extensions...)
			go func() {
				value := e.Path
				s, err := a.Paths()
				if err == nil {
					switch e.Action {
					case ui.FilePickerBrowse:
						value, err = s.ChoosePath(w.Context(), w, app.PathOptions{Kind: e.Kind, Title: "Choose a " + string(e.Kind), InitialPath: e.Path, Extensions: ext})
					case ui.FilePickerReveal:
						err = s.RevealPath(w.Context(), e.Path)
					case ui.FilePickerDrop:
						var info os.FileInfo
						info, err = os.Stat(e.Path)
						if err == nil && (e.Kind == ui.FilePickerFolder && !info.IsDir() || e.Kind == ui.FilePickerFile && !info.Mode().IsRegular()) {
							err = fmt.Errorf("Path kind differs")
						}
					}
				}
				_ = ticket.Publish(func() {
					if m.filePickerTicket != ticket || p.Kind != e.Kind {
						return
					}
					m.filePickerTicket = nil
					if err != nil {
						if err != app.ErrCanceled {
							m.status = err.Error()
						}
					} else if e.Action != ui.FilePickerReveal {
						p.Path = value
						m.status = "Path accepted"
						m.count++
					}
					a.Invalidate(w.ID())
				})
			}()
			return nil
		}
		m.filePicker = p
	}
	p := m.filePicker
	p.Busy = m.filePickerActive()
	n, e := p.View()
	if e != nil {
		return ui.Node{ID: "file-picker/error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "file-picker/page", Kind: ui.Column, Children: []ui.Node{{ID: "file-picker/kind", Kind: ui.SegmentedControl, SelectedID: string(p.Kind), Disabled: p.Busy, Options: []ui.Option{{ID: "file", Label: "File"}, {ID: "folder", Label: "Folder"}}, Change: func(e ui.ChangeEvent) {
		if m.filePickerActive() {
			return
		}
		p.Kind = ui.FilePickerKind(e.SelectedID)
		p.Path = ""
		p.Extensions = nil
		if p.Kind == ui.FilePickerFile {
			p.Extensions = []string{"txt", "png", "pdf"}
		}
	}}, n}}
}

func (m *macComponentsGallery) filePickerActive() bool {
	return m.filePickerTicket != nil && m.filePickerTicket.Err() == nil
}
