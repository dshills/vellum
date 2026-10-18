package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestTextCatalogueSharesValidOrdinaryControls(t *testing.T) {
	g := newGallery()
	ordinary, compact := g.view(), g.textCatalogueView()
	for _, view := range []ui.Node{ordinary, compact} {
		if err := ui.Validate(view); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"text-name", "text-notes", "text-preferences", "preferences", "dialog-name", "apply", "text-status"} {
		a, b := node(ordinary, id), node(compact, id)
		if a.ID == "" || a.Kind != b.Kind || a.Text != b.Text || a.Editor != b.Editor || a.Disabled != b.Disabled {
			t.Fatal("shared public constructor differs", id)
		}
	}
	before := g.textModel()
	steps := g.textScript()
	if before != g.textModel() || len(steps) != 33 {
		t.Fatal("script mutates model or changed record size")
	}
	packets := 0
	seen := map[string]bool{}
	for _, s := range steps {
		if s.Label == "" || seen[s.Label] {
			t.Fatal("ambiguous step", s.Label)
		}
		seen[s.Label] = true
		switch s.Kind {
		case "observe":
		case "click":
			packets += 2
		case "key", "refuse", "background", "retired", "stale":
			packets++
		case "replace":
			packets += 4 + len([]rune(s.Text))
		default:
			t.Fatal("unknown scripted action", s.Kind)
		}
	}
	if packets != 68 {
		t.Fatal("native input record changed", packets)
	}
	if node(ordinary, "notes").Kind != ui.Column || node(ordinary, "text-notes").Kind != ui.MultilineText {
		t.Fatal("document page and editor identity collide")
	}
}
