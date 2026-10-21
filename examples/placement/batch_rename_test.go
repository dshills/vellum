package main

import (
	"errors"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestBatchRenameGroupUnitsHistoryAndRoundTrip(t *testing.T) {
	m := groupedModel(t)
	m.setSelection([]string{"block-1", "connector-2"}, "block-1")
	m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Member"}}
	before := slices.Clone(m.scene.Objects)
	ids := m.selectedIDs()
	n := m.journal.Snapshot().UndoCount
	if err := m.beginBatchRename(); err != nil {
		t.Fatal(err)
	}
	d := m.batchRename
	d.prefix = "Part"
	d.start = "9"
	names, preview, err := m.batchNames(d)
	if err != nil || names["group-1"] != "Part 9" || names["connector-2"] != "Part 10" || len(preview) != 2 {
		t.Fatal(names, err)
	}
	m.prompt = ""
	if err = m.applyBatchRename(d); err != nil {
		t.Fatal(err)
	}
	after := maps.Clone(m.layerInfo)
	if m.layerInfo["block-1"].Name != "Member" || !slices.Equal(before, m.scene.Objects) || !slices.Equal(ids, m.selectedIDs()) || m.journal.Snapshot().UndoCount != n+1 {
		t.Fatal("changed geometry/selection/member")
	}
	if err = m.journal.Undo(); err != nil || m.layerInfo["group-1"].Name != "" {
		t.Fatal("undo", err)
	}
	if err = m.journal.Redo(); err != nil || !maps.Equal(after, m.layerInfo) {
		t.Fatal("redo", err)
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	b, err := drawingCodec(m).Encode(f)
	if err != nil {
		t.Fatal(err)
	}
	v, err := drawingCodec(m).Decode(b)
	if err != nil || !maps.Equal(v.Layers, after) {
		t.Fatal("roundtrip", err)
	}
}
func TestBatchRenameAtomicInvalidAndStale(t *testing.T) {
	for _, prefix := range []string{"", "bad\nname", strings.Repeat("a", 80)} {
		m := groupedModel(t)
		_ = m.beginBatchRename()
		d := m.batchRename
		d.prefix = prefix
		m.prompt = ""
		state := m.journal.Snapshot()
		if m.applyBatchRename(d) == nil || state != m.journal.Snapshot() || len(m.layerInfo) != 0 {
			t.Fatal("invalid partial edit")
		}
	}
	m := groupedModel(t)
	_ = m.beginBatchRename()
	d := m.batchRename
	m.prompt = ""
	m.setSelection([]string{"connector-2"}, "connector-2")
	if !errors.Is(m.applyBatchRename(d), history.ErrStale) {
		t.Fatal("stale")
	}
}
func TestBatchRenameCancelRetiredAndPreview(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		_ = f.m.beginBatchRename()
		d := f.m.batchRenameDialog(f.a, f.ref)
		geometryDialogControl(t, d, "batch-prefix").Change(ui.ChangeEvent{Text: "Name"})
		d = f.m.batchRenameDialog(f.a, f.ref)
		if !strings.Contains(geometryDialogControl(t, d, "batch-preview").Text, "Name 1") {
			t.Fatal("preview")
		}
		geometryDialogControl(t, d, "batch-cancel").Activate()
		f.m.prompt = "close"
		geometryDialogControl(t, d, "batch-apply").Activate()
		if len(f.m.layerInfo) != 0 || f.m.prompt != "close" {
			t.Fatal("retired")
		}
	})
}
