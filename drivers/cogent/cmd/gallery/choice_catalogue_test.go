package main

import (
	"encoding/json"
	"image"
	"reflect"
	"testing"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func TestChoiceCatalogueUsesOrdinaryControlledConstructors(t *testing.T) {
	g := newGallery()
	for _, view := range []ui.Node{g.view(), g.choiceCatalogueView()} {
		if err := ui.Validate(view); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"enable", "snapping", "pinned", "radio-preset", "segmented-preset", "slider", "progress", "progress-ring", "progress-gauge"} {
		a, b := node(g.view(), id), node(g.choiceCatalogueView(), id)
		if a.ID == "" || a.ID != b.ID || a.Kind != b.Kind || a.Checked != b.Checked || a.Number != b.Number || a.Numeric != b.Numeric || a.SelectedID != b.SelectedID || !reflect.DeepEqual(a.Options, b.Options) {
			t.Fatal("fixture diverged from ordinary public control", id, a, b)
		}
	}
	node(g.view(), "radio-preset").Change(ui.ChangeEvent{SelectedID: "advanced"})
	g.selectionReject = true
	node(g.choiceCatalogueView(), "segmented-preset").Change(ui.ChangeEvent{SelectedID: "basic"})
	if g.selection != "advanced" || g.selectionRequests != 2 || node(g.view(), "segmented-preset").SelectedID != "advanced" {
		t.Fatal("shared controlled choice/rejection", g.selection, g.selectionRequests)
	}
	node(g.view(), "slider").Change(ui.ChangeEvent{Number: 42})
	g.numberReject = true
	node(g.choiceCatalogueView(), "slider").Change(ui.ChangeEvent{Number: 43})
	if g.number != 42 || g.numberRequests != 2 {
		t.Fatal("controlled slider/rejection")
	}
	for _, id := range []string{"progress", "progress-ring", "progress-gauge"} {
		if node(g.choiceCatalogueView(), id).Number != float32(.42) {
			t.Fatal("progress publication", id)
		}
	}
	node(g.view(), "snapping").Change(ui.ChangeEvent{Checked: true})
	node(g.view(), "pinned").Change(ui.ChangeEvent{Checked: true})
	node(g.view(), "enable").Change(ui.ChangeEvent{Checked: false})
	if !g.snapping || !g.pinned || g.snappingRequests != 1 || g.pinnedRequests != 1 || g.enabledRequests != 1 || !node(g.choiceCatalogueView(), "fields").Disabled {
		t.Fatal("boolean ownership/counters")
	}
	if g.choiceCatalogueView().Focus != (ui.FocusRequest{ID: "enable", Sequence: 1}) {
		t.Fatal("initial focus must stay on independent enable control")
	}
}

func TestChoiceOptionLookupQualifiesDuplicateDomainIDs(t *testing.T) {
	s := cogent.DesktopSnapshot{Elements: []cogent.DesktopElement{
		{ID: "radio", DomainID: "radio-preset"}, {ID: "segments", DomainID: "segmented-preset"},
		{ID: "radio-basic", DomainID: "basic", Parent: "radio", Bounds: image.Rect(0, 0, 10, 10), Value: json.RawMessage("true")},
		{ID: "segments-basic", DomainID: "basic", Parent: "segments", Bounds: image.Rect(20, 0, 30, 10), Value: json.RawMessage("true")},
	}}
	el, err := choiceElement(s, "segmented-preset", "basic")
	if err != nil || el.ID != "segments-basic" {
		t.Fatal("option ID resolved outside its group", el, err)
	}
	if _, err = choiceElement(s, "radio-preset", "missing"); err == nil {
		t.Fatal("missing option admitted")
	}
}
