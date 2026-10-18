package main

import "github.com/dshills/vellum/ui"

func (g *gallery) actionMenuControl() ui.Node {
	return ui.Node{ID: "menu-actions", Kind: ui.MenuButton, Text: "Document actions", Icon: ui.SymbolSettings, Disabled: g.menuDisabled,
		Options: []ui.Option{{ID: "blocked", Label: "Unavailable", Disabled: true}, {ID: "first", Label: "Run action"}, {ID: "second", Label: "Run action"}},
		Change:  func(e ui.ChangeEvent) { g.menuCount++; g.menuLast = e.SelectedID }}
}
func (g *gallery) viewMenuControl() ui.Node {
	return ui.Node{ID: "view-menu", Kind: ui.MenuButton, Text: "View options", Options: []ui.Option{
		{ID: "grid", Label: "Show grid", Checkable: true, Checked: g.menuChecked}, {ID: "group-divider", Separator: true}, {ID: "reset", Label: "Reset view"},
	}, Change: func(e ui.ChangeEvent) {
		g.viewMenuRequests++
		if g.viewMenuReject {
			return
		}
		if e.SelectedID == "grid" {
			g.menuChecked = e.Checked
		} else if e.SelectedID == "reset" {
			g.menuChecked = false
		}
	}}
}
