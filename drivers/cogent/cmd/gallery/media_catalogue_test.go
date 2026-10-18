package main

import (
	"github.com/dshills/vellum/ui"
	"reflect"
	"testing"
)

func TestMediaCatalogueSharedViewsAndFrozenModel(t *testing.T) {
	g := newGallery()
	ordinary, media := g.view(), g.mediaCatalogueView()
	for _, view := range []ui.Node{ordinary, media} {
		if e := ui.Validate(view); e != nil {
			t.Fatal(e)
		}
	}
	for _, id := range []string{"image-fit", "image-preview", "image-replace", "ink", "media-tools", "media-form", "media-status"} {
		a, b := node(ordinary, id), node(media, id)
		if a.ID == "" || a.Kind != b.Kind || a.Text != b.Text || a.Image != b.Image || a.Disabled != b.Disabled || !reflect.DeepEqual(a.Options, b.Options) {
			t.Fatal("ordinary constructor differs", id)
		}
	}
	before := g.mediaModel()
	steps := g.mediaScript()
	if before != g.mediaModel() || len(steps) != 42 {
		t.Fatal("script mutates state or changed size", len(steps))
	}
	packets := 0
	seen := map[string]bool{}
	for _, s := range steps {
		if s.Label == "" || seen[s.Label] {
			t.Fatal("ambiguous label", s.Label)
		}
		seen[s.Label] = true
		switch s.Kind {
		case "observe":
		case "click":
			packets += 2
		case "key", "refuse", "stale-path", "stale-color":
			packets++
		case "text":
			packets += 4 + len([]rune(s.Text))
		default:
			t.Fatal("unknown scripted action", s.Kind)
		}
	}
	if packets != 71 {
		t.Fatal("native input packet budget", packets)
	}
	if !seen["hidden-color-refusal"] || !seen["image-replace"] || !seen["breadcrumb-stale-frame"] {
		t.Fatal("lifecycle cases missing")
	}
}
