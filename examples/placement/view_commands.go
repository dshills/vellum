package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"slices"
)

func (m *model) bindViewCommands() error {
	for _, c := range []struct{ id, label string }{{"zoom-in", "Zoom In"}, {"zoom-out", "Zoom Out"}, {"zoom-100", "100%"}, {"fit-selection", "Fit Selection"}, {"fit", "Fit All"}, {"reset", "Reset View"}} {
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Available: func() error {
			if err := m.selectionAvailable(); err != nil {
				return err
			}
			if c.id == "fit-selection" && len(m.selectedIDs()) == 0 {
				return errors.New("select components to fit")
			}
			return nil
		}, Execute: func(context.Context) error { return m.viewCommand(c.id) }}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) viewCommand(action string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	v := m.scene.View
	if err := (&canvas.Scene{View: v}).Validate(); err != nil {
		return err
	}
	switch action {
	case "zoom-100":
		v.Zoom = 1
	case "zoom-in", "zoom-out":
		factor := 1.25
		if action == "zoom-out" {
			factor = 1 / 1.25
		}
		if err := v.ZoomAt(canvas.Point{X: v.Width / 2, Y: v.Height / 2}, factor); err != nil {
			return err
		}
	case "fit-selection", "fit":
		objects := m.visibleObjects()
		if action == "fit-selection" {
			ids := m.selectedIDs()
			objects = slices.DeleteFunc(objects, func(o canvas.Rect) bool { return !slices.Contains(ids, o.ID) })
			if len(objects) == 0 {
				return errors.New("select components to fit")
			}
		}
		if len(objects) > 0 {
			s := canvas.Scene{View: v, Objects: objects}
			s.Fit()
			v = s.View
		}
	case "reset":
		v.Center = canvas.Point{}
		v.Zoom = 1
	default:
		return errors.New("unknown view command")
	}
	if err := (&canvas.Scene{View: v}).Validate(); err != nil {
		return err
	}
	m.scene.View = v
	m.status = fmt.Sprintf("Zoom %.0f%%", v.Zoom*100)
	return nil
}
