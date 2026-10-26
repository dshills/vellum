package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacShareBusyExpiresWithExactWindow(t *testing.T) {
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	if e = a.Open(app.Window{ID: "share", Title: "Synthetic", Width: 300, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); e != nil {
		t.Fatal(e)
	}
	ref := a.Windows()[0].Ref()
	ticket, e := a.ReserveWindowCompletion(ref, nil)
	if e != nil {
		t.Fatal(e)
	}
	m := newMacComponentsGallery(false)
	m.shareTicket = ticket
	if !m.shareActive() {
		t.Fatal("live share did not disable button")
	}
	if e = a.CloseWindow(ref); e != nil {
		t.Fatal(e)
	}
	if m.shareActive() {
		t.Fatal("revoked share stayed busy without a publish callback")
	}
}
