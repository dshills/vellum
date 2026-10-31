package main

import (
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"testing"
)

func TestAllPagesAndModalContracts(t *testing.T) {
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	doc, e := editor.New(sampleText, editor.Config{})
	if e != nil {
		t.Fatal(e)
	}
	grid, e := virtual.NewGrid(rows{}, []virtual.GridColumn{{ID: "part", Title: "Part", Width: 180}, {ID: "value", Title: "Value", Width: 180}}, virtual.GridConfig{RowHeight: 30, HeaderHeight: 34, FrozenColumns: 1}, virtual.GridViewport{Width: 800, Height: 480})
	if e != nil {
		t.Fatal(e)
	}
	d := &demo{preset: "basic", owner: a, document: doc, grid: grid, scene: &canvas.Scene{}, name: "Developer", enabled: true, level: 65}
	d.latest = newLatestFeatures(d)
	for _, page := range []string{"latest", "overview", "editing", "data", "services", "suite"} {
		for _, modal := range []bool{false, true} {
			d.tab = page
			d.modal = modal
			if e := ui.Validate(d.view()); e != nil {
				t.Errorf("%s modal=%t: %v", page, modal, e)
			}
		}
	}
	d.modal = false
	d.tab = "overview"
	r, e := d.commands()
	if e != nil {
		t.Fatal(e)
	}
	d.enabled = false
	if e := r.Invoke(context.Background(), "increment-demo"); e == nil {
		t.Fatal("disabled counter invoked through menu")
	}
	if d.count != 0 {
		t.Fatal("disabled command changed state")
	}
	d.enabled = true
	if e := r.Invoke(context.Background(), "increment-demo"); e != nil {
		t.Fatal(e)
	}
	if d.count != 1 || d.tab != "overview" {
		t.Fatal("menu did not update shared model")
	}
	if e := r.Invoke(context.Background(), "show-services"); e != nil {
		t.Fatal(e)
	}
	if d.tab != "services" {
		t.Fatal("service menu routing")
	}
}

func TestServicePayloadContracts(t *testing.T) {
	if err := (app.Notification{ID: notificationID, Title: "Vellum Showcase", Body: "Demo"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (app.TrayOptions{Title: "Vellum", Tooltip: "Demo", Menu: []app.TrayMenuItem{{Command: "show-services"}}}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (app.FileOptions{Mode: app.SaveFile, Filename: "Vellum-demo.txt", Extensions: []string{"txt"}}).Validate(); err != nil {
		t.Fatal(err)
	}
}
