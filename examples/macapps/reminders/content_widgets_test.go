package reminders

import (
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
	"time"
)

func TestReusableChecklistSettlementHistoryAndRetirement(t *testing.T) {
	m := New()
	h := installHeld(t, m, t.Context())
	n := find(t, m.View(), ui.ChecklistCheckID(checklistID("r1"), "r1"))
	n.Change(ui.ChangeEvent{Checked: true})
	if !m.PendingInput() || item(t, m, "r1").Completed {
		t.Fatal("bypassed settlement")
	}
	h.settle(t, nil)
	if !item(t, m, "r1").Completed || m.History().UndoCount != 1 {
		t.Fatal("completion missed history")
	}
	if e := m.Undo(); e != nil {
		t.Fatal(e)
	}
	h.settle(t, nil)
	if item(t, m, "r1").Completed {
		t.Fatal("undo")
	}
	if e := m.Redo(); e != nil {
		t.Fatal(e)
	}
	h.settle(t, nil)
	if !item(t, m, "r1").Completed {
		t.Fatal("redo")
	}
	m.Close()
	n.Change(ui.ChangeEvent{Checked: false})
	if !item(t, m, "r1").Completed || len(m.checklists) != 0 {
		t.Fatal("late callback after close")
	}
}
func TestScheduleRangeFilteringDoesNotEditReminderHistory(t *testing.T) {
	m := New()
	m.Select("scheduled")
	m.rangeExpanded = true
	m.View()
	start, _ := time.Parse("2006-01-02", Today)
	m.schedule.Change(ui.DateRangeEvent{Value: ui.DateRange{Start: start, End: start}})
	for _, r := range m.Visible() {
		if r.Due != Today {
			t.Fatal("range filter", r)
		}
	}
	if m.History().UndoCount != 0 {
		t.Fatal("filter edited reminders")
	}
	find(t, m.View(), "reminders-clear-range").Activate()
	if m.rangeActive {
		t.Fatal("filter not cleared")
	}
}
func TestChecklistLongTitlesRetainModelAndFullAccessibleName(t *testing.T) {
	m := New()
	long := strings.Repeat("é", 200)
	id, e := m.Add(long)
	if e != nil {
		t.Fatal(e)
	}
	n := m.View()
	name := find(t, n, ui.ChecklistNameID(checklistID(id), id))
	if name.Accessibility.Name != long || len(name.Text) > 256 || item(t, m, id).Title != long {
		t.Fatal("long title lost")
	}
	if e = ui.Validate(n); e != nil {
		t.Fatal(e)
	}
}

func TestScheduledSettlementAllowsDateFieldsAndRejectsCalendarActions(t *testing.T) {
	m := New()
	m.Select("scheduled")
	m.rangeExpanded = true
	tree := m.View()
	var calendar []ui.Node
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if n.CalendarCell != nil {
			calendar = append(calendar, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(tree)
	if len(calendar) < 2 {
		t.Fatal("calendar missing")
	}
	h := installHeld(t, m, t.Context())
	m.Undo()
	if !m.PendingInput() {
		t.Fatal("settlement not pending")
	}
	calendar[0].Activate()
	calendar[1].Activate()
	if m.rangeActive {
		t.Fatal("calendar changed filter during settlement")
	}
	if e := ui.Validate(m.View()); e != nil {
		t.Fatal("disabled calendar description invalid", e)
	}
	h.settle(t, map[string]string{ui.DateRangeStartID(m.schedule.ID): "2026-10-20", ui.DateRangeEndID(m.schedule.ID): "2026-10-26"})
	if !m.rangeActive || m.schedule.Value.Start.Format("2006-01-02") != "2026-10-20" || m.PendingInput() {
		t.Fatal("settlement lost controlled range commit")
	}
}
