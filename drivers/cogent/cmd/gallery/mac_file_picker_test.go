package main

import (
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestFilePickerRejectsOverlappingRequestAndRetiresBusy(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(app.Window{ID: "path", Title: "Synthetic", Width: 300, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	ref := a.Windows()[0].Ref()
	ticket, e := a.ReserveWindowCompletion(ref, nil)
	if e != nil {
		t.Fatal(e)
	}
	m := newMacComponentsGallery(false)
	m.filePickerView()
	m.filePickerTicket = ticket
	if e = m.filePicker.Request(ui.FilePickerEvent{ID: m.filePicker.ID, Action: ui.FilePickerBrowse}); !errors.Is(e, app.ErrServiceBusy) {
		t.Fatal("overlap not refused", e)
	}
	a.CloseWindow(ref)
	m.filePickerView()
	if m.filePicker.Busy || m.filePickerActive() {
		t.Fatal("closed window kept picker busy")
	}
}
