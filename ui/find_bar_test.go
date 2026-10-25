package ui

import (
	"github.com/dshills/vellum/editor"
	"strings"
	"testing"
)

func TestFindUnicodeWordAndBoundedMatches(t *testing.T) {
	m, e := FindMatches("Σ σ ς plan planner plan_2", "σ", FindOptions{})
	if e != nil || len(m) != 3 || m[2].Start != 4 {
		t.Fatal(m, e)
	}
	m, e = FindMatches("plan planner plan_2 plan", "plan", FindOptions{WholeWord: true})
	if e != nil || len(m) != 2 {
		t.Fatal(m, e)
	}
	if _, e = FindMatches(strings.Repeat("a", MaxFindMatches+1), "a", FindOptions{}); e != ErrFindLimit {
		t.Fatal(e)
	}
}
func TestFindReplaceAllOneUndoAndCompositionRefusal(t *testing.T) {
	m, _ := editor.New("α plan PLAN", editor.Config{})
	r := FindRequest{Action: FindReplaceAll, Query: "plan", Replacement: "drawing"}
	if e := r.Apply(m); e != nil {
		t.Fatal(e)
	}
	if m.Snapshot().Text != "α drawing drawing" {
		t.Fatal(m.Snapshot())
	}
	if e := m.Journal().Undo(); e != nil || m.Snapshot().Text != "α plan PLAN" {
		t.Fatal(m.Snapshot(), e)
	}
	if e := m.Journal().Redo(); e != nil || m.Snapshot().Text != "α drawing drawing" {
		t.Fatal(e)
	}
	s := m.Snapshot()
	if e := m.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: editor.Range{}, Insert: "x", Selection: editor.Selection{Anchor: 1, Focus: 1}, Marked: editor.Range{Start: 0, End: 1}}); e != nil {
		t.Fatal(e)
	}
	if e := r.Apply(m); e != editor.ErrComposition {
		t.Fatal(e)
	}
}
func TestFindPreviousWrapReplaceAndClosedCallbacks(t *testing.T) {
	m, _ := editor.New("one one", editor.Config{})
	q := FindRequest{Query: "one"}
	if q.Apply(m) != nil || m.Snapshot().Selection.Focus != 3 {
		t.Fatal(m.Snapshot())
	}
	q.Action = FindPrevious
	if q.Apply(m) != nil || m.Snapshot().Selection.Focus != 7 {
		t.Fatal("wrap")
	}
	q.Action = FindReplace
	q.Replacement = "two"
	if q.Apply(m) != nil || m.Snapshot().Text != "one two" {
		t.Fatal(m.Snapshot())
	}
	calls := 0
	f := &FindBar{ID: "find", Open: true, Query: "one", Model: m, Request: func(FindRequest) { calls++ }}
	n, e := f.View()
	if e != nil {
		t.Fatal(e)
	}
	old := pickerNode(n, "find/find/next")
	f.Open = false
	old.Activate()
	if calls != 0 {
		t.Fatal("closed bar action")
	}
}

func TestFindNonoverlapCacheAndReplaceAllSelection(t *testing.T) {
	matches, e := FindMatches("aaaaa", "aa", FindOptions{})
	if e != nil || len(matches) != 2 || matches[1].Start != 2 {
		t.Fatal(matches, e)
	}
	m, _ := editor.New("one one", editor.Config{})
	_ = m.Select(editor.Selection{Anchor: 4, Focus: 7})
	f := &FindBar{ID: "cache", Model: m, Open: true, Query: "one"}
	if _, e = f.View(); e != nil {
		t.Fatal(e)
	}
	first := &f.cacheMatches[0]
	_ = m.Select(editor.Selection{})
	if _, e = f.View(); e != nil || first != &f.cacheMatches[0] {
		t.Fatal("selection invalidated content cache")
	}
	_ = m.Select(editor.Selection{Anchor: 4, Focus: 7})
	if e := (FindRequest{Action: FindReplaceAll, Query: "one", Replacement: "two words"}).Apply(m); e != nil {
		t.Fatal(e)
	}
	if m.Snapshot().Selection != (editor.Selection{Anchor: 10, Focus: 19}) {
		t.Fatal(m.Snapshot().Selection)
	}
	f.Open = false
	_, _ = f.View()
	if f.cacheModel != nil || f.cacheMatches != nil {
		t.Fatal("closed bar retained cache")
	}
}

func TestFindNoMatchDoesNotCommit(t *testing.T) {
	m, _ := editor.New("text", editor.Config{})
	before := m.Snapshot()
	if e := (FindRequest{Action: FindReplaceAll, Query: "missing", Replacement: "x"}).Apply(m); e != ErrFindNoMatch {
		t.Fatal(e)
	}
	if m.Snapshot() != before {
		t.Fatal("no match changed model")
	}
}
