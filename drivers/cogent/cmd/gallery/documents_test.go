package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestGalleryDocumentsCloseAcceptanceAndEmptyState(t *testing.T) {
	g := newGallery()
	find := func(n ui.Node, id string) ui.Node { return findGalleryDocumentNode(n, id) }
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	tabs := find(g.documentsView(), "document-tabs")
	g.documentReject = true
	tabs.TabClose(ui.TabCloseEvent{ID: "document-tabs", PageID: "sketch"})
	if len(g.documents) != 2 || g.documentSelected != "sketch" {
		t.Fatal("rejected close changed documents")
	}
	g.documentReject = false
	tabs.TabClose(ui.TabCloseEvent{ID: "document-tabs", PageID: "sketch"})
	if len(g.documents) != 1 || g.documentSelected != "notes" {
		t.Fatal("accepted selected close did not choose surviving page")
	}
	tabs = find(g.documentsView(), "document-tabs")
	tabs.TabClose(ui.TabCloseEvent{ID: "document-tabs", PageID: "notes"})
	n := g.documentsView()
	if len(g.documents) != 0 || find(n, "document-tabs").ID != "" || ui.Validate(g.view()) != nil {
		t.Fatal("last close retained invalid empty Tabs")
	}
	find(n, "document-reset").Activate()
	if len(g.documents) != 2 || ui.Validate(g.view()) != nil {
		t.Fatal("reopen failed")
	}
}
func findGalleryDocumentNode(n ui.Node, id string) ui.Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := findGalleryDocumentNode(c, id); found.ID != "" {
			return found
		}
	}
	return ui.Node{}
}
