package main

import (
	"context"
	"github.com/dshills/vellum/ui"
)

// The callback is owner-side; actual OS/clipboard work runs on a worker with an
// exact-window completion ticket, as in the existing native panel examples.
func (m *macComponentsGallery) linkAction(e ui.LinkEvent, copy bool) {
	a, w := m.essentials.application, m.essentials.window
	if a == nil {
		m.status = "Native links unavailable offscreen"
		return
	}
	ticket, err := a.ReserveWindowCompletion(w, nil)
	if err != nil {
		m.status = err.Error()
		return
	}
	go func() {
		var err error
		if copy {
			var s interface {
				WriteClipboard(context.Context, string) error
			}
			s, err = a.Services()
			if err == nil {
				err = s.WriteClipboard(w.Context(), e.URL)
			}
		} else {
			s, serviceErr := a.URLs()
			err = serviceErr
			if err == nil {
				err = s.OpenURL(w.Context(), e.URL)
			}
		}
		_ = ticket.Publish(func() {
			if err != nil {
				m.status = err.Error()
			} else {
				m.status = "Link action accepted"
				m.count++
			}
			a.Invalidate(w.ID())
		})
	}()
}
func (m *macComponentsGallery) linksView() ui.Node {
	l := &ui.Link{ID: "links/project", Label: "Vellum on GitHub", URL: "https://github.com/dshills/vellum", Dark: m.dark, IconImage: m.artwork[ui.SymbolDocument], Open: func(e ui.LinkEvent) { m.linkAction(e, false) }, Copy: func(e ui.LinkEvent) { m.linkAction(e, true) }}
	n, e := l.View()
	if e != nil {
		return ui.Node{ID: "links/error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "links/page", Kind: ui.Column, Children: []ui.Node{{ID: "links/help", Kind: ui.Label, Text: "Tab to the link and press Return/Space to open it. Right-click or Shift-F10 offers Open Link and Copy Link. The destination also appears as a help tag."}, n}}
}
