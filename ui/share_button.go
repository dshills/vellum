package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
)

type ShareEvent struct{ ID string }

// ShareButton proposes sharing on the owner. Use Application.Sharing from a
// worker with this ID as AnchorID. Payload and result belong to the application.
type ShareButton struct {
	ID, Label string
	Disabled  bool
	IconImage *canvas.TiledImage
	Share     func(ShareEvent)
	Style     Style
}

func (s *ShareButton) View() (Node, error) {
	if s == nil || s.ID == "" {
		return Node{}, fmt.Errorf("ui: share button requires an ID")
	}
	id := s.ID
	label := s.Label
	if label == "" {
		label = "Share…"
	}
	n := Node{ID: id, Kind: Button, Text: label, IconImage: s.IconImage, Style: s.Style, Disabled: s.Disabled || s.Share == nil, Accessibility: Accessibility{Name: label, Help: "Choose a macOS sharing service"}, Activate: func() {
		if s.ID == id && !s.Disabled && s.Share != nil {
			s.Share(ShareEvent{ID: id})
		}
	}}
	if n.IconImage == nil {
		n.Icon = SymbolShare
	}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
