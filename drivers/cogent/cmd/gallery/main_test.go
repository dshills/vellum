package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func node(n ui.Node, id string) ui.Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if result := node(c, id); result.ID != "" {
			return result
		}
	}
	return ui.Node{}
}
func TestGalleryControlledValuesDialogAndFocus(t *testing.T) {
	g := newGallery()
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	node(g.view(), "slider").Change(ui.ChangeEvent{Number: 60})
	if node(g.view(), "number").Number != 60 {
		t.Fatal("numeric controls do not share accepted model value")
	}
	node(g.view(), "focus-name").Activate()
	if g.view().Focus.ID != "project-name" || g.view().Focus.Sequence != 1 {
		t.Fatal("focus request not issued")
	}
	node(g.view(), "enable").Change(ui.ChangeEvent{Checked: false})
	if !node(g.view(), "fields").Disabled {
		t.Fatal("checkbox doesn't disable fields")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal("consumed focus request should survive later disabled state", err)
	}
	node(g.view(), "show-dialog").Activate()
	if !node(g.view(), "preferences").Open {
		t.Fatal("dialog didn't open")
	}
	node(g.view(), "preferences").Dismiss()
	if g.dialog {
		t.Fatal("dismissal didn't close model state")
	}
}
