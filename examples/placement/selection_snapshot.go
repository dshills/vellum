package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"maps"
	"slices"
)

// Bounded owner-side snapshot for app dialogs; never a second document model.
type selectionSnapshot struct {
	scene    *canvas.Scene
	journal  *history.Journal
	revision uint64
	objects  []canvas.Rect
	groups   map[string]string
	layers   map[string]layerMetadata
	ids      []string
	primary  string
}

func (m *model) selectionSnapshot() selectionSnapshot {
	return selectionSnapshot{m.scene, m.journal, m.journal.Snapshot().Revision, slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo), m.selectedIDs(), m.scene.Selected}
}
func (s selectionSnapshot) current(m *model) bool {
	return s.scene == m.scene && s.journal == m.journal && s.revision == m.journal.Snapshot().Revision && slices.Equal(s.objects, m.scene.Objects) && maps.Equal(s.groups, m.groups) && maps.Equal(s.layers, m.layerInfo) && slices.Equal(s.ids, m.selectedIDs()) && s.primary == m.scene.Selected
}
