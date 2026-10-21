package main

import (
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"strings"
	"unicode/utf8"
)

func (m *model) layerFilterActive() bool { return strings.TrimSpace(m.layerFilter) != "" }
func (m *model) emptyLayersLabel() string {
	if m.layerFilterActive() {
		return "No matching layers"
	}
	return "No components"
}
func (m *model) setLayerFilter(value string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > 128 {
		return errors.New("layer filter is limited to 128 characters")
	}
	m.layerFilter = value
	return nil
}

// Filtering changes only this projection. A matching group reveals all members;
// a matching member keeps its group marker as context. Search reveals matches
// even inside collapsed groups; clearing it restores the prior collapse state.
func (m *model) layerRows() []layerRow {
	all := m.allLayerRows()
	query := strings.ToLower(strings.TrimSpace(m.layerFilter))
	matches := map[string]bool{}
	groups := map[string]bool{}
	if query != "" {
		for _, r := range all {
			if strings.Contains(strings.ToLower(m.layerLabel(r.id)), query) {
				matches[r.id] = true
				if g := m.groups[r.id]; g != "" {
					groups[g] = true
				}
			}
		}
	}
	rows := make([]layerRow, 0, len(all))
	for _, r := range all {
		g := m.groups[r.id]
		marker := g == "" && strings.HasPrefix(r.id, "group-")
		if query != "" {
			if !matches[r.id] && !matches[g] && !(marker && groups[r.id]) {
				continue
			}
		} else if g != "" && m.collapsedGroups[g] {
			continue
		}
		if marker && m.collapsedGroups[r.id] && query == "" {
			r.label = "▸ " + r.label
		}
		rows = append(rows, r)
	}
	return rows
}
func (m *model) setGroupRows(action string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	live := map[string]bool{}
	for _, g := range m.groups {
		live[g] = true
	}
	next := make(map[string]bool, len(live))
	for g := range live {
		if m.collapsedGroups[g] {
			next[g] = true
		}
	}
	switch action {
	case "collapse-all":
		for g := range live {
			next[g] = true
		}
	case "expand-all":
		clear(next)
	default:
		id, ok := strings.CutPrefix(action, "toggle-")
		if !ok || !live[id] {
			return errors.New("group no longer available")
		}
		if next[id] {
			delete(next, id)
		} else {
			next[id] = true
		}
	}
	m.collapsedGroups = next
	return nil
}
func (m *model) layerOrganizationControls(ref app.WindowRef, update func()) ui.Node {
	disabled := m.controlDisabled("interaction", m.selectionAvailable() != nil)
	finish := func(action func() error) {
		if ref.Context().Err() != nil {
			return
		}
		if err := action(); err != nil {
			m.status = err.Error()
		}
		update()
	}
	options := []ui.Option{{ID: "collapse-all", Label: "Collapse All"}, {ID: "expand-all", Label: "Expand All"}}
	for _, row := range m.allLayerRows() {
		if !strings.HasPrefix(row.id, "group-") {
			continue
		}
		label := "Collapse "
		if m.collapsedGroups[row.id] {
			label = "Expand "
		}
		options = append(options, ui.Option{ID: "toggle-" + row.id, Label: label + m.layerLabel(row.id)})
	}
	help := "Filter names or IDs; matching groups show members."
	if m.layerFilterActive() {
		help = "Filtering layers; drag reordering disabled. Clear the filter to reorder."
	}
	return ui.Node{ID: "layers-organization", Kind: ui.Column, Children: []ui.Node{
		{ID: "layers-filter-label", Kind: ui.Label, Text: "Filter layers"},
		{ID: "layers-filter", Kind: ui.TextField, Text: m.layerFilter, Disabled: disabled, Accessibility: ui.Accessibility{Name: "Filter layers", Help: help}, Change: func(e ui.ChangeEvent) { finish(func() error { return m.setLayerFilter(e.Text) }) }},
		{ID: "layers-filter-tools", Kind: ui.Row, Children: []ui.Node{
			{ID: "layers-filter-clear", Kind: ui.Button, Text: "Clear filter", Style: compactControlStyle(), Disabled: disabled || m.layerFilter == "", Activate: func() { finish(func() error { return m.setLayerFilter("") }) }},
			{ID: "layers-groups", Kind: ui.MenuButton, Text: "Groups", Style: compactControlStyle(), Disabled: disabled || len(options) == 2, Options: options, Change: func(e ui.ChangeEvent) { finish(func() error { return m.setGroupRows(e.SelectedID) }) }},
		}},
	}}
}
