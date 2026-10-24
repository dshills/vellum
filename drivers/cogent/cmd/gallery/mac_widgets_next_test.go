package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacWidgetsNextPagesUsePublicComponents(t *testing.T) {
	for _, dark := range []bool{false, true} {
		inputs := newInputHelpersGallery()
		inputs.setAppearance(dark)
		m := newMacWidgetsGallery(inputs, dark)
		pages := []string{"navigation", "sheets", "mac-collection-page", "shortcuts", "levels", "toolbars", "inputs"}
		for _, page := range pages {
			m.page = page
			n := m.view()
			if err := ui.Validate(n); err != nil {
				t.Fatal(page, dark, err)
			}
			tabs := n.Children[0]
			if !n.Style.Background.Set || n.Style.Background.Value.A != 255 {
				t.Fatal("gallery must supply an opaque window surface", dark)
			}
			if !tabs.Style.Foreground.Set {
				t.Fatal("tab strip must supply its appearance ink", dark)
			}
			for _, pageNode := range tabs.Children {
				if !pageNode.Style.GrowX.Set || pageNode.Style.GrowX.Value != 1 || pageNode.Style.Background != n.Style.Background {
					t.Fatal("page must fill and paint its surface", pageNode.ID, dark)
				}
			}
			found := false
			for i, o := range tabs.Options {
				if o.ID == page {
					found = true
					if tabs.Children[i].ID != page {
						t.Fatal("page order mismatch", page)
					}
				}
			}
			if !found {
				t.Fatal("missing demo page", page)
			}
		}
	}
}
