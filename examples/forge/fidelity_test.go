package main

import (
	"github.com/dshills/vellum/ui"
	"math"
	"testing"
)

func TestInspectorFidelityKeepsQuantityValidationAndCommands(t *testing.T) {
	m := newWorkbench()
	original := m.selected()
	n := m.inspectorView()
	input := requireFidelityNode(t, n, "width/quantity/input")
	if input == nil || input.Style.MinSize.Value.Width != 64 {
		t.Fatal("compact field missing")
	}
	input.Change(ui.ChangeEvent{Text: "width / 2"})
	n = m.inspectorView()
	requireFidelityNode(t, n, "width/quantity/apply").Activate()
	resized := m.selected()
	if math.Abs((resized.Bounds.Max.X-resized.Bounds.Min.X)*40-(original.Bounds.Max.X-original.Bounds.Min.X)*20) > 1e-8 || m.journal.Snapshot().UndoCount != 1 {
		t.Fatal("quantity command changed")
	}
	if e := m.journal.Undo(); e != nil || m.selected() != original {
		t.Fatal("quantity undo", e)
	}
	m.draft = "not-a-length"
	n = m.inspectorView()
	requireFidelityNode(t, n, "width/quantity/apply").Activate()
	if m.validation == "" || m.journal.Snapshot().UndoCount != 0 {
		t.Fatal("invalid quantity admitted")
	}
	m.state.Selected = ""
	n = m.inspectorView()
	for _, id := range []string{"nudge", "remove"} {
		if !requireFidelityNode(t, n, id).Disabled {
			t.Fatal("empty-selection action enabled", id)
		}
	}
	if e := ui.Validate(n); e != nil {
		t.Fatal(e)
	}
}

func requireFidelityNode(t *testing.T, n ui.Node, id string) *ui.Node {
	t.Helper()
	found := fidelityNode(n, id)
	if found == nil {
		t.Fatal("missing fidelity control", id)
	}
	return found
}

func TestProjectionCaptionMatchesExistingCameraState(t *testing.T) {
	m := newWorkbench()
	if requireFidelityNode(t, m.View(), "canvas-mode").Text != "Solid · Perspective" {
		t.Fatal("perspective caption")
	}
	m.action("projection")
	if !m.scene.Camera.Orthographic || requireFidelityNode(t, m.View(), "canvas-mode").Text != "Solid · Orthographic" {
		t.Fatal("projection caption ignored state")
	}
}
