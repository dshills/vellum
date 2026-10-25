package sdkcontract_test

import (
	"errors"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"testing"
)

// This external-package consumer verifies public editor/find contracts together;
// it does not simulate a human IME or VoiceOver session.
func TestPersonalMacUnicodeFindCompositionAndHistory(t *testing.T) {
	original := "😀 cafe\u0301 PLAN\r\nplan\nאבג\n"
	m, e := editor.New(original, editor.Config{})
	if e != nil {
		t.Fatal(e)
	}
	snap := m.Snapshot()
	e = m.Preview(editor.Edit{BaseRevision: snap.Revision, BaseVisibleEpoch: snap.VisibleEpoch, Range: editor.Range{}, Insert: "x", Selection: editor.Selection{Anchor: 1, Focus: 1}, Marked: editor.Range{Start: 0, End: 1}})
	if e != nil {
		t.Fatal(e)
	}
	request := ui.FindRequest{Action: ui.FindReplaceAll, Query: "plan", Replacement: "drawing"}
	if e = request.Apply(m); !errors.Is(e, editor.ErrComposition) {
		t.Fatal("replace modified active composition", e)
	}
	if m.Snapshot().Text != original || m.Snapshot().History.UndoCount != 0 {
		t.Fatal("composition leaked into committed text/history")
	}
	if e = m.CancelComposition(); e != nil {
		t.Fatal(e)
	}
	if e = request.Apply(m); e != nil {
		t.Fatal(e)
	}
	replaced := "😀 cafe\u0301 drawing\r\ndrawing\nאבג\n"
	if m.Snapshot().Text != replaced || m.Snapshot().History.UndoCount != 1 {
		t.Fatal("Unicode replace was not one exact-byte operation", m.Snapshot())
	}
	if e = m.Journal().Undo(); e != nil || m.Snapshot().Text != original {
		t.Fatal("Unicode/CRLF undo", e)
	}
	if e = m.Journal().Redo(); e != nil || m.Snapshot().Text != replaced {
		t.Fatal("Unicode/CRLF redo", e)
	}
}
