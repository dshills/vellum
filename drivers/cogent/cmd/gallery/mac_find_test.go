package main

import (
	"context"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacFindCommandsAndUndo(t *testing.T) {
	g := newMacWidgetsGallery(nil, false)
	g.components.page = "find-replace"
	g.components.findView()
	f := g.components.find
	r := g.findCommands()
	f.bar.Open = false
	if ok, e := r.Shortcut(context.Background(), "Meta+F"); e != nil || !ok || !f.bar.Open || f.focus.ID != ui.FindBarInputID(f.bar.ID) {
		t.Fatal(e)
	}
	f.bar.Request(ui.FindRequest{Action: ui.FindReplaceAll, Query: "plan", Replacement: "drawing"})
	text := f.model.Snapshot().Text
	if e := f.model.Journal().Undo(); e != nil || text == f.model.Snapshot().Text {
		t.Fatal(e)
	}
	g.page = "inputs"
	if _, e := r.Shortcut(context.Background(), "Meta+F"); e == nil {
		t.Fatal("inactive page command")
	}
}
