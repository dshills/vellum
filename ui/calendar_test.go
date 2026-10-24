package ui

import (
	"testing"
	"time"
)

func TestCalendarLeapBoundsTodayAndStaleMonth(t *testing.T) {
	value := time.Date(2024, 2, 20, 12, 30, 0, 0, time.UTC)
	c := &Calendar{ID: "calendar", Value: value, Today: value, FirstWeekday: time.Monday, AllowDate: func(v time.Time) bool { return v.Day() != 21 }}
	calls := 0
	c.Change = func(e CalendarEvent) { calls++; c.Value = e.Value }
	c.Navigate = func(e CalendarMonthEvent) { c.Month = e.Month }
	n, err := c.View()
	if err != nil {
		t.Fatal(err)
	}
	leap := pickerNode(n, CalendarDayID(c.ID, time.Date(2024, 2, 29, 12, 30, 0, 0, time.UTC)))
	if leap.ID == "" || leap.Disabled {
		t.Fatal("leap day absent")
	}
	blocked := pickerNode(n, CalendarDayID(c.ID, time.Date(2024, 2, 21, 12, 30, 0, 0, time.UTC)))
	blocked.Activate()
	if calls != 0 {
		t.Fatal("disabled predicate admitted")
	}
	leap.Activate()
	if c.Value.Day() != 29 || c.Value.Hour() != 12 || c.Value.Minute() != 30 {
		t.Fatal("civil fields changed")
	}
	pickerNode(n, "calendar/calendar/next").Activate()
	leap.Activate()
	if calls != 1 {
		t.Fatal("stale month chose date")
	}
	c.Month = time.Time{}
	n, _ = c.View()
	pickerNode(n, "calendar/calendar/today").Activate()
	if c.Value.Day() != 20 {
		t.Fatal("Today not accepted")
	}
}
func TestCalendarCellSkipsDisabledAndKeepsColumns(t *testing.T) {
	c := CalendarCell{Index: 7}
	c.IDs[9] = "right"
	c.IDs[21] = "down"
	if c.Neighbor(1) != "right" || c.Neighbor(7) != "down" || c.Neighbor(2) != "" {
		t.Fatal("grid navigation")
	}
}

func TestDateCalendarReopenRevokesOldDayAndDismiss(t *testing.T) {
	p := &DateTimePicker{ID: "due", Value: time.Date(2026, 3, 16, 12, 30, 0, 0, time.UTC)}
	calls := 0
	p.Change = func(e DateTimeEvent) { calls++; p.Value = e.Value }
	n, _ := p.View()
	pickerNode(n, DateTimeCalendarID(p.ID)).Activate()
	n, _ = p.View()
	oldDay := pickerNode(n, "due/date-time/cell/17")
	oldDismiss := pickerNode(n, "due/date-time/popup").Dismiss
	pickerNode(n, DateTimeCalendarID(p.ID)).Activate()
	n, _ = p.View()
	pickerNode(n, DateTimeCalendarID(p.ID)).Activate()
	n, _ = p.View()
	oldDay.Activate()
	oldDismiss()
	if calls != 0 || !p.calendarOpen {
		t.Fatal("obsolete popup session acted", calls, p.calendarOpen)
	}
	pickerNode(n, "due/date-time/cell/17").Activate()
	if calls != 1 || p.calendarOpen {
		t.Fatal("current day failed", calls)
	}
}

func TestCalendarSkippedCivilDayHasUniqueDisabledID(t *testing.T) {
	loc, err := time.LoadLocation("Pacific/Apia")
	if err != nil {
		t.Fatal(err)
	}
	c := &Calendar{ID: "apia", Value: time.Date(2011, 12, 29, 12, 0, 0, 0, loc), Change: func(CalendarEvent) {}}
	n, err := c.View()
	if err != nil {
		t.Fatal(err)
	}
	skipped := pickerNode(n, "apia/calendar/day/2011-12-30")
	if skipped.ID == "" || !skipped.Disabled {
		t.Fatal("nonexistent civil day admitted")
	}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
}

func TestDateCalendarSourceLocationRevokesDetachedView(t *testing.T) {
	p := &DateTimePicker{ID: "due", Value: time.Date(2026, 3, 16, 12, 30, 0, 0, time.UTC)}
	calls := 0
	p.Change = func(DateTimeEvent) { calls++ }
	n, _ := p.View()
	pickerNode(n, DateTimeCalendarID(p.ID)).Activate()
	n, _ = p.View()
	oldDay := pickerNode(n, "due/date-time/cell/17")
	p.Value = p.Value.In(time.FixedZone("other", 3600))
	oldDay.Activate()
	pickerNode(n, "due/date-time/next").Activate()
	if calls != 0 || !p.month.IsZero() {
		t.Fatal("obsolete source location acted")
	}
}
