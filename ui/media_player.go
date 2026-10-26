package ui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/png"
)

// MediaPlayer embeds native transport, seek and volume controls for an explicit
// local audio/video file. macOS uses AVKit. New sources start paused; settings
// marked Initial apply only when Path/Revision creates a new native player.
// The source remains a mutable path: AVFoundation reopens it after admission;
// applications needing immutable bytes must provide their own stable snapshot.
// Owner-only. Native playback state is transient and released on node retirement.
type MediaPlayer struct {
	ID, Label, Path              string
	Revision                     uint64
	InitialMuted, Dark, Disabled bool
	IconImage                    *canvas.TiledImage
	// PosterImage is immutable canvas artwork and preserves original colors. At most 256x256 pixels; it is shown
	// over the paused initial frame until playback advances. Nil uses no poster.
	PosterImage  *canvas.TiledImage
	Style        Style
	posterSource *canvas.TiledImage
	posterPNG    string
}

func (p *MediaPlayer) View() (Node, error) {
	if p == nil || p.ID == "" || !paletteText(p.Label, 256, false) {
		return Node{}, fmt.Errorf("ui: invalid media player")
	}
	if p.PosterImage != p.posterSource {
		if p.PosterImage == nil {
			p.posterPNG = ""
		} else {
			size := p.PosterImage.Size()
			if size.X < 1 || size.Y < 1 || size.X > 256 || size.Y > 256 {
				return Node{}, fmt.Errorf("ui: media poster must be at most 256x256")
			}
			var data bytes.Buffer
			if e := png.Encode(&data, p.PosterImage); e != nil {
				return Node{}, e
			}
			if data.Len() > 256<<10 {
				return Node{}, fmt.Errorf("ui: media poster exceeds 256 KiB")
			}
			p.posterPNG = base64.StdEncoding.EncodeToString(data.Bytes())
		}
		p.posterSource = p.PosterImage
	}
	header := nativeContentHeader(p.ID, p.Label, p.IconImage)
	body := Node{ID: p.ID + "/content", Kind: Label, Text: "Choose a local audio or video file.", MinHeight: 300, Grow: true}
	if p.Path != "" {
		body = Node{ID: p.ID + "/content", Kind: NativeContent, Text: "AVKit player — available in a native macOS window", MinWidth: 280, MinHeight: 300, Grow: true, Disabled: p.Disabled, Accessibility: Accessibility{Name: p.Label}, NativeContent: &NativeContentInput{Type: "media", Path: p.Path, Revision: p.Revision, InitialMuted: p.InitialMuted, Dark: p.Dark, PosterPNG: p.posterPNG, generated: true}}
	}
	n := Node{ID: p.ID, Kind: Column, Grow: true, Style: p.Style, Children: []Node{header, body}}
	return n, Validate(n)
}

func validateNativePoster(value string) bool {
	if value == "" {
		return true
	}
	if len(value) > 350000 {
		return false
	}
	data, e := base64.StdEncoding.DecodeString(value)
	if e != nil || len(data) > 256<<10 {
		return false
	}
	c, e := png.DecodeConfig(bytes.NewReader(data))
	return e == nil && c.Width > 0 && c.Height > 0 && c.Width <= 256 && c.Height <= 256
}
