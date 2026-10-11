//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"testing"
)

func TestReviewGridRetiredTabCannotApplyAXPress(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "one-page", Options: []ui.Option{{ID: "one-page", Label: "One"}, {ID: "two-page", Label: "Two"}}, Children: []ui.Node{
			{ID: "one-page", Kind: ui.Column, Children: []ui.Node{{ID: "g", Kind: ui.DataGrid, Grid: g, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 240}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 240})}}}},
			{ID: "two-page", Kind: ui.Column, Children: []ui.Node{{ID: "other", Kind: ui.Label, Text: "Other page"}}},
		}}
	})
	gridFrame(t, m)
	ins := gridInspection(t, a, g)
	visible := ""
	for _, c := range ins.Cells {
		if !c.Bounds.Empty() {
			visible = c.ID
			break
		}
	}
	if visible == "" {
		t.Fatal("no current visible cell")
	}
	if err := a.Do(func() {
		root := m.implementation.find("one").state.root
		_, handlers := accessibilitySnapshot(root, nil)
		press := handlers[accessibilityManagedID(visible)]
		if press == nil {
			t.Fatal("visible cell has no handler")
		}
		root.widget.(*core.Tabs).SelectTabByName("two-page")
		bindGridScopes(root)
		before, _ := g.State()
		accepted := press(accessibilityAction{Kind: "press"})
		after, _ := g.State()
		if accepted || before != after {
			t.Errorf("AX press mutated grid in inactive actual native tab: accepted=%v before=%+v after=%+v", accepted, before.Active, after.Active)
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestManagedGridQueuedIntentRefusesOptimisticNativeTab(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "one-page", Options: []ui.Option{{ID: "one-page", Label: "One"}, {ID: "two-page", Label: "Two"}}, Children: []ui.Node{
			{ID: "one-page", Kind: ui.Column, Children: []ui.Node{{ID: "g", Kind: ui.DataGrid, Grid: g, GridChange: func(e ui.GridEvent) {
				if e.Reason == ui.GridSelection {
					changes++
				}
			}, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 240}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 240})}}}},
			{ID: "two-page", Kind: ui.Column, Children: []ui.Node{{ID: "other", Kind: ui.Label, Text: "Other"}}},
		}}
	})
	gridFrame(t, m)
	var before, after virtual.GridState
	if err := a.Do(func() {
		root := m.implementation.find("one").state.root
		b := root.children[0].children[0].control.grid
		before, _ = g.State()
		b.post(ui.GridSelection, func() error { return g.Move(1, 0) })
		root.widget.(*core.Tabs).SelectTabByName("two-page")
		// Deliberately do not refresh scopes: queued delivery must check the retained
		// owner identity against the actual tab, including a global other-window drain.
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := a.Do(func() { after, _ = g.State() }); err != nil {
		t.Fatal(err)
	}
	if before != after || changes != 0 {
		t.Fatal("queued intent crossed optimistic native tab switch", changes, before, after)
	}
}
