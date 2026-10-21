package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
)

var matchDimensionOptions = []ui.Option{{ID: "match-width", Label: "Match width to primary"}, {ID: "match-height", Label: "Match height to primary"}, {ID: "match-size", Label: "Match size to primary"}}

func (m *model) bindMatchDimensions() error {
	for _, o := range matchDimensionOptions {
		if err := m.commands.Register(command.Command{ID: o.ID, Label: o.Label, Available: func() error {
			if err := m.selectionAvailable(); err != nil {
				return err
			}
			return m.arrangementAvailable(2)
		}, Execute: func(context.Context) error { return m.matchDimensions(o.ID) }}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) matchDimensions(action string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if err := m.arrangementAvailable(2); err != nil {
		return err
	}
	if action != "match-width" && action != "match-height" && action != "match-size" {
		return errors.New("unknown dimension match")
	}
	units, keys := m.selectionUnits()
	primary := keys[m.scene.Selected]
	reference := -1
	for i, u := range units {
		if u.ID == primary {
			reference = i
			break
		}
	}
	if reference < 0 {
		return errNoSelection
	}
	ref := units[reference]
	after := slices.Clone(m.scene.Objects)
	for _, u := range units {
		if u.ID == primary {
			continue
		}
		ids := []string{}
		group := false
		for _, o := range m.scene.Objects {
			if keys[o.ID] == u.ID {
				ids = append(ids, o.ID)
				group = group || m.groups[o.ID] != ""
			}
		}
		target := u
		uniform := group || m.aspectLocked
		switch action {
		case "match-width":
			target.Width = ref.Width
			if uniform {
				target.Height = u.Height * (ref.Width / u.Width)
			}
		case "match-height":
			target.Height = ref.Height
			if uniform {
				target.Width = u.Width * (ref.Height / u.Height)
			}
		case "match-size":
			target.Width, target.Height = ref.Width, ref.Height
			if uniform {
				sx, sy := ref.Width/u.Width, ref.Height/u.Height
				if math.Abs(sx-sy) > 1e-12*math.Max(1, math.Max(sx, sy)) {
					return errors.New("matching both dimensions would change group or locked aspect proportions; match one dimension instead")
				}
			}
		}
		var err error
		after, err = transformedSelection(after, ids, target)
		if err != nil {
			return err
		}
	}
	if err := m.applyObjects("Match selection dimensions", after, m.scene.Selected); err != nil {
		return err
	}
	m.status = "Matched dimensions to primary component/group"
	return nil
}
