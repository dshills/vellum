package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
)

func TestSelectedSVGWholeGroupOrderBoundsAndCameraIndependence(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Housing <&>"}}
	m.setSelection([]string{"note-3"}, "note-3")
	before := m.journal.Snapshot()
	image, err := m.captureSelectedSVG()
	if err != nil {
		t.Fatal(err)
	}
	if len(image) != 2 || image[0].Rect.ID != "block-1" || image[1].Rect.ID != "note-3" {
		t.Fatal("whole group selection or paint order", image)
	}
	data, err := encodeSVG(image)
	if err != nil {
		t.Fatal(err)
	}
	var out parsedSVG
	if err := xml.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.ViewBox != "0 0 490 65" || len(out.Rects) != 2 || out.Rects[0].Title != m.layerLabel("block-1") || out.Rects[0].Fill != "#3075d1" {
		t.Fatal("selection bounds, names or colors", out)
	}
	m.scene.View.Center = canvas.Point{X: 930, Y: -230}
	m.scene.View.Zoom = 4
	m.gridHidden, m.gridSpacing = true, 100
	again, err := m.captureSelectedSVG()
	if err != nil {
		t.Fatal(err)
	}
	againData, err := encodeSVG(again)
	if err != nil || !bytes.Equal(data, againData) {
		t.Fatal("camera/grid changed selected SVG", err)
	}
	image[0].Rect.X = 999
	image[0].Name = "changed"
	if m.scene.Objects[0].X != 0 || m.layerLabel("block-1") == "changed" || m.journal.Snapshot() != before {
		t.Fatal("capture changed drawing or history")
	}
}

func TestSelectedSVGRejectsMissingHiddenLockedAndBusySelection(t *testing.T) {
	m := groupedModel(t)
	for _, metadata := range []layerMetadata{{Hidden: true}, {Locked: true}} {
		m.layerInfo = map[string]layerMetadata{"note-3": metadata}
		if _, err := m.captureSelectedSVG(); err == nil {
			t.Fatal("hidden/locked group exported")
		}
	}
	m.layerInfo = nil
	m.setSelection(nil, "")
	if m.selectedSVGAvailable() == nil {
		t.Fatal("empty selection admitted")
	}
	m.setSelection([]string{"connector-2"}, "connector-2")
	m.prompt = "save"
	if m.selectedSVGAvailable() == nil {
		t.Fatal("busy selection admitted")
	}
	m.prompt = ""
	m.layerInfo = map[string]layerMetadata{"note-3": {Hidden: true}}
	image, err := m.captureSelectedSVG()
	if err != nil || len(image) != 1 || image[0].Rect.ID != "connector-2" {
		t.Fatal("hidden and unselected objects leaked", image, err)
	}
}

func TestSelectedSVGDetachedChooserFreezesSnapshotAndPreservesActiveDrawing(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{X: -40, Y: 40})
	f.place(t, "note", canvas.Point{X: 200, Y: 100})
	f.owner(t, func() { _ = f.m.fileAction(f.a, f.ref, "save", false) })
	f.wait(t)
	active := f.path
	dest := filepath.Join(t.TempDir(), "selection.svg")
	entered, release := make(chan app.FileOptions, 1), make(chan struct{})
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1"}, "block-1")
		f.m.chooseFile = func(ctx context.Context, options app.FileOptions) (string, error) {
			entered <- options
			select {
			case <-release:
				return dest, nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		if err := f.m.exportSelectedSVG(f.a, f.ref, nil); err != nil {
			t.Fatal(err)
		}
	})
	var options app.FileOptions
	select {
	case options = <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("selected export chooser did not start")
	}
	if options.Filename != "round trip Selection" || options.Directory != filepath.Dir(active) || options.Mode != app.SaveFile {
		t.Fatal("selected export file suggestion", options)
	}
	var before drawingFile
	var state document.State
	f.owner(t, func() {
		// Deliberately mutate the source after admission to prove the worker has
		// a detached image; normal commands are gated while the chooser is open.
		f.m.scene.Objects[0].X = 880
		f.m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Later name"}}
		f.m.setSelection([]string{"note-2"}, "note-2")
		before, _ = drawingCodec(f.m).Capture()
		state = f.m.files.State()
	})
	journal := f.m.journal.Snapshot()
	close(release)
	waitOutput(t, f)
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	var out parsedSVG
	if err := xml.Unmarshal(data, &out); err != nil || len(out.Rects) != 1 || out.Rects[0].ID != "block-1" || out.Rects[0].X != "-40" || out.Rects[0].Title != "block-1" || out.ViewBox != "-40 40 80 50" {
		t.Fatal("chooser did not retain original selection snapshot", out, err)
	}
	f.owner(t, func() {
		after, _ := drawingCodec(f.m).Capture()
		if !reflect.DeepEqual(before, after) || f.m.files.State() != state || f.m.journal.Snapshot() != journal {
			t.Fatal("export changed active drawing, path or history")
		}
	})
}

func TestSelectedSVGCancelAndWriteFailurePreserveDrawing(t *testing.T) {
	for _, canceled := range []bool{true, false} {
		t.Run(map[bool]string{true: "cancel", false: "write failure"}[canceled], func(t *testing.T) {
			f := fileModel(t, canceled)
			f.place(t, "block", canvas.Point{})
			var before drawingFile
			var state document.State
			f.owner(t, func() {
				f.m.setSelection([]string{"block-1"}, "block-1")
				before, _ = drawingCodec(f.m).Capture()
				state = f.m.files.State()
			})
			journal := f.m.journal.Snapshot()
			f.owner(t, func() {
				var store document.Store
				if !canceled {
					store = failingOutputStore{}
				}
				if err := f.m.exportSelectedSVG(f.a, f.ref, store); err != nil {
					t.Fatal(err)
				}
			})
			waitOutput(t, f)
			f.owner(t, func() {
				after, _ := drawingCodec(f.m).Capture()
				if !reflect.DeepEqual(before, after) || f.m.files.State() != state || f.m.journal.Snapshot() != journal {
					t.Fatal("canceled/failed export changed drawing")
				}
			})
		})
	}
}
