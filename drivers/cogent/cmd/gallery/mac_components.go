package main

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image/color"
)

type macComponentsGallery struct {
	inspector *ui.Inspector
	width     float32
	splits    []*ui.SplitButton
	menu      *ui.MenuBranch
	dark      bool
	accent    string
	page      string
	name      string
	status    string
	count     int
	artwork   map[ui.Symbol]*canvas.TiledImage
}

func newMacComponentsGallery(dark bool) *macComponentsGallery {
	artwork, err := toolbarColorIcons()
	if err != nil {
		panic(err)
	}
	branch, err := ui.NewMenuBranch([]ui.Option{{ID: "export-png", Label: "PNG image", IconImage: artwork[ui.SymbolSave], Shortcut: "⇧⌘E"}, {ID: "export-pdf", Label: "PDF document", IconImage: artwork[ui.SymbolFolder]}, {ID: "export-blocked", Label: "Cloud export (unavailable)", Disabled: true}})
	if err != nil {
		panic(err)
	}
	return &macComponentsGallery{menu: branch, dark: dark, accent: "blue", page: "theme", name: "Untitled", status: "No action yet", artwork: artwork}
}
func (m *macComponentsGallery) theme() ui.Theme {
	accent := color.NRGBA{G: 112, B: 245, A: 255}
	switch m.accent {
	case "purple":
		accent = color.NRGBA{R: 126, G: 65, B: 194, A: 255}
	case "green":
		accent = color.NRGBA{R: 25, G: 125, B: 80, A: 255}
	}
	return ui.MacThemeWithAccent(m.dark, accent)
}
func (m *macComponentsGallery) view() ui.Node {
	theme := m.theme()
	palette := theme.Palette.Value
	appearance := []ui.Node{}
	for _, choice := range []struct {
		label string
		dark  bool
	}{{"Light", false}, {"Dark", true}} {
		style := ui.Style{GrowX: ui.Some(float32(0))}
		if m.dark == choice.dark {
			style.Background = ui.Some(palette.Accent)
			style.Foreground = ui.Some(palette.AccentInk)
		}
		appearance = append(appearance, ui.Node{ID: "components/appearance/" + choice.label, Kind: ui.Button, Text: choice.label, Style: style, Activate: func() { m.dark = choice.dark }})
	}
	appearance = append(appearance, ui.Node{ID: "components/accent", Kind: ui.Selector, SelectedID: m.accent, Options: []ui.Option{{ID: "blue", Label: "Blue accent"}, {ID: "purple", Label: "Purple accent"}, {ID: "green", Label: "Green accent"}}, Accessibility: ui.Accessibility{Name: "Component accent"}, Change: func(e ui.ChangeEvent) {
		if e.SelectedID == "blue" || e.SelectedID == "purple" || e.SelectedID == "green" {
			m.accent = e.SelectedID
		}
	}})
	content := m.themePreview()
	if m.page == "inspector" {
		content = m.inspectorPreview()
	}
	if m.page == "split" {
		content = m.splitPreview()
	}
	if m.page == "menus" {
		content = m.menuPreview()
	}
	navigation := ui.Node{ID: "components/pages", Kind: ui.Selector, SelectedID: m.page, Options: []ui.Option{{ID: "theme", Label: "Themes"}, {ID: "menus", Label: "Cascading menus"}, {ID: "split", Label: "Split buttons"}, {ID: "inspector", Label: "Inspector panel"}}, Accessibility: ui.Accessibility{Name: "Component demonstration"}, Change: func(e ui.ChangeEvent) { m.page = e.SelectedID }}

	return ui.Node{ID: "mac-components", Kind: ui.Scroll, Theme: &theme, Grow: true, Style: ui.Style{Background: ui.Some(palette.Window), Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}, Children: []ui.Node{{ID: "components/title", Kind: ui.Label, Text: "Mac components", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "components/appearance", Kind: ui.Row, Children: appearance}, navigation, content, {ID: "components/status", Kind: ui.Label, Text: fmt.Sprintf("%s · actions: %d", m.status, m.count)}}}
}
func (m *macComponentsGallery) themePreview() ui.Node {
	return ui.Node{ID: "components/theme-preview", Kind: ui.FormGroup, Text: "Reusable scoped themes", Children: []ui.Node{{ID: "components/theme-help", Kind: ui.Label, Text: "Switch appearance and accent. Controls retain their values and full-color artwork."}, {ID: "components/name", Kind: ui.TextField, Text: m.name, Accessibility: ui.Accessibility{Name: "Theme preview name"}, Change: func(e ui.ChangeEvent) { m.name = e.Text }}, {ID: "components/color-button", Kind: ui.Button, Text: "Save preview", IconImage: m.artwork[ui.SymbolSave], Style: ui.Style{GrowX: ui.Some(float32(0))}, Activate: func() { m.count++; m.status = "Theme preview saved" }}}}
}

func (m *macComponentsGallery) menuPreview() ui.Node {
	return ui.Node{ID: "components/menu-demo", Kind: ui.FormGroup, Text: "Cascading action menus", Children: []ui.Node{
		{ID: "components/menu-help", Kind: ui.Label, Text: "Open Export, then use Right / Left to navigate. Escape cancels. Shortcut captions do not install bindings."},
		{ID: "components/actions", Kind: ui.MenuButton, Text: "Document actions", IconImage: m.artwork[ui.SymbolFolder], Options: []ui.Option{{ID: "export", Label: "Export", IconImage: m.artwork[ui.SymbolSave], Branch: m.menu}, {ID: "separator", Separator: true}, {ID: "grid", Label: "Show grid", Checkable: true, Checked: m.count%2 == 1}}, Change: func(e ui.ChangeEvent) { m.count++; m.status = "Menu action: " + e.SelectedID }},
	}}
}

func (m *macComponentsGallery) splitPreview() ui.Node {
	if len(m.splits) == 0 {
		for i, small := range []bool{false, true} {
			b := &ui.SplitButton{ID: fmt.Sprintf("components/split/%d", i), Label: "Save", Small: small, IconImage: m.artwork[ui.SymbolSave], Options: []ui.Option{{ID: "save-as", Label: "Save As…", IconImage: m.artwork[ui.SymbolFolder]}, {ID: "formats", Label: "Export", Branch: m.menu}}, Primary: func() { m.count++; m.status = "Primary Save action" }, Change: func(e ui.ChangeEvent) { m.count++; m.status = "Split option: " + e.SelectedID }}
			m.splits = append(m.splits, b)
		}
	}
	children := []ui.Node{{ID: "components/split/help", Kind: ui.Label, Text: "Regular and small sizes. Save runs the primary action; the adjacent arrow opens options."}}
	for _, b := range m.splits {
		n, err := b.View()
		if err != nil {
			children = append(children, ui.Node{ID: b.ID + "/error", Kind: ui.Label, Text: err.Error()})
			continue
		}
		children = append(children, n)
	}
	return ui.Node{ID: "components/split-demo", Kind: ui.FormGroup, Text: "Split buttons", Children: children}
}

func (m *macComponentsGallery) inspectorPreview() ui.Node {
	if m.inspector == nil {
		m.width = 120
		m.inspector = &ui.Inspector{ID: "components/inspector", Title: "Object properties", IconImage: m.artwork[ui.SymbolFolder], Sections: []ui.InspectorSection{
			{ID: "geometry", Title: "Geometry", Expanded: true, Properties: []ui.InspectorProperty{
				{ID: "name", Label: "Name", Editor: ui.Node{ID: "components/inspector-name", Kind: ui.TextField, Text: m.name, Change: func(e ui.ChangeEvent) { m.name = e.Text }}},
				{ID: "width", Label: "Width", Mixed: true, Editor: ui.Node{ID: "components/inspector-width", Kind: ui.NumberField, Number: m.width, Numeric: ui.Numeric{Min: 1, Max: 1000, Step: 1}, NumberButtons: ui.NumberButtonsStepper, Change: func(e ui.ChangeEvent) { m.width = e.Number }}},
			}},
			{ID: "appearance", Title: "Appearance", Expanded: true, Properties: []ui.InspectorProperty{{ID: "visible", Label: "Visibility", Editor: ui.Node{ID: "components/inspector-visible", Kind: ui.Checkbox, Text: "Visible", Checked: true, Change: func(e ui.ChangeEvent) { m.inspector.Sections[1].Properties[0].Editor.Checked = e.Checked }}}}},
		}, Disclosure: func(e ui.InspectorDisclosureEvent) {
			for i := range m.inspector.Sections {
				if m.inspector.Sections[i].ID == e.SectionID {
					m.inspector.Sections[i].Expanded = e.Expanded
				}
			}
		}}
	}
	m.inspector.Sections[0].Properties[0].Editor.Text = m.name
	m.inspector.Sections[0].Properties[1].Editor.Number = m.width
	n, err := m.inspector.View()
	if err != nil {
		return ui.Node{ID: "components/inspector-error", Kind: ui.Label, Text: err.Error()}
	}
	return n
}
