package main

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

func layerKeys(rows []layerRow) []string {
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.id)
	}
	return ids
}
func layerNode(t *testing.T, n ui.Node, id string) ui.Node {
	t.Helper()
	var found ui.Node
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if n.ID == id {
			found = n
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	if found.ID == "" {
		t.Fatal("missing layer node", id)
	}
	return found
}

func TestLayersProjectionSelectionAndHistory(t *testing.T) {
	m := groupedModel(t)
	before, groups := slices.Clone(m.scene.Objects), maps.Clone(m.groups)
	_ = m.journal.MarkSaved()
	h := m.journal.Snapshot()
	rows := m.layerRows()
	if !slices.Equal(layerKeys(rows), []string{"group-1", "note-3", "connector-2", "block-1"}) {
		t.Fatal("not front-first or group marker misplaced", rows)
	}
	if !rows[0].selected || !rows[1].selected || rows[2].selected || !rows[3].selected {
		t.Fatal("selection projection")
	}
	if !slices.Equal(m.scene.Objects, before) || !maps.Equal(m.groups, groups) || !reflect.DeepEqual(h, m.journal.Snapshot()) {
		t.Fatal("projection mutated model")
	}
	if err := m.selectLayer("connector-2"); err != nil || m.scene.Selected != "connector-2" || len(m.selectedIDs()) != 1 {
		t.Fatal("single layer selection", err)
	}
	if err := m.selectLayer("group-1"); err != nil || m.scene.Selected != "note-3" || !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) {
		t.Fatal("group selection", err)
	}
	if err := m.selectLayer("block-1"); err != nil || m.scene.Selected != "block-1" || len(m.selectedIDs()) != 2 {
		t.Fatal("member expands whole group", err)
	}
	m.layerAdd = true
	_ = m.selectLayer("connector-2")
	if len(m.selectedIDs()) != 3 || m.scene.Selected != "connector-2" {
		t.Fatal("add to selection")
	}
	_ = m.selectLayer("group-1")
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) {
		t.Fatal("toggle group out")
	}
	_ = m.selectLayer("group-1")
	_ = m.selectLayer("note-3")
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) || m.scene.Selected != "connector-2" {
		t.Fatal("toggle primary group fallback")
	}
	if !reflect.DeepEqual(h, m.journal.Snapshot()) {
		t.Fatal("selection changed dirty/history state")
	}
	m.layerAdd = false
	_ = m.selectLayer("connector-2")
	if err := m.commands.Invoke(context.Background(), "bring-forward"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(layerKeys(m.layerRows()), []string{"connector-2", "group-1", "note-3", "block-1"}) {
		t.Fatal("projection did not follow reorder")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(layerKeys(m.layerRows()), layerKeys(rows)) || m.journal.Snapshot().Dirty {
		t.Fatal("panel Undo", err)
	}
	if err := m.journal.Redo(); err != nil || m.layerRows()[0].id != "connector-2" {
		t.Fatal("panel Redo", err)
	}
	// Selection must not discard redo or restore a deleted cached row.
	_ = m.journal.Undo()
	_ = m.selectLayer("group-1")
	if m.journal.Snapshot().RedoCount != 1 {
		t.Fatal("selection branched document history")
	}
	_ = m.selectLayer("")
	if len(m.selectedIDs()) != 0 {
		t.Fatal("clear")
	}
}

func TestLayersRefuseBusyAndStaleSelection(t *testing.T) {
	m := groupedModel(t)
	for _, busy := range []string{"gesture", "prompt", "clipboard", "missing-member", "missing-group"} {
		switch busy {
		case "gesture":
			m.move = &objectMove{}
		case "prompt":
			m.prompt = "open"
		case "clipboard":
			m.clipboardJob = &clipboardOperation{}
		}
		id := "connector-2"
		if strings.HasPrefix(busy, "missing") {
			id = busy
		}
		before, selected, h := slices.Clone(m.scene.Objects), m.selectedIDs(), m.journal.Snapshot()
		if m.selectLayer(id) == nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), selected) || !reflect.DeepEqual(h, m.journal.Snapshot()) {
			t.Fatal("busy/stale layer accepted", busy)
		}
		if busy == "gesture" || busy == "prompt" || busy == "clipboard" {
			if m.selectLayer("") == nil {
				t.Fatal("busy clear accepted", busy)
			}
		}
		m.move = nil
		m.prompt = ""
		m.clipboardJob = nil
	}
	_ = m.selectLayer("group-1")
	if err := m.deleteSelected(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(layerKeys(m.layerRows()), []string{"connector-2"}) || m.selectLayer("group-1") == nil || m.selectLayer("block-1") == nil {
		t.Fatal("removed layers retained")
	}
	_ = m.journal.Undo()
	if m.selectLayer("group-1") != nil || len(m.selectedIDs()) != 2 {
		t.Fatal("restored group unavailable")
	}
}

func TestLayersPanelCallbacksAndLifecycle(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200})
	var panel, stale ui.Node
	f.owner(t, func() {
		m := f.m
		updates := 0
		invoke := func(id string) func() {
			return func() {
				if err := m.commands.Invoke(f.ref.Context(), id); err != nil {
					t.Fatal(err)
				}
				updates++
			}
		}
		panel = m.layersPanel(f.ref, func() { updates++ }, invoke)
		if err := ui.Validate(panel); err != nil {
			t.Fatal(err)
		}
		stale = layerNode(t, panel, "layer-select-block-1")
		stale.Activate()
		if m.scene.Selected != "block-1" {
			t.Fatal("row closure selected wrong object")
		}
		layerNode(t, panel, "layers-add").Change(ui.ChangeEvent{Checked: true})
		layerNode(t, panel, "layer-select-note-2").Activate()
		if !m.layerAdd || len(m.selectedIDs()) != 2 || updates != 3 {
			t.Fatal("add callback")
		}
		m.layerAdd = false
		stale.Activate()
		layerNode(t, panel, "layers-forward").Activate()
		if m.scene.Objects[1].ID != "block-1" || m.journal.Snapshot().UndoCount != 3 {
			t.Fatal("panel command history")
		}
		_ = m.journal.Undo()
		layerNode(t, panel, "layers-backward").Activate()
		if m.scene.Objects[0].ID != "block-1" || m.journal.Snapshot().UndoCount != 2 {
			t.Fatal("backward edge no-op")
		}
		layerNode(t, panel, "layers-clear").Activate()
		if len(m.selectedIDs()) != 0 {
			t.Fatal("clear callback")
		}
		m.move = &objectMove{}
		blocked := m.layersPanel(f.ref, func() { updates++ }, invoke)
		if !layerNode(t, blocked, "layer-select-block-1").Disabled || !layerNode(t, blocked, "layers-add").Disabled || !layerNode(t, blocked, "layers-forward").Disabled {
			t.Fatal("busy panel enabled")
		}
		stale.Activate()
		if len(m.selectedIDs()) != 0 {
			t.Fatal("stale enabled callback bypassed gate")
		}
		m.move = nil
	})
	f.a.Close()
	prior := f.m.scene.Selected
	stale.Activate()
	if f.m.scene.Selected != prior {
		t.Fatal("closed-window callback changed selection")
	}
}

func TestLayersMaximumBoundAndNewDocumentProjection(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.palette.Close)
	if len(m.layerRows()) != 0 {
		t.Fatal("empty scene")
	}
	m.groups = make(map[string]string)
	for i := 0; i < 512; i++ {
		id := fmt.Sprintf("block-%d", i+1)
		m.scene.Objects = append(m.scene.Objects, canvas.Rect{ID: id, Width: 80, Height: 50})
		m.groups[id] = fmt.Sprintf("group-%d", i/2+1)
	}
	m.next = 512
	m.nextGroup = 256
	rows := m.layerRows()
	if len(rows) != 768 || rows[0].id != "group-256" || rows[1].id != "block-512" {
		t.Fatal("maximum projection", len(rows))
	}
	empty, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(empty.palette.Close)
	snapshot, err := drawingCodec(empty).Capture()
	if err != nil {
		t.Fatal(err)
	}
	if err := drawingCodec(m).Install(snapshot); err != nil || len(m.layerRows()) != 0 {
		t.Fatal("Open retained previous document rows", err)
	}
}
