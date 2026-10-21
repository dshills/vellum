package main

import (
	"errors"
	"fmt"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

// The managed canvas keeps one primary selection. The app adds a bounded set,
// ordered by paint order, and reconciles managed keyboard/accessibility changes.
func (m *model) setSelection(ids []string, primary string) {
	selected := m.expandedSelection(ids)
	if !slices.Contains(selected, primary) {
		primary = ""
		if len(selected) > 0 {
			primary = selected[len(selected)-1]
		}
	}
	m.selectionIDs, m.selectionPrimary, m.scene.Selected = selected, primary, primary
	m.layerFocus = primary
}

// Reading selection is side-effect free, including during View construction.
func (m *model) selectedIDs() []string {
	if m.scene.Selected == "" {
		return nil
	}
	ids := m.selectionIDs
	if m.scene.Selected != m.selectionPrimary {
		ids = []string{m.scene.Selected}
	}
	return m.expandedSelection(ids)
}

// Managed callbacks represent explicit single-selection actions. Navigation is
// routed separately by the app adapter so reselecting the primary also collapses.
func (m *model) managedSelection(id string) {
	m.cancelLayerDrag(m.layerDrag)
	if m.move != nil {
		m.cancelMove(m.move)
	}
	m.setSelection([]string{id}, id)
}

func (m *model) resizeAvailable() error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	if len(m.selectedIDs()) != 1 {
		return errors.New("select exactly one component to resize")
	}
	return nil
}
func (m *model) selectionToggle(g placement.Geometry, p canvas.Point) *objectMove {
	if m.fileBusy() || m.prompt != "" || m.move != nil || m.palette.State().Active {
		return nil
	}
	world, err := g.World(p)
	if err != nil {
		return nil
	}
	ids := m.selectedIDs()
	id := m.objectAtWorld(world)
	primary := m.scene.Selected
	if id == "" {
		m.scene.View = g.View
		s := &objectMove{id: primary, before: slices.Clone(m.scene.Objects), selection: ids, start: p, grab: world, geometry: g, valid: true, marquee: true}
		m.move = s
		return s
	}
	unit := m.expandedSelection([]string{id})
	if slices.Contains(ids, id) {
		ids = slices.DeleteFunc(ids, func(member string) bool { return slices.Contains(unit, member) })
	} else {
		ids = append(ids, unit...)
		primary = id
	}
	m.setSelection(ids, primary)
	m.scene.View = g.View
	s := &objectMove{id: m.scene.Selected, before: slices.Clone(m.scene.Objects), selection: m.selectedIDs(), start: p, grab: world, geometry: g, valid: true, selectionOnly: true}
	m.move = s
	m.status = fmt.Sprintf("%d selected components", len(s.selection))
	return s
}

func movedSelection(before []canvas.Rect, ids []string, dx, dy float64) ([]canvas.Rect, error) {
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	after := slices.Clone(before)
	for i := range after {
		if !selected[after[i].ID] {
			continue
		}
		x, y := after[i].X+dx, after[i].Y+dy
		if !validPosition(x, y) {
			return nil, errors.New("all selected components must remain within ±10000")
		}
		after[i].X, after[i].Y = x, y
	}
	return after, nil
}
