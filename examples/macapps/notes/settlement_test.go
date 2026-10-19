package notes

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

func deferSettlement(t *testing.T, m *Model) (*func(error), *int) {
	t.Helper()
	var complete func(error)
	calls := 0
	must(t, m.SetSettlement(func(callback func(error)) error { calls++; complete = callback; return nil }))
	return &complete, &calls
}

func TestSettlementAdmitsPendingTitleBeforeUndoExactlyOnce(t *testing.T) {
	m := New()
	title, _ := find(m.View(), "note-title-n1")
	title.Change(ui.ChangeEvent{Text: "Native sample"})
	complete, calls := deferSettlement(t, m)
	must(t, m.Commands().Invoke(context.Background(), "notes.undo"))
	if !m.Pending() || *calls != 1 || m.lookup("n1").value.Title != "Native sample" {
		t.Fatal("Undo ran before settlement")
	}
	// This represents the public controlled Change delivered by the native FIFO
	// barrier. It must remain admissible while structural actions are pending.
	title.Change(ui.ChangeEvent{Text: "A fictional note created during native UI testing.Native sample"})
	if m.lookup("n1").value.Title == "Native sample" || m.status != "Finishing text input" || m.History().UndoCount != 2 {
		t.Fatal("pending native title commit was rejected", m.status, m.History())
	}
	(*complete)(nil)
	if m.Pending() || m.lookup("n1").value.Title != "Native sample" || m.History().UndoCount != 1 || m.History().RedoCount != 1 {
		t.Fatal("Undo targeted an older title", m.Items(), m.History())
	}
	(*complete)(nil)
	if m.lookup("n1").value.Title != "Native sample" || m.History().UndoCount != 1 {
		t.Fatal("duplicate completion repeated Undo")
	}
}

func TestEmptyHistoryUndoStillSettlesUnadmittedInput(t *testing.T) {
	m := New()
	before := m.lookup("n1").value.Title
	complete, _ := deferSettlement(t, m)
	if m.History().UndoCount != 0 {
		t.Fatal("fixture")
	}
	must(t, m.Commands().Invoke(context.Background(), "notes.undo"))
	title, _ := find(m.View(), "note-title-n1")
	title.Change(ui.ChangeEvent{Text: "Typed but not previously published"})
	(*complete)(nil)
	if m.lookup("n1").value.Title != before || m.History().RedoCount != 1 {
		t.Fatal("history availability rejected native pending edit")
	}
}

func TestPendingSearchPublishesBeforeQueryClearingStructuralAction(t *testing.T) {
	for _, emptySelection := range []bool{false, true} {
		m := New()
		if emptySelection {
			for len(m.notes) > 0 {
				must(t, m.Delete(m.notes[0].value.ID))
			}
			if id, body := m.CurrentInput(); id != "" || body != nil {
				t.Fatal("empty selection exposes stale input")
			}
		}
		complete, calls := deferSettlement(t, m)
		before := len(m.Items())
		must(t, m.Commands().Invoke(context.Background(), "notes.new"))
		search, _ := find(m.View(), "notes-search")
		if search.Disabled || !m.Pending() || *calls != 1 {
			t.Fatal("search cannot publish while structural request waits")
		}
		search.Change(ui.ChangeEvent{Text: "Query still in the native buffer before the boundary"})
		if m.query == "" || len(m.Items()) != before {
			t.Fatal("search publication did not precede the new-note action")
		}
		(*complete)(nil)
		if m.Pending() || m.query != "" || len(m.Items()) != before+1 || m.Selected() == "" || len(m.Visible()) == 0 {
			t.Fatal("unadmitted query re-filtered or hid new note", emptySelection, m.query, m.Items())
		}
	}
}

func TestPendingRequestBoundsActionsAndKeepsNativeChangesAdmissible(t *testing.T) {
	m := New()
	complete, calls := deferSettlement(t, m)
	before := m.Items()
	must(t, m.Request(func() error { return m.Select("n2") }))
	if !errors.Is(m.Request(func() error { return m.Delete("n1") }), ErrPending) || !errors.Is(m.Delete("n1"), ErrPending) || !errors.Is(m.SetSettlement(nil), ErrPending) || *calls != 1 {
		t.Fatal("pending boundary admitted another structural action")
	}
	for _, id := range []string{"notes-new", "notes-delete", "notes-pin", "notes-undo", "notes-redo", "folder-projects", "note-select-n2"} {
		n, ok := find(m.View(), id)
		if !ok || !n.Disabled {
			t.Fatal("structural action stays enabled", id)
		}
	}
	n := m.lookup("n1")
	must(t, n.editor.Replace("Native body committed while pending", editor.Selection{}))
	body, _ := find(m.View(), "note-editor-n1")
	if body.Disabled {
		t.Fatal("body disabled before its settlement")
	}
	body.Change(ui.ChangeEvent{})
	if n.value.Body != "Native body committed while pending" || !m.Pending() {
		t.Fatal("pending body callback was rejected")
	}
	(*complete)(nil)
	if m.Selected() != "n2" || len(m.Items()) != len(before) || m.Pending() {
		t.Fatal("deferred selection failed")
	}
}

func TestSettlementErrorAndImmediateRefusalNeverRunAction(t *testing.T) {
	refused := errors.New("native input refused")
	for _, immediate := range []bool{false, true} {
		m := New()
		before := m.Items()
		var complete func(error)
		must(t, m.SetSettlement(func(callback func(error)) error {
			complete = callback
			if immediate {
				return refused
			}
			return nil
		}))
		err := m.Request(func() error { return m.Delete("n1") })
		if immediate {
			if !errors.Is(err, refused) {
				t.Fatal("immediate refusal was hidden")
			}
		} else {
			must(t, err)
			complete(refused)
		}
		complete(nil) // Defensive protection against a buggy or late native callback.
		if m.Pending() || !slices.Equal(before, m.Items()) || m.status != refused.Error() {
			t.Fatal("failed boundary changed data", immediate, m.status)
		}
	}
}

func TestTitleCommitDuringBodyCompositionDoesNotDiscardEither(t *testing.T) {
	m := New()
	n := m.lookup("n1")
	s := n.editor.Snapshot()
	must(t, n.editor.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Insert: "あ", Selection: editor.Selection{Anchor: 1, Focus: 1}, Marked: editor.Range{Start: 0, End: 1}}))
	preedit := n.editor.Snapshot()
	complete, _ := deferSettlement(t, m)
	must(t, m.Request(func() error { return m.Delete("n1") }))
	title, _ := find(m.View(), "note-title-n1")
	title.Change(ui.ChangeEvent{Text: "Title committed before body acknowledgment"})
	if n.value.Title != "Title committed before body acknowledgment" || n.editor.Snapshot() != preedit {
		t.Fatal("title admission was coupled to independent body composition")
	}
	(*complete)(editor.ErrComposition)
	if m.Pending() || m.lookup("n1") != n || n.editor.Snapshot() != preedit {
		t.Fatal("composition refusal lost the note or preedit")
	}
}

func TestRejectedNativeTitleDuringSettlementDoesNotUndoOlderHistory(t *testing.T) {
	m := New()
	title, _ := find(m.View(), "note-title-n1")
	title.Change(ui.ChangeEvent{Text: "Previously committed title"})
	state := m.History()
	complete, _ := deferSettlement(t, m)
	must(t, m.Request(m.Undo))
	title.Change(ui.ChangeEvent{Text: strings.Repeat("x", MaxTitleBytes+1)})
	(*complete)(nil)
	if m.Pending() || m.lookup("n1").value.Title != "Previously committed title" || m.History() != state || m.status != ErrInvalid.Error() {
		t.Fatal("rejected native input caused an older undo", m.status, m.History())
	}
}

func TestEveryStructuralViewRoutePassesThroughSettlement(t *testing.T) {
	for _, kind := range []string{"new", "delete", "select", "folder", "pin", "undo", "redo"} {
		t.Run(kind, func(t *testing.T) {
			m := New()
			if kind == "undo" || kind == "redo" {
				must(t, m.Update("n1", "Changed", m.lookup("n1").value.Body))
				if kind == "redo" {
					must(t, m.Undo())
				}
			}
			before, selected, folder := m.Items(), m.Selected(), m.SelectedFolder()
			complete, calls := deferSettlement(t, m)
			id := map[string]string{"new": "notes-new", "delete": "notes-delete", "select": "note-select-n2", "folder": "folder-projects", "pin": "notes-pin", "undo": "notes-undo", "redo": "notes-redo"}[kind]
			n, _ := find(m.View(), id)
			if n.Disabled {
				t.Fatal("route unexpectedly disabled")
			}
			if kind == "pin" {
				n.Change(ui.ChangeEvent{Checked: false})
			} else {
				n.Activate()
			}
			if !m.Pending() || *calls != 1 || !slices.Equal(before, m.Items()) || selected != m.Selected() || folder != m.SelectedFolder() {
				t.Fatal("route bypassed boundary", kind)
			}
			(*complete)(nil)
			if m.Pending() || slices.Equal(before, m.Items()) && selected == m.Selected() && folder == m.SelectedFolder() {
				t.Fatal("deferred action did not run", kind)
			}
		})
	}
}

func TestSynchronousBoundaryAndNilBoundaryRemainPortable(t *testing.T) {
	m := New()
	calls := 0
	must(t, m.SetSettlement(func(done func(error)) error { done(nil); done(nil); return nil }))
	must(t, m.Request(func() error { calls++; return nil }))
	if calls != 1 || m.Pending() {
		t.Fatal("synchronous boundary duplicated action")
	}
	must(t, m.SetSettlement(nil))
	must(t, m.Request(func() error { calls++; return nil }))
	if calls != 2 || m.Pending() {
		t.Fatal("nil boundary did not execute directly")
	}
}
