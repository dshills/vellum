package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"path/filepath"
	"strings"
)

type ImageWellAction string

const (
	ImageWellChoose ImageWellAction = "choose"
	ImageWellClear  ImageWellAction = "clear"
	ImageWellDrop   ImageWellAction = "drop"
)

type ImageWellEvent struct {
	ID     string
	Action ImageWellAction
	Path   string
}

// ImageWell displays immutable pixels and requests choose/replace/clear or one
// PNG/JPEG path. Request runs on owner and must not read files or block. Decode
// on a worker and assign Image/Status/Busy to accept results. Retain on one owner.
// Drop acceptance is framework propagation; the OS currently acknowledges file
// drags before model admission. Status provides application-owned result feedback.
type ImageWell struct {
	ID, Label, Status            string
	Image                        *canvas.TiledImage
	PlaceholderImage, ChooseIcon *canvas.TiledImage
	Disabled, Busy               bool
	Size                         float32
	Request                      func(ImageWellEvent) bool
	Style                        Style
}

func ImageWellTargetID(id string) string { return id + "/image-well/target" }
func (w *ImageWell) View() (Node, error) {
	if w == nil || !validTokenText(w.ID) || !validTokenText(w.Label) || len(w.Status) > 4096 {
		return Node{}, fmt.Errorf("ui: image well requires bounded ID, label and status")
	}
	size := w.Size
	if size == 0 {
		size = 144
	}
	if !finite(size) || size < 64 || size > 512 {
		return Node{}, fmt.Errorf("ui: image well size must be 64..512")
	}
	for _, im := range []*canvas.TiledImage{w.Image, w.PlaceholderImage, w.ChooseIcon} {
		if im != nil && im.TileCount() == 0 {
			return Node{}, fmt.Errorf("ui: empty image well pixels")
		}
	}
	id := w.ID
	live := func() bool { return w.ID == id && !w.Disabled && !w.Busy && w.Request != nil }
	request := func(action ImageWellAction, path string) bool {
		if !live() || action == ImageWellClear && w.Image == nil {
			return false
		}
		return w.Request(ImageWellEvent{ID: id, Action: action, Path: path})
	}
	preview := Node{ID: id + "/image-well/empty", Kind: Icon, Icon: SymbolDocument, Text: "Drop a PNG or JPEG", Style: Style{MinSize: Some(Size{Width: size, Height: size}), MaxSize: Some(Size{Width: size, Height: size})}}
	image := w.Image
	if image == nil {
		image = w.PlaceholderImage
	}
	if image != nil {
		preview = Node{ID: id + "/image-well/preview", Kind: Image, Image: image, Text: w.Label, Fit: ImageContain, Style: preview.Style}
	}
	target := Node{ID: ImageWellTargetID(id), Kind: FormGroup, Style: Style{Padding: Some(UniformInsets(6)), GrowX: Some(float32(0))}, Children: []Node{preview}, FileDrop: func(e FileDropEvent) bool {
		paths, err := CopyFileDropPaths(e.Paths)
		if err != nil || len(paths) != 1 {
			return false
		}
		ext := strings.ToLower(filepath.Ext(paths[0]))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			return false
		}
		return request(ImageWellDrop, paths[0])
	}}
	choose := "Choose image…"
	if w.Image != nil {
		choose = "Replace image…"
	}
	actions := Node{ID: id + "/image-well/actions", Kind: Row, Children: []Node{{ID: id + "/image-well/choose", Kind: Button, Text: choose, IconImage: w.ChooseIcon, Activate: func() { request(ImageWellChoose, "") }}, {ID: id + "/image-well/clear", Kind: Button, Text: "Clear", Disabled: w.Image == nil, Activate: func() { request(ImageWellClear, "") }}}}
	status := w.Status
	if w.Busy {
		status = "Loading image…"
	}
	if status == "" {
		status = "Drop one PNG or JPEG, or choose an image."
	}
	n := Node{ID: id, Kind: Column, Disabled: !live(), Style: Style{Gap: Some(float32(8)), GrowX: Some(float32(1))}.Overlay(w.Style), Accessibility: Accessibility{Name: w.Label}, Children: []Node{{ID: id + "/image-well/label", Kind: Label, Text: w.Label}, target, actions, {ID: id + "/image-well/status", Kind: Label, Text: status}}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
