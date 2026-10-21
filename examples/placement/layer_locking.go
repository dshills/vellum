package main

import (
	"errors"
	"maps"
)

func (m *model) layerLocked(id string) bool {
	unit := m.layerUnit(id)
	if m.layerInfo[unit].Locked {
		return true
	}
	for member, g := range m.groups {
		if g == unit && m.layerInfo[member].Locked {
			return true
		}
	}
	return false
}
func (m *model) toggleLayerLock(id string) error {
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	if !liveLayerIDs(m.scene.Objects, m.groups)[id] {
		return errors.New("layer no longer available")
	}
	unit := m.layerUnit(id)
	info := maps.Clone(m.layerInfo)
	if info == nil {
		info = make(map[string]layerMetadata)
	}
	locked := !m.layerLocked(id)
	v := info[unit]
	v.Locked = locked
	info[unit] = v
	if !locked {
		for member, g := range m.groups {
			if g == unit {
				v := info[member]
				v.Locked = false
				info[member] = v
			}
		}
	}
	if err := m.applyLayerProperties("Change layer lock", prunedLayerInfo(info, m.scene.Objects, m.groups)); err != nil {
		return err
	}
	m.layerFocus = id
	m.status = "Layer lock changed"
	return nil
}
