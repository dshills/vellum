package main

import (
	"github.com/dshills/vellum/ui"
)

func (m *macComponentsGallery) checklistView() ui.Node {
	if m.checklist == nil {
		c := &ui.Checklist{ID: "checklist/tasks", Label: "Today", Editable: true, Items: []ui.ChecklistItem{{ID: "drawing", Text: "Review drawing", Due: "Today", Priority: 3, IconImage: m.artwork[ui.SymbolDocument]}, {ID: "references", Text: "Collect reference images", Due: "Tomorrow", IconImage: m.artwork[ui.SymbolFolder]}, {ID: "done", Text: "Create project", Completed: true, IconImage: m.artwork[ui.SymbolSave]}}}
		c.Change = func(e ui.ChecklistEvent) {
			for i := range c.Items {
				if c.Items[i].ID == e.ItemID {
					if e.Action == ui.ChecklistToggle {
						c.Items[i].Completed = e.Checked
					} else {
						c.Items[i].Text = e.Text
					}
					m.count++
					m.status = "Checklist updated"
				}
			}
		}
		c.Invalid = func(e error) { m.status = e.Error() }
		m.checklist = c
	}
	m.checklist.Dark = m.dark
	n, e := m.checklist.View()
	if e != nil {
		return ui.Node{ID: "checklist/error", Kind: ui.Label, Text: e.Error()}
	}
	return n
}
