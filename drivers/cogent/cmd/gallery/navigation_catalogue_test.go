package main

import (
	"github.com/dshills/vellum/ui"
	"reflect"
	"testing"
)

func TestNavigationCatalogueReusesOrdinaryPublicNodes(t *testing.T) {
	g := newGallery()
	ordinary, compact := g.view(), g.navigationCatalogueView()
	for _, id := range []string{"layers", "menu-actions", "view-menu"} {
		a, b := node(ordinary, id), node(compact, id)
		if a.ID == "" || a.Kind != b.Kind || a.SelectedID != b.SelectedID || !reflect.DeepEqual(a.Options, b.Options) || a.Disabled != b.Disabled {
			t.Fatal("different public constructor", id)
		}
	}
	if e := ui.Validate(compact); e != nil {
		t.Fatal(e)
	}
	for _, reverse := range []bool{false, true} {
		g.listReversed = reverse
		opts := g.layerListControl().Options
		if len(opts) != 24 || opts[0].ID != map[bool]string{false: "layer-00", true: "layer-23"}[reverse] {
			t.Fatal("ordinary list order changed")
		}
	}
}
func TestNavigationScriptConstructionIsInert(t *testing.T) {
	g := newGallery()
	before := g.navigationModel()
	steps := g.navigationScript()
	if g.navigationModel() != before || len(steps) != 41 {
		t.Fatal("authored script mutated model or record size", len(steps))
	}
	inputs := 0
	labels := map[string]bool{}
	for _, s := range steps {
		if labels[s.Label] || s.Label == "" {
			t.Fatal("duplicate or empty step")
		}
		labels[s.Label] = true
		switch s.Kind {
		case "observe":
		case "click", "popup-click", "outside":
			inputs += 2
		default:
			inputs++
		}
	}
	if inputs != 47 {
		t.Fatal("authored packet count", inputs)
	}
}
