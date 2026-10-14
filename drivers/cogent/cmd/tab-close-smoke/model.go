package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"slices"
)

type model struct {
	documents                              []string
	selected, draft                        string
	reject, disabled                       bool
	requests, selections, edits, decisions int
	failure                                error
}

func newModel() *model {
	return &model{documents: []string{"a", "b"}, selected: "a", draft: "Original draft", reject: true}
}
func (m *model) view() ui.Node {
	root := ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "reject", Kind: ui.Checkbox, Text: "Reject close requests", Checked: m.reject, Change: func(e ui.ChangeEvent) { m.reject = e.Checked; m.decisions++ }},
	}}
	if len(m.documents) == 0 {
		root.Children = append(root.Children, ui.Node{ID: "empty", Kind: ui.Label, Text: "All documents closed"})
		return root
	}
	tabs := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: m.selected, Disabled: m.disabled, Grow: true, Change: func(e ui.ChangeEvent) { m.selected = e.SelectedID; m.selections++ }, TabClose: func(e ui.TabCloseEvent) {
		m.requests++
		if e.ID != "tabs" || !slices.Contains(m.documents, e.PageID) {
			m.failure = fmt.Errorf("unexpected close request %+v", e)
			return
		}
		if m.reject {
			return
		}
		i := slices.Index(m.documents, e.PageID)
		m.documents = slices.Delete(m.documents, i, i+1)
		if m.selected == e.PageID && len(m.documents) > 0 {
			m.selected = m.documents[0]
		}
	}}
	for _, id := range m.documents {
		label := "First"
		if id == "b" {
			label = "Second"
		}
		tabs.Options = append(tabs.Options, ui.Option{ID: id, Label: label, Closable: true})
		page := ui.Node{ID: id, Kind: ui.Column, Grow: true}
		if id == "a" {
			page.Children = []ui.Node{{ID: "draft", Kind: ui.TextField, Text: m.draft, Change: func(e ui.ChangeEvent) { m.draft = e.Text; m.edits++ }, Accessibility: ui.Accessibility{Name: "Document draft"}}}
		} else {
			page.Children = []ui.Node{{ID: "second-content", Kind: ui.Label, Text: "Second document"}}
		}
		tabs.Children = append(tabs.Children, page)
	}
	root.Children = append(root.Children, tabs)
	return root
}
