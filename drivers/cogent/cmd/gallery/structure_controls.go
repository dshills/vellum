package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
)

// Ordinary gallery controls shared with the bounded native consumer.
func (g *gallery) detailsControl(children ...ui.Node) ui.Node {
	var change func(ui.ChangeEvent)
	if !g.detailsReadOnly {
		change = func(e ui.ChangeEvent) {
			g.detailsRequests++
			if !g.detailsReject {
				g.details = e.Checked
			}
		}
	}
	return ui.Node{ID: "details", Kind: ui.Disclosure, Text: "Editing controls", Checked: g.details, Disabled: g.detailsDisabled, Change: change, Children: children}
}
func (g *gallery) splitControl() ui.Node {
	var change func(ui.ChangeEvent)
	if !g.splitReadOnly {
		change = func(e ui.ChangeEvent) {
			g.splitRequests++
			if !g.splitReject {
				g.split = e.Number
			}
		}
	}
	content := []ui.Node{
		{ID: "split-content", Kind: ui.Label, Text: "Workspace content retains its native controls as the panes resize."},
		{ID: "split-reset", Kind: ui.Button, Text: "Reset divider", Activate: func() { g.split = 0.35 }},
	}
	for i := range 24 {
		content = append(content, ui.Node{ID: fmt.Sprintf("split-line-%02d", i), Kind: ui.Label, Text: fmt.Sprintf("Workspace row %02d", i+1)})
	}
	return ui.Node{ID: "panes", Kind: ui.SplitPane, Number: g.split, Grow: true, Disabled: g.splitDisabled, Change: change, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(0))}, Accessibility: ui.Accessibility{Name: "Inspector width", Help: "Resize the inspector and workspace"}, Children: []ui.Node{
		{ID: "split-inspector", Kind: ui.FormGroup, Text: "Inspector", Grow: true, Children: []ui.Node{{ID: "split-name", Kind: ui.TextField, Text: g.text, Change: func(e ui.ChangeEvent) { g.text = e.Text }, Accessibility: ui.Accessibility{Name: "Project name"}}}},
		{ID: "split-workspace", Kind: ui.Scroll, Grow: true, Children: content},
	}}
}
func (g *gallery) structureCatalogueView() ui.Node {
	focus := g.focus
	if focus.Sequence == 0 {
		focus = ui.FocusRequest{ID: "details", Sequence: 1}
	}
	sections := ui.Node{ID: "sections", Kind: ui.Column, Grow: true, Children: []ui.Node{
		g.detailsControl(
			ui.Node{ID: "section-draft", Kind: ui.TextField, Text: g.text, Change: func(e ui.ChangeEvent) { g.text = e.Text }},
			ui.Node{ID: "section-action", Kind: ui.Button, Text: "Section action", Activate: func() { g.outsideClicks++ }},
		),
		{ID: "section-after", Kind: ui.Button, Text: "After section"},
	}}
	return ui.Node{ID: "structure-catalogue", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{
		{ID: "structure-title", Kind: ui.Label, Text: "Sections, split panes and document tabs"},
		{ID: "structure-pages", Kind: ui.Tabs, Grow: true, SelectedID: g.structurePage, Options: []ui.Option{{ID: "sections", Label: "Sections"}, {ID: "unavailable", Label: "Unavailable", Disabled: true}, {ID: "split-page", Label: "Split panes"}, {ID: "documents", Label: "Documents"}}, Change: func(e ui.ChangeEvent) { g.structurePageRequests++; g.structurePage = e.SelectedID }, Children: []ui.Node{sections, {ID: "unavailable", Kind: ui.Column}, {ID: "split-page", Kind: ui.Column, Grow: true, Children: []ui.Node{g.splitControl()}}, g.documentsView()}},
		{ID: "structure-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Page %s / expanded %t / split %.2f / documents %d", g.structurePage, g.details, g.split, len(g.documents))},
	}}
}
