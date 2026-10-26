package main

import (
	"github.com/dshills/vellum/ui"
	"time"
)

func (m *macComponentsGallery) dateRangeView() ui.Node {
	if m.dateRange == nil {
		today := time.Date(2026, 10, 26, 12, 0, 0, 0, time.Local)
		p := &ui.DateRangePicker{ID: "date-range/dates", Label: "Project dates", Value: ui.DateRange{Start: today, End: today.AddDate(0, 0, 6)}, Presets: ui.StandardDateRangePresets(today), FirstWeekday: time.Monday, IconImage: m.artwork[ui.SymbolDocument]}
		p.Change = func(e ui.DateRangeEvent) {
			p.Value = e.Value
			m.status = e.Value.Start.Format("Jan 2") + " – " + e.Value.End.Format("Jan 2")
			m.count++
		}
		p.Invalid = func(e error) { m.status = e.Error() }
		m.dateRange = p
	}
	m.dateRange.Dark = m.dark
	n, e := m.dateRange.View()
	if e != nil {
		return ui.Node{ID: "date-range/error", Kind: ui.Label, Text: e.Error()}
	}
	return n
}
