package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image/color"
)

type macComponentsGallery struct {
	followSystem bool
	system       cogent.SystemAppearance

	ranges           *macRangeGallery
	gradient         *ui.GradientWell
	scopes           *macScopeGallery
	essentials       *macEssentialsGallery
	find             *macFindGallery
	preview          *macPreviewGallery
	documents        *macDocumentGallery
	rules            *ui.RuleEditor
	browser          *ui.ColumnBrowser
	inspector        *ui.Inspector
	width            float32
	splits           []*ui.SplitButton
	menuGrid         bool
	menu             *ui.MenuBranch
	dark             bool
	accent           string
	page             string
	name             string
	status           string
	count            int
	shareTicket      *app.Completion
	dropTicket       *app.Completion
	dropZone         *ui.FileDropZone
	commandPalette   *ui.CommandPalette
	checklist        *ui.Checklist
	filePicker       *ui.FilePickerField
	filePickerTicket *app.Completion
	dateRange        *ui.DateRangePicker
	pdfPath          string
	mediaPath        string
	mediaPlayer      *ui.MediaPlayer
	mediaTicket      *app.Completion
	pdfTicket        *app.Completion
	artwork          map[ui.Symbol]*canvas.TiledImage
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
	if m.followSystem && m.system.Known {
		return cogent.MacSystemTheme(m.system)
	}
	accent := color.NRGBA{G: 112, B: 245, A: 255}
	switch m.accent {
	case "purple":
		accent = color.NRGBA{R: 126, G: 65, B: 194, A: 255}
	case "green":
		accent = color.NRGBA{R: 25, G: 125, B: 80, A: 255}
	}
	if m.system.Known {
		v := m.system
		v.Dark = m.dark
		v.Accent = accent
		return cogent.MacSystemTheme(v)
	}
	return ui.MacThemeWithAccent(m.dark, accent)
}
func (m *macComponentsGallery) view() ui.Node {
	theme := m.theme()
	palette := theme.Palette.Value
	appearance := []ui.Node{}
	if m.system.Known {
		appearance = append(appearance, ui.Node{ID: "components/appearance/system", Kind: ui.Button, Text: "System", Activate: func() { m.followSystem = true; m.dark = m.system.Dark }})
	}
	for _, choice := range []struct {
		label string
		dark  bool
	}{{"Light", false}, {"Dark", true}} {
		style := ui.Style{GrowX: ui.Some(float32(0))}
		if m.dark == choice.dark {
			style.Background = ui.Some(palette.Accent)
			style.Foreground = ui.Some(palette.AccentInk)
		}
		appearance = append(appearance, ui.Node{ID: "components/appearance/" + choice.label, Kind: ui.Button, Text: choice.label, Style: style, Activate: func() { m.followSystem = false; m.dark = choice.dark }})
	}
	appearance = append(appearance, ui.Node{ID: "components/accent", Kind: ui.Selector, SelectedID: m.accent, Options: []ui.Option{{ID: "blue", Label: "Blue accent"}, {ID: "purple", Label: "Purple accent"}, {ID: "green", Label: "Green accent"}}, Accessibility: ui.Accessibility{Name: "Component accent"}, Change: func(e ui.ChangeEvent) {
		if e.SelectedID == "blue" || e.SelectedID == "purple" || e.SelectedID == "green" {
			m.followSystem = false
			m.accent = e.SelectedID
		}
	}})
	if m.essentials == nil {
		m.essentials = newMacEssentialsGallery(m)
	}
	content := m.themePreview()
	if m.page == "media-player" {
		content = m.mediaPlayerView()
	}
	if m.page == "pdf" {
		content = m.pdfViewerView()
	}
	if m.page == "date-range" {
		content = m.dateRangeView()
	}
	if m.page == "path-picker" {
		content = m.filePickerView()
	}
	if m.page == "checklist" {
		content = m.checklistView()
	}
	if m.page == "commands" {
		content = m.commandPaletteView()
	}
	if m.page == "file-drop" {
		content = m.fileDropZoneView()
	}
	if m.page == "sharing" {
		content = m.sharingView()
	}
	if m.page == "links" {
		content = m.linksView()
	}
	if m.page == "help-tags" {
		content = m.helpTagsView()
	}
	if m.page == "range-slider" {
		content = m.rangeView()
	}
	if m.page == "gradient" {
		content = m.gradientView()
	}
	if m.page == "search-scopes" {
		content = m.scopeView()
	}
	if m.page == "find-replace" {
		content = m.findView()
	}
	if m.page == "quick-look" {
		content = m.quickLookView()
	}
	if m.page == "angle-dial" {
		content = m.angleDialView()
	}
	if m.page == "native-fonts" {
		content = m.nativeFontView()
	}
	if m.page == "native-colors" {
		content = m.nativeColorView()
	}
	if m.page == "outline" {
		content = m.outlineView()
	}
	if m.page == "documents" {
		content = m.documentListView()
	}
	if m.page == "calendar" {
		content = m.essentials.calendarView()
	}
	if m.page == "alerts" {
		content = m.essentials.alertView()
	}
	if m.page == "settings" {
		content = m.essentials.settingsView()
	}
	if m.page == "icon-picker" {
		content = m.essentials.iconPickerView()
	}
	if m.page == "image-well" {
		content = m.essentials.imageWellView()
	}
	if m.page == "rules" {
		content = m.rulesPreview()
	}
	if m.page == "browser" {
		content = m.browserPreview()
	}
	if m.page == "inspector" {
		content = m.inspectorPreview()
	}
	if m.page == "split" {
		content = m.splitPreview()
	}
	if m.page == "menus" {
		content = m.menuPreview()
	}
	navigation := ui.Node{ID: "components/pages", Kind: ui.Selector, SelectedID: m.page, Options: []ui.Option{{ID: "media-player", Label: "Media player"}, {ID: "pdf", Label: "PDF viewer"}, {ID: "date-range", Label: "Date range"}, {ID: "path-picker", Label: "File / folder field"}, {ID: "checklist", Label: "Checklist"}, {ID: "commands", Label: "Command palette"}, {ID: "file-drop", Label: "File drop zone"}, {ID: "sharing", Label: "Native Share"}, {ID: "links", Label: "Links"}, {ID: "help-tags", Label: "Help tags"}, {ID: "range-slider", Label: "Range slider"}, {ID: "gradient", Label: "Gradient well"}, {ID: "search-scopes", Label: "Search scope bar"}, {ID: "find-replace", Label: "Find / Replace"}, {ID: "quick-look", Label: "Quick Look"}, {ID: "theme", Label: "Themes"}, {ID: "menus", Label: "Cascading menus"}, {ID: "split", Label: "Split buttons"}, {ID: "inspector", Label: "Inspector panel"}, {ID: "browser", Label: "Column browser"}, {ID: "rules", Label: "Filter / rule editor"}, {ID: "image-well", Label: "Image well"}, {ID: "icon-picker", Label: "Searchable icon picker"}, {ID: "settings", Label: "Settings pane"}, {ID: "alerts", Label: "Native alert"}, {ID: "calendar", Label: "Calendar picker"}, {ID: "documents", Label: "Document list"}, {ID: "outline", Label: "Mac outline"}, {ID: "native-colors", Label: "Native Colors panel"}, {ID: "native-fonts", Label: "Native Fonts panel"}, {ID: "angle-dial", Label: "Circular angle dial"}}, Accessibility: ui.Accessibility{Name: "Component demonstration"}, Change: func(e ui.ChangeEvent) { m.page = e.SelectedID }}

	return ui.Node{ID: "mac-components", Kind: ui.Scroll, Theme: &theme, Grow: true, Style: ui.Style{Background: ui.Some(palette.Window), Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}, Children: []ui.Node{{ID: "components/title", Kind: ui.Label, Text: "Mac components", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "components/appearance", Kind: ui.Row, Children: appearance}, navigation, content, {ID: "components/status", Kind: ui.Label, Text: fmt.Sprintf("%s · actions: %d", m.status, m.count)}}}
}
func (m *macComponentsGallery) themePreview() ui.Node {
	return ui.Node{ID: "components/theme-preview", Kind: ui.FormGroup, Text: "Reusable scoped themes", Children: []ui.Node{{ID: "components/theme-help", Kind: ui.Label, Text: "Switch appearance and accent. Controls retain their values and full-color artwork."}, {ID: "components/name", Kind: ui.TextField, Text: m.name, Accessibility: ui.Accessibility{Name: "Theme preview name"}, Change: func(e ui.ChangeEvent) { m.name = e.Text }}, {ID: "components/color-button", Kind: ui.Button, Text: "Save preview", IconImage: m.artwork[ui.SymbolSave], Style: ui.Style{GrowX: ui.Some(float32(0))}, Activate: func() { m.count++; m.status = "Theme preview saved" }}}}
}

func (m *macComponentsGallery) menuPreview() ui.Node {
	return ui.Node{ID: "components/menu-demo", Kind: ui.FormGroup, Text: "Cascading action menus", Children: []ui.Node{
		{ID: "components/menu-help", Kind: ui.Label, Text: "Open Export, then use Right / Left to navigate. Escape cancels. Shortcut captions do not install bindings."},
		{ID: "components/actions", Kind: ui.MenuButton, Text: "Document actions", IconImage: m.artwork[ui.SymbolFolder], Options: []ui.Option{{ID: "export", Label: "Export", IconImage: m.artwork[ui.SymbolSave], Branch: m.menu}, {ID: "separator", Separator: true}, {ID: "grid", Label: "Show grid", Checkable: true, Checked: m.menuGrid}}, Change: func(e ui.ChangeEvent) {
			m.count++
			m.status = "Menu action: " + e.SelectedID
			if e.SelectedID == "grid" {
				m.menuGrid = e.Checked
			}
		}},
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

func (m *macComponentsGallery) browserPreview() ui.Node {
	if m.browser == nil {
		m.browser = &ui.ColumnBrowser{ID: "components/browser", Label: "Component library", Fractions: map[string]float32{}, Items: []ui.BrowserItem{
			{ID: "drawings", Label: "Drawings", Thumbnail: m.artwork[ui.SymbolFolder], Children: []ui.BrowserItem{{ID: "floor", Label: "Floor plan", Thumbnail: m.artwork[ui.SymbolDocument]}, {ID: "detail", Label: "Detail drawing", Thumbnail: m.artwork[ui.SymbolSave]}}},
			{ID: "components", Label: "Components", Thumbnail: m.artwork[ui.SymbolFolder], Children: []ui.BrowserItem{{ID: "symbols", Label: "Symbols", Thumbnail: m.artwork[ui.SymbolFolder], Children: []ui.BrowserItem{{ID: "connector", Label: "Connector", Thumbnail: m.artwork[ui.SymbolDocument]}}}, {ID: "notes", Label: "Notes", Thumbnail: m.artwork[ui.SymbolDocument]}}},
			{ID: "cloud", Label: "Cloud (unavailable)", Disabled: true, Thumbnail: m.artwork[ui.SymbolFolder]},
		}, Change: func(e ui.BrowserEvent) {
			m.browser.Path = e.Path
			m.browser.Fractions = map[string]float32{}
			m.status = "Browser path: " + fmt.Sprint(e.Path)
		}, Resize: func(e ui.BrowserResizeEvent) {
			if m.browser.Fractions == nil {
				m.browser.Fractions = map[string]float32{}
			}
			m.browser.Fractions[e.ColumnID] = e.Fraction
		}}
	}
	label := "Choose an item"
	if len(m.browser.Path) > 0 {
		label = m.browser.Path[len(m.browser.Path)-1]
	}
	preview := ui.Node{ID: "components/browser-preview-content", Kind: ui.Column, Children: []ui.Node{{ID: "components/browser-preview-art", Kind: ui.Image, Text: "Document preview", Image: m.artwork[ui.SymbolDocument], Fit: ui.ImageContain, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 64, Height: 64}), MaxSize: ui.Some(ui.Size{Width: 64, Height: 64})}}, {ID: "components/browser-preview-name", Kind: ui.Label, Text: label}}}
	m.browser.Preview = &preview
	n, err := m.browser.View()
	if err != nil {
		return ui.Node{ID: "components/browser-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "components/browser-demo", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "components/browser-help", Kind: ui.Label, Text: "Select a folder to reveal its column. Tab moves through columns/dividers; drag a divider to resize. Scroll horizontally for deeper paths."}, n}}
}

func (m *macComponentsGallery) rulesPreview() ui.Node {
	if m.rules == nil {
		m.rules = &ui.RuleEditor{ID: "components/rules", Fields: []ui.RuleField{{ID: "name", Label: "Name", Kind: ui.RuleText}, {ID: "size", Label: "Size", Kind: ui.RuleNumber}, {ID: "visible", Label: "Visible", Kind: ui.RuleBoolean}}, Root: ui.Rule{ID: "root", Group: &ui.RuleGroup{Match: ui.RuleAll, Rules: []ui.Rule{
			{ID: "name-rule", FieldID: "name", Operator: ui.RuleContains, Value: ui.RuleValue{Text: "Drawing"}},
			{ID: "size-rule", FieldID: "size", Operator: ui.RuleGreater, Value: ui.RuleValue{Number: 10}},
			{ID: "nested", Group: &ui.RuleGroup{Match: ui.RuleAny, Rules: []ui.Rule{{ID: "visible-rule", FieldID: "visible", Operator: ui.RuleIs, Value: ui.RuleValue{Checked: true}}}}},
		}}}, AddIconImage: m.artwork[ui.SymbolAdd], GroupIconImage: m.artwork[ui.SymbolFolder], Change: func(e ui.RuleEditEvent) { m.rules.Root = e.Root; m.count++; m.status = "Rule updated: " + e.RuleID }}
	}
	n, err := m.rules.View()
	if err != nil {
		return ui.Node{ID: "components/rules-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "components/rules-demo", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "components/rules-help", Kind: ui.Label, Text: "Compose typed conditions and All/Any groups. Changes update the model; this demonstration does not execute a search."}, n}}
}
