package main

import (
	"context"
	"maps"
	"slices"
	"testing"
)

func TestLayerLockBlocksEditingAndRestoresSelectionThroughHistory(t *testing.T) {
	m := groupedModel(t)
	before := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	if err := m.toggleLayerLock("block-1"); err != nil {
		t.Fatal(err)
	}
	if !m.layerLocked("note-3") || len(m.selectedIDs()) != 0 || m.scene.Selected != "" {
		t.Fatal("group not protected")
	}
	if len(m.canvasForView().Objects) != len(before) {
		t.Fatal("lock hid objects")
	}
	for _, id := range []string{"move-left", "width-more", "delete", "group", "ungroup", "bring-front"} {
		if m.commands.Invoke(context.Background(), id) == nil {
			t.Fatal("locked edit allowed", id)
		}
	}
	if m.deleteSelected() == nil || m.duplicateSelected() == nil {
		t.Fatal("direct edit on locked selection")
	}
	g := testLayerLayout(m, 1)
	if m.beginLayerDrag("group-1", layerPoint(g, "group-1", true), g) != nil {
		t.Fatal("locked layer drag")
	}
	_ = m.selectLayer("group-1")
	if m.layerActionTarget() != "group-1" || len(m.selectedIDs()) != 0 {
		t.Fatal("Unlock target inaccessible")
	}
	if err := m.journal.Undo(); err != nil || m.layerLocked("block-1") || len(m.selectedIDs()) != 2 {
		t.Fatal("lock Undo", err)
	}
	if err := m.journal.Redo(); err != nil || !m.layerLocked("group-1") || len(m.selectedIDs()) != 0 {
		t.Fatal("lock Redo", err)
	}
	if !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("locked geometry/history")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	n, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer n.palette.Close()
	if err = drawingCodec(n).Install(f); err != nil || !maps.Equal(n.layerInfo, m.layerInfo) || !n.layerLocked("note-3") {
		t.Fatal("locked roundtrip", err)
	}
	if err = m.toggleLayerLock("group-1"); err != nil || m.layerLocked("block-1") {
		t.Fatal("unlock", err)
	}
}
func TestImportedMemberLockProtectsGroupUntilExplicitUnlock(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Housing", Locked: true}}
	m.managedSelection("note-3")
	if len(m.selectedIDs()) != 0 || !m.layerLocked("group-1") {
		t.Fatal("imported member lock bypass")
	}
	if err := m.toggleLayerLock("group-1"); err != nil || m.layerLocked("note-3") || m.layerInfo["block-1"].Name != "Housing" {
		t.Fatal("unlock dropped name", err)
	}
	m.prompt = "rename"
	if m.toggleLayerLock("group-1") == nil {
		t.Fatal("busy lock mutation")
	}
}
