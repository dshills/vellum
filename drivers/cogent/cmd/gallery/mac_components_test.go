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

func TestMacComponentPagesAreConsumable(t *testing.T) {
	for _, dark := range []bool{false, true} {
		m := newMacComponentsGallery(dark)
		for _, page := range []string{"theme", "menus", "split", "inspector", "browser", "rules"} {
			m.page = page
			n := m.view()
			if err := ui.Validate(n); err != nil {
				t.Fatal(page, err)
			}
			if page == "browser" && findGalleryDocumentNode(n, "components/browser").ID == "" {
				t.Fatal("browser failed to construct")
			}
		}
	}
}

func TestMacComponentRuleEditsRetainAcrossTheme(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.page = "rules"
	n := m.view()
	add := findGalleryDocumentNode(n, ui.RuleControlID("components/rules", "root", "add-rule"))
	if add.IconImage == nil {
		t.Fatal("missing color artwork")
	}
	add.Activate()
	m.dark = true
	m.accent = "green"
	n = m.view()
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	if len(m.rules.Root.Group.Rules) != 4 || findGalleryDocumentNode(n, ui.RuleControlID("components/rules", "root", "add-rule")).IconImage != add.IconImage {
		t.Fatal("theme lost rules or art")
	}
}
