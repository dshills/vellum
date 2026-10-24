package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"strconv"
	"time"
)

type CalendarEvent struct {
	ID    string
	Value time.Time
}
type CalendarMonthEvent struct {
	ID    string
	Month time.Time
}

// CalendarCell is detached immutable focus metadata produced by Calendar.View.
// Empty entries represent disabled/out-of-month cells; never mutate after View.
type CalendarCell struct {
	IDs   [42]string
	Index int
}

func (c CalendarCell) Neighbor(delta int) string {
	if c.Index < 0 || c.Index >= 42 || (delta != -1 && delta != 1 && delta != -7 && delta != 7) {
		return ""
	}
	for i := c.Index + delta; i >= 0 && i < 42; i += delta {
		if c.IDs[i] != "" {
			return c.IDs[i]
		}
	}
	return ""
}

// Calendar is an owner-held controlled month view. Zero Month follows Value;
// zero Today hides Today. English labels; FirstWeekday selects weekday order.
// AllowDate must be pure/nonblocking. Civil dates preserve Value's location and
// time-of-day using DateTimePicker's DST/range validation. Change proposes a date.
type Calendar struct {
	ID                  string
	Value, Month, Today time.Time
	Min, Max            *time.Time
	FirstWeekday        time.Weekday
	AllowDate           func(time.Time) bool
	Disabled            bool
	HeaderIconImage     *canvas.TiledImage
	Change              func(CalendarEvent)
	Navigate            func(CalendarMonthEvent)
	Style               Style
}

func CalendarDayID(id string, value time.Time) string {
	return id + "/calendar/day/" + value.Format("2006-01-02")
}
func (c *Calendar) month() time.Time {
	v := c.Month
	if v.IsZero() {
		v = c.Value
	}
	return time.Date(v.Year(), v.Month(), 1, 12, 0, 0, 0, c.Value.Location())
}
func (c *Calendar) View() (Node, error) {
	if c == nil {
		return Node{}, fmt.Errorf("ui: calendar required")
	}
	return c.view(c.ID+"/calendar", false)
}

// legacy preserves DateTimePicker's published focus IDs.
func (c *Calendar) view(prefix string, legacy bool) (Node, error) {
	if c == nil || !validTokenText(c.ID) || c.Value.IsZero() || c.FirstWeekday < time.Sunday || c.FirstWeekday > time.Saturday {
		return Node{}, fmt.Errorf("ui: calendar requires ID/value and valid weekday")
	}
	engine := DateTimePicker{Value: c.Value, Min: c.Min, Max: c.Max}
	if !engine.valid(c.Value) || c.Min != nil && c.Max != nil && c.Min.After(*c.Max) {
		return Node{}, fmt.Errorf("ui: calendar value/bounds invalid")
	}
	month := c.month()
	if month.Year() < 1 || month.Year() > 9999 {
		return Node{}, fmt.Errorf("ui: calendar month outside 1..9999")
	}
	id, location := c.ID, c.Value.Location()
	live := func() bool {
		return c.ID == id && !c.Disabled && c.Change != nil && c.Value.Location() == location && c.month().Equal(month)
	}
	candidate := func(year int, m time.Month, day int) (time.Time, bool) {
		current := DateTimePicker{Value: c.Value, Min: c.Min, Max: c.Max}
		v, err := current.localDate(year, m, day)
		return v, err == nil && current.valid(v) && (c.AllowDate == nil || c.AllowDate(v))
	}
	change := func(year int, m time.Month, day int) {
		if !live() {
			return
		}
		if v, ok := candidate(year, m, day); ok {
			c.Change(CalendarEvent{ID: id, Value: v})
		}
	}
	dayStyle := Style{MinSize: Some(Size{Width: 28, Height: 26}), MaxSize: Some(Size{Width: 32}), GrowX: Some(float32(0)), Padding: Some(UniformInsets(3)), TextAlign: Some(AlignCenter)}
	nav := Node{ID: prefix + "/month-nav", Kind: Row, Children: []Node{{ID: prefix + "/previous", Kind: Button, Icon: SymbolLeft, Appearance: ButtonTool, Accessibility: Accessibility{Name: "Previous month"}, Disabled: c.Navigate == nil || month.Year() == 1 && month.Month() == time.January, Activate: func() {
		if live() && c.Navigate != nil && (month.Year() > 1 || month.Month() > time.January) {
			c.Navigate(CalendarMonthEvent{ID: id, Month: month.AddDate(0, -1, 0)})
		}
	}}, {ID: prefix + "/month-label", Kind: Label, Text: month.Format("January 2006"), Grow: true}, {ID: prefix + "/next", Kind: Button, Icon: SymbolRight, Appearance: ButtonTool, Accessibility: Accessibility{Name: "Next month"}, Disabled: c.Navigate == nil || month.Year() == 9999 && month.Month() == time.December, Activate: func() {
		if live() && c.Navigate != nil && (month.Year() < 9999 || month.Month() < time.December) {
			c.Navigate(CalendarMonthEvent{ID: id, Month: month.AddDate(0, 1, 0)})
		}
	}}}}
	if c.HeaderIconImage != nil {
		nav.Children = append([]Node{{ID: prefix + "/icon", Kind: Image, Image: c.HeaderIconImage, Text: "Calendar", Fit: ImageContain, Style: Style{MinSize: Some(Size{Width: 20, Height: 20}), MaxSize: Some(Size{Width: 20, Height: 20})}}}, nav.Children...)
	}
	content := []Node{nav}
	weekdays := []Node{}
	names := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i := 0; i < 7; i++ {
		weekdays = append(weekdays, Node{ID: prefix + "/weekday/" + strconv.Itoa(i), Kind: Label, Text: names[(int(c.FirstWeekday)+i)%7], Style: dayStyle})
	}
	content = append(content, Node{ID: prefix + "/weekdays", Kind: Row, Style: Style{Gap: Some(float32(2))}, Children: weekdays})
	last := month.AddDate(0, 1, -1).Day()
	offset := (int(month.Weekday()) - int(c.FirstWeekday) + 7) % 7
	ids := [42]string{}
	for i := 0; i < 42; i++ {
		day := i - offset + 1
		if day < 1 || day > last {
			continue
		}
		if v, ok := candidate(month.Year(), month.Month(), day); ok {
			ids[i] = CalendarDayID(id, v)
			if legacy {
				ids[i] = prefix + "/cell/" + strconv.Itoa(i)
			}
		}
	}
	for week := 0; week < 6; week++ {
		cells := []Node{}
		for col := 0; col < 7; col++ {
			i := week*7 + col
			day := i - offset + 1
			cellID := prefix + "/cell/" + strconv.Itoa(i)
			if day < 1 || day > last {
				cells = append(cells, Node{ID: cellID, Kind: Label, Style: dayStyle})
				continue
			}
			if !legacy {
				cellID = fmt.Sprintf("%s/calendar/day/%04d-%02d-%02d", id, month.Year(), month.Month(), day)
			}
			appearance := ButtonAppearance("")
			if c.Value.Year() == month.Year() && c.Value.Month() == month.Month() && c.Value.Day() == day {
				appearance = ButtonPrimary
			}
			cell := Node{ID: cellID, Kind: Button, Text: strconv.Itoa(day), Disabled: ids[i] == "", Appearance: appearance, Style: dayStyle, Accessibility: Accessibility{Name: month.Format("January") + " " + strconv.Itoa(day) + ", " + strconv.Itoa(month.Year()), Help: "Arrow keys move; Return selects"}, Activate: func() { change(month.Year(), month.Month(), day) }}
			if ids[i] != "" {
				cell.CalendarCell = &CalendarCell{IDs: ids, Index: i}
			}
			cells = append(cells, cell)
		}
		content = append(content, Node{ID: prefix + "/week/" + strconv.Itoa(week), Kind: Row, Style: Style{Gap: Some(float32(2))}, Children: cells})
	}
	if !c.Today.IsZero() {
		today := c.Today.In(location)
		_, allowed := candidate(today.Year(), today.Month(), today.Day())
		content = append(content, Node{ID: prefix + "/today", Kind: Button, Text: "Today", Disabled: !allowed, Activate: func() { change(today.Year(), today.Month(), today.Day()) }})
	}
	n := Node{ID: id, Kind: Column, Disabled: c.Disabled || c.Change == nil, Style: Style{GrowX: Some(float32(0)), Gap: Some(float32(2))}.Overlay(c.Style), Children: content}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
