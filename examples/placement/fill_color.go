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
	m.fillDraft = &fillDraft{value: value, scene: m.scene, revision: m.journal.Snapshot().Revision, before: slices.Clone(m.scene.Objects), ids: ids, primary: m.scene.Selected}
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
	return ui.Node{ID: "fill-dialog", Kind: ui.Dialog, Text: "Selection fill color", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "fill-help", Kind: ui.Label, Text: "Opaque #RRGGBB, or default to restore each type. Blank means mixed colors. Applies to all selected group members in one Undo.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 420})}},
		{ID: "fill-value", Kind: ui.TextField, Text: value, Accessibility: ui.Accessibility{Name: "Fill color"}, Change: func(e ui.ChangeEvent) {
			if active() {
				if len(e.Text) <= 32 {
					d.value = e.Text
				} else {
					d.value = "invalid"
				}
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
