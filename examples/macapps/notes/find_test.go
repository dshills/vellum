package notes

import (
	"context"
	"errors"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"slices"
	"testing"
)

func findFixture(t *testing.T) *Model {
	t.Helper()
	m := New()
	must(t, m.Update("n1", "Fixture", "A😀 plan PLAN e\u0301 plan"))
	return m
}
func TestFindUnicodeNavigationAndReplaceAllUsesDocumentHistory(t *testing.T) {
	m := findFixture(t)
	before := m.Items()
	state := m.History()
	must(t, m.Commands().Invoke(context.Background(), "notes.find"))
	if m.View().Focus.ID != ui.FindBarInputID(findID) {
		t.Fatal("query focus")
	}
	m.find.Query = "plan"
	m.find.Replacement = "drawing"
	must(t, m.Commands().Invoke(context.Background(), "notes.find-next"))
	if got := m.lookup("n1").editor.Snapshot().Selection; got != (editor.Selection{Anchor: 3, Focus: 7}) {
		t.Fatal(got)
	}
	if m.History() != state {
		t.Fatal("navigation dirtied history")
	}
	must(t, m.Commands().Invoke(context.Background(), "notes.find-previous"))
	if got := m.lookup("n1").editor.Snapshot().Selection; got != (editor.Selection{Anchor: 16, Focus: 20}) {
		t.Fatal(got)
	}
	must(t, m.requestFind(ui.FindReplaceAll, "n1"))
	if m.lookup("n1").value.Body != "A😀 drawing drawing e\u0301 drawing" || m.History().UndoCount != state.UndoCount+1 {
		t.Fatal("replace not one app operation", m.History())
	}
	must(t, m.Undo())
	if !slices.Equal(m.Items(), before) {
		t.Fatal("Undo differs")
	}
	must(t, m.Redo())
	if m.lookup("n1").value.Body != "A😀 drawing drawing e\u0301 drawing" {
		t.Fatal("Redo differs")
	}
	must(t, m.Commands().Invoke(context.Background(), "notes.find-close"))
	if m.find.Open || m.find.Model != nil || m.View().Focus.ID != "note-editor-n1" {
		t.Fatal("close focus/reference")
	}
	if m.Commands().Enabled("notes.find-close") == nil {
		t.Fatal("closed Escape consumed")
	}
}
func TestFindSettlementReadsLatestControlledFieldsOnce(t *testing.T) {
	m := findFixture(t)
	must(t, m.openFind(true))
	m.find.Query = "missing"
	m.find.Replacement = "old"
	complete, _ := deferSettlement(t, m)
	before := m.History().UndoCount
	must(t, m.requestFind(ui.FindReplaceAll, "n1"))
	pending := m.View()
	query, _ := find(pending, ui.FindBarInputID(findID))
	replacement, _ := find(pending, findID+"/find/replacement")
	query.Change(ui.ChangeEvent{Text: "plan"})
	replacement.Change(ui.ChangeEvent{Text: "new"})
	(*complete)(nil)
	(*complete)(nil)
	if m.lookup("n1").value.Body != "A😀 new new e\u0301 new" || m.History().UndoCount != before+1 {
		t.Fatal("stale fields or duplicate replace")
	}
}
func TestFindRejectsMissingStaleAndCompositionWithoutMutation(t *testing.T) {
	m := findFixture(t)
	must(t, m.openFind(false))
	m.find.Query = "missing"
	before := m.Items()
	state := m.History()
	if err := m.requestFind(ui.FindReplaceAll, "n1"); err != ui.ErrFindNoMatch {
		t.Fatal(err)
	}
	if !slices.Equal(m.Items(), before) || m.History() != state {
		t.Fatal("no-match changed document")
	}
	m.find.Query = "plan"
	must(t, m.Select("n2"))
	if err := m.requestFind(ui.FindReplaceAll, "n1"); err != ErrInvalid {
		t.Fatal("stale target", err)
	}
	n := m.lookup("n2")
	s := n.editor.Snapshot()
	must(t, n.editor.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: editor.Range{}, Insert: "marked", Selection: editor.Selection{Anchor: 6, Focus: 6}, Marked: editor.Range{Start: 0, End: 6}}))
	if err := m.requestFind(ui.FindNext, "n2"); !errors.Is(err, editor.ErrComposition) {
		t.Fatal("composition bypassed", err)
	}
	if !slices.Equal(m.Items(), before) {
		t.Fatal("refusal changed model")
	}
}
func TestFindSharedSettlementOrderAndDuplicateRefusal(t *testing.T) {
	m := New()
	must(t, m.openFind(true))
	var steps []string
	var callbacks []func(error)
	finished := 0
	field := func(id string, done func(error)) error {
		steps = append(steps, id)
		callbacks = append(callbacks, done)
		return nil
	}
	body := func(model *editor.Model, done func(error)) error {
		steps = append(steps, "body")
		callbacks = append(callbacks, done)
		return nil
	}
	must(t, m.SettleInputs(field, body, func(err error) {
		if err != nil {
			t.Fatal(err)
		}
		finished++
	}))
	for i := 0; i < 5; i++ {
		if len(callbacks) != i+1 {
			t.Fatal("not serialized", steps)
		}
		callbacks[i](nil)
		callbacks[i](nil)
	}
	expected := []string{"notes-search", ui.FindBarInputID(findID), findID + "/find/replacement", "note-title-n1", "body"}
	if !slices.Equal(steps, expected) || finished != 1 {
		t.Fatal(steps, finished)
	}
	steps = nil
	callbacks = nil
	failure := errors.New("settlement refused")
	must(t, m.SettleInputs(field, body, func(err error) {
		if err != failure {
			t.Fatal(err)
		}
		finished++
	}))
	callbacks[0](failure)
	callbacks[0](nil)
	if len(steps) != 1 || finished != 2 {
		t.Fatal("failure advanced sequence")
	}
}

func TestFindNextSettlesInitiallyUnpublishedQuery(t *testing.T) {
	m := findFixture(t)
	must(t, m.openFind(false))
	m.find.Query = ""
	complete, _ := deferSettlement(t, m)
	must(t, m.Commands().Invoke(context.Background(), "notes.find-next"))
	node, _ := find(m.View(), ui.FindBarInputID(findID))
	node.Change(ui.ChangeEvent{Text: "plan"})
	(*complete)(nil)
	if got := m.lookup("n1").editor.Snapshot().Selection; got != (editor.Selection{Anchor: 3, Focus: 7}) {
		t.Fatal("unpublished query bypassed", got)
	}
	if err := ui.Validate(m.View()); err != nil {
		t.Fatal(err)
	}
}
