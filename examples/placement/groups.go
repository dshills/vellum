package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
)

// Groups are flat membership metadata, never synthetic scene objects.
// Every selection is expanded to whole groups in the existing paint order.
func (m *model) expandedSelection(ids []string) []string {
	wanted, groups := make(map[string]bool, len(ids)), make(map[string]bool)
	for _, id := range ids {
		wanted[id] = true
		if g := m.groups[id]; g != "" {
			groups[g] = true
		}
	}
	result := make([]string, 0, len(ids))
	for _, o := range m.scene.Objects {
		if wanted[o.ID] || (m.groups[o.ID] != "" && groups[m.groups[o.ID]]) {
			result = append(result, o.ID)
		}
	}
	return result
}

func prunedGroups(groups map[string]string, objects []canvas.Rect) map[string]string {
	counts := make(map[string]int)
	for _, o := range objects {
		if g := groups[o.ID]; g != "" {
			counts[g]++
		}
	}
	result := make(map[string]string)
	for _, o := range objects {
		if g := groups[o.ID]; g != "" && counts[g] >= 2 {
			result[o.ID] = g
		}
	}
	return result
}

// Every document edit snapshots geometry, membership and selection together.
// ID counters remain monotonic outside history so a new branch cannot reuse IDs.
func (m *model) applyDocument(label string, objects []canvas.Rect, groups map[string]string, ids []string, primary string) error {
	return m.applyDocumentMetadata(label, objects, groups, prunedLayerInfo(m.layerInfo, objects, groups), ids, primary)
}
func (m *model) applyDocumentMetadata(label string, objects []canvas.Rect, groups map[string]string, info map[string]layerMetadata, ids []string, primary string) error {
	if err := validateLayerInfo(info, objects, groups); err != nil {
		return err
	}
	beforeInfo, afterInfo := maps.Clone(m.layerInfo), maps.Clone(info)
	before, beforeGroups, priorIDs, priorPrimary := slices.Clone(m.scene.Objects), maps.Clone(m.groups), m.selectedIDs(), m.scene.Selected
	after, afterGroups, nextIDs := slices.Clone(objects), maps.Clone(groups), slices.Clone(ids)
	if slices.Equal(before, after) && maps.Equal(beforeGroups, afterGroups) && maps.Equal(beforeInfo, afterInfo) {
		return nil
	}
	return m.journal.Apply(history.Change{Label: label, Apply: func() error {
		if !slices.Equal(m.scene.Objects, before) || !maps.Equal(m.groups, beforeGroups) || !maps.Equal(m.layerInfo, beforeInfo) {
			return history.ErrStale
		}
		m.scene.Objects, m.groups, m.layerInfo = slices.Clone(after), maps.Clone(afterGroups), maps.Clone(afterInfo)
		m.setSelection(nextIDs, primary)
		return nil
	}, Revert: func() error {
		if !slices.Equal(m.scene.Objects, after) || !maps.Equal(m.groups, afterGroups) || !maps.Equal(m.layerInfo, afterInfo) {
			return history.ErrStale
		}
		m.scene.Objects, m.groups, m.layerInfo = slices.Clone(before), maps.Clone(beforeGroups), maps.Clone(beforeInfo)
		m.setSelection(priorIDs, priorPrimary)
		return nil
	}})
}

// Units are group bounding boxes or single components. They are ordered by the
// first member's paint position; mapping a unit moves all its members rigidly.
func (m *model) selectionUnits() ([]canvas.Rect, map[string]string) {
	ids := m.selectedIDs()
	keys := make(map[string]string, len(ids))
	members := make(map[string][]string)
	order := make([]string, 0, len(ids))
	for _, id := range ids {
		unit := m.groups[id]
		if unit == "" {
			unit = id
		}
		keys[id] = unit
		if len(members[unit]) == 0 {
			order = append(order, unit)
		}
		members[unit] = append(members[unit], id)
	}
	units := make([]canvas.Rect, 0, len(order))
	for _, unit := range order {
		b, _ := selectionBounds(m.scene.Objects, members[unit])
		b.ID = unit
		units = append(units, b)
	}
	return units, keys
}
func (m *model) groupingAvailable(ungroup bool) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	if ungroup {
		for _, id := range m.selectedIDs() {
			if m.groups[id] != "" {
				return nil
			}
		}
		return errors.New("select a group to ungroup")
	}
	units, _ := m.selectionUnits()
	if len(units) < 2 {
		return errors.New("select at least two components or groups")
	}
	if m.nextGroup == math.MaxUint64 {
		return errors.New("group IDs exhausted")
	}
	return nil
}
func (m *model) changeGrouping(ungroup bool) error {
	if err := m.groupingAvailable(ungroup); err != nil {
		return err
	}
	ids := m.selectedIDs()
	groups := maps.Clone(m.groups)
	if groups == nil {
		groups = make(map[string]string)
	}
	label, group := "Ungroup selection", ""
	if !ungroup {
		label = "Group selection"
		group = fmt.Sprintf("group-%d", m.nextGroup+1)
	}
	for _, id := range ids {
		if group == "" {
			delete(groups, id)
		} else {
			groups[id] = group
		}
	}
	if err := m.applyDocument(label, m.scene.Objects, groups, ids, m.scene.Selected); err != nil {
		return err
	}
	if !ungroup {
		m.nextGroup++
	}
	m.status = label
	return nil
}
func (m *model) bindGrouping() error {
	for _, c := range []struct {
		id, label, shortcut string
		ungroup             bool
	}{{"group", "Group", "Meta+G", false}, {"ungroup", "Ungroup", "Meta+Shift+G", true}} {
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Shortcut: c.shortcut, ShortcutScope: command.AfterFocusShortcuts, Available: func() error { return m.groupingAvailable(c.ungroup) }, Execute: func(context.Context) error { return m.changeGrouping(c.ungroup) }}); err != nil {
			return err
		}
	}
	return nil
}
