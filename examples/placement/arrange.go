package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
)

var arrangementActions = []struct {
	id, label string
	minimum   int
}{
	{"align-left", "Left edges", 2}, {"align-hcenter", "H centers", 2}, {"align-right", "Right edges", 2},
	{"align-top", "Top edges", 2}, {"align-vcenter", "V centers", 2}, {"align-bottom", "Bottom edges", 2},
	{"distribute-h", "H gaps", 3}, {"distribute-v", "V gaps", 3},
}

func (m *model) arrangementAvailable(minimum int) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	units, _ := m.selectionUnits()
	if len(units) < minimum {
		return fmt.Errorf("select at least %d components or groups", minimum)
	}
	return nil
}
func (m *model) bindArrangement() error {
	for _, action := range arrangementActions {
		if err := m.commands.Register(command.Command{ID: action.id, Label: action.label, Available: func() error { return m.arrangementAvailable(action.minimum) }, Execute: func(context.Context) error { return m.arrangeSelected(action.id) }}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) arrangeSelected(action string) error {
	for _, a := range arrangementActions {
		if a.id == action {
			if err := m.arrangementAvailable(a.minimum); err != nil {
				return err
			}
			after, err := m.arrangedSelection(action)
			if err != nil {
				return err
			}
			if err = m.applyObjects(a.label, after, m.scene.Selected); err != nil {
				return err
			}
			m.status = "Arranged selection: " + a.label
			return nil
		}
	}
	return errors.New("unknown arrangement")
}

func (m *model) arrangedSelection(action string) ([]canvas.Rect, error) {
	units, keys := m.selectionUnits()
	ids := make([]string, 0, len(units))
	for _, u := range units {
		ids = append(ids, u.ID)
	}
	arranged, err := arrangedObjects(units, ids, action)
	if err != nil {
		return nil, err
	}
	deltas := make(map[string]canvas.Point, len(units))
	for i, u := range units {
		deltas[u.ID] = canvas.Point{X: arranged[i].X - u.X, Y: arranged[i].Y - u.Y}
	}
	after := slices.Clone(m.scene.Objects)
	for i, o := range after {
		if unit := keys[o.ID]; unit != "" {
			d := deltas[unit]
			if d.X != 0 {
				after[i].X += d.X
			}
			if d.Y != 0 {
				after[i].Y += d.Y
			}
			if !validPosition(after[i].X, after[i].Y) {
				return nil, errors.New("arrangement would exceed member origin limits")
			}
		}
	}
	return after, nil
}

// Work in world coordinates without snapping: alignment/gap equalities must
// remain exact under any camera. The scene's paint order is never reordered.
func arrangedObjects(before []canvas.Rect, ids []string, action string) ([]canvas.Rect, error) {
	b, ok := selectionBounds(before, ids)
	if !ok {
		return nil, errNoSelection
	}
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	indices := make([]int, 0, len(ids))
	after := slices.Clone(before)
	for i, o := range before {
		if wanted[o.ID] {
			indices = append(indices, i)
		}
	}
	if action == "distribute-h" || action == "distribute-v" {
		if len(indices) < 3 {
			return nil, errors.New("distribution requires three components")
		}
		horizontal := action == "distribute-h"
		position := func(o canvas.Rect) float64 {
			if horizontal {
				return o.X
			}
			return o.Y
		}
		size := func(o canvas.Rect) float64 {
			if horizontal {
				return o.Width
			}
			return o.Height
		}
		slices.SortStableFunc(indices, func(i, j int) int { return cmp.Compare(position(before[i]), position(before[j])) })
		start, span := b.X, b.Width
		if !horizontal {
			start, span = b.Y, b.Height
		}
		total := 0.0
		for _, i := range indices {
			total += size(before[i])
		}
		gap := (span - total) / float64(len(indices)-1)
		if gap < -1e-9 {
			return nil, errors.New("not enough space for non-overlapping equal gaps")
		}
		// Ignore sub-nanounit noise when a fractional-size set fits tightly.
		if gap < 0 {
			gap = 0
		}
		prefix := 0.0
		for j, i := range indices {
			cursor := start + prefix + float64(j)*gap
			if j == len(indices)-1 {
				cursor = start + span - size(before[i])
			}
			if horizontal {
				after[i].X = cursor
			} else {
				after[i].Y = cursor
			}
			prefix += size(before[i])
		}
	} else {
		for _, i := range indices {
			switch action {
			case "align-left":
				after[i].X = b.X
			case "align-hcenter":
				after[i].X = b.X + (b.Width-before[i].Width)/2
			case "align-right":
				after[i].X = b.X + b.Width - before[i].Width
			case "align-top":
				after[i].Y = b.Y
			case "align-vcenter":
				after[i].Y = b.Y + (b.Height-before[i].Height)/2
			case "align-bottom":
				after[i].Y = b.Y + b.Height - before[i].Height
			default:
				return nil, errors.New("unknown arrangement")
			}
		}
	}
	for _, i := range indices {
		if !validPosition(after[i].X, after[i].Y) {
			return nil, errors.New("arrangement would exceed drawing origin limits")
		}
	}
	return after, nil
}
