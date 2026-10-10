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
	for _, text := range []string{"a\x00b", "\xff", strings.Repeat("a", (1<<20)+1)} {
		if nativePublicationSupported(text) {
			t.Fatal("unsupported document accepted")
		}
	}
	if !nativePublicationSupported("a\nאבג👩") {
		t.Fatal("multiline UTF-8 refused")
	}
}
