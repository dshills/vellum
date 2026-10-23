//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
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

func TestPersonalToolbarsUseDarkFormBackground(t *testing.T) {
	path := &ui.PathControl{ID: "path", Items: []ui.PathItem{{ID: "home", Label: "Home"}}}
	tokens := &ui.TokenField{ID: "tags", Tokens: []ui.Token{{ID: "design", Text: "Design"}}, Change: func(ui.TokenEvent) {}}
	view := func() ui.Node {
		p, err := path.View()
		if err != nil {
			panic(err)
		}
		tags, err := tokens.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "group", Kind: ui.FormGroup, Text: "Dark controls", Children: []ui.Node{p, tags}}}}
	}
	config := manualConfig(t, 929)
	config.Theme = ui.MacTheme(true)
	config.Theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
	config.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
	m, _ := simpleManual(t, config, view)
	unitSettle(t, m)
	frame, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	root := m.implementation.find("one").state.root
	for _, id := range []string{"path", "tags/tokens/chips"} {
		box := findRecord(root, id).widget.AsWidget().Geom.ContentBBox
		point := image.Pt(box.Max.X-3, box.Min.Y+2)
		got := frame.RGBAAt(point.X, point.Y)
		expected := color.RGBA{R: 65, G: 65, B: 68, A: 255}
		if got != expected {
			t.Fatal("toolbar does not inherit dark form surface", id, point, got, box)
		}
	}
}
