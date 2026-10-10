// Package history provides bounded, UI-owner document undo and redo.
package history

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrEmpty             = errors.New("history: no change available")
	ErrBusy              = errors.New("history: change callback is executing")
	ErrInvalid           = errors.New("history: reset required after a callback panic")
	ErrStale             = errors.New("history: saved revision is no longer current")
	ErrRevisionExhausted = errors.New("history: revision space exhausted")
)

// Change owns immutable before/after model data captured by its callbacks.
// Apply and Revert run synchronously on the application's UI owner. Each must
// either complete its entire model change or return an error without changing
// the model. Validate before assigning; do IO and asynchronous work separately.
// The journal cannot roll back partial mutations or infer their memory cost.
// Callbacks must not reenter this journal. A panic invalidates the journal and
// propagates to the caller; Reset is required if the application recovers.
type Change struct {
	Label  string
	Apply  func() error
	Revert func() error
}

type entry struct {
	change        Change
	before, after uint64
}

// State is an immutable presentation snapshot. Dirty compares revision identity,
// so a new branch cannot impersonate a saved state at the same undo depth.
type State struct {
	UndoLabel, RedoLabel string
	UndoCount, RedoCount int
	Revision             uint64
	Dirty, Busy, Invalid bool
}

// Journal retains at most limit changes, including redo entries. It is owned by
// one Application UI owner; it does not synchronize or clone callback captures.
// The bound counts changes, not the bytes of application-owned captured data.
// A newly created journal represents a clean document.
type Journal struct {
	entries              []entry
	cursor, limit        int
	current, saved, next uint64
	busy, invalid        bool
}

func New(limit int) (*Journal, error) {
	if limit < 1 {
		return nil, fmt.Errorf("history: positive change limit required")
	}
	return &Journal{limit: limit, current: 1, saved: 1, next: 1}, nil
}

func (j *Journal) ready() error {
	if j.busy {
		return ErrBusy
	}
	if j.invalid {
		return ErrInvalid
	}
	if j.limit < 1 {
		return fmt.Errorf("history: use New to initialize a journal")
	}
	return nil
}

func (j *Journal) invoke(callback func() error) (err error) {
	j.busy = true
	completed := false
	defer func() {
		j.busy = false
		if !completed {
			// A panicking callback may have changed the model. Release stale closures
			// and refuse further replay rather than claiming a known history state.
			j.entries = nil
			j.cursor = 0
			j.invalid = true
		}
	}()
	err = callback()
	completed = true
	return err
}

// Apply executes a new change. Failed admission or callback preserves both
// stacks. A successful new change discards redo and evicts the oldest changes.
func (j *Journal) Apply(change Change) error {
	if err := j.ready(); err != nil {
		return err
	}
	if strings.TrimSpace(change.Label) == "" || change.Apply == nil || change.Revert == nil {
		return fmt.Errorf("history: label, apply and revert required")
	}
	if j.next == math.MaxUint64 {
		return ErrRevisionExhausted
	}
	if err := j.invoke(change.Apply); err != nil {
		return err
	}
	j.next++
	e := entry{change: change, before: j.current, after: j.next}
	clear(j.entries[j.cursor:])
	j.entries = j.entries[:j.cursor]
	if len(j.entries) == j.limit {
		copy(j.entries, j.entries[1:])
		j.entries[len(j.entries)-1] = entry{}
		j.entries = j.entries[:len(j.entries)-1]
	}
	j.entries = append(j.entries, e)
	j.cursor = len(j.entries)
	j.current = e.after
	return nil
}

func (j *Journal) Undo() error {
	if err := j.ready(); err != nil {
		return err
	}
	if j.cursor == 0 {
		return ErrEmpty
	}
	e := j.entries[j.cursor-1]
	if err := j.invoke(e.change.Revert); err != nil {
		return err
	}
	j.cursor--
	j.current = e.before
	return nil
}

func (j *Journal) Redo() error {
	if err := j.ready(); err != nil {
		return err
	}
	if j.cursor == len(j.entries) {
		return ErrEmpty
	}
	e := j.entries[j.cursor]
	if err := j.invoke(e.change.Apply); err != nil {
		return err
	}
	j.cursor++
	j.current = e.after
	return nil
}

func (j *Journal) Snapshot() State {
	s := State{Revision: j.current, UndoCount: j.cursor, RedoCount: len(j.entries) - j.cursor,
		Dirty: j.invalid || j.current != j.saved, Busy: j.busy, Invalid: j.invalid}
	if s.UndoCount > 0 {
		s.UndoLabel = j.entries[j.cursor-1].change.Label
	}
	if s.RedoCount > 0 {
		s.RedoLabel = j.entries[j.cursor].change.Label
	}
	return s
}

// MarkSaved acknowledges successful persistence of the current owner-held model.
// An asynchronous save must first verify the captured model revision is still
// current. A stale save completion must not call this method.
func (j *Journal) MarkSaved() error {
	if err := j.ready(); err != nil {
		return err
	}
	return j.MarkSavedRevision(j.current)
}

// MarkSavedRevision accepts a completed save only while its captured revision
// is still current. Invalidation and document replacement also reject stale saves.
func (j *Journal) MarkSavedRevision(revision uint64) error {
	if err := j.ready(); err != nil {
		return err
	}
	if revision != j.current {
		return ErrStale
	}
	j.saved = revision
	return nil
}

// Clear releases all undo/redo captures, preserving current dirty status.
// Reset is used instead when replacing the model or recovering after a panic.
func (j *Journal) Clear() error {
	if err := j.ready(); err != nil {
		return err
	}
	j.entries = nil
	j.cursor = 0
	return nil
}

// Reset releases all captures and represents a replaced document. saved says
// whether that new model is already persisted. It does not modify model data.
func (j *Journal) Reset(saved bool) error {
	return j.Replace(saved, func() error { return nil })
}

// Replace installs a complete replacement model, then releases history and
// advances revision identity. The install callback must be atomic on error and
// must not reenter the journal. Failed admission or installation preserves both
// history stacks. Like Reset, Replace can recover an invalid journal explicitly.
func (j *Journal) Replace(saved bool, install func() error) error {
	if j.busy {
		return ErrBusy
	}
	if j.limit < 1 {
		return fmt.Errorf("history: use New to initialize a journal")
	}
	if j.next == math.MaxUint64 {
		return ErrRevisionExhausted
	}
	if install == nil {
		return fmt.Errorf("history: replacement callback required")
	}
	if err := j.invoke(install); err != nil {
		return err
	}
	j.next++
	j.current = j.next
	j.saved = 0
	if saved {
		j.saved = j.current
	}
	j.entries = nil
	j.cursor = 0
	j.invalid = false
	return nil
}
