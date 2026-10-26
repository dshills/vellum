package sdkcontract_test

import (
	"fmt"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"time"
)

func contentNode(n ui.Node, id string) ui.Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := contentNode(c, id); found.ID != "" {
			return found
		}
	}
	return ui.Node{}
}

// This is application-owned history. The component proposes values; it never
// owns an application's journal or mutates accepted data itself.
func ExampleChecklist_history() {
	journal, _ := history.New(16)
	item := ui.ChecklistItem{ID: "review", Text: "Review drawing"}
	checklist := ui.Checklist{ID: "tasks", Items: []ui.ChecklistItem{item}}
	checklist.Change = func(e ui.ChecklistEvent) {
		before := item
		after := item
		after.Completed = e.Checked
		if before == after {
			return
		}
		install := func(v ui.ChecklistItem) error { item = v; checklist.Items = []ui.ChecklistItem{v}; return nil }
		if err := journal.Apply(history.Change{Label: "Complete task", Apply: func() error { return install(after) }, Revert: func() error { return install(before) }}); err != nil {
			panic(err)
		}
	}
	n, _ := checklist.View()
	contentNode(n, ui.ChecklistCheckID("tasks", "review")).Change(ui.ChangeEvent{Checked: true})
	fmt.Println(item.ID, item.Completed)
	journal.Undo()
	fmt.Println(item.ID, item.Completed)
	journal.Redo()
	fmt.Println(item.ID, item.Completed)
	// Output:
	// review true
	// review false
	// review true
}
func ExampleDateRangePicker() {
	today := time.Date(2026, time.October, 26, 12, 0, 0, 0, time.UTC)
	rangePicker := ui.DateRangePicker{ID: "dates", Label: "Schedule", Value: ui.DateRange{Start: today, End: today.AddDate(0, 0, 4)}}
	rangePicker.Change = func(e ui.DateRangeEvent) { rangePicker.Value = e.Value }
	node, _ := rangePicker.View()
	contentNode(node, ui.DateRangeEndID("dates")).Change(ui.ChangeEvent{Text: "2026-11-02"})
	fmt.Println(rangePicker.Value.Start.Format("2006-01-02"), rangePicker.Value.End.Format("2006-01-02"))
	// Output: 2026-10-26 2026-11-02
}
