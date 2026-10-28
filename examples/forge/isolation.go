package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/viewport"
	"slices"
)

func (m *workbench) isolateSelection() error {
	if e := m.clipboardReady(); e != nil {
		return e
	}
	ids := m.selectionSet()
	if len(ids) == 0 {
		return fmt.Errorf("select bodies to isolate")
	}
	m.isolatedIDs = nil
	for _, o := range m.state.Objects {
		if ids[o.ID] {
			m.isolatedIDs = append(m.isolatedIDs, o.ID)
		}
	}
	m.status = fmt.Sprintf("Isolated %d bodies · Restore Visibility to return", len(m.isolatedIDs))
	m.refresh()
	return nil
}
func (m *workbench) restoreVisibility() {
	m.isolatedIDs = nil
	m.status = "Restored document visibility"
	m.refresh()
}
func (m *workbench) applyIsolation(objects []viewport.Object) {
	if len(m.isolatedIDs) == 0 {
		return
	}
	alive := false
	for _, o := range objects {
		alive = alive || slices.Contains(m.isolatedIDs, o.ID)
	}
	if !alive {
		m.isolatedIDs = nil
		return
	}
	for i := range objects {
		objects[i].Visible = objects[i].Visible && slices.Contains(m.isolatedIDs, objects[i].ID)
	}
}
func (m *workbench) isolationControls() ui.Node {
	label := "Isolate Selection"
	if len(m.isolatedIDs) > 0 {
		label = fmt.Sprintf("Isolate Selection (%d active)", len(m.isolatedIDs))
	}
	return row("isolation-actions", button("isolate-selection", label, func() { m.result(m.isolateSelection()) }), button("restore-visibility", "Restore", m.restoreVisibility))
}
func (m *workbench) bindIsolationCommands() {
	m.register("isolate", "Isolate Selection", "", func() error {
		if len(m.selectionSet()) == 0 {
			return fmt.Errorf("select bodies")
		}
		return m.clipboardReady()
	}, m.isolateSelection)
	m.register("restore-visibility", "Restore Visibility", "", func() error {
		if len(m.isolatedIDs) == 0 {
			return fmt.Errorf("not isolated")
		}
		return nil
	}, func() error { m.restoreVisibility(); return nil })
}
