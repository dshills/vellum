package main

import (
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/examples/macapps/notes"
	"testing"
)

func TestActiveDocumentMenusHistoryAndClosedRefusal(t *testing.T) {
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	w := &workspace{owner: a, documents: map[app.WindowRef]*documentWindow{}}
	var refs []app.WindowRef
	for _, id := range []string{"A", "B"} {
		m := notes.New()
		r, e := a.OpenWindow(app.Window{ID: id, Title: id, Width: 600, Height: 400, View: m.View})
		if e != nil {
			t.Fatal(e)
		}
		w.documents[r] = &documentWindow{ref: r, model: m}
		refs = append(refs, r)
	}
	if e = w.bindMenus(); e != nil {
		t.Fatal(e)
	}
	if w.menus.Enabled("notes.new") == nil {
		t.Fatal("no active document admitted")
	}
	for _, r := range refs {
		w.activate(r)
		if e = w.menus.Invoke(context.Background(), "notes.new"); e != nil {
			t.Fatal(e)
		}
	}
	if w.documents[refs[0]].model.History().UndoCount != 1 || w.documents[refs[1]].model.History().UndoCount != 1 {
		t.Fatal("histories not independent")
	}
	if e = w.menus.Invoke(context.Background(), "notes.undo"); e != nil {
		t.Fatal(e)
	}
	if w.documents[refs[0]].model.History().UndoCount != 1 || w.documents[refs[1]].model.History().UndoCount != 0 {
		t.Fatal("wrong history routed")
	}
	if e = a.CloseWindow(refs[1]); e != nil {
		t.Fatal(e)
	}
	if w.menus.Enabled("notes.new") == nil {
		t.Fatal("retired active document admitted")
	}
	w.activate(refs[0])
	if e = w.menus.Invoke(context.Background(), "notes.new"); e != nil {
		t.Fatal(e)
	}
}
