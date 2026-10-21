package main

import (
	"bytes"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image/color"
	"slices"
	"testing"
)

func TestFillColorHistoryDrawingClipboardAndSVG(t *testing.T) {
	m := groupedModel(t)
	before := slices.Clone(m.scene.Objects)
	ids := m.selectedIDs()
	n := m.journal.Snapshot().UndoCount
	if err := m.beginFillColor(); err != nil {
		t.Fatal(err)
	}
	d := m.fillDraft
	d.value = "#Ab12Ef"
	m.prompt = ""
	if err := m.applyFillColor(d); err != nil {
		t.Fatal(err)
	}
	want := color.RGBA{171, 18, 239, 255}
	after := slices.Clone(m.scene.Objects)
	for _, o := range after {
		if slices.Contains(ids, o.ID) && o.Color != want {
			t.Fatal("fill", o)
		}
	}
	if m.journal.Snapshot().UndoCount != n+1 {
		t.Fatal("not one edit")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil || f.Version != 5 {
		t.Fatal(f.Version, err)
	}
	data, err := drawingCodec(m).Encode(f)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := drawingCodec(m).Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	scene, err := decoded.scene()
	if err != nil || !slices.Equal(scene.Objects, after) {
		t.Fatal("drawing round trip", err)
	}
	copied, err := m.copiedSelection()
	if err != nil {
		t.Fatal(err)
	}
	wire, err := encodeSelection(copied)
	if err != nil {
		t.Fatal(err)
	}
	copied, err = decodeSelection(wire)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.insertCopies(copied, 20, "Paste selection"); err != nil {
		t.Fatal(err)
	}
	for _, o := range m.scene.Objects[len(after):] {
		if o.Color != want {
			t.Fatal("clipboard fill lost")
		}
	}
	if err = m.journal.Undo(); err != nil {
		t.Fatal(err)
	}
	if err = m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("undo fill", err)
	}
	if err = m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
		t.Fatal("redo fill", err)
	}
	image, err := m.captureSVG()
	if err != nil {
		t.Fatal(err)
	}
	svg, err := encodeSVG(image)
	if err != nil || !bytes.Contains(svg, []byte(`fill="#ab12ef"`)) {
		t.Fatal("SVG fill", err)
	}
	if err = m.renameLayer(ids[0], "Colored"); err != nil {
		t.Fatal("rename colored document", err)
	}
}
func TestFillColorRejectsAndRestoresDefaults(t *testing.T) {
	for _, bad := range []string{"", "red", "#xyzxyz", "#12345678", "#123", "NaN"} {
		m := groupedModel(t)
		before := slices.Clone(m.scene.Objects)
		state := m.journal.Snapshot()
		_ = m.beginFillColor()
		d := m.fillDraft
		d.value = bad
		m.prompt = ""
		if m.applyFillColor(d) == nil || !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() {
			t.Fatal("bad fill mutation", bad)
		}
	}
	m := groupedModel(t)
	_ = m.beginFillColor()
	d := m.fillDraft
	d.value = "#112233"
	m.prompt = ""
	_ = m.applyFillColor(d)
	m.fillDraft = nil
	_ = m.beginFillColor()
	d = m.fillDraft
	d.value = "default"
	m.prompt = ""
	if err := m.applyFillColor(d); err != nil {
		t.Fatal(err)
	}
	f, err := drawingCodec(m).Capture()
	if err != nil || f.Version == 5 {
		t.Fatal("default format compatibility", err)
	}
	n := m.journal.Snapshot().UndoCount
	m.fillDraft = nil
	_ = m.beginFillColor()
	d = m.fillDraft
	d.value = "default"
	m.prompt = ""
	if err = m.applyFillColor(d); err != nil || m.journal.Snapshot().UndoCount != n {
		t.Fatal("no-op history", err)
	}
	m.fillDraft = nil
	_ = m.beginFillColor()
	d = m.fillDraft
	m.prompt = ""
	m.scene.Objects[0].X++
	if m.applyFillColor(d) == nil {
		t.Fatal("stale")
	}
}
func TestFillColorWireCompatibilityAndValidation(t *testing.T) {
	m := groupedModel(t)
	f, _ := drawingCodec(m).Capture()
	f.Objects[0].Color = "#123456"
	if _, err := f.scene(); err == nil {
		t.Fatal("old file admitted custom fill")
	}
	f.Version = 5
	f.Objects[0].Color = "no"
	if _, err := f.scene(); err == nil {
		t.Fatal("malformed fill")
	}
	for _, wire := range []string{`{"format":"vellum-placement-selection","version":2,"objects":[{"kind":"block","x":0,"y":0,"width":80,"height":50,"color":"#123456"}]}`, `{"format":"vellum-placement-selection","version":3,"objects":[{"kind":"block","x":0,"y":0,"width":80,"height":50,"color":"bad"}]}`} {
		if _, err := decodeSelection(wire); err == nil {
			t.Fatal("invalid clipboard fill")
		}
	}
	legacy, err := decodeSelection(`{"format":"vellum-placement-selection","version":1,"objects":[{"kind":"block","x":0,"y":0,"width":80,"height":50}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.insertCopiesOffset(legacy, canvas.Point{}, "legacy"); err != nil {
		t.Fatal(err)
	}
}

func TestFillColorDialogInvalidRetryCancelAndOldCallbacks(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.beginFillColor(); err != nil {
			t.Fatal(err)
		}
		old := f.m.fillColorDialog(f.a, f.ref)
		d := f.m.fillDraft
		epoch := f.m.modalEpoch
		state := f.m.journal.Snapshot()
		geometryDialogControl(t, old, "fill-value").Change(ui.ChangeEvent{Text: "bad"})
		geometryDialogControl(t, old, "fill-apply").Activate()
		if f.m.fillDraft != d || f.m.prompt != "fill-color" || d.message == "" || f.m.modalEpoch != epoch || state != f.m.journal.Snapshot() {
			t.Fatal("invalid draft did not stay open")
		}
		geometryDialogControl(t, old, "fill-value").Change(ui.ChangeEvent{Text: "#123456"})
		geometryDialogControl(t, old, "fill-apply").Activate()
		if f.m.fillDraft != nil || f.m.prompt != "" || f.m.scene.Objects[0].Color != (color.RGBA{18, 52, 86, 255}) {
			t.Fatal("valid retry failed")
		}
		if err := f.m.beginFillColor(); err != nil {
			t.Fatal(err)
		}
		current := f.m.fillDraft
		old.Dismiss()
		geometryDialogControl(t, old, "fill-apply").Activate()
		if f.m.fillDraft != current || f.m.prompt != "fill-color" {
			t.Fatal("old callback changed current draft")
		}
		f.m.fillColorDialog(f.a, f.ref).Dismiss()
		if f.m.fillDraft != nil || f.m.prompt != "" {
			t.Fatal("cancel")
		}
	})
}
