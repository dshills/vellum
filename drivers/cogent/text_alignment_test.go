//go:build offscreen

package cogent

import (
	"image"
	"image/color"
	"math"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
)

// These tests inspect completed software-rendered glyph pixels and retained
// widget identity, rather than treating a native style enum as layout proof.
func TestManagedTextAlignmentPixelsKeyedUpdatesAndRemoval(t *testing.T) {
	for _, kind := range []ui.Kind{ui.Label, ui.Button, ui.MenuButton} {
		t.Run(string(kind), func(t *testing.T) {
			alignment := ui.Optional[ui.Alignment]{}
			view := func() ui.Node {
				n := ui.Node{ID: "aligned", Kind: kind, Text: "12345", Style: ui.Style{
					MinSize: ui.Some(ui.Size{Width: 240, Height: 44}), MaxSize: ui.Some(ui.Size{Width: 240, Height: 44}),
					Padding: ui.Some(ui.Insets{}), Grow: ui.Some(float32(0)), TextAlign: alignment,
					Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}),
					Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Typography: ui.Some(ui.Typography{Size: 24}),
					Border: ui.Some(ui.Border{}), CornerRadius: ui.Some(float32(0)),
				}}
				if kind != ui.Label {
					n.Appearance = ui.ButtonQuiet
					n.Activate = func() {}
				}
				if kind == ui.MenuButton {
					n.Activate = nil
					n.Options = []ui.Option{{ID: "one", Label: "One"}}
					n.Change = func(ui.ChangeEvent) {}
				}
				return n
			}
			m, a := simpleManual(t, manualConfig(t, 7), view)
			unitSettle(t, m)
			r := findRecord(m.implementation.find("one").state.root, "aligned")
			widget := r.widget
			capture := func() image.Rectangle {
				t.Helper()
				frame, err := m.Capture(testContext, "one")
				if err != nil {
					t.Fatal(err)
				}
				bounds := widget.AsWidget().Geom.ContentBBox.Intersect(frame.Bounds())
				ink := image.Rectangle{}
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						c := frame.RGBAAt(x, y)
						if c.R < 90 && c.G < 90 && c.B < 90 {
							p := image.Rect(x, y, x+1, y+1)
							if ink.Empty() {
								ink = p
							} else {
								ink = ink.Union(p)
							}
						}
					}
				}
				if ink.Empty() {
					t.Fatal("no completed glyph pixels")
				}
				return ink
			}
			baseline := capture()
			if kind == ui.Label && widget.(*core.Text).AllocationLayout {
				t.Fatal("omitted allocation style enabled new layout")
			}
			positions := map[ui.Alignment]image.Rectangle{}
			for _, value := range []ui.Alignment{ui.AlignStart, ui.AlignEnd, ui.AlignCenter} {
				if err := a.Post(func() { alignment = ui.Some(value); _ = a.Invalidate("one") }); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				positions[value] = capture()
				if kind == ui.Label && !widget.(*core.Text).AllocationLayout {
					t.Fatal("explicit alignment did not enable allocation layout")
				}
				if findRecord(m.implementation.find("one").state.root, "aligned").widget != widget {
					t.Fatal("alignment update replaced keyed widget")
				}
			}
			start, end, center := positions[ui.AlignStart], positions[ui.AlignEnd], positions[ui.AlignCenter]
			if end.Min.X-start.Min.X < 100 || math.Abs(float64(center.Min.X*2-start.Min.X-end.Min.X)) > 3 {
				t.Fatalf("glyph content did not align across allocation: start=%v center=%v end=%v", start, center, end)
			}
			if err := a.Post(func() { alignment = ui.Optional[ui.Alignment]{}; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if kind == ui.Label && widget.(*core.Text).AllocationLayout {
				t.Fatal("removing alignment retained allocation layout")
			}
			if got := capture(); got != baseline {
				t.Fatalf("removing alignment did not restore native geometry: %v want %v", got, baseline)
			}
		})
	}
}

func TestManagedGrowingLabelCentersInsideCenteredWideColumn(t *testing.T) {
	view := func() ui.Node {
		return ui.Node{ID: "wide-root", Kind: ui.Column, Grow: true, Children: []ui.Node{
			{ID: "weather-hero", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1)), Align: ui.Some(ui.AlignCenter), Padding: ui.Some(ui.Insets{Top: 24, Bottom: 24})}, Children: []ui.Node{
				{ID: "growing-heading", Kind: ui.Label, Text: "Cupertino", Style: ui.Style{GrowX: ui.Some(float32(1)), TextAlign: ui.Some(ui.AlignCenter), Typography: ui.Some(ui.Typography{Size: 29})}},
			}},
		}}
	}
	m, _ := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	hero := findRecord(root, "weather-hero").widget.AsWidget()
	label := findRecord(root, "growing-heading").widget.(*core.Text)
	frame, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	bounds := hero.Geom.ContentBBox.Intersect(frame.Bounds())
	ink := image.Rectangle{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := frame.RGBAAt(x, y)
			if c.R < 90 && c.G < 90 && c.B < 90 {
				p := image.Rect(x, y, x+1, y+1)
				if ink.Empty() {
					ink = p
				} else {
					ink = ink.Union(p)
				}
			}
		}
	}
	if ink.Empty() {
		t.Fatal("no centered heading glyph pixels")
	}
	if math.Abs(float64(ink.Min.X+ink.Max.X-bounds.Min.X-bounds.Max.X)) > 4 {
		t.Fatalf("growing heading did not center in parent: ink=%v parent=%v textGeom=%+v shaped=%+v", ink, bounds, label.Geom, label.PaintText().Bounds)
	}
}
