package main

import (
	"errors"
	"maps"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

type layerMetadata struct {
	Name   string `json:"name,omitempty"`
	Hidden bool   `json:"hidden,omitempty"`
	Locked bool   `json:"locked,omitempty"`
}

func validLayerName(name string) bool {
	if !utf8.ValidString(name) || len(name) > 128 || utf8.RuneCountInString(name) > 80 || strings.TrimSpace(name) != name {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func liveLayerIDs(objects []canvas.Rect, groups map[string]string) map[string]bool {
	live := make(map[string]bool, len(objects)+len(groups))
	for _, o := range objects {
		live[o.ID] = true
		if g := groups[o.ID]; g != "" {
			live[g] = true
		}
	}
	return live
}
func prunedLayerInfo(info map[string]layerMetadata, objects []canvas.Rect, groups map[string]string) map[string]layerMetadata {
	live := liveLayerIDs(objects, groups)
	result := make(map[string]layerMetadata)
	for id, v := range info {
		if live[id] && v != (layerMetadata{}) {
			result[id] = v
		}
	}
	return result
}
func validateLayerInfo(info map[string]layerMetadata, objects []canvas.Rect, groups map[string]string) error {
	live := liveLayerIDs(objects, groups)
	if len(info) > 768 {
		return errors.New("too many layer properties")
	}
	for id, v := range info {
		if !live[id] || !validLayerName(v.Name) {
			return errors.New("invalid layer properties")
		}
	}
	return nil
}
func (m *model) layerLabel(id string) string {
	if name := m.layerInfo[id].Name; name != "" {
		return name + " · " + id
	}
	return id
}
func (m *model) renameLayer(id, name string) error {
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if !liveLayerIDs(m.scene.Objects, m.groups)[id] || !validLayerName(name) {
		return errors.New("choose a live layer and a name of at most80 characters/128 bytes without control characters")
	}
	info := maps.Clone(m.layerInfo)
	if info == nil {
		info = make(map[string]layerMetadata)
	}
	v := info[id]
	v.Name = name
	info[id] = v
	info = prunedLayerInfo(info, m.scene.Objects, m.groups)
	if err := m.applyLayerProperties("Rename layer", info); err != nil {
		return err
	}
	m.status = "Renamed " + id
	return nil
}
func (m *model) renameDialog(a *app.Application, ref app.WindowRef) ui.Node {
	epoch, target := m.renameEpoch, m.renameID
	active := func() bool {
		return ref.Context().Err() == nil && m.prompt == "rename" && m.renameEpoch == epoch && m.renameID == target
	}
	cancel := func() {
		if !active() {
			return
		}
		m.renameID = ""
		m.prompt = ""
		_ = a.Invalidate("placement")
	}
	return ui.Node{ID: "layer-rename-dialog", Kind: ui.Dialog, Text: "Rename layer", Open: m.prompt == "rename", Dismiss: cancel, Children: []ui.Node{
		{ID: "layer-rename-target", Kind: ui.Label, Text: m.renameID},
		{ID: "layer-name", Kind: ui.TextField, Text: m.renameDraft, Accessibility: ui.Accessibility{Name: "Layer name"}, Change: func(e ui.ChangeEvent) {
			if active() {
				m.renameDraft = e.Text
			}
		}},
		{ID: "layer-name-save", Kind: ui.Button, Text: "Rename", Activate: func() {
			if !active() {
				return
			}
			id, name := m.renameID, m.renameDraft
			m.prompt = ""
			if err := m.renameLayer(id, name); err != nil {
				m.status = err.Error()
				m.prompt = "rename"
			} else {
				m.renameID = ""
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "layer-name-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}
func (m *model) layerActionTarget() string {
	if m.layerFocus != "" && liveLayerIDs(m.scene.Objects, m.groups)[m.layerFocus] {
		return m.layerFocus
	}
	return m.scene.Selected
}

func (m *model) applyLayerProperties(label string, info map[string]layerMetadata) error {
	f, err := drawingCodec(m).Capture()
	if err != nil {
		return err
	}
	if f.Version < 4 {
		f.Version = 4
	}
	f.Layers = info
	if _, err = drawingCodec(m).Encode(f); err != nil {
		return err
	}
	return m.applyDocumentMetadata(label, m.scene.Objects, m.groups, info, m.selectedIDs(), m.scene.Selected)
}
