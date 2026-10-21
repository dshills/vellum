package main

import (
	"context"
	"errors"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func primaryAlignOptions() []ui.Option {
	return []ui.Option{
		{ID: "align-primary-left", Label: "Left edges to primary"},
		{ID: "align-primary-hcenter", Label: "H centers to primary"},
		{ID: "align-primary-right", Label: "Right edges to primary"},
		{ID: "align-primary-top", Label: "Top edges to primary"},
		{ID: "align-primary-vcenter", Label: "V centers to primary"},
		{ID: "align-primary-bottom", Label: "Bottom edges to primary"},
	}
}
func (m *model) primaryAlignmentAvailable() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	return m.arrangementAvailable(2)
}
func (m *model) bindPrimaryAlignment() error {
	for _, o := range primaryAlignOptions() {
		if err := m.commands.Register(command.Command{ID: o.ID, Label: o.Label, Available: m.primaryAlignmentAvailable, Execute: func(context.Context) error { return m.alignToPrimary(o.ID) }}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) alignToPrimary(action string) error {
	if err := m.primaryAlignmentAvailable(); err != nil {
		return err
	}
	options := primaryAlignOptions()
	choice := slices.IndexFunc(options, func(o ui.Option) bool { return o.ID == action })
	if choice < 0 {
		return errors.New("unknown primary alignment")
	}
	units, keys := m.selectionUnits()
	primary := keys[m.scene.Selected]
	reference := slices.IndexFunc(units, func(u canvas.Rect) bool { return u.ID == primary })
	if reference < 0 {
		return errNoSelection
	}
	ref := units[reference]
	after := slices.Clone(m.scene.Objects)
	for _, u := range units {
		if u.ID == primary {
			continue
		}
		dx, dy := 0.0, 0.0
		switch action {
		case "align-primary-left":
			dx = ref.X - u.X
		case "align-primary-hcenter":
			dx = ref.X + (ref.Width-u.Width)/2 - u.X
		case "align-primary-right":
			dx = ref.X + ref.Width - u.Width - u.X
		case "align-primary-top":
			dy = ref.Y - u.Y
		case "align-primary-vcenter":
			dy = ref.Y + (ref.Height-u.Height)/2 - u.Y
		case "align-primary-bottom":
			dy = ref.Y + ref.Height - u.Height - u.Y
		}
		ids := []string{}
		for _, o := range m.scene.Objects {
			if keys[o.ID] == u.ID {
				ids = append(ids, o.ID)
			}
		}
		var err error
		after, err = movedSelection(after, ids, dx, dy)
		if err != nil {
			return err
		}
	}
	if err := m.applyObjects("Align to primary", after, m.scene.Selected); err != nil {
		return err
	}
	m.status = "Aligned selection: " + options[choice].Label
	return nil
}
