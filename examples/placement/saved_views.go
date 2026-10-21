package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"strings"
	"unicode/utf8"
)

type savedView struct {
	Name   string       `json:"name"`
	Center canvas.Point `json:"center"`
	Zoom   float64      `json:"zoom"`
}
type savedViewsDraft struct {
	snapshot              selectionSnapshot
	before                []savedView
	camera                savedView
	name, choice, message string
	invalid               bool
}

func validSavedViews(views []savedView) error {
	if len(views) > 16 {
		return errors.New("at most 16 saved views per drawing")
	}
	seen := map[string]bool{}
	for _, v := range views {
		if v.Name == "" || !validLayerName(v.Name) || seen[v.Name] || math.IsNaN(v.Center.X) || math.IsNaN(v.Center.Y) || math.IsInf(v.Center.X, 0) || math.IsInf(v.Center.Y, 0) || math.Abs(v.Center.X) > 1e12 || math.Abs(v.Center.Y) > 1e12 || math.IsNaN(v.Zoom) || v.Zoom < .01 || v.Zoom > 100 {
			return errors.New("invalid or duplicate saved view")
		}
		seen[v.Name] = true
	}
	return nil
}
func (m *model) bindSavedViews() error {
	return m.commands.Register(command.Command{ID: "saved-views", Label: "Saved views…", Available: m.selectionAvailable, Execute: func(context.Context) error { return m.beginSavedViews() }})
}
func (m *model) beginSavedViews() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	d := &savedViewsDraft{snapshot: m.selectionSnapshot(), before: slices.Clone(m.savedViews), camera: savedView{Center: m.scene.View.Center, Zoom: m.scene.View.Zoom}, name: "View"}
	if len(d.before) > 0 {
		d.choice = d.before[0].Name
	}
	m.viewsDraft = d
	m.prompt = "saved-views"
	m.modalEpoch++
	return nil
}
func (m *model) applySavedViews(d *savedViewsDraft, action string) error {
	if d == nil || m.viewsDraft != d || !d.snapshot.current(m) || !slices.Equal(m.savedViews, d.before) {
		return history.ErrStale
	}
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if action == "restore" {
		for _, v := range d.before {
			if v.Name == d.choice {
				m.scene.View.Center, m.scene.View.Zoom = v.Center, v.Zoom
				m.status = "Restored view " + v.Name
				return nil
			}
		}
		return errors.New("choose a saved view")
	}
	after := slices.Clone(d.before)
	label := ""
	switch action {
	case "save":
		if d.invalid {
			return errors.New("view name exceeded 128 bytes; edit it before saving")
		}
		name := strings.TrimSpace(d.name)
		if name == "" {
			return errors.New("enter a view name")
		}
		v := d.camera
		v.Name = name
		idx := slices.IndexFunc(after, func(v savedView) bool { return v.Name == name })
		if idx >= 0 {
			after[idx] = v
		} else {
			after = append(after, v)
		}
		label = "Save view " + name
	case "remove":
		idx := slices.IndexFunc(after, func(v savedView) bool { return v.Name == d.choice })
		if idx < 0 {
			return errors.New("choose a saved view")
		}
		after = slices.Delete(after, idx, idx+1)
		label = "Remove view " + d.choice
	default:
		return errors.New("unknown saved view action")
	}
	if err := validSavedViews(after); err != nil {
		return err
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		return err
	}
	f.Version = 6
	f.Views = after
	if _, err = drawingCodec(m).Encode(f); err != nil {
		return err
	}
	if slices.Equal(after, d.before) {
		return nil
	}
	before := slices.Clone(d.before)
	err = m.journal.Apply(history.Change{Label: label, Apply: func() error {
		if !slices.Equal(m.savedViews, before) {
			return history.ErrStale
		}
		m.savedViews = slices.Clone(after)
		return nil
	}, Revert: func() error {
		if !slices.Equal(m.savedViews, after) {
			return history.ErrStale
		}
		m.savedViews = slices.Clone(before)
		return nil
	}})
	if err == nil {
		m.status = label
	}
	return err
}
func (m *model) savedViewsDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.viewsDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.viewsDraft == d && m.prompt == "saved-views"
	}
	cancel := func() {
		if active() {
			m.viewsDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	options := []ui.Option{}
	choice, name, message, position := "", "", "", ""
	if d != nil {
		choice, name, message = d.choice, d.name, d.message
		position = fmt.Sprintf("Current view: X %.3f, Y %.3f, %.1f%%", d.camera.Center.X, d.camera.Center.Y, d.camera.Zoom*100)
		for _, v := range d.before {
			options = append(options, ui.Option{ID: v.Name, Label: v.Name})
		}
	}
	action := func(id string) func() {
		return func() {
			if !active() {
				return
			}
			m.prompt = ""
			if err := m.applySavedViews(d, id); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.viewsDraft = nil
				} else {
					d.message = err.Error()
					m.prompt = "saved-views"
				}
			} else {
				m.viewsDraft = nil
			}
			_ = a.Invalidate("placement")
		}
	}
	return ui.Node{ID: "saved-views-dialog", Kind: ui.Dialog, Text: "Saved views", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "views-help", Kind: ui.Label, Text: "Save the current pan/zoom under a name (up to 16). Saving an existing name replaces its camera. Save/Remove is undoable; Restore changes only navigation. Bookmarks travel with the drawing.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 480})}},
		{ID: "views-position", Kind: ui.Label, Text: position},
		{ID: "view-name", Kind: ui.TextField, Text: name, Accessibility: ui.Accessibility{Name: "Saved view name"}, Change: func(e ui.ChangeEvent) {
			if active() {
				d.invalid = len(e.Text) > 128
				if d.invalid {
					d.name = e.Text[:128]
					for !utf8.ValidString(d.name) && len(d.name) > 0 {
						d.name = d.name[:len(d.name)-1]
					}
					d.message = "view name exceeded 128 bytes and was shortened; edit it before saving"
				} else {
					d.name = e.Text
					d.message = ""
				}
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "views-choice", Kind: ui.Selector, SelectedID: choice, Options: options, Accessibility: ui.Accessibility{Name: "Saved views"}, Change: func(e ui.ChangeEvent) {
			if active() && slices.ContainsFunc(options, func(o ui.Option) bool { return o.ID == e.SelectedID }) {
				d.choice = e.SelectedID
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "views-error", Kind: ui.Label, Text: message},
		{ID: "view-save", Kind: ui.Button, Text: "Save current view", Disabled: d == nil || d.invalid, Activate: action("save")},
		{ID: "view-restore", Kind: ui.Button, Text: "Restore view", Disabled: choice == "", Activate: action("restore")},
		{ID: "view-remove", Kind: ui.Button, Text: "Remove view", Disabled: choice == "", Activate: action("remove")},
		{ID: "views-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel}}}
}
