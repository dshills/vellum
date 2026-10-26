package main

import "github.com/dshills/vellum/ui"

func (m *macComponentsGallery) helpTagsView() ui.Node {
	save, _ := ui.WithTooltip(ui.Node{ID: "help/save", Kind: ui.Button, Text: "Save drawing", IconImage: m.artwork[ui.SymbolSave], Activate: func() { m.count++; m.status = "Save demonstration" }}, "Save the current drawing (⌘S)")
	field, _ := ui.WithTooltip(ui.Node{ID: "help/name", Kind: ui.TextField, Text: m.name, Accessibility: ui.Accessibility{Name: "Drawing name"}, Change: func(e ui.ChangeEvent) { m.name = e.Text }}, "Give your drawing a descriptive name.")
	return ui.Node{ID: "help/page", Kind: ui.Column, Children: []ui.Node{{ID: "help/explanation", Kind: ui.Label, Text: "Pause over a control for help. With the control focused, press F1. Escape, Tab, clicking or leaving the target dismisses the tag."}, save, field}}
}
