package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"maps"
	"math"
	"slices"
	"testing"
)

func TestPasteCenterPreservesGeometryMetadataFreshIDsAndHistory(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"group-1": {Name: "Assembly"}, "block-1": {Name: "Housing"}}
	objects, err := m.copiedSelection()
	if err != nil {
		t.Fatal(err)
	}
	text, err := encodeSelection(objects)
	if err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(m.scene.Objects)
	metadata := maps.Clone(m.layerInfo)
	groups := maps.Clone(m.groups)
	state := m.journal.Snapshot()
	next := m.next
	m.scene.View.Zoom = 3
	m.scene.View.Center = canvas.Point{X: 413.25, Y: -227.5}
	m.snap = true
	m.gridSpacing = 100
	m.pasteText, m.pasteCount = "unchanged", 17
	if err := m.pasteSelectionAt(text, m.scene.View.Center); err != nil {
		t.Fatal(err)
	}
	ids := m.selectedIDs()
	after := slices.Clone(m.scene.Objects)
	bounds, ok := selectionBounds(after, ids)
	if !ok || math.Abs(bounds.X+bounds.Width/2-413.25) > 1e-9 || math.Abs(bounds.Y+bounds.Height/2+227.5) > 1e-9 {
		t.Fatal("not centered", bounds)
	}
	if len(ids) != 2 || ids[0] == "block-1" || m.groups[ids[0]] == "group-1" || m.groups[ids[0]] != m.groups[ids[1]] || m.layerInfo[ids[0]].Name != "Housing" || m.layerInfo[m.groups[ids[0]]].Name != "Assembly" {
		t.Fatal("identities/metadata", ids, m.groups)
	}
	if m.pasteText != "unchanged" || m.pasteCount != 17 || m.journal.Snapshot().UndoCount != state.UndoCount+1 || m.next != next+2 {
		t.Fatal("paste counter/history")
	}
	if after[len(before)+1].X-after[len(before)].X != before[2].X-before[0].X {
		t.Fatal("relative position changed")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) || !maps.Equal(metadata, m.layerInfo) || !maps.Equal(groups, m.groups) {
		t.Fatal("Undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) {
		t.Fatal("Redo changed identities", err)
	}
}
func TestPasteCenterRejectsInvalidBoundsPayloadAndCapacity(t *testing.T) {
	m := groupedModel(t)
	text := copiedText(t, m)
	before := slices.Clone(m.scene.Objects)
	state := m.journal.Snapshot()
	next := m.next
	for _, p := range []canvas.Point{{X: math.NaN()}, {X: 20000}} {
		if m.pasteSelectionAt(text, p) == nil {
			t.Fatal("invalid center accepted")
		}
	}
	if m.pasteSelectionAt("bad", canvas.Point{}) == nil {
		t.Fatal("bad clipboard")
	}
	m.locked = true
	if m.pasteSelectionAt(text, canvas.Point{}) == nil {
		t.Fatal("placement rejection")
	}
	m.locked = false
	m.prompt = "geometry"
	if m.pasteSelectionAt(text, canvas.Point{}) == nil {
		t.Fatal("prompt admitted paste")
	}
	m.prompt = ""
	if !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() || m.next != next {
		t.Fatal("failed paste mutated")
	}
}
func TestPasteCenterUsesClipboardWorkerAndCapturedCenter(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 150, Y: 70})
	service := &appClipboard{}
	f.a.SetServices(service)
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		service.text = copiedText(t, f.m)
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		f.m.scene.View.Center = canvas.Point{X: -321.25, Y: 123.5}
		if err := f.m.commands.Invoke(f.ref.Context(), "paste-center"); err != nil {
			t.Fatal(err)
		}
		if f.m.commands.Enabled("new") == nil {
			t.Fatal("worker did not gate New")
		}
	})
	waitClipboard(t, f)
	f.owner(t, func() {
		b, ok := selectionBounds(f.m.scene.Objects, f.m.selectedIDs())
		if !ok || math.Abs(b.X+b.Width/2+321.25) > 1e-9 || math.Abs(b.Y+b.Height/2-123.5) > 1e-9 {
			t.Fatal("worker center", b)
		}
	})
}

func TestPasteCenterWindowRetirementCancelsReadAndWithdrawsCompletion(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	service := &appClipboard{started: make(chan struct{}), done: make(chan struct{})}
	f.a.SetServices(service)
	if _, err := f.a.OpenWindow(app.Window{ID: "keeper", Title: "Keeper", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "keeper", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		service.text = copiedText(t, f.m)
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		if err := f.m.commands.Invoke(f.ref.Context(), "paste-center"); err != nil {
			t.Fatal(err)
		}
	})
	cutAwait(t, service.started)
	f.owner(t, func() {
		if err := f.a.CloseWindow(f.ref); err != nil {
			t.Fatal(err)
		}
	})
	cutAwait(t, service.done)
	_ = f.a.Drain()
	if len(f.m.scene.Objects) != 1 || f.a.Pending().ReadyCompletions != 0 {
		t.Fatal("retired paste mutated or queued")
	}
}
