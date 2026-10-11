// Package editor provides an owner-confined, UTF-8 document model. Offsets are
// document-wide Unicode scalar offsets, including each CR and LF separately.
package editor

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/history"
)

const (
	DefaultMaxBytes     = 4 << 20
	DefaultHistoryLimit = 128
)

var (
	ErrInvalidModel      = errors.New("editor: use New to initialize a model")
	ErrInvalidText       = errors.New("editor: valid UTF-8 without NUL required")
	ErrRange             = errors.New("editor: invalid scalar range or selection")
	ErrTooLarge          = errors.New("editor: document byte budget exceeded")
	ErrStale             = errors.New("editor: edit does not match current content")
	ErrComposition       = errors.New("editor: settle composition before committed changes")
	ErrInstallContext    = errors.New("editor: Install requires a journal installation callback")
	ErrRevisionExhausted = history.ErrRevisionExhausted
)

// Config bounds document bytes and the number of retained sparse undo changes.
// Zero fields select defaults; negative fields are invalid. HistoryLimit counts
// changes, not bytes: each change retains its removed and inserted span.
type Config struct{ MaxBytes, HistoryLimit int }

type Range struct{ Start, End int }

// Selection preserves direction. Anchor and Focus may be equal at a caret.
type Selection struct{ Anchor, Focus int }

// Snapshot is immutable. Text contains committed bytes; VisibleText includes
// preedit while Composing. The three monotonic clocks are independent of the
// journal's saved-state identity. Revision advances on committed content changes
// and installs; VisibleEpoch on visible content changes; SelectionEpoch on
// directional selection changes. They never reuse values after undo or cancel.
type Snapshot struct {
	Text, VisibleText                      string
	Selection                              Selection
	Revision, VisibleEpoch, SelectionEpoch uint64
	Composing                              bool
	Marked                                 Range
	History                                history.State
}

// Edit replaces Range in the current VisibleText. Before must exactly match
// that scalar span, and both epochs must match the captured snapshot. Selection
// and (for Preview) Marked refer to the resulting document.
type Edit struct {
	BaseRevision, BaseVisibleEpoch uint64
	Range                          Range
	Before, Insert                 string
	Selection                      Selection
	Marked                         Range
}

type preedit struct {
	text      string
	selection Selection
	marked    Range
}
type patch struct {
	start                           int
	before, insert                  string
	beforeSelection, afterSelection Selection
}

// Model and its Journal belong to one application UI owner. They deliberately
// do not synchronize: perform worker IO using detached immutable snapshots.
// Native adapters must use this journal instead of a second native undo stack.
// A Model must not be copied after creation; use snapshots or New for new models.
type Model struct {
	self                                   *Model
	text                                   string
	selection                              Selection
	preview                                *preedit
	maxBytes                               int
	journal                                *history.Journal
	revision, visibleEpoch, selectionEpoch uint64
}

func New(text string, config Config) (*Model, error) {
	if config.MaxBytes < 0 || config.HistoryLimit < 0 {
		return nil, fmt.Errorf("editor: nonnegative configuration required")
	}
	if config.MaxBytes == 0 {
		config.MaxBytes = DefaultMaxBytes
	}
	if config.HistoryLimit == 0 {
		config.HistoryLimit = DefaultHistoryLimit
	}
	if err := validText(text, config.MaxBytes); err != nil {
		return nil, err
	}
	j, err := history.New(config.HistoryLimit)
	if err != nil {
		return nil, err
	}
	m := &Model{text: text, maxBytes: config.MaxBytes, journal: j, revision: 1, visibleEpoch: 1, selectionEpoch: 1}
	m.self = m
	return m, nil
}

// Validate safely refuses nil, zero, and copied models for UI admission.
func (m *Model) Validate() error {
	if m == nil || m.self != m || m.journal == nil || m.maxBytes < 1 || m.revision == 0 || m.visibleEpoch == 0 || m.selectionEpoch == 0 {
		return ErrInvalidModel
	}
	return nil
}

// Journal is the sole undo and saved-state authority. It returns nil for an
// uninitialized or copied model. Undo/Redo refuse active composition without moving history.
func (m *Model) Journal() *history.Journal {
	if m.Validate() != nil {
		return nil
	}
	return m.journal
}

// MaxBytes returns the document byte budget, or zero for an uninitialized model.
// Native adapters use this bound to refuse platform mutations before admission.
func (m *Model) MaxBytes() int {
	if m.Validate() != nil {
		return 0
	}
	return m.maxBytes
}

// Snapshot returns a zero snapshot for an uninitialized model.
func (m *Model) Snapshot() Snapshot {
	if m.Validate() != nil {
		return Snapshot{}
	}
	s := Snapshot{Text: m.text, VisibleText: m.text, Selection: m.selection, Revision: m.revision, VisibleEpoch: m.visibleEpoch, SelectionEpoch: m.selectionEpoch, History: m.journal.Snapshot()}
	if m.preview != nil {
		s.VisibleText, s.Selection, s.Composing, s.Marked = m.preview.text, m.preview.selection, true, m.preview.marked
	}
	return s
}

func (m *Model) ready() error {
	if err := m.Validate(); err != nil {
		return err
	}
	s := m.journal.Snapshot()
	if s.Busy {
		return history.ErrBusy
	}
	if s.Invalid {
		return history.ErrInvalid
	}
	return nil
}

func validText(text string, limit int) error {
	if len(text) > limit {
		return ErrTooLarge
	}
	if !utf8.ValidString(text) || strings.IndexByte(text, 0) >= 0 {
		return ErrInvalidText
	}
	return nil
}

func validSelection(s Selection, n int) bool {
	return s.Anchor >= 0 && s.Focus >= 0 && s.Anchor <= n && s.Focus <= n
}
func selectionWithin(s Selection, r Range) bool {
	return s.Anchor >= r.Start && s.Focus >= r.Start && s.Anchor <= r.End && s.Focus <= r.End
}
func validRange(r Range, n int) bool { return r.Start >= 0 && r.End >= r.Start && r.End <= n }

func span(text string, r Range) (int, int, error) {
	if !validRange(r, utf8.RuneCountInString(text)) {
		return 0, 0, ErrRange
	}
	start, end, scalar := len(text), len(text), 0
	for b := range text {
		if scalar == r.Start {
			start = b
		}
		if scalar == r.End {
			end = b
			break
		}
		scalar++
	}
	return start, end, nil
}

func (m *Model) replaceSpan(text string, r Range, before, insert string) (string, error) {
	start, end, err := span(text, r)
	if err != nil {
		return "", err
	}
	if text[start:end] != before {
		return "", ErrStale
	}
	if err := validText(insert, m.maxBytes); err != nil {
		return "", err
	}
	if len(insert) > m.maxBytes-(len(text)-(end-start)) {
		return "", ErrTooLarge
	}
	return text[:start] + insert + text[end:], nil
}

func (m *Model) edited(e Edit) (string, error) {
	if e.BaseRevision != m.revision || e.BaseVisibleEpoch != m.visibleEpoch {
		return "", ErrStale
	}
	next, err := m.replaceSpan(m.Snapshot().VisibleText, e.Range, e.Before, e.Insert)
	if err != nil {
		return "", err
	}
	if !validSelection(e.Selection, utf8.RuneCountInString(next)) {
		return "", ErrRange
	}
	return next, nil
}

func (m *Model) clocks(committed, visible, selection bool) error {
	if committed && m.revision == math.MaxUint64 || visible && m.visibleEpoch == math.MaxUint64 || selection && m.selectionEpoch == math.MaxUint64 {
		return ErrRevisionExhausted
	}
	return nil
}

// Apply commits one sparse edit. A selection-only change preserves history and
// redo. Refused edits leave content, selection, clocks, and history unchanged.
func (m *Model) Apply(e Edit) error {
	if err := m.ready(); err != nil {
		return err
	}
	if m.preview != nil {
		return ErrComposition
	}
	next, err := m.edited(e)
	if err != nil {
		return err
	}
	if next == m.text {
		return m.Select(e.Selection)
	}
	p := patch{start: e.Range.Start, before: strings.Clone(e.Before), insert: strings.Clone(e.Insert), beforeSelection: m.selection, afterSelection: e.Selection}
	return m.record(p, false)
}

// Replace is an undoable replacement, retaining its smallest scalar-aligned
// changed span. Sparse Apply additionally validates the caller's captured epochs.
func (m *Model) Replace(text string, selection Selection) error {
	if err := m.ready(); err != nil {
		return err
	}
	if m.preview != nil {
		return ErrComposition
	}
	if err := validText(text, m.maxBytes); err != nil {
		return err
	}
	p := difference(m.text, text)
	return m.Apply(Edit{BaseRevision: m.revision, BaseVisibleEpoch: m.visibleEpoch, Range: Range{p.start, p.start + utf8.RuneCountInString(p.before)}, Before: p.before, Insert: p.insert, Selection: selection})
}

// Preview changes preedit without changing committed Text or history. Multiple
// previews form one composition whose eventual commit is one journal entry.
func (m *Model) Preview(e Edit) error {
	if err := m.ready(); err != nil {
		return err
	}
	next, err := m.edited(e)
	if err != nil {
		return err
	}
	if !validRange(e.Marked, utf8.RuneCountInString(next)) || !selectionWithin(e.Selection, e.Marked) {
		return ErrRange
	}
	s := m.Snapshot()
	textChanged, selChanged := next != s.VisibleText, e.Selection != s.Selection
	if err := m.clocks(false, textChanged, selChanged); err != nil {
		return err
	}
	m.preview = &preedit{text: next, selection: e.Selection, marked: e.Marked}
	if textChanged {
		m.visibleEpoch++
	}
	if selChanged {
		m.selectionEpoch++
	}
	return nil
}

// CommitComposition removes preedit and commits its final bytes exactly once.
// Unmarking need not change visible bytes. A no-net composition adds no history.
// Calling this method when no composition is active is a no-op.
func (m *Model) CommitComposition() error {
	if err := m.ready(); err != nil {
		return err
	}
	if m.preview == nil {
		return nil
	}
	if m.preview.text == m.text {
		m.selection = m.preview.selection
		m.preview = nil
		return nil
	}
	p := difference(m.text, m.preview.text)
	p.beforeSelection, p.afterSelection = m.selection, m.preview.selection
	return m.record(p, true)
}

// CancelComposition restores the committed selection and exact bytes without
// affecting history. Visible and selection clocks still invalidate stale input.
func (m *Model) CancelComposition() error {
	if err := m.ready(); err != nil {
		return err
	}
	if m.preview == nil {
		return nil
	}
	textChanged, selChanged := m.preview.text != m.text, m.preview.selection != m.selection
	if err := m.clocks(false, textChanged, selChanged); err != nil {
		return err
	}
	m.preview = nil
	if textChanged {
		m.visibleEpoch++
	}
	if selChanged {
		m.selectionEpoch++
	}
	return nil
}

// Select changes the visible directional selection, preserving content/history.
func (m *Model) Select(selection Selection) error {
	if err := m.ready(); err != nil {
		return err
	}
	s := m.Snapshot()
	if !validSelection(selection, utf8.RuneCountInString(s.VisibleText)) || m.preview != nil && !selectionWithin(selection, m.preview.marked) {
		return ErrRange
	}
	if selection == s.Selection {
		return nil
	}
	if err := m.clocks(false, false, true); err != nil {
		return err
	}
	if m.preview != nil {
		m.preview.selection = selection
	} else {
		m.selection = selection
	}
	m.selectionEpoch++
	return nil
}

func (m *Model) record(p patch, composition bool) error {
	firstComposition := composition
	return m.journal.Apply(history.Change{Label: "Edit text", Apply: func() error {
		err := m.replay(p, false, firstComposition)
		if err == nil {
			firstComposition = false
		}
		return err
	}, Revert: func() error { return m.replay(p, true, false) }})
}

func (m *Model) replay(p patch, reverse, composition bool) error {
	if m.preview != nil && !composition {
		return ErrComposition
	}
	before, insert, selection := p.before, p.insert, p.afterSelection
	if reverse {
		before, insert, selection = p.insert, p.before, p.beforeSelection
	}
	next, err := m.replaceSpan(m.text, Range{p.start, p.start + utf8.RuneCountInString(before)}, before, insert)
	if err != nil {
		return err
	}
	s := m.Snapshot()
	visibleChanged, selChanged := next != s.VisibleText, selection != s.Selection
	if err := m.clocks(true, visibleChanged, selChanged); err != nil {
		return err
	}
	m.text, m.selection, m.preview = next, selection, nil
	m.revision++
	if visibleChanged {
		m.visibleEpoch++
	}
	if selChanged {
		m.selectionEpoch++
	}
	return nil
}

// difference retains the smallest scalar-aligned contiguous replacement. Clone
// detaches small retained spans from the document backing strings.
func difference(before, after string) patch {
	prefix, start := 0, 0
	for prefix < len(before) && prefix < len(after) {
		r1, n1 := utf8.DecodeRuneInString(before[prefix:])
		r2, n2 := utf8.DecodeRuneInString(after[prefix:])
		if r1 != r2 || n1 != n2 {
			break
		}
		prefix += n1
		start++
	}
	bend, aend := len(before), len(after)
	for bend > prefix && aend > prefix {
		r1, n1 := utf8.DecodeLastRuneInString(before[:bend])
		r2, n2 := utf8.DecodeLastRuneInString(after[:aend])
		if r1 != r2 {
			break
		}
		bend -= n1
		aend -= n2
	}
	return patch{start: start, before: strings.Clone(before[prefix:bend]), insert: strings.Clone(after[prefix:aend])}
}

// Install atomically installs decoded bytes inside Journal.Replace's callback.
// It does not reenter/reset the journal: document.Session performs that once.
// Calling Install without an executing journal callback is refused. The owner
// must use Journal.Replace; the busy gate cannot distinguish callback kinds.
// Even an identical install advances Revision to invalidate old document edits.
func (m *Model) Install(text string, selection Selection) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if !m.journal.Snapshot().Busy {
		return ErrInstallContext
	}
	if m.preview != nil {
		return ErrComposition
	}
	if err := validText(text, m.maxBytes); err != nil {
		return err
	}
	if !validSelection(selection, utf8.RuneCountInString(text)) {
		return ErrRange
	}
	visibleChanged, selChanged := text != m.text, selection != m.selection
	if err := m.clocks(true, visibleChanged, selChanged); err != nil {
		return err
	}
	m.text, m.selection = text, selection
	m.revision++
	if visibleChanged {
		m.visibleEpoch++
	}
	if selChanged {
		m.selectionEpoch++
	}
	return nil
}
