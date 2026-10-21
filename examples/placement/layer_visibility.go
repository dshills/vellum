package main

import (
	"errors"
	"github.com/dshills/vellum/canvas"
	"maps"
)

func (m *model) layerUnit(id string) string {
	if g := m.groups[id]; g != "" {
		return g
	}
	return id
}
func (m *model) layerUnitIDs(id string) []string {
	unit := m.layerUnit(id)
	ids := []string{}
	for _, o := range m.scene.Objects {
		if o.ID == unit || m.groups[o.ID] == unit {
			ids = append(ids, o.ID)
		}
	}
	return ids
}
func (m *model) layerHidden(id string) bool {
	unit := m.layerUnit(id)
	if m.layerInfo[unit].Hidden {
		return true
	}
	for member, g := range m.groups {
		if g == unit && m.layerInfo[member].Hidden {
			return true
		}
	}
	return false
}
func (m *model) toggleLayerVisibility(id string) error {
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
	hidden := !m.layerHidden(id)
	v := info[unit]
	v.Hidden = hidden
	info[unit] = v
	// Groups remain atomic interaction units. Showing a group clears any
	// member flags from imported files, without changing names or geometry.
	if !hidden {
		for member, g := range m.groups {
			if g == unit {
				v := info[member]
				v.Hidden = false
				info[member] = v
			}
		}
	}
	info = prunedLayerInfo(info, m.scene.Objects, m.groups)
	if err := m.applyLayerProperties("Change layer visibility", info); err != nil {
		return err
	}
	m.layerFocus = id
	m.status = "Layer visibility changed"
	return nil
}
func (m *model) visibleObjects() []canvas.Rect {
	objects := make([]canvas.Rect, 0, len(m.scene.Objects))
	for _, o := range m.scene.Objects {
		if !m.layerHidden(o.ID) {
			objects = append(objects, o)
		}
	}
	return objects
}

// The document retains all objects; this app-local projection controls painting
// and managed semantic picking without changing framework canvas behavior.
func (m *model) canvasForView() *canvas.Scene {
	visible := m.visibleObjects()
	if len(visible) == len(m.scene.Objects) {
		m.visibleScene = nil
		return m.scene
	}
	if m.visibleScene == nil {
		m.visibleScene = &canvas.Scene{View: m.scene.View}
	} else if m.visibleScene.View.Width > 0 && m.visibleScene.View.Height > 0 {
		m.scene.View.Width, m.scene.View.Height = m.visibleScene.View.Width, m.visibleScene.View.Height
	}
	m.visibleScene.Objects = visible
	m.visibleScene.View = m.scene.View
	m.visibleScene.Selected = ""
	for _, o := range visible {
		if o.ID == m.scene.Selected {
			m.visibleScene.Selected = o.ID
			break
		}
	}
	return m.visibleScene
}
func (m *model) objectAtWorld(p canvas.Point) string {
	for i := len(m.scene.Objects) - 1; i >= 0; i-- {
		o := m.scene.Objects[i]
		if !m.layerHidden(o.ID) && !m.layerLocked(o.ID) && p.X >= o.X && p.X <= o.X+o.Width && p.Y >= o.Y && p.Y <= o.Y+o.Height {
			return o.ID
		}
	}
	return ""
}
func (m *model) fitVisible() {
	visible := m.visibleObjects()
	if len(visible) == 0 {
		return
	}
	s := canvas.Scene{Objects: visible, View: m.scene.View}
	s.Fit()
	m.scene.View = s.View
}
