package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestBreadcrumbSmokeControlledPath(t *testing.T) {
	m := newModel()
	if err := ui.Validate(m.view()); err != nil {
		t.Fatal(err)
	}
	m.path.Navigate(ui.BreadcrumbEvent{ID: "path", ItemID: "folder"})
	if m.requests != 1 || len(m.path.Items) != 3 || m.failure != nil {
		t.Fatal("rejected navigation changed path")
	}
	m.reject = false
	m.path.Navigate(ui.BreadcrumbEvent{ID: "path", ItemID: "folder"})
	if m.requests != 2 || len(m.path.Items) != 2 || m.path.Items[1].ID != "folder" || m.failure != nil {
		t.Fatal("accepted ancestor did not become current")
	}
	m.path.Items = fullPath()
	if err := ui.Validate(m.view()); err != nil {
		t.Fatal(err)
	}
}
