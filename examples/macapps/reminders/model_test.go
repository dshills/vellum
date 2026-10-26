package reminders

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func find(t *testing.T, n ui.Node, id string) ui.Node {
	t.Helper()
	var visit func(ui.Node) (ui.Node, bool)
	visit = func(n ui.Node) (ui.Node, bool) {
		if n.ID == id {
			return n, true
		}
		for _, c := range n.Children {
			if result, ok := visit(c); ok {
				return result, true
			}
		}
		return ui.Node{}, false
	}
	result, ok := visit(n)
	if !ok {
		t.Fatalf("missing node %q", id)
	}
	return result
}
func item(t *testing.T, m *Model, id string) Reminder {
	t.Helper()
	for _, r := range m.Items() {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("missing item %q", id)
	return Reminder{}
}
func checkView(t *testing.T, m *Model) {
	t.Helper()
	if err := ui.Validate(m.View()); err != nil {
		t.Fatal(err)
	}
	if err := Theme().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCompleteEditHistoryRestoresIdentity(t *testing.T) {
	m := New()
	before := m.Items()
	id, err := m.Add("  Buy coffee  ")
	if err != nil {
		t.Fatal(err)
	}
	created := item(t, m, id)
	if created.Title != "Buy coffee" || created.ListID != "groceries" {
		t.Fatalf("created=%+v", created)
	}
	if err = m.SetCompleted(id, true); err != nil {
		t.Fatal(err)
	}
	if err = m.Update(id, "Coffee beans", "Decaf", "2026-10-20", "high", true); err != nil {
		t.Fatal(err)
	}
	edited := item(t, m, id)
	for i := 0; i < 3; i++ {
		if err = m.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(before, m.Items()) {
		t.Fatal("undo did not restore original collection")
	}
	for i := 0; i < 3; i++ {
		if err = m.Redo(); err != nil {
			t.Fatal(err)
		}
	}
	if got := item(t, m, id); got != edited {
		t.Fatalf("redo restored different object: %+v", got)
	}
	if err = m.SetCompleted(id, false); err != nil {
		t.Fatal(err)
	}
	if item(t, m, id).Completed {
		t.Fatal("cannot reopen reminder")
	}
	checkView(t, m)
}

func TestValidationIsAtomicAndHistoryIsBounded(t *testing.T) {
	m := New()
	before := m.Items()
	h := m.History()
	for _, title := range []string{" ", strings.Repeat("x", MaxTitleBytes+1), "\xff", "bad\x00title"} {
		if _, err := m.Add(title); err == nil {
			t.Fatalf("accepted invalid title %q", title)
		}
	}
	for _, tc := range []struct{ due, priority string }{{"2026-02-30", ""}, {"", "urgent"}} {
		if err := m.Update("r1", "New", "", tc.due, tc.priority, false); err == nil {
			t.Fatal("invalid details accepted")
		}
	}
	if err := m.Update("missing", "New", "", "", "", false); err == nil {
		t.Fatal("accepted missing identity")
	}
	if !slices.Equal(before, m.Items()) || m.History().UndoCount != h.UndoCount {
		t.Fatal("failed mutation changed state/history")
	}
	for i := 0; i < 150; i++ {
		if err := m.SetCompleted("r1", i%2 == 0); err != nil {
			t.Fatal(err)
		}
	}
	if m.History().UndoCount != 64 {
		t.Fatalf("history grew beyond bound: %+v", m.History())
	}
	for len(m.items) < MaxReminders {
		if _, err := m.Add(fmt.Sprintf("Item %d", len(m.items))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Add("Overflow"); !errors.Is(err, ErrLimit) {
		t.Fatalf("capacity error=%v", err)
	}
	checkView(t, m)
}

func TestBranchingHistoryNeverReusesIDs(t *testing.T) {
	m := New()
	a, _ := m.Add("First")
	if err := m.Undo(); err != nil {
		t.Fatal(err)
	}
	b, _ := m.Add("Second")
	if a == b {
		t.Fatal("new branch reused stable ID")
	}
	if m.History().RedoCount != 0 {
		t.Fatal("new edit did not discard redo branch")
	}
	if err := m.SetCompleted(b, false); err != nil {
		t.Fatal(err)
	}
	if m.History().UndoCount != 1 {
		t.Fatal("no-op completion added history entry")
	}
	items := m.Items()
	items[0].Title = "external mutation"
	if m.Items()[0].Title == items[0].Title {
		t.Fatal("Items leaked internal slice")
	}
}

func TestFiltersSearchAndSmartListCreation(t *testing.T) {
	m := New()
	for _, tc := range []struct {
		view string
		want int
	}{{"groceries", 6}, {"all", 11}, {"today", 3}, {"scheduled", 6}, {"flagged", 2}, {"completed", 1}, {"assigned", 2}, {"work", 2}} {
		if err := m.Select(tc.view); err != nil {
			t.Fatal(err)
		}
		if got := len(m.Visible()); got != tc.want {
			t.Errorf("%s=%d want %d", tc.view, got, tc.want)
		}
		checkView(t, m)
	}
	if err := m.SetSearch("DRAWINGS"); err != nil {
		t.Fatal(err)
	}
	if got := m.Visible(); len(got) != 1 || got[0].ID != "r9" {
		t.Fatalf("search=%v", got)
	}
	if err := m.SetSearch(strings.Repeat("x", MaxTitleBytes+1)); err == nil || m.Search() != "DRAWINGS" {
		t.Fatal("oversized search changed controlled text")
	}
	if err := m.Select("unknown"); err == nil || m.Selected() != "work" {
		t.Fatal("invalid selection changed list")
	}
	_ = m.SetSearch("")
	_ = m.Select("today")
	id, err := m.Add("Due today")
	if err != nil {
		t.Fatal(err)
	}
	if r := item(t, m, id); r.Due != Today || r.ListID != "work" {
		t.Fatalf("smart-list creation=%+v", r)
	}
	_ = m.Select("flagged")
	id, err = m.Add("Flag me")
	if err != nil || !item(t, m, id).Flagged {
		t.Fatal("flagged creation did not seed flag")
	}
}

func TestAccessibleCallbacksAndModalCancelSave(t *testing.T) {
	m := New()
	checkView(t, m)
	complete := find(t, m.View(), ui.ChecklistCheckID(checklistID("r1"), "r1"))
	if complete.Kind != ui.ToggleButton || complete.Checked {
		t.Fatal("completion action lacks accessible state")
	}
	complete.Change(ui.ChangeEvent{Checked: true})
	if !item(t, m, "r1").Completed || m.History().UndoCount != 1 {
		t.Fatal("completion action missed history")
	}
	find(t, m.View(), "reminders-show-completed").Activate()
	if !find(t, m.View(), ui.ChecklistCheckID(checklistID("r1"), "r1")).Checked {
		t.Fatal("completed action lacks state")
	}
	find(t, m.View(), "details-r1").Activate()
	checkView(t, m)
	if got := ui.FocusOrder(m.View()); slices.Contains(got, "reminders-new-title") || !slices.Contains(got, "details-title") {
		t.Fatalf("modal focus escaped: %v", got)
	}
	find(t, m.View(), "details-title").Change(ui.ChangeEvent{Text: "Discard me"})
	find(t, m.View(), "details-cancel").Activate()
	if item(t, m, "r1").Title != "Croissants" || m.History().UndoCount != 1 {
		t.Fatal("cancel committed draft")
	}
	find(t, m.View(), "details-r1").Activate()
	find(t, m.View(), "details-title").Change(ui.ChangeEvent{Text: "Almond croissants"})
	find(t, m.View(), "details-date").Change(ui.ChangeEvent{Text: "2026-02-30"})
	find(t, m.View(), "details-save").Activate()
	if m.editing == "" || m.History().UndoCount != 1 {
		t.Fatal("invalid date committed/closed draft")
	}
	find(t, m.View(), "details-date").Change(ui.ChangeEvent{Text: Today})
	find(t, m.View(), "details-save").Activate()
	if item(t, m, "r1").Title != "Almond croissants" || m.History().UndoCount != 2 || m.editing != "" {
		t.Fatal("save did not commit atomic edit")
	}
	find(t, m.View(), "reminders-undo").Activate()
	if item(t, m, "r1").Title != "Croissants" {
		t.Fatal("undo failed from view")
	}
	find(t, m.View(), "reminders-new").Activate()
	if m.View().Focus.ID != "reminders-new-title" || m.History().UndoCount != 1 {
		t.Fatal("New Reminder created object before input")
	}
	find(t, m.View(), "reminders-new-title").Change(ui.ChangeEvent{Text: "Berries"})
	find(t, m.View(), "reminders-add").Activate()
	if len(m.Items()) != 13 || m.composer != "" {
		t.Fatal("composer did not create exactly once and clear")
	}
	checkView(t, m)
}

func TestCommandHistoryUsesAfterFocusPhase(t *testing.T) {
	m := New()
	if _, err := m.Add("Command test"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if handled, err := m.Commands().ShortcutInScope(ctx, "Meta+Z", command.WindowShortcuts); handled || err != nil {
		t.Fatalf("command stole focused editor shortcut: %v %v", handled, err)
	}
	if handled, err := m.Commands().ShortcutInScope(ctx, "Meta+Z", command.AfterFocusShortcuts); !handled || err != nil {
		t.Fatalf("undo: %v %v", handled, err)
	}
	if len(m.Items()) != 12 {
		t.Fatal("command undo did not remove new object")
	}
	if handled, err := m.Commands().ShortcutInScope(ctx, "Meta+Shift+Z", command.AfterFocusShortcuts); !handled || err != nil {
		t.Fatalf("redo: %v %v", handled, err)
	}
	if len(m.Items()) != 13 {
		t.Fatal("command redo did not restore object")
	}
}
