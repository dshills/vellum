package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"slices"
	"strings"
	"unicode/utf8"
)

type paletteCommand struct{ id, label, shortcut, reason string }
type commandPaletteDraft struct {
	query, selected, message string
	entries                  []paletteCommand
	scene                    *canvas.Scene
	revision                 uint64
	ids                      []string
	primary                  string
}

func (m *model) bindCommandPalette() error {
	return m.commands.Register(command.Command{ID: "command-palette", Label: "Search Commands", Shortcut: "Meta+Shift+P", Available: m.selectionAvailable, Execute: func(context.Context) error { return m.beginCommandPalette() }})
}
func (m *model) beginCommandPalette() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	d := &commandPaletteDraft{scene: m.scene, revision: m.journal.Snapshot().Revision, ids: m.selectedIDs(), primary: m.scene.Selected}
	for _, c := range m.commands.Search("") {
		if c.ID == "place" || c.ID == "command-palette" {
			continue
		}
		label := c.Label
		if c.ID == "undo" {
			label = m.historyCaption("Undo")
		}
		if c.ID == "redo" {
			label = m.historyCaption("Redo")
		}
		entry := paletteCommand{id: c.ID, label: label, shortcut: c.Shortcut}
		if err := m.commands.Enabled(c.ID); err != nil {
			entry.reason = err.Error()
		}
		d.entries = append(d.entries, entry)
	}
	m.commandDraft = d
	m.prompt = "commands"
	m.modalEpoch++
	m.updateCommandQuery(d, "")
	return nil
}
func (d *commandPaletteDraft) matches() []paletteCommand {
	out := []paletteCommand{}
	q := strings.ToLower(strings.TrimSpace(d.query))
	for _, c := range d.entries {
		if strings.Contains(strings.ToLower(c.id+" "+c.label), q) {
			out = append(out, c)
			if len(out) == 12 {
				break
			}
		}
	}
	return out
}
func (m *model) commandDraftActive(d *commandPaletteDraft) bool {
	return d != nil && m.commandDraft == d && m.prompt == "commands"
}
func (m *model) updateCommandQuery(d *commandPaletteDraft, value string) bool {
	if !m.commandDraftActive(d) {
		return false
	}
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > 128 {
		d.message = "Search is limited to 128 characters"
		return false
	}
	if d.query != value {
		d.query = value
		d.selected = ""
	}
	rows := d.matches()
	if !slices.ContainsFunc(rows, func(c paletteCommand) bool { return c.id == d.selected }) {
		d.selected = ""
		if len(rows) > 0 {
			d.selected = rows[0].id
		}
	}
	d.message = ""
	return true
}
func (m *model) moveCommandChoice(d *commandPaletteDraft, delta int) {
	if !m.commandDraftActive(d) {
		return
	}
	rows := d.matches()
	if len(rows) == 0 {
		return
	}
	i := slices.IndexFunc(rows, func(c paletteCommand) bool { return c.id == d.selected })
	i = (i + delta + len(rows)) % len(rows)
	d.selected = rows[i].id
}
func (m *model) cancelCommandPalette(d *commandPaletteDraft) {
	if m.commandDraftActive(d) {
		m.commandDraft = nil
		m.prompt = ""
	}
}
func (m *model) runCommandChoice(ctx context.Context, d *commandPaletteDraft) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !m.commandDraftActive(d) {
		return history.ErrStale
	}
	if d.message != "" {
		return errors.New(d.message)
	}
	rows := d.matches()
	i := slices.IndexFunc(rows, func(c paletteCommand) bool { return c.id == d.selected })
	if i < 0 {
		return errors.New("choose a matching command")
	}
	if rows[i].reason != "" {
		d.message = rows[i].reason
		return errors.New(rows[i].reason)
	}
	if m.scene != d.scene || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		m.cancelCommandPalette(d)
		return history.ErrStale
	}
	id := d.selected
	m.cancelCommandPalette(d)
	return m.commands.Invoke(ctx, id) // Revalidate actual availability after closing the palette.
}
func (m *model) commandPaletteDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.commandDraft
	active := func() bool { return ref.Context().Err() == nil && m.commandDraftActive(d) }
	update := func() { _ = a.Invalidate("placement") }
	cancel := func() {
		if active() {
			m.cancelCommandPalette(d)
			update()
		}
	}
	options := []ui.Option{}
	query, selected, reason, message := "", "", "No matching commands", ""
	if d != nil {
		query, selected, message = d.query, d.selected, d.message
		for _, c := range d.matches() {
			label := c.label
			if c.shortcut != "" {
				label += "  " + c.shortcut
			}
			if c.reason != "" {
				label += " (unavailable)"
			}
			options = append(options, ui.Option{ID: c.id, Label: label})
			if c.id == selected {
				reason = c.reason
			}
		}
	}
	if m.commandKeys == nil || m.commandKeys.a != a || m.commandKeys.ref != ref {
		m.commandKeys = &commandPaletteKeys{a: a, ref: ref, m: m}
	}
	return ui.Node{ID: "command-dialog", Kind: ui.Dialog, Text: "Search Commands", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "command-guide", Kind: ui.Label, Text: "Type to filter • Up/Down chooses • Return runs • Escape cancels. First 12 matches; narrow your search.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 540})}},
		{ID: "command-query", Kind: ui.TextField, Text: query, MinWidth: 480, Accessibility: ui.Accessibility{Name: "Search commands"}, Change: func(e ui.ChangeEvent) {
			if active() {
				m.updateCommandQuery(d, e.Text)
				update()
			}
		}},
		{ID: "command-results", Kind: ui.ListBox, Options: options, SelectedID: selected, Change: func(e ui.ChangeEvent) {
			if active() && slices.ContainsFunc(d.matches(), func(c paletteCommand) bool { return c.id == e.SelectedID }) {
				d.selected = e.SelectedID
				update()
			}
		}},
		{ID: "command-reason", Kind: ui.Label, Text: reason, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 540})}},
		{ID: "command-error", Kind: ui.Label, Text: message},
		{ID: "command-run", Kind: ui.Button, Text: "Run command", Disabled: reason != "" || selected == "" || message != "", Activate: func() {
			if active() {
				if err := m.runCommandChoice(ref.Context(), d); err != nil {
					m.status = err.Error()
				}
				update()
			}
		}},
		{ID: "command-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
		{ID: "command-palette-keys", Kind: ui.Extension, Custom: m.commandKeys},
	}}
}
