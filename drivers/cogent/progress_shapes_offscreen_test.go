//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dshills/vellum/ui"
)

func TestProgressShapesCompletedFramesAcrossDPI(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			fraction := float32(0)
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			view := func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(24))}, Children: []ui.Node{
					{ID: "ring", Kind: ui.ProgressRing, Number: fraction, Text: fmt.Sprintf("%.0f%%", fraction*100), Accessibility: ui.Accessibility{Name: "Circular progress"}},
					{ID: "gauge", Kind: ui.ProgressGauge, Number: fraction, Text: fmt.Sprintf("%.0f%%", fraction*100), Accessibility: ui.Accessibility{Name: "Semicircular progress"}},
				}}
			}
			m, a := simpleManual(t, config, view)
			unitSettle(t, m)
			before, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			initial, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			identities := map[string]uint64{}
			for _, e := range initial.Windows[0].Elements {
				identities[e.ID] = e.Incarnation
			}
			for _, value := range []float32{0.25, 0.75, 1} {
				if err = a.Do(func() { fraction = value; _ = a.Invalidate("one") }); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				after, err := m.Capture(testContext, "one")
				if err != nil {
					t.Fatal(err)
				}
				snapshot, err := m.Snapshot(testContext)
				if err != nil {
					t.Fatal(err)
				}
				if snapshot.Windows[0].Timers != 0 || snapshot.Windows[0].FuturePaint {
					t.Fatal("passive progress introduced ongoing work", snapshot.Windows[0])
				}
				found := 0
				for _, e := range snapshot.Windows[0].Elements {
					if e.ID == "ring" || e.ID == "gauge" {
						found++
						if e.Incarnation != identities[e.ID] || e.Value != value || e.Bounds.Empty() || !e.Visible || e.Focused {
							t.Fatal("progress shape identity/value/geometry", e)
						}
						changed := 0
						for y := e.Bounds.Min.Y; y < e.Bounds.Max.Y; y++ {
							for x := e.Bounds.Min.X; x < e.Bounds.Max.X; x++ {
								if before.RGBAAt(x, y) != after.RGBAAt(x, y) {
									changed++
								}
							}
						}
						if changed == 0 {
							t.Fatal("accepted fraction did not change actual shape pixels", e.ID, value)
						}
					}
				}
				if found != 2 || bytes.Equal(before.Pix, after.Pix) {
					t.Fatal("progress frame missing changes")
				}
				before = after
			}
		})
	}
}
