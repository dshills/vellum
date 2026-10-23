package ui

import (
	"strings"
	"testing"
	"time"
)

func TestDateTimePickerControlledParseBoundsAndCalendar(t *testing.T) {
	value := time.Date(2024, 2, 28, 13, 45, 12, 0, time.UTC)
	min := value.AddDate(0, 0, -1)
	max := value.AddDate(0, 0, 2)
	p := &DateTimePicker{ID: "due", Value: value, Min: &min, Max: &max}
	calls, invalid := 0, 0
	p.Change = func(e DateTimeEvent) { calls++ }
	p.Invalid = func(error) { invalid++ }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, DateTimeInputID(p.ID))
	field.Change(ChangeEvent{Text: "2024-02-29"})
	if calls != 1 || p.Value != value {
		t.Fatal("controlled rejection")
	}
	field.Change(ChangeEvent{Text: "2024-02-30"})
	field.Change(ChangeEvent{Text: "2024-03-03"})
	if invalid != 2 {
		t.Fatal("invalid/bounds", invalid)
	}
	p.Change = func(e DateTimeEvent) { p.Value = e.Value }
	field.Change(ChangeEvent{Text: "2024-02-29"})
	if p.Value.Day() != 29 || p.Value.Hour() != 13 || p.Value.Second() != 12 {
		t.Fatal("date did not preserve time", p.Value)
	}
	helperNode(t, n, DateTimeCalendarID(p.ID)).Activate()
	n, err = p.View()
	if err != nil {
		t.Fatal(err)
	}
	if !helperNode(t, n, "due/date-time/popup").Open {
		t.Fatal("calendar not open")
	}
	var days []Node
	var walk func(Node)
	walk = func(n Node) {
		if strings.HasPrefix(n.ID, "due/date-time/cell/") && n.Kind == Button {
			days = append(days, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	if len(days) != 29 {
		t.Fatal("leap month", len(days))
	}
	for _, day := range days {
		if day.Text == "29" {
			day.Activate()
		}
	}
	if p.calendarOpen {
		t.Fatal("calendar placement did not close")
	}
	p.ID = "new"
	field.Change(ChangeEvent{Text: "2024-02-28"})
	if p.Value.Day() != 29 {
		t.Fatal("stale component callback")
	}
}
func TestDateTimePickerTimeLayoutAndDST(t *testing.T) {
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	p := &DateTimePicker{ID: "clock", Mode: TimeOnly, Layout: "3:04 PM", Value: time.Date(2026, 10, 23, 9, 0, 0, 0, loc)}
	p.Change = func(e DateTimeEvent) { p.Value = e.Value }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	helperNode(t, n, DateTimeInputID(p.ID)).Change(ChangeEvent{Text: "2:30 PM"})
	if p.Value.Hour() != 14 || p.Value.Day() != 23 || p.Value.Location() != loc {
		t.Fatal(p.Value)
	}
	if _, ok := helperFind(n, DateTimeCalendarID(p.ID)); ok {
		t.Fatal("time-only calendar")
	}
	p.Mode = DateAndTime
	p.Layout = ""
	p.Value = time.Date(2026, 3, 8, 1, 30, 0, 0, loc)
	if _, err := p.parse("2026-03-08 02:30"); err == nil {
		t.Fatal("DST gap admitted")
	}
	p.Mode = DateOnly
	p.Value = time.Date(2026, 3, 7, 2, 30, 0, 0, loc)
	if _, err := p.parse("2026-03-08"); err == nil {
		t.Fatal("preserved time normalized across gap")
	}
	p.Disabled = true
	n, err = p.View()
	if err != nil || !n.Disabled {
		t.Fatal(n, err)
	}
	p.Mode = "bad"
	if _, err := p.View(); err == nil {
		t.Fatal("unknown mode")
	}
}

func TestDateTimePickerAcceptsParserVariantsAndRevokesOldLayout(t *testing.T) {
	p := &DateTimePicker{ID: "due", Value: time.Date(2024, 2, 28, 9, 0, 0, 0, time.UTC), Layout: "Jan 2, 2006"}
	p.Change = func(e DateTimeEvent) { p.Value = e.Value }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, DateTimeInputID(p.ID))
	field.Change(ChangeEvent{Text: "feb 29, 2024"})
	if p.Value.Day() != 29 {
		t.Fatal("case variant refused", p.Value)
	}
	p.Layout = "2006-01-02"
	field.Change(ChangeEvent{Text: "2024-02-28"})
	if p.Value.Day() != 29 {
		t.Fatal("obsolete layout callback admitted")
	}
}

func TestDateTimePickerRejectsLayoutModeMismatch(t *testing.T) {
	for _, tc := range []struct {
		mode   DateTimeMode
		layout string
	}{{DateOnly, "Jan 2"}, {DateAndTime, "15:04"}, {TimeOnly, "2006-01-02"}, {DateOnly, "2006-01-02 15:04"}, {TimeOnly, "2006-01-02 15:04"}, {DateAndTime, "2006-01-02"}} {
		p := &DateTimePicker{ID: "date", Value: time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC), Mode: tc.mode, Layout: tc.layout}
		if _, err := p.View(); err == nil {
			t.Fatal("mode/layout mismatch", tc)
		}
	}
}
