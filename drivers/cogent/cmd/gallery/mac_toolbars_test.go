package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacToolbarCatalogueVariants(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, mode := range []string{"icons", "text", "both", "above"} {
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
						if mode == "above" && (item.IconPlacement != ui.IconAboveText || item.Text == "" || (item.IconImage == nil && item.Icon == "")) {
							t.Fatal("stacked mode", item.ID)
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

func TestMacToolbarCustomizationIsIndependentPerBar(t *testing.T) {
	for _, id := range []string{"toolbar-Small", "toolbar-Regular", "toolbar-Large", "toolbar-mixed"} {
		m := newMacToolbarGallery(false)
		findGalleryDocumentNode(m.view(), id+"/customize").Activate()
		panel := m.view()
		if !m.customizer.Open || m.customizer.Toolbar.ID != id {
			t.Fatal("wrong customization target", id)
		}
		original := findGalleryDocumentNode(panel, id+"/Save").IconImage
		findGalleryDocumentNode(panel, ui.ToolbarCustomizationEarlierID("toolbar-customizer", id+"/Save")).Activate()
		panel = m.view()
		findGalleryDocumentNode(panel, ui.ToolbarCustomizationVisibilityID("toolbar-customizer", id+"/New")).Change(ui.ChangeEvent{Checked: false})
		panel = m.view()
		bar := findGalleryDocumentNode(panel, id)
		if bar.Children[0].ID != id+"/Save" || bar.Children[0].IconImage != original {
			t.Fatal("configuration lost order/artwork", id)
		}
		other := "toolbar-Regular"
		if id == other {
			other = "toolbar-Small"
		}
		if findGalleryDocumentNode(panel, other).Children[0].ID != other+"/New" {
			t.Fatal("configuration changed peer toolbar")
		}
		for _, name := range []string{"Open", "Save", "Undo", "Redo", "Delete"} {
			findGalleryDocumentNode(m.view(), ui.ToolbarCustomizationVisibilityID("toolbar-customizer", id+"/"+name)).Change(ui.ChangeEvent{Checked: false})
		}
		panel = m.view()
		bar = findGalleryDocumentNode(panel, id)
		if len(bar.Children) != 0 {
			t.Fatal("all-hidden toolbar retained overflow or separator", id)
		}
		if e := ui.Validate(panel); e != nil {
			t.Fatal(e)
		}
		findGalleryDocumentNode(panel, ui.ToolbarCustomizationRestoreID("toolbar-customizer")).Activate()
		panel = m.view()
		if findGalleryDocumentNode(panel, id).Children[0].ID != id+"/New" {
			t.Fatal("restore failed", id)
		}
		findGalleryDocumentNode(panel, ui.ToolbarCustomizationDoneID("toolbar-customizer")).Activate()
		if m.customizer.Open || m.count != 0 {
			t.Fatal("customization ran action or stayed open", id)
		}
	}
}

func TestMacToolbarCustomizationQueuedTargetChangeCannotCrossSave(t *testing.T) {
	m := newMacToolbarGallery(false)
	initial := m.view()
	findGalleryDocumentNode(initial, "toolbar-Regular/customize").Activate()
	regular := m.view()
	oldHide := findGalleryDocumentNode(regular, ui.ToolbarCustomizationVisibilityID("toolbar-customizer", "toolbar-Regular/New"))
	// The new opener has run, but the owner has not built the new View yet.
	findGalleryDocumentNode(initial, "toolbar-Small/customize").Activate()
	oldHide.Change(ui.ChangeEvent{Checked: false})
	if len(m.layouts["toolbar-Regular"].Hidden) != 0 || len(m.layouts["toolbar-Small"].Hidden) != 0 {
		t.Fatal("stale edit saved to old/new target")
	}
	m.view()
	oldHide.Change(ui.ChangeEvent{Checked: false})
	if len(m.layouts["toolbar-Small"].Hidden) != 0 {
		t.Fatal("old sheet edited replacement toolbar")
	}
}

func TestMacToolbarContextDisplayAndCustomization(t *testing.T) {
	m := newMacToolbarGallery(false)
	n := m.view()
	bar := findGalleryDocumentNode(n, "toolbar-Regular")
	image := findGalleryDocumentNode(n, "toolbar-Regular/Save").IconImage
	bar.ContextMenu.Change(ui.ChangeEvent{SelectedID: "icons"})
	n = m.view()
	if findGalleryDocumentNode(n, "toolbar-Regular/Save").Text != "" || findGalleryDocumentNode(n, "toolbar-Small/Save").Text != "Save" || findGalleryDocumentNode(n, "toolbar-Regular/Save").IconImage != image {
		t.Fatal("per-toolbar display lost independence/artwork")
	}
	ctx := findGalleryDocumentNode(n, "toolbar-Regular").ContextMenu
	checked := false
	for _, o := range ctx.Options {
		if o.ID == "icons" {
			checked = o.Checked
		}
	}
	if !checked {
		t.Fatal("context menu lost current display check")
	}
	ctx.Change(ui.ChangeEvent{SelectedID: "customize"})
	m.view()
	if !m.customizer.Open || m.customizeTarget != "toolbar-Regular" || m.count != 0 {
		t.Fatal("context opened wrong customizer or command")
	}
	ctx.Change(ui.ChangeEvent{SelectedID: "text"})
	if m.displays["toolbar-Regular"] != "icons" {
		t.Fatal("stale context edited open modal")
	}
	m.customizer.Dismiss()
	n = m.view()
	findGalleryDocumentNode(n, "toolbar-display/text").Activate()
	n = m.view()
	if len(m.displays) != 0 || findGalleryDocumentNode(n, "toolbar-Regular/Save").IconImage != nil {
		t.Fatal("gallery display override did not reset per-bar mode")
	}
}
