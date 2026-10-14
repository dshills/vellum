package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestGalleryBreadcrumbControlledNavigationAndReset(t *testing.T) {
	g := newGallery()
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	g.breadcrumbReject = true
	g.breadcrumb.Navigate(ui.BreadcrumbEvent{ID: "project-path", ItemID: "workspace"})
	if len(g.breadcrumb.Items) != 3 || g.breadcrumbRequests != 1 {
		t.Fatal("rejected gallery path")
	}
	g.breadcrumbReject = false
	g.breadcrumb.Navigate(ui.BreadcrumbEvent{ID: "project-path", ItemID: "workspace"})
	if len(g.breadcrumb.Items) != 2 || g.breadcrumbRequests != 2 {
		t.Fatal("accepted gallery path")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	g.breadcrumb.Items = galleryBreadcrumbItems()
	if err := ui.Validate(g.view()); err != nil || len(g.breadcrumb.Items) != 3 {
		t.Fatal("reset path", err)
	}
}
