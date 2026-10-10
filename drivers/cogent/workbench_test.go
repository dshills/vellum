package cogent

import (
	"cogentcore.org/core/core"
	"example.com/vellum/app"
	"example.com/vellum/ui"
	"example.com/vellum/workbench"
	"math"
	"testing"
)

func TestDockReconcileRestoresTabsAndRatio(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	layout := workbench.Layout{Version: 1, Root: &workbench.Group{ID: "split", Axis: workbench.Horizontal, Ratio: .6, First: &workbench.Group{ID: "left", Tabs: []string{"one", "two"}, Active: "one"}, Second: &workbench.Group{ID: "right", Tabs: []string{"three"}, Active: "three"}}}
	model, err := workbench.New(layout, []string{"one", "two", "three"})
	if err != nil {
		t.Fatal(err)
	}
	w := &Workbench{Owner: a, Workspace: model}
	for _, id := range []string{"one", "two", "three"} {
		id := id
		w.Panels = append(w.Panels, Panel{ID: id, Title: id, View: func() ui.Node { return ui.Node{ID: id, Kind: ui.Label, Text: id} }})
	}
	host := core.NewBody("headless")
	native := w.Mount(host)
	if err := w.Sync(native); err != nil {
		t.Fatal(err)
	}
	if err := model.Activate("left", "two"); err != nil {
		t.Fatal(err)
	}
	if err := model.Resize("split", .7); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(native); err != nil {
		t.Fatal(err)
	}
	_, tab := w.tabs["left"].CurrentTab()
	if ratios := w.splits["split"].Splits(); len(ratios) != 2 || math.Abs(float64(ratios[0])-.7) > .001 {
		t.Fatal("split ratio", ratios)
	}
	if tab != 1 {
		t.Fatal("restored tab", tab)
	}
	if err := model.Move("three", "left", 0); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(native); err != nil {
		t.Fatal(err)
	}
	_, selected := w.tabs["left"].CurrentTab()
	if selected != 0 || w.tabs["left"].NumTabs() != 3 {
		t.Fatal("moved panel placement", selected)
	}
	if len(w.records) != 3 {
		t.Fatal("panel lost")
	}
}
