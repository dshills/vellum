package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
)

func (g *gallery) documentsView() ui.Node {
	content := ui.Node{ID: "documents", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "document-help", Kind: ui.Label, Text: "Close a document with its X button, or focus its tab and press Delete. Rejecting a request retains its draft."},
		{ID: "document-reject", Kind: ui.Checkbox, Text: "Reject close requests", Checked: g.documentReject, Change: func(e ui.ChangeEvent) { g.documentReject = e.Checked }},
		{ID: "document-reset", Kind: ui.Button, Text: "Reopen both documents", Activate: func() { g.documents = []string{"sketch", "notes"}; g.documentSelected = "sketch" }},
	}}
	if len(g.documents) == 0 {
		content.Children = append(content.Children, ui.Node{ID: "documents-empty", Kind: ui.Label, Text: "All documents are closed."})
	} else {
		tabs := ui.Node{ID: "document-tabs", Kind: ui.Tabs, SelectedID: g.documentSelected, Grow: true, Change: func(e ui.ChangeEvent) { g.documentSelectRequests++; g.documentSelected = e.SelectedID }, TabClose: func(e ui.TabCloseEvent) {
			g.documentCloseRequests++
			if g.documentReject {
				return
			}
			for i, id := range g.documents {
				if id == e.PageID {
					g.documents = append(g.documents[:i], g.documents[i+1:]...)
					break
				}
			}
			if g.documentSelected == e.PageID && len(g.documents) > 0 {
				g.documentSelected = g.documents[0]
			}
		}}
		for _, id := range g.documents {
			label := "Sketch"
			if id == "notes" {
				label = "Notes"
			}
			tabs.Options = append(tabs.Options, ui.Option{ID: id, Label: label, Closable: true})
			tabs.Children = append(tabs.Children, ui.Node{ID: id, Kind: ui.Column, Children: []ui.Node{{ID: "document-content-" + id, Kind: ui.TextField, Text: g.documentText[id], Accessibility: ui.Accessibility{Name: label + " draft"}, Change: func(e ui.ChangeEvent) { g.documentText[id] = e.Text }}}})
		}
		content.Children = append(content.Children, tabs)
	}
	content.Children = append(content.Children, ui.Node{ID: "document-close-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Close requests: %d · open documents: %d", g.documentCloseRequests, len(g.documents))})
	return content
}
