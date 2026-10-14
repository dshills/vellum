package app

import (
	"github.com/dshills/vellum/ui"
	"runtime"
	"strconv"
	"testing"
	"weak"
)

type closedWindowPayload struct {
	id   int
	data [1 << 20]byte
}

//go:noinline
func openRetainedWindow(a *Application) (weak.Pointer[closedWindowPayload], error) {
	payload := &closedWindowPayload{id: 42}
	err := a.Open(Window{ID: "temporary", Title: "Temporary", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: strconv.Itoa(payload.id)} }})
	return weak.Make(payload), err
}
func TestClosedWindowReleasesViewCaptureWithSurvivor(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	if err := a.Open(sampleWindow("survivor")); err != nil {
		t.Fatal(err)
	}
	reference, err := openRetainedWindow(a)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.CloseWindow("temporary"); err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	if reference.Value() != nil {
		t.Fatal("closed View capture retained while survivor remains")
	}
	runtime.KeepAlive(a)
	if windows := a.Windows(); len(windows) != 1 || windows[0].ID != "survivor" {
		t.Fatal("surviving registration changed")
	}
}
