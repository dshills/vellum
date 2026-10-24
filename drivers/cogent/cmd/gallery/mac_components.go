package main

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image/color"
)

type macComponentsGallery struct {
	dark    bool
	accent  string
	page    string
	name    string
	status  string
	count   int
	artwork map[ui.Symbol]*canvas.TiledImage
}

func newMacComponentsGallery(dark bool) *macComponentsGallery {
	artwork, err := toolbarColorIcons()
	if err != nil {
		panic(err)
	}
	return &macComponentsGallery{dark: dark, accent: "blue", page: "theme", name: "Untitled", status: "No action yet", artwork: artwork}
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
	return ui.Node{ID: "mac-components", Kind: ui.Scroll, Theme: &theme, Grow: true, Style: ui.Style{Background: ui.Some(palette.Window), Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}, Children: []ui.Node{{ID: "components/title", Kind: ui.Label, Text: "Mac components", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "components/appearance", Kind: ui.Row, Children: appearance}, content, {ID: "components/status", Kind: ui.Label, Text: fmt.Sprintf("%s · actions: %d", m.status, m.count)}}}
}
func (m *macComponentsGallery) themePreview() ui.Node {
	return ui.Node{ID: "components/theme-preview", Kind: ui.FormGroup, Text: "Reusable scoped themes", Children: []ui.Node{{ID: "components/theme-help", Kind: ui.Label, Text: "Switch appearance and accent. Controls retain their values and full-color artwork."}, {ID: "components/name", Kind: ui.TextField, Text: m.name, Accessibility: ui.Accessibility{Name: "Theme preview name"}, Change: func(e ui.ChangeEvent) { m.name = e.Text }}, {ID: "components/color-button", Kind: ui.Button, Text: "Save preview", IconImage: m.artwork[ui.SymbolSave], Style: ui.Style{GrowX: ui.Some(float32(0))}, Activate: func() { m.count++; m.status = "Theme preview saved" }}}}
}
