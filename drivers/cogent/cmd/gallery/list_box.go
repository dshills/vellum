package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"slices"
)

func (g *gallery) listBoxView() ui.Node {
	options := make([]ui.Option, 24)
	for i := range options {
		options[i] = ui.Option{ID: fmt.Sprintf("layer-%02d", i), Label: fmt.Sprintf("Layer %02d", i+1), Disabled: i == 2}
	}
	options[0].Label = "Background"
	options[1].Label = "Guides"
	options[2].Label = "Locked template"
	options[23].Label = "Annotations"
	if g.listReversed {
		slices.Reverse(options)
	}
	var change func(ui.ChangeEvent)
	if !g.listReadOnly {
		change = func(e ui.ChangeEvent) {
			g.listRequests++
			if !g.listReject {
				g.listSelection = e.SelectedID
			}
		}
	}
	return ui.Node{ID: "lists", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "list-help", Kind: ui.Label, Text: "Choose a layer using the pointer or Up/Down, Home/End and Page Up/Down. Type a letter to cycle matching rows. Tab leaves the list."},
		{ID: "layers", Kind: ui.ListBox, Options: options, SelectedID: g.listSelection, Disabled: g.listDisabled, Change: change, Accessibility: ui.Accessibility{Name: "Layers", Help: "Always-visible single selection. Locked template is unavailable."}},
		{ID: "list-reject", Kind: ui.Checkbox, Text: "Reject selection requests", Checked: g.listReject, Change: func(e ui.ChangeEvent) { g.listReject = e.Checked }},
		{ID: "list-readonly", Kind: ui.Checkbox, Text: "Read-only selection", Checked: g.listReadOnly, Change: func(e ui.ChangeEvent) { g.listReadOnly = e.Checked }},
		{ID: "list-disabled", Kind: ui.Checkbox, Text: "Disable list", Checked: g.listDisabled, Change: func(e ui.ChangeEvent) { g.listDisabled = e.Checked }},
		{ID: "list-reverse", Kind: ui.Checkbox, Text: "Reverse row order", Checked: g.listReversed, Change: func(e ui.ChangeEvent) { g.listReversed = e.Checked }},
		{ID: "list-clear", Kind: ui.Button, Text: "Clear selection", Activate: func() { g.listSelection = "" }},
		{ID: "list-status", Kind: ui.Label, Text: fmt.Sprintf("Accepted layer: %s · selection requests: %d", g.listSelection, g.listRequests)},
	}}
}
