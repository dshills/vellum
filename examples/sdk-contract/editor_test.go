package sdkcontract_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

func TestPublicEditorExactBytesHistoryAndVisibleEpoch(t *testing.T) {
	original := "A\r\n🙂\r\n"
	m, err := editor.New(original, editor.Config{HistoryLimit: 4})
	if err != nil {
		t.Fatal(err)
	}
	first := m.Snapshot()
	change := editor.Edit{BaseRevision: first.Revision, BaseVisibleEpoch: first.VisibleEpoch, Range: editor.Range{Start: 3, End: 4}, Before: "🙂", Insert: "ש", Selection: editor.Selection{Anchor: 4, Focus: 3}}
	if err := m.Apply(change); err != nil {
		t.Fatal(err)
	}
	changed := m.Snapshot()
	if changed.Text != "A\r\nש\r\n" || changed.Selection != change.Selection || !changed.History.Dirty {
		t.Fatalf("accepted edit: %+v", changed)
	}
	if err := m.Journal().Undo(); err != nil {
		t.Fatal(err)
	}
	undone := m.Snapshot()
	if undone.Text != original || undone.History.Dirty || undone.Revision <= changed.Revision || undone.VisibleEpoch <= changed.VisibleEpoch {
		t.Fatalf("undo reused content identity: %+v", undone)
	}
	if err := m.Apply(change); err == nil || m.Snapshot() != undone {
		t.Fatal("ABA edit was accepted or mutated document")
	}
	preview := editor.Edit{BaseRevision: undone.Revision, BaseVisibleEpoch: undone.VisibleEpoch, Range: editor.Range{Start: 3, End: 4}, Before: "🙂", Insert: "に", Selection: editor.Selection{Anchor: 4, Focus: 4}, Marked: editor.Range{Start: 3, End: 4}}
	if err := m.Preview(preview); err != nil {
		t.Fatal(err)
	}
	pending := m.Snapshot()
	if pending.Text != original || pending.VisibleText != "A\r\nに\r\n" || !pending.Composing || pending.History.UndoCount != 0 {
		t.Fatalf("composition entered committed history: %+v", pending)
	}
	// Finishing native marked text may have no additional visible text delta.
	if err := m.CommitComposition(); err != nil {
		t.Fatal(err)
	}
	committed := m.Snapshot()
	if committed.Composing || committed.Text != pending.VisibleText || committed.History.UndoCount != 1 {
		t.Fatalf("composition did not commit once: %+v", committed)
	}
	if err := m.Journal().Undo(); err != nil || m.Snapshot().Text != original {
		t.Fatal("composition undo did not restore exact document", err)
	}
	if err := m.Journal().Redo(); err != nil || m.Snapshot().Text != pending.VisibleText {
		t.Fatal("composition redo failed", err)
	}
}

func TestPublicMultilineDescriptorAdmission(t *testing.T) {
	m, err := editor.New("literal <text>\n", editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	node := ui.Node{ID: "document", Kind: ui.MultilineText, Editor: m, Accessibility: ui.Accessibility{Name: "Document"}}
	if err := ui.Validate(node); err != nil {
		t.Fatal(err)
	}
	if !ui.IsFocusable(node.Kind) {
		t.Fatal("document cannot participate in managed focus")
	}
	copied := *m
	node.Editor = &copied
	if err := ui.Validate(node); err == nil {
		t.Fatal("copied model accepted by descriptor")
	}
	node.Editor = nil
	if err := ui.Validate(node); err == nil {
		t.Fatal("nil editor model accepted")
	}
	node.Editor = &editor.Model{}
	if err := ui.Validate(node); err == nil {
		t.Fatal("uninitialized editor model accepted")
	}
}

func TestPublicEditorSessionAtomicInstallAndSave(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.Open(app.Window{ID: "editor", Title: "SDK editor", Width: 200, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	m, err := editor.New("exact\r\n🙂\r\n", editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan document.Result, 1)
	session, err := document.New(document.Config[string]{Owner: a, Window: "editor", Journal: m.Journal(), Codec: document.Codec[string]{
		Capture: func() (string, error) {
			s := m.Snapshot()
			if s.Composing {
				return "", errors.New("unsettled composition")
			}
			return s.Text, nil
		},
		Encode: func(s string) ([]byte, error) { return []byte(s), nil }, Decode: func(b []byte) (string, error) { return string(b), nil },
		Install: func(s string) error { return m.Install(s, editor.Selection{}) },
	}, OnComplete: func(r document.Result) { results <- r }})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "SDK editor Ω.txt")
	var operationErr error
	if err := a.Do(func() { operationErr = session.Save(context.Background(), document.SaveOptions{Path: path}) }); err != nil {
		t.Fatal(err)
	}
	if operationErr != nil {
		t.Fatal(operationErr)
	}
	finish := func() {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if err := a.Drain(); err != nil {
				t.Fatal(err)
			}
			select {
			case r := <-results:
				if r.Err != nil {
					t.Fatal(r.Err)
				}
				return
			default:
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("document operation did not complete")
	}
	finish()
	bytes, err := os.ReadFile(path)
	if err != nil || string(bytes) != m.Snapshot().Text {
		t.Fatal("exact saved bytes", string(bytes), err)
	}
	replacement := "loaded\r\nאב🙂\r\n"
	if err := os.WriteFile(path, []byte(replacement), 0600); err != nil {
		t.Fatal(err)
	}
	epoch := m.Snapshot().Revision
	if err := a.Do(func() { operationErr = session.Open(context.Background(), document.OpenOptions{Path: path}) }); err != nil {
		t.Fatal(err)
	}
	if operationErr != nil {
		t.Fatal(operationErr)
	}
	finish()
	s := m.Snapshot()
	if s.Text != replacement || s.Revision <= epoch || s.History.Dirty || s.History.UndoCount != 0 {
		t.Fatalf("atomic loaded document: %+v", s)
	}
}
