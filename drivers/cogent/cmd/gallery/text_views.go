package main

import (
	"fmt"

	"github.com/dshills/vellum/ui"
)

func (g *gallery) projectTextField(id, name string) ui.Node {
	return ui.Node{ID: id, Kind: ui.TextField, Text: g.text, Disabled: g.textDisabled, Change: func(e ui.ChangeEvent) {
		g.textRequests++
		if !g.textReject {
			g.text = e.Text
		}
	}, Accessibility: ui.Accessibility{Name: name, Help: "Single-line native text field"}}
}
func (g *gallery) openPreferences() { g.dialogOpens++; g.dialog = true }
func (g *gallery) preferencesDialog() ui.Node {
	return ui.Node{ID: "preferences", Kind: ui.Dialog, Text: "Preferences", Open: g.dialog, Dismiss: func() { g.dialogDismisses++; g.dialog = false }, Children: []ui.Node{
		{ID: "dialog-description", Kind: ui.Label, Text: "Native modal dialog with retained content. Close with Escape or the Close button."},
		g.projectTextField("dialog-name", "Project name in preferences"),
		{ID: "apply", Kind: ui.Button, Text: "Apply and close", Activate: func() { g.dialogApplies++; g.dialog = false }},
	}}
}
func (g *gallery) textPage() ui.Node {
	s := g.notes.Snapshot()
	return ui.Node{ID: "text-page", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "text-instructions", Kind: ui.Label, Text: "Edit exact Unicode text. Command+Z / Shift+Command+Z use the document journal. Tab moves focus."},
		g.projectTextField("text-name", "Project name"),
		{ID: "text-notes", Kind: ui.MultilineText, Editor: g.notes, Grow: true, MinHeight: 180, SoftWrap: true, Disabled: g.notesDisabled, Change: func(ui.ChangeEvent) { g.editRequests++ }, Accessibility: ui.Accessibility{Name: "Project notes"}},
		{ID: "text-tools", Kind: ui.Row, Children: []ui.Node{
			{ID: "text-preferences", Kind: ui.Button, Text: "Open preferences", Activate: g.openPreferences},
			{ID: "text-disable-notes", Kind: ui.Checkbox, Text: "Disable editor", Checked: g.notesDisabled, Change: func(e ui.ChangeEvent) { g.notesDisabled = e.Checked }},
		}},
		{ID: "text-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Project %s / edits %d / undo %d / redo %d", g.text, g.editRequests, s.History.UndoCount, s.History.RedoCount)},
	}}
}
func (g *gallery) textCatalogueView() ui.Node {
	focus := g.focus
	if focus.Sequence == 0 {
		focus = ui.FocusRequest{ID: "text-name", Sequence: 1}
	}
	return ui.Node{ID: "text-catalogue", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{
		{ID: "text-title", Kind: ui.Label, Text: "Text, document history and modal ownership"}, g.textPage(), g.preferencesDialog(),
	}}
}
