package app

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"
)

func TestAlertOptionsBoundsAndRoles(t *testing.T) {
	o := AlertOptions{Title: "Confirm", Buttons: []AlertButton{{ID: "ok", Label: "OK", Role: AlertDefault}, {ID: "cancel", Label: "Cancel", Role: AlertCancel}, {ID: "delete", Label: "Delete", Role: AlertDestructive}}}
	if err := o.Validate(); err != nil {
		t.Fatal(err)
	}
	var pngData bytes.Buffer
	png.Encode(&pngData, image.NewNRGBA(image.Rect(0, 0, 8, 8)))
	o.IconPNG = pngData.Bytes()
	if o.Validate() != nil {
		t.Fatal("color icon rejected")
	}
	o.Buttons[1].Role = AlertDefault
	if o.Validate() == nil {
		t.Fatal("multiple defaults admitted")
	}
	o.Buttons[1].Role = AlertCancel
	o.IconPNG = make([]byte, 256<<10+1)
	if o.Validate() == nil {
		t.Fatal("oversize icon admitted")
	}
}

type alertTestServices struct{ legacyClipboardServices }

func (alertTestServices) ShowAlert(context.Context, WindowRef, AlertOptions) (AlertResult, error) {
	return AlertResult{ButtonID: "ok"}, nil
}
func TestOptionalAlertService(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	a.SetServices(legacyClipboardServices{})
	if _, err := a.Alerts(); err != ErrUnavailable {
		t.Fatal(err)
	}
	a.SetServices(alertTestServices{})
	if _, err := a.Alerts(); err != nil {
		t.Fatal(err)
	}
}
