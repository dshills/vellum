package main

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

type geometryDraft struct {
	values   [4]string
	message  string
	before   []canvas.Rect
	ids      []string
	primary  string
	revision uint64
	scene    *canvas.Scene
}

// Exact numeric edits use world units and deliberately bypass grid snapping.
// A multi-selection is one affine transform of its complete bounding rectangle.
func transformedSelection(before []canvas.Rect, ids []string, target canvas.Rect) ([]canvas.Rect, error) {
	bounds, ok := selectionBounds(before, ids)
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if !ok || !finite(target.X) || !finite(target.Y) || !finite(target.Width) || !finite(target.Height) || target.Width <= 0 || target.Height <= 0 {
		return nil, errors.New("enter finite X/Y and positive width/height")
	}
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	sx, sy := target.Width/bounds.Width, target.Height/bounds.Height
	after := slices.Clone(before)
	for i, o := range after {
		if wanted[o.ID] {
			if sx == 1 {
				o.X += target.X - bounds.X
			} else {
				o.X = target.X + (o.X-bounds.X)*sx
			}
			if sy == 1 {
				o.Y += target.Y - bounds.Y
			} else {
				o.Y = target.Y + (o.Y-bounds.Y)*sy
			}
			o.Width *= sx
			o.Height *= sy
			if len(ids) == 1 {
				o.X = target.X
				o.Y = target.Y
				o.Width = target.Width
				o.Height = target.Height
			} else {
				roundLimit := func(v float64) float64 {
					if v < minComponentSize && v >= math.Nextafter(minComponentSize, math.Inf(-1)) {
						return minComponentSize
					}
					if v > maxComponentSize && v <= math.Nextafter(maxComponentSize, math.Inf(1)) {
						return maxComponentSize
					}
					return v
				}
				o.Width = roundLimit(o.Width)
				o.Height = roundLimit(o.Height)
			}
			if !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) {
				return nil, errors.New("every component must keep origins within ±10000 and dimensions within 8–2000")
			}
			after[i] = o
		}
	}
	return after, nil
}
func (m *model) beginGeometry() error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	ids := m.selectedIDs()
	bounds, ok := selectionBounds(m.scene.Objects, ids)
	if !ok {
		return errNoSelection
	}
	d := &geometryDraft{before: slices.Clone(m.scene.Objects), ids: ids, primary: m.scene.Selected, revision: m.journal.Snapshot().Revision, scene: m.scene}
	for i, v := range []float64{bounds.X, bounds.Y, bounds.Width, bounds.Height} {
		d.values[i] = strconv.FormatFloat(v, 'g', -1, 64)
	}
	m.geometry = d
	m.prompt = "geometry"
	return nil
}
func (m *model) applyGeometry(d *geometryDraft) error {
	if d == nil || m.geometry != d || m.scene != d.scene || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.scene.Objects, d.before) || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		return history.ErrStale
	}
	if err := m.editAvailable(); err != nil {
		return err
	}
	values := [4]float64{}
	for i, s := range d.values {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("%s must be a finite number", []string{"X", "Y", "Width", "Height"}[i])
		}
		values[i] = v
	}
	after, err := transformedSelection(d.before, d.ids, canvas.Rect{X: values[0], Y: values[1], Width: values[2], Height: values[3]})
	if err != nil {
		return err
	}
	if err = m.applyObjects("Edit selection geometry", after, d.primary); err != nil {
		return err
	}
	m.status = fmt.Sprintf("Updated geometry of %d components", len(d.ids))
	return nil
}
func (m *model) geometryDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.geometry
	active := func() bool {
		return d != nil && m.geometry == d && m.prompt == "geometry" && ref.Context().Err() == nil
	}
	cancel := func() {
		if active() {
			m.geometry = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	children := []ui.Node{{ID: "geometry-help", Kind: ui.Label, Text: "World units; exact values ignore grid snapping. Groups scale together.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 380})}}}
	for i, label := range []string{"X", "Y", "Width", "Height"} {
		text := ""
		if d != nil {
			text = d.values[i]
		}
		children = append(children, ui.Node{ID: "geometry-row-" + label, Kind: ui.Row, Children: []ui.Node{{ID: "geometry-label-" + label, Kind: ui.Label, Text: label, MinWidth: 60}, {ID: "geometry-" + label, Kind: ui.TextField, Text: text, MinWidth: 180, Accessibility: ui.Accessibility{Name: "Selection " + label}, Change: func(e ui.ChangeEvent) {
			if active() {
				d.values[i] = e.Text
			}
		}}}})
	}
	message := ""
	if d != nil {
		message = d.message
	}
	children = append(children, ui.Node{ID: "geometry-error", Kind: ui.Label, Text: message})
	children = append(children, ui.Node{ID: "geometry-apply", Kind: ui.Button, Text: "Apply geometry", Activate: func() {
		if !active() {
			return
		}
		m.prompt = ""
		if err := m.applyGeometry(d); err != nil {
			m.status = err.Error()
			if errors.Is(err, history.ErrStale) {
				m.geometry = nil
			} else {
				d.message = err.Error()
				m.prompt = "geometry"
			}
		} else {
			m.geometry = nil
		}
		_ = a.Invalidate("placement")
	}}, ui.Node{ID: "geometry-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel})
	return ui.Node{ID: "geometry-dialog", Kind: ui.Dialog, Text: "Selection geometry", Open: active(), Dismiss: cancel, Children: children}
}
