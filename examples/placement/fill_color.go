package main

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

// Empty wire values retain the component's original default; custom fills are
// opaque RGB, which the existing frozen canvas and SVG APIs already support.
func componentFill(kind, value string) (color.RGBA, error) {
	c, ok := componentByKind(kind)
	if !ok {
		return color.RGBA{}, errors.New("unknown component type")
	}
	if value == "" {
		return c.Color, nil
	}
	if len(value) != 7 || value[0] != '#' {
		return color.RGBA{}, errors.New("enter an opaque color as #RRGGBB")
	}
	n, err := strconv.ParseUint(value[1:], 16, 24)
	if err != nil {
		return color.RGBA{}, errors.New("enter an opaque color as #RRGGBB")
	}
	return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}, nil
}
func savedFill(kind string, value color.RGBA) string {
	c, ok := componentByKind(kind)
	if ok && c.Color == value {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", value.R, value.G, value.B)
}

type fillDraft struct {
	value, message string
	picker         ui.ColorPicker
	pickerInvalid  bool
	scene          *canvas.Scene
	revision       uint64
	before         []canvas.Rect
	ids            []string
	primary        string
}

func (m *model) bindFillColor() error {
	return m.commands.Register(command.Command{ID: "fill-color", Label: "Edit fill color", Available: m.editAvailable, Execute: func(context.Context) error { return m.beginFillColor() }})
}
func (m *model) beginFillColor() error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	ids := m.selectedIDs()
	value := ""
	mixed := false
	for _, o := range m.scene.Objects {
		if slices.Contains(ids, o.ID) {
			v := fmt.Sprintf("#%02x%02x%02x", o.Color.R, o.Color.G, o.Color.B)
			if value != "" && value != v {
				mixed = true
			}
			value = v
		}
	}
	if mixed {
		value = ""
	}
	index := m.selectedIndex()
	if index < 0 {
		return errNoSelection
	}
	primary := m.scene.Objects[index].Color
	m.fillDraft = &fillDraft{value: value, scene: m.scene, revision: m.journal.Snapshot().Revision, before: slices.Clone(m.scene.Objects), ids: ids, primary: m.scene.Selected}
	m.fillDraft.picker = ui.ColorPicker{ID: "fill-picker", Label: "Fill", Color: color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 255}}
	m.prompt = "fill-color"
	m.modalEpoch++
	return nil
}
func (m *model) applyFillColor(d *fillDraft) error {
	if d == nil || m.fillDraft != d || m.scene != d.scene || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.scene.Objects, d.before) || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		return history.ErrStale
	}
	if err := m.editAvailable(); err != nil {
		return err
	}
	if d.pickerInvalid {
		return ui.ErrColorChannel
	}
	value := strings.TrimSpace(d.value)
	if value == "" {
		return errors.New("enter #RRGGBB, or default to restore each component type")
	}
	after := slices.Clone(m.scene.Objects)
	for i, o := range after {
		if slices.Contains(d.ids, o.ID) {
			kind, _, _ := strings.Cut(o.ID, "-")
			wire := value
			if value == "default" {
				wire = ""
			}
			c, err := componentFill(kind, wire)
			if err != nil {
				return err
			}
			after[i].Color = c
		}
	}
	// Keep the serialized document within the same bounds before any history edit.
	copyModel := *m
	scene := *m.scene
	scene.Objects = after
	copyModel.scene = &scene
	f, err := drawingCodec(&copyModel).Capture()
	if err != nil {
		return err
	}
	if _, err = drawingCodec(&copyModel).Encode(f); err != nil {
		return err
	}
	if err = m.applyObjects("Change fill color", after, d.primary); err != nil {
		return err
	}
	m.status = "Updated selection fill color"
	return nil
}
func (m *model) fillColorDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.fillDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.fillDraft == d && m.prompt == "fill-color"
	}
	cancel := func() {
		if active() {
			m.fillDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	value, message := "", ""
	if d != nil {
		value, message = d.value, d.message
	}
	picker, swatches := m.fillPickerNodes(a, ref, d, active)
	return ui.Node{ID: "fill-dialog", Kind: ui.Dialog, Text: "Selection fill color", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "fill-help", Kind: ui.Label, Text: "Choose a swatch or adjust RGB. Blank hex means mixed fills; preview starts with the primary. Default previews the primary but restores each type. Apply changes all selected members in one Undo.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 740})}},
		picker, swatches,
		{ID: "fill-value", Kind: ui.TextField, Text: value, Accessibility: ui.Accessibility{Name: "Fill color"}, Change: func(e ui.ChangeEvent) {
			if active() {
				if len(e.Text) <= 32 {
					d.value = e.Text
				} else {
					d.value = "invalid"
				}
				d.pickerInvalid = false
				d.message = ""
				kind, _, _ := strings.Cut(d.primary, "-")
				wire := strings.TrimSpace(d.value)
				if wire == "default" {
					wire = ""
				}
				if c, err := componentFill(kind, wire); err == nil && strings.TrimSpace(d.value) != "" {
					d.picker.Color = color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
				}
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "fill-error", Kind: ui.Label, Text: message},
		{ID: "fill-apply", Kind: ui.Button, Text: "Apply fill", Activate: func() {
			if !active() {
				return
			}
			m.prompt = ""
			if err := m.applyFillColor(d); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.fillDraft = nil
				} else {
					d.message = err.Error()
					m.prompt = "fill-color"
				}
			} else {
				m.fillDraft = nil
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "fill-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}

// The app supports opaque RGB fills. Compose the existing public picker while
// keeping its alpha/hex inputs out of this dialog; the bounded legacy hex field
// also accepts the app-specific "default" action.
func (m *model) fillPickerNodes(a *app.Application, ref app.WindowRef, d *fillDraft, active func() bool) (ui.Node, ui.Node) {
	if d == nil {
		return ui.Node{ID: "fill-picker-empty", Kind: ui.Label}, ui.Node{ID: "fill-swatches-empty", Kind: ui.Label}
	}
	choose := func(c color.NRGBA) {
		if active() {
			c.A = 255
			d.picker.Color = c
			d.value = fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
			d.pickerInvalid = false
			d.message = ""
			_ = a.Invalidate("placement")
		}
	}
	d.picker.Change = func(e ui.ColorEdit) { choose(e.Apply(d.picker.Color)) }
	d.picker.Invalid = func(err error) {
		if active() {
			d.pickerInvalid = true
			d.message = err.Error()
			_ = a.Invalidate("placement")
		}
	}
	picker := d.picker.View()
	picker.Children = slices.DeleteFunc(picker.Children, func(n ui.Node) bool {
		return fillNodeContains(n, ui.ColorFieldID(d.picker.ID, ui.ColorAlpha)) || fillNodeContains(n, ui.ColorFieldID(d.picker.ID, ui.ColorHex))
	})
	row := ui.Node{ID: "fill-swatches", Kind: ui.Row}
	for _, s := range []struct {
		name string
		c    color.NRGBA
	}{{"Blue", color.NRGBA{55, 120, 225, 255}}, {"Purple", color.NRGBA{160, 90, 204, 255}}, {"Red", color.NRGBA{210, 55, 65, 255}}, {"Green", color.NRGBA{45, 150, 85, 255}}, {"Gold", color.NRGBA{220, 170, 35, 255}}, {"Gray", color.NRGBA{100, 110, 120, 255}}} {
		c := s.c
		row.Children = append(row.Children, ui.Node{ID: "fill-swatch-" + strings.ToLower(s.name), Kind: ui.Button, Text: s.name, Accessibility: ui.Accessibility{Name: "Fill swatch " + s.name}, Style: ui.Style{Background: ui.Some(c), Foreground: ui.Some(color.NRGBA{255, 255, 255, 255})}, Activate: func() { choose(c) }})
	}
	row.Children = append(row.Children, ui.Node{ID: "fill-default", Kind: ui.Button, Text: "Default", Activate: func() {
		if active() {
			kind, _, _ := strings.Cut(d.primary, "-")
			c, _ := componentFill(kind, "")
			d.picker.Color = color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
			d.value = "default"
			d.pickerInvalid = false
			d.message = ""
			_ = a.Invalidate("placement")
		}
	}})
	return picker, row
}
func fillNodeContains(n ui.Node, id string) bool {
	if n.ID == id {
		return true
	}
	for _, c := range n.Children {
		if fillNodeContains(c, id) {
			return true
		}
	}
	return false
}
