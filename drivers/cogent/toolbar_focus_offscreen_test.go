//go:build offscreen

package cogent

import (
	"cogentcore.org/core/styles/states"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestToolbarCustomizationFocusCuesFollowActualScope(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			cfg := manualConfig(t, 7)
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			c := ui.ToolbarCustomizer{ID: "customize", ReturnFocusID: "opener"}
			c.Dismiss = func() { c.Open = false }
			c.Change = func(v ui.ToolbarLayout) { c.Value = v }
			c.Toolbar = ui.Node{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{{ID: "a", Kind: ui.Button, Text: "One", Activate: func() {}}, {ID: "b", Kind: ui.Button, Text: "Two", Activate: func() {}}}}
			view := func() ui.Node {
				n, e := c.View()
				if e != nil {
					panic(e)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "opener", Sequence: 1}, Children: []ui.Node{{ID: "opener", Kind: ui.Button, Text: "Customize", Style: ui.Style{Border: ui.Some(ui.Border{})}, Activate: func() { c.Open = true }}, n}}
			}
			a, e := app.New(app.Config{})
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(a.Close)
			if e = a.Open(app.Window{ID: "one", Title: "Focus", Width: 660, Height: 780, View: view}); e != nil {
				t.Fatal(e)
			}
			m, e := NewManual(testContext, a, cfg)
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(func() { m.Close(testContext) })
			unitSettle(t, m)
			root := m.implementation.find("one").state.root
			opener := findRecord(root, "opener")
			for range 3 {
				c.Open = true
				a.Invalidate("one")
				unitSettle(t, m)
				if opener.focusVisibility.Load() != 2 || opener.widget.AsWidget().Styles.Border.Width.Top.Value != 0 {
					t.Fatal("background opener retained modal focus outline")
				}
				first := findRecord(root, ui.ToolbarCustomizationLaterID(c.ID, "a")).widget.AsWidget()
				next := findRecord(root, ui.ToolbarCustomizationEarlierID(c.ID, "b")).widget.AsWidget()
				box := first.Geom.Size.Actual.Total
				first.SetFocus()
				unitSettle(t, m)
				if first.Styles.Border.Width.Top.Value != 2 || !first.ContainsFocus() {
					t.Fatal("focused reorder button lacks cue", first.Styles.Border.Width.Top, first.ContainsFocus(), first.Styles.State, findRecord(root, ui.ToolbarCustomizationLaterID(c.ID, "a")).focusVisibility.Load(), first.Scene.SceneGeom, first.Styles.MaxBorder.Width.Top, controlDisabled(findRecord(root, ui.ToolbarCustomizationLaterID(c.ID, "a"))), nativeWindowActive(findRecord(root, ui.ToolbarCustomizationLaterID(c.ID, "a"))))
				}
				next.SetFocus()
				unitSettle(t, m)
				first.SetState(true, states.Focused)
				first.Restyle()
				unitSettle(t, m)
				if first.Styles.Border.Width.Top.Value != 1 || next.Styles.Border.Width.Top.Value != 2 || first.Geom.Size.Actual.Total != box {
					t.Fatal("stale focus cue/geometry", first.Styles.Border.Width.Top, next.Styles.Border.Width.Top, first.Geom.Size.Actual.Total, box)
				}
				c.Open = false
				a.Invalidate("one")
				unitSettle(t, m)
				if opener.focusVisibility.Load() != 1 || !opener.widget.AsWidget().ContainsFocus() {
					t.Fatal("closing panel did not restore current scope/focus")
				}
			}
		})
	}
}
