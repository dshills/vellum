package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// NativeContent is a driver-owned embedded view emitted by PDFViewer.View.
// Unsupported drivers show an availability placeholder. Native controls own
// their transient selection/zoom; applications own the accepted source path.
const NativeContent Kind = "native-content"

// NativeContentInput is detached presentation metadata, not a service handle.
// Construct it using a content widget's View method. Revision reloads the same path.
type NativeContentInput struct {
	Type         string `json:"type"`
	Path         string `json:"path"`
	Revision     uint64 `json:"revision"`
	InitialMuted bool   `json:"initial_muted"`
	PosterPNG    string `json:"poster_png,omitempty"`
	Thumbnails   bool   `json:"thumbnails"`
	Dark         bool   `json:"dark"`
	generated    bool
}

func validateNativeContent(n Node) error {
	c := n.NativeContent
	if c == nil {
		if n.Kind == NativeContent {
			return fmt.Errorf("ui: missing native content metadata")
		}
		return nil
	}
	if !validateNativePoster(c.PosterPNG) {
		return fmt.Errorf("ui: invalid native poster")
	}
	if n.Kind != NativeContent || !c.generated || (c.Type != "pdf" && c.Type != "media") || len(c.Path) > 4096 || !utf8.ValidString(c.Path) || strings.ContainsRune(c.Path, 0) || !filepath.IsAbs(c.Path) {
		return fmt.Errorf("ui: invalid native content source")
	}
	return nil
}

// PDFViewer embeds the platform PDF viewer for one explicit local PDF file.
// macOS uses PDFKit with page, thumbnail, zoom, find and selection controls.
// Reads are limited to 64 MiB and documents to 1024 pages; PDF decoding has OS
// resource costs beyond these admission limits. Empty Path displays a prompt.
// Owner-only. Hiding/removal releases the native document and transient state.
type PDFViewer struct {
	ID, Label, Path                string
	Revision                       uint64
	ShowThumbnails, Dark, Disabled bool
	IconImage                      *canvas.TiledImage
	Style                          Style
}

func (p *PDFViewer) View() (Node, error) {
	if p == nil || p.ID == "" || !paletteText(p.Label, 256, false) {
		return Node{}, fmt.Errorf("ui: invalid PDF viewer")
	}
	header := nativeContentHeader(p.ID, p.Label, p.IconImage)
	body := Node{ID: p.ID + "/content", Kind: Label, Text: "Choose a local PDF document.", MinHeight: 320, Grow: true}
	if p.Path != "" {
		body = Node{ID: p.ID + "/content", Kind: NativeContent, Text: "PDFKit viewer — available in a native macOS window", MinWidth: 280, MinHeight: 320, Grow: true, Disabled: p.Disabled, Accessibility: Accessibility{Name: p.Label}, NativeContent: &NativeContentInput{Type: "pdf", Path: p.Path, Revision: p.Revision, Thumbnails: p.ShowThumbnails, Dark: p.Dark, generated: true}}
	}
	n := Node{ID: p.ID, Kind: Column, Children: []Node{header, body}, Grow: true, Style: p.Style}
	return n, Validate(n)
}

// A content title is a label with decorative artwork, not a disabled action.
func nativeContentHeader(id, label string, icon *canvas.TiledImage) Node {
	row := Node{ID: id + "/label", Kind: Row, Style: Style{Gap: Some(float32(8)), Align: Some(AlignCenter)}}
	if icon != nil {
		row.Children = append(row.Children, Node{ID: id + "/label/icon", Kind: Image, Image: icon, Fit: ImageContain, Style: Style{MinSize: Some(Size{Width: 24, Height: 24}), MaxSize: Some(Size{Width: 24, Height: 24})}})
	}
	row.Children = append(row.Children, Node{ID: id + "/label/text", Kind: Label, Text: label, Style: Style{Typography: Some(MacTypography(13, 600))}})
	return row
}
