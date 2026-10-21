package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"strconv"
)

type overlapDraft struct {
	snapshot              selectionSnapshot
	x, y, choice, message string
}

func (m *model) bindOverlapPicker() error {
	return m.commands.Register(command.Command{ID: "pick-overlap", Label: "Pick overlapping component…", Available: m.selectionAvailable, Execute: func(context.Context) error { return m.beginOverlap(m.scene.View.Center) }})
}
func (m *model) beginOverlapAt(g placement.Geometry, p canvas.Point) error {
	world, err := g.World(p)
	if err != nil {
		return err
	}
	return m.beginOverlap(world)
}
func (m *model) beginOverlap(point canvas.Point) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	d := &overlapDraft{snapshot: m.selectionSnapshot(), x: strconv.FormatFloat(point.X, 'f', -1, 64), y: strconv.FormatFloat(point.Y, 'f', -1, 64)}
	m.overlapDraft = d
	m.refreshOverlap(d)
	m.prompt = "overlap-picker"
	m.modalEpoch++
	return nil
}
func (m *model) overlapOptions(d *overlapDraft) ([]ui.Option, error) {
	x, e := strconv.ParseFloat(d.x, 64)
	if e != nil || math.IsNaN(x) || math.IsInf(x, 0) || math.Abs(x) > 1e12 {
		return nil, errors.New("enter finite world X within ±1e12")
	}
	y, e := strconv.ParseFloat(d.y, 64)
	if e != nil || math.IsNaN(y) || math.IsInf(y, 0) || math.Abs(y) > 1e12 {
		return nil, errors.New("enter finite world Y within ±1e12")
	}
	options := []ui.Option{}
	for i := len(m.scene.Objects) - 1; i >= 0; i-- {
		o := m.scene.Objects[i]
		if m.layerHidden(o.ID) || m.layerLocked(o.ID) || x < o.X || x > o.X+o.Width || y < o.Y || y > o.Y+o.Height {
			continue
		}
		label := m.layerLabel(o.ID)
		if group := m.groups[o.ID]; group != "" {
			label += " • group " + m.layerLabel(group)
		}
		options = append(options, ui.Option{ID: o.ID, Label: fmt.Sprintf("%d. %s", len(options)+1, label)})
	}
	return options, nil
}
func (m *model) refreshOverlap(d *overlapDraft) {
	options, err := m.overlapOptions(d)
	d.message = ""
	if err != nil {
		d.choice = ""
		d.message = err.Error()
		return
	}
	if len(options) == 0 {
		d.choice = ""
		d.message = "No visible, unlocked component at this point"
		return
	}
	if !slices.ContainsFunc(options, func(o ui.Option) bool { return o.ID == d.choice }) {
		d.choice = options[0].ID
	}
}
func (m *model) applyOverlap(d *overlapDraft) error {
	if d == nil || m.overlapDraft != d || !d.snapshot.current(m) {
		return history.ErrStale
	}
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	options, err := m.overlapOptions(d)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(options, func(o ui.Option) bool { return o.ID == d.choice }) {
		return errors.New("choose a component at this point")
	}
	m.setSelection([]string{d.choice}, d.choice)
	m.layerFocus = d.choice
	m.status = "Selected " + m.layerLabel(d.choice)
	return nil
}
func (m *model) overlapDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.overlapDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.overlapDraft == d && m.prompt == "overlap-picker"
	}
	cancel := func() {
		if active() {
			m.overlapDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	x, y, choice, message := "", "", "", ""
	options := []ui.Option{}
	if d != nil {
		x, y, choice, message = d.x, d.y, d.choice, d.message
		options, _ = m.overlapOptions(d)
	}
	field := func(id, label, value string, isX bool) ui.Node {
		return ui.Node{ID: id, Kind: ui.TextField, Text: value, Accessibility: ui.Accessibility{Name: label}, Change: func(e ui.ChangeEvent) {
			if active() {
				if len(e.Text) > 64 {
					d.message = "coordinate text is limited to 64 bytes"
					return
				}
				if isX {
					d.x = e.Text
				} else {
					d.y = e.Text
				}
				m.refreshOverlap(d)
				_ = a.Invalidate("placement")
			}
		}}
	}
	return ui.Node{ID: "overlap-dialog", Kind: ui.Dialog, Text: "Pick overlapping component", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "overlap-help", Kind: ui.Label, Text: "Option-click a canvas point, or enter world X/Y. Frontmost first. Choose a component to select its whole group. Selection adds no Undo entry.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 460})}}, field("overlap-x", "World X", x, true), field("overlap-y", "World Y", y, false),
		{ID: "overlap-choice", Kind: ui.Selector, Options: options, SelectedID: choice, Accessibility: ui.Accessibility{Name: "Components at point"}, Change: func(e ui.ChangeEvent) {
			if active() && slices.ContainsFunc(options, func(o ui.Option) bool { return o.ID == e.SelectedID }) {
				d.choice = e.SelectedID
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "overlap-message", Kind: ui.Label, Text: message},
		{ID: "overlap-select", Kind: ui.Button, Text: "Select", Disabled: message != "" || len(options) == 0, Activate: func() {
			if !active() || d.message != "" {
				return
			}
			m.prompt = ""
			if err := m.applyOverlap(d); err != nil {
				m.status = err.Error()
			}
			m.overlapDraft = nil
			_ = a.Invalidate("placement")
		}},
		{ID: "overlap-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel}}}
}
