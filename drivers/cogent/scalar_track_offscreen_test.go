//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"math"
	"testing"
)

func TestGradientTrackNativeCaptureScaleAndRevocation(t *testing.T) {
	for _, dpi := range []float32{160, 320} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			g := &ui.GradientWell{ID: "gradient", Label: "Gradient", Value: ui.GradientValue{Stops: []ui.GradientStop{{ID: "a", Position: 0, Color: color.NRGBA{R: 255, A: 255}}, {ID: "b", Position: .5, Color: color.NRGBA{B: 255, A: 255}}, {ID: "c", Position: 1, Color: color.NRGBA{G: 255, A: 255}}}}}
			calls := 0
			g.Change = func(e ui.GradientEdit) {
				v, err := e.Apply(g.Value)
				if err != nil {
					t.Error(err)
					return
				}
				g.Value = v
				calls++
			}
			n, _ := g.View()
			findTrackNode(n, ui.GradientWellButtonID(g.ID)).Activate()
			view := func() ui.Node {
				n, e := g.View()
				if e != nil {
					t.Fatal(e)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.GradientTrackID(g.ID), Sequence: 1}, Children: []ui.Node{n}}
			}
			cfg := manualConfig(t, 945)
			cfg.LogicalDPI = dpi
			m, a := simpleManual(t, cfg, view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			face := findRecord(w.state.root, ui.GradientTrackID(g.ID)).widget.(*core.Button).ChildByName("scalar-track").(*core.Canvas)
			rect := face.Geom.ContentBBox.Add(face.Scene.SceneGeom.Pos)
			at := func(f float32) image.Point {
				inset := float32(rect.Dx()) * 12 / 320
				return image.Pt(rect.Min.X+int(inset+f*(float32(rect.Dx())-2*inset)), rect.Min.Y+rect.Dy()/2)
			}
			input := func(typ events.Types, p image.Point) {
				if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: p}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
			}
			input(events.MouseDown, at(.5))
			input(events.MouseMove, at(.8))
			before := calls
			input(events.MouseUp, at(.8))
			if math.Abs(float64(g.Value.Stops[1].Position-.8)) > .011 || calls != before {
				t.Fatal("transformed move/release", g.Value, calls, before)
			}
			if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeRightArrow, Modifiers: 1 << key.NumLock}); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			if g.Value.Stops[1].Position < .8 {
				t.Fatal("lock key")
			}
			input(events.MouseDown, at(.8))
			before = calls
			face.Scene.Events.CancelPointerCapture()
			input(events.MouseMove, at(.2))
			input(events.MouseUp, at(.2))
			if calls != before {
				t.Fatal("lost capture changed gradient")
			}
			input(events.MouseDown, at(.8))
			before = calls
			if e := a.Do(func() { g.Disabled = true; _ = a.Invalidate("one") }); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			input(events.MouseUp, at(.2))
			if calls != before {
				t.Fatal("disabled gesture committed")
			}
		})
	}
}
func findTrackNode(n ui.Node, id string) ui.Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := findTrackNode(c, id); found.ID != "" {
			return found
		}
	}
	return ui.Node{}
}
