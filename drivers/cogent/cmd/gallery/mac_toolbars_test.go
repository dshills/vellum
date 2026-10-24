package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacToolbarCatalogueVariants(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, mode := range []string{"icons", "text", "both"} {
			m := newMacToolbarGallery(dark)
			m.display = mode
			for _, narrow := range []bool{false, true} {
				m.narrow = narrow
				n := m.view()
				if err := ui.Validate(n); err != nil {
					t.Fatal(dark, mode, narrow, err)
				}
				for _, size := range macToolbarSizes {
					bar := findGalleryDocumentNode(n, "toolbar-"+size.name)
					for _, item := range bar.Children {
						if item.Kind != ui.Button {
							continue
						}
						if item.Accessibility.Name == "" {
							t.Fatal("unnamed action")
						}
						if mode == "icons" && (item.Text != "" || (item.Icon == "" && item.IconImage == nil)) {
							t.Fatal("icon mode", item.ID)
						}
						if mode == "text" && (item.Icon != "" || item.IconImage != nil || item.Text == "") {
							t.Fatal("text mode", item.ID)
						}
					}
				}
				newAction := findGalleryDocumentNode(n, "toolbar-mixed/New")
				openAction := findGalleryDocumentNode(n, "toolbar-mixed/Open")
				if newAction.Text != "New" || (newAction.Icon == "" && newAction.IconImage == nil) || openAction.Text != "Open" || openAction.Icon != "" || openAction.IconImage != nil {
					t.Fatal("mixed action presentation")
				}
				findGalleryDocumentNode(n, "toolbar-Regular/Save").Activate()
				if m.last != "toolbar-Regular: Save" {
					t.Fatal("action lost")
				}
				before := m.count
				m.disabled = true
				findGalleryDocumentNode(n, "toolbar-Regular/Save").Activate()
				findGalleryDocumentNode(n, "toolbar-Regular/Delete").Activate()
				if m.count != before {
					t.Fatal("disabled callback admitted")
				}
				m.disabled = false
			}
		}
	}
}

func TestToolbarFullColorComparison(t *testing.T) {
	m := newMacToolbarGallery(false)
	first := findGalleryDocumentNode(m.view(), "toolbar-Regular/New")
	if first.IconImage == nil || first.Icon != "" {
		t.Fatal("full-color artwork not initial")
	}
	findGalleryDocumentNode(m.view(), "toolbar-color").Change(ui.ChangeEvent{Checked: false})
	mono := findGalleryDocumentNode(m.view(), first.ID)
	if mono.Icon != ui.SymbolAdd || mono.IconImage != nil {
		t.Fatal("monochrome comparison failed")
	}
	findGalleryDocumentNode(m.view(), "toolbar-color").Change(ui.ChangeEvent{Checked: true})
	if next := findGalleryDocumentNode(m.view(), first.ID); next.IconImage != first.IconImage {
		t.Fatal("artwork not shared")
	}
}
