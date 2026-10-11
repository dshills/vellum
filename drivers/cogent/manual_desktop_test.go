//go:build !offscreen

package cogent

import (
	"context"
	"errors"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func TestManualDesktopRefusesBeforeConstruction(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	views := 0
	if err = a.Open(app.Window{ID: "one", Title: "Refused manual", Width: 320, Height: 200, View: func() ui.Node { views++; return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	if _, err = NewManual(context.Background(), a, ManualConfig{}); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal(err)
	}
	if views != 0 || a.Context().Err() != nil || len(a.Windows()) != 1 || a.Pending().DirtyWindows != 1 {
		t.Fatal("desktop refusal changed app or constructed widgets")
	}
}
