package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"image/color"
)

// Separate pages keep these larger controls independently inspectable.
type macWidgetsGallery struct {
	sheet      *macSheetGallery
	collection *macCollectionGallery
	shortcut   *macShortcutGallery
	level      *macLevelGallery
	toolbars   *macToolbarGallery
	components *macComponentsGallery
	inputs     *inputHelpersGallery
	sidebar    *ui.Sidebar
	page       string
	dark       bool
	last       string
}

func newMacWidgetsGallery(inputs *inputHelpersGallery, dark bool) *macWidgetsGallery {
	m := &macWidgetsGallery{inputs: inputs, dark: dark, page: "mac-components"}
	m.components = newMacComponentsGallery(dark)
	m.sheet = newMacSheetGallery(dark)
	m.shortcut = newMacShortcutGallery()
	m.level = newMacLevelGallery(dark)
	m.toolbars = newMacToolbarGallery(dark)
	m.collection = newMacCollectionGallery()
	m.collection.dark = dark
	m.sidebar = &ui.Sidebar{ID: "mac-sidebar", Label: "Library navigation", Dark: dark, SelectedID: "all", Sections: []ui.SidebarSection{{ID: "favorites", Label: "Favorites", Items: []ui.Option{{ID: "all", Label: "All Drawings", Icon: ui.SymbolFolder, Badge: "12"}, {ID: "recent", Label: "Recent", Icon: ui.SymbolDocument, Badge: "3"}, {ID: "shared", Label: "Shared", Icon: ui.SymbolFolder}}}, {ID: "projects", Label: "Projects", Items: []ui.Option{{ID: "house", Label: "House", Icon: ui.SymbolFolder}, {ID: "garden", Label: "Garden", Icon: ui.SymbolFolder}, {ID: "archive", Label: "Archive", Icon: ui.SymbolFolder, Disabled: true}}}}}
	m.sidebar.Navigate = func(e ui.SidebarEvent) { m.sidebar.SelectedID = e.ItemID; m.last = e.ItemID }
	m.sidebar.Disclosure = func(e ui.SidebarDisclosureEvent) {
		for i := range m.sidebar.Sections {
			if m.sidebar.Sections[i].ID == e.SectionID {
				m.sidebar.Sections[i].Collapsed = !e.Expanded
			}
		}
	}
	return m
}
func (m *macWidgetsGallery) view() ui.Node {
	nav, err := m.sidebar.View()
	if err != nil {
		return ui.Node{ID: "mac-widget-error", Kind: ui.Label, Text: err.Error()}
	}
	nav.Style.MinSize = ui.Some(ui.Size{Width: 210, Height: 260})
	nav.Style.MaxSize = ui.Some(ui.Size{Width: 260})
	navigation := ui.Node{ID: "navigation", Kind: ui.Row, Grow: true, Children: []ui.Node{nav, {ID: "navigation-detail", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(20))}, Children: []ui.Node{{ID: "navigation-heading", Kind: ui.Label, Text: "Library", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "navigation-value", Kind: ui.Label, Text: fmt.Sprintf("Selected: %s", m.sidebar.SelectedID)}, {ID: "navigation-help", Kind: ui.Label, Text: "Choose a location. Collapse a section with its disclosure triangle. Tab moves between lists; arrows choose items."}}}}}
	options := []ui.Option{{ID: "mac-components", Label: "Components"}, {ID: "navigation", Label: "Sidebar"}, {ID: "sheets", Label: "Sheet"}, {ID: "mac-collection-page", Label: "Collection"}, {ID: "shortcuts", Label: "Shortcuts"}, {ID: "levels", Label: "Levels"}, {ID: "toolbars", Label: "Toolbars"}, {ID: "inputs", Label: "Inputs"}}
	children := []ui.Node{m.components.view(), navigation, m.sheet.view(), m.collection.view(), m.shortcut.view(), m.level.view(), m.toolbars.view(), m.inputs.view()}
	background := color.NRGBA{R: 246, G: 246, B: 248, A: 255}
	if m.dark {
		background = color.NRGBA{R: 29, G: 29, B: 31, A: 255}
	}
	for i := range children {
		if !children[i].Style.GrowX.Set {
			children[i].Style.GrowX = ui.Some(float32(1))
		}
		if !children[i].Style.Background.Set {
			children[i].Style.Background = ui.Some(background)
		}
	}
	root := ui.Node{ID: "mac-widgets-root", Kind: ui.Column, Grow: true, Style: ui.Style{Background: ui.Some(background)}, Children: []ui.Node{{ID: "mac-widgets-pages", Kind: ui.Tabs, Grow: true, Style: ui.Style{Foreground: ui.MacTheme(m.dark).Default.Foreground}, SelectedID: m.page, Options: options, Change: func(e ui.ChangeEvent) { m.page = e.SelectedID }, Children: children}}}
	if m.page == "inputs" {
		root.Focus = m.inputs.focus
	}
	return root
}
