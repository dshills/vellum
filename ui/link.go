package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateLinkURL permits bounded absolute HTTP(S) links without credentials or
// control characters. It never opens or fetches the destination.
func ValidateLinkURL(value string) error {
	if len(value) == 0 || len(value) > 4096 || !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 || strings.ContainsAny(value, " \\") {
		return fmt.Errorf("ui: invalid link URL")
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Opaque != "" {
		return fmt.Errorf("ui: link requires an absolute HTTP(S) URL without credentials")
	}
	return nil
}

type LinkEvent struct{ ID, URL string }

// Link is a controlled Mac-style text action. Open and Copy run on the owner;
// dispatch native services from a worker. Contextual Open/Copy use the same live
// validation as keyboard/button activation. Nil actions are disabled individually.
// One retained instance belongs to one application owner. Style overrides defaults.
type Link struct {
	ID, Label, URL string
	Dark, Disabled bool
	IconImage      *canvas.TiledImage
	Open, Copy     func(LinkEvent)
	Style          Style
}

func (l *Link) View() (Node, error) {
	if l == nil || l.ID == "" || strings.TrimSpace(l.Label) == "" || len(l.Label) > 1024 || !utf8.ValidString(l.Label) || strings.ContainsRune(l.Label, 0) {
		return Node{}, fmt.Errorf("ui: link requires an ID and bounded label")
	}
	if err := ValidateLinkURL(l.URL); err != nil {
		return Node{}, err
	}
	id, target := l.ID, l.URL
	invoke := func(copy bool) {
		if l.ID != id || l.URL != target || l.Disabled || ValidateLinkURL(l.URL) != nil {
			return
		}
		fn := l.Open
		if copy {
			fn = l.Copy
		}
		if fn != nil {
			fn(LinkEvent{ID: id, URL: target})
		}
	}
	p := MacPalette(l.Dark, color.NRGBA{})
	style := Style{Padding: Some(Insets{}), Border: Some(Border{}), Background: Some(color.NRGBA{}), Foreground: Some(p.Accent), Typography: Some(MacTypography(13, 400)), GrowX: Some(float32(0))}.Overlay(l.Style)
	n := Node{ID: id, Kind: Button, Text: l.Label, Appearance: ButtonTool, IconImage: l.IconImage, Style: style, Disabled: l.Disabled || (l.Open == nil && l.Copy == nil), Accessibility: Accessibility{Name: l.Label, Help: target}, Tooltip: target, Activate: func() { invoke(false) }}
	n.ContextMenu = &ContextMenu{Options: []Option{{ID: "open", Label: "Open Link", Disabled: l.Open == nil}, {ID: "copy", Label: "Copy Link", Icon: SymbolCopy, Disabled: l.Copy == nil}}, Change: func(e ChangeEvent) {
		if e.SelectedID == "open" {
			invoke(false)
		}
		if e.SelectedID == "copy" {
			invoke(true)
		}
	}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
