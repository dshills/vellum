package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

type primarySelectionDraft struct {
	scene                    *canvas.Scene
	revision                 uint64
	before                   []canvas.Rect
	groups                   map[string]string
	ids                      []string
	primary, choice, message string
}

func (m *model) primarySelectionAvailable() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	units, _ := m.selectionUnits()
	if len(units) == 0 {
		return errNoSelection
	}
	return nil
}
func (m *model) bindPrimarySelection() error {
	return m.commands.Register(command.Command{ID: "primary-selection", Label: "Choose primary selection", Available: m.primarySelectionAvailable, Execute: func(context.Context) error { return m.beginPrimarySelection() }})
}
func (m *model) beginPrimarySelection() error {
	if err := m.primarySelectionAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	_, keys := m.selectionUnits()
	m.primaryDraft = &primarySelectionDraft{scene: m.scene, revision: m.journal.Snapshot().Revision, before: slices.Clone(m.scene.Objects), groups: maps.Clone(m.groups), ids: m.selectedIDs(), primary: m.scene.Selected, choice: keys[m.scene.Selected]}
	m.prompt = "primary-selection"
	m.modalEpoch++
	return nil
}
func (m *model) applyPrimarySelection(d *primarySelectionDraft) error {
	if d == nil || m.primaryDraft != d || m.scene != d.scene || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.scene.Objects, d.before) || !maps.Equal(m.groups, d.groups) || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		return history.ErrStale
	}
	if err := m.primarySelectionAvailable(); err != nil {
		return err
	}
	units, keys := m.selectionUnits()
	if !slices.ContainsFunc(units, func(u canvas.Rect) bool { return u.ID == d.choice }) {
		return errors.New("choose an already selected component or group")
	}
	primary := d.primary
	if keys[primary] != d.choice {
		primary = ""
		for _, id := range d.ids {
			if keys[id] == d.choice {
				primary = id
			}
		}
	}
	if primary == "" {
		return errNoSelection
	}
	m.setSelection(d.ids, primary)
	m.status = fmt.Sprintf("Primary selection: %s", m.layerLabel(d.choice))
	return nil
}

// A group marker represents the primary interaction unit; the actual primary
// member also remains identifiable without changing the selected set.
func (m *model) primaryLayerSuffix(id string) string {
	if id == m.scene.Selected && slices.Contains(m.selectedIDs(), id) {
		return " [primary]"
	}
	if id != "" && m.groups[m.scene.Selected] == id && slices.Contains(m.selectedIDs(), m.scene.Selected) {
		return " [primary group]"
	}
	return ""
}
func (m *model) primarySelectionDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.primaryDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.primaryDraft == d && m.prompt == "primary-selection"
	}
	cancel := func() {
		if active() {
			m.primaryDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	options := []ui.Option{}
	choice, message := "", ""
	if d != nil {
		choice, message = d.choice, d.message
		units, _ := m.selectionUnits()
		for _, u := range units {
			label := m.layerLabel(u.ID)
			if d.groups[u.ID] == "" && slices.ContainsFunc(d.ids, func(id string) bool { return d.groups[id] == u.ID }) {
				label = "Group " + label
			}
			options = append(options, ui.Option{ID: u.ID, Label: label})
		}
	}
	return ui.Node{ID: "primary-selection-dialog", Kind: ui.Dialog, Text: "Choose primary selection", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "primary-selection-help", Kind: ui.Label, Text: "Choose a component or whole group already in the selection. The selected set stays unchanged; this does not add Undo history.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 440})}},
		{ID: "primary-choice", Kind: ui.Selector, Options: options, SelectedID: choice, Accessibility: ui.Accessibility{Name: "Primary selected unit"}, Change: func(e ui.ChangeEvent) {
			if active() && slices.ContainsFunc(options, func(o ui.Option) bool { return o.ID == e.SelectedID }) {
				d.choice = e.SelectedID
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "primary-error", Kind: ui.Label, Text: message},
		{ID: "primary-apply", Kind: ui.Button, Text: "Make primary", Activate: func() {
			if !active() {
				return
			}
			m.prompt = ""
			if err := m.applyPrimarySelection(d); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.primaryDraft = nil
				} else {
					d.message = err.Error()
					m.prompt = "primary-selection"
				}
			} else {
				m.primaryDraft = nil
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "primary-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}
