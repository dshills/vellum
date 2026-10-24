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
		pages := []string{"navigation", "sheets", "mac-collection-page", "shortcuts", "levels", "inputs"}
		for _, page := range pages {
			m.page = page
			n := m.view()
			if err := ui.Validate(n); err != nil {
				t.Fatal(page, dark, err)
			}
			tabs := n.Children[0]
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
