package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/command"
	"maps"
)

var bulkLayerActions = []struct{ id, label string }{
	{"hide-selected", "Hide Selected"}, {"show-all", "Show All"},
	{"lock-selected", "Lock Selected"}, {"unlock-all", "Unlock All"},
}

func (m *model) bulkLayerAvailable(action string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if (action == "hide-selected" || action == "lock-selected") && len(m.selectedIDs()) == 0 {
		return errNoSelection
	}
	return nil
}
func (m *model) bindBulkLayers() error {
	for _, a := range bulkLayerActions {
		if err := m.commands.Register(command.Command{ID: a.id, Label: a.label, Available: func() error { return m.bulkLayerAvailable(a.id) }, Execute: func(context.Context) error { return m.bulkLayers(a.id) }}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) bulkLayers(action string) error {
	if err := m.bulkLayerAvailable(action); err != nil {
		return err
	}
	info := maps.Clone(m.layerInfo)
	if info == nil {
		info = map[string]layerMetadata{}
	}
	label := ""
	switch action {
	case "hide-selected", "lock-selected":
		for _, id := range m.selectedIDs() {
			unit := m.layerUnit(id)
			v := info[unit]
			if action == "hide-selected" {
				v.Hidden = true
				label = "Hide selected layers"
			} else {
				v.Locked = true
				label = "Lock selected layers"
			}
			info[unit] = v
		}
	case "show-all", "unlock-all":
		for id, v := range info {
			if action == "show-all" {
				v.Hidden = false
			} else {
				v.Locked = false
			}
			info[id] = v
		}
		if action == "show-all" {
			label = "Show all layers"
		} else {
			label = "Unlock all layers"
		}
	default:
		return errors.New("unknown bulk layer action")
	}
	if err := m.applyLayerProperties(label, prunedLayerInfo(info, m.scene.Objects, m.groups)); err != nil {
		return err
	}
	m.status = label
	return nil
}
