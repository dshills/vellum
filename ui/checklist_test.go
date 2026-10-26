package ui

import "testing"

func TestChecklistControlledProposalsAndStaleIdentity(t *testing.T) {
	events := []ChecklistEvent{}
	invalid := 0
	c := &Checklist{ID: "tasks", Editable: true, Items: []ChecklistItem{{ID: "a", Text: "First", Due: "Today", Priority: 2}, {ID: "b", Text: "Second", Disabled: true}}, Change: func(e ChecklistEvent) { events = append(events, e) }, Invalid: func(error) { invalid++ }}
	n, e := c.View()
	if e != nil {
		t.Fatal(e)
	}
	toggle := paletteNode(n, ChecklistCheckID(c.ID, "a"))
	name := paletteNode(n, ChecklistNameID(c.ID, "a"))
	toggle.Change(ChangeEvent{Checked: true})
	name.Change(ChangeEvent{Text: "Renamed"})
	name.Change(ChangeEvent{Text: ""})
	if len(events) != 2 || !events[0].Checked || events[1].Text != "Renamed" || invalid != 1 || c.Items[0].Completed {
		t.Fatal(events, invalid)
	}
	c.Items[0].Disabled = true
	toggle.Change(ChangeEvent{Checked: true})
	c.Items = nil
	name.Change(ChangeEvent{Text: "stale"})
	if len(events) != 2 {
		t.Fatal("stale or disabled event")
	}
	c.Items = []ChecklistItem{{ID: "a", Text: "One"}, {ID: "a", Text: "Two"}}
	if _, e = c.View(); e == nil {
		t.Fatal("duplicate item")
	}
}
