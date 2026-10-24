package main

import (
	"fmt"

	"github.com/dshills/vellum/ui"
)

type macShortcutGallery struct {
	duplicate, export           *ui.ShortcutRecorder
	duplicateError, exportError string
	disabled                    bool
}

func newMacShortcutGallery() *macShortcutGallery {
	m := &macShortcutGallery{}
	m.duplicate = &ui.ShortcutRecorder{ID: "duplicate-shortcut", Label: "Duplicate shortcut", Value: ui.Shortcut{Key: "D", Modifiers: ui.ShortcutCommand}}
	m.export = &ui.ShortcutRecorder{ID: "export-shortcut", Label: "Export shortcut", Value: ui.Shortcut{Key: "E", Modifiers: ui.ShortcutCommand | ui.ShortcutShift}}
	m.duplicate.Validate = func(v ui.Shortcut) error { return m.validate(v, m.export.Value) }
	m.export.Validate = func(v ui.Shortcut) error { return m.validate(v, m.duplicate.Value) }
	m.duplicate.Change = func(e ui.ShortcutEvent) { m.duplicate.Value = e.Value; m.duplicateError = "" }
	m.export.Change = func(e ui.ShortcutEvent) { m.export.Value = e.Value; m.exportError = "" }
	m.duplicate.Invalid = func(err error) { m.duplicateError = inputFeedback(err) }
	m.export.Invalid = func(err error) { m.exportError = inputFeedback(err) }
	return m
}
func (m *macShortcutGallery) validate(value, other ui.Shortcut) error {
	if value == (ui.Shortcut{}) {
		return nil
	}
	if value == other {
		return fmt.Errorf("This shortcut is already assigned to the other action.")
	}
	if value.Modifiers == ui.ShortcutCommand && (value.Key == "Q" || value.Key == "W") {
		return fmt.Errorf("Choose another shortcut; this one is reserved for closing the window or app.")
	}
	return nil
}
func (m *macShortcutGallery) view() ui.Node {
	// Beginning a new attempt retires feedback from the previous chord.
	if m.duplicate.Recording() {
		m.duplicateError = ""
	}
	if m.export.Recording() {
		m.exportError = ""
	}
	children := []ui.Node{{ID: "shortcut-heading", Kind: ui.Label, Text: "Keyboard shortcuts", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "shortcut-help", Kind: ui.Label, Text: "Click a shortcut or focus it and press Return. Then press a key with ⌘, ⌃ or ⌥, or a function key. Escape cancels; Tab cancels and moves focus. Delete clears."}}
	for _, item := range []struct {
		id, label string
		recorder  *ui.ShortcutRecorder
		feedback  string
	}{{"duplicate", "Duplicate", m.duplicate, m.duplicateError}, {"export", "Export", m.export, m.exportError}} {
		item.recorder.Disabled = m.disabled
		n, err := item.recorder.View()
		if err != nil {
			return ui.Node{ID: "shortcut-error", Kind: ui.Label, Text: err.Error()}
		}
		row := ui.Node{ID: item.id + "-shortcut-row", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(20))}, Children: []ui.Node{{ID: item.id + "-shortcut-name", Kind: ui.Label, Text: item.label, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 100})}}, n}}
		children = append(children, row)
		if item.feedback != "" {
			children = append(children, ui.Node{ID: item.id + "-shortcut-feedback", Kind: ui.Label, Text: item.feedback})
		}
	}
	children = append(children, ui.Node{ID: "shortcut-local-help", Kind: ui.Label, Text: "This editor changes these sample assignments only. It does not install global shortcuts. Assigning the other action’s shortcut demonstrates conflict validation."}, ui.Node{ID: "disable-shortcut-recorders", Kind: ui.Checkbox, Text: "Disable shortcut editors", Checked: m.disabled, Change: func(e ui.ChangeEvent) { m.disabled = e.Checked }}, ui.Node{ID: "reset-shortcut-recorders", Kind: ui.Button, Text: "Reset shortcuts", Activate: func() {
		m.duplicate.Cancel(m.duplicate.Session())
		m.export.Cancel(m.export.Session())
		m.duplicate.Value = ui.Shortcut{Key: "D", Modifiers: ui.ShortcutCommand}
		m.export.Value = ui.Shortcut{Key: "E", Modifiers: ui.ShortcutCommand | ui.ShortcutShift}
		m.duplicateError = ""
		m.exportError = ""
	}})
	return ui.Node{ID: "shortcuts", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(20)), Gap: ui.Some(float32(14))}, Children: children}
}
