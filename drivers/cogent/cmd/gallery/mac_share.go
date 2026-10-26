package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func (m *macComponentsGallery) sharingView() ui.Node {
	b := &ui.ShareButton{ID: "share/demo", Disabled: m.shareActive(), Share: func(e ui.ShareEvent) {
		if m.shareActive() {
			return
		}
		if m.essentials == nil {
			m.status = "Native Share unavailable before gallery binding"
			return
		}
		a, w := m.essentials.application, m.essentials.window
		if a == nil {
			m.status = "Native Share unavailable offscreen"
			return
		}
		ticket, err := a.ReserveWindowCompletion(w, nil)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.shareTicket = ticket
		go func() {
			service, err := a.Sharing()
			var result app.ShareResult
			if err == nil {
				result, err = service.Share(w.Context(), w, app.ShareOptions{AnchorID: e.ID, Items: []app.ShareItem{{Kind: app.ShareText, Value: "Synthetic Vellum sharing demonstration"}, {Kind: app.ShareURL, Value: "https://github.com/dshills/vellum"}}})
			}
			_ = ticket.Publish(func() {
				if m.shareTicket != ticket {
					return
				}
				m.shareTicket = nil
				if err != nil {
					m.status = err.Error()
				} else {
					m.status = "Selected sharing service: " + result.Service
					m.count++
				}
				a.Invalidate(w.ID())
			})
		}()
	}}
	n, e := b.View()
	if e != nil {
		return ui.Node{ID: "share/error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "share/page", Kind: ui.Column, Children: []ui.Node{{ID: "share/help", Kind: ui.Label, Text: "Uses the real macOS Share picker and system service icons. This sample offers only synthetic text and the project URL. Choosing a service may transmit that content; Escape dismisses the picker. Selection is not a delivery receipt."}, n}}
}

// A revoked completion is terminal even when its owner callback cannot run.
func (m *macComponentsGallery) shareActive() bool {
	return m.shareTicket != nil && m.shareTicket.Err() == nil
}
