package notes

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func find(n ui.Node, id string) (ui.Node, bool) {
	if n.ID == id {
		return n, true
	}
	for _, c := range n.Children {
		if got, ok := find(c, id); ok {
			return got, true
		}
	}
	return ui.Node{}, false
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateEditDeleteUndoRedoRestoresIdentityAndEditor(t *testing.T) {
	m := New()
	initial := m.Items()
	id, err := m.Add()
	must(t, err)
	created := m.lookup(id)
	must(t, m.Update(id, "Design ideas", "A careful plan.\nA second line."))
	must(t, m.Delete(id))
	if m.lookup(id) != nil || m.History().UndoCount != 3 {
		t.Fatal("delete/history", m.Items(), m.History())
	}
	must(t, m.Undo())
	if m.lookup(id) != created || m.lookup(id).value.Title != "Design ideas" {
		t.Fatal("undo delete changed object")
	}
	must(t, m.Undo())
	if m.lookup(id).value.Title != "" || m.lookup(id).value.Body != "" {
		t.Fatal("undo edit")
	}
	must(t, m.Undo())
	if !slices.Equal(m.Items(), initial) {
		t.Fatal("undo create")
	}
	must(t, m.Redo())
	if m.lookup(id) != created {
		t.Fatal("redo create replaced retained editor")
	}
	must(t, m.Redo())
	if m.lookup(id).value.Body != "A careful plan.\nA second line." {
		t.Fatal("redo edit")
	}
	must(t, m.Redo())
	if m.lookup(id) != nil {
		t.Fatal("redo delete")
	}
	other, err := m.Add()
	must(t, err)
	if other == id || m.History().RedoCount != 0 {
		t.Fatal("identity reuse or stale redo branch")
	}
}

func TestNativeBodyCallbackIsRecordedExactlyOnce(t *testing.T) {
	m := New()
	n := m.lookup("n1")
	original := n.value.Body
	must(t, n.editor.Replace("A😀e\u0301\r\nאבג", editor.Selection{Anchor: 2, Focus: 2}))
	view := m.View()
	node, ok := find(view, "note-editor-n1")
	if !ok || node.Change == nil {
		t.Fatal("native editor callback missing")
	}
	node.Change(ui.ChangeEvent{Text: "deliberately ignored event copy"})
	node.Change(ui.ChangeEvent{})
	if m.History().UndoCount != 1 || n.value.Body != "A😀e\u0301\r\nאבג" {
		t.Fatal("duplicate or wrong admitted native change", m.History(), n.value.Body)
	}
	must(t, m.Undo())
	if n.value.Body != original || n.editor.Snapshot().Text != original {
		t.Fatal("undo did not replay editor and app text together")
	}
	must(t, m.Redo())
	if n.value.Body != n.editor.Snapshot().Text || n.value.Body != "A😀e\u0301\r\nאבג" {
		t.Fatal("redo native edit")
	}
}

func TestTitleOnlyChangesPreserveBodyCaretAndNewTitleStartsEmpty(t *testing.T) {
	m := New()
	n := m.lookup("n1")
	selection := editor.Selection{Anchor: 3, Focus: 9}
	must(t, n.editor.Select(selection))
	before := n.editor.Snapshot()
	must(t, m.Update("n1", "A different title", n.value.Body))
	if n.editor.Snapshot() != before {
		t.Fatal("title edit moved body selection or created text history")
	}
	must(t, m.Undo())
	if n.editor.Snapshot() != before {
		t.Fatal("title undo moved body selection")
	}
	id, err := m.Add()
	must(t, err)
	title, ok := find(m.View(), "note-title-"+id)
	if !ok || title.Text != "" || displayTitle(title.Text) != "New Note" {
		t.Fatal("new note inserts a placeholder into editable user data")
	}
}

func TestToolbarFlushesEarlierCommittedTextBeforeDelete(t *testing.T) {
	m := New()
	n := m.lookup("n1")
	before := n.value.Body
	must(t, n.editor.Replace("Committed before delete", editor.Selection{})) // No Change callback delivered yet.
	must(t, m.Delete("n1"))
	if m.History().UndoCount != 2 {
		t.Fatal("earlier native edit was omitted")
	}
	must(t, m.Undo())
	if m.lookup("n1") != n || n.value.Body != "Committed before delete" {
		t.Fatal("delete undo lost preceding text")
	}
	must(t, m.Undo())
	if n.value.Body != before || n.editor.Snapshot().Text != before {
		t.Fatal("native text undo lost ordering")
	}
}

func TestCompositionIsPreservedByRefusedDestructiveActions(t *testing.T) {
	m := New()
	n := m.lookup("n1")
	s := n.editor.Snapshot()
	must(t, n.editor.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: editor.Range{}, Insert: "あ", Selection: editor.Selection{Anchor: 1, Focus: 1}, Marked: editor.Range{Start: 0, End: 1}}))
	before, items := n.editor.Snapshot(), m.Items()
	journal := m.History()
	for _, action := range []func() error{func() error { return m.Select("n2") }, func() error { return m.SelectFolder("projects") }, func() error { return m.Delete("n1") }, func() error { return m.Update("n1", "Changed", "Changed") }, m.Undo, m.Redo, func() error { _, err := m.Add(); return err }} {
		if !errors.Is(action(), editor.ErrComposition) {
			t.Fatal("action did not preserve live composition")
		}
		if !slices.Equal(items, m.Items()) || m.History() != journal || n.editor.Snapshot() != before || m.Selected() != "n1" {
			t.Fatal("refusal mutated state")
		}
	}
	must(t, n.editor.CancelComposition())
	must(t, m.Select("n2"))
}

func TestFolderSearchPinAndSelectionUseStableIdentities(t *testing.T) {
	m := New()
	must(t, m.SelectFolder("all"))
	if len(m.Visible()) != 6 {
		t.Fatal("all notes")
	}
	must(t, m.SetSearch("  RECYCLED  "))
	v := m.Visible()
	if len(v) != 1 || v[0].ID != "n4" {
		t.Fatal("body search", v)
	}
	open, _ := find(m.View(), "note-select-n4")
	open.Activate()
	if m.Selected() != "n4" {
		t.Fatal("public selection callback")
	}
	must(t, m.SetSearch("no matching note"))
	if len(m.Visible()) != 0 || m.Selected() != "n4" {
		t.Fatal("empty filter changed retained editor")
	}
	if _, ok := find(m.View(), "notes-empty-search"); !ok {
		t.Fatal("empty state missing")
	}
	must(t, m.SelectFolder("projects"))
	if m.SelectedFolder() != "projects" || m.query != "" {
		t.Fatal("folder transition")
	}
	must(t, m.SetPinned("n5", true))
	if m.Visible()[0].ID != "n5" {
		t.Fatal("pin order")
	}
	must(t, m.Undo())
	if m.lookup("n5").value.Pinned {
		t.Fatal("pin undo")
	}
	if m.Select("nonesuch") == nil || m.SelectFolder("nonesuch") == nil {
		t.Fatal("invalid identity admitted")
	}
	copy := m.Items()
	copy[0].Body = "mutated detached data"
	if m.Items()[0].Body == copy[0].Body {
		t.Fatal("snapshot aliases mutable state")
	}
}

func TestBoundsRejectWithoutMutationAndHistoryIsBounded(t *testing.T) {
	m := New()
	before := m.Items()
	state := m.History()
	for _, title := range []string{strings.Repeat("x", MaxTitleBytes+1), "bad\x00title", "bad\xfftitle"} {
		if m.Update("n1", title, "body") == nil {
			t.Fatal("invalid title admitted")
		}
	}
	for _, body := range []string{strings.Repeat("x", MaxBodyBytes+1), "bad\x00body", "bad\xffbody"} {
		if m.Update("n1", "title", body) == nil {
			t.Fatal("invalid body admitted")
		}
	}
	if m.SetSearch(strings.Repeat("界", MaxTitleBytes)) == nil || m.SetSearch("\xff") == nil {
		t.Fatal("invalid/oversized search retained")
	}
	if !slices.Equal(before, m.Items()) || m.History() != state {
		t.Fatal("rejected input mutated model or history")
	}
	for len(m.notes) < MaxNotes {
		_, err := m.Add()
		must(t, err)
	}
	if _, err := m.Add(); !errors.Is(err, ErrLimit) {
		t.Fatal("capacity not enforced")
	}
	if m.History().UndoCount != HistoryLimit {
		t.Fatal("history not bounded", m.History())
	}
	for range HistoryLimit {
		must(t, m.Undo())
	}
	if !errors.Is(m.Undo(), history.ErrEmpty) {
		t.Fatal("evicted history replayed")
	}
}

func TestGlobalCommandsUndoAcrossSelectionsAndRefuseCanceledContext(t *testing.T) {
	m := New()
	must(t, m.Update("n1", "Edited first note", m.lookup("n1").value.Body))
	must(t, m.Select("n2"))
	matched, err := m.Commands().ShortcutInScope(context.Background(), "Meta+Z", command.WindowShortcuts)
	must(t, err)
	if !matched || m.Selected() != "n1" || m.lookup("n1").value.Title == "Edited first note" {
		t.Fatal("global history did not reveal affected note")
	}
	matched, err = m.Commands().ShortcutInScope(context.Background(), "Meta+Z", command.AfterFocusShortcuts)
	must(t, err)
	if matched {
		t.Fatal("duplicate native fallback history route")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	count := len(m.notes)
	if !errors.Is(m.Commands().Invoke(ctx, "notes.new"), context.Canceled) || len(m.notes) != count {
		t.Fatal("canceled command mutated notes")
	}
}

func TestEveryFolderAndEmptyStateProduceValidAccessibleTrees(t *testing.T) {
	m := New()
	must(t, Theme().Validate())
	check := func() {
		t.Helper()
		root := m.View()
		must(t, ui.Validate(root))
		for _, id := range ui.FocusOrder(root) {
			n, _ := find(root, id)
			if n.Accessibility.Name == "" {
				t.Fatal("unnamed control", id)
			}
		}
	}
	for _, id := range []string{"all", "notes", "projects", "personal"} {
		must(t, m.SelectFolder(id))
		check()
		must(t, m.SetSearch("no matches"))
		check()
	}
	id, err := m.Add()
	must(t, err)
	check()
	must(t, m.Select("n6"))
	check() // Focus request for another note must disappear.
	must(t, m.Delete(id))
	check()
	for len(m.notes) > 0 {
		must(t, m.Delete(m.notes[0].value.ID))
	}
	check()
	if _, ok := find(m.View(), "notes-empty-document"); !ok {
		t.Fatal("empty document state omitted")
	}
}

func TestNativeShellKeepsInputAndUnifiedCommands(t *testing.T) {
	m := New()
	before := m.Items()
	m.SetNativeToolbar(true)
	view := m.View()
	if err := ui.Validate(view); err != nil {
		t.Fatal(err)
	}
	if _, ok := find(view, "notes-toolbar"); ok {
		t.Fatal("native shell duplicated client toolbar")
	}
	if _, ok := find(view, "note-editor-n1"); !ok {
		t.Fatal("native shell removed editor")
	}
	if err := m.Commands().Invoke(context.Background(), "notes.delete"); err != nil {
		t.Fatal(err)
	}
	if len(m.Items()) != len(before)-1 {
		t.Fatal("native delete bypassed model")
	}
	must(t, m.Undo())
	if !slices.Equal(before, m.Items()) {
		t.Fatal("native command history diverged")
	}
}
