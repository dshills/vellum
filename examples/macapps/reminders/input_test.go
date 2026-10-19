package reminders

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/ui"
)

type heldField struct {
	id   string
	ctx  context.Context
	done func(error)
}
type heldInput struct {
	m     *Model
	calls []*heldField
	next  int
}

func installHeld(t *testing.T, m *Model, ctx context.Context) *heldInput {
	t.Helper()
	h := &heldInput{m: m}
	if err := m.SetInputBoundary(ctx, func(ctx context.Context, id string, done func(error)) error {
		h.calls = append(h.calls, &heldField{id: id, ctx: ctx, done: done})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return h
}
func (h *heldInput) deliver(t *testing.T, text string) *heldField {
	t.Helper()
	if h.next >= len(h.calls) {
		t.Fatal("no admitted input field")
	}
	call := h.calls[h.next]
	h.next++
	find(t, h.m.View(), call.id).Change(ui.ChangeEvent{Text: text})
	call.done(nil)
	return call
}
func (h *heldInput) settle(t *testing.T, values map[string]string) {
	t.Helper()
	for h.m.PendingInput() {
		if h.next >= len(h.calls) {
			t.Fatal("pending action without admitted field")
		}
		id := h.calls[h.next].id
		value, ok := values[id]
		if !ok {
			value = find(t, h.m.View(), id).Text
		}
		h.deliver(t, value)
	}
}

func TestInputBoundaryCommitsComposerBeforeExactlyOneAdd(t *testing.T) {
	m := New()
	h := installHeld(t, m, context.Background())
	add := find(t, m.View(), "reminders-add")
	if add.Disabled {
		t.Fatal("stale blank model disabled Add before native settlement")
	}
	add.Activate()
	if !m.PendingInput() || len(m.Items()) != 12 || len(h.calls) != 1 {
		t.Fatal("Add ran before field publication")
	}
	if err := m.Select("work"); !errors.Is(err, ErrInputPending) {
		t.Fatal("mutation bypassed pending structural guard", err)
	}
	add.Activate()
	if len(h.calls) != 1 || len(m.Items()) != 12 {
		t.Fatal("second activation queued another action")
	}
	first := h.deliver(t, "")
	first.done(nil)
	if len(h.calls) != 2 {
		t.Fatal("duplicate callback advanced settlement twice")
	}
	last := h.deliver(t, "Fresh native title")
	last.done(nil)
	if m.PendingInput() || len(m.Items()) != 13 || m.Items()[12].Title != "Fresh native title" || m.History().UndoCount != 1 || m.composer != "" {
		t.Fatal("Add did not use settled composer exactly once", m.Items(), m.History())
	}
	if got := []string{h.calls[0].id, h.calls[1].id}; !reflect.DeepEqual(got, []string{"reminders-search", "reminders-new-title"}) {
		t.Fatal("wrong mounted fields", got)
	}
	checkView(t, m)
}

func TestInputBoundaryDoneAndCancelSequenceMountedDraftFields(t *testing.T) {
	for _, commit := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancel", true: "done"}[commit], func(t *testing.T) {
			m := New()
			m.openDetails("r1")
			h := installHeld(t, m, context.Background())
			id := "details-cancel"
			if commit {
				id = "details-save"
			}
			find(t, m.View(), id).Activate()
			h.settle(t, map[string]string{"details-title": "Native draft title", "details-notes": "Native draft notes", "details-date": Today})
			var fields []string
			for _, call := range h.calls {
				fields = append(fields, call.id)
			}
			if !reflect.DeepEqual(fields, []string{"reminders-search", "reminders-new-title", "details-title", "details-notes", "details-date"}) {
				t.Fatal("settled absent/missing fields", fields)
			}
			if m.editing != "" || m.PendingInput() {
				t.Fatal("completion retained dialog/action")
			}
			r := item(t, m, "r1")
			if commit {
				if r.Title != "Native draft title" || r.Notes != "Native draft notes" || r.Due != Today || m.History().UndoCount != 1 {
					t.Fatal("Done did not commit settled fields atomically", r, m.History())
				}
			} else if r.Title != "Croissants" || m.History().UndoCount != 0 {
				t.Fatal("Cancel committed draft", r, m.History())
			}
			checkView(t, m)
		})
	}
}

func TestInputBoundaryUndoCommitsDraftBeforeHistoryReplay(t *testing.T) {
	m := New()
	if err := m.SetCompleted("r1", true); err != nil {
		t.Fatal(err)
	}
	m.openDetails("r1")
	h := installHeld(t, m, context.Background())
	if err := m.Commands().Invoke(context.Background(), "reminders.undo"); err != nil {
		t.Fatal(err)
	}
	if m.History().UndoCount != 1 || !item(t, m, "r1").Completed {
		t.Fatal("Undo ran before settlement")
	}
	if err := m.Undo(); !errors.Is(err, ErrInputPending) {
		t.Fatal("second Undo bypassed pending guard", err)
	}
	h.settle(t, map[string]string{"details-title": "Pending native edit"})
	if r := item(t, m, "r1"); r.Title != "Croissants" || !r.Completed || m.History().UndoCount != 1 || m.History().RedoCount != 1 {
		t.Fatal("Undo reverted earlier completion instead of newly committed draft", r, m.History())
	}
	if err := m.Redo(); err != nil {
		t.Fatal(err)
	}
	h.settle(t, nil)
	if r := item(t, m, "r1"); r.Title != "Pending native edit" || !r.Completed {
		t.Fatal("Redo did not restore stable reminder edit", r)
	}
}

func TestInputBoundaryRefusalValidationAndSelection(t *testing.T) {
	t.Run("immediate refusal", func(t *testing.T) {
		m := New()
		refusal := errors.New("boundary refused")
		if err := m.SetInputBoundary(context.Background(), func(context.Context, string, func(error)) error { return refusal }); err != nil {
			t.Fatal(err)
		}
		if err := m.AfterInput(func() error { _, err := m.Add("No"); return err }); !errors.Is(err, refusal) || m.PendingInput() || len(m.Items()) != 12 {
			t.Fatal("refusal mutated or retained action", err)
		}
	})
	t.Run("deferred refusal", func(t *testing.T) {
		m := New()
		h := installHeld(t, m, context.Background())
		find(t, m.View(), "reminders-add").Activate()
		h.calls[0].done(errors.New("composition unresolved"))
		if m.PendingInput() || len(m.Items()) != 12 || m.status != "composition unresolved" {
			t.Fatal("deferred refusal did not preserve collection")
		}
	})
	t.Run("invalid native text", func(t *testing.T) {
		m := New()
		h := installHeld(t, m, context.Background())
		find(t, m.View(), "reminders-add").Activate()
		h.settle(t, map[string]string{"reminders-new-title": strings.Repeat("x", MaxTitleBytes+1)})
		if len(m.Items()) != 12 || m.History().UndoCount != 0 || m.PendingInput() || m.status == "" {
			t.Fatal("rejected native field allowed structural action")
		}
	})
	t.Run("selection preserves composer", func(t *testing.T) {
		m := New()
		h := installHeld(t, m, context.Background())
		find(t, m.View(), "list-work").Activate()
		if m.Selected() != "groceries" {
			t.Fatal("selection changed before settlement")
		}
		h.settle(t, map[string]string{"reminders-new-title": "Unsubmitted draft"})
		if m.Selected() != "work" || m.composer != "Unsubmitted draft" || len(m.Items()) != 12 {
			t.Fatal("selection lost native composer or created reminder")
		}
	})
}

func TestInputBoundaryWindowLifetimeAndCloseRejectLateCompletion(t *testing.T) {
	for _, closeModel := range []bool{false, true} {
		t.Run(map[bool]string{false: "window canceled", true: "model closed"}[closeModel], func(t *testing.T) {
			m := New()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			h := installHeld(t, m, ctx)
			staleSearch := find(t, m.View(), "reminders-search")
			staleComposer := find(t, m.View(), "reminders-new-title")
			find(t, m.View(), "reminders-add").Activate()
			call := h.calls[0]
			if closeModel {
				m.Close()
			} else {
				cancel()
			}
			staleSearch.Change(ui.ChangeEvent{Text: "Late query"})
			staleComposer.Change(ui.ChangeEvent{Text: "Late title"})
			call.done(nil)
			call.done(nil)
			if m.PendingInput() || len(m.Items()) != 12 || m.History().UndoCount != 0 || m.Search() != "" || m.composer != "" {
				t.Fatal("late completion/Change mutated retired lifetime")
			}
			if err := m.AfterInput(func() error { _, err := m.Add("Late"); return err }); err == nil {
				t.Fatal("retired lifetime admitted another action")
			}
			if closeModel && call.ctx.Err() == nil {
				t.Fatal("Close did not cancel admitted boundary context")
			}
		})
	}
}

func TestRedoChangedDraftIsNonmutating(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		t.Run(map[bool]string{false: "changed title", true: "invalid date"}[invalid], func(t *testing.T) {
			m := New()
			if err := m.SetCompleted("r1", true); err != nil {
				t.Fatal(err)
			}
			if err := m.Undo(); err != nil {
				t.Fatal(err)
			}
			m.openDetails("r1")
			if invalid {
				m.draft.Due = "2026-02-30"
			} else {
				m.draft.Title = "Unsaved title"
			}
			beforeItems, beforeHistory, beforeDraft, beforeEditing := m.Items(), m.History(), m.draft, m.editing
			if err := m.Redo(); !errors.Is(err, ErrDraftForRedo) {
				t.Fatal("Redo did not require explicit draft decision", err)
			}
			if !slices.Equal(beforeItems, m.Items()) || beforeHistory != m.History() || beforeDraft != m.draft || beforeEditing != m.editing {
				t.Fatal("failed Redo mutated collection/history/draft", m.Items(), m.History(), m.draft)
			}
			if !strings.Contains(m.status, "Done or Cancel") {
				t.Fatal("missing actionable refusal", m.status)
			}
		})
	}
}

func TestRedoSettlesNativeDraftBeforeNonmutatingRefusal(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		t.Run(map[bool]string{false: "changed title", true: "invalid date"}[invalid], func(t *testing.T) {
			m := New()
			if err := m.SetCompleted("r1", true); err != nil {
				t.Fatal(err)
			}
			if err := m.Undo(); err != nil {
				t.Fatal(err)
			}
			m.openDetails("r1")
			h := installHeld(t, m, context.Background())
			beforeItems, beforeHistory := m.Items(), m.History()
			if err := m.Redo(); err != nil {
				t.Fatal(err)
			}
			if !m.PendingInput() || beforeHistory != m.History() {
				t.Fatal("Redo bypassed native settlement")
			}
			values := map[string]string{"details-title": "Native unsaved title"}
			if invalid {
				values = map[string]string{"details-date": "2026-02-30"}
			}
			h.settle(t, values)
			if m.PendingInput() || !slices.Equal(beforeItems, m.Items()) || beforeHistory != m.History() || m.editing != "r1" {
				t.Fatal("native draft refusal changed collection/history/open state")
			}
			if invalid && m.draft.Due != "2026-02-30" || !invalid && m.draft.Title != "Native unsaved title" {
				t.Fatal("refusal discarded settled draft")
			}
			if m.status != ErrDraftForRedo.Error() {
				t.Fatal("missing deferred draft refusal", m.status)
			}
		})
	}
}

func TestRedoUnchangedOpenDraftCanReplayExistingHistory(t *testing.T) {
	m := New()
	if err := m.SetCompleted("r1", true); err != nil {
		t.Fatal(err)
	}
	if err := m.Undo(); err != nil {
		t.Fatal(err)
	}
	m.openDetails("r1")
	h := installHeld(t, m, context.Background())
	if err := m.Redo(); err != nil {
		t.Fatal(err)
	}
	h.settle(t, nil)
	if !item(t, m, "r1").Completed || m.editing != "" || m.History().RedoCount != 0 || m.History().UndoCount != 1 {
		t.Fatal("unchanged draft blocked or replaced existing redo", m.History())
	}
}
