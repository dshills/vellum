package main

import (
	"github.com/dshills/vellum/ui"
	"reflect"
	"testing"
)

func TestStructureCatalogueSharedConstructorsAndScript(t *testing.T) {
	g := newGallery()
	view := g.structureCatalogueView()
	if err := ui.Validate(view); err != nil {
		t.Fatal(err)
	}
	ordinary := g.view()
	for _, id := range []string{"panes", "details", "document-tabs"} {
		a, b := node(ordinary, id), node(view, id)
		if a.ID == "" || a.Kind != b.Kind || a.Number != b.Number || a.Checked != b.Checked || a.Disabled != b.Disabled || !reflect.DeepEqual(a.Options, b.Options) {
			t.Fatal("shared public control differs", id)
		}
	}
	before := g.structureModel()
	steps := g.structureScript()
	if len(steps) != 42 || g.structureModel() != before {
		t.Fatal("script construction mutated state or size", len(steps))
	}
	labels := map[string]bool{}
	packets := 0
	for _, s := range steps {
		if s.Label == "" || labels[s.Label] {
			t.Fatal("duplicate/empty label")
		}
		labels[s.Label] = true
		switch s.Kind {
		case "observe":
		case "click":
			packets += 2
		case "key", "scroll", "refuse", "stale":
			packets++
		default:
			t.Fatal("unknown script operation", s.Kind)
		}
	}
	if packets != 50 {
		t.Fatal("packet count", packets)
	}
	if !structureModelsEqual(structureModel{Fraction: 0.35}, structureModel{Fraction: 0.35000002}) || structureModelsEqual(structureModel{SectionActions: 1}, structureModel{}) {
		t.Fatal("fraction tolerance hid discrete model change")
	}
}
