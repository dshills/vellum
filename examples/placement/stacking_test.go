package main

import (
	"context"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
)

func objectIDs(objects []canvas.Rect) []string {
	ids := make([]string, 0, len(objects))
	for _, o := range objects {
		ids = append(ids, o.ID)
	}
	return ids
}

func TestStackingStableOrderHistoryAndState(t *testing.T) {
	for _, front := range []bool{true, false} {
		t.Run(map[bool]string{true: "front", false: "back"}[front], func(t *testing.T) {
			m := groupedModel(t) // Noncontiguous block/note with connector between them.
			m.scene.Objects = append(m.scene.Objects,
				canvas.Rect{ID: "block-4", X: 33, Y: 42, Width: 80, Height: 50},
				canvas.Rect{ID: "note-5", X: 80, Y: 90, Width: 90, Height: 65})
			m.next = 5
			// An independent selected component joins the group; selection order is
			// intentionally reversed to verify paint order, rather than click order.
			m.setSelection([]string{"block-4", "note-3"}, "note-3")
			m.scene.View.Zoom = 2.25
			m.scene.View.Center = canvas.Point{X: 117, Y: -82}
			before, groups := slices.Clone(m.scene.Objects), maps.Clone(m.groups)
			view, next, nextGroup := m.scene.View, m.next, m.nextGroup
			_ = m.journal.MarkSaved()
			count := m.journal.Snapshot().UndoCount
			commandID := "send-back"
			want := []string{"block-1", "note-3", "block-4", "connector-2", "note-5"}
			if front {
				commandID = "bring-front"
				want = []string{"connector-2", "note-5", "block-1", "note-3", "block-4"}
			}
			if err := m.commands.Invoke(context.Background(), commandID); err != nil {
				t.Fatal(err)
			}
			after := slices.Clone(m.scene.Objects)
			if !slices.Equal(objectIDs(after), want) {
				t.Fatal("wrong stable order", objectIDs(after))
			}
			original := make(map[string]canvas.Rect)
			for _, o := range before {
				original[o.ID] = o
			}
			for _, o := range after {
				if o != original[o.ID] {
					t.Fatal("geometry or identity changed")
				}
			}
			if !maps.Equal(groups, m.groups) || m.scene.View != view || m.next != next || m.nextGroup != nextGroup || m.scene.Selected != "note-3" || !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3", "block-4"}) || m.journal.Snapshot().UndoCount != count+1 || !m.journal.Snapshot().Dirty {
				t.Fatal("unrelated state or history")
			}
			if err := m.stackSelected(front); err != nil || m.journal.Snapshot().UndoCount != count+1 {
				t.Fatal("edge no-op history", err)
			}
			if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !maps.Equal(m.groups, groups) || m.scene.Selected != "note-3" || m.journal.Snapshot().Dirty {
				t.Fatal("stack Undo", err)
			}
			if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) || !maps.Equal(m.groups, groups) {
				t.Fatal("stack Redo", err)
			}
		})
	}
}

func TestStackingChangesTopmostPickingAndPersists(t *testing.T) {
	m := groupedModel(t)
	for i := range m.scene.Objects {
		m.scene.Objects[i].X, m.scene.Objects[i].Y = 0, 0
	}
	p := m.scene.View.Screen(canvas.Point{X: 4, Y: 4})
	if m.scene.ObjectAt(p) != "note-3" {
		t.Fatal("fixture topmost")
	}
	m.setSelection([]string{"connector-2"}, "connector-2")
	if err := m.stackSelected(true); err != nil || m.scene.ObjectAt(p) != "connector-2" {
		t.Fatal("front not picked", err)
	}
	codec := drawingCodec(m)
	saved, err := codec.Capture()
	if err != nil {
		t.Fatal(err)
	}
	data, err := codec.Encode(saved)
	if err != nil {
		t.Fatal(err)
	}
	store := document.FileStore{}
	path := filepath.Join(t.TempDir(), "stack.vellum")
	if written, err := store.Write(context.Background(), path, data, drawingByteLimit); err != nil || !written {
		t.Fatal(err)
	}
	data, err = store.Read(context.Background(), path, drawingByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := codec.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	other, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(other.palette.Close)
	if err := drawingCodec(other).Install(loaded); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(other.scene.Objects, m.scene.Objects) || !maps.Equal(other.groups, m.groups) || other.scene.ObjectAt(other.scene.View.Screen(canvas.Point{X: 4, Y: 4})) != "connector-2" {
		t.Fatal("file lost stack/group state")
	}
	if err := m.stackSelected(false); err != nil || m.scene.ObjectAt(p) != "note-3" {
		t.Fatal("back not picked", err)
	}
	if err := m.journal.Undo(); err != nil || m.scene.ObjectAt(p) != "connector-2" {
		t.Fatal("picking Undo", err)
	}
	if err := m.journal.Redo(); err != nil || m.scene.ObjectAt(p) != "note-3" {
		t.Fatal("picking Redo", err)
	}
}

func TestStackingNoopAndBusyGuards(t *testing.T) {
	m := editModel(t)
	m.setSelection(nil, "")
	if m.commands.Enabled("bring-front") == nil || m.stackSelected(false) == nil {
		t.Fatal("empty selection enabled")
	}
	m.setSelection(objectIDs(m.scene.Objects), "note-3")
	_ = m.journal.MarkSaved()
	before, count := slices.Clone(m.scene.Objects), m.journal.Snapshot().UndoCount
	for _, front := range []bool{true, false} {
		if err := m.stackSelected(front); err != nil || m.journal.Snapshot().UndoCount != count || m.journal.Snapshot().Dirty {
			t.Fatal("all-selected no-op", err)
		}
	}
	for _, busy := range []string{"gesture", "prompt", "clipboard"} {
		switch busy {
		case "gesture":
			m.move = &objectMove{}
		case "prompt":
			m.prompt = "close"
		case "clipboard":
			m.clipboardJob = &clipboardOperation{}
		}
		if m.commands.Enabled("bring-front") == nil || m.commands.Enabled("send-back") == nil || m.stackSelected(true) == nil || !slices.Equal(m.scene.Objects, before) {
			t.Fatal("busy edit allowed", busy)
		}
		m.move = nil
		m.prompt = ""
		m.clipboardJob = nil
	}
}
