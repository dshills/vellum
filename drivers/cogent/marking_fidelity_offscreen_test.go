//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMarkingPanelFidelityGeometryFocusAndDismiss(t *testing.T) {
	for _, dpi := range []float32{160, 240, 320} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			cfg := manualConfig(t, 7)
			cfg.LogicalDPI = dpi
			bg := cfg.Theme.Default.Background
			cfg.Theme = ui.MacTheme(false)
			cfg.Theme.Default.Background = bg
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			chosen, dismissed := 0, 0
			p := &ui.MarkingMenu{ID: "panel", Label: "Workbench actions"}
			for i := 0; i < 8; i++ {
				p.Actions = append(p.Actions, ui.MarkingAction{ID: fmt.Sprint(i), Label: fmt.Sprint("Action ", i), IconImage: fullColorButtonSource(t, false)})
			}
			p.Choose = func(string) { chosen++; p.Open = false }
			p.Dismiss = func() { dismissed++; p.Open = false }
			view := func() ui.Node {
				n, e := p.View()
				if e != nil {
					panic(e)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "opener", Sequence: 1}, Children: []ui.Node{{ID: "opener", Kind: ui.Button, Text: "Actions", Activate: func() { p.Open = true }}, n}}
			}
			a, e := app.New(app.Config{})
			if e != nil {
				t.Fatal(e)
			}
			defer a.Close()
			if e = a.Open(app.Window{ID: "one", Title: "Panel geometry", Width: 1200, Height: 900, View: view}); e != nil {
				t.Fatal(e)
			}
			m, e := NewManual(testContext, a, cfg)
			if e != nil {
				t.Fatal(e)
			}
			defer m.Close(testContext)
			unitSettle(t, m)
			w := m.implementation.find("one")
			click := func(id string) {
				t.Helper()
				_, el := imageManualElement(t, m, id)
				for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
					if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Target: id, TargetIncarnation: el.Incarnation, Type: kind, Button: events.Left}); e != nil {
						t.Fatal(e)
					}
				}
				unitSettle(t, m)
			}
			click("opener")
			panel := findRecord(w.state.root, "panel")
			if !panel.control.dialog.active {
				t.Fatal("panel closed")
			}
			north := findRecord(w.state.root, "panel/cell/0/1")
			east := findRecord(w.state.root, "panel/cell/1/2")
			if north == nil || east == nil || north.widget.AsWidget().Geom.TotalBBox.Empty() {
				t.Fatal("missing compact actions")
			}
			aBox, bBox := north.widget.AsWidget().Geom.TotalBBox, east.widget.AsWidget().Geom.TotalBBox
			if aBox.Overlaps(bBox) || float32(aBox.Dy()) > 72*dpi/160 {
				t.Fatal("panel targets overlap or oversized", aBox, bBox)
			}
			if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeEscape}); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			if p.Open || dismissed != 1 || chosen != 0 || panel.control.dialog.active {
				t.Fatal("Escape/dismiss lifecycle", p.Open, dismissed, chosen)
			}
			click("opener")
			click("panel/cell/0/1")
			if chosen != 1 || p.Open {
				t.Fatal("action duplicated or panel retained")
			}
			p.Open = true
			if e := a.Invalidate("one"); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			if e := m.Close(testContext); e != nil {
				t.Fatal(e)
			}
			if panel.control.dialog.active || panel.control.dialog.body != nil {
				t.Fatal("modal retained on close")
			}
		})
	}
}
