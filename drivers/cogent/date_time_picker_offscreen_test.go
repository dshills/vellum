//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
	"time"
)

func TestDateTimePickerRoutedCalendarAndNativeCommit(t *testing.T) {
	p := &ui.DateTimePicker{ID: "due", Value: time.Date(2024, 2, 28, 9, 15, 0, 0, time.UTC)}
	calls, invalid := 0, 0
	p.Change = func(e ui.DateTimeEvent) { calls++; p.Value = e.Value }
	p.Invalid = func(error) { invalid++ }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{n}}
	}
	a, err := app.New(app.Config{QueueCapacity: 32})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	if err = a.Open(app.Window{ID: "one", Title: "Date", Width: 700, Height: 700, View: view}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 920))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	unitSettle(t, m)
	w, button := imageManualElement(t, m, ui.DateTimeCalendarID(p.ID))
	click := func(target ManualElement) {
		t.Helper()
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
				t.Fatal(err)
			}
		}
		unitSettle(t, m)
	}
	click(button)
	// February 2024 starts Thursday; day 29 occupies cell 32.
	_, day := imageManualElement(t, m, "due/date-time/cell/32")
	click(day)
	if p.Value.Day() != 29 || calls != 1 || p.Value.Hour() != 9 {
		t.Fatal("calendar selection", p.Value, calls)
	}
	root := m.implementation.find("one").state.root
	field := findRecord(root, ui.DateTimeInputID(p.ID)).widget.(*core.TextField)
	field.SetText("2024-02-30").SendChange()
	a.Drain()
	unitSettle(t, m)
	if invalid != 1 || calls != 1 || field.Text() != "2024-02-29" {
		t.Fatal("invalid text restore", invalid, calls, field.Text())
	}
	_, button = imageManualElement(t, m, ui.DateTimeCalendarID(p.ID))
	click(button)
	// The opening anchor is outside the modal scope, so route by observed position.
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Position: button.Bounds.Min.Add(button.Bounds.Size().Div(2)), Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if findRecord(root, "due/date-time/popup").control.dialog.active {
		t.Fatal("anchor outside-release reopened calendar")
	}
	_, button = imageManualElement(t, m, ui.DateTimeCalendarID(p.ID))
	click(button)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeEscape}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 1 {
		t.Fatal("Escape created change")
	}
	_, button = imageManualElement(t, m, ui.DateTimeCalendarID(p.ID))
	click(button)
	if err := a.Do(func() { p.Disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if err := a.Do(func() { p.Disabled = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if findRecord(root, "due/date-time/popup").control.dialog.active {
		t.Fatal("disabled picker reopened calendar")
	}
	if findRecord(root, ui.DateTimeInputID(p.ID)).widget != field {
		t.Fatal("field identity lost")
	}
}
