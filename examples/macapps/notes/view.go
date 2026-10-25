package notes

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/ui"
)

const notesFont = "Helvetica Neue, emoji"

var (
	ink     = color.NRGBA{R: 35, G: 35, B: 37, A: 255}
	muted   = color.NRGBA{R: 133, G: 133, B: 136, A: 255}
	white   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	sidebar = color.NRGBA{R: 246, G: 246, B: 247, A: 255}
	line    = color.NRGBA{R: 229, G: 229, B: 231, A: 255}
	amber   = color.NRGBA{R: 255, G: 227, B: 146, A: 255}
)

// Theme uses the shared Mac control metrics; the app's containers own spacing.
func Theme() ui.Theme {
	theme := ui.MacTheme(false)
	theme.Default.Typography = ui.Some(ui.Typography{Family: notesFont, Size: 13})
	theme.Kinds[ui.Row] = ui.Style{Padding: ui.Some(ui.Insets{}), Gap: ui.Some(float32(0)), Align: ui.Some(ui.AlignCenter)}
	theme.Kinds[ui.Column] = ui.Style{Padding: ui.Some(ui.Insets{}), Gap: ui.Some(float32(0))}
	return theme
}
func label(id, text string, size float32, weight int) ui.Node {
	return ui.Node{ID: id, Kind: ui.Label, Text: text, Style: ui.Style{Padding: ui.Some(ui.Insets{}), Typography: ui.Some(ui.Typography{Family: notesFont, Size: size, Weight: weight})}}
}
func box(kind ui.Kind, id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: kind, Children: children}
}
func spacer(id string) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}}
}
func action(id, name string, symbol ui.Symbol, f func()) ui.Node {
	return ui.Node{ID: id, Kind: ui.Button, Icon: symbol, Activate: f, Appearance: ui.ButtonTool, Accessibility: ui.Accessibility{Name: name}, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(6)), MinSize: ui.Some(ui.Size{Width: 16, Height: 16}), MaxSize: ui.Some(ui.Size{Height: 16}), Background: ui.Some(color.NRGBA{}), CornerRadius: ui.Some(float32(7)), Border: ui.Some(ui.Border{})}}
}

// View returns a fresh tree. Each note retains its own public editor model and
// editor identity, so selection never reuses a native binding for another note.
func (m *Model) View() ui.Node {
	content := box(ui.Row, "notes-content", m.sidebarView(), divider("folders-divider"), m.listView(), divider("list-divider"), m.documentView())
	content.Grow = true
	root := box(ui.Column, "notes-root", content)
	if m.files != nil {
		root.Children = append(root.Children, m.files.promptView())
	}
	root.Grow = true
	root.Style.Background = ui.Some(white)
	// A removed note cannot leave a stale focus target in the admitted tree.
	if m.focus.ID != "" && m.focus.ID == "note-title-"+m.selected {
		root.Focus = m.focus
	}
	return root
}
func divider(id string) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 1}), MaxSize: ui.Some(ui.Size{Width: 1}), Background: ui.Some(line), GrowY: ui.Some(float32(1))}}
}
func (m *Model) folderCount(id string) int {
	n := 0
	for _, note := range m.notes {
		if id == "all" || note.value.FolderID == id {
			n++
		}
	}
	return n
}
func (m *Model) folderName() string {
	if m.selectedFolder == "all" {
		return "All Notes"
	}
	for _, f := range m.folders {
		if f.id == m.selectedFolder {
			return f.name
		}
	}
	return "Notes"
}
func (m *Model) sidebarView() ui.Node {
	heading := label("folders-heading", "On My Mac", 12, 600)
	heading.Style.Foreground = ui.Some(muted)
	heading.Style.Padding = ui.Some(ui.Insets{Left: 9, Top: 14, Bottom: 6})
	heading.Style.GrowX = ui.Some(float32(1))
	children := []ui.Node{heading}
	folders := append([]folder{{"all", "All Notes"}}, m.folders...)
	for _, f := range folders {
		f := f
		n := ui.Node{ID: "folder-" + f.id, Kind: ui.Button, Text: f.name, Icon: ui.SymbolFolder, Appearance: ui.ButtonQuiet, Disabled: m.actionReady() != nil, Activate: func() { _ = m.Request(func() error { return m.SelectFolder(f.id) }) }, Accessibility: ui.Accessibility{Name: fmt.Sprintf("%s folder, %d notes", f.name, m.folderCount(f.id))}, Style: ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)), TextAlign: ui.Some(ui.AlignStart), Padding: ui.Some(ui.Insets{Top: 4, Right: 4, Bottom: 4, Left: 8}), MinSize: ui.Some(ui.Size{Height: 30}), MaxSize: ui.Some(ui.Size{Height: 30}), Background: ui.Some(color.NRGBA{}), Border: ui.Some(ui.Border{})}}
		count := label("folder-count-"+f.id, fmt.Sprint(m.folderCount(f.id)), 13, 400)
		count.Style.Foreground = ui.Some(muted)
		count.Style.TextAlign = ui.Some(ui.AlignEnd)
		count.Style.MinSize = ui.Some(ui.Size{Width: 24})
		count.Style.MaxSize = ui.Some(ui.Size{Width: 24})
		countBox := box(ui.Column, "folder-count-box-"+f.id, count)
		countBox.Style = ui.Style{MinSize: ui.Some(ui.Size{Width: 24}), MaxSize: ui.Some(ui.Size{Width: 24}), GrowX: ui.Some(float32(0))}
		row := box(ui.Row, "folder-row-"+f.id, n, spacer("folder-space-"+f.id), countBox)
		row.Style = ui.Style{GrowX: ui.Some(float32(1)), MinSize: ui.Some(ui.Size{Height: 30}), MaxSize: ui.Some(ui.Size{Height: 30}), CornerRadius: ui.Some(float32(6))}
		if f.id == m.selectedFolder {
			row.Style.Background = ui.Some(color.NRGBA{R: 235, G: 234, B: 231, A: 255})
			n.Style.Foreground = ui.Some(color.NRGBA{R: 176, G: 117, B: 0, A: 255})
			row.Children[0] = n
		}
		children = append(children, row)
	}
	space := box(ui.Column, "folders-space")
	space.Style.GrowY = ui.Some(float32(1))
	children = append(children, space)
	footerText := "In-memory sample notes"
	if m.files != nil {
		footerText = "Local notebook"
	}
	footer := label("notes-demo-label", footerText, 11, 400)
	footer.Style.Foreground = ui.Some(muted)
	children = append(children, footer)
	n := box(ui.Column, "notes-folders", children...)
	n.Style = ui.Style{Padding: ui.Some(ui.Insets{Left: 10, Right: 10, Bottom: 12}), Gap: ui.Some(float32(2)), Background: ui.Some(sidebar), MinSize: ui.Some(ui.Size{Width: 190}), MaxSize: ui.Some(ui.Size{Width: 190}), GrowY: ui.Some(float32(1))}
	return n
}
func preview(body string) string {
	body = strings.Join(strings.Fields(body), " ")
	if utf8.RuneCountInString(body) > 46 {
		return string([]rune(body)[:46]) + "…"
	}
	if body == "" {
		return "No additional text"
	}
	return body
}
func displayTitle(title string) string {
	if strings.TrimSpace(title) == "" {
		return "New Note"
	}
	r := []rune(title)
	if len(r) > 36 {
		return string(r[:36]) + "…"
	}
	return title
}
func (m *Model) listView() ui.Node {
	header := box(ui.Column, "notes-list-header", label("notes-folder-name", m.folderName(), 15, 600), label("notes-folder-count", fmt.Sprintf("%d notes", m.folderCount(m.selectedFolder)), 11, 400))
	header.Style = ui.Style{Padding: ui.Some(ui.Insets{Left: 16, Top: 12, Right: 14, Bottom: 10}), Gap: ui.Some(float32(2)), GrowX: ui.Some(float32(1))}
	for i := range header.Children {
		header.Children[i].Style.GrowX = ui.Some(float32(1))
	}
	search := ui.Node{ID: "notes-search", Kind: ui.TextField, Placeholder: "Search notes", Disabled: m.files != nil && m.files.blocked(), Text: m.query, Change: func(e ui.ChangeEvent) { m.accept(m.SetSearch(e.Text)) }, Accessibility: ui.Accessibility{Name: "Search notes", Help: "Search titles and text in the selected folder"}, Style: ui.Style{GrowX: ui.Some(float32(1)), CornerRadius: ui.Some(float32(7)), MinSize: ui.Some(ui.Size{Height: 16}), MaxSize: ui.Some(ui.Size{Height: 16}), Padding: ui.Some(ui.Insets{Left: 8, Right: 8, Top: 4, Bottom: 4}), GrowY: ui.Some(float32(0)), Background: ui.Some(white)}}
	searchBox := box(ui.Column, "notes-search-box", search)
	searchBox.Style.Padding = ui.Some(ui.Insets{Left: 12, Right: 12, Bottom: 4})
	searchBox.Style.GrowX = ui.Some(float32(1))
	items := box(ui.Column, "notes-list-rows")
	items.Style.Padding = ui.Some(ui.Insets{Left: 8, Right: 8, Bottom: 12})
	items.Style.GrowX = ui.Some(float32(1))
	visible := m.Visible()
	lastGroup := ""
	for _, v := range visible {
		v := v
		group := "Today"
		if v.Pinned {
			group = "Pinned"
		}
		if group != lastGroup {
			h := label("notes-group-"+v.ID, group, 12, 600)
			h.Style.Padding = ui.Some(ui.Insets{Top: 14, Bottom: 7, Left: 8})
			h.Style.GrowX = ui.Some(float32(1))
			items.Children = append(items.Children, h)
			lastGroup = group
		}
		button := ui.Node{ID: "note-select-" + v.ID, Kind: ui.Button, Text: displayTitle(v.Title), Appearance: ui.ButtonQuiet, Disabled: m.actionReady() != nil, Activate: func() { _ = m.Request(func() error { return m.Select(v.ID) }) }, Accessibility: ui.Accessibility{Name: "Open note: " + displayTitle(v.Title)}, Style: ui.Style{Typography: ui.Some(ui.Typography{Family: notesFont, Size: 13, Weight: 600}), MinSize: ui.Some(ui.Size{Height: 16}), MaxSize: ui.Some(ui.Size{Height: 16}), Background: ui.Some(color.NRGBA{}), Padding: ui.Some(ui.Insets{}), TextAlign: ui.Some(ui.AlignStart), GrowX: ui.Some(float32(1)), Border: ui.Some(ui.Border{})}}
		summary := label("note-preview-"+v.ID, preview(v.Body), 12, 400)
		summary.Style.Foreground = ui.Some(muted)
		summary.Style.GrowX = ui.Some(float32(1))
		dateText := "10/19/26"
		if m.files != nil {
			dateText = "Local note"
		}
		date := label("note-date-"+v.ID, dateText, 11, 400)
		date.Style.Foreground = ui.Some(muted)
		date.Style.GrowX = ui.Some(float32(1))
		card := box(ui.Column, "note-card-"+v.ID, button, date, summary)
		card.Style = ui.Style{Padding: ui.Some(ui.Insets{Top: 6, Right: 12, Bottom: 6, Left: 16}), Gap: ui.Some(float32(1)), CornerRadius: ui.Some(float32(8)), GrowX: ui.Some(float32(1))}
		if v.ID == m.selected {
			card.Style.Background = ui.Some(amber)
		}
		items.Children = append(items.Children, card, ui.Node{ID: "note-line-" + v.ID, Kind: ui.Column, Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 1}), MaxSize: ui.Some(ui.Size{Height: 1}), GrowX: ui.Some(float32(1)), Background: ui.Some(line)}})
	}
	if len(visible) == 0 {
		empty := label("notes-empty-search", "No notes found", 13, 400)
		empty.Style.Padding = ui.Some(ui.UniformInsets(16))
		items.Children = append(items.Children, empty)
	}
	scroll := ui.Node{ID: "notes-list-scroll", Kind: ui.Scroll, Grow: true, Children: []ui.Node{items}}
	n := box(ui.Column, "notes-list", header, searchBox, scroll)
	n.Style = ui.Style{MinSize: ui.Some(ui.Size{Width: 320}), MaxSize: ui.Some(ui.Size{Width: 320}), GrowY: ui.Some(float32(1)), Background: ui.Some(white)}
	return n
}
func (m *Model) documentView() ui.Node {
	selected := m.lookup(m.selected)
	busy := m.actionReady() != nil
	newNote := action("notes-new", "Create a note", ui.SymbolAdd, func() { _ = m.Request(func() error { _, err := m.Add(); return err }) })
	newNote.Disabled = busy || len(m.notes) >= MaxNotes
	deleteNote := action("notes-delete", "Delete selected note", ui.SymbolDelete, func() { _ = m.Request(func() error { return m.Delete(m.selected) }) })
	deleteNote.Disabled = busy || selected == nil
	undo := action("notes-undo", "Undo", ui.SymbolUndo, func() { _ = m.Request(m.Undo) })
	undo.Disabled = busy || m.settlement == nil && m.History().UndoCount == 0
	redo := action("notes-redo", "Redo", ui.SymbolRedo, func() { _ = m.Request(m.Redo) })
	redo.Disabled = busy || m.settlement == nil && m.History().RedoCount == 0
	toolbar := box(ui.Row, "notes-toolbar", deleteNote, newNote, spacer("notes-tool-space"), undo, redo)
	if m.files != nil {
		for _, v := range []struct {
			id, label string
			icon      ui.Symbol
			run       func() error
		}{{"open", "Open Notebook…", ui.SymbolOpen, m.files.Open}, {"save", "Save Notebook", ui.SymbolSave, func() error { return m.files.Save(false) }}, {"save-as", "Save Notebook As…", ui.SymbolSave, func() error { return m.files.Save(true) }}} {
			v := v
			n := action("notes-file-"+v.id, v.label, v.icon, func() { _ = m.Request(v.run) })
			n.Disabled = busy
			toolbar.Children = append(toolbar.Children, n)
		}
	}
	toolbar.Style = ui.Style{Padding: ui.Some(ui.Insets{Left: 18, Right: 18, Top: 7, Bottom: 7}), Gap: ui.Some(float32(6)), Align: ui.Some(ui.AlignCenter), GrowX: ui.Some(float32(1))}
	if selected != nil {
		id := selected.value.ID
		pin := ui.Node{ID: "notes-pin", Kind: ui.ToggleButton, Text: "Pin", Checked: selected.value.Pinned, Disabled: busy, Change: func(e ui.ChangeEvent) { _ = m.Request(func() error { return m.SetPinned(id, e.Checked) }) }, Accessibility: ui.Accessibility{Name: "Pin selected note"}, Style: ui.Style{Padding: ui.Some(ui.Insets{Left: 9, Right: 9, Top: 5, Bottom: 5}), CornerRadius: ui.Some(float32(7)), MinSize: ui.Some(ui.Size{Height: 30}), MaxSize: ui.Some(ui.Size{Height: 30}), Border: ui.Some(ui.Border{})}}
		toolbar.Children = append(toolbar.Children, pin)
	}
	content := box(ui.Column, "notes-document", toolbar)
	content.Style = ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(1)), MinSize: ui.Some(ui.Size{Width: 340}), Background: ui.Some(white)}
	if selected == nil {
		empty := label("notes-empty-document", "No Note Selected", 22, 500)
		empty.Style = empty.Style.Overlay(ui.Style{Foreground: ui.Some(muted), TextAlign: ui.Some(ui.AlignCenter), GrowX: ui.Some(float32(1)), Padding: ui.Some(ui.Insets{Top: 100})})
		content.Children = append(content.Children, empty)
	} else {
		id := selected.value.ID
		dateText := "October 19, 2026 at 9:41 AM"
		if m.files != nil {
			dateText = "Local plain-text notebook"
		}
		date := label("notes-edited-date", dateText, 12, 400)
		date.Style.Foreground = ui.Some(muted)
		date.Style.TextAlign = ui.Some(ui.AlignCenter)
		date.Style.GrowX = ui.Some(float32(1))
		title := ui.Node{ID: "note-title-" + id, Kind: ui.TextField, Disabled: m.files != nil && m.files.blocked(), Text: selected.value.Title, Change: func(e ui.ChangeEvent) {
			m.accept(m.titleChanged(id, e.Text))
		}, Accessibility: ui.Accessibility{Name: "Note title"}, Style: ui.Style{Typography: ui.Some(ui.Typography{Family: notesFont, Size: 20, Weight: 600}), MinSize: ui.Some(ui.Size{Height: 30}), MaxSize: ui.Some(ui.Size{Height: 30}), CornerRadius: ui.Some(float32(0)), GrowY: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{}), Border: ui.Some(ui.Border{}), Background: ui.Some(white), GrowX: ui.Some(float32(1))}}
		bodyHelp := "Plain text editor; changes remain only in this demonstration"
		if m.files != nil {
			bodyHelp = "Plain text editor; Command-S saves the local notebook"
		}
		body := ui.Node{ID: "note-editor-" + id, Kind: ui.MultilineText, Disabled: m.files != nil && m.files.blocked(), Editor: selected.editor, Grow: true, SoftWrap: true, Prose: true, Change: func(ui.ChangeEvent) { m.accept(m.BodyChanged(id)) }, Accessibility: ui.Accessibility{Name: "Note text", Help: bodyHelp}, Style: ui.Style{Typography: ui.Some(ui.Typography{Family: notesFont, Size: 14}), Padding: ui.Some(ui.Insets{}), Border: ui.Some(ui.Border{}), Background: ui.Some(white)}}
		document := box(ui.Column, "note-paper", title, body)
		if m.files == nil {
			document.Children = append([]ui.Node{date}, document.Children...)
		}
		document.Style = ui.Style{Padding: ui.Some(ui.Insets{Left: 24, Right: 24, Top: 16, Bottom: 20}), Gap: ui.Some(float32(10)), GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(1))}
		content.Children = append(content.Children, document)
	}
	status := m.status
	if status == "" {
		status = "Sample notes · changes are not saved"
		if m.files != nil {
			status = m.files.status()
		}
	}
	fullStatus := status
	if strings.HasPrefix(status, "Saved · ") {
		status = "Saved"
	}
	if strings.HasPrefix(status, "Opened · ") {
		status = "Opened notebook"
	}
	footer := label("notes-status", status, 11, 400)
	footer.Accessibility = ui.Accessibility{Name: status, Help: fullStatus}
	footer.Style.Foreground = ui.Some(muted)
	footer.Style.Padding = ui.Some(ui.Insets{Left: 24, Right: 16, Top: 6, Bottom: 10})
	content.Children = append(content.Children, footer)
	return content
}
