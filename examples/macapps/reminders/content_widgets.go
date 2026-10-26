package reminders

import (
	"github.com/dshills/vellum/examples/macapps/artwork"
	"github.com/dshills/vellum/ui"
	"image/color"
	"time"
	"unicode/utf8"
)

func coloredAction(n ui.Node, symbol ui.Symbol) ui.Node {
	image, e := artwork.ToolbarIcon(symbol)
	if e != nil {
		panic(e)
	}
	n.Icon = ""
	n.IconImage = image
	return n
}
func checklistID(id string) string { return "reminder-checklist/" + id }
func checklistTitle(s string) string {
	if len(s) <= 256 {
		return s
	}
	for len(s) > 253 {
		_, n := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-n]
	}
	return s + "…"
}
func (m *Model) checklistFor(r Reminder) *ui.Checklist {
	if m.checklists == nil {
		m.checklists = map[string]*ui.Checklist{}
	}
	c := m.checklists[r.ID]
	if c == nil {
		c = &ui.Checklist{ID: checklistID(r.ID), Style: ui.Style{Background: ui.Some(color.NRGBA{}), Padding: ui.Some(ui.Insets{}), GrowX: ui.Some(float32(1))}}
		c.Change = func(e ui.ChecklistEvent) {
			if e.Action != ui.ChecklistToggle || !m.inputAlive() {
				return
			}
			_ = m.AfterInput(func() error { return m.SetCompleted(e.ItemID, e.Checked) })
		}
		m.checklists[r.ID] = c
	}
	priority := map[string]uint8{"low": 1, "medium": 2, "high": 3}[r.Priority]
	c.Disabled = !m.inputAlive() || m.PendingInput()
	c.Items = []ui.ChecklistItem{{ID: r.ID, Text: checklistTitle(r.Title), Completed: r.Completed, Priority: priority}}
	return c
}
func setChecklistName(n *ui.Node, id, name string) {
	if n.ID == id {
		n.Accessibility.Name = name
		n.Style.Typography = ui.Some(ui.MacTypography(14, 400))
	}
	for i := range n.Children {
		setChecklistName(&n.Children[i], id, name)
	}
}
func (m *Model) inScheduleRange(r Reminder) bool {
	if m.selected != "scheduled" || !m.rangeActive || m.schedule == nil {
		return true
	}
	return r.Due >= m.schedule.Value.Start.Format("2006-01-02") && r.Due <= m.schedule.Value.End.Format("2006-01-02")
}
func (m *Model) scheduleView() ui.Node {
	if m.selected != "scheduled" {
		return ui.Node{ID: "reminders-schedule-empty", Kind: ui.Label}
	}
	if m.schedule == nil {
		today, _ := time.Parse("2006-01-02", Today)
		p := &ui.DateRangePicker{ID: "reminders-schedule-dates", Label: "Scheduled dates", Value: ui.DateRange{Start: today, End: today.AddDate(0, 0, 7)}, FirstWeekday: time.Monday, Presets: ui.StandardDateRangePresets(today)}
		p.Change = func(e ui.DateRangeEvent) {
			if m.inputAlive() && m.selected == "scheduled" {
				p.Value = e.Value
				m.rangeActive = true
			}
		}
		p.Invalid = m.inputError
		m.schedule = p
	}
	// Field Change must stay admitted while native settlement publishes text.
	// Calendar/preset actions have a separate pending-action guard below.
	m.schedule.Disabled = !m.inputAlive()
	n, e := m.schedule.View()
	if e != nil {
		panic(e)
	}
	m.guardScheduleActions(&n)
	clear := button("reminders-clear-range", "All scheduled dates", "Clear date range filter", "", func() { _ = m.AfterInput(func() error { m.rangeActive = false; return nil }) })
	return ui.Node{ID: "reminders-schedule", Kind: ui.Disclosure, Text: func() string {
		if m.rangeActive {
			return "Date range filter (active)"
		}
		return "Date range filter"
	}(), Checked: m.rangeExpanded, Change: func(e ui.ChangeEvent) { _ = m.AfterInput(func() error { m.rangeExpanded = e.Checked; return nil }) }, Children: []ui.Node{n, clear}}
}

func (m *Model) guardScheduleActions(n *ui.Node) {
	if n.Kind != ui.TextField {
		if fn := n.Activate; fn != nil {
			n.Activate = func() {
				if !m.PendingInput() && m.inputAlive() {
					fn()
				}
			}
		}
		if fn := n.Change; fn != nil {
			n.Change = func(e ui.ChangeEvent) {
				if !m.PendingInput() && m.inputAlive() {
					fn(e)
				}
			}
		}
	}
	for i := range n.Children {
		m.guardScheduleActions(&n.Children[i])
	}
}
