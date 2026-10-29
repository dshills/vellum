package notes

import (
	"context"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

const findID = "notes-find"

func (m *Model) initFind() {
	m.find = &ui.FindBar{ID: findID, Label: "Find in this note", Small: true}
	m.find.Change = func(e ui.FindBarChange) {
		switch e.Field {
		case "query":
			m.find.Query = e.Text
		case "replacement":
			m.find.Replacement = e.Text
		case "case":
			m.find.Options.CaseSensitive = e.Checked
		case "word":
			m.find.Options.WholeWord = e.Checked
		case "replace-open":
			m.find.ReplaceOpen = e.Checked
		}
	}
	m.find.Close = func() { _ = m.closeFind() }
	for _, entry := range []struct {
		id, label, key string
		action         int
	}{
		{"find", "Find…", "Meta+F", -1},
		{"find-replace", "Find and Replace…", "Meta+Alt+F", -2},
		{"find-next", "Find Next", "Meta+G", int(ui.FindNext)},
		{"find-previous", "Find Previous", "Meta+Shift+G", int(ui.FindPrevious)},
		{"find-close", "Close Find", "Escape", -3},
	} {
		entry := entry
		err := m.commands.Register(command.Command{ID: "notes." + entry.id, Label: entry.label, Shortcut: entry.key, ShortcutScope: command.WindowShortcuts, Available: func() error {
			if err := m.actionReady(); err != nil {
				return err
			}
			if m.lookup(m.selected) == nil || entry.action == -3 && !m.find.Open {
				return ErrInvalid
			}
			if entry.action >= 0 && m.find.Query == "" && m.settlement == nil {
				return ui.ErrFindNoMatch
			}
			return nil
		}, Execute: func(context.Context) error {
			switch entry.action {
			case -1:
				return m.openFind(false)
			case -2:
				return m.openFind(true)
			case -3:
				return m.closeFind()
			default:
				return m.requestFind(ui.FindAction(entry.action), m.selected)
			}
		}})
		if err != nil {
			panic(err)
		}
	}
}
func (m *Model) openFind(replace bool) error {
	return m.Request(func() error {
		if err := m.flush(); err != nil {
			return err
		}
		n := m.lookup(m.selected)
		if n == nil {
			return ErrInvalid
		}
		m.find.Model = n.editor
		m.find.Open = true
		m.find.ReplaceOpen = m.find.ReplaceOpen || replace
		m.requestFocus(ui.FindBarInputID(findID))
		return nil
	})
}
func (m *Model) closeFind() error {
	return m.Request(func() error {
		if err := m.flush(); err != nil {
			return err
		}
		m.find.Open = false
		m.find.Model = nil
		if m.lookup(m.selected) != nil {
			m.requestFocus("note-editor-" + m.selected)
		}
		return nil
	})
}
func (m *Model) requestFind(action ui.FindAction, expected string) error {
	return m.Request(func() error {
		// Resolve controlled fields after the native FIFO settlement, not from a
		// pre-settlement button snapshot. Selection changes do not dirty the notebook.
		if expected != m.selected {
			return ErrInvalid
		}
		if err := m.flush(); err != nil {
			return err
		}
		n := m.lookup(expected)
		if n == nil {
			return ErrInvalid
		}
		q := ui.FindRequest{Action: action, Query: m.find.Query, Replacement: m.find.Replacement, Options: m.find.Options}
		if err := q.Apply(n.editor); err != nil {
			return err
		}
		return m.BodyChanged(expected)
	})
}
func (m *Model) findView(n *note) ui.Node {
	m.find.Model = n.editor
	m.find.Disabled = m.files != nil && m.files.blockedWithoutQuit()
	// Capture note identity; a retained action cannot edit a different selection.
	m.find.Request = func(q ui.FindRequest) { _ = m.requestFind(q.Action, n.value.ID) }
	bar, err := m.find.View()
	if err != nil {
		panic(err)
	}
	bar.Style = bar.Style.Overlay(ui.Style{Padding: ui.Some(ui.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8}), Background: ui.Some(sidebar), Gap: ui.Some(float32(6))})
	// Compact navigation leaves space for prose in the narrow document column.
	for i := range bar.Children {
		row := &bar.Children[i]
		row.Style.Gap = ui.Some(float32(6))
		for j := range row.Children {
			item := &row.Children[j]
			switch item.ID {
			case findID + "/find/previous":
				item.Text = "‹"
				item.Accessibility.Name = "Previous match"
			case findID + "/find/next":
				item.Text = "›"
				item.Accessibility.Name = "Next match"
			case findID + "/find/count":
				item.Style.MinSize = ui.Some(ui.Size{Width: 78})
				item.Style.MaxSize = ui.Some(ui.Size{Width: 78})
			case ui.FindBarInputID(findID):
				item.Style.MinSize = ui.Some(ui.Size{Width: 100, Height: 24})
			}
		}
	}
	return bar
}

// SettleInputs is shared app-host wiring. Each step is asynchronous on the UI
// owner; duplicate acknowledgments cannot skip a field or finish twice. No
// native driver import is needed in this application-owned portable package.
func (m *Model) SettleInputs(field func(string, func(error)) error, body func(*editor.Model, func(error)) error, done func(error)) error {
	if field == nil || body == nil || done == nil {
		return ErrInvalid
	}
	fields := []string{"notes-search"}
	title, model := m.CurrentInput()
	if title != "" && m.find.Open {
		fields = append(fields, ui.FindBarInputID(findID))
		if m.find.ReplaceOpen {
			fields = append(fields, findID+"/find/replacement")
		}
	}
	if title != "" {
		fields = append(fields, title)
	}
	finished := false
	finish := func(err error) {
		if !finished {
			finished = true
			done(err)
		}
	}
	var next func(int)
	next = func(index int) {
		once := false
		complete := func(err error) {
			if once || finished {
				return
			}
			once = true
			if err != nil {
				finish(err)
				return
			}
			if index < len(fields) {
				next(index + 1)
			} else {
				finish(nil)
			}
		}
		if index < len(fields) {
			if err := field(fields[index], complete); err != nil {
				complete(err)
			}
		} else if model != nil {
			if err := body(model, complete); err != nil {
				complete(err)
			}
		} else {
			finish(nil)
		}
	}
	next(0)
	return nil
}
