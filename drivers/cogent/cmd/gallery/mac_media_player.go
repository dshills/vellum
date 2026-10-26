package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func (m *macComponentsGallery) mediaPlayerView() ui.Node {
	if m.mediaPlayer == nil {
		m.mediaPlayer = &ui.MediaPlayer{ID: "media/player", Label: "Audio and video", IconImage: m.artwork[ui.SymbolDocument], PosterImage: m.artwork[ui.SymbolFolder]}
	}
	p := m.mediaPlayer
	p.Path = m.mediaPath
	p.Dark = m.dark
	viewer, e := p.View()
	if e != nil {
		return ui.Node{ID: "media/error", Kind: ui.Label, Text: e.Error()}
	}
	choose := ui.Node{ID: "media/open", Kind: ui.Button, Text: "Open audio or video…", IconImage: m.artwork[ui.SymbolFolder], Disabled: m.mediaTicket != nil && m.mediaTicket.Err() == nil, Activate: func() {
		if m.mediaTicket != nil && m.mediaTicket.Err() == nil || m.essentials == nil || m.essentials.application == nil {
			return
		}
		a, w := m.essentials.application, m.essentials.window
		t, e := a.ReserveWindowCompletion(w, nil)
		if e != nil {
			m.status = e.Error()
			return
		}
		m.mediaTicket = t
		go func() {
			s, e := a.Paths()
			var path string
			if e == nil {
				path, e = s.ChoosePath(w.Context(), w, app.PathOptions{Kind: ui.FilePickerFile, Title: "Open local media", Extensions: []string{"mp4", "mov", "m4v", "mp3", "m4a", "wav", "aiff"}})
			}
			_ = t.Publish(func() {
				if m.mediaTicket != t {
					return
				}
				m.mediaTicket = nil
				if e == nil {
					m.mediaPath = path
				} else if e != app.ErrCanceled {
					m.status = e.Error()
				}
				a.Invalidate(w.ID())
			})
		}()
	}}
	return ui.Node{ID: "media/page", Kind: ui.Column, Grow: true, Children: []ui.Node{choose, viewer}}
}
