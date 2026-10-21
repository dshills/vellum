package main

import (
	"cmp"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"slices"
	"strconv"
)

type spacingDraft struct {
	axis, gap, message string
	scene              *canvas.Scene
	revision           uint64
	before             []canvas.Rect
	ids                []string
	primary            string
}

func (m *model) spacingAvailable() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	return m.arrangementAvailable(2)
}
func (m *model) bindNumericSpacing() error {
	return m.commands.Register(command.Command{ID: "numeric-spacing", Label: "Set exact spacing", Available: m.spacingAvailable, Execute: func(context.Context) error { return m.beginSpacing() }})
}
func (m *model) beginSpacing() error {
	if err := m.spacingAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	m.spacingDraft = &spacingDraft{axis: "horizontal", gap: "20", scene: m.scene, revision: m.journal.Snapshot().Revision, before: slices.Clone(m.scene.Objects), ids: m.selectedIDs(), primary: m.scene.Selected}
	m.prompt = "spacing"
	m.modalEpoch++
	return nil
}
func (m *model) spacedSelection(axis string, gap float64) ([]canvas.Rect, error) {
	if (axis != "horizontal" && axis != "vertical") || !finiteResize(gap) || gap < 0 || gap > 10000 {
		return nil, errors.New("choose horizontal/vertical and a finite gap from 0 to 10000")
	}
	units, keys := m.selectionUnits()
	if len(units) < 2 {
		return nil, errors.New("select at least two components or groups")
	}
	position := func(o canvas.Rect) float64 {
		if axis == "horizontal" {
			return o.X
		}
		return o.Y
	}
	size := func(o canvas.Rect) float64 {
		if axis == "horizontal" {
			return o.Width
		}
		return o.Height
	}
	slices.SortStableFunc(units, func(a, b canvas.Rect) int { return cmp.Compare(position(a), position(b)) })
	cursor := position(units[0])
	after := slices.Clone(m.scene.Objects)
	for _, u := range units {
		delta := cursor - position(u)
		dx, dy := delta, 0.0
		if axis == "vertical" {
			dx, dy = 0, delta
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
			return nil, err
		}
		cursor += size(u) + gap
	}
	return after, nil
}
func (m *model) applySpacing(d *spacingDraft) error {
	if d == nil || m.spacingDraft != d || m.scene != d.scene || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.scene.Objects, d.before) || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		return history.ErrStale
	}
	if err := m.spacingAvailable(); err != nil {
		return err
	}
	gap, err := strconv.ParseFloat(d.gap, 64)
	if err != nil {
		return errors.New("gap must be a finite number from 0 to 10000")
	}
	after, err := m.spacedSelection(d.axis, gap)
	if err != nil {
		return err
	}
	if err = m.applyObjects("Set exact spacing", after, d.primary); err != nil {
		return err
	}
	m.status = "Applied exact " + d.axis + " spacing"
	return nil
}
func (m *model) spacingDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.spacingDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.spacingDraft == d && m.prompt == "spacing"
	}
	cancel := func() {
		if active() {
			m.spacingDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	axis, gap, message := "horizontal", "20", ""
	if d != nil {
		axis, gap, message = d.axis, d.gap, d.message
	}
	return ui.Node{ID: "spacing-dialog", Kind: ui.Dialog, Text: "Exact spacing", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "spacing-help", Kind: ui.Label, Text: "World-unit gap, 0–10000. Units sort left-to-right or top-to-bottom; the first stays fixed. Groups translate rigidly. Snapping is ignored.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 440})}},
		{ID: "spacing-axis", Kind: ui.Selector, Options: []ui.Option{{ID: "horizontal", Label: "Horizontal gap"}, {ID: "vertical", Label: "Vertical gap"}}, SelectedID: axis, Change: func(e ui.ChangeEvent) {
			if active() && (e.SelectedID == "horizontal" || e.SelectedID == "vertical") {
				d.axis = e.SelectedID
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "spacing-gap", Kind: ui.TextField, Text: gap, Accessibility: ui.Accessibility{Name: "Exact gap"}, Change: func(e ui.ChangeEvent) {
			if active() {
				if len(e.Text) <= 64 {
					d.gap = e.Text
				} else {
					d.gap = "invalid"
				}
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "spacing-error", Kind: ui.Label, Text: message},
		{ID: "spacing-apply", Kind: ui.Button, Text: "Apply spacing", Activate: func() {
			if !active() {
				return
			}
			m.prompt = ""
			if err := m.applySpacing(d); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.spacingDraft = nil
				} else {
					d.message = err.Error()
					m.prompt = "spacing"
				}
			} else {
				m.spacingDraft = nil
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "spacing-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}
