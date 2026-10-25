//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/ui"
	"image"
	"math"
	"testing"
)

func TestRangeTrackNativeBothHandlesKeyboardAXAndCapture(t *testing.T) {
	for _, dpi := range []float32{160, 320} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			r := &ui.RangeSlider{ID: "range", Label: "Range", Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Value: ui.RangeValue{Lower: 25, Upper: 75}, TickCount: 5}
			calls := 0
			r.Change = func(e ui.RangeEvent) {
				v, err := e.Apply(r.Value, r.Numeric)
				if err != nil {
					t.Error(err)
					return
				}
				r.Value = v
				calls++
			}
			view := func() ui.Node {
				n, e := r.View()
				if e != nil {
					t.Fatal(e)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.RangeTrackID(r.ID), Sequence: 1}, Children: []ui.Node{n}}
			}
			cfg := manualConfig(t, 946)
			cfg.LogicalDPI = dpi
			m, a := simpleManual(t, cfg, view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			record := findRecord(w.state.root, ui.RangeTrackID(r.ID))
			face := record.widget.(*core.Button).ChildByName("scalar-track").(*core.Canvas)
			rect := face.Geom.ContentBBox.Add(face.Scene.SceneGeom.Pos)
			at := func(f float32) image.Point {
				inset := float32(rect.Dx()) * 12 / 320
				return image.Pt(rect.Min.X+int(inset+f*(float32(rect.Dx())-2*inset)), rect.Min.Y+rect.Dy()/2)
			}
			input := func(typ events.Types, p image.Point) {
				if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Position: p, Button: events.Left}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
			}
			input(events.MouseDown, at(.25))
			input(events.MouseMove, at(.4))
			before := calls
			input(events.MouseUp, at(.4))
			if math.Abs(float64(r.Value.Lower-40)) > 1 || calls != before {
				t.Fatal("lower/release", r.Value, calls, before)
			}
			input(events.MouseDown, at(.75))
			input(events.MouseUp, at(.9))
			if r.Value.Upper < 89 {
				t.Fatal("coalesced final coordinate", r.Value)
			}
			if e := a.Do(func() {
				record = findRecord(w.state.root, ui.RangeTrackID(r.ID))
				if !applyAccessibilityAction(record, accessibilityAction{Kind: "decrement"}) {
					t.Error("AX decrement")
				}
			}); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
			if r.Value.Upper > 90 {
				t.Fatal("AX route")
			}
			for _, code := range []key.Codes{key.CodeUpArrow, key.CodeRightArrow} {
				if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Modifiers: 1 << key.NumLock}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
			}
			if r.Value.Lower < 40 {
				t.Fatal("select/lower key", r.Value)
			}
			input(events.MouseDown, at(.4))
			before = calls
			face.Scene.Events.CancelPointerCapture()
			input(events.MouseUp, at(.1))
			if calls != before {
				t.Fatal("capture loss")
			}
		})
	}
}
