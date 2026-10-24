//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
)

func TestAngleDialNativeKeysPointerCaptureAndLoss(t *testing.T) {
	for _, dpi := range []float32{160, 320} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			d := &ui.AngleDial{ID: "angle", Label: "Rotation", Degrees: 345, Step: 15, Size: 64}
			calls := 0
			d.Change = func(e ui.AngleEvent) { d.Degrees = e.Degrees; calls++ }
			view := func() ui.Node {
				n, err := d.View()
				if err != nil {
					t.Fatal(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.AngleDialButtonID(d.ID), Sequence: 1}, Children: []ui.Node{n}}
			}
			cfg := manualConfig(t, 913)
			cfg.LogicalDPI = dpi
			m, a := simpleManual(t, cfg, view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			keyInput := func(code key.Codes) {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			keyInput(key.CodeRightArrow)
			if d.Degrees != 0 {
				t.Fatal("keyboard did not wrap", d.Degrees)
			}
			keyInput(key.CodeLeftArrow)
			if d.Degrees != 345 {
				t.Fatal("reverse wrap", d.Degrees)
			}
			keyInput(key.CodeHome)
			if err := a.Do(func() {
				r := findRecord(w.state.root, ui.AngleDialButtonID(d.ID))
				if !applyAccessibilityAction(r, accessibilityAction{Kind: "increment"}) || d.Degrees != 15 {
					t.Error("AX increment route")
				}
				if !applyAccessibilityAction(r, accessibilityAction{Kind: "decrement"}) || d.Degrees != 0 {
					t.Error("AX decrement route")
				}
			}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			b := findRecord(w.state.root, ui.AngleDialButtonID(d.ID)).widget.(*core.Button)
			face := b.ChildByName("angle-face").(*core.Canvas)
			rect := face.Geom.ContentBBox.Add(face.Scene.SceneGeom.Pos)
			center := rect.Min.Add(rect.Size().Div(2))
			top := image.Pt(center.X, rect.Min.Y+rect.Dy()/8)
			right := image.Pt(rect.Max.X-rect.Dx()/8, center.Y)
			input := func(typ events.Types, p image.Point) {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Position: p, Button: events.Left}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			input(events.MouseDown, top)
			input(events.MouseMove, right)
			beforeRelease := calls
			input(events.MouseUp, right)
			if d.Degrees != 90 || calls != beforeRelease {
				t.Fatal("pointer angle/release click", d.Degrees, calls, beforeRelease)
			}
			input(events.MouseDown, top)
			input(events.MouseUp, right)
			if d.Degrees != 90 {
				t.Fatal("release coordinate was lost when moves were coalesced")
			}

			input(events.MouseDown, top)
			atLoss := calls
			face.Scene.Events.CancelPointerCapture()
			input(events.MouseMove, right)
			input(events.MouseUp, right)
			if calls != atLoss {
				t.Fatal("capture loss leaked a click or proposal", calls, atLoss)
			}
			input(events.MouseDown, top)
			if err := a.Do(func() { d.Disabled = true; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			beforeDisabled := calls
			if err := a.Do(func() {
				if applyAccessibilityAction(findRecord(w.state.root, ui.AngleDialButtonID(d.ID)), accessibilityAction{Kind: "increment"}) {
					t.Error("disabled AX action admitted")
				}
			}); err != nil {
				t.Fatal(err)
			}
			input(events.MouseMove, right)
			input(events.MouseUp, right)
			if calls != beforeDisabled {
				t.Fatal("disabled drag acted")
			}
		})
	}
}
