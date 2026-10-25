package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestGallerySearchScopesFilterWithoutDisk(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.scopeView()
	m.scopes.bar.Value = ui.SearchScopeValue{ScopeID: "project", Query: "plan"}
	n := m.scopeView()
	summary := findGalleryDocumentNode(n, "preview/scopes-summary")
	if summary.Text != "1 matching sample documents" {
		t.Fatal(summary.Text)
	}
	m.scopes.bar.Value.Query = "no such item"
	n = m.scopeView()
	if findGalleryDocumentNode(n, "preview/scopes-summary").Text != "0 matching sample documents" || ui.Validate(n) != nil {
		t.Fatal("empty results")
	}
}
