// Package reminders is a bounded, in-memory Reminders-style Vellum consumer.
package reminders

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

const (
	MaxReminders  = 256
	MaxTitleBytes = 512
	MaxNotesBytes = 4096
	Today         = "2026-10-19" // Stable demonstration date, never the user's calendar.
)

var ErrLimit = errors.New("reminders: demonstration capacity reached")

type Reminder struct {
	ID, ListID, Title, Notes, Due, Section, Assignee string
	Priority                                         string
	Completed, Flagged                               bool
}

type List struct{ ID, Name string }

// Model and all callbacks are owned by the application's UI owner. It performs
// no persistence, account access, notification scheduling or background work.
type Model struct {
	items                                                []Reminder
	lists                                                []List
	selected, lastList, query, composer, editing, status string
	draft                                                Reminder
	nextID                                               uint64
	showCompleted                                        bool
	journal                                              *history.Journal
	commands                                             *command.Registry
	focus                                                ui.FocusRequest
	inputBoundary                                        InputBoundary
	inputLifetime                                        context.Context
	pending                                              *inputAction
	closed                                               bool
}

func New() *Model {
	j, _ := history.New(64)
	m := &Model{selected: "groceries", lastList: "groceries", nextID: 20, journal: j, commands: command.New(), lists: []List{
		{"reminders", "Reminders"}, {"family", "Family"}, {"work", "Work"}, {"camping", "Camping Trip"}, {"groceries", "Groceries"},
	}, items: []Reminder{
		{ID: "r1", ListID: "groceries", Title: "Croissants", Section: "Bakery"},
		{ID: "r2", ListID: "groceries", Title: "Bread", Section: "Bakery"},
		{ID: "r3", ListID: "groceries", Title: "Milk", Section: "Dairy", Due: Today},
		{ID: "r4", ListID: "groceries", Title: "Cheese", Section: "Dairy"},
		{ID: "r5", ListID: "groceries", Title: "Butter", Section: "Dairy", Flagged: true},
		{ID: "r6", ListID: "groceries", Title: "Bananas", Section: "Fruit"},
		{ID: "r7", ListID: "groceries", Title: "Apples", Section: "Fruit", Completed: true},
		{ID: "r8", ListID: "work", Title: "Review the launch checklist", Due: Today, Priority: "high", Flagged: true},
		{ID: "r9", ListID: "work", Title: "Send the updated drawings", Due: "2026-10-20", Assignee: "You"},
		{ID: "r10", ListID: "family", Title: "Book a table for Saturday", Due: "2026-10-24", Assignee: "You"},
		{ID: "r11", ListID: "camping", Title: "Check the tent", Notes: "Check the poles and groundsheet.", Due: "2026-10-21"},
		{ID: "r12", ListID: "reminders", Title: "Return library books", Due: Today},
	}}
	for _, c := range []command.Command{
		{ID: "reminders.undo", Label: "Undo", Shortcut: "Meta+Z", ShortcutScope: command.AfterFocusShortcuts, Available: func() error {
			if err := m.mutationReady(); err != nil {
				return err
			}
			if m.journal.Snapshot().UndoCount == 0 && m.editing == "" {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error { return m.Undo() }},
		{ID: "reminders.redo", Label: "Redo", Shortcut: "Meta+Shift+Z", ShortcutScope: command.AfterFocusShortcuts, Available: func() error {
			if err := m.mutationReady(); err != nil {
				return err
			}
			if m.journal.Snapshot().RedoCount == 0 {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error { return m.Redo() }},
	} {
		if err := m.commands.Register(c); err != nil {
			panic(err)
		}
	}
	return m
}

func (m *Model) Commands() *command.Registry { return m.commands }
func (m *Model) Items() []Reminder           { return slices.Clone(m.items) }
func (m *Model) History() history.State      { return m.journal.Snapshot() }
func (m *Model) Selected() string            { return m.selected }
func (m *Model) Search() string              { return m.query }

func validText(value string, max int, blank bool) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && len(value) <= max && (blank || strings.TrimSpace(value) != "")
}
func validReminder(r Reminder) bool {
	if !validText(r.Title, MaxTitleBytes, false) || !validText(r.Notes, MaxNotesBytes, true) {
		return false
	}
	if r.Due != "" {
		if _, err := time.Parse("2006-01-02", r.Due); err != nil {
			return false
		}
	}
	return r.Priority == "" || r.Priority == "low" || r.Priority == "medium" || r.Priority == "high"
}
func (m *Model) index(id string) int {
	return slices.IndexFunc(m.items, func(r Reminder) bool { return r.ID == id })
}
func (m *Model) listExists(id string) bool {
	return slices.ContainsFunc(m.lists, func(l List) bool { return l.ID == id })
}
func (m *Model) Select(id string) error {
	if err := m.mutationReady(); err != nil {
		return err
	}
	if !m.listExists(id) && !slices.Contains([]string{"today", "scheduled", "all", "flagged", "completed", "assigned"}, id) {
		return errors.New("reminders: unknown list")
	}
	m.selected = id
	if m.listExists(id) {
		m.lastList = id
	}
	m.status = ""
	return nil
}
func (m *Model) SetSearch(query string) error {
	if m.closed {
		return ErrClosed
	}
	if m.inputLifetime != nil {
		if err := m.inputLifetime.Err(); err != nil {
			return err
		}
	}
	if !validText(query, MaxTitleBytes, true) {
		return errors.New("reminders: search is too long or invalid")
	}
	m.query = query
	return nil
}
func (m *Model) accept(err error) {
	m.status = ""
	if err != nil {
		m.status = err.Error()
	}
}

// Add creates one stable object; redo restores the identical ID and data.
func (m *Model) Add(title string) (string, error) {
	if err := m.mutationReady(); err != nil {
		return "", err
	}
	if len(m.items) >= MaxReminders {
		return "", ErrLimit
	}
	if m.nextID == ^uint64(0) {
		return "", ErrLimit
	}
	r := Reminder{ID: fmt.Sprintf("r%d", m.nextID), ListID: m.lastList, Title: strings.TrimSpace(title)}
	if m.selected == "today" {
		r.Due = Today
	}
	if m.selected == "flagged" {
		r.Flagged = true
	}
	if m.selected == "assigned" {
		r.Assignee = "You"
	}
	if !validReminder(r) {
		return "", errors.New("reminders: enter a title of at most 512 bytes")
	}
	position := len(m.items)
	err := m.journal.Apply(history.Change{Label: "Add Reminder", Apply: func() error {
		if m.index(r.ID) >= 0 || len(m.items) >= MaxReminders || position > len(m.items) {
			return errors.New("reminders: stale insertion")
		}
		m.items = slices.Insert(m.items, position, r)
		return nil
	}, Revert: func() error {
		at := m.index(r.ID)
		if at < 0 {
			return errors.New("reminders: added reminder missing")
		}
		m.items = slices.Delete(m.items, at, at+1)
		return nil
	}})
	if err != nil {
		return "", err
	}
	m.nextID++
	m.composer = ""
	return r.ID, nil
}

func (m *Model) replace(id, label string, change func(*Reminder)) error {
	if err := m.mutationReady(); err != nil {
		return err
	}
	index := m.index(id)
	if index < 0 {
		return errors.New("reminders: unknown reminder")
	}
	before := m.items[index]
	after := before
	change(&after)
	if !validReminder(after) || after.ID != before.ID || after.ListID != before.ListID {
		return errors.New("reminders: invalid reminder details")
	}
	if after == before {
		return nil
	}
	install := func(expected, replacement Reminder) error {
		at := m.index(id)
		if at < 0 || m.items[at] != expected {
			return errors.New("reminders: stale edit")
		}
		m.items[at] = replacement
		return nil
	}
	return m.journal.Apply(history.Change{Label: label, Apply: func() error { return install(before, after) }, Revert: func() error { return install(after, before) }})
}
func (m *Model) SetCompleted(id string, completed bool) error {
	return m.replace(id, "Complete Reminder", func(r *Reminder) { r.Completed = completed })
}
func (m *Model) SetFlagged(id string, flagged bool) error {
	return m.replace(id, "Flag Reminder", func(r *Reminder) { r.Flagged = flagged })
}
func (m *Model) Update(id, title, notes, due, priority string, flagged bool) error {
	return m.replace(id, "Edit Reminder", func(r *Reminder) {
		r.Title = strings.TrimSpace(title)
		r.Notes = notes
		r.Due = due
		r.Priority = priority
		r.Flagged = flagged
	})
}
func (m *Model) Undo() error {
	return m.AfterInput(func() error {
		if err := m.saveDraft(); err != nil {
			return err
		}
		err := m.journal.Undo()
		if err == nil {
			m.editing = ""
		}
		m.accept(err)
		return err
	})
}
func (m *Model) Redo() error {
	return m.AfterInput(func() error {
		if m.draftBlocksRedo() {
			return ErrDraftForRedo
		}
		err := m.journal.Redo()
		if err == nil {
			m.editing = ""
		}
		m.accept(err)
		return err
	})
}

func (m *Model) matches(r Reminder, view string, includeCompleted bool) bool {
	if view == "completed" {
		return r.Completed
	}
	if r.Completed && !includeCompleted {
		return false
	}
	switch view {
	case "today":
		return r.Due != "" && r.Due <= Today
	case "scheduled":
		return r.Due != ""
	case "all":
		return true
	case "flagged":
		return r.Flagged
	case "assigned":
		return r.Assignee != ""
	default:
		return r.ListID == view
	}
}
func (m *Model) Visible() []Reminder {
	query := strings.ToLower(strings.TrimSpace(m.query))
	out := make([]Reminder, 0, len(m.items))
	for _, r := range m.items {
		if m.matches(r, m.selected, m.showCompleted) && (query == "" || strings.Contains(strings.ToLower(r.Title+"\n"+r.Notes), query)) {
			out = append(out, r)
		}
	}
	return out
}
func (m *Model) count(view string) int {
	n := 0
	for _, r := range m.items {
		if m.matches(r, view, false) {
			n++
		}
	}
	return n
}
func (m *Model) title() string {
	if m.query != "" {
		return "Search Results"
	}
	for _, l := range m.lists {
		if l.ID == m.selected {
			return l.Name
		}
	}
	return map[string]string{"today": "Today", "scheduled": "Scheduled", "all": "All", "flagged": "Flagged", "completed": "Completed", "assigned": "Assigned"}[m.selected]
}
