package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

// Search is bounded and remains a transient selection operation. Its draft owns
// the originating selection/document so a retained dialog cannot affect a later one.
type selectMatchingDraft struct {
	kind, query, message string
	add                  bool
	invalidQuery         bool
	scene                *canvas.Scene
	journal              *history.Journal
	revision             uint64
	objects              []canvas.Rect
	groups               map[string]string
	layers               map[string]layerMetadata
	ids                  []string
	primary              string
}

func (m *model) bindSelectMatching() error {
	return m.commands.Register(command.Command{ID: "select-matching", Label: "Select matching components", Available: m.selectionAvailable, Execute: func(context.Context) error { return m.beginSelectMatching() }})
}
func (m *model) beginSelectMatching() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	m.matchingDraft = &selectMatchingDraft{kind: "all", scene: m.scene, journal: m.journal, revision: m.journal.Snapshot().Revision, objects: slices.Clone(m.scene.Objects), groups: maps.Clone(m.groups), layers: maps.Clone(m.layerInfo), ids: m.selectedIDs(), primary: m.scene.Selected}
	m.prompt = "select-matching"
	m.modalEpoch++
	return nil
}
func validMatchingQuery(kind, query string) error {
	if kind != "all" && kind != "block" && kind != "connector" && kind != "note" {
		return errors.New("choose a component type")
	}
	if !utf8.ValidString(query) || utf8.RuneCountInString(query) > 128 {
		return errors.New("name or ID search is limited to 128 characters")
	}
	if kind == "all" && strings.TrimSpace(query) == "" {
		return errors.New("choose a type or enter a name or ID")
	}
	return nil
}
func (m *model) matchingSelection(kind, query string, add bool) ([]string, error) {
	if err := validMatchingQuery(kind, query); err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	ids := []string{}
	if add {
		ids = append(ids, m.selectedIDs()...)
	}
	for _, o := range m.scene.Objects {
		component, _, _ := strings.Cut(o.ID, "-")
		if kind != "all" && kind != component {
			continue
		}
		nameMatch := strings.Contains(strings.ToLower(m.layerLabel(o.ID)), query)
		if group := m.groups[o.ID]; group != "" {
			nameMatch = nameMatch || strings.Contains(strings.ToLower(m.layerLabel(group)), query)
		}
		if nameMatch {
			ids = append(ids, o.ID)
		}
	}
	return m.expandedSelection(ids), nil
}
func (m *model) applySelectMatching(d *selectMatchingDraft) error {
	if d == nil || m.matchingDraft != d || m.scene != d.scene || m.journal != d.journal || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.scene.Objects, d.objects) || !maps.Equal(m.groups, d.groups) || !maps.Equal(m.layerInfo, d.layers) || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		return history.ErrStale
	}
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if d.invalidQuery {
		return errors.New("edit the search to at most 128 characters")
	}
	ids, err := m.matchingSelection(d.kind, d.query, d.add)
	if err != nil {
		return err
	}
	m.setSelection(ids, d.primary)
	m.status = fmt.Sprintf("Matching selection: %d components", len(m.selectedIDs()))
	return nil
}
func (m *model) selectMatchingDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.matchingDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.matchingDraft == d && m.prompt == "select-matching"
	}
	cancel := func() {
		if active() {
			m.matchingDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	kind, query, message, add := "all", "", "", false
	if d != nil {
		kind, query, message, add = d.kind, d.query, d.message, d.add
	}
	return ui.Node{ID: "matching-dialog", Kind: ui.Dialog, Text: "Select matching components", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "matching-help", Kind: ui.Label, Text: "Type and name/ID filters combine. Names and IDs match without case, including group names. Matching any member selects its whole group; hidden or locked units are excluded. Replace with no matches clears selection. Selection changes add no Undo entry.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 500})}},
		{ID: "matching-kind", Kind: ui.Selector, SelectedID: kind, Options: []ui.Option{{ID: "all", Label: "Any type"}, {ID: "block", Label: "Block"}, {ID: "connector", Label: "Connector"}, {ID: "note", Label: "Note"}}, Change: func(e ui.ChangeEvent) {
			if active() {
				switch e.SelectedID {
				case "all", "block", "connector", "note":
					d.kind = e.SelectedID
					if !d.invalidQuery {
						d.message = ""
					}
					_ = a.Invalidate("placement")
				}
			}
		}},
		{ID: "matching-query", Kind: ui.TextField, Text: query, Accessibility: ui.Accessibility{Name: "Name or ID contains"}, Change: func(e ui.ChangeEvent) {
			if active() {
				if !utf8.ValidString(e.Text) || utf8.RuneCountInString(e.Text) > 128 {
					d.invalidQuery = true
					d.message = "name or ID search is limited to 128 characters"
				} else {
					d.query = e.Text
					d.invalidQuery = false
					d.message = ""
				}
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "matching-add", Kind: ui.Checkbox, Text: "Add to current selection", Checked: add, Change: func(e ui.ChangeEvent) {
			if active() {
				d.add = e.Checked
				_ = a.Invalidate("placement")
			}
		}},
		{ID: "matching-error", Kind: ui.Label, Text: message},
		{ID: "matching-apply", Kind: ui.Button, Text: "Select matches", Disabled: message != "", Activate: func() {
			if !active() || d.message != "" || d.invalidQuery {
				return
			}
			m.prompt = ""
			if err := m.applySelectMatching(d); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.matchingDraft = nil
				} else {
					d.message = err.Error()
					m.prompt = "select-matching"
				}
			} else {
				m.matchingDraft = nil
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "matching-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}
