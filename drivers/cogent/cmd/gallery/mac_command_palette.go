package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func (m *macComponentsGallery) commandPaletteView() ui.Node {
	if m.commandPalette == nil {
		registry := command.New()
		items := []ui.CommandPaletteItem{}
		for _, spec := range []struct {
			id, label, group, shortcut string
			icon                       ui.Symbol
		}{{"new", "New drawing", "File", "⌘N", ui.SymbolDocument}, {"open", "Open drawing", "File", "⌘O", ui.SymbolFolder}, {"save", "Save drawing", "File", "⌘S", ui.SymbolSave}, {"copy", "Copy selection", "Edit", "⌘C", ui.SymbolCopy}, {"paste", "Paste selection", "Edit", "⌘V", ui.SymbolPaste}, {"undo", "Undo", "Edit", "⌘Z", ui.SymbolUndo}, {"redo", "Redo", "Edit", "⇧⌘Z", ui.SymbolRedo}} {
			id, label := spec.id, spec.label
			_ = registry.Register(command.Command{ID: id, Label: label, Execute: func(context.Context) error { m.status = "Command: " + label; m.count++; return nil }})
			items = append(items, ui.CommandPaletteItem{ID: id, Label: label, Group: spec.group, Shortcut: spec.shortcut, IconImage: m.artwork[spec.icon], Available: func() error { return registry.Enabled(id) }})
		}
		m.commandPalette = &ui.CommandPalette{ID: "commands/palette", ReturnFocusID: "commands/open", Items: items, Invoke: func(e ui.CommandPaletteEvent) error { return registry.Invoke(context.Background(), e.CommandID) }, OnError: func(e error) { m.status = fmt.Sprint(e) }}
	}
	p := m.commandPalette
	p.Dark = m.dark
	n, e := p.View()
	if e != nil {
		return ui.Node{ID: "commands/error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "commands/page", Kind: ui.Column, Children: []ui.Node{{ID: "commands/help", Kind: ui.Label, Text: "A reusable command search panel with grouped results and full-color icons. Type to filter, use ↑↓ and Return, or choose a result. Escape cancels. The demo invokes a command registry and reports actions; shortcut captions illustrate app-owned bindings."}, {ID: "commands/open", Kind: ui.Button, Text: "Search Commands…", IconImage: m.artwork[ui.SymbolSearch], Activate: func() { p.Query = ""; p.SelectedID = ""; p.Open = true }}, n}}
}
