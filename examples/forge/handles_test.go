package main

import (
	"github.com/dshills/vellum/drivers/cogent"
	"image"
	"reflect"
	"testing"
)

func TestHandlePreviewCancelAndSingleCommit(t *testing.T) {
	m := newWorkbench()
	m.handleMode = "move"
	before := clone(m.state)
	m.drag = &handleDrag{sequence: 9, axis: 0, origin: m.selected().Bounds.Min, start: 0, before: before, after: before, mode: "move"}
	after, e := m.transform(before, 0, .2, "move")
	if e != nil {
		t.Fatal(e)
	}
	m.drag.after = after
	m.drag.moved = true
	if !m.handleInput(cogent.GPUEvent{Kind: cogent.GPUKey, Key: "Escape"}) || !reflect.DeepEqual(m.state, before) || m.drag != nil {
		t.Fatal("cancel")
	}
	m.drag = &handleDrag{sequence: 10, before: before, after: after, mode: "move", moved: true}
	up := cogent.GPUEvent{Kind: cogent.GPUPointerUp, Sequence: 10, Inside: true, Size: image.Pt(800, 600)}
	m.handleInput(up)
	m.handleInput(up)
	if m.journal.Snapshot().UndoCount != 1 {
		t.Fatal("commit count")
	}
	if e = m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, m.state) {
		t.Fatal("undo")
	}
}
func TestHandleResizeRejectAndAxisProjection(t *testing.T) {
	m := newWorkbench()
	before := m.captureStudy()
	if _, e := m.transform(m.state, 0, -100, "resize"); e == nil {
		t.Fatal("negative size")
	}
	if !reflect.DeepEqual(before, m.captureStudy()) {
		t.Fatal("mutated")
	}
	for i := 0; i < 3; i++ {
		v, ok := axisAt(m.scene.Camera, cogent.GPUEvent{Position: image.Pt(400, 300), Size: image.Pt(800, 600)}, m.scene.Camera.Target, i)
		if !ok || v > 1e-6 || v < -1e-6 {
			t.Fatalf("axis projection %d %g %v", i, v, ok)
		}
	}
}
