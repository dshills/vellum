package main

import (
	"bytes"
	"github.com/dshills/vellum/viewport"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestHistoricalStudyPreservesBodiesSelectionCameraAndHistory(t *testing.T) {
	path := filepath.Join("testdata", "compatibility", "study-build22.vforge")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	old, err := decodeStudy(original)
	if err != nil {
		t.Fatal(err)
	}
	m := newWorkbench()
	if err := m.journal.Replace(true, func() error { return m.installStudy(old) }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(old, m.captureStudy()) || m.journal.Snapshot().Dirty || m.journal.Snapshot().UndoCount != 0 || m.scene.Camera.Distance != 12 || len(m.state.Selection) != 2 || m.next != 9 {
		t.Fatal("historical study/history changed")
	}
	shaft := false
	for _, o := range m.state.Objects {
		if o.ID == "shaft" {
			shaft = o.Shape == viewport.CylinderXShape
		}
	}
	if !shaft {
		t.Fatal("round shaft shape lost")
	}
	encoded, err := encodeStudy(m.captureStudy())
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeStudy(encoded)
	if err != nil || !reflect.DeepEqual(old, decoded) {
		t.Fatal("study round trip changed", err)
	}
	if output := os.Getenv("VELLUM_COMPAT_ROUNDTRIP"); output != "" {
		if err := os.WriteFile(filepath.Join(output, "study-build22.vforge"), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.add(); err != nil {
		t.Fatal(err)
	}
	if m.state.Selected != "body-9" {
		t.Fatal("historical identity reused")
	}
	if err := m.journal.Undo(); err != nil || len(m.state.Objects) != 8 {
		t.Fatal("Undo", err)
	}
	if err := m.journal.Redo(); err != nil || len(m.state.Objects) != 9 {
		t.Fatal("Redo", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(original, after) {
		t.Fatal("historical source file overwritten", err)
	}
}
