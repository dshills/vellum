package ui

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDropZoneAtomicSharedChooserValidationAndFeedback(t *testing.T) {
	calls := 0
	var got FileDropZoneEvent
	z := &FileDropZone{ID: "files", Extensions: []string{"txt"}, MaxFiles: 2, Change: func(e FileDropZoneEvent) error { calls++; got = e; return nil }, Choose: func() {}}
	path := filepath.Join(t.TempDir(), "A.TXT")
	bad := filepath.Join(t.TempDir(), "B.png")
	n, e := z.View()
	if e != nil {
		t.Fatal(e)
	}
	if e = z.ProposeFiles([]string{path, bad}); e == nil || calls != 0 || z.Feedback() == "" {
		t.Fatal("partial admission", e, calls)
	}
	paths := []string{path}
	if e = z.ProposeFiles(paths); e != nil {
		t.Fatal(e)
	}
	got.Paths[0] = "private mutation"
	if paths[0] != path {
		t.Fatal("shared payload")
	}
	z.Change = func(e FileDropZoneEvent) error { calls++; return errors.New(strings.Repeat("x", 5000)) }
	if !n.FileDrop(FileDropEvent{Paths: []string{path}, WindowID: "one"}) || len(z.Feedback()) > 512 {
		t.Fatal("rejection was not consumed/bounded")
	}
	z.Disabled = true
	if n.FileDrop(FileDropEvent{Paths: []string{path}}) {
		t.Fatal("disabled target handled")
	}
	z.Disabled = false
	z.ID = "reused"
	if n.FileDrop(FileDropEvent{Paths: []string{path}}) {
		t.Fatal("retired ID handled")
	}
	z.ID = "files"
	z.Extensions = []string{".txt"}
	if _, e = z.View(); e == nil {
		t.Fatal("bad extension")
	}
}

func TestDropZoneFeedbackByteBoundWithInvalidAndSplitUTF8(t *testing.T) {
	for _, s := range []string{strings.Repeat("€", 300), strings.Repeat(string([]byte{255}), 600), strings.Repeat("x", 511) + "€", strings.Repeat("a\x00", 400)} {
		got := boundedWidgetMessage(s)
		if len(got) > 512 || !utf8.ValidString(got) || strings.ContainsRune(got, 0) {
			t.Fatalf("invalid bounded feedback: %d bytes", len(got))
		}
	}
}
