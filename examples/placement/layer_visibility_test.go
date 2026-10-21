package main

import (
	"github.com/dshills/vellum/canvas"
	"maps"
	"slices"
	"testing"
)

func TestLayerVisibilityProjectionPickingHistoryAndPersistence(t *testing.T) {
	m := groupedModel(t)
	m.scene.View.Center = canvas.Point{X: 12, Y: 34}
	m.scene.View.Zoom = 2
	before := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	if err := m.toggleLayerVisibility("block-1"); err != nil {
		t.Fatal(err)
	}
	if !m.layerHidden("group-1") || !m.layerHidden("note-3") || len(m.selectedIDs()) != 0 || m.scene.Selected != "" {
		t.Fatal("hidden group selectable")
	}
	display := m.canvasForView()
	if len(display.Objects) != 1 || display.Objects[0].ID != "connector-2" || display.View != m.scene.View {
		t.Fatal("projection/camera")
	}
	if !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("hide deleted objects or history")
	}
	// Managed callback cannot select a hidden object even through a stale ID.
	m.managedSelection("block-1")
	if len(m.selectedIDs()) != 0 {
		t.Fatal("AX hidden selection")
	}
	m.setSelection(marqueeIDs(m.scene.Objects, nil, canvas.Rect{X: -10000, Y: -10000, Width: 20000, Height: 20000}), "")
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) {
		t.Fatal("marquee includes hidden")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := drawingCodec(m).Encode(f)
	if err != nil {
		t.Fatal(err)
	}
	n, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer n.palette.Close()
	decoded, err := drawingCodec(n).Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err = drawingCodec(n).Install(decoded); err != nil || !maps.Equal(m.layerInfo, n.layerInfo) || len(n.canvasForView().Objects) != 1 {
		t.Fatal("hide roundtrip", err)
	}
	if err = m.journal.Undo(); err != nil || m.layerHidden("block-1") {
		t.Fatal("undo visibility", err)
	}
	if err = m.journal.Redo(); err != nil || !m.layerHidden("group-1") {
		t.Fatal("redo visibility", err)
	}
	if err = m.toggleLayerVisibility("group-1"); err != nil || m.layerHidden("note-3") || m.canvasForView() != m.scene {
		t.Fatal("show", err)
	}
}
func TestHiddenLayersCanBeFocusedAndReorderedButNotHit(t *testing.T) {
	m := editModel(t)
	p := canvas.Point{X: m.scene.Objects[2].X + 1, Y: m.scene.Objects[2].Y + 1}
	if err := m.toggleLayerVisibility("note-3"); err != nil {
		t.Fatal(err)
	}
	if m.objectAtWorld(p) == "note-3" {
		t.Fatal("hidden pointer target")
	}
	if err := m.selectLayer("note-3"); err != nil || m.layerActionTarget() != "note-3" || len(m.selectedIDs()) != 0 {
		t.Fatal("hidden properties target", err)
	}
	g := testLayerLayout(m, 1)
	s := m.beginLayerDrag("note-3", layerPoint(g, "note-3", true), g)
	if s == nil || len(s.ids) != 1 {
		t.Fatal("hidden layer reorder")
	}
	m.cancelLayerDrag(s)
	m.move = &objectMove{}
	if m.toggleLayerVisibility("note-3") == nil {
		t.Fatal("busy visibility")
	}
	m.move = nil
}

func TestVisibilityProjectionCannotRestoreStaleSizeAndEmptyFitKeepsCamera(t *testing.T) {
	m := editModel(t)
	_ = m.toggleLayerVisibility("note-3")
	display := m.canvasForView()
	display.View.Width = 1200
	_ = m.canvasForView()
	if m.scene.View.Width != 1200 {
		t.Fatal("active projection resize lost")
	}
	_ = m.toggleLayerVisibility("note-3")
	_ = m.canvasForView()
	m.scene.View.Width = 1600
	_ = m.toggleLayerVisibility("note-3")
	if m.canvasForView().View.Width != 1600 {
		t.Fatal("inactive cache overwrote resize")
	}
	_ = m.toggleLayerVisibility("block-1")
	_ = m.toggleLayerVisibility("connector-2")
	m.scene.View.Center = canvas.Point{X: 53, Y: 26}
	m.scene.View.Zoom = 3
	before := m.scene.View
	m.fitVisible()
	if m.scene.View != before {
		t.Fatal("empty Fit changed camera")
	}
}
