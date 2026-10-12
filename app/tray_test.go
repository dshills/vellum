package app

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"strings"
	"testing"
)

func TestTrayOptionsBoundsAndPNGHeader(t *testing.T) {
	var icon bytes.Buffer
	if err := png.Encode(&icon, image.NewNRGBA(image.Rect(0, 0, 2, 1))); err != nil {
		t.Fatal(err)
	}
	if err := (TrayOptions{IconPNG: icon.Bytes(), Tooltip: "Vellum status", Template: true, Menu: []TrayMenuItem{{Command: "show"}, {Separator: true}}}).Validate(); err != nil {
		t.Fatal(err)
	}
	for _, o := range []TrayOptions{{}, {Title: "a\x00b"}, {Title: "\xff"}, {Title: "Vellum", Template: true}, {Title: "Vellum", IconPNG: []byte("not PNG")}, {Title: "Vellum", Menu: []TrayMenuItem{{}}}, {Title: "Vellum", Menu: []TrayMenuItem{{Separator: true, Command: "show"}}}} {
		if o.Validate() == nil {
			t.Fatal("invalid tray options accepted", o.Title)
		}
	}
	if err := (TrayOptions{IconPNG: icon.Bytes()}).Validate(); err == nil {
		t.Fatal("unnamed icon-only tray accepted")
	}
	for _, o := range []TrayOptions{{Title: strings.Repeat("a", 257)}, {Title: "Vellum", Tooltip: strings.Repeat("a", 1025)}, {Title: "Vellum", IconPNG: make([]byte, TrayMaxPNGBytes+1)}, {Title: "Vellum", Menu: make([]TrayMenuItem, 129)}} {
		if err := o.Validate(); !errors.Is(err, ErrServiceTooLarge) {
			t.Fatal(err)
		}
	}
	var tall bytes.Buffer
	png.Encode(&tall, image.NewNRGBA(image.Rect(0, 0, 1, 257)))
	if err := (TrayOptions{Title: "Vellum", IconPNG: tall.Bytes()}).Validate(); !errors.Is(err, ErrServiceTooLarge) {
		t.Fatal(err)
	}
}

type trayFake struct{ legacyClipboardServices }

func (trayFake) TrayCapabilities() TrayCapabilities         { return TrayCapabilities{StatusItem: true} }
func (trayFake) SetTray(context.Context, TrayOptions) error { return nil }
func (trayFake) ClearTray(context.Context) error            { return nil }
func TestTrayOptionalCompatibility(t *testing.T) {
	a, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if _, err = a.Tray(); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	a.SetServices(legacyClipboardServices{})
	if _, err = a.Tray(); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	a.SetServices(trayFake{})
	s, err := a.Tray()
	if err != nil || !s.TrayCapabilities().StatusItem {
		t.Fatal(err)
	}
	a.Close()
	if _, err = a.Tray(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
