package main

import (
	"errors"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"testing"
)

func saveTestView(t *testing.T, m *model, name string) {
	t.Helper()
	if err := m.beginSavedViews(); err != nil {
		t.Fatal(err)
	}
	d := m.viewsDraft
	d.name = name
	m.prompt = ""
	if err := m.applySavedViews(d, "save"); err != nil {
		t.Fatal(err)
	}
	m.viewsDraft = nil
}
func TestSavedViewsHistoryRoundTripRestoreAndLegacy(t *testing.T) {
	m := editModel(t)
	m.scene.View.Center = canvas.Point{X: 123.5, Y: -44}
	m.scene.View.Zoom = 2.5
	n := m.journal.Snapshot().UndoCount
	saveTestView(t, m, "Detail")
	if len(m.savedViews) != 1 || m.journal.Snapshot().UndoCount != n+1 {
		t.Fatal("save history")
	}
	original := slices.Clone(m.savedViews)
	if err := m.journal.Undo(); err != nil || len(m.savedViews) != 0 {
		t.Fatal("undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(original, m.savedViews) {
		t.Fatal("redo", err)
	}
	f, err := drawingCodec(m).Capture()
	if err != nil || f.Version != 6 {
		t.Fatal("capture", err)
	}
	data, err := drawingCodec(m).Encode(f)
	if err != nil {
		t.Fatal(err)
	}
	other := editModel(t)
	decoded, err := drawingCodec(other).Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	other.scene.View.Width = 987
	if err = drawingCodec(other).Install(decoded); err != nil || !slices.Equal(other.savedViews, original) || other.scene.View.Width != 987 {
		t.Fatal("roundtrip", err)
	}
	m.scene.View.Center = canvas.Point{}
	m.scene.View.Zoom = 1
	state := m.journal.Snapshot()
	ids := m.selectedIDs()
	_ = m.beginSavedViews()
	d := m.viewsDraft
	m.prompt = ""
	if err = m.applySavedViews(d, "restore"); err != nil || m.scene.View.Center != original[0].Center || m.scene.View.Zoom != 2.5 || state != m.journal.Snapshot() || !slices.Equal(ids, m.selectedIDs()) {
		t.Fatal("restore", err)
	}
	m.viewsDraft = nil
	saveTestView(t, m, "Detail")
	if m.journal.Snapshot() != state {
		t.Fatal("no-op history")
	}
	m.scene.View.Zoom = 4
	saveTestView(t, m, "Detail")
	if len(m.savedViews) != 1 || m.savedViews[0].Zoom != 4 {
		t.Fatal("upsert")
	}
	_ = m.beginSavedViews()
	d = m.viewsDraft
	m.prompt = ""
	if err = m.applySavedViews(d, "remove"); err != nil || len(m.savedViews) != 0 {
		t.Fatal("remove", err)
	}
	if err = m.journal.Undo(); err != nil || m.savedViews[0].Zoom != 4 {
		t.Fatal("remove undo", err)
	}
	f.Version = 5
	if _, err = f.scene(); err == nil {
		t.Fatal("legacy with views")
	}
	f.Views = nil
	if err = drawingCodec(m).Install(f); err != nil || len(m.savedViews) != 0 {
		t.Fatal("legacy clears views", err)
	}
}
func TestSavedViewsValidationCapacityAndStale(t *testing.T) {
	for _, v := range []savedView{{Name: "", Zoom: 1}, {Name: "X", Zoom: math.NaN()}, {Name: "X", Zoom: 1, Center: canvas.Point{X: math.Inf(1)}}} {
		if validSavedViews([]savedView{v}) == nil {
			t.Fatal("invalid", v)
		}
	}
	m := editModel(t)
	for i := 0; i < 16; i++ {
		saveTestView(t, m, fmt.Sprint(i))
	}
	before := m.journal.Snapshot()
	_ = m.beginSavedViews()
	d := m.viewsDraft
	d.name = "overflow"
	m.prompt = ""
	if m.applySavedViews(d, "save") == nil || m.journal.Snapshot() != before || len(m.savedViews) != 16 {
		t.Fatal("capacity")
	}
	m.viewsDraft = nil
	_ = m.beginSavedViews()
	d = m.viewsDraft
	m.prompt = ""
	m.savedViews[0].Name = "external"
	if !errors.Is(m.applySavedViews(d, "remove"), history.ErrStale) {
		t.Fatal("stale")
	}
}
func TestSavedViewsDialogCancelRetiredCallbacks(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		_ = f.m.beginSavedViews()
		d := f.m.savedViewsDialog(f.a, f.ref)
		geometryDialogControl(t, d, "view-name").Change(ui.ChangeEvent{Text: "Detail"})
		geometryDialogControl(t, d, "views-cancel").Activate()
		f.m.prompt = "close"
		geometryDialogControl(t, d, "view-save").Activate()
		if len(f.m.savedViews) != 0 || f.m.prompt != "close" {
			t.Fatal("retired")
		}
	})
}
