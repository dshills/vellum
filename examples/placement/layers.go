package main

import (
	"errors"
	"fmt"
	"image/color"
	"slices"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

type layerRow struct {
	id, label string
	selected  bool
}

// A fresh projection, bounded by 512 components plus at most 256 group markers.
// Only component rows encode paint order. Each group's selector precedes its
// frontmost member, without pretending interleaved flat groups are a tree.
func (m *model) allLayerRows() []layerRow {
	selected := make(map[string]bool)
	for _, id := range m.selectedIDs() {
		selected[id] = true
	}
	counts, seen := make(map[string]int), make(map[string]bool)
	for _, o := range m.scene.Objects {
		if g := m.groups[o.ID]; g != "" {
			counts[g]++
		}
	}
	rows := make([]layerRow, 0, len(m.scene.Objects)+len(counts))
	for i := len(m.scene.Objects) - 1; i >= 0; i-- {
		o := m.scene.Objects[i]
		g := m.groups[o.ID]
		if g != "" && !seen[g] {
			label := fmt.Sprintf("Group %s · %d", m.layerLabel(g), counts[g]) + m.primaryLayerSuffix(g)
			if m.layerLocked(g) {
				label += " [locked]"
			}
			if m.layerHidden(g) {
				label += " [hidden]"
			}
			rows = append(rows, layerRow{g, label, selected[o.ID]})
			seen[g] = true
		}
		label := m.layerLabel(o.ID) + m.primaryLayerSuffix(o.ID)
		if m.layerLocked(o.ID) {
			label += " [locked]"
		}
		if m.layerHidden(o.ID) {
			label += " [hidden]"
		}
		if g != "" {
			label += " · " + m.layerLabel(g)
		}
		rows = append(rows, layerRow{o.ID, label, selected[o.ID]})
	}
	return rows
}

// Resolve stable IDs at activation time, not indices from a prior View. Group
// selection uses its current frontmost member as primary. Selection is transient
// and shares the canvas set; it never creates or branches document history.
func (m *model) selectLayer(id string) error {
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	if id == "" {
		m.setSelection(nil, "")
		m.layerFocus = ""
		m.status = "Cleared selection"
		return nil
	}
	primary := ""
	for i := len(m.scene.Objects) - 1; i >= 0; i-- {
		o := m.scene.Objects[i]
		if o.ID == id || m.groups[o.ID] == id {
			primary = o.ID
			break
		}
	}
	if primary == "" {
		return errors.New("layer is no longer available")
	}
	ids := m.expandedSelection([]string{primary})
	if m.layerAdd {
		current := m.selectedIDs()
		if slices.Contains(current, primary) {
			ids = slices.DeleteFunc(current, func(member string) bool { return slices.Contains(ids, member) })
			primary = m.scene.Selected
		} else {
			ids = append(current, ids...)
		}
	}
	m.setSelection(ids, primary)
	m.layerFocus = id
	m.status = fmt.Sprintf("Layer selection: %d components", len(m.selectedIDs()))
	return nil
}

func (m *model) layersPanel(ref app.WindowRef, update func(), invoke func(string) func()) ui.Node {
	disabled := m.clipboardAvailable() != nil && m.layerDrag == nil
	finish := func(action func() error) {
		if ref.Context().Err() != nil {
			return
		}
		if err := action(); err != nil {
			m.status = err.Error()
		}
		update()
	}
	rows := []ui.Node{}
	for _, row := range m.layerRows() {
		text := row.label
		if row.selected {
			text = "✓ " + text
		}
		style := ui.Style{Padding: ui.Some(ui.Insets{Top: 2, Right: 4, Bottom: 2, Left: 4}), GrowX: ui.Some(float32(1)), TextAlign: ui.Some(ui.AlignStart), Typography: ui.Some(ui.Typography{Size: 12}), CornerRadius: ui.Some(float32(3))}
		if row.selected {
			style.Background = ui.Some(color.NRGBA{R: 218, G: 232, B: 252, A: 255})
		}
		rows = append(rows, ui.Node{ID: "layer-select-" + row.id, Kind: ui.Button, Text: text, Appearance: ui.ButtonQuiet, Disabled: disabled, Style: style, Accessibility: ui.Accessibility{Name: "Layer " + text, Help: "Click to select; drag to reorder when no filter is active. Groups menu collapses or expands member rows. Escape cancels. Add selection toggles membership."}, Activate: func() { finish(func() error { return m.selectLayer(row.id) }) }})
	}
	if len(rows) == 0 {
		rows = append(rows, ui.Node{ID: "layers-empty", Kind: ui.Label, Text: m.emptyLayersLabel()})
	}
	compact := ui.Style{Padding: ui.Some(ui.UniformInsets(4)), Typography: ui.Some(ui.Typography{Size: 12})}
	lockLabel := "Lock"
	if m.layerLocked(m.layerActionTarget()) {
		lockLabel = "Unlock"
	}
	visibilityLabel := "Hide"
	if m.layerHidden(m.layerActionTarget()) {
		visibilityLabel = "Show"
	}
	return ui.Node{ID: "layers-panel", Kind: ui.Column, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 260}), MaxSize: ui.Some(ui.Size{Width: 260}), GrowY: ui.Some(float32(1)), Gap: ui.Some(float32(4)), Padding: ui.Some(ui.UniformInsets(4))}, Accessibility: ui.Accessibility{Name: "Layers panel"}, Children: []ui.Node{
		{ID: "layers-heading", Kind: ui.Label, Text: "Layers · front first"},
		m.layerOrganizationControls(ref, update),
		{ID: "layers-options", Kind: ui.Row, Children: []ui.Node{
			{ID: "layers-add", Kind: ui.Checkbox, Text: "Add selection", Checked: m.layerAdd, Disabled: disabled, Change: func(e ui.ChangeEvent) {
				finish(func() error {
					if err := m.clipboardAvailable(); err != nil {
						return err
					}
					m.layerAdd = e.Checked
					return nil
				})
			}},
			{ID: "layers-clear", Kind: ui.Button, Text: "Clear", Style: compact, Disabled: disabled || len(m.selectedIDs()) == 0, Activate: func() { finish(func() error { return m.selectLayer("") }) }},
		}},
		{ID: "layers-order", Kind: ui.Row, Children: []ui.Node{
			{ID: "layers-forward", Kind: ui.Button, Text: "Forward", Style: compact, Disabled: m.commandDisabled("bring-forward"), Activate: invoke("bring-forward")},
			{ID: "layers-backward", Kind: ui.Button, Text: "Backward", Style: compact, Disabled: m.commandDisabled("send-backward"), Activate: invoke("send-backward")},
		}},
		{ID: "layers-properties", Kind: ui.Row, Children: []ui.Node{{ID: "layers-rename", Kind: ui.Button, Text: "Rename…", Style: compact, Disabled: disabled || m.layerActionTarget() == "", Activate: func() {
			if ref.Context().Err() != nil || m.clipboardAvailable() != nil {
				return
			}
			if m.renameEpoch == ^uint64(0) {
				return
			}
			m.renameEpoch++
			m.renameID = m.layerActionTarget()
			m.renameDraft = m.layerInfo[m.renameID].Name
			m.prompt = "rename"
			update()
		}}, {ID: "layers-visible", Kind: ui.Button, Text: visibilityLabel, Style: compact, Disabled: disabled || m.layerActionTarget() == "", Activate: func() { finish(func() error { return m.toggleLayerVisibility(m.layerActionTarget()) }) }}, {ID: "layers-locked", Kind: ui.Button, Text: lockLabel, Style: compact, Disabled: disabled || m.layerActionTarget() == "", Activate: func() { finish(func() error { return m.toggleLayerLock(m.layerActionTarget()) }) }}}},
		{ID: "layers-scroll", Kind: ui.Scroll, Grow: true, Children: []ui.Node{{ID: "layers-rows", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1)), Gap: ui.Some(float32(1))}, Children: rows}}},
	}}
}
