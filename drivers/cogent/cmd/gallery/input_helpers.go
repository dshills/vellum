package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/dshills/vellum/ui"
)

type inputHelpersGallery struct {
	combo            *ui.ComboBox
	search           *ui.SearchBox
	notice           *ui.InlineNotice
	name             string
	requests         int
	quantity         float32
	reject, disabled bool
	dark             bool
}

func newInputHelpersGallery() *inputHelpersGallery {
	m := &inputHelpersGallery{name: "Untitled", quantity: 1}
	m.combo = &ui.ComboBox{ID: "unit-combo", Label: "Drawing units", Text: "Millimeters", Options: []ui.Option{{ID: "mm", Label: "Millimeters"}, {ID: "cm", Label: "Centimeters"}, {ID: "in", Label: "Inches"}}}
	m.combo.Change = func(e ui.ComboEvent) { m.combo.Text = e.Text }
	m.search = &ui.SearchBox{ID: "object-search", Label: "Search objects", Help: "Filter the component names below"}
	m.search.Change = func(e ui.SearchEvent) {
		m.requests++
		if !m.reject {
			m.search.Text = e.Text
		}
	}
	m.notice = &ui.InlineNotice{ID: "input-notice", Tone: ui.NoticeInformation, Text: "These widgets use application-owned values.", ActionLabel: "Reset fields", Activate: func() { m.name = "Untitled"; m.search.Text = ""; m.quantity = 1 }}
	m.setAppearance(false)
	return m
}
func (m *inputHelpersGallery) setDisabled(disabled bool) {
	m.disabled = disabled
	m.combo.Disabled = disabled
	m.search.Disabled = disabled
	m.notice.Disabled = disabled
}
func (m *inputHelpersGallery) setAppearance(dark bool) {
	m.dark = dark
	theme := ui.MacTheme(m.dark)
	m.combo.Style = theme.Kinds[ui.TextField]
	m.combo.Style.Padding = ui.Some(ui.Insets{})
	m.combo.Style.MaxSize = ui.Some(ui.Size{Width: 360})
	m.search.Style = theme.Kinds[ui.TextField]
	m.search.Style.MaxSize = ui.Some(ui.Size{Width: 360})
	m.search.Style.CornerRadius = ui.Some(float32(8))
	m.search.Style.Gap = ui.Some(float32(2))
	m.search.Style.Padding = ui.Some(ui.Insets{Top: 1, Right: 4, Bottom: 1, Left: 6})
	m.search.InputStyle = ui.Style{Border: ui.Some(ui.Border{}), Background: ui.Some(color.NRGBA{}), Padding: ui.Some(ui.Insets{Top: 2, Right: 2, Bottom: 2, Left: 2})}
	m.notice.Style = ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)), Padding: ui.Some(ui.UniformInsets(12)), Gap: ui.Some(float32(8)), Background: theme.Kinds[ui.FormGroup].Background, CornerRadius: ui.Some(float32(8))}
	m.search.ClearStyle = ui.Style{Border: ui.Some(ui.Border{}), Background: ui.Some(color.NRGBA{}), Padding: ui.Some(ui.UniformInsets(2))}
}
func (m *inputHelpersGallery) view() ui.Node {
	combo, err := m.combo.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	search, err := m.search.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: "Cannot render widgets: " + err.Error()}
	}
	errorText := ""
	if strings.TrimSpace(m.name) == "" {
		errorText = "Enter a drawing name."
	}
	field, err := (ui.FormField{ID: "drawing-name-field", Label: "Drawing name", Required: true, Help: "A name for your drawing.", ValidationMessage: errorText, Disabled: m.disabled,
		Content: ui.Node{ID: "drawing-name", Kind: ui.TextField, Text: m.name, Change: func(e ui.ChangeEvent) { m.name = e.Text }}}).View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: "Cannot render widgets: " + err.Error()}
	}
	notice, err := m.notice.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: "Cannot render widgets: " + err.Error()}
	}
	matches := []ui.Node{}
	for i, name := range []string{"Block", "Connector", "Note", "Rounded rectangle"} {
		if strings.Contains(strings.ToLower(name), strings.ToLower(m.search.Text)) {
			matches = append(matches, ui.Node{ID: fmt.Sprintf("match-%d", i), Kind: ui.Label, Text: name})
		}
	}
	if len(matches) == 0 {
		matches = append(matches, ui.Node{ID: "no-matches", Kind: ui.Label, Text: "No matching components"})
	}
	background := color.NRGBA{R: 246, G: 246, B: 248, A: 255}
	if m.dark {
		background = color.NRGBA{R: 29, G: 29, B: 31, A: 255}
	}
	return ui.Node{ID: "inputs", Kind: ui.Scroll, Grow: true, Style: ui.Style{Background: ui.Some(background), Padding: ui.Some(ui.UniformInsets(20)), Gap: ui.Some(float32(16))}, Children: []ui.Node{
		{ID: "inputs-title", Kind: ui.Label, Text: "Mac input widgets", Style: ui.Style{Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 20, Weight: 600})}},
		{ID: "inputs-description", Kind: ui.Label, Text: "Search, validation and inline feedback through public Go APIs."},
		search,
		{ID: "combo-group", Kind: ui.FormGroup, Text: "Drawing units — type a value or choose", Children: []ui.Node{combo}},
		{ID: "search-results", Kind: ui.FormGroup, Text: "Components", Children: matches},
		field,
		{ID: "quantity-field", Kind: ui.FormGroup, Text: "Copies", Children: []ui.Node{
			{ID: "quantity", Kind: ui.NumberField, NumberButtons: ui.NumberButtonsStepper, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 88, Height: 28}), MaxSize: ui.Some(ui.Size{Width: 104})}, Number: m.quantity, Numeric: ui.Numeric{Min: 0, Max: 1000, Step: 1}, Accessibility: ui.Accessibility{Name: "Copies"}, Disabled: m.disabled, Change: func(e ui.ChangeEvent) { m.quantity = e.Number }},
		}},
		notice,
		{ID: "input-requests", Kind: ui.Label, Text: fmt.Sprintf("Search requests: %d", m.requests)},
		{ID: "input-options", Kind: ui.Row, Children: []ui.Node{
			{ID: "input-reject", Kind: ui.Checkbox, Text: "Reject search changes", Checked: m.reject, Change: func(e ui.ChangeEvent) { m.reject = e.Checked }},
			{ID: "input-disabled", Kind: ui.Checkbox, Text: "Disable inputs", Checked: m.disabled, Change: func(e ui.ChangeEvent) { m.setDisabled(e.Checked) }},
		}},
	}}
}
