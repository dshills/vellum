package ui

import (
	"testing"
	"time"
)

func TestDateRangeCivilDraftReversalAndPresets(t *testing.T) {
	loc, e := time.LoadLocation("America/Chicago")
	if e != nil {
		t.Fatal(e)
	}
	date := func(d int) time.Time { return time.Date(2026, 3, d, 12, 0, 0, 0, loc) }
	calls, invalid := 0, 0
	p := &DateRangePicker{ID: "range", Value: DateRange{date(7), date(9)}, Presets: StandardDateRangePresets(date(9)), Change: func(e DateRangeEvent) {
		calls++
		if e.Value.Start.Location() != loc || e.Value.Start.Hour() != 12 {
			t.Fatal(e)
		}
	}, Invalid: func(error) { invalid++ }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	paletteNode(n, DateRangeStartID(p.ID)).Change(ChangeEvent{Text: "2026-03-10"})
	if calls != 0 || invalid != 1 {
		t.Fatal("reversed field range accepted")
	}
	first := paletteNode(n, CalendarDayID(p.ID+"/range/calendar", date(12)))
	first.Activate()
	if calls != 0 {
		t.Fatal("draft committed early")
	}
	n, e = p.View()
	if e != nil {
		t.Fatal(e)
	}
	paletteNode(n, CalendarDayID(p.ID+"/range/calendar", date(10))).Activate()
	if calls != 1 {
		t.Fatal("reverse click range not proposed")
	}
	if len(p.Presets) != 3 || civilKey(p.Presets[1].Value.Start) != 20260303 {
		t.Fatal("civil DST preset", p.Presets)
	}
	p.Disabled = true
	first.Activate()
	if calls != 1 {
		t.Fatal("disabled draft accepted")
	}
}
func TestDateRangeBoundsAndHighlight(t *testing.T) {
	d := func(n int) time.Time { return time.Date(2026, 10, n, 1, 0, 0, 0, time.UTC) }
	min, max := d(5), d(20)
	p := &DateRangePicker{ID: "r", Value: DateRange{d(8), d(12)}, Min: &min, Max: &max, Change: func(DateRangeEvent) {}}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	inside := paletteNode(n, CalendarDayID("r/range/calendar", d(10)))
	if !inside.Style.Background.Set || inside.CalendarCell.Selected {
		t.Fatal("interior highlight")
	}
	if !paletteNode(n, CalendarDayID("r/range/calendar", d(8))).CalendarCell.Selected {
		t.Fatal("endpoint highlight")
	}
	p.Value.End = d(21)
	if _, e = p.View(); e == nil {
		t.Fatal("out of bounds range")
	}
}

func TestDateRangeCivilRepresentationAndStaleCallbacks(t *testing.T) {
	d := func(n int) time.Time { return time.Date(2026, 10, n, 1, 0, 0, 0, time.UTC) }
	calls := 0
	p := &DateRangePicker{ID: "stable", Value: DateRange{d(2), d(4)}, Change: func(DateRangeEvent) { calls++ }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	old := paletteNode(n, CalendarDayID(p.ID+"/range/calendar", d(6)))
	old.Activate()
	if p.anchor == nil {
		t.Fatal("missing anchor")
	}
	p.Value.Start = p.Value.Start.Add(time.Hour)
	p.Value.End = p.Value.End.Add(time.Hour)
	if _, e = p.View(); e != nil || p.anchor == nil {
		t.Fatal("same civil range reset draft", e)
	}
	p.Value = DateRange{d(10), d(12)}
	old.Activate()
	if calls != 0 {
		t.Fatal("stale accepted-range callback")
	}
}
