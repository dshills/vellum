package reminders

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/dshills/vellum/ui"
)

// Native demo configuration measured on macOS Tahoe 26.6.2 with unified chrome.
// This is an example-specific reserve, not a framework safe-area calculation.
const nativeHeaderReserve float32 = 52

var (
	ink     = color.NRGBA{R: 35, G: 35, B: 39, A: 255}
	muted   = color.NRGBA{R: 125, G: 125, B: 132, A: 255}
	blue    = color.NRGBA{R: 0, G: 122, B: 255, A: 255}
	orange  = color.NRGBA{R: 235, G: 137, B: 26, A: 255}
	white   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	sidebar = color.NRGBA{R: 239, G: 239, B: 242, A: 255}
)

// Theme is application-local. It never changes the system appearance.
func Theme() ui.Theme {
	return ui.Theme{Default: ui.Style{
		Foreground: ui.Some(ink), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 13}),
	}, Kinds: map[ui.Kind]ui.Style{
		ui.Row:    {Gap: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{})},
		ui.Column: {Gap: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{})},
		ui.Button: {CornerRadius: ui.Some(float32(6))},
	}}
}

func box(kind ui.Kind, id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: kind, Children: children, Style: ui.Style{Gap: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{}), GrowX: ui.Some(float32(1))}}
}
func label(id, text string, size float32, weight int, foreground color.NRGBA) ui.Node {
	return ui.Node{ID: id, Kind: ui.Label, Text: text, Style: ui.Style{Foreground: ui.Some(foreground), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: size, Weight: weight})}}
}
func button(id, text, name string, icon ui.Symbol, action func()) ui.Node {
	return ui.Node{ID: id, Kind: ui.Button, Text: text, Icon: icon, Appearance: ui.ButtonQuiet, Activate: action, Accessibility: ui.Accessibility{Name: name}}
}
func growSpacer(id string) ui.Node {
	n := box(ui.Column, id)
	n.Style.GrowX = ui.Some(float32(1))
	return n
}

// Numeric labels use a fixed column so their intrinsic wrapping measurement
// does not participate in the surrounding row's flexible-width allocation.
func countColumn(id string, width float32, child ui.Node) ui.Node {
	n := box(ui.Column, id, child)
	n.Style.GrowX = ui.Some(float32(0))
	n.Style.MinSize = ui.Some(ui.Size{Width: width})
	n.Style.MaxSize = ui.Some(ui.Size{Width: width})
	return n
}
func (m *Model) accent() color.NRGBA {
	switch m.selected {
	case "groceries", "flagged":
		return orange
	case "scheduled", "family":
		return color.NRGBA{R: 239, G: 74, B: 67, A: 255}
	default:
		return blue
	}
}

// View returns a fresh declarative tree. All edits use the same model and
// bounded command/history path as keyboard-accessible controls.
func (m *Model) View() ui.Node {
	content := box(ui.Column, "reminders-content", m.toolbar(), m.heading(), m.rows(), m.composerView())
	content.Style = ui.Style{Padding: ui.Some(ui.Insets{Top: nativeHeaderReserve + 10, Right: 28, Bottom: 20, Left: 28}), Gap: ui.Some(float32(10)), GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(1)), Background: ui.Some(white)}
	root := box(ui.Row, "reminders-root", m.sidebarView(), content)
	root.Style.Grow = ui.Some(float32(1))
	root.Style.Background = ui.Some(white)
	root.Focus = m.focus
	if m.editing != "" {
		root.Children = append(root.Children, m.detailsView())
	}
	if m.PendingInput() {
		disableButtons(&root)
	}
	return root
}

func disableButtons(n *ui.Node) {
	if n.Kind == ui.Button {
		n.Disabled = true
	}
	for i := range n.Children {
		disableButtons(&n.Children[i])
	}
}

func (m *Model) sidebarView() ui.Node {
	search := ui.Node{ID: "reminders-search", Kind: ui.TextField, Text: m.query, Accessibility: ui.Accessibility{Name: "Search reminders", Help: "Searches titles and notes in the current list"}, Change: func(e ui.ChangeEvent) { m.inputError(m.SetSearch(e.Text)) }}
	search.Style = ui.Style{GrowX: ui.Some(float32(1)), MaxSize: ui.Some(ui.Size{}), CornerRadius: ui.Some(float32(7)), Background: ui.Some(color.NRGBA{R: 225, G: 225, B: 229, A: 255})}
	specs := []struct {
		id, name string
		icon     ui.Symbol
		c        color.NRGBA
	}{
		{"today", "Today", ui.SymbolSun, blue}, {"scheduled", "Scheduled", ui.SymbolMenu, color.NRGBA{R: 238, G: 76, B: 69, A: 255}},
		{"all", "All", ui.SymbolMenu, muted}, {"flagged", "Flagged", ui.SymbolWarning, orange},
		{"completed", "Completed", ui.SymbolCircleCheck, muted}, {"assigned", "Assigned", ui.SymbolInfo, color.NRGBA{R: 46, G: 167, B: 77, A: 255}},
	}
	children := []ui.Node{search}
	for i := 0; i < len(specs); i += 2 {
		row := box(ui.Row, fmt.Sprintf("smart-row-%d", i))
		row.Style.Gap = ui.Some(float32(8))
		for _, s := range specs[i : i+2] {
			id := s.id
			icon := ui.Node{ID: "smart-icon-" + id, Kind: ui.Icon, Icon: s.icon, Text: s.name, Style: ui.Style{Foreground: ui.Some(s.c), MinSize: ui.Some(ui.Size{Width: 22, Height: 22}), MaxSize: ui.Some(ui.Size{Width: 22, Height: 22})}}
			count := label("smart-count-"+id, fmt.Sprint(m.count(id)), 24, 700, ink)
			count.Style.MinSize = ui.Some(ui.Size{Width: 48})
			count.Style.MaxSize = ui.Some(ui.Size{Width: 48})
			count.Style.TextAlign = ui.Some(ui.AlignEnd)
			top := box(ui.Row, "smart-top-"+id, icon, growSpacer("smart-space-"+id), countColumn("smart-count-box-"+id, 48, count))
			top.Style.Gap = ui.Some(float32(8))
			top.Style.Align = ui.Some(ui.AlignCenter)
			action := button("smart-select-"+id, s.name, fmt.Sprintf("%s, %d reminders", s.name, m.count(id)), "", func() { _ = m.AfterInput(func() error { return m.Select(id) }) })
			action.Style = ui.Style{Padding: ui.Some(ui.Insets{}), Foreground: ui.Some(muted), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 12, Weight: 600}), TextAlign: ui.Some(ui.AlignStart)}
			card := box(ui.Column, "smart-card-"+id, top, action)
			card.Style = ui.Style{Padding: ui.Some(ui.UniformInsets(9)), Gap: ui.Some(float32(2)), GrowX: ui.Some(float32(1)), MinSize: ui.Some(ui.Size{Width: 90, Height: 54}), CornerRadius: ui.Some(float32(9)), Background: ui.Some(white)}
			row.Children = append(row.Children, card)
		}
		children = append(children, row)
	}
	myLists := label("my-lists", "My Lists", 12, 600, muted)
	myLists.Style.Padding = ui.Some(ui.Insets{Top: 14, Left: 8})
	children = append(children, myLists)
	for _, l := range m.lists {
		id := l.ID
		n := button("list-"+id, l.Name, fmt.Sprintf("%s list, %d reminders", l.Name, m.count(id)), ui.SymbolMenu, func() { _ = m.AfterInput(func() error { return m.Select(id) }) })
		n.Style = ui.Style{GrowX: ui.Some(float32(1)), TextAlign: ui.Some(ui.AlignStart), Padding: ui.Some(ui.Insets{})}
		count := label("list-count-"+id, fmt.Sprint(m.count(id)), 13, 0, muted)
		count.Style.MinSize = ui.Some(ui.Size{Width: 28})
		count.Style.MaxSize = ui.Some(ui.Size{Width: 28})
		count.Style.TextAlign = ui.Some(ui.AlignEnd)
		row := box(ui.Row, "list-row-"+id, n, countColumn("list-count-box-"+id, 28, count))
		row.Style = row.Style.Overlay(ui.Style{Padding: ui.Some(ui.Insets{Top: 8, Right: 9, Bottom: 8, Left: 9}), Gap: ui.Some(float32(8)), CornerRadius: ui.Some(float32(6)), Align: ui.Some(ui.AlignCenter)})
		if id == m.selected {
			row.Style.Background = ui.Some(blue)
			row.Children[0].Style.Foreground = ui.Some(white)
			row.Children[1].Children[0].Style.Foreground = ui.Some(white)
		}
		children = append(children, row)
	}
	n := box(ui.Column, "reminders-sidebar", children...)
	n.Style = ui.Style{Padding: ui.Some(ui.Insets{Top: nativeHeaderReserve + 18, Right: 12, Bottom: 18, Left: 12}), Gap: ui.Some(float32(8)), MinSize: ui.Some(ui.Size{Width: 260}), MaxSize: ui.Some(ui.Size{Width: 260}), GrowY: ui.Some(float32(1)), Background: ui.Some(sidebar)}
	return n
}

func (m *Model) toolbar() ui.Node {
	undo := button("reminders-undo", "", "Undo", ui.SymbolUndo, func() { _ = m.Undo() })
	undo.Disabled = m.History().UndoCount == 0 && m.editing == ""
	redo := button("reminders-redo", "", "Redo", ui.SymbolRedo, func() { _ = m.Redo() })
	redo.Disabled = m.History().RedoCount == 0
	add := button("reminders-new", "New Reminder", "Add a new reminder", ui.SymbolAdd, func() {
		_ = m.AfterInput(func() error { m.focus.ID = "reminders-new-title"; m.focus.Sequence++; return nil })
	})
	add.Disabled = len(m.items) >= MaxReminders
	n := box(ui.Row, "reminders-toolbar", undo, redo, growSpacer("reminders-toolbar-space"), add)
	n.Style.Gap = ui.Some(float32(5))
	n.Style.GrowX = ui.Some(float32(1))
	return n
}
func (m *Model) heading() ui.Node {
	title := label("reminders-title", m.title(), 34, 700, m.accent())
	title.Style.GrowX = ui.Some(float32(1))
	count := label("reminders-count", fmt.Sprint(len(m.Visible())), 34, 700, m.accent())
	count.Style.MinSize = ui.Some(ui.Size{Width: 64})
	count.Style.MaxSize = ui.Some(ui.Size{Width: 64})
	count.Style.TextAlign = ui.Some(ui.AlignEnd)
	row := box(ui.Row, "reminders-title-row", title, countColumn("reminders-count-box", 64, count))
	completed := 0
	for _, r := range m.items {
		if r.Completed && m.matches(r, m.selected, true) {
			completed++
		}
	}
	summary := label("reminders-completed-count", fmt.Sprintf("%d Completed", completed), 12, 0, muted)
	summary.Style.MinSize = ui.Some(ui.Size{Width: 100})
	verb := "Show"
	if m.showCompleted {
		verb = "Hide"
	}
	toggle := button("reminders-show-completed", verb, verb+" completed reminders", "", func() { _ = m.AfterInput(func() error { m.showCompleted = !m.showCompleted; return nil }) })
	toggle.Style.Foreground = ui.Some(m.accent())
	toggle.Style.Padding = ui.Some(ui.Insets{})
	line := box(ui.Row, "reminders-completed-summary", summary, toggle)
	line.Style.Gap = ui.Some(float32(8))
	line.Style.GrowX = ui.Some(float32(0))
	n := box(ui.Column, "reminders-heading", row)
	n.Style.Gap = ui.Some(float32(7))
	if m.selected != "completed" {
		n.Children = append(n.Children, line)
	}
	return n
}

func (m *Model) rows() ui.Node {
	rows := []ui.Node{}
	previous := ""
	visible := m.Visible()
	for _, r := range visible {
		if m.selected == "groceries" && m.query == "" && r.Section != "" && r.Section != previous {
			section := label("section-"+r.ID, r.Section, 16, 600, ink)
			section.Style.Padding = ui.Some(ui.Insets{Top: 16, Bottom: 4})
			rows = append(rows, section)
			previous = r.Section
		}
		rows = append(rows, m.reminderRow(r))
	}
	if len(visible) == 0 {
		rows = append(rows, label("reminders-empty", "No Reminders", 20, 600, muted))
	}
	column := box(ui.Column, "reminders-list", rows...)
	column.Style.GrowX = ui.Some(float32(1))
	n := box(ui.Scroll, "reminders-scroll", column)
	n.Style.Grow = ui.Some(float32(1))
	return n
}
func (m *Model) reminderRow(r Reminder) ui.Node {
	icon := ui.SymbolCircle
	state := "Incomplete. Mark complete"
	fg := muted
	if r.Completed {
		icon = ui.SymbolCircleCheck
		state = "Completed. Mark incomplete"
		fg = m.accent()
	}
	toggle := button("complete-"+r.ID, "", r.Title+". "+state, icon, func() { _ = m.AfterInput(func() error { return m.SetCompleted(r.ID, !r.Completed) }) })
	toggle.Style = ui.Style{Foreground: ui.Some(fg), Padding: ui.Some(ui.Insets{}), MinSize: ui.Some(ui.Size{Width: 28, Height: 28})}
	title := button("title-"+r.ID, r.Title, "Edit "+r.Title, "", func() { m.openDetails(r.ID) })
	title.Style = ui.Style{Foreground: ui.Some(ink), Padding: ui.Some(ui.Insets{}), TextAlign: ui.Some(ui.AlignStart), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 14})}
	title.Style.GrowX = ui.Some(float32(1))
	if r.Completed {
		title.Style.Foreground = ui.Some(muted)
	}
	text := box(ui.Column, "text-"+r.ID, title)
	text.Style.GrowX = ui.Some(float32(1))
	text.Style.Gap = ui.Some(float32(2))
	parts := []string{}
	if r.Due != "" {
		due := r.Due
		if due == Today {
			due = "Today"
		}
		parts = append(parts, due)
	}
	if r.Priority != "" {
		parts = append(parts, strings.ToUpper(r.Priority[:1])+r.Priority[1:]+" Priority")
	}
	if r.Flagged {
		parts = append(parts, "Flagged")
	}
	if r.Assignee != "" {
		parts = append(parts, r.Assignee)
	}
	if len(parts) > 0 {
		text.Children = append(text.Children, label("metadata-"+r.ID, strings.Join(parts, " · "), 11, 0, muted))
	}
	if r.Notes != "" {
		text.Children = append(text.Children, label("notes-"+r.ID, r.Notes, 11, 0, muted))
	}
	details := button("details-"+r.ID, "", "Details for "+r.Title, ui.SymbolInfo, func() { m.openDetails(r.ID) })
	details.Style.Foreground = ui.Some(m.accent())
	row := box(ui.Row, "row-"+r.ID, toggle, text, details)
	row.Style = ui.Style{Padding: ui.Some(ui.Insets{Top: 8, Bottom: 8}), Gap: ui.Some(float32(8)), GrowX: ui.Some(float32(1)), Align: ui.Some(ui.AlignCenter)}
	n := box(ui.Column, "reminder-"+r.ID, row, ui.Node{ID: "rule-" + r.ID, Kind: ui.Separator})
	return n
}

func (m *Model) composerView() ui.Node {
	field := ui.Node{ID: "reminders-new-title", Kind: ui.TextField, Text: m.composer, Accessibility: ui.Accessibility{Name: "New reminder title"}, Change: func(e ui.ChangeEvent) {
		if !m.inputAlive() {
			return
		}
		if validText(e.Text, MaxTitleBytes, true) {
			m.composer = e.Text
		} else {
			m.inputError(fmt.Errorf("reminders: title must be valid text of at most 512 bytes"))
		}
	}}
	field.Style.GrowX = ui.Some(float32(1))
	field.Style.MaxSize = ui.Some(ui.Size{})
	add := button("reminders-add", "Add Reminder", "Add reminder", ui.SymbolAdd, func() { _ = m.AfterInput(func() error { _, err := m.Add(m.composer); return err }) })
	add.Disabled = len(m.items) >= MaxReminders
	row := box(ui.Row, "reminders-compose-row", field, add)
	row.Style.Gap = ui.Some(float32(8))
	row.Style.GrowX = ui.Some(float32(1))
	n := box(ui.Column, "reminders-compose", row)
	n.Style.Gap = ui.Some(float32(5))
	if m.status != "" {
		n.Children = append(n.Children, label("reminders-status", m.status, 11, 0, color.NRGBA{R: 190, G: 45, B: 40, A: 255}))
	}
	return n
}

func (m *Model) openDetails(id string) {
	_ = m.AfterInput(func() error {
		at := m.index(id)
		if at >= 0 {
			m.editing = id
			m.draft = m.items[at]
			m.status = ""
			return nil
		}
		return fmt.Errorf("reminders: unknown reminder")
	})
}
func (m *Model) detailsView() ui.Node {
	field := func(id, name, value string, limit int, set func(string)) ui.Node {
		return ui.Node{ID: id, Kind: ui.TextField, Text: value, Accessibility: ui.Accessibility{Name: name}, Change: func(e ui.ChangeEvent) {
			if !m.inputAlive() {
				return
			}
			if validText(e.Text, limit, true) {
				set(e.Text)
			} else {
				m.inputError(fmt.Errorf("reminders: invalid %s", name))
			}
		}}
	}
	priority := m.draft.Priority
	if priority == "" {
		priority = "none"
	}
	rows := []ui.Node{
		label("details-heading", "Reminder Details", 20, 600, ink),
		label("details-title-label", "Title", 12, 600, muted), field("details-title", "Reminder title", m.draft.Title, MaxTitleBytes, func(v string) { m.draft.Title = v }),
		label("details-notes-label", "Notes", 12, 600, muted), field("details-notes", "Reminder notes", m.draft.Notes, MaxNotesBytes, func(v string) { m.draft.Notes = v }),
		label("details-date-label", "Date (YYYY-MM-DD)", 12, 600, muted), field("details-date", "Due date, YYYY-MM-DD", m.draft.Due, 10, func(v string) { m.draft.Due = v }),
		ui.Node{ID: "details-priority", Kind: ui.Selector, SelectedID: priority, Options: []ui.Option{{ID: "none", Label: "No Priority"}, {ID: "low", Label: "Low Priority"}, {ID: "medium", Label: "Medium Priority"}, {ID: "high", Label: "High Priority"}}, Accessibility: ui.Accessibility{Name: "Priority"}, Change: func(e ui.ChangeEvent) {
			if !m.inputAlive() {
				return
			}
			if e.SelectedID == "none" {
				m.draft.Priority = ""
			} else {
				m.draft.Priority = e.SelectedID
			}
		}},
		ui.Node{ID: "details-flagged", Kind: ui.Checkbox, Text: "Flagged", Checked: m.draft.Flagged, Change: func(e ui.ChangeEvent) {
			if m.inputAlive() {
				m.draft.Flagged = e.Checked
			}
		}},
	}
	save := button("details-save", "Done", "Save reminder details", "", m.doneDetails)
	save.Appearance = ui.ButtonPrimary
	cancel := button("details-cancel", "Cancel", "Cancel editing reminder", "", m.cancelDetails)
	row := box(ui.Row, "details-actions", growSpacer("details-space"), cancel, save)
	row.Style.Gap = ui.Some(float32(8))
	rows = append(rows, row)
	if m.status != "" {
		rows = append(rows, label("details-status", m.status, 11, 0, color.NRGBA{R: 190, G: 45, B: 40, A: 255}))
	}
	form := box(ui.Column, "details-form", rows...)
	form.Style = ui.Style{Padding: ui.Some(ui.UniformInsets(20)), Gap: ui.Some(float32(10)), MinSize: ui.Some(ui.Size{Width: 380}), Background: ui.Some(white)}
	return ui.Node{ID: "reminders-details", Kind: ui.Dialog, Text: "Reminder Details", Open: true, Dismiss: m.cancelDetails, Children: []ui.Node{form}}
}
