//go:build offscreen

package cogent

import (
	"image"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
)

func TestMacNumberStepperPointerKeyboardBoundsAndRetention(t *testing.T) {
	value := float32(1)
	requests := 0
	reject, disabled := false, false
	buttons := ui.NumberButtonsStepper
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "number", Sequence: 1}, Children: []ui.Node{
			{ID: "number", Kind: ui.NumberField, Number: value, Numeric: ui.Numeric{Min: 0, Max: 3, Step: 1}, NumberButtons: buttons, Disabled: disabled, Accessibility: ui.Accessibility{Name: "Copies"},
				Change: func(e ui.ChangeEvent) {
					requests++
					if !reject {
						value = e.Number
					}
				}},
		}}
	}
	cfg := manualConfig(t, 7)
	theme := ui.MacTheme(false)
	theme.Default.Typography = cfg.Theme.Default.Typography
	theme.Default.Background = cfg.Theme.Default.Background
	cfg.Theme = theme
	m, a := simpleManual(t, cfg, view)
	unitSettle(t, m)
	record := findRecord(m.implementation.find("one").state.root, "number")
	spinner := record.widget.(*core.Spinner)
	var lastBox image.Rectangle
	click := func(upper bool) {
		t.Helper()
		window, _ := imageManualElement(t, m, "number")
		if part := spinner.ChildByName("trail-icon", 0); part != nil {
			lastBox = part.(core.Widget).AsWidget().Geom.TotalBBox
		}
		box := lastBox
		if box.Empty() {
			t.Fatal("stepper has no arrow geometry")
		}
		y := box.Min.Y + box.Dy()/4
		if !upper {
			y = box.Min.Y + 3*box.Dy()/4
		}
		point := image.Pt(box.Min.X+box.Dx()/2, y)
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: typ, Button: events.Left, Position: point}); err != nil {
				t.Fatal(err)
			}
		}
		unitSettle(t, m)
	}
	if !spinner.ChildByName("lead-icon", 0).(core.Widget).AsWidget().StateIs(states.Invisible) || spinner.TrailingIcon != icons.UnfoldMore {
		t.Fatal("horizontal buttons still present")
	}
	click(true)
	if value != 2 || requests != 1 {
		t.Fatal("upper half", value, requests)
	}
	click(false)
	if value != 1 || requests != 2 {
		t.Fatal("lower half", value, requests)
	}
	if err := a.Do(func() { value = 3; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	click(true)
	if value != 3 {
		t.Fatal("maximum wrapped", value)
	}
	window, _ := imageManualElement(t, m, "number")
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.KeyChord, Code: key.CodeDownArrow}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if value != 2 {
		t.Fatal("keyboard down", value)
	}
	if err := a.Do(func() { value = 1; reject = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	click(true)
	if value != 1 || spinner.Value != 1 {
		t.Fatal("rejected step changed accepted/native value", value, spinner.Value)
	}

	if err := a.Do(func() { value = 3; reject = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	heldWindow, _ := imageManualElement(t, m, "number")
	heldBox := spinner.ChildByName("trail-icon", 0).(core.Widget).AsWidget().Geom.TotalBBox
	heldPoint := image.Pt(heldBox.Min.X+heldBox.Dx()/2, heldBox.Min.Y+3*heldBox.Dy()/4)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: heldWindow.Incarnation, Type: events.MouseDown, Button: events.Left, Position: heldPoint}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if _, err := cfg.Clock.Advance(8*core.TimingSettings.RepeatClickTime + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if value != 2 {
		t.Fatal("held lower arrow reversed direction", value)
	}
	if _, err := cfg.Clock.Advance(core.TimingSettings.RepeatClickTime + time.Millisecond); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if value != 1 {
		t.Fatal("second held lower tick reversed direction", value)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: heldWindow.Incarnation, Type: events.MouseUp, Button: events.Left, Position: heldPoint}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if value != 1 {
		t.Fatal("held release direction", value)
	}
	before := requests
	if err := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	click(true)
	if requests != before || value != 1 {
		t.Fatal("disabled step reached model")
	}
	if err := a.Do(func() { disabled = false; buttons = ""; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if findRecord(m.implementation.find("one").state.root, "number").widget != spinner || spinner.LeadingIcon != icons.Remove || spinner.TrailingIcon != icons.Add || spinner.StepFunc != nil || spinner.ChildByName("lead-icon", 0).(core.Widget).AsWidget().StateIs(states.Invisible) {
		t.Fatal("inline presentation did not restore without replacing field")
	}
}
