package main

import (
	"context"
	"github.com/dshills/vellum/command"
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
	// Escape must run before the native text field consumes its normal Abort key.
	if ok, e := r.ShortcutInScope(context.Background(), "Escape", command.AfterFocusShortcuts); e != nil || ok {
		t.Fatal("Escape leaked into text-widget fallback", ok, e)
	}
	if ok, e := r.ShortcutInScope(context.Background(), "Escape", command.WindowShortcuts); e != nil || !ok || f.bar.Open || f.focus.ID != "preview/find-editor" {
		t.Fatal("Escape did not close and restore editor focus", ok, e)
	}
	if ok, e := r.ShortcutInScope(context.Background(), "Meta+F", command.WindowShortcuts); e != nil || !ok || !f.bar.Open {
		t.Fatal("Find could not reopen", ok, e)
	}
	g.page = "inputs"
	if _, e := r.Shortcut(context.Background(), "Meta+F"); e == nil {
		t.Fatal("inactive page command")
	}
}
