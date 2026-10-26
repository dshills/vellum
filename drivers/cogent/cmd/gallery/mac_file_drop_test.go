package main

import (
	"github.com/dshills/vellum/app"
	"path/filepath"
	"testing"
)

func TestMacFileChoiceCancellationPreservesFeedback(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.fileDropZoneView()
	z := m.dropZone
	path := filepath.Join(t.TempDir(), "synthetic.txt")
	if err := z.ProposeFiles([]string{path}); err != nil {
		t.Fatal(err)
	}
	before, feedback := m.count, z.Feedback()
	m.applyFileChoice(z, "", app.ErrCanceled)
	m.applyFileChoice(z, "", nil)
	if m.count != before || z.Feedback() != feedback {
		t.Fatal("cancel proposed an empty path or discarded existing feedback")
	}
	m.applyFileChoice(z, path, nil)
	if m.count != before+1 {
		t.Fatal("chooser did not use common admission path")
	}
}
