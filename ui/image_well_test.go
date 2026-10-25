package ui

import (
	"github.com/dshills/vellum/canvas"
	"image"
	"path/filepath"
	"testing"
)

func TestImageWellControlledPathsAndStaleRequests(t *testing.T) {
	art, _ := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 4, 4)))
	calls := 0
	w := &ImageWell{ID: "well", Label: "Artwork", Image: art, Request: func(e ImageWellEvent) bool { calls++; return e.Action == ImageWellDrop }}
	n, err := w.View()
	if err != nil {
		t.Fatal(err)
	}
	drop := n.Children[1].FileDrop
	directory := t.TempDir()
	path := func(name string) string { return filepath.Join(directory, name) }
	for _, paths := range [][]string{{path("a.txt")}, {"relative.png"}, {path("a.png"), path("b.png")}} {
		if drop(FileDropEvent{Paths: paths}) {
			t.Fatal("invalid drop admitted")
		}
	}
	if calls != 0 || !drop(FileDropEvent{Paths: []string{path("A.PNG")}}) || calls != 1 {
		t.Fatal("path request", calls)
	}
	if n.Children[1].Children[0].Image != art {
		t.Fatal("color source replaced")
	}
	w.Busy = true
	drop(FileDropEvent{Paths: []string{path("a.png")}})
	if calls != 1 {
		t.Fatal("busy drop admitted")
	}
	w.Busy = false
	w.ID = "replacement"
	n.Children[2].Children[0].Activate()
	if calls != 1 {
		t.Fatal("old identity acted")
	}
}
