package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacComponentThemePreviewRetainsArtworkAndValue(t *testing.T) {
	m := newMacComponentsGallery(false)
	n := m.view()
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	original := findGalleryDocumentNode(n, "components/color-button").IconImage
	findGalleryDocumentNode(n, "components/name").Change(ui.ChangeEvent{Text: "Drawing"})
	findGalleryDocumentNode(n, "components/appearance/Dark").Activate()
	findGalleryDocumentNode(n, "components/accent").Change(ui.ChangeEvent{SelectedID: "purple"})
	next := m.view()
	if err := ui.Validate(next); err != nil {
		t.Fatal(err)
	}
	if !m.dark || m.accent != "purple" || m.name != "Drawing" || findGalleryDocumentNode(next, "components/color-button").IconImage != original {
		t.Fatal("theme transition lost state or artwork")
	}
}
