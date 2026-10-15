//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"fmt"
	"image"
	"math"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"github.com/dshills/vellum/workbench"
)

func manualDockDrag(t *testing.T, m *Manual, batch int, from, to image.Point) {
	t.Helper()
	s, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := s.Windows[0]
	for _, step := range []struct {
		kind  events.Types
		point image.Point
	}{{events.MouseDown, from}, {events.MouseMove, from.Add(to.Sub(from).Div(2))}, {events.MouseMove, to}, {events.MouseUp, to}} {
		if err = m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: step.kind, Position: step.point, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
		manualDockSettle(t, m, batch)
		if step.kind == events.MouseDown {
			if _, err = m.implementation.config.Clock.Advance(core.TimingSettings.SlideStartTime + time.Millisecond); err != nil {
				t.Fatal(err)
			}
			manualDockSettle(t, m, batch)
		}
	}
}

func TestManualWorkbenchActualTabDragAndSplitResize(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			m, a, dock, _ := manualDockFixture(t, batch)
			_, right := manualDockElement(t, m, "one", ManualWorkbenchTabID("dock", "right"))
			_, left := manualDockElement(t, m, "one", ManualWorkbenchTabID("dock", "left"))
			leftWidget, rightWidget := dock.records["left"].widget, dock.records["right"].widget
			manualDockDrag(t, m, batch, right.Bounds.Min.Add(right.Bounds.Size().Div(2)), left.Bounds.Min.Add(left.Bounds.Size().Div(3)))
			if got := dock.Workspace.Layout().Root.Tabs; len(got) != 2 || got[0] != "right" {
				t.Fatal("actual drag did not reorder model", got)
			}
			if dock.records["left"].widget != leftWidget || dock.records["right"].widget != rightWidget {
				t.Fatal("same-window drag destroyed retained panel controls")
			}
			if err := a.Do(func() {
				if err := dock.Workspace.Restore(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "split", Axis: workbench.Horizontal, Ratio: .6, First: &workbench.Group{ID: "first", Tabs: []string{"left"}, Active: "left"}, Second: &workbench.Group{ID: "second", Tabs: []string{"right"}, Active: "right"}}}); err != nil {
					t.Error(err)
				}
				a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			manualDockSettle(t, m, batch)
			// Pilot coordinates come from the actual native divider handle, then
			// mouse events use raw physical coordinates through the normal router.
			handle := dock.splits["split"].Parts.Child(0)
			box := handle.(core.Widget).AsWidget().Geom.TotalBBox
			from := box.Min.Add(box.Size().Div(2))
			manualDockDrag(t, m, batch, from, from.Add(image.Pt(80, 0)))
			ratio := dock.Workspace.Layout().Root.Ratio
			if math.Abs(ratio-.6) < .02 || ratio <= .05 || ratio >= .95 {
				t.Fatal("actual divider drag did not update accepted ratio", ratio)
			}
			if got := dock.splits["split"].Splits(); len(got) != 2 || math.Abs(float64(got[0])-ratio) > .001 {
				t.Fatal("accepted divider ratio not reconciled", got, ratio)
			}
		})
	}
}
