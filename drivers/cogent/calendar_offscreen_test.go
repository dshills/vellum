//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
	"time"
)

func TestCalendarNativeArrowsSkipDisabledAndChooseOnce(t *testing.T) {
	value := time.Date(2026, 3, 16, 12, 30, 0, 0, time.UTC)
	c := &ui.Calendar{ID: "calendar", Value: value, AllowDate: func(v time.Time) bool { return v.Day() != 17 }}
	calls := 0
	c.Change = func(e ui.CalendarEvent) { calls++; c.Value = e.Value }
	view := func() ui.Node {
		n, err := c.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.CalendarDayID(c.ID, value), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 828), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	for _, code := range []key.Codes{key.CodeRightArrow, key.CodeReturnEnter} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if calls != 1 || c.Value.Day() != 18 || c.Value.Hour() != 12 || c.Value.Minute() != 30 {
		t.Fatal("grid navigation/selection failed", calls, c.Value)
	}
}
