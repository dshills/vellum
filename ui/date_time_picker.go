package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"strings"
	"time"
)

type DateTimeMode string

const (
	DateOnly    DateTimeMode = "date"
	TimeOnly    DateTimeMode = "time"
	DateAndTime DateTimeMode = "date-time"
)

type DateTimeEvent struct {
	ID    string
	Value time.Time
}

// DateTimePicker is a controlled compact field plus optional calendar popover.
// Retain on one owner. Value's location is retained; Min/Max are inclusive.
// Layout selects the caller's locale format (Go time layout); default ISO date,
// 24-hour time. Calendar labels/weekday order are English/Sunday-first initially.
// Invalid reports invalid syntax, normalized/nonexistent local times or bounds.
// View is pure; private calendar navigation changes only on owner callbacks.
type DateTimePicker struct {
	ID, Label           string
	Value               time.Time
	Mode                DateTimeMode
	Layout              string
	Min, Max            *time.Time
	Disabled            bool
	Change              func(DateTimeEvent)
	Invalid             func(error)
	Style, PopoverStyle Style
	Today               time.Time
	FirstWeekday        time.Weekday
	AllowDate           func(time.Time) bool
	CalendarIconImage   *canvas.TiledImage
	calendarEpoch       *dateCalendarEpoch
	calendarOpen        bool
	month               time.Time
}

type dateCalendarEpoch struct{ marker byte }

func DateTimeInputID(root string) string    { return root + "/date-time/input" }
func DateTimeCalendarID(root string) string { return root + "/date-time/calendar-button" }
func (p *DateTimePicker) mode() DateTimeMode {
	if p.Mode == "" {
		return DateOnly
	}
	return p.Mode
}
func (p *DateTimePicker) layout() string {
	if p.Layout != "" {
		return p.Layout
	}
	switch p.mode() {
	case TimeOnly:
		return "15:04"
	case DateAndTime:
		return "2006-01-02 15:04"
	}
	return "2006-01-02"
}
func (p *DateTimePicker) valid(v time.Time) bool {
	return !v.IsZero() && v.Year() >= 1 && v.Year() <= 9999 && (p.Min == nil || !v.Before(*p.Min)) && (p.Max == nil || !v.After(*p.Max))
}
func (p *DateTimePicker) localDate(year int, month time.Month, day int) (time.Time, error) {
	v := time.Date(year, month, day, p.Value.Hour(), p.Value.Minute(), p.Value.Second(), p.Value.Nanosecond(), p.Value.Location())
	if v.Year() != year || v.Month() != month || v.Day() != day || v.Hour() != p.Value.Hour() || v.Minute() != p.Value.Minute() {
		return time.Time{}, fmt.Errorf("ui: nonexistent local date/time")
	}
	return v, nil
}
func (p *DateTimePicker) parse(text string) (time.Time, error) {
	if len(text) > 256 || strings.ContainsRune(text, 0) {
		return time.Time{}, fmt.Errorf("ui: invalid date/time text")
	}
	// Parse civil fields in UTC first so a local DST gap cannot normalize the
	// requested hour before we compare it. Accept normal parser variants (month
	// case, optional padding) rather than demanding identical formatted text.
	wall, err := time.ParseInLocation(p.layout(), strings.TrimSpace(text), time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	var v time.Time
	switch p.mode() {
	case DateOnly:
		v, err = p.localDate(wall.Year(), wall.Month(), wall.Day())
	case TimeOnly:
		v = time.Date(p.Value.Year(), p.Value.Month(), p.Value.Day(), wall.Hour(), wall.Minute(), wall.Second(), wall.Nanosecond(), p.Value.Location())
		if v.Hour() != wall.Hour() || v.Minute() != wall.Minute() || v.Second() != wall.Second() {
			err = fmt.Errorf("ui: nonexistent local time")
		}
	default:
		v = time.Date(wall.Year(), wall.Month(), wall.Day(), wall.Hour(), wall.Minute(), wall.Second(), wall.Nanosecond(), p.Value.Location())
		if v.Year() != wall.Year() || v.Month() != wall.Month() || v.Day() != wall.Day() || v.Hour() != wall.Hour() || v.Minute() != wall.Minute() || v.Second() != wall.Second() {
			err = fmt.Errorf("ui: nonexistent local date/time")
		}
	}
	if err == nil && (!p.valid(v) || p.mode() != TimeOnly && p.AllowDate != nil && !p.AllowDate(v)) {
		err = fmt.Errorf("ui: date/time is outside inclusive bounds")
	}
	return v, err
}
func (p *DateTimePicker) View() (Node, error) {
	if p == nil || p.ID == "" || !p.valid(p.Value) || len(p.Layout) > 128 || (p.Min != nil && p.Max != nil && p.Min.After(*p.Max)) {
		return Node{}, fmt.Errorf("ui: date/time picker requires ID, valid value and ordered bounds")
	}
	if strings.Contains(p.Layout, "MST") || strings.Contains(p.Layout, "Z07") || strings.Contains(p.Layout, "-07") {
		return Node{}, fmt.Errorf("ui: layout must contain local date/time fields without a time zone")
	}
	switch p.mode() {
	case DateOnly, TimeOnly, DateAndTime:
	default:
		return Node{}, fmt.Errorf("ui: invalid date/time mode")
	}
	probe := time.Date(2037, 11, 23, 17, 46, 52, 0, time.UTC)
	parsed, parseErr := time.ParseInLocation(p.layout(), probe.Format(p.layout()), time.UTC)
	dateFields := parseErr == nil && parsed.Year() == 2037 && parsed.Month() == 11 && parsed.Day() == 23
	timeFields := parseErr == nil && parsed.Hour() == 17 && parsed.Minute() == 46
	if (p.mode() != TimeOnly && !dateFields) || (p.mode() != DateOnly && !timeFields) || (p.mode() == DateOnly && (parsed.Hour() != 0 || parsed.Minute() != 0 || parsed.Second() != 0)) || (p.mode() == TimeOnly && (parsed.Year() != 0 || parsed.Month() != 1 || parsed.Day() != 1)) {
		return Node{}, fmt.Errorf("ui: layout must represent all fields for its date/time mode")
	}
	if p.FirstWeekday < time.Sunday || p.FirstWeekday > time.Saturday {
		return Node{}, fmt.Errorf("ui: invalid first weekday")
	}
	id := p.ID
	mode := p.mode()
	layout := p.layout()
	location := p.Value.Location()
	label := p.Label
	if strings.TrimSpace(label) == "" {
		label = "Date and time"
	}
	available := func() bool {
		return p.ID == id && !p.Disabled && p.Change != nil && p.mode() == mode && p.layout() == layout && p.Value.Location() == location
	}
	report := func(err error) {
		if p.Invalid != nil {
			p.Invalid(err)
		}
	}
	root := Node{ID: id, Kind: Column, Disabled: !available(), Style: Style{Gap: Some(float32(4))}}
	row := Node{ID: id + "/date-time/row", Kind: Row, Children: []Node{{ID: DateTimeInputID(id), Kind: TextField, Text: p.Value.Format(p.layout()), Style: p.Style, Accessibility: Accessibility{Name: label, Help: "Enter " + p.layout()}, Change: func(e ChangeEvent) {
		if !available() {
			return
		}
		v, err := p.parse(e.Text)
		if err != nil {
			report(err)
			return
		}
		p.Change(DateTimeEvent{ID: id, Value: v})
	}}}}
	if mode != TimeOnly {
		row.Children = append(row.Children, Node{ID: DateTimeCalendarID(id), Kind: Button, Text: "Choose date…", IconImage: p.CalendarIconImage, Accessibility: Accessibility{Name: label + " calendar"}, Activate: func() {
			if available() {
				p.calendarOpen = !p.calendarOpen
				if p.calendarOpen {
					p.calendarEpoch = &dateCalendarEpoch{}
				} else {
					p.calendarEpoch = nil
				}
				p.month = time.Time{}
			}
		}})
		month := p.month
		if month.IsZero() {
			month = time.Date(p.Value.Year(), p.Value.Month(), 1, 12, 0, 0, 0, p.Value.Location())
		}

		if !available() {
			p.calendarOpen = false
			p.calendarEpoch = nil
			p.month = time.Time{}
		}
		if p.calendarOpen && p.calendarEpoch == nil {
			p.calendarEpoch = &dateCalendarEpoch{}
		}
		epoch := p.calendarEpoch
		sameMonth := func() bool {
			current := p.month
			if current.IsZero() {
				current = time.Date(p.Value.Year(), p.Value.Month(), 1, 12, 0, 0, 0, p.Value.Location())
			}
			return current.Equal(month)
		}
		liveCalendar := func() bool {
			return available() && p.calendarOpen && epoch != nil && p.calendarEpoch == epoch && sameMonth()
		}
		calendar := &Calendar{ID: id + "/date-time/month-view", Value: p.Value, Month: month, Today: p.Today, FirstWeekday: p.FirstWeekday, Min: p.Min, Max: p.Max, HeaderIconImage: p.CalendarIconImage, AllowDate: p.AllowDate, Disabled: !available(), Navigate: func(e CalendarMonthEvent) {
			if liveCalendar() {
				p.month = e.Month
			}
		}, Change: func(e CalendarEvent) {
			if !liveCalendar() {
				return
			}
			v, err := p.localDate(e.Value.Year(), e.Value.Month(), e.Value.Day())
			if err != nil || !p.valid(v) || p.AllowDate != nil && !p.AllowDate(v) {
				return
			}
			p.Change(DateTimeEvent{ID: id, Value: v})
			p.calendarOpen = false
			p.calendarEpoch = nil
			p.month = time.Time{}
		}}
		calendarNode, err := calendar.view(id+"/date-time", true)
		if err != nil {
			return Node{}, err
		}
		content := calendarNode.Children

		panel, err := (Popover{ID: id + "/date-time/popup", AnchorID: DateTimeCalendarID(id), Label: label, Open: p.calendarOpen, Disabled: !available(), Dismiss: func() {
			if liveCalendar() {
				p.calendarOpen = false
				p.calendarEpoch = nil
				p.month = time.Time{}
			}
		}, Style: p.PopoverStyle, Children: content}).View()
		if err != nil {
			return Node{}, err
		}
		root.Children = []Node{row, panel}
	} else {
		root.Children = []Node{row}
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
