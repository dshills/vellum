//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
)

func TestReducedMotionToolbarDragKeepsPreviewAndCommit(t *testing.T) {
	cfg := manualConfig(t, 958)
	cfg.Theme = ui.MacTheme(false)
	cfg.Theme.ReducedMotion = true
	cfg.Theme.Default.Background = ui.Some(cfg.Theme.Palette.Value.Window)
	cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
	c := ui.ToolbarCustomizer{ID: "custom", Open: true, Toolbar: ui.Node{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{{ID: "a", Kind: ui.Button, Text: "A"}, {ID: "b", Kind: ui.Button, Text: "B"}, {ID: "c", Kind: ui.Button, Text: "C"}}}}
	changes := 0
	c.Change = func(v ui.ToolbarLayout) { changes++; c.Value = v }
	c.Dismiss = func() { c.Open = false }
	m, _ := simpleManual(t, cfg, func() ui.Node {
		n, e := c.View()
		if e != nil {
			t.Fatal(e)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{n}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	_, grip := imageManualElement(t, m, ui.ToolbarCustomizationDragID(c.ID, "b"))
	start := grip.Bounds.Min.Add(grip.Bounds.Size().Div(2))
	_, target := imageManualElement(t, m, ui.ToolbarCustomizationDragID(c.ID, "a"))
	end := image.Pt(start.X, target.Bounds.Min.Y+1)
	send := func(typ events.Types, p image.Point) {
		if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: p}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
	}
	send(events.MouseDown, start)
	send(events.MouseMove, end)
	binding := findRecord(w.state.root, ui.ToolbarCustomizationDragID(c.ID, "b")).control.toolbarDrag
	g := binding.active
	if g == nil || !g.hasLine || g.animation != nil || g.alpha != 1 || g.lineY != g.targetY {
		t.Fatal("reduced motion lost immediate insertion feedback", g)
	}
	send(events.MouseUp, end)
	if changes != 1 || binding.active != nil || c.Value.Order[0] != "b" {
		t.Fatal("motion preference broke reorder or cleanup", changes, c.Value)
	}
}
