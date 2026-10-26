package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"strings"
	"time"
)

type DateRange struct{ Start, End time.Time }
type DateRangeEvent struct {
	ID    string
	Value DateRange
}
type DateRangePreset struct {
	ID, Label string
	Value     DateRange
}

// DateRangePicker is a controlled inclusive civil-date range, in Start's exact
// location. Time-of-day is ignored; proposals use local noon. It owns only the
// incomplete calendar anchor/month; the app owns accepted data/history/clock.
// Click a start then an end (either direction). Cancel selection discards the
// draft. Native fields propose on commit and reject reversed endpoints.
type DateRangePicker struct {
	ID, Label      string
	Value          DateRange
	Min, Max       *time.Time
	Today          time.Time
	FirstWeekday   time.Weekday
	Presets        []DateRangePreset
	Disabled, Dark bool
	IconImage      *canvas.TiledImage
	Change         func(DateRangeEvent)
	Invalid        func(error)
	Style          Style
	anchor         *time.Time
	month          time.Time
	accepted       DateRange
}

func DateRangeStartID(id string) string { return id + "/range/start" }
func DateRangeEndID(id string) string   { return id + "/range/end" }
func civilDate(v time.Time) time.Time {
	return time.Date(v.Year(), v.Month(), v.Day(), 12, 0, 0, 0, v.Location())
}
func civilKey(v time.Time) int    { return v.Year()*10000 + int(v.Month())*100 + v.Day() }
func validCivil(v time.Time) bool { return !v.IsZero() && v.Year() >= 1 && v.Year() <= 9999 }
func (p *DateRangePicker) validateValue(v DateRange) error {
	if !validCivil(v.Start) || !validCivil(v.End) || v.Start.Location() != v.End.Location() || civilKey(v.Start) > civilKey(v.End) {
		return fmt.Errorf("ui: range requires ordered civil dates in one location")
	}
	loc := v.Start.Location()
	if p.Min != nil && (!validCivil(*p.Min) || civilKey(v.Start) < civilKey(p.Min.In(loc))) {
		return fmt.Errorf("ui: range is before minimum")
	}
	if p.Max != nil && (!validCivil(*p.Max) || civilKey(v.End) > civilKey(p.Max.In(loc))) {
		return fmt.Errorf("ui: range is after maximum")
	}
	return nil
}

// StandardDateRangePresets uses an explicit caller clock; no time.Now call.
// Values outside years1..9999 are omitted.
func StandardDateRangePresets(today time.Time) []DateRangePreset {
	if !validCivil(today) {
		return nil
	}
	today = civilDate(today)
	out := []DateRangePreset{{ID: "today", Label: "Today", Value: DateRange{today, today}}}
	start := today.AddDate(0, 0, -6)
	if validCivil(start) {
		out = append(out, DateRangePreset{ID: "last-seven", Label: "Last 7 days", Value: DateRange{start, today}})
	}
	start = time.Date(today.Year(), today.Month(), 1, 12, 0, 0, 0, today.Location())
	return append(out, DateRangePreset{ID: "month-to-date", Label: "Month to date", Value: DateRange{start, today}})
}
func (p *DateRangePicker) View() (Node, error) {
	if p == nil || !validTokenText(p.ID) || len(p.Presets) > 12 || p.FirstWeekday < 0 || p.FirstWeekday > 6 {
		return Node{}, fmt.Errorf("ui: invalid date range picker")
	}
	if e := p.validateValue(p.Value); e != nil {
		return Node{}, e
	}
	if p.Min != nil && p.Max != nil && civilKey(p.Min.In(p.Value.Start.Location())) > civilKey(p.Max.In(p.Value.Start.Location())) {
		return Node{}, fmt.Errorf("ui: reversed range bounds")
	}
	seen := map[string]bool{}
	for _, preset := range p.Presets {
		if !validTokenText(preset.ID) || !validTokenText(preset.Label) || seen[preset.ID] || !validCivil(preset.Value.Start) || !validCivil(preset.Value.End) || preset.Value.Start.Location() != p.Value.Start.Location() || preset.Value.End.Location() != p.Value.Start.Location() || civilKey(preset.Value.Start) > civilKey(preset.Value.End) {
			return Node{}, fmt.Errorf("ui: invalid date range preset")
		}
		seen[preset.ID] = true
	}
	if civilKey(p.accepted.Start) != civilKey(p.Value.Start) || civilKey(p.accepted.End) != civilKey(p.Value.End) || p.accepted.Start.Location() != p.Value.Start.Location() {
		p.anchor = nil
		p.accepted = p.Value
	}
	id, loc := p.ID, p.Value.Start.Location()
	renderedStart, renderedEnd := civilKey(p.Value.Start), civilKey(p.Value.End)
	renderedMonth := p.month
	live := func() bool {
		return p.ID == id && !p.Disabled && p.Change != nil && p.Value.Start.Location() == loc && civilKey(p.Value.Start) == renderedStart && civilKey(p.Value.End) == renderedEnd
	}
	propose := func(v DateRange) {
		if !live() {
			return
		}
		v.Start = civilDate(v.Start)
		v.End = civilDate(v.End)
		if e := p.validateValue(v); e != nil {
			if p.Invalid != nil {
				p.Invalid(e)
			}
			return
		}
		p.Change(DateRangeEvent{ID: id, Value: v})
	}
	fields := []Node{}
	for _, start := range []bool{true, false} {
		value := p.Value.End
		fieldID, label := DateRangeEndID(id), "End date"
		if start {
			value = p.Value.Start
			fieldID = DateRangeStartID(id)
			label = "Start date"
		}
		fields = append(fields, Node{ID: fieldID, Kind: TextField, Text: value.Format("2006-01-02"), Accessibility: Accessibility{Name: label, Help: "YYYY-MM-DD"}, Change: func(e ChangeEvent) {
			if !live() {
				return
			}
			base := p.Value.End
			if start {
				base = p.Value.Start
			}
			engine := DateTimePicker{Value: civilDate(base), Mode: DateOnly}
			candidate, err := engine.parse(e.Text)
			if err != nil {
				if p.Invalid != nil {
					p.Invalid(err)
				}
				return
			}
			v := p.Value
			if start {
				v.Start = candidate
			} else {
				v.End = candidate
			}
			propose(v)
		}})
	}
	selected := p.Value.Start
	if p.anchor != nil {
		selected = *p.anchor
	}
	minimum, maximum := (*time.Time)(nil), (*time.Time)(nil)
	if p.Min != nil {
		v := civilDate(p.Min.In(loc))
		minimum = &v
	}
	if p.Max != nil {
		v := civilDate(p.Max.In(loc))
		maximum = &v
	}
	calendarLive := func() bool { return live() && civilKey(p.month) == civilKey(renderedMonth) }
	c := Calendar{ID: id + "/range/calendar", Value: civilDate(selected), Month: p.month, Min: minimum, Max: maximum, FirstWeekday: p.FirstWeekday, HeaderIconImage: p.IconImage, Change: func(e CalendarEvent) {
		if !calendarLive() {
			return
		}
		v := civilDate(e.Value)
		if p.anchor == nil {
			p.anchor = &v
			return
		}
		rangeValue := DateRange{Start: *p.anchor, End: v}
		if civilKey(rangeValue.Start) > civilKey(rangeValue.End) {
			rangeValue.Start, rangeValue.End = rangeValue.End, rangeValue.Start
		}
		p.anchor = nil
		propose(rangeValue)
	}, Navigate: func(e CalendarMonthEvent) {
		if calendarLive() {
			p.month = e.Month
		}
	}}
	calendar, err := c.View()
	if err != nil {
		return Node{}, err
	}
	palette := MacPalette(p.Dark, color.NRGBA{})
	lo, hi := civilKey(p.Value.Start), civilKey(p.Value.End)
	if p.anchor != nil {
		lo = civilKey(*p.anchor)
		hi = lo
	}
	var decorate func(*Node)
	decorate = func(n *Node) {
		if n.CalendarCell != nil && n.Kind == Button && len(n.ID) >= 10 && strings.HasPrefix(n.ID, c.ID+"/calendar/day/") {
			day, err := time.Parse("2006-01-02", n.ID[len(n.ID)-10:])
			if err == nil {
				key := civilKey(day)
				meta := *n.CalendarCell
				meta.Selected = key == lo || key == hi
				n.CalendarCell = &meta
				if key > lo && key < hi {
					mix := func(a, b uint8) uint8 { return uint8((uint16(a)*3 + uint16(b)) / 4) }
					n.Style.Background = Some(color.NRGBA{mix(palette.Field.R, palette.Accent.R), mix(palette.Field.G, palette.Accent.G), mix(palette.Field.B, palette.Accent.B), 255})
					n.Accessibility.Help += "; within selected range"
				}
			}
		}
		for i := range n.Children {
			decorate(&n.Children[i])
		}
	}
	decorate(&calendar)
	root := Node{ID: id, Kind: Column, Disabled: p.Disabled || p.Change == nil, Style: p.Style, Children: []Node{{ID: id + "/range/label", Kind: Label, Text: p.Label}, {ID: id + "/range/fields", Kind: Row, Children: fields}, calendar}}
	if len(p.Presets) > 0 {
		options := []Option{}
		for _, preset := range p.Presets {
			options = append(options, Option{ID: preset.ID, Label: preset.Label, Disabled: p.validateValue(preset.Value) != nil})
		}
		root.Children = append(root.Children, Node{ID: id + "/range/presets", Kind: MenuButton, Text: "Presets", Options: options, Change: func(e ChangeEvent) {
			if !live() {
				return
			}
			for _, preset := range p.Presets {
				if preset.ID == e.SelectedID {
					p.anchor = nil
					propose(preset.Value)
					return
				}
			}
		}})
	}
	caption := "Choose a start date, then an end date"
	if p.anchor != nil {
		caption = "Start: " + p.anchor.Format("Jan 2, 2006") + " — choose an end date"
		root.Children = append(root.Children, Node{ID: id + "/range/cancel", Kind: Button, Text: "Cancel selection", Activate: func() {
			if live() {
				p.anchor = nil
			}
		}})
	}
	root.Children = append(root.Children, Node{ID: id + "/range/help", Kind: Label, Text: caption})
	if err = Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
