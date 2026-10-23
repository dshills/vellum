package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/dshills/vellum/ui"
)

type inputHelpersGallery struct {
	secureField      *ui.SecureField
	pathControl      *ui.PathControl
	fontPicker       *ui.FontPicker
	fontError        string
	tokens           *ui.TokenField
	tokenNext        int
	tokenError       string
	colorWell        *ui.ColorWell
	colorError       string
	datePicker       *ui.DateTimePicker
	dateError        string
	selectedObjects  []string
	rejectSelection  bool
	combo            *ui.ComboBox
	search           *ui.SearchBox
	notice           *ui.InlineNotice
	name             string
	requests         int
	contextAction    string
	contextChecked   bool
	popoverOpen      bool
	quantity         float32
	reject, disabled bool
	dark             bool
}

func newInputHelpersGallery() *inputHelpersGallery {
	m := &inputHelpersGallery{}
	m.resetFields()
	return m
}

// Rebuild sample components on this owner so reset also closes private editor
// state. Revoke the old components first; queued callbacks must stay inert.
func (m *inputHelpersGallery) resetFields() {
	dark, disabled := m.dark, m.disabled
	if m.search != nil {
		m.setDisabled(true)
	}
	*m = inputHelpersGallery{name: "Untitled", quantity: 1}
	m.secureField = &ui.SecureField{ID: "demo-password", Label: "Demo password", Text: "Vellum-demo", Placeholder: "Enter demo text", AllowReveal: true}
	m.secureField.Change = func(e ui.ChangeEvent) { m.secureField.Text = e.Text }
	m.pathControl = &ui.PathControl{ID: "drawing-path", Label: "Drawing location", MaxVisible: 3, MaxLabelRunes: 12, Items: []ui.PathItem{{ID: "root", Label: "Home"}, {ID: "documents", Label: "Documents"}, {ID: "projects", Label: "Projects"}, {ID: "drawings", Label: "Drawings"}, {ID: "drawing", Label: "Untitled.vellum", Kind: ui.PathFile}}}
	m.pathControl.Navigate = func(e ui.PathEvent) {
		for i, item := range m.pathControl.Items {
			if item.ID == e.ItemID {
				m.pathControl.Items = append([]ui.PathItem(nil), m.pathControl.Items[:i+1]...)
				break
			}
		}
	}
	m.fontPicker = &ui.FontPicker{ID: "drawing-font", Label: "Drawing font", Preview: "A drawing starts with an idea.", Value: ui.FontSelection{FamilyID: "helvetica", TypefaceID: "regular", Size: 16}, Families: []ui.FontFamily{{ID: "helvetica", Name: "Helvetica Neue", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}, {ID: "bold", Name: "Bold", Weight: 700}, {ID: "italic", Name: "Italic", Weight: 400, Italic: true}}}, {ID: "times", Name: "Times New Roman", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}, {ID: "bold", Name: "Bold", Weight: 700}}}, {ID: "menlo", Name: "Menlo", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}}}
	m.fontPicker.Change = func(e ui.FontSelectionEvent) { m.fontPicker.Value = e.Apply(m.fontPicker.Value); m.fontError = "" }
	m.fontPicker.Invalid = func(err error) { m.fontError = err.Error() }
	m.tokens = &ui.TokenField{ID: "drawing-tags", Label: "Drawing tags", Tokens: []ui.Token{{ID: "tag-1", Text: "Design"}}, Suggestions: []ui.Option{{ID: "work", Label: "Work"}, {ID: "personal", Label: "Personal"}, {ID: "review", Label: "Review"}}}
	m.tokenNext = 1
	m.tokens.Change = func(e ui.TokenEvent) {
		if e.Remove {
			for i, t := range m.tokens.Tokens {
				if t.ID == e.TokenID {
					m.tokens.Tokens = append(m.tokens.Tokens[:i], m.tokens.Tokens[i+1:]...)
					break
				}
			}
		} else {
			m.tokenNext++
			m.tokens.Tokens = append(m.tokens.Tokens, ui.Token{ID: fmt.Sprintf("tag-%d", m.tokenNext), Text: e.Text})
			m.tokens.Draft = ""
		}
		m.tokenError = ""
	}
	m.tokens.Invalid = func(err error) { m.tokenError = err.Error() }
	m.colorWell = &ui.ColorWell{ID: "drawing-color", Label: "Drawing color", Color: color.NRGBA{R: 52, G: 120, B: 246, A: 255}, Recent: []color.NRGBA{{R: 52, G: 120, B: 246, A: 255}, {R: 237, G: 90, B: 80, A: 255}, {R: 64, G: 180, B: 110, A: 255}}}
	m.colorWell.Change = func(e ui.ColorEdit) { m.colorWell.Color = e.Apply(m.colorWell.Color); m.colorError = "" }
	m.colorWell.Invalid = func(err error) { m.colorError = err.Error() }

	m.datePicker = &ui.DateTimePicker{ID: "reminder-date", Label: "Reminder date and time", Mode: ui.DateAndTime, Value: time.Date(2026, 10, 23, 9, 0, 0, 0, time.Local)}
	m.datePicker.Change = func(e ui.DateTimeEvent) { m.datePicker.Value = e.Value; m.dateError = "" }
	m.datePicker.Invalid = func(err error) { m.dateError = err.Error() }
	m.combo = &ui.ComboBox{ID: "unit-combo", Label: "Drawing units", Text: "Millimeters", Options: []ui.Option{{ID: "mm", Label: "Millimeters"}, {ID: "cm", Label: "Centimeters"}, {ID: "in", Label: "Inches"}}}
	m.combo.Change = func(e ui.ComboEvent) { m.combo.Text = e.Text }
	m.search = &ui.SearchBox{ID: "object-search", Label: "Search objects", Help: "Filter the component names below"}
	m.search.Change = func(e ui.SearchEvent) {
		m.requests++
		if !m.reject {
			m.search.Text = e.Text
		}
	}
	m.notice = &ui.InlineNotice{ID: "input-notice", Tone: ui.NoticeInformation, Text: "These widgets use application-owned values.", ActionLabel: "Reset fields", Activate: m.resetFields}
	m.setAppearance(dark)
	m.setDisabled(disabled)
}
func (m *inputHelpersGallery) setDisabled(disabled bool) {
	m.disabled = disabled
	m.colorWell.Disabled = disabled
	m.tokens.Disabled = disabled
	m.fontPicker.Disabled = disabled
	m.pathControl.Disabled = disabled
	m.secureField.Disabled = disabled
	m.combo.Disabled = disabled
	m.datePicker.Disabled = disabled
	m.search.Disabled = disabled
	m.notice.Disabled = disabled
}
func (m *inputHelpersGallery) setAppearance(dark bool) {
	m.dark = dark
	theme := ui.MacTheme(m.dark)
	m.colorWell.PopoverStyle = theme.Kinds[ui.FormGroup]
	m.secureField.Style = theme.Kinds[ui.TextField]
	m.secureField.Style.Padding = ui.Some(ui.Insets{})
	m.secureField.Style.MaxSize = ui.Some(ui.Size{Width: 360})
	m.fontPicker.FieldStyle = theme.Kinds[ui.TextField]
	m.tokens.InputStyle = theme.Kinds[ui.TextField]
	m.tokens.InputStyle.Padding = ui.Some(ui.Insets{})
	m.tokens.InputStyle.MaxSize = ui.Some(ui.Size{Width: 360})
	m.datePicker.Style = theme.Kinds[ui.TextField]
	m.datePicker.Style.MinSize = ui.Some(ui.Size{Width: 180, Height: 28})
	m.datePicker.Style.MaxSize = ui.Some(ui.Size{Width: 240})
	m.datePicker.PopoverStyle = theme.Kinds[ui.FormGroup]
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
	secure, err := m.secureField.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	path, err := m.pathControl.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	font, err := m.fontPicker.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	tokens, err := m.tokens.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	m.colorWell.ValidationMessage = m.colorError
	well, err := m.colorWell.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}

	date, err := m.datePicker.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	combo, err := m.combo.View()
	if err != nil {
		return ui.Node{ID: "inputs", Kind: ui.Label, Text: err.Error()}
	}
	popover, err := (ui.Popover{ID: "quick-inspector", AnchorID: "inspector-anchor", Label: "Quick inspector", Open: m.popoverOpen, Disabled: m.disabled, Dismiss: func() { m.popoverOpen = false }, Style: ui.MacTheme(m.dark).Kinds[ui.FormGroup], Children: []ui.Node{{ID: "inspector-label", Kind: ui.Label, Text: "Drawing inspector"}, {ID: "inspector-close", Kind: ui.Button, Text: "Done", Activate: func() { m.popoverOpen = false }}}}).View()
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
		{ID: "inputs-description", Kind: ui.Label, Text: "Mac-style colors, tags, fonts, paths and protected text through public Go APIs."},
		{ID: "secure-group", Kind: ui.FormGroup, Text: "Secure entry — synthetic demo text", Children: []ui.Node{secure}},
		{ID: "path-group", Kind: ui.FormGroup, Text: "Location — ancestor menus and overflow", Children: []ui.Node{path}},
		{ID: "font-group", Kind: ui.FormGroup, Text: "Font", Children: []ui.Node{font, {ID: "font-feedback", Kind: ui.Label, Text: m.fontError}}},
		{ID: "tokens-group", Kind: ui.FormGroup, Text: "Tags — Return adds, chip button removes", Children: []ui.Node{tokens, {ID: "token-feedback", Kind: ui.Label, Text: m.tokenError}}},
		{ID: "color-well-group", Kind: ui.FormGroup, Text: "Drawing color", Children: []ui.Node{well}},
		search,
		{ID: "combo-group", Kind: ui.FormGroup, Text: "Drawing units — type a value or choose", Children: []ui.Node{combo}},
		{ID: "search-results", Kind: ui.FormGroup, Text: "Components", Children: matches},
		field,
		{ID: "quantity-field", Kind: ui.FormGroup, Text: "Copies", Children: []ui.Node{
			{ID: "quantity", Kind: ui.NumberField, NumberButtons: ui.NumberButtonsStepper, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 88, Height: 28}), MaxSize: ui.Some(ui.Size{Width: 104})}, Number: m.quantity, Numeric: ui.Numeric{Min: 0, Max: 1000, Step: 1}, Accessibility: ui.Accessibility{Name: "Copies"}, Disabled: m.disabled, Change: func(e ui.ChangeEvent) { m.quantity = e.Number }},
		}},
		{ID: "context-demo", Kind: ui.Button, Text: "Right-click or Control-click for actions", Disabled: m.disabled, Accessibility: ui.Accessibility{Name: "Context menu demo", Help: "Shift-F10 also opens the menu"}, ContextMenu: &ui.ContextMenu{Options: []ui.Option{{ID: "inspect", Label: "Inspect"}, {ID: "sep", Separator: true}, {ID: "grid", Label: "Show grid", Checkable: true, Checked: m.contextChecked}}, Change: func(e ui.ChangeEvent) {
			m.contextAction = e.SelectedID
			if e.SelectedID == "grid" {
				m.contextChecked = e.Checked
			}
		}}},
		{ID: "context-result", Kind: ui.Label, Text: "Context action: " + m.contextAction},
		{ID: "inspector-anchor", Kind: ui.Button, Text: "Quick inspector…", Disabled: m.disabled, Activate: func() { m.popoverOpen = !m.popoverOpen }},
		popover,
		{ID: "source-group", Kind: ui.FormGroup, Text: "Objects — Command-click / Shift-click", Disabled: m.disabled, Children: []ui.Node{
			{ID: "object-list", Kind: ui.ListBox, Multiple: true, SelectedIDs: m.selectedObjects, Options: []ui.Option{{ID: "block", Label: "Block"}, {ID: "connector", Label: "Connector"}, {ID: "note", Label: "Note"}, {ID: "round", Label: "Rounded rectangle"}}, SelectionChange: func(e ui.ListSelectionEvent) {
				if !m.rejectSelection {
					m.selectedObjects = e.Apply(m.selectedObjects)
				}
			}, Style: ui.Style{MaxSize: ui.Some(ui.Size{Height: 140})}},
			{ID: "selected-objects", Kind: ui.Label, Text: "Selected: " + strings.Join(m.selectedObjects, ", ")},
			{ID: "reject-selection", Kind: ui.Checkbox, Text: "Reject selection edits", Checked: m.rejectSelection, Change: func(e ui.ChangeEvent) { m.rejectSelection = e.Checked }},
		}},
		{ID: "date-group", Kind: ui.FormGroup, Text: "Date and time", Children: []ui.Node{date, {ID: "date-error", Kind: ui.Label, Text: m.dateError}}},
		notice,
		{ID: "input-requests", Kind: ui.Label, Text: fmt.Sprintf("Search requests: %d", m.requests)},
		{ID: "input-options", Kind: ui.Row, Children: []ui.Node{
			{ID: "input-reject", Kind: ui.Checkbox, Text: "Reject search changes", Checked: m.reject, Change: func(e ui.ChangeEvent) { m.reject = e.Checked }},
			{ID: "input-disabled", Kind: ui.Checkbox, Text: "Disable inputs", Checked: m.disabled, Change: func(e ui.ChangeEvent) { m.setDisabled(e.Checked) }},
		}},
	}}
}
