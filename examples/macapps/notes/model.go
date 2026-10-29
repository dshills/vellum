// Package notes is a bounded Notes-style Vellum consumer.
package notes

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

const (
	MaxNotes      = 64
	MaxBodyBytes  = 32 << 10
	MaxTitleBytes = 256
	HistoryLimit  = 32
)

var (
	ErrLimit   = errors.New("notes: demonstration capacity reached")
	ErrInvalid = errors.New("notes: invalid text or identity")
	ErrPending = errors.New("notes: finishing earlier text input")
)

// Note is a detached presentation snapshot. Body is plain UTF-8 text.
type Note struct {
	ID, FolderID, Title, Body string
	Pinned                    bool
}
type note struct {
	value  Note
	editor *editor.Model
}
type folder struct{ id, name string }

// Model and its Commands belong to one application UI owner. There is no
// account access or network service. Optional Files owns persistence workers.
type Model struct {
	files                                            *Files
	notes                                            []*note
	folders                                          []folder
	selected, selectedFolder, query, status, touched string
	nextID                                           uint64
	journal                                          *history.Journal
	commands                                         *command.Registry
	focus                                            ui.FocusRequest
	settlement                                       func(func(error)) error
	pending                                          bool
	nativeToolbar                                    bool
	pendingInputError                                error
}

func New() *Model {
	j, _ := history.New(HistoryLimit)
	m := &Model{selected: "n1", selectedFolder: "notes", nextID: 10, journal: j, commands: command.New(), folders: []folder{{"notes", "Notes"}, {"projects", "Projects"}, {"personal", "Personal"}}}
	for _, value := range []Note{
		{ID: "n1", FolderID: "notes", Title: "Monday Morning Meeting", Body: "Planning for the week\n\nStudio updates\nDiscuss the new exhibition space and the sketches from Friday.\n\nNext steps\nReview the material samples.\nChoose a time for the site visit.\nBring a notebook and a camera.", Pinned: true},
		{ID: "n2", FolderID: "notes", Title: "Family Fun", Body: "Ideas for the weekend\n\nVisit the botanical gardens.\nMake pizza together.\nWalk along the waterfront."},
		{ID: "n3", FolderID: "notes", Title: "Places to hike", Body: "Coast trail\nA gentle walk with a view of the bay.\n\nForest loop\nRemember to bring water and a light jacket."},
		{ID: "n4", FolderID: "projects", Title: "Customized Filtration", Body: "Our mission is to provide a simpler way to enjoy clean water.\n\nSimple installation\nRecycled materials\nLong-lasting design\n\nReview the prototype and document the next round of improvements."},
		{ID: "n5", FolderID: "projects", Title: "Studio ideas", Body: "Light, space, and a place to think.\n\nKeep the workbench near the window.\nAdd shelves for materials and reference books."},
		{ID: "n6", FolderID: "personal", Title: "Reading list", Body: "The Art of Noticing\nA Field Guide to Getting Lost\nThe Creative Act"},
	} {
		e, err := editor.New(value.Body, editor.Config{MaxBytes: MaxBodyBytes, HistoryLimit: 1})
		if err != nil {
			panic(err)
		}
		m.notes = append(m.notes, &note{value: value, editor: e})
	}
	for _, c := range []command.Command{
		{ID: "notes.delete", Label: "Delete Note", Available: func() error {
			if err := m.actionReady(); err != nil {
				return err
			}
			if m.lookup(m.selected) == nil {
				return ErrInvalid
			}
			return nil
		}, Execute: func(context.Context) error { return m.Request(func() error { return m.Delete(m.selected) }) }},
		{ID: "notes.pin", Label: "Pin or Unpin Note", Available: func() error {
			if err := m.actionReady(); err != nil {
				return err
			}
			if m.lookup(m.selected) == nil {
				return ErrInvalid
			}
			return nil
		}, Execute: func(context.Context) error {
			return m.Request(func() error {
				n := m.lookup(m.selected)
				if n == nil {
					return ErrInvalid
				}
				return m.SetPinned(n.value.ID, !n.value.Pinned)
			})
		}},
		{ID: "notes.new", Label: "New Note", Shortcut: "Meta+N", Available: func() error {
			if len(m.notes) >= MaxNotes {
				return ErrLimit
			}
			return m.actionReady()
		}, Execute: func(context.Context) error { return m.Request(func() error { _, err := m.Add(); return err }) }},
		{ID: "notes.undo", Label: "Undo", Shortcut: "Meta+Z", Available: func() error {
			if err := m.actionReady(); err != nil {
				return err
			}
			if m.settlement == nil && m.journal.Snapshot().UndoCount == 0 {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error { return m.Request(m.Undo) }},
		{ID: "notes.redo", Label: "Redo", Shortcut: "Meta+Shift+Z", Available: func() error {
			if err := m.actionReady(); err != nil {
				return err
			}
			if m.settlement == nil && m.journal.Snapshot().RedoCount == 0 {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error { return m.Request(m.Redo) }},
	} {
		if err := m.commands.Register(c); err != nil {
			panic(err)
		}
	}
	return m
}

func (m *Model) Commands() *command.Registry { return m.commands }
func (m *Model) History() history.State      { return m.journal.Snapshot() }
func (m *Model) Selected() string            { return m.selected }
func (m *Model) SelectedFolder() string      { return m.selectedFolder }
func (m *Model) Pending() bool               { return m.pending }

// CurrentInput is owner-only native host wiring. The title ID and public body
// model are stable until a structural request completes; neither is present
// when no note is selected. The model must not be copied or used on a worker.
func (m *Model) CurrentInput() (string, *editor.Model) {
	if n := m.lookup(m.selected); n != nil {
		return "note-title-" + n.value.ID, n.editor
	}
	return "", nil
}

// SetSettlement installs an owner-confined asynchronous input boundary. Its
// callback must run on that same owner. Immediate errors need not call back.
// The host must invalidate after asynchronous completion. Nil keeps direct,
// portable model behavior. Swapping a pending boundary is refused.
func (m *Model) SetSettlement(boundary func(func(error)) error) error {
	if m.pending {
		return ErrPending
	}
	m.settlement = boundary
	return nil
}

// Request settles originating native input once before executing a synchronous
// model action. At most one request is retained. Duplicate or late completion
// callbacks cannot repeat an action; committed Change callbacks remain admitted
// while pending. Direct mutators remain available for portable model consumers.
func (m *Model) Request(action func() error) error { return m.request(action, false) }
func (m *Model) request(action func() error, preparingQuit bool) error {
	if action == nil {
		return ErrInvalid
	}
	if m.pending || m.files != nil && (m.files.blockedWithoutQuit() || m.files.quitHeld && !preparingQuit) {
		return ErrPending
	}
	if m.settlement == nil {
		err := action()
		m.accept(err)
		return err
	}
	m.pending = true
	m.pendingInputError = nil
	m.status = "Finishing text input"
	completed := false
	var outcome error
	finish := func(err error) {
		if completed {
			return
		}
		completed = true
		if err == nil {
			err = m.pendingInputError
		}
		m.pendingInputError = nil
		m.pending = false
		if err == nil {
			err = action()
		}
		outcome = err
		m.accept(err)
		if err != nil && preparingQuit && m.files != nil {
			m.files.finishQuit(false)
		}
	}
	if err := m.settlement(finish); err != nil {
		finish(err)
	}
	if completed {
		return outcome
	}
	return nil
}

func (m *Model) actionReady() error {
	if m.files != nil && m.files.blocked() {
		return ErrPending
	}
	if m.pending {
		return ErrPending
	}
	if m.settlement != nil {
		return nil
	}
	return m.ready()
}
func (m *Model) Items() []Note {
	out := make([]Note, 0, len(m.notes))
	for _, n := range m.notes {
		out = append(out, n.value)
	}
	return out
}
func (m *Model) lookup(id string) *note {
	for _, n := range m.notes {
		if n.value.ID == id {
			return n
		}
	}
	return nil
}
func validText(s string, limit int) bool {
	return len(s) <= limit && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}
func (m *Model) accept(err error) {
	m.status = ""
	if err != nil {
		m.status = err.Error()
		if m.pending {
			m.pendingInputError = err
		}
	} else if m.pending {
		m.status = "Finishing text input"
	}
}

func (m *Model) ready() error {
	if m.files != nil && m.files.blocked() {
		return ErrPending
	}
	for _, n := range m.notes {
		if n.editor.Snapshot().Composing {
			return editor.ErrComposition
		}
	}
	return nil
}

// flush admits any earlier committed public-editor changes before a toolbar
// action, even if the native Change callback has not yet rendered the list.
func (m *Model) flush() error {
	if err := m.ready(); err != nil {
		return err
	}
	for _, n := range m.notes {
		if err := m.BodyChanged(n.value.ID); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) Select(id string) error {
	if m.pending {
		return ErrPending
	}
	if err := m.flush(); err != nil {
		return err
	}
	if m.lookup(id) == nil {
		return ErrInvalid
	}
	m.selected = id
	m.status = ""
	return nil
}
func (m *Model) SelectFolder(id string) error {
	if m.pending {
		return ErrPending
	}
	if id != "all" && !slices.ContainsFunc(m.folders, func(f folder) bool { return f.id == id }) {
		return ErrInvalid
	}
	if err := m.flush(); err != nil {
		return err
	}
	m.selectedFolder = id
	m.query = ""
	visible := m.Visible()
	m.selected = ""
	if len(visible) > 0 {
		m.selected = visible[0].ID
	}
	return nil
}
func (m *Model) SetSearch(query string) error {
	if !validText(query, MaxTitleBytes) {
		return ErrInvalid
	}
	m.query = query
	return nil
}
func (m *Model) Visible() []Note {
	query := strings.ToLower(strings.TrimSpace(m.query))
	out := make([]Note, 0, len(m.notes))
	for _, n := range m.notes {
		if (m.selectedFolder == "all" || n.value.FolderID == m.selectedFolder) && (query == "" || strings.Contains(strings.ToLower(n.value.Title+"\n"+n.value.Body), query)) {
			out = append(out, n.value)
		}
	}
	slices.SortStableFunc(out, func(a, b Note) int {
		if a.Pinned && !b.Pinned {
			return -1
		}
		if !a.Pinned && b.Pinned {
			return 1
		}
		return 0
	})
	return out
}
func (m *Model) reveal(id string) {
	m.selected = id
	m.query = ""
	if n := m.lookup(id); n != nil {
		m.selectedFolder = n.value.FolderID
	}
}
func (m *Model) requestFocus(id string) {
	if m.focus.Sequence != ^uint64(0) {
		m.focus = ui.FocusRequest{ID: id, Sequence: m.focus.Sequence + 1}
	}
}

// Add creates one stable note. Undo removes it; redo restores the same editor,
// identity and contents. New identities never reuse the deleted/undone identity.
func (m *Model) Add() (string, error) {
	if m.pending {
		return "", ErrPending
	}
	if err := m.flush(); err != nil {
		return "", err
	}
	if len(m.notes) >= MaxNotes || m.nextID == ^uint64(0) {
		return "", ErrLimit
	}
	folderID := m.selectedFolder
	if folderID == "all" {
		folderID = "notes"
	}
	e, err := editor.New("", editor.Config{MaxBytes: MaxBodyBytes, HistoryLimit: 1})
	if err != nil {
		return "", err
	}
	n := &note{value: Note{ID: fmt.Sprintf("n%d", m.nextID), FolderID: folderID}, editor: e}
	m.nextID++
	err = m.journal.Apply(history.Change{Label: "New Note", Apply: func() error {
		if m.lookup(n.value.ID) != nil || len(m.notes) >= MaxNotes {
			return ErrLimit
		}
		m.notes = append(m.notes, n)
		m.reveal(n.value.ID)
		return nil
	}, Revert: func() error { return m.remove(n) }})
	if err == nil {
		m.requestFocus("note-title-" + n.value.ID)
	}
	m.accept(err)
	return n.value.ID, err
}
func (m *Model) remove(n *note) error {
	i := slices.Index(m.notes, n)
	if i < 0 {
		return ErrInvalid
	}
	m.notes = slices.Delete(m.notes, i, i+1)
	if m.selected == n.value.ID {
		m.selected = ""
		visible := m.Visible()
		if len(visible) > 0 {
			m.selected = visible[0].ID
		}
	}
	return nil
}
func (m *Model) Delete(id string) error {
	if m.pending {
		return ErrPending
	}
	if err := m.flush(); err != nil {
		return err
	}
	n := m.lookup(id)
	if n == nil {
		return ErrInvalid
	}
	index := slices.Index(m.notes, n)
	err := m.journal.Apply(history.Change{Label: "Delete Note", Apply: func() error { return m.remove(n) }, Revert: func() error {
		if len(m.notes) >= MaxNotes || m.lookup(id) != nil {
			return ErrLimit
		}
		m.notes = slices.Insert(m.notes, min(index, len(m.notes)), n)
		m.reveal(id)
		return nil
	}})
	m.accept(err)
	return err
}

// Update is atomic and undoable; it rejects invalid/oversized UTF-8 before any
// mutation and preserves a live composition until the application settles it.
func (m *Model) Update(id, title, body string) error {
	if !validText(title, MaxTitleBytes) || !validText(body, MaxBodyBytes) {
		return ErrInvalid
	}
	if err := m.flush(); err != nil {
		return err
	}
	n := m.lookup(id)
	if n == nil {
		return ErrInvalid
	}
	before := n.value
	after := before
	after.Title = title
	after.Body = body
	if before == after {
		return nil
	}
	install := func(expected, replacement Note) error {
		if n.value != expected || m.lookup(id) != n {
			return history.ErrStale
		}
		if replacement.Body != expected.Body {
			if err := n.editor.Replace(replacement.Body, editor.Selection{}); err != nil {
				return err
			}
		}
		n.value = replacement
		m.touched = id
		return nil
	}
	err := m.journal.Apply(history.Change{Label: "Edit Note", Apply: func() error { return install(before, after) }, Revert: func() error { return install(after, before) }})
	m.accept(err)
	return err
}

// titleChanged admits the controlled native title field directly, including
// during settlement. A composition in the independent body must not reject an
// earlier committed title. The normal native FIFO orders these callbacks.
func (m *Model) titleChanged(id, title string) error {
	if !validText(title, MaxTitleBytes) {
		return ErrInvalid
	}
	n := m.lookup(id)
	if n == nil {
		return ErrInvalid
	}
	if !n.editor.Snapshot().Composing {
		if err := m.BodyChanged(id); err != nil {
			return err
		}
	}
	before := n.value.Title
	if before == title {
		return nil
	}
	install := func(expected, value string) error {
		if m.lookup(id) != n || n.value.Title != expected {
			return history.ErrStale
		}
		n.value.Title = value
		m.touched = id
		return nil
	}
	return m.journal.Apply(history.Change{Label: "Edit Note", Apply: func() error { return install(before, title) }, Revert: func() error { return install(title, before) }})
}

// BodyChanged records an already-committed native editor update once, using the
// same app journal as structural and title edits. Programmatic replay goes
// through editor.Replace; the native adapter remains the sole text bridge.
func (m *Model) BodyChanged(id string) error {
	n := m.lookup(id)
	if n == nil {
		return ErrInvalid
	}
	s := n.editor.Snapshot()
	if s.Composing {
		return editor.ErrComposition
	}
	if s.Text == n.value.Body {
		return nil
	}
	before, after := n.value.Body, s.Text
	first := true
	err := m.journal.Apply(history.Change{Label: "Edit Note", Apply: func() error {
		if m.lookup(id) != n || n.value.Body != before {
			return history.ErrStale
		}
		if first {
			if n.editor.Snapshot().Text != after {
				return history.ErrStale
			}
			first = false
		} else if err := n.editor.Replace(after, editor.Selection{}); err != nil {
			return err
		}
		n.value.Body = after
		m.touched = id
		return nil
	}, Revert: func() error {
		if m.lookup(id) != n || n.value.Body != after {
			return history.ErrStale
		}
		if err := n.editor.Replace(before, editor.Selection{}); err != nil {
			return err
		}
		n.value.Body = before
		m.touched = id
		return nil
	}})
	m.accept(err)
	return err
}
func (m *Model) SetPinned(id string, pinned bool) error {
	if m.pending {
		return ErrPending
	}
	if err := m.flush(); err != nil {
		return err
	}
	n := m.lookup(id)
	if n == nil {
		return ErrInvalid
	}
	before := n.value.Pinned
	if before == pinned {
		return nil
	}
	install := func(expected, value bool) error {
		if m.lookup(id) != n || n.value.Pinned != expected {
			return history.ErrStale
		}
		n.value.Pinned = value
		m.touched = id
		return nil
	}
	err := m.journal.Apply(history.Change{Label: "Pin Note", Apply: func() error { return install(before, pinned) }, Revert: func() error { return install(pinned, before) }})
	m.accept(err)
	return err
}
func (m *Model) Undo() error {
	if m.pending {
		return ErrPending
	}
	if err := m.flush(); err != nil {
		return err
	}
	m.touched = ""
	err := m.journal.Undo()
	if err == nil && m.lookup(m.touched) != nil {
		m.reveal(m.touched)
	}
	m.accept(err)
	return err
}
func (m *Model) Redo() error {
	if m.pending {
		return ErrPending
	}
	if err := m.flush(); err != nil {
		return err
	}
	m.touched = ""
	err := m.journal.Redo()
	if err == nil && m.lookup(m.touched) != nil {
		m.reveal(m.touched)
	}
	m.accept(err)
	return err
}

// SetNativeToolbar selects the host-owned native app shell; it does not alter
// notes, history or input IDs. Configure on the application owner before View.
func (m *Model) SetNativeToolbar(enabled bool) { m.nativeToolbar = enabled }
