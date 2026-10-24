//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"image"
	"slices"
	"testing"
)

func TestMacOutlineNativeKeysDisclosureAndHiddenSelection(t *testing.T) {
	p := &ui.OutlineView{ID: "outline", Label: "Project", Items: []ui.OutlineItem{{ID: "root", Label: "Root"}, {ID: "child", Parent: "root", Label: "Child"}}}
	p.Change = func(e ui.ListSelectionEvent) { p.SelectedIDs = e.Apply(p.SelectedIDs) }
	p.Disclosure = func(e ui.OutlineDisclosureEvent) { p.ExpandedIDs = e.Apply(p.ExpandedIDs) }
	calls := 0
	p.Activate = func(ui.OutlineEvent) { calls++ }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root-view", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: p.ID, Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 912), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	keyInput := func(code key.Codes) {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	keyInput(key.CodeRightArrow)
	keyInput(key.CodeRightArrow)
	keyInput(key.CodeReturnEnter)
	if !slices.Equal(p.SelectedIDs, []string{"child"}) || calls != 1 {
		t.Fatal("native outline key route", p.SelectedIDs, calls)
	}
	root := findRecord(w.state.root, p.ID).control.choice.buttons["root"]
	slot := root.ChildByName("outline-disclosure-slot").(*core.Frame)
	disclosure := slot.ChildByName("disclosure").(*core.Button)
	at := disclosure.Geom.TotalBBox.Min.Add(disclosure.Geom.TotalBBox.Size().Div(2)).Add(disclosure.Scene.SceneGeom.Pos)
	if at == (image.Point{}) {
		t.Fatal("no disclosure geometry")
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: at}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if len(p.ExpandedIDs) != 0 || !slices.Equal(p.SelectedIDs, []string{"child"}) || calls != 1 {
		t.Fatal("disclosure click leaked selection/activation", p.ExpandedIDs, p.SelectedIDs, calls)
	}
	if err := a.Do(func() { p.ExpandedIDs = []string{"root"}; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !findRecord(w.state.root, p.ID).control.choice.buttons["child"].IsChecked() {
		t.Fatal("hidden accepted selection not restored")
	}
	retained := findRecord(w.state.root, p.ID).control.choice.buttons["root"]
	if err := a.Do(func() { p.Items = p.Items[:1]; p.ExpandedIDs = nil; p.SelectedIDs = nil; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if retained.ChildByName("outline-disclosure-slot").(*core.Frame).ChildByName("disclosure") != nil {
		t.Fatal("removed child left disclosure")
	}
	if err := a.Do(func() {
		p.Items = append(p.Items, ui.OutlineItem{ID: "new-child", Parent: "root", Label: "New child"})
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if findRecord(w.state.root, p.ID).control.choice.buttons["root"] != retained || retained.ChildByName("outline-disclosure-slot").(*core.Frame).ChildByName("disclosure") == nil {
		t.Fatal("new parent failed to rebuild retained disclosure")
	}

}
