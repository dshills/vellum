package editor_test

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

func TestDocumentSessionSharesJournalAndPreservesExactBytes(t *testing.T) {
	const original = "😀 e\u0301\r\nlast\r\n\r\n"
	m, err := editor.New(original, editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "main", Title: "Editor", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	var results []document.Result
	s, err := document.New(document.Config[string]{Owner: a, Window: "main", Journal: m.Journal(), Codec: document.Codec[string]{
		Capture: func() (string, error) {
			snapshot := m.Snapshot()
			if snapshot.Composing {
				return "", editor.ErrComposition
			}
			return snapshot.Text, nil
		},
		Encode:  func(text string) ([]byte, error) { return []byte(text), nil },
		Decode:  func(bytes []byte) (string, error) { return string(bytes), nil },
		Install: func(text string) error { return m.Install(text, editor.Selection{}) },
	}, OnComplete: func(r document.Result) { results = append(results, r) }})
	if err != nil {
		t.Fatal(err)
	}
	owner := func(fn func()) {
		t.Helper()
		if e := a.Do(fn); e != nil {
			t.Fatal(e)
		}
	}
	finish := func(n int) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if e := a.Drain(); e != nil {
				t.Fatal(e)
			}
			if len(results) >= n {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("document completion timed out")
	}
	path := filepath.Join(t.TempDir(), "exact.txt")
	owner(func() {
		if e := m.Replace(original+"edit", editor.Selection{0, 2}); e != nil {
			t.Fatal(e)
		}
		if e := s.Save(context.Background(), document.SaveOptions{Path: path}); e != nil {
			t.Fatal(e)
		}
	})
	finish(1)
	if results[0].Err != nil || m.Journal().Snapshot().Dirty {
		t.Fatal(results, m.Snapshot())
	}
	bytes, err := os.ReadFile(path)
	if err != nil || string(bytes) != original+"edit" {
		t.Fatal(string(bytes), err)
	}
	savedRevision := m.Journal().Snapshot().Revision
	owner(func() {
		if e := m.Replace("other", editor.Selection{}); e != nil {
			t.Fatal(e)
		}
		if e := s.Open(context.Background(), document.OpenOptions{Path: path, DiscardChanges: true}); e != nil {
			t.Fatal(e)
		}
	})
	finish(2)
	if results[1].Err != nil || m.Snapshot().Text != original+"edit" || m.Journal().Snapshot().Dirty || m.Journal().Snapshot().UndoCount != 0 || m.Journal().Snapshot().Revision != savedRevision+2 {
		t.Fatal(results, m.Snapshot())
	}
	// A decoding install failure must preserve model and history despite worker IO.
	if err = os.WriteFile(path, []byte{0xff}, 0600); err != nil {
		t.Fatal(err)
	}
	before := m.Snapshot()
	owner(func() {
		if e := s.Open(context.Background(), document.OpenOptions{Path: path}); e != nil {
			t.Fatal(e)
		}
	})
	finish(3)
	if !errors.Is(results[2].Err, editor.ErrInvalidText) || m.Snapshot() != before {
		t.Fatal(results, m.Snapshot())
	}
}
