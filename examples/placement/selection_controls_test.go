package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"testing"
)

func presentationNode(t *testing.T, n ui.Node, id string) ui.Node {
	t.Helper()
	var result ui.Node
	var visit func(ui.Node)
	visit = func(v ui.Node) {
		if v.ID == id {
			result = v
		}
		for _, c := range v.Children {
			visit(c)
		}
	}
	visit(n)
	if result.ID == "" {
		t.Fatal("missing", id)
	}
	return result
}
func TestActualNewControlsKeepLayoutButRefuseActionsDuringLayerCapture(t *testing.T) {
	for _, selected := range []bool{false, true} {
		f := fileModel(t, false)
		f.place(t, "block", canvas.Point{})
		f.place(t, "note", canvas.Point{X: 200, Y: 100})
		f.owner(t, func() {
			m := f.m
			if !selected {
				m.setSelection(nil, "")
			}
			before := view(f.a, f.ref, m)
			g := testLayerLayout(m, 1)
			drag := m.beginLayerDrag("block-1", layerPoint(g, "block-1", true), g)
			if drag == nil {
				t.Fatal("capture")
			}
			after := view(f.a, f.ref, m)
			for _, id := range []string{"geometry-open", "aspect-lock", "selection-subtract", "recent-drawings"} {
				if presentationNode(t, before, id).Disabled != presentationNode(t, after, id).Disabled {
					t.Fatal("presentation changed", id)
				}
			}
			presentationNode(t, after, "geometry-open").Activate()
			presentationNode(t, after, "aspect-lock").Change(ui.ChangeEvent{Checked: true})
			presentationNode(t, after, "selection-subtract").Change(ui.ChangeEvent{Checked: true})
			if m.geometry != nil || m.prompt != "" || m.aspectLocked || m.selectionSubtract {
				t.Fatal("frozen presentation admitted edit")
			}
			m.cancelLayerDrag(drag)
		})
	}
}
