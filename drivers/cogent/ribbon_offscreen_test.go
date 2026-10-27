//go:build offscreen

package cogent

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestRibbonWholeGroupOverflowNestedVisibilityAndRetention(t *testing.T) {
	for _, dpi := range []float32{160, 240, 320} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			a, _ := app.New(app.Config{})
			defer a.Close()
			source := fullColorButtonSource(t, false)
			width := float32(230)
			calls := 0
			r := &ui.Ribbon{ID: "r", Label: "Design tools", SelectedTab: "solid", Action: func(ui.RibbonEvent) { calls++ }, Tabs: []ui.RibbonTab{{ID: "solid", Label: "Solid"}}}
			for i := 0; i < 3; i++ {
				r.Tabs[0].Groups = append(r.Tabs[0].Groups, ui.RibbonGroup{ID: fmt.Sprint(i), Label: fmt.Sprint("Group ", i), Actions: []ui.RibbonAction{{ID: fmt.Sprint("a", i), Label: "Create", IconImage: source}, {ID: fmt.Sprint("b", i), Label: "Modify", IconImage: source, Menu: []ui.Option{{ID: fmt.Sprint("variant", i), Label: "Variant", IconImage: source}}}}})
			}
			view := func() ui.Node {
				r.Style.MaxSize = ui.Some(ui.Size{Width: width})
				n, e := r.View()
				if e != nil {
					panic(e)
				}
				return n
			}
			_ = a.Open(app.Window{ID: "one", Title: "Ribbon", Width: 2200, Height: 800, View: view})
			cfg := manualConfig(t, 7)
			cfg.LogicalDPI = dpi
			m, e := NewManual(testContext, a, cfg)
			if e != nil {
				t.Fatal(e)
			}
			defer m.Close(testContext)
			unitSettle(t, m)
			w := m.implementation.find("one")
			page := findRecord(w.state.root, "r/ribbon/c29saWQ")
			more := page.children[len(page.children)-1]
			if toolbarOverflowHidden(more) || !toolbarOverflowHidden(page.children[2]) {
				t.Fatal("group did not collapse")
			}
			child := findRecord(w.state.root, ui.RibbonActionID("r", "solid", "a2"))
			native := child.widget
			if !toolbarOverflowHidden(child) {
				t.Fatal("nested hidden control exposed")
			}
			_, handlers := accessibilitySnapshot(w.state.root, nil)
			if handlers[accessibilityManagedID(child.node.ID)] != nil {
				t.Fatal("hidden group action remained actionable")
			}
			options := toolbarOverflowOptions(page, more.node.Options)
			if len(options) == 0 || options[0].Branch == nil {
				t.Fatal("missing grouped submenu")
			}
			leaf := options[0].Branch.Items()[0]
			if leaf.IconImage != source {
				t.Fatal("artwork lost")
			}
			more.node.Change(ui.ChangeEvent{SelectedID: leaf.ID})
			if calls != 1 {
				t.Fatal("overflow did not execute action", calls)
			}
			width = 850
			a.Invalidate("one")
			unitSettle(t, m)
			if !toolbarOverflowHidden(more) || toolbarOverflowHidden(child) || child.widget != native {
				t.Fatal("widen did not retain/reveal groups")
			}
		})
	}
}
