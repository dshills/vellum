//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
)

func TestWorkbenchFidelityCompactGeometryAndRouting(t *testing.T) {
	for _, dpi := range []float32{160, 240, 320} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			source := fullColorButtonSource(t, false)
			var browserEvents []ui.ModelBrowserEvent
			var timelineEvents []ui.FeatureTimelineEvent
			b := &ui.ModelBrowser{ID: "browser", Label: "BROWSER", Items: []ui.ModelBrowserItem{{ID: "assembly", Label: "Assembly", Visible: true, Activatable: true}, {ID: "part", Parent: "assembly", Label: "Carriage", Visible: true, Activatable: true, IconImage: source}}, SelectedID: "part", ExpandedIDs: []string{"assembly"}, Change: func(e ui.ModelBrowserEvent) { browserEvents = append(browserEvents, e) }}
			f := &ui.FeatureTimeline{ID: "history", Label: "History", Items: []ui.FeatureItem{{ID: "f0", Label: "Create base", IconImage: source}, {ID: "f1", Label: "Create carriage", IconImage: source}}, SelectedID: "f1", Marker: 2, Change: func(e ui.FeatureTimelineEvent) { timelineEvents = append(timelineEvents, e) }}
			a, _ := app.New(app.Config{})
			defer a.Close()
			_ = a.Open(app.Window{ID: "one", Title: "Compact workbench", Width: 800, Height: 500, View: func() ui.Node {
				bn, e := b.View()
				if e != nil {
					panic(e)
				}
				bn.Style.GrowY = ui.Some(float32(1))
				fn, e := f.View()
				if e != nil {
					panic(e)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: ui.Style{Gap: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{})}, Children: []ui.Node{bn, fn}}
			}})
			cfg := manualConfig(t, 7)
			cfg.LogicalDPI = dpi
			cfg.Theme = ui.MacTheme(false)
			cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			m, e := NewManual(testContext, a, cfg)
			if e != nil {
				t.Fatal(e)
			}
			defer m.Close(testContext)
			unitSettle(t, m)
			w := m.implementation.find("one")
			scale := dpi / 160
			row := findRecord(w.state.root, ui.ModelBrowserRowID("browser", "part"))
			if row == nil {
				t.Fatal("missing row")
			}
			box := row.widget.AsWidget().Geom.TotalBBox
			if float32(box.Dy()) > 32*scale {
				t.Fatal("browser row lost compact geometry", box, dpi)
			}
			history := findRecord(w.state.root, "history")
			if history == nil {
				t.Fatal("missing history strip")
			}
			if float32(history.widget.AsWidget().Geom.TotalBBox.Dy()) > 48*scale {
				t.Fatal("timeline became a card panel", history.widget.AsWidget().Geom.TotalBBox, dpi)
			}
			var previous image.Rectangle
			for _, suffix := range []string{"/visibility", "/select", "/active"} {
				c := findRecord(w.state.root, ui.ModelBrowserRowID("browser", "part")+suffix)
				if c == nil {
					t.Fatal("missing target", suffix)
				}
				bounds := c.widget.AsWidget().Geom.TotalBBox
				if bounds.Empty() || !bounds.In(box) || !previous.Empty() && bounds.Min.X < previous.Max.X {
					t.Fatal("dense targets overlap/clip", suffix, box, previous, bounds)
				}
				previous = bounds
			}
			click := func(id string) {
				t.Helper()
				_, el := imageManualElement(t, m, id)
				for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
					if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}); e != nil {
						t.Fatal(e)
					}
					unitSettle(t, m)
				}
			}
			click(ui.ModelBrowserRowID("browser", "part") + "/visibility")
			click(ui.ModelBrowserRowID("browser", "part") + "/select")
			click(ui.ModelBrowserRowID("browser", "part") + "/active")
			if len(browserEvents) != 3 || browserEvents[0].Kind != ui.ModelBrowserVisibility || browserEvents[1].Kind != ui.ModelBrowserSelect || browserEvents[2].Kind != ui.ModelBrowserActivate {
				t.Fatal(browserEvents)
			}
			click("history/earlier")
			click("history/back")
			if len(timelineEvents) != 2 || timelineEvents[0].Action != "move" || timelineEvents[0].Index != 0 || timelineEvents[1].Action != "marker" || timelineEvents[1].Index != 1 {
				t.Fatal(timelineEvents)
			}
		})
	}
}
