package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
)

type fileFixture struct {
	m    *model
	a    *app.Application
	ref  app.WindowRef
	path string
}

func fileModel(t *testing.T, canceled bool) *fileFixture {
	t.Helper()
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.palette.Close)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	ref, err := a.OpenWindow(app.Window{ID: "placement", Title: "File test", Width: 800, Height: 600, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	if err != nil {
		t.Fatal(err)
	}
	f := &fileFixture{m: m, a: a, ref: ref, path: filepath.Join(t.TempDir(), "round trip.vellum")}
	f.owner(t, func() {
		if err := m.bindFiles(a, ref, func(context.Context, app.FileOptions) (string, error) {
			if canceled {
				return "", app.ErrCanceled
			}
			return f.path, nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	return f
}
func (f *fileFixture) owner(t *testing.T, fn func()) {
	t.Helper()
	if err := f.a.Do(fn); err != nil {
		t.Fatal(err)
	}
}
func (f *fileFixture) wait(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := f.a.Drain(); err != nil {
			t.Fatal(err)
		}
		done := false
		f.owner(t, func() { s := f.m.files.State(); done = !s.Busy && !s.WorkerPending && !s.OwnerPending })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("file operation did not settle")
}
func (f *fileFixture) place(t *testing.T, kind string, p canvas.Point) {
	t.Helper()
	f.owner(t, func() {
		if err := f.m.palette.Source(kind).PlaceKeyboard(f.ref.Context(), f.m.target, p); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDrawingFileRoundTripAndRestoredHistory(t *testing.T) {
	f := fileModel(t, false)
	for i, kind := range []string{"block", "connector", "note"} {
		f.place(t, kind, canvas.Point{X: float64(i*100 + 33), Y: float64(i*-60 - 27)})
	}
	var saved drawingFile
	f.owner(t, func() {
		f.m.scene.View.Center = canvas.Point{X: 129.25, Y: -48.5}
		f.m.scene.View.Zoom = 1.75
		f.m.snap = false
		var err error
		saved, err = drawingCodec(f.m).Capture()
		if err != nil {
			t.Fatal(err)
		}
		if err = f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if s := f.m.files.State(); s.Last.Err != nil || s.Dirty || s.Path != f.path {
			t.Fatal(s)
		}
	})
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := drawingCodec(f.m).Decode(data)
	if err != nil || !reflect.DeepEqual(decoded, saved) {
		t.Fatal("disk snapshot differs", err, decoded, saved)
	}
	f.place(t, "block", canvas.Point{X: 500, Y: 300})
	f.owner(t, func() {
		f.m.scene.View.Center = canvas.Point{}
		f.m.scene.View.Zoom = .5
		f.m.scene.View.Width = 1234
		f.m.scene.View.Height = 567
		f.m.snap = true
		if err := f.m.fileAction(f.a, f.ref, "open", false); err != nil {
			t.Fatal(err)
		}
		if f.m.prompt != "open" || f.m.fileBusy() || len(f.m.scene.Objects) != 4 {
			t.Fatal("dirty open did not ask before replacement")
		}
		f.m.prompt = ""
		if err := f.m.fileAction(f.a, f.ref, "open", true); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		got, err := drawingCodec(f.m).Capture()
		if err != nil || !reflect.DeepEqual(got, saved) {
			t.Fatal("open did not restore exact drawing", err, got, saved)
		}
		if f.m.scene.View.Width != 1234 || f.m.scene.View.Height != 567 {
			t.Fatal("open restored old window dimensions")
		}
		s := f.m.journal.Snapshot()
		if s.Dirty || s.UndoCount != 0 || s.RedoCount != 0 {
			t.Fatal("open retained previous document history", s)
		}
	})
	f.place(t, "note", canvas.Point{X: 700, Y: 200})
	f.owner(t, func() {
		if f.m.scene.Objects[3].ID != "note-4" {
			t.Fatal("restored object counter collided")
		}
		if err := f.m.journal.Undo(); err != nil {
			t.Fatal(err)
		}
		if len(f.m.scene.Objects) != 3 {
			t.Fatal("undo after load")
		}
		if err := f.m.journal.Redo(); err != nil || f.m.scene.Objects[3].ID != "note-4" {
			t.Fatal("redo after load", err)
		}
	})
}

func TestInvalidDrawingRefusesWithoutChangingModel(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{X: 20, Y: 40})
	original, _ := drawingCodec(f.m).Capture()
	encoded, _ := drawingCodec(f.m).Encode(original)
	for _, name := range []string{"version", "kind", "duplicate", "coordinates", "counter", "selected", "zoom", "center"} {
		t.Run(name, func(t *testing.T) {
			bad := original
			bad.Objects = append([]drawingObject(nil), original.Objects...)
			switch name {
			case "version":
				bad.Version = 4
			case "kind":
				bad.Objects[0].Kind = "foreign"
			case "duplicate":
				bad.Objects = append(bad.Objects, bad.Objects[0])
			case "coordinates":
				bad.Objects[0].X = 10001
			case "counter":
				bad.NextID = 0
			case "selected":
				bad.Selected = "note-999"
			case "zoom":
				bad.Zoom = 0
			case "center":
				bad.Center.X = math.Inf(1)
			}
			if err := drawingCodec(f.m).Install(bad); err == nil {
				t.Fatal("invalid model installed")
			}
			got, _ := drawingCodec(f.m).Capture()
			if !reflect.DeepEqual(got, original) {
				t.Fatal("invalid install changed model")
			}
		})
	}
	for _, data := range [][]byte{[]byte("not JSON"), append(append([]byte{}, encoded...), []byte("{}")...), bytes.Replace(encoded, []byte(`"version": 2`), []byte(`"unexpected": 1, "version": 2`), 1), []byte(strings.Repeat(" ", drawingByteLimit+1))} {
		if _, err := drawingCodec(f.m).Decode(data); err == nil {
			t.Fatal("malformed/oversize file accepted")
		}
	}
	if err := os.WriteFile(f.path, []byte(`{"format":"not-vellum","version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{Path: f.path, DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		got, _ := drawingCodec(f.m).Capture()
		if f.m.files.State().Last.Err == nil || !reflect.DeepEqual(got, original) || f.m.journal.Snapshot().UndoCount != 1 {
			t.Fatal("failed open changed drawing/history")
		}
	})
}

func TestCanceledSaveAndFailedWriteKeepDirtyDrawing(t *testing.T) {
	f := fileModel(t, true)
	f.place(t, "note", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		s := f.m.files.State()
		if !errors.Is(s.Last.Err, app.ErrCanceled) || !s.Dirty || s.Path != "" || len(f.m.scene.Objects) != 1 {
			t.Fatal(s)
		}
	})
	if _, err := os.Stat(f.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("canceled save wrote a file", err)
	}
	f.owner(t, func() {
		if err := f.m.files.Save(f.ref.Context(), document.SaveOptions{Path: filepath.Join(filepath.Dir(f.path), "missing", "draw.vellum")}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if s := f.m.files.State(); s.Last.Err == nil || !s.Dirty || s.Path != "" {
			t.Fatal(s)
		}
	})
}

func TestCaptureIsDetachedAndFormatIsVersioned(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "connector", canvas.Point{})
	snapshot, _ := drawingCodec(f.m).Capture()
	f.m.scene.Objects[0].X = 999
	data, err := drawingCodec(f.m).Encode(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err = json.Unmarshal(data, &raw); err != nil || raw["format"] != "vellum-placement" || raw["version"] != float64(2) {
		t.Fatal(err, raw)
	}
	decoded, err := drawingCodec(f.m).Decode(data)
	if err != nil || decoded.Objects[0].X != 0 {
		t.Fatal("snapshot retained mutable owner data", err)
	}
}

func TestExhaustedIDsRemainSaveable(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		codec := drawingCodec(f.m)
		f.m.next = math.MaxUint64 - 1
		if err := f.m.palette.Source("block").PlaceKeyboard(f.ref.Context(), f.m.target, canvas.Point{}); err != nil {
			t.Fatal(err)
		}
		for _, undo := range []bool{false, true} {
			if undo {
				if err := f.m.journal.Undo(); err != nil {
					t.Fatal(err)
				}
			}
			image, err := codec.Capture()
			if err != nil {
				t.Fatal("exhausted drawing cannot be saved", err)
			}
			data, err := codec.Encode(image)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := codec.Decode(data)
			if err != nil || decoded.NextID != math.MaxUint64 {
				t.Fatal("exhausted counter did not round trip", err)
			}
		}
		if err := f.m.palette.Source("note").PlaceKeyboard(f.ref.Context(), f.m.target, canvas.Point{}); err == nil {
			t.Fatal("exhausted counter reused an ID")
		}
	})
}
