//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"slices"
	"testing"
)

func TestToolbarCustomizationNativeReorderHideRestoreAndLifecycle(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			cfg := manualConfig(t, 7)
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			source := fullColorButtonSource(t, false)
			calls, changes := 0, 0
			c := ui.ToolbarCustomizer{ID: "customize", ReturnFocusID: "opener", Style: cfg.Theme.Kinds[ui.FormGroup]}
			c.Style.Padding = ui.Some(ui.UniformInsets(20))
			c.Dismiss = func() { c.Open = false }
			c.Change = func(next ui.ToolbarLayout) { changes++; c.Value = next }
			view := func() ui.Node {
				items := []ui.Node{}
				for i, name := range []string{"one", "two", "three", "disabled", "extra", "last"} {
					if i == 2 {
						items = append(items, ui.Node{ID: "rule", Kind: ui.Separator})
					}
					items = append(items, ui.Node{ID: name, Kind: ui.Button, Text: name, IconImage: source, Disabled: name == "disabled", Activate: func() { calls++ }, Style: ui.Style{GrowX: ui.Some(float32(0))}})
				}
				c.Toolbar = ui.Node{ID: "tools", Kind: ui.Toolbar, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 220})}, Children: items}
				configured, e := c.Value.Apply(c.Toolbar)
				if e != nil {
					panic(e)
				}
				if len(configured.Children) > 0 {
					configured = ui.WithToolbarOverflow(configured)
				}
				panel, e := c.View()
				if e != nil {
					panic(e)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "opener", Sequence: 1}, Children: []ui.Node{
					{ID: "opener", Kind: ui.Button, Text: "Customize Toolbar…", Activate: func() { c.Open = true }}, configured, panel,
				}}
			}
			a, e := app.New(app.Config{})
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(a.Close)
			if e = a.Open(app.Window{ID: "one", Title: "Toolbar customization", Width: 660, Height: 780, View: view}); e != nil {
				t.Fatal(e)
			}
			m, e := NewManual(testContext, a, cfg)
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(func() {
				if e := m.Close(testContext); e != nil {
					t.Error(e)
				}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			root := w.state.root
			retained := findRecord(root, "three").widget
			tap := func(id string) {
				t.Helper()
				window, target := imageManualElement(t, m, id)
				if !target.Visible || !target.Enabled || target.Bounds.Empty() {
					t.Fatal("unavailable routed target", id, target)
				}
				for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
					if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Target: id, TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}); e != nil {
						t.Fatal(e)
					}
				}
				unitSettle(t, m)
			}
			tap("opener")
			panel := findRecord(root, "customize")
			if !panel.control.dialog.active {
				t.Fatal("customizer not modal")
			}
			arrowX := -1
			for _, id := range []string{"one", "two", "rule", "three", "disabled", "extra", "last"} {
				checkbox := findRecord(root, ui.ToolbarCustomizationVisibilityID(c.ID, id)).widget.(*core.Switch)
				indicator := checkbox.ChildByName("stack").(*core.Frame)
				if indicator.Geom.TotalBBox.Min.X-checkbox.Geom.ContentBBox.Min.X > 2 {
					t.Fatal("visibility checkbox content centered", id, indicator.Geom.TotalBBox, checkbox.Geom.ContentBBox)
				}
				arrow := findRecord(root, ui.ToolbarCustomizationEarlierID(c.ID, id)).widget.AsWidget().Geom.TotalBBox
				if arrowX >= 0 && arrow.Min.X != arrowX {
					t.Fatal("reorder columns do not align", id, arrow.Min.X, arrowX)
				}
				arrowX = arrow.Min.X
				row := checkbox.Parent.(core.Widget).AsWidget()
				viewport := findRecord(root, c.ID+"/items").widget.AsWidget().Geom.ContentBBox
				if row.Geom.TotalBBox.Empty() || !row.Geom.TotalBBox.In(viewport) || float32(row.Geom.TotalBBox.Dy())+1 < row.Geom.Size.Actual.Total.Y {
					t.Fatal("customization row clipped", id, row.Geom.TotalBBox, viewport, row.Geom.Size.Actual.Total)
				}
			}
			earlier := ui.ToolbarCustomizationEarlierID(c.ID, "three")
			tap(earlier)
			tap(earlier)
			if !slices.Equal(c.Value.Order, []string{"one", "three", "two", "rule", "disabled", "extra", "last"}) || changes != 2 {
				t.Fatal("reorder did not merge exactly once", c.Value, changes)
			}
			if findRecord(root, "three").widget != retained {
				t.Fatal("reorder replaced original action")
			}
			if retained.(*core.Button).ChildByName(buttonImageName).(*managedImage).source != source {
				t.Fatal("reorder lost original artwork")
			}
			tap(ui.ToolbarCustomizationVisibilityID(c.ID, "one"))
			if findRecord(root, "one") != nil || !slices.Contains(c.Value.Hidden, "one") {
				t.Fatal("hide retained action or lost setting")
			}
			if calls != 0 {
				t.Fatal("customization activated a toolbar command")
			}
			staleNodes, staleHandlers := accessibilitySnapshot(root, nil)
			_ = staleNodes
			oldMove := staleHandlers[accessibilityManagedID(earlier)]
			if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeEscape, Modifiers: 1 << key.NumLock}); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			if c.Open || panel.control.dialog.active || !findRecord(root, "opener").widget.AsWidget().ContainsFocus() {
				t.Fatal("Escape did not dismiss and restore focus")
			}
			before := changes
			if oldMove != nil {
				oldMove(accessibilityAction{Kind: "press"})
				unitSettle(t, m)
			}
			if changes != before {
				t.Fatal("closed panel accepted stale native action")
			}
			tap("opener")
			for _, id := range []string{"three", "two", "disabled", "extra", "last"} {
				tap(ui.ToolbarCustomizationVisibilityID(c.ID, id))
			}
			if bar := findRecord(root, "tools"); len(bar.children) != 0 {
				t.Fatal("all-hidden toolbar retained actions/overflow")
			}
			tap(ui.ToolbarCustomizationRestoreID(c.ID))
			if len(c.Value.Hidden) != 0 || !slices.Equal(c.Value.Order, []string{"one", "two", "rule", "three", "disabled", "extra", "last"}) {
				t.Fatal("restore defaults failed", c.Value)
			}
			tap(ui.ToolbarCustomizationDoneID(c.ID))
			if c.Open || !findRecord(root, "opener").widget.AsWidget().ContainsFocus() {
				t.Fatal("Done did not restore opener")
			}
			tap("opener")
			c.Disabled = true
			if e := a.Invalidate("one"); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			if panel.control.dialog.active {
				t.Fatal("disabled customizer retained modal")
			}
			c.Disabled = false
			c.Open = false
			a.Invalidate("one")
			unitSettle(t, m)
			tap("opener")
			if e := m.Close(testContext); e != nil {
				t.Fatal(e)
			}
			if panel.control.dialog.active || panel.control.dialog.body != nil {
				t.Fatal("window close retained customizer")
			}
			if calls != 0 {
				t.Fatal("configuration invoked toolbar actions", calls)
			}
		})
	}
}
