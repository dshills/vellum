package main

import (
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestGallerySystemAppearancePreservesValuesAndExplicitOverride(t *testing.T) {
	inputs := newInputHelpersGallery()
	m := newMacWidgetsGallery(inputs, false)
	m.components.followSystem = true
	m.components.name = "Retained"
	selected := m.collection.collection.SelectedIDs[0]
	v := cogent.SystemAppearance{Known: true, Dark: true, HighContrast: true, ReduceMotion: true, Accent: color.NRGBA{R: 180, G: 80, B: 210, A: 255}}
	m.acceptSystemAppearance(v)
	if !m.dark || !m.components.dark || !m.components.theme().ReducedMotion || m.components.name != "Retained" || m.collection.collection.SelectedIDs[0] != selected {
		t.Fatal("system preference reset values or theme")
	}
	n := m.components.view()
	if e := ui.Validate(n); e != nil {
		t.Fatal(e)
	}
	// The existing Light action must remain an explicit override.
	var apply func(ui.Node)
	apply = func(n ui.Node) {
		if n.ID == "components/appearance/Light" {
			n.Activate()
		}
		for _, c := range n.Children {
			apply(c)
		}
	}
	apply(n)
	m.acceptSystemAppearance(v)
	if m.components.followSystem || m.components.dark || !m.components.theme().ReducedMotion || m.components.theme().Palette.Value.Line == ui.MacTheme(false).Palette.Value.Line {
		t.Fatal("system observation overwrote explicit light choice")
	}
}
