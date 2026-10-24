package main

import (
	"bytes"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image/png"
	"testing"
)

func TestMacNativeTitlebarPrototypeUsesIndependentFullColorActions(t *testing.T) {
	for _, dark := range []bool{false, true} {
		m := newMacWidgetsGallery(newInputHelpersGallery(), dark)
		a, err := app.New(app.Config{})
		if err != nil {
			t.Fatal(err)
		}
		ref, err := a.OpenWindow(app.Window{ID: "gallery", Title: "Prototype", Width: 660, Height: 740, View: m.view})
		if err != nil {
			t.Fatal(err)
		}
		p, err := m.titlebarPrototype(ref)
		if err != nil {
			t.Fatal(err)
		}
		if p.Window != ref || len(p.Items) != 8 {
			t.Fatal("prototype config")
		}
		m.toolbars.disabled = true
		m.toolbars.fullColor = false
		for _, item := range p.Items {
			if item.Kind != ui.Button {
				continue
			}
			if _, err = png.Decode(bytes.NewReader(item.PNG)); err != nil {
				t.Fatal(err)
			}
			before := m.toolbars.count
			item.Activate()
			if item.ID == "Delete" {
				if m.toolbars.count != before || !item.Disabled {
					t.Fatal("disabled Delete")
				}
			} else if m.toolbars.count != before+1 || m.toolbars.last != "Native titlebar: "+item.Label {
				t.Fatal("native action did not reach shared model")
			}
		}
		if m.toolbars.count != 5 {
			t.Fatal("unexpected action count")
		}
		a.Close()
	}
}
