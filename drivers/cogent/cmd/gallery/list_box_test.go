package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestGalleryListBoxControlledSelection(t *testing.T) {
	g := newGallery()
	list := func() ui.Node { return findGalleryDocumentNode(g.view(), "layers") }
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	n := list()
	n.Change(ui.ChangeEvent{SelectedID: "layer-23"})
	if list().SelectedID != "layer-23" || g.listRequests != 1 {
		t.Fatal("accepted selection")
	}
	g.listReject = true
	list().Change(ui.ChangeEvent{SelectedID: "layer-00"})
	if list().SelectedID != "layer-23" || g.listRequests != 2 {
		t.Fatal("rejected selection")
	}
	g.listReversed = true
	if list().Options[0].ID != "layer-23" || list().SelectedID != "layer-23" {
		t.Fatal("reordering lost identity")
	}
	g.listReadOnly = true
	if list().Change != nil {
		t.Fatal("readonly selection writable")
	}
	g.listDisabled = true
	if !list().Disabled {
		t.Fatal("disabled state missing")
	}
	findGalleryDocumentNode(g.view(), "list-clear").Activate()
	if list().SelectedID != "" {
		t.Fatal("clear selection")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}
