package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image/color"
	"math"
	"slices"
	"testing"
)

func TestFillPickerPreviewAtomicApplyCancelAndStale(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		before := slices.Clone(f.m.scene.Objects)
		state := f.m.journal.Snapshot()
		if err := f.m.beginFillColor(); err != nil {
			t.Fatal(err)
		}
		old := f.m.fillColorDialog(f.a, f.ref)
		if fillNodeContains(old, ui.ColorFieldID("fill-picker", ui.ColorAlpha)) || fillNodeContains(old, ui.ColorFieldID("fill-picker", ui.ColorHex)) {
			t.Fatal("unsupported alpha or duplicate hex")
		}
		geometryDialogControl(t, old, "fill-swatch-purple").Activate()
		if !slices.Equal(before, f.m.scene.Objects) || state != f.m.journal.Snapshot() {
			t.Fatal("preview committed")
		}
		red := geometryDialogControl(t, old, ui.ColorFieldID("fill-picker", ui.ColorRed))
		red.Change(ui.ChangeEvent{Number: float32(math.NaN())})
		geometryDialogControl(t, old, "fill-apply").Activate()
		if f.m.prompt != "fill-color" || state != f.m.journal.Snapshot() {
			t.Fatal("invalid slider committed")
		}
		red.Change(ui.ChangeEvent{Number: 12})
		if f.m.fillDraft.value != "#0c5acc" {
			t.Fatal(f.m.fillDraft.value)
		}
		geometryDialogControl(t, old, "fill-apply").Activate()
		after := slices.Clone(f.m.scene.Objects)
		if after[0].Color != (color.RGBA{12, 90, 204, 255}) || f.m.journal.Snapshot().UndoCount != state.UndoCount+1 {
			t.Fatal("apply")
		}
		if err := f.m.journal.Undo(); err != nil || !slices.Equal(before, f.m.scene.Objects) {
			t.Fatal("undo", err)
		}
		if err := f.m.journal.Redo(); err != nil || !slices.Equal(after, f.m.scene.Objects) {
			t.Fatal("redo", err)
		}
		if err := f.m.beginFillColor(); err != nil {
			t.Fatal(err)
		}
		current := f.m.fillDraft
		red.Change(ui.ChangeEvent{Number: 200})
		if current.value != "#0c5acc" {
			t.Fatal("old picker callback")
		}
		dialog := f.m.fillColorDialog(f.a, f.ref)
		geometryDialogControl(t, dialog, "fill-default").Activate()
		dialog.Dismiss()
		if !slices.Equal(after, f.m.scene.Objects) {
			t.Fatal("cancel changed fill")
		}
	})
}

func TestFillPickerMixedSelectionAndHexSync(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200})
	f.owner(t, func() {
		f.m.setSelection([]string{f.m.scene.Objects[0].ID, f.m.scene.Objects[1].ID}, f.m.scene.Objects[0].ID)
		if err := f.m.beginFillColor(); err != nil {
			t.Fatal(err)
		}
		d := f.m.fillDraft
		if d.value != "" || d.picker.Color.R != f.m.scene.Objects[0].Color.R {
			t.Fatal("mixed primary preview")
		}
		dialog := f.m.fillColorDialog(f.a, f.ref)
		geometryDialogControl(t, dialog, "fill-value").Change(ui.ChangeEvent{Text: "#123456"})
		if d.picker.Color != (color.NRGBA{18, 52, 86, 255}) {
			t.Fatal("hex preview")
		}
		geometryDialogControl(t, dialog, "fill-value").Change(ui.ChangeEvent{Text: "default"})
		c := f.m.scene.Objects[0].Color
		if d.picker.Color != (color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}) {
			t.Fatal("typed default preview")
		}
		geometryDialogControl(t, dialog, "fill-swatch-green").Activate()
		geometryDialogControl(t, dialog, "fill-apply").Activate()
		for _, o := range f.m.scene.Objects {
			if o.Color != (color.RGBA{45, 150, 85, 255}) {
				t.Fatal("mixed apply", o)
			}
		}
	})
}
