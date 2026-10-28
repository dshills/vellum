package main

import (
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"strings"
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
	if requireFidelityNode(t, m.View(), "canvas-mode").Text != "Orbit · Perspective" {
		t.Fatal("perspective caption")
	}
	m.action("projection")
	if !m.scene.Camera.Orthographic || requireFidelityNode(t, m.View(), "canvas-mode").Text != "Orbit · Orthographic" {
		t.Fatal("projection caption ignored state")
	}
}

func TestRibbonExplicitLightSurfaceAndInk(t *testing.T) {
	m := newWorkbench()
	n := m.ribbonView()
	if !n.Style.Foreground.Set || !n.Style.Background.Set {
		t.Fatal("ribbon must own light surface and ink")
	}
	for _, page := range n.Children {
		if page.Style.Background != n.Style.Background || page.Style.Foreground != n.Style.Foreground {
			t.Fatalf("native toolbar page %s inherited system appearance", page.ID)
		}
	}
}

func TestNavigationFeedbackTracksExistingModesWithoutHistory(t *testing.T) {
	m := newWorkbench()
	before := clone(m.state)
	for _, tool := range []string{"orbit", "pan", "select"} {
		m.action(tool)
		view := m.View()
		for _, candidate := range []string{"orbit", "pan", "select"} {
			n := fidelityNode(view, ui.RibbonActionID("ribbon", "solid", candidate))
			if n == nil {
				t.Fatal("missing mode", candidate)
			}
			active := strings.Contains(n.Accessibility.Name, "(active)")
			if active != (candidate == tool) {
				t.Fatal("incorrect active mode", tool, candidate)
			}
			if active && (!n.Style.Background.Set || !n.Style.Border.Set || n.Tooltip == "") {
				t.Fatal("missing mode feedback")
			}
		}
		if !strings.HasPrefix(fidelityNode(view, "canvas-mode").Text, m.toolLabel()+" ·") {
			t.Fatal("stale mode caption")
		}
		if m.journal.Snapshot().UndoCount != 0 || m.state.Selected != before.Selected || !slices.Equal(m.state.Objects, before.Objects) {
			t.Fatal("transient mode changed document")
		}
	}
	m.marking.Open = true
	view := m.View()
	active := fidelityNode(view, "action-wheel/cell/2/2")
	if active.Accessibility.Name != "Select (active)" {
		t.Fatal("panel disagrees with ribbon", active.Accessibility.Name)
	}
}
