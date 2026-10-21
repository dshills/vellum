package main

import (
	"maps"
	"slices"
	"testing"
)

func TestBulkLayersHistoryAndRoundTrip(t *testing.T) {
	for _, action := range []string{"hide-selected", "lock-selected"} {
		t.Run(action, func(t *testing.T) {
			m := groupedModel(t)
			m.setSelection([]string{"block-1", "connector-2"}, "block-1")
			m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Housing"}, "group-1": {Name: "Assembly"}}
			before := slices.Clone(m.scene.Objects)
			ids := m.selectedIDs()
			prior := maps.Clone(m.layerInfo)
			n := m.journal.Snapshot().UndoCount
			if err := m.bulkLayers(action); err != nil {
				t.Fatal(err)
			}
			if len(m.selectedIDs()) != 0 || m.journal.Snapshot().UndoCount != n+1 || !slices.Equal(before, m.scene.Objects) {
				t.Fatal("bulk semantics")
			}
			after := maps.Clone(m.layerInfo)
			f, err := drawingCodec(m).Capture()
			if err != nil {
				t.Fatal(err)
			}
			b, err := drawingCodec(m).Encode(f)
			if err != nil {
				t.Fatal(err)
			}
			d, err := drawingCodec(m).Decode(b)
			if err != nil || !maps.Equal(d.Layers, after) {
				t.Fatal("roundtrip", err)
			}
			if err = m.journal.Undo(); err != nil || !maps.Equal(m.layerInfo, prior) || !slices.Equal(ids, m.selectedIDs()) {
				t.Fatal("undo", err)
			}
			if err = m.journal.Redo(); err != nil || !maps.Equal(m.layerInfo, after) {
				t.Fatal("redo", err)
			}
		})
	}
}
func TestBulkLayersAllClearsImportedMemberFlagsAndNoOps(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"group-1": {Hidden: true, Locked: true, Name: "A"}, "block-1": {Hidden: true, Locked: true, Name: "B"}}
	for _, action := range []string{"show-all", "unlock-all"} {
		n := m.journal.Snapshot().UndoCount
		if err := m.bulkLayers(action); err != nil {
			t.Fatal(err)
		}
		if m.journal.Snapshot().UndoCount != n+1 {
			t.Fatal("one change")
		}
		if err := m.bulkLayers(action); err != nil || m.journal.Snapshot().UndoCount != n+1 {
			t.Fatal("no-op", err)
		}
	}
	if m.layerHidden("block-1") || m.layerLocked("block-1") || m.layerInfo["block-1"].Name != "B" {
		t.Fatal("flags/names")
	}
	m.prompt = "help-guide"
	before := maps.Clone(m.layerInfo)
	if m.bulkLayers("hide-selected") == nil || !maps.Equal(m.layerInfo, before) {
		t.Fatal("busy")
	}
	m.prompt = ""
	m.setSelection(nil, "")
	if m.bulkLayers("lock-selected") == nil {
		t.Fatal("empty")
	}
	if m.bulkLayers("bad") == nil {
		t.Fatal("unknown")
	}
}
