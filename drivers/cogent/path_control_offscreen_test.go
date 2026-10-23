//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestPathControlManagedOverflowNavigation(t *testing.T) {
	p := &ui.PathControl{ID: "path", MaxVisible: 3, Items: []ui.PathItem{{ID: "root", Label: "Home"}, {ID: "docs", Label: "Documents"}, {ID: "projects", Label: "Projects"}, {ID: "drawings", Label: "Drawings"}, {ID: "file", Label: "Draft.vellum", Kind: ui.PathFile}}}
	calls := 0
	p.Navigate = func(e ui.PathEvent) {
		calls++
		for i, item := range p.Items {
			if item.ID == e.ItemID {
				p.Items = append([]ui.PathItem(nil), p.Items[:i+1]...)
				return
			}
		}
	}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.PathOverflowID(p.ID), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 924), view)
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.PathOverflowID(p.ID))
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	menu := findRecord(m.implementation.find("one").state.root, ui.PathOverflowID(p.ID))
	if !menuOpen(menu) {
		t.Fatal("overflow did not open")
	}
	_, item := imageManualElement(t, m, menuActionID(ui.PathOverflowID(p.ID), "docs"))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: item.ID, TargetIncarnation: item.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if calls != 1 || len(p.Items) != 2 || p.Items[1].ID != "docs" {
		t.Fatal(calls, p.Items)
	}
}

func TestPathControlDisabledReadOnlyChildren(t *testing.T) {
	p := &ui.PathControl{ID: "path", Items: []ui.PathItem{{ID: "home", Label: "Home", Disabled: true}}}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return n
	}
	m, _ := simpleManual(t, manualConfig(t, 926), view)
	unitSettle(t, m)
	for _, suffix := range []string{"/icon", "/label"} {
		_, e := imageManualElement(t, m, ui.PathSegmentID(p.ID, "home")+suffix)
		if e.Enabled {
			t.Fatal("disabled ancestry lost", suffix)
		}
	}
}

func TestPathControlInsideMacFormHasVisibleWidth(t *testing.T) {
	p := &ui.PathControl{ID: "path", MaxVisible: 3, Items: []ui.PathItem{{ID: "home", Label: "Home"}, {ID: "docs", Label: "Documents"}, {ID: "projects", Label: "Projects"}, {ID: "drawings", Label: "Drawings"}, {ID: "file", Label: "Draft.vellum", Kind: ui.PathFile}}, Navigate: func(ui.PathEvent) {}}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Scroll, Grow: true, Children: []ui.Node{{ID: "group", Kind: ui.FormGroup, Text: "Location", Children: []ui.Node{n}}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 928), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	path := findRecord(root, p.ID).widget.AsWidget()
	home := findRecord(root, ui.PathSegmentID(p.ID, "home")).widget.AsWidget()
	if path.Geom.ContentBBox.Dx() < 100 || home.Geom.TotalBBox.Empty() || !home.Geom.TotalBBox.In(path.Geom.ContentBBox) {
		t.Fatal("path toolbar collapsed inside form", path.Geom, home.Geom)
	}
}
