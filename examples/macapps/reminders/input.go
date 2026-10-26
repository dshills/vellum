package reminders

import (
	"context"
	"errors"
	"github.com/dshills/vellum/ui"
)

var (
	ErrInputPending = errors.New("reminders: finish the pending input action first")
	ErrClosed       = errors.New("reminders: model is closed")
	ErrDraftForRedo = errors.New("reminders: choose Done or Cancel for the changed details before Redo")
)

// InputBoundary settles one currently mounted TextField. It publishes its
// controlled Change before done, on the same UI owner. Immediate refusal must
// return an error without calling done; accepted completion may be asynchronous.
// Native hosts can implement this with cogent.SettleTextField. No editor/Core
// type is required by this portable model.
type InputBoundary func(context.Context, string, func(error)) error

type inputAction struct {
	ctx             context.Context
	cancel          context.CancelFunc
	fields          []string
	next            int
	action          func() error
	done            bool
	err, validation error
}

// SetInputBoundary installs the optional native settlement boundary and its
// window lifetime. Call on the UI owner before running the host. Without one,
// model/routing tests and other already-committed consumers execute immediately.
func (m *Model) SetInputBoundary(lifetime context.Context, boundary InputBoundary) error {
	if m.closed {
		return ErrClosed
	}
	if m.pending != nil {
		return ErrInputPending
	}
	if lifetime == nil || boundary == nil {
		return errors.New("reminders: input lifetime and boundary are required")
	}
	if err := lifetime.Err(); err != nil {
		return err
	}
	m.inputLifetime = lifetime
	m.inputBoundary = boundary
	return nil
}

// Close releases the single pending action and refuses late native callbacks.
// It must run on the UI owner after the native host stops dispatching.
func (m *Model) Close() {
	m.closed = true
	for _, c := range m.checklists {
		c.Disabled = true
		c.Change = nil
		c.Items = nil
	}
	m.checklists = nil
	if m.schedule != nil {
		m.schedule.Disabled = true
		m.schedule.Change = nil
		m.schedule.Invalid = nil
	}
	m.schedule = nil
	if op := m.pending; op != nil {
		op.done = true
		op.action = nil
		op.fields = nil
		op.cancel()
		m.pending = nil
	}
	m.inputBoundary = nil
	m.inputLifetime = nil
}

func (m *Model) PendingInput() bool { return m.pending != nil }

func (m *Model) inputAlive() bool {
	return !m.closed && (m.inputLifetime == nil || m.inputLifetime.Err() == nil)
}

func (m *Model) mutationReady() error {
	if m.closed {
		return ErrClosed
	}
	if m.inputLifetime != nil {
		if err := m.inputLifetime.Err(); err != nil {
			return err
		}
	}
	if m.pending != nil {
		return ErrInputPending
	}
	return nil
}

func (m *Model) inputError(err error) {
	if !m.inputAlive() {
		return
	}
	if err != nil && m.pending != nil {
		m.pending.validation = err
	}
	m.accept(err)
}

// AfterInput admits at most one structural action, after all currently mounted
// fields have published their committed values. The callback reads the model
// at execution time, not stale values captured before settlement. Immediate
// refusal is returned; deferred errors appear in the view's status. Ordinary
// controlled Change callbacks remain available during settlement.
func (m *Model) AfterInput(action func() error) error {
	if err := m.mutationReady(); err != nil {
		m.accept(err)
		return err
	}
	if action == nil {
		return errors.New("reminders: input action required")
	}
	if m.inputBoundary == nil {
		err := action()
		m.accept(err)
		return err
	}
	if err := m.inputLifetime.Err(); err != nil {
		m.accept(err)
		return err
	}
	ctx, cancel := context.WithCancel(m.inputLifetime)
	op := &inputAction{ctx: ctx, cancel: cancel, action: action, fields: []string{"reminders-search", "reminders-new-title"}}
	if m.selected == "scheduled" && m.rangeExpanded && m.schedule != nil {
		op.fields = append(op.fields, ui.DateRangeStartID(m.schedule.ID), ui.DateRangeEndID(m.schedule.ID))
	}
	if m.editing != "" {
		op.fields = append(op.fields, "details-title", "details-notes", "details-date")
	}
	m.pending = op
	m.status = ""
	finish := func(err error) {
		if m.pending != op || op.done {
			return
		}
		if err == nil {
			err = op.ctx.Err()
		}
		if err == nil {
			err = op.validation
		}
		op.done = true
		m.pending = nil
		op.cancel()
		fn := op.action
		op.action = nil
		op.fields = nil
		if err == nil {
			err = fn()
		}
		op.err = err
		m.accept(err)
	}
	var advance func()
	advance = func() {
		if m.pending != op || op.done {
			return
		}
		if err := op.ctx.Err(); err != nil {
			finish(err)
			return
		}
		if op.next == len(op.fields) {
			finish(nil)
			return
		}
		id := op.fields[op.next]
		op.next++
		completed := false
		err := m.inputBoundary(op.ctx, id, func(err error) {
			if completed || m.pending != op || op.done {
				return
			}
			completed = true
			if err != nil {
				finish(err)
				return
			}
			advance()
		})
		if err != nil && !completed {
			completed = true
			finish(err)
		}
	}
	advance()
	if op.done {
		return op.err
	}
	return nil
}

func (m *Model) saveDraft() error {
	if m.editing == "" {
		return nil
	}
	return m.Update(m.editing, m.draft.Title, m.draft.Notes, m.draft.Due, m.draft.Priority, m.draft.Flagged)
}

// Redo may not create a new edit branch while checking its own availability.
// The draft remains open for an explicit Done/Cancel decision.
func (m *Model) draftBlocksRedo() bool {
	if m.editing == "" {
		return false
	}
	at := m.index(m.editing)
	if at < 0 || !validReminder(m.draft) {
		return true
	}
	r := m.items[at]
	return m.draft.Title != r.Title || m.draft.Notes != r.Notes || m.draft.Due != r.Due || m.draft.Priority != r.Priority || m.draft.Flagged != r.Flagged
}

func (m *Model) doneDetails() {
	_ = m.AfterInput(func() error {
		if err := m.saveDraft(); err != nil {
			return err
		}
		m.editing = ""
		return nil
	})
}
func (m *Model) cancelDetails() { _ = m.AfterInput(func() error { m.editing = ""; return nil }) }
