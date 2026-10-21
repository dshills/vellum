package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestPNGPaintOrderBoundsScaleAndBackground(t *testing.T) {
	objects := []canvas.Rect{{ID: "red", X: -.25, Y: -.5, Width: 10, Height: 10, Color: color.RGBA{255, 0, 0, 255}}, {ID: "blue", X: 4.75, Y: 4.5, Width: 10, Height: 10, Color: color.RGBA{0, 0, 255, 255}}}
	for _, scale := range []int{1, 2, 4} {
		for _, bg := range []string{"transparent", "white", "dark"} {
			b, err := encodePNG(pngImage{objects, scale, bg})
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			if im.Bounds().Dx() != 15*scale || im.Bounds().Dy() != 15*scale {
				t.Fatal("bounds", im.Bounds())
			}
			red := color.NRGBAModel.Convert(im.At(0, 0)).(color.NRGBA)
			blue := color.NRGBAModel.Convert(im.At(6*scale, 6*scale)).(color.NRGBA)
			corner := color.NRGBAModel.Convert(im.At(14*scale, 0)).(color.NRGBA)
			if red != (color.NRGBA{255, 0, 0, 255}) || blue != (color.NRGBA{0, 0, 255, 255}) {
				t.Fatal("paint order", red, blue)
			}
			want := color.NRGBA{}
			if bg == "white" {
				want = color.NRGBA{255, 255, 255, 255}
			}
			if bg == "dark" {
				want = color.NRGBA{32, 35, 40, 255}
			}
			if corner != want {
				t.Fatal("background", corner, want)
			}
		}
	}
}
func TestPNGRejectsOversizeAndInvalidBeforeAllocation(t *testing.T) {
	for _, p := range []pngImage{{scale: 1, background: "white"}, {objects: []canvas.Rect{{ID: "a", Width: 2000, Height: 2000}, {ID: "b", X: 10000, Y: 10000, Width: 2000, Height: 2000}}, scale: 1, background: "white"}, {objects: []canvas.Rect{{ID: "a", Width: 2000, Height: 2000}}, scale: 4, background: "white"}, {objects: []canvas.Rect{{ID: "a", Width: 10, Height: 10}}, scale: 3, background: "white"}} {
		if _, err := encodePNG(p); err == nil {
			t.Fatal("invalid PNG admitted")
		}
	}
	var b pngBuffer
	if _, err := b.Write(make([]byte, pngByteLimit+1)); !errors.Is(err, document.ErrTooLarge) || b.Len() != 0 {
		t.Fatal("byte limit", err)
	}
}
func TestPNGSnapshotSelectionGroupsVisibilityAndStale(t *testing.T) {
	m := groupedModel(t)
	m.setSelection([]string{"block-1"}, "block-1")
	if err := m.beginPNG(); err != nil {
		t.Fatal(err)
	}
	d := m.pngDraft
	d.selectionOnly = true
	p, err := m.pngSnapshot(d)
	if err != nil || len(p.objects) != 2 || p.objects[0].ID != "block-1" || p.objects[1].ID != "note-3" {
		t.Fatal("group selection", err)
	}
	d.selectionOnly = false
	p, err = m.pngSnapshot(d)
	if err != nil || len(p.objects) != 3 {
		t.Fatal("all", err)
	}
	m.prompt = ""
	m.layerInfo = map[string]layerMetadata{"connector-2": {Hidden: true}}
	if _, err = m.pngSnapshot(d); !errors.Is(err, history.ErrStale) {
		t.Fatal("stale")
	}
	m.pngDraft = nil
	if err = m.beginPNG(); err != nil {
		t.Fatal(err)
	}
	p, err = m.pngSnapshot(m.pngDraft)
	if err != nil || len(p.objects) != 2 {
		t.Fatal("hidden", err)
	}
}
func TestPNGDetachedChooserFreezeAndDrawingPreservation(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{X: 0, Y: 0})
	dest := filepath.Join(t.TempDir(), "Drawing.png")
	entered, release := make(chan struct{}), make(chan struct{})
	f.owner(t, func() {
		f.m.chooseFile = func(ctx context.Context, o app.FileOptions) (string, error) {
			if o.Extensions[0] != "png" {
				return "", errors.New("wrong extension")
			}
			close(entered)
			select {
			case <-release:
				return dest, nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		p := pngImage{objects: slices.Clone(f.m.scene.Objects), scale: 2, background: "transparent"}
		if err := f.m.exportPNG(f.a, f.ref, nil, p); err != nil {
			t.Fatal(err)
		}
	})
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("chooser")
	}
	var before drawingFile
	f.owner(t, func() { f.m.scene.Objects[0].Width = 100; before, _ = drawingCodec(f.m).Capture() })
	state := f.m.journal.Snapshot()
	close(release)
	waitOutput(t, f)
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil || im.Width != 160 || im.Height != 100 {
		t.Fatal("frozen export", im, err)
	}
	f.owner(t, func() {
		after, _ := drawingCodec(f.m).Capture()
		if !reflect.DeepEqual(before, after) || state != f.m.journal.Snapshot() || f.m.files.State().Path != "" {
			t.Fatal("active drawing changed")
		}
	})
}
func TestPNGCancelWriteFailureAndRetiredDialog(t *testing.T) {
	for _, cancel := range []bool{true, false} {
		f := fileModel(t, cancel)
		f.place(t, "block", canvas.Point{})
		before := slices.Clone(f.m.scene.Objects)
		state := f.m.journal.Snapshot()
		f.owner(t, func() {
			var store document.Store
			if !cancel {
				store = failingOutputStore{}
			}
			if err := f.m.exportPNG(f.a, f.ref, store, pngImage{before, 1, "white"}); err != nil {
				t.Fatal(err)
			}
		})
		waitOutput(t, f)
		if !slices.Equal(before, f.m.scene.Objects) || state != f.m.journal.Snapshot() {
			t.Fatal("failure mutated")
		}
	}
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		_ = f.m.beginPNG()
		d := f.m.pngDialog(f.a, f.ref)
		geometryDialogControl(t, d, "png-cancel").Activate()
		f.m.prompt = "close"
		geometryDialogControl(t, d, "png-scale").Change(ui.ChangeEvent{SelectedID: "4"})
		geometryDialogControl(t, d, "png-export").Activate()
		if f.m.prompt != "close" || f.m.pngDraft != nil || f.m.output != nil {
			t.Fatal("retired callback")
		}
	})
}
