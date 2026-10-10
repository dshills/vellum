package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/text/lines"
	"cogentcore.org/core/text/textcore"
	"example.com/vellum/foundationintegration/inputbridge"
	"strings"
	"testing"
)

func TestNativeQueuedClientRoutingAndStaleShadow(t *testing.T) {
	ed := textcore.NewEditor()
	old := ed.Lines
	ls := lines.NewLinesExact().SetString("editor\n")
	ls.Autosave = false
	ed.SetLines(ls)
	old.Close()
	defer ls.Close()
	f := fixture{field: core.NewTextField().SetText("field"), editor: ed, nativeClientID: 40}
	apply := func(u inputbridge.Update) bool {
		r, err := inputbridge.Decode(u)
		return err == nil && f.applyNativeSnapshot(u, r)
	}
	if !apply(inputbridge.Update{ClientID: 8, BeforeText: "field", Text: "fieldに", Kind: "mark", Anchor: 6, Focus: 6, MarkedStart: 5, MarkedEnd: 6}) {
		t.Fatal("queued old field update lost")
	}
	if ls.String() != "editor\n" {
		t.Fatal("old field update routed to editor")
	}
	f.nativeClientID = 8
	if !apply(inputbridge.Update{ClientID: 40, BeforeText: "editor\n", Text: "editorに\n", Kind: "mark", Anchor: 7, Focus: 7, MarkedStart: 6, MarkedEnd: 7}) {
		t.Fatal("queued old editor update lost")
	}
	if f.field.EditingState().Text != "fieldに" {
		t.Fatal("editor update routed to field")
	}
	ls.InsertText(ls.EndPos(), []rune("new"))
	before := ls.String()
	if apply(inputbridge.Update{ClientID: 40, BeforeText: "editorに\n", Text: "editor\n", Kind: "cancel", Anchor: 0, Focus: 0, MarkedStart: -1, MarkedEnd: -1}) || ls.String() != before {
		t.Fatal("stale queued cancellation erased shared edit")
	}
	if apply(inputbridge.Update{ClientID: 99, Text: "bad", Kind: "insert", MarkedStart: -1, MarkedEnd: -1}) {
		t.Fatal("unknown client accepted")
	}
}

func TestNativePublicationUnsupportedDocument(t *testing.T) {
	for _, text := range []string{"a\x00b", "\xff", strings.Repeat("a", inputbridge.MaxDocumentBytes+1)} {
		if nativePublicationSupported(text) {
			t.Fatal("unsupported document accepted")
		}
	}
	if !nativePublicationSupported("a\nאבג👩") {
		t.Fatal("multiline UTF-8 refused")
	}
}

func TestNativeLargeDocumentQueuedDeltasAndSharedEdit(t *testing.T) {
	text := strings.Repeat("abcdefghij\n", 190000)
	ed := textcore.NewEditor()
	old := ed.Lines
	ls := lines.NewLinesExact().SetString(text)
	ls.Autosave = false
	ed.SetLines(ls)
	old.Close()
	defer ls.Close()
	f := fixture{editor: ed, nativeShadows: map[uint64]string{40: text}}
	apply := func(u inputbridge.Update) bool {
		expanded, err := f.expandNativeUpdate(u)
		if err != nil {
			return false
		}
		decoded, err := inputbridge.Decode(expanded)
		if err != nil || !f.applyNativeSnapshot(expanded, decoded) {
			return false
		}
		f.nativeShadows[40] = expanded.Text
		return true
	}
	start := len(text) - 2
	u := inputbridge.Update{ClientID: 40, Delta: true, ReplaceStart: start, ReplaceEnd: start, Text: "に", Kind: "mark", Anchor: start + 1, Focus: start + 1, MarkedStart: start, MarkedEnd: start + 1}
	if !apply(u) {
		t.Fatal("large-document preedit refused")
	}
	u.ReplaceEnd = start + 1
	u.BeforeText = "に"
	u.Text = "日本"
	u.Anchor = start + 2
	u.Focus = start + 2
	u.MarkedEnd = start + 2
	if !apply(u) {
		t.Fatal("queued large-document replacement refused")
	}
	u.ReplaceEnd = start + 2
	u.BeforeText = "日本"
	u.Text = ""
	u.Kind = "cancel"
	u.Anchor = start
	u.Focus = start
	u.MarkedStart = -1
	u.MarkedEnd = -1
	if !apply(u) || ls.String() != text {
		t.Fatal("large-document cancellation did not restore exact bytes")
	}
	u.ReplaceEnd = start
	u.BeforeText = ""
	u.Text = "字"
	u.Kind = "insert"
	u.Anchor = start + 1
	u.Focus = start + 1
	ls.InsertText(ls.EndPos(), []rune("external"))
	before := ls.String()
	if apply(u) || ls.String() != before {
		t.Fatal("outside-range shared edit was overwritten")
	}
}
