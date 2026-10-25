package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"strings"
)

type macScopeGallery struct {
	bar       *ui.SearchScopeBar
	selection []string
}

func (m *macComponentsGallery) scopeView() ui.Node {
	if m.scopes == nil {
		p := &macScopeGallery{bar: &ui.SearchScopeBar{ID: "preview/scopes", Label: "Search sample documents", Value: ui.SearchScopeValue{ScopeID: "all"}}}
		p.bar.Change = func(e ui.SearchScopeEvent) {
			p.bar.Value = e.Apply(p.bar.Value)
			m.count++
			m.status = "Search scope: " + p.bar.Value.ScopeID
		}
		m.scopes = p
	}
	p := m.scopes
	theme := m.theme()
	p.bar.SelectedStyle = ui.Style{Background: ui.Some(theme.Palette.Value.Accent), Foreground: ui.Some(theme.Palette.Value.AccentInk)}
	catalog := []struct {
		id, name, scope string
		selected        bool
	}{{"floor", "Floor plan.vellum", "project", true}, {"detail", "Details.vellum", "project", false}, {"notes", "Project notes.txt", "project", true}, {"other", "Garden drawing.vellum", "other", false}}
	counts := map[string]int{}
	items := []ui.DocumentItem{}
	for _, doc := range catalog {
		if !strings.Contains(strings.ToLower(doc.name), strings.ToLower(p.bar.Value.Query)) {
			continue
		}
		counts["all"]++
		if doc.scope == "project" {
			counts["project"]++
		}
		if doc.selected {
			counts["selection"]++
		}
		if p.bar.Value.ScopeID == "all" || p.bar.Value.ScopeID == "project" && doc.scope == "project" || p.bar.Value.ScopeID == "selection" && doc.selected {
			items = append(items, ui.DocumentItem{ID: doc.id, Name: doc.name, Kind: "Sample", Image: m.artwork[ui.SymbolDocument]})
		}
	}
	p.bar.Items = []ui.SearchScopeItem{{ID: "all", Label: "All Documents", Count: counts["all"], IconImage: m.artwork[ui.SymbolFolder]}, {ID: "project", Label: "Current Project", Count: counts["project"], IconImage: m.artwork[ui.SymbolDocument]}, {ID: "selection", Label: "Selected Files", Count: counts["selection"], IconImage: m.artwork[ui.SymbolSave]}, {ID: "archive", Label: "Archive", Count: -1, Disabled: true, IconImage: m.artwork[ui.SymbolFolder]}}
	bar, e := p.bar.View()
	if e != nil {
		return ui.Node{ID: "preview/scopes-error", Kind: ui.Label, Text: e.Error()}
	}
	selected := []string{}
	for _, item := range items {
		for _, id := range p.selection {
			if item.ID == id {
				selected = append(selected, id)
			}
		}
	}
	list := &ui.DocumentList{ID: "preview/scope-results", Label: "Search results", Columns: []ui.DocumentColumn{{ID: "name", Title: "Name", Width: 300}}, Items: items, SelectedIDs: selected, Change: func(e ui.ListSelectionEvent) { p.selection = e.Apply(p.selection) }}
	results, e := list.View()
	if e != nil {
		return ui.Node{ID: "preview/scopes-error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "preview/scopes-page", Kind: ui.Column, Children: []ui.Node{{ID: "preview/scopes-help", Kind: ui.Label, Text: "Finder-style scope choices with original-color artwork and query-dependent counts. Arrows move toolbar focus; Return chooses. Sample data only, no disk search."}, {ID: "preview/scopes-small", Kind: ui.Checkbox, Text: "Small controls", Checked: p.bar.Small, Change: func(e ui.ChangeEvent) { p.bar.Small = e.Checked }}, bar, {ID: "preview/scopes-summary", Kind: ui.Label, Text: fmt.Sprintf("%d matching sample documents", len(items))}, results}}
}
