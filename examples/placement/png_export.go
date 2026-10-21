package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"path/filepath"
	"slices"
)

const pngByteLimit = 16 << 20
const pngPixelLimit = 16 << 20

type pngImage struct {
	objects    []canvas.Rect
	scale      int
	background string
}
type pngDraft struct {
	snapshot            selectionSnapshot
	all, selected       []canvas.Rect
	scale               int
	background, message string
	selectionOnly       bool
}

func pngBounds(p pngImage) (canvas.Rect, int, int, error) {
	if len(p.objects) == 0 || len(p.objects) > 512 {
		return canvas.Rect{}, 0, 0, errors.New("PNG requires 1..512 visible components")
	}
	if p.scale != 1 && p.scale != 2 && p.scale != 4 {
		return canvas.Rect{}, 0, 0, errors.New("choose PNG scale 1, 2 or 4")
	}
	if p.background != "transparent" && p.background != "white" && p.background != "dark" {
		return canvas.Rect{}, 0, 0, errors.New("choose a PNG background")
	}
	left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	seen := map[string]bool{}
	for _, o := range p.objects {
		if o.ID == "" || seen[o.ID] || !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) {
			return canvas.Rect{}, 0, 0, errors.New("invalid PNG geometry")
		}
		seen[o.ID] = true
		left = math.Min(left, o.X)
		top = math.Min(top, o.Y)
		right = math.Max(right, o.X+o.Width)
		bottom = math.Max(bottom, o.Y+o.Height)
	}
	width, height := math.Ceil((right-left)*float64(p.scale)), math.Ceil((bottom-top)*float64(p.scale))
	if width < 1 || height < 1 || width > 8192 || height > 8192 || width*height > pngPixelLimit {
		return canvas.Rect{}, 0, 0, errors.New("PNG exceeds 8192 pixels per side or 16 megapixels; use a lower scale or smaller selection")
	}
	return canvas.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}, int(width), int(height), nil
}

type pngBuffer struct{ bytes.Buffer }

func (b *pngBuffer) Write(p []byte) (int, error) {
	if len(p) > pngByteLimit-b.Len() {
		return 0, document.ErrTooLarge
	}
	return b.Buffer.Write(p)
}

// Geometry-only CPU export on the existing detached document worker. No live
// canvas, GPU presentation, editor overlays or mutable owner state is accessed.
func encodePNG(p pngImage) ([]byte, error) {
	bounds, w, h, err := pngBounds(p)
	if err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	if p.background != "transparent" {
		bg := color.NRGBA{255, 255, 255, 255}
		if p.background == "dark" {
			bg = color.NRGBA{32, 35, 40, 255}
		}
		draw.Draw(out, out.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	}
	for _, o := range p.objects {
		x0, y0 := (o.X-bounds.X)*float64(p.scale), (o.Y-bounds.Y)*float64(p.scale)
		x1, y1 := (o.X+o.Width-bounds.X)*float64(p.scale), (o.Y+o.Height-bounds.Y)*float64(p.scale)
		r := image.Rect(int(math.Floor(x0)), int(math.Floor(y0)), int(math.Ceil(x1)), int(math.Ceil(y1)))
		draw.Draw(out, r, image.NewUniform(o.Color), image.Point{}, draw.Over)
	}
	var b pngBuffer
	if err = png.Encode(&b, out); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func (m *model) beginPNG() error {
	if err := m.svgAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	d := &pngDraft{snapshot: m.selectionSnapshot(), all: slices.Clone(m.visibleObjects()), scale: 1, background: "transparent"}
	for _, o := range d.all {
		if slices.Contains(d.snapshot.ids, o.ID) {
			d.selected = append(d.selected, o)
		}
	}
	m.pngDraft = d
	m.prompt = "png-export"
	m.modalEpoch++
	return nil
}
func (m *model) pngSnapshot(d *pngDraft) (pngImage, error) {
	if d == nil || m.pngDraft != d || !d.snapshot.current(m) {
		return pngImage{}, history.ErrStale
	}
	objects := d.all
	if d.selectionOnly {
		objects = d.selected
	}
	p := pngImage{objects: slices.Clone(objects), scale: d.scale, background: d.background}
	_, _, _, err := pngBounds(p)
	return p, err
}
func (m *model) exportPNG(a *app.Application, ref app.WindowRef, store document.Store, p pngImage) error {
	if _, _, _, err := pngBounds(p); err != nil {
		return err
	}
	p.objects = slices.Clone(p.objects)
	codec := document.Codec[pngImage]{Capture: func() (pngImage, error) { return p, nil }, Encode: encodePNG, Decode: func([]byte) (pngImage, error) { return pngImage{}, app.ErrUnavailable }, Install: func(pngImage) error { return app.ErrUnavailable }}
	name := "Drawing"
	if m.files != nil && m.files.State().Path != "" {
		name = drawingSuggestion(filepath.Base(m.files.State().Path))
	}
	return startDetachedSave(m, a, ref, detachedSaveConfig[pngImage]{title: "Export PNG", filename: name, extension: "png", success: "Exported PNG", codec: codec, store: store, maxBytes: pngByteLimit})
}
func (m *model) pngDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.pngDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.pngDraft == d && m.prompt == "png-export"
	}
	cancel := func() {
		if active() {
			m.pngDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	scale, bg, message, dimensions := "1", "transparent", "", ""
	selected, valid := false, false
	if d != nil {
		scale = fmt.Sprint(d.scale)
		bg, message, selected = d.background, d.message, d.selectionOnly
		p, err := m.pngSnapshot(d)
		if err != nil {
			message = err.Error()
		} else {
			_, w, h, _ := pngBounds(p)
			dimensions = fmt.Sprintf("%d × %d pixels", w, h)
			valid = true
		}
	}
	return ui.Node{ID: "png-export-dialog", Kind: ui.Dialog, Text: "Export PNG", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "png-help", Kind: ui.Label, Text: "Fit visible drawing or selected units to their world bounds. One world unit is one pixel at 1×. Grid, camera, selection and handles are omitted. Rectangles round outward to pixel edges. Maximum 8192 per side, 16 megapixels and 16 MiB output.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 490})}},
		{ID: "png-selected", Kind: ui.Checkbox, Text: "Selected components only", Checked: selected, Change: func(e ui.ChangeEvent) {
			if active() {
				d.selectionOnly = e.Checked
				d.message = ""
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "png-scale", Kind: ui.Selector, SelectedID: scale, Options: []ui.Option{{ID: "1", Label: "1×"}, {ID: "2", Label: "2×"}, {ID: "4", Label: "4×"}}, Accessibility: ui.Accessibility{Name: "PNG scale"}, Change: func(e ui.ChangeEvent) {
			if active() {
				switch e.SelectedID {
				case "1":
					d.scale = 1
				case "2":
					d.scale = 2
				case "4":
					d.scale = 4
				default:
					return
				}
				d.message = ""
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "png-background", Kind: ui.Selector, SelectedID: bg, Options: []ui.Option{{ID: "transparent", Label: "Transparent"}, {ID: "white", Label: "White"}, {ID: "dark", Label: "Dark"}}, Accessibility: ui.Accessibility{Name: "PNG background"}, Change: func(e ui.ChangeEvent) {
			if active() {
				switch e.SelectedID {
				case "transparent", "white", "dark":
					d.background = e.SelectedID
				default:
					return
				}
				d.message = ""
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "png-dimensions", Kind: ui.Label, Text: dimensions}, {ID: "png-error", Kind: ui.Label, Text: message},
		{ID: "png-export", Kind: ui.Button, Text: "Choose destination…", Disabled: !valid, Activate: func() {
			if !active() {
				return
			}
			p, err := m.pngSnapshot(d)
			if err != nil {
				d.message = err.Error()
				_ = a.Invalidate("placement")
				return
			}
			m.prompt = ""
			if err = m.exportPNG(a, ref, nil, p); err != nil {
				m.status = err.Error()
			}
			m.pngDraft = nil
			_ = a.Invalidate("placement")
		}},
		{ID: "png-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel}}}
}
