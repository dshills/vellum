package main

import "testing"

func TestControlPresentationStaysStableDuringLayerCapture(t *testing.T) {
	for _, selected := range []bool{false, true} {
		m := editModel(t)
		if !selected {
			m.setSelection(nil, "")
		}
		beforeGeometry := m.editAvailable() != nil
		beforeInteraction := m.clipboardAvailable() != nil
		g := testLayerLayout(m, 1)
		s := m.beginLayerDrag("block-1", layerPoint(g, "block-1", true), g)
		if s == nil {
			t.Fatal("capture")
		}
		if m.controlDisabled("geometry", true) != beforeGeometry {
			t.Fatal("geometry presentation changed")
		}
		if m.controlDisabled("interaction", true) != beforeInteraction {
			t.Fatal("interaction presentation changed")
		}
		if m.editAvailable() == nil || m.clipboardAvailable() == nil {
			t.Fatal("presentation admitted live edit")
		}
		m.cancelLayerDrag(s)
		if !m.controlDisabled("geometry", true) {
			t.Fatal("capture snapshot retained")
		}
	}
}
