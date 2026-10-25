package sdkcontract_test

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestPublicQuickLookContract(t *testing.T) {
	o := app.QuickLookOptions{Items: []app.QuickLookItem{{ID: "a", Path: "/tmp/a.txt", Title: "A"}}, SelectedID: "a"}
	if o.Validate() != nil {
		t.Fatal("options")
	}
	d := &ui.DocumentList{ID: "d", Label: "Files", Items: []ui.DocumentItem{{ID: "a", Name: "A"}}, Preview: func(ui.DocumentEvent) {}}
	if _, e := d.View(); e != nil {
		t.Fatal(e)
	}
}

func TestPublicFindBarContract(t *testing.T) {
	m, _ := editor.New("Plan plan", editor.Config{})
	f := &ui.FindBar{ID: "find", Model: m, Open: true, Query: "plan", Request: func(r ui.FindRequest) {
		if e := r.Apply(m); e != nil {
			t.Error(e)
		}
	}}
	if _, e := f.View(); e != nil {
		t.Fatal(e)
	}
	if e := (ui.FindRequest{Action: ui.FindReplaceAll, Query: "plan", Replacement: "drawing"}).Apply(m); e != nil {
		t.Fatal(e)
	}
	if e := m.Journal().Undo(); e != nil || m.Snapshot().Text != "Plan plan" {
		t.Fatal(e)
	}
}

func TestPublicSearchScopeContract(t *testing.T) {
	bar := &ui.SearchScopeBar{ID: "scopes", Label: "Search", Value: ui.SearchScopeValue{ScopeID: "all"}, Items: []ui.SearchScopeItem{{ID: "all", Label: "All", Count: 0}}}
	bar.Change = func(e ui.SearchScopeEvent) { bar.Value = e.Apply(bar.Value) }
	if _, e := bar.View(); e != nil {
		t.Fatal(e)
	}
}

func TestPublicGradientWellContract(t *testing.T) {
	g := &ui.GradientWell{ID: "gradient", Label: "Gradient", Value: ui.GradientValue{Stops: []ui.GradientStop{{ID: "a", Position: 0, Color: color.NRGBA{R: 255, A: 255}}, {ID: "b", Position: 1, Color: color.NRGBA{B: 255, A: 255}}}}}
	g.Change = func(e ui.GradientEdit) {
		v, err := e.Apply(g.Value)
		if err != nil {
			t.Error(err)
		} else {
			g.Value = v
		}
	}
	if _, e := g.View(); e != nil {
		t.Fatal(e)
	}
}
