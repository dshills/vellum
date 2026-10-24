package main

import (
	"fmt"
	"image/color"

	"github.com/dshills/vellum/ui"
)

// The toolbar catalogue composes only existing public nodes and local styles.
// Applications can copy these recipes without depending on gallery internals.
type macToolbarGallery struct {
	dark             bool
	display          string
	disabled, narrow bool
	count            int
	last             string
}

type macToolbarMetrics struct {
	name                  string
	height, font, padding float32
}

var macToolbarSizes = []macToolbarMetrics{{"Small", 26, 12, 6}, {"Regular", 32, 13, 8}, {"Large", 40, 16, 10}}

func newMacToolbarGallery(dark bool) *macToolbarGallery {
	return &macToolbarGallery{dark: dark, display: "both", last: "No action yet"}
}
func (m *macToolbarGallery) action(prefix, name string, icon ui.Symbol, size macToolbarMetrics, display string) ui.Node {
	text, art := name, icon
	if display == "icons" {
		text = ""
	}
	if display == "text" {
		art = ""
	}
	style := ui.Style{Grow: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{Top: 3, Bottom: 3, Left: size.padding, Right: size.padding}), Gap: ui.Some(float32(5)), MinSize: ui.Some(ui.Size{Width: size.height, Height: size.height}), MaxSize: ui.Some(ui.Size{Height: size.height}), Typography: ui.Some(ui.Typography{Size: size.font}), CornerRadius: ui.Some(float32(5)), Border: ui.Some(ui.Border{}), Background: ui.Some(color.NRGBA{})}
	if display == "icons" {
		style.MaxSize = ui.Some(ui.Size{Width: size.height, Height: size.height})
		style.Padding = ui.Some(ui.UniformInsets(3))
	}
	return ui.Node{ID: prefix + "/" + name, Kind: ui.Button, Text: text, Icon: art, Appearance: ui.ButtonTool, Style: style, Disabled: m.disabled || name == "Delete", Accessibility: ui.Accessibility{Name: name, Help: name + " — sample action"}, Activate: func() {
		if !m.disabled && name != "Delete" {
			m.count++
			m.last = prefix + ": " + name
		}
	}}
}
func (m *macToolbarGallery) toolbar(id string, size macToolbarMetrics, display string, vertical bool) ui.Node {
	theme := ui.MacTheme(m.dark)
	style := ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)), Padding: ui.Some(ui.UniformInsets(5)), Gap: ui.Some(float32(4)), Background: theme.Kinds[ui.FormGroup].Background, Border: theme.Kinds[ui.FormGroup].Border, CornerRadius: ui.Some(float32(8)), Align: ui.Some(ui.AlignCenter)}
	orientation := ui.Horizontal
	if m.narrow {
		style.MaxSize = ui.Some(ui.Size{Width: 280})
	}
	if vertical {
		orientation = ui.Vertical
		style.GrowX = ui.Some(float32(0))
		style.MaxSize = ui.Some(ui.Size{Width: 72, Height: 154})
	}
	children := []ui.Node{}
	for i, item := range []struct {
		name string
		icon ui.Symbol
	}{{"New", ui.SymbolAdd}, {"Open", ui.SymbolFolder}, {"Save", ui.SymbolSave}, {"Undo", ui.SymbolUndo}, {"Redo", ui.SymbolRedo}, {"Delete", ui.SymbolDelete}} {
		if i == 3 {
			sep := ui.Node{ID: id + "/separator", Kind: ui.Separator, Orientation: ui.Vertical, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 1, Height: size.height - 8}), MaxSize: ui.Some(ui.Size{Width: 1, Height: size.height - 8})}}
			if vertical {
				sep.Orientation = ui.Horizontal
				sep.Style.MinSize = ui.Some(ui.Size{Width: 24, Height: 1})
				sep.Style.MaxSize = sep.Style.MinSize
			}
			children = append(children, sep)
		}
		children = append(children, m.action(id, item.name, item.icon, size, display))
	}
	return ui.Node{ID: id, Kind: ui.Toolbar, Orientation: orientation, Disabled: m.disabled, Style: style, Accessibility: ui.Accessibility{Name: size.name + " " + display + " toolbar", Help: "Tab enters controls. Arrows move between enabled actions; Home and End reveal the first and last."}, Children: children}
}
func (m *macToolbarGallery) view() ui.Node {
	background := color.NRGBA{R: 246, G: 246, B: 248, A: 255}
	if m.dark {
		background = color.NRGBA{R: 29, G: 29, B: 31, A: 255}
	}
	modes := []ui.Node{}
	for _, mode := range []struct{ id, label string }{{"icons", "Icons"}, {"text", "Text"}, {"both", "Icons and text"}} {
		mode := mode
		n := ui.Node{ID: "toolbar-display/" + mode.id, Kind: ui.Button, Text: mode.label, Activate: func() { m.display = mode.id }, Accessibility: ui.Accessibility{Help: "Show toolbar actions as " + mode.label}, Style: ui.Style{Grow: ui.Some(float32(0))}}
		if m.display == mode.id {
			n.Style.Background = ui.Some(color.NRGBA{R: 0, G: 112, B: 245, A: 255})
			n.Style.Foreground = ui.Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
		modes = append(modes, n)
	}
	children := []ui.Node{{ID: "toolbar-title", Kind: ui.Label, Text: "Toolbars", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "toolbar-help", Kind: ui.Label, Text: "Compare sizes and display modes. Actions update the status below; Delete is deliberately disabled."}, {ID: "toolbar-display", Kind: ui.Row, Children: modes}, {ID: "toolbar-options", Kind: ui.Row, Children: []ui.Node{{ID: "toolbar-narrow", Kind: ui.Checkbox, Text: "Narrow preview", Checked: m.narrow, Change: func(e ui.ChangeEvent) { m.narrow = e.Checked }}, {ID: "toolbar-disable", Kind: ui.Checkbox, Text: "Disable actions", Checked: m.disabled, Change: func(e ui.ChangeEvent) { m.disabled = e.Checked }}}}}
	children = append(children, ui.Node{ID: "toolbar-status", Kind: ui.Label, Text: fmt.Sprintf("%s · actions: %d", m.last, m.count)})
	for _, size := range macToolbarSizes {
		id := "toolbar-" + size.name
		children = append(children, ui.Node{ID: id + "/label", Kind: ui.Label, Text: fmt.Sprintf("%s · %g pt controls", size.name, size.height)}, m.toolbar(id, size, m.display, false))
	}
	mixed := m.toolbar("toolbar-mixed", macToolbarSizes[1], "icons", false)
	mixed.Accessibility.Name = "Mixed document toolbar"
	// These are ordinary actions: explicit per-item presentation can coexist.
	for i, child := range mixed.Children {
		switch child.ID {
		case "toolbar-mixed/New":
			mixed.Children[i] = m.action("toolbar-mixed", "New", ui.SymbolAdd, macToolbarSizes[1], "both")
		case "toolbar-mixed/Open":
			mixed.Children[i] = m.action("toolbar-mixed", "Open", ui.SymbolFolder, macToolbarSizes[1], "text")
		}
	}
	children = append(children, ui.Node{ID: "toolbar-mixed-label", Kind: ui.Label, Text: "Mixed · text, icons and labels together"}, mixed, ui.Node{ID: "toolbar-vertical-label", Kind: ui.Label, Text: "Vertical · scroll to reveal more tools"}, m.toolbar("toolbar-vertical", macToolbarSizes[1], "icons", true), ui.Node{ID: "toolbar-limits", Kind: ui.Label, Text: "Content toolbars with retained scrolling overflow. No title-bar integration, customization palette, or overflow menu."})
	return ui.Node{ID: "toolbars", Kind: ui.Scroll, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(10)), Background: ui.Some(background)}, Children: children}
}
