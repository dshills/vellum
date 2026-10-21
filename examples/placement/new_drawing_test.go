package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"os"
	"reflect"
	"testing"
)

func TestNewDrawingCancelReplaceAndSaveDestination(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	disk, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	f.place(t, "note", canvas.Point{X: 200})
	f.owner(t, func() {
		f.m.groups = map[string]string{"block-1": "group-1", "note-2": "group-1"}
		f.m.nextGroup = 1
		f.m.layerInfo = map[string]layerMetadata{"group-1": {Name: "Old group"}}
		f.m.scene.View.Center = canvas.Point{X: 100, Y: 50}
		f.m.scene.View.Zoom = 2
		scene := f.m.scene
		before, _ := drawingCodec(f.m).Capture()
		state := f.m.journal.Snapshot()
		session := f.m.files
		if err := f.m.commands.Invoke(f.ref.Context(), "new"); err != nil {
			t.Fatal(err)
		}
		if f.m.prompt != "new" {
			t.Fatal("missing new prompt")
		}
		old := f.m.filePrompt(f.a, f.ref)
		geometryDialogControl(t, old, "file-keep").Activate()
		after, _ := drawingCodec(f.m).Capture()
		if !reflect.DeepEqual(before, after) || state != f.m.journal.Snapshot() || f.m.files != session {
			t.Fatal("cancel changed document")
		}
		if err := f.m.commands.Invoke(f.ref.Context(), "new"); err != nil {
			t.Fatal(err)
		}
		geometryDialogControl(t, f.m.filePrompt(f.a, f.ref), "file-discard").Activate()
		s := f.m.journal.Snapshot()
		if f.m.scene != scene || len(f.m.groups) != 0 || len(f.m.layerInfo) != 0 || f.m.nextGroup != 0 || f.m.scene.View.Center != (canvas.Point{}) || f.m.scene.View.Zoom != 1 || f.m.layerFocus != "" {
			t.Fatal("old document state survived New")
		}
		if f.m.prompt != "" || f.m.files.State().Path != "" || len(f.m.scene.Objects) != 0 || s.Dirty || s.UndoCount != 0 || s.RedoCount != 0 || f.m.next != 0 || len(f.m.selectedIDs()) != 0 {
			t.Fatal("new did not reset", s)
		}
		// Old callbacks must not dismiss a different prompt.
		f.m.prompt = "close"
		old.Dismiss()
		if f.m.prompt != "close" {
			t.Fatal("stale prompt")
		}
		f.m.prompt = ""
		node := geometryDialogControl(t, view(f.a, f.ref, f.m), "new")
		if node.Kind != ui.Button || node.Disabled {
			t.Fatal("New control")
		}
	})
	now, err := os.ReadFile(f.path)
	if err != nil || !reflect.DeepEqual(disk, now) {
		t.Fatal("New overwrote old drawing", err)
	}
	// A fresh Save must invoke the chooser, rather than overwrite the prior path.
	f.path = f.path + "-new.vellum"
	f.place(t, "connector", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	if _, err := os.Stat(f.path); err != nil {
		t.Fatal("fresh Save did not choose destination", err)
	}
}
func TestNewDrawingGatesAndCleanReplacement(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		before, _ := drawingCodec(f.m).Capture()
		f.m.prompt = "geometry"
		if err := f.m.newDrawing(f.a, f.ref, true); err == nil {
			t.Fatal("prompt admitted New")
		}
		f.m.prompt = ""
		f.m.move = &objectMove{}
		if err := f.m.newDrawing(f.a, f.ref, true); err == nil {
			t.Fatal("gesture admitted New")
		}
		f.m.move = nil
		after, _ := drawingCodec(f.m).Capture()
		if !reflect.DeepEqual(before, after) {
			t.Fatal("rejected New changed drawing")
		}
		if err := f.m.journal.MarkSaved(); err != nil {
			t.Fatal(err)
		}
		if err := f.m.commands.Invoke(f.ref.Context(), "new"); err != nil {
			t.Fatal(err)
		}
		if len(f.m.scene.Objects) != 0 || f.m.prompt != "" {
			t.Fatal("clean New prompted")
		}
	})
}
