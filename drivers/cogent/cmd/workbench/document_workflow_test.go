package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type workflowStore struct {
	fail    bool
	entered chan struct{}
	release chan struct{}
}

func (s *workflowStore) Read(c context.Context, p string, m int64) ([]byte, error) {
	return (document.FileStore{}).Read(c, p, m)
}
func (s *workflowStore) Write(c context.Context, p string, b []byte, m int64) (bool, error) {
	if s.entered != nil {
		close(s.entered)
		<-s.release
	}
	if s.fail {
		return false, errors.New("save rejected")
	}
	return (document.FileStore{}).Write(c, p, b, m)
}
func TestWorkflowSaveDiscardCancelAndFailedSave(t *testing.T) {
	for _, decision := range []string{"save", "discard", "cancel", "failure", "save-open", "discard-open"} {
		t.Run(decision, func(t *testing.T) {
			a, _ := app.New(app.Config{})
			defer a.Close()
			j, _ := history.New(8)
			value := 1
			status := ""
			w := &documentWorkflow{owner: a, window: "main", status: func(s string) { status = s }}
			a.Open(app.Window{ID: "main", Title: "Document", Width: 200, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }, CloseRequest: w.requestClose})
			w.windowRef = currentFixtureRef(a, "main")
			path := filepath.Join(t.TempDir(), "doc")
			store := &workflowStore{fail: decision == "failure"}
			session, err := document.New(document.Config[int]{Owner: a, Window: "main", Journal: j, Store: store, Choose: func(context.Context, app.FileOptions) (string, error) { return path, nil }, Codec: document.Codec[int]{Capture: func() (int, error) { return value, nil }, Encode: func(v int) ([]byte, error) { return []byte(strconv.Itoa(v)), nil }, Decode: func(b []byte) (int, error) { return strconv.Atoi(string(b)) }, Install: func(v int) error { value = v; return nil }}, OnComplete: w.complete})
			if err != nil {
				t.Fatal(err)
			}
			w.session = session
			if decision == "discard-open" {
				if written, err := (document.FileStore{}).Write(context.Background(), path, []byte("99"), 32); !written || err != nil {
					t.Fatal(err)
				}
			}
			a.Do(func() {
				if err := j.Apply(history.Change{Label: "Edit", Apply: func() error { value = 7; return nil }, Revert: func() error { value = 1; return nil }}); err != nil {
					t.Fatal(err)
				}
				if decision == "save-open" || decision == "discard-open" {
					if err := w.open(a.Context()); err != nil || w.pending != "open" {
						t.Fatal(err, w.pending)
					}
				} else {
					if !errors.Is(a.RequestCloseWindow(currentFixtureRef(a, "main")), app.ErrCloseRefused) || w.pending != "close" {
						t.Fatal("dirty close not deferred")
					}
				}
				switch decision {
				case "save", "failure", "save-open":
					w.saveDecision()
				case "discard", "discard-open":
					w.discard()
				case "cancel":
					w.cancel()
				}
			})
			deadline := time.Now().Add(3 * time.Second)
			for a.Context().Err() == nil && session.State().Busy && time.Now().Before(deadline) {
				a.Drain()
				time.Sleep(time.Millisecond)
			}
			switch decision {
			case "save":
				if a.Context().Err() == nil || j.Snapshot().Dirty {
					t.Fatal("acknowledged save didn't close", status)
				}
			case "discard":
				if a.Context().Err() == nil {
					t.Fatal("discard didn't close")
				}
			case "cancel":
				if a.Context().Err() != nil || w.pending != "" || !j.Snapshot().Dirty || value != 7 {
					t.Fatal("cancel lost changes")
				}
			case "failure":
				if a.Context().Err() != nil || !j.Snapshot().Dirty || w.pending != "close" || session.State().Busy {
					t.Fatal("failed save closed", status)
				}
			case "save-open":
				if a.Context().Err() != nil || j.Snapshot().Dirty || value != 7 || w.pending != "" || session.State().Busy {
					t.Fatal("save then open failed", status)
				}
			case "discard-open":
				if value != 99 || j.Snapshot().Dirty || w.pending != "" || a.Context().Err() != nil {
					t.Fatal("discard then open failed", status)
				}
			}
		})
	}
}

func TestSaveDecisionCancelAndLateEditKeepWindow(t *testing.T) {
	for _, lateEdit := range []bool{false, true} {
		t.Run(strconv.FormatBool(lateEdit), func(t *testing.T) {
			a, _ := app.New(app.Config{})
			defer a.Close()
			j, _ := history.New(8)
			value := 1
			store := &workflowStore{entered: make(chan struct{}), release: make(chan struct{})}
			flow := &documentWorkflow{owner: a, window: "main", status: func(string) {}}
			a.Open(app.Window{ID: "main", Title: "Document", Width: 100, Height: 100, CloseRequest: flow.requestClose, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
			flow.windowRef = currentFixtureRef(a, "main")
			path := filepath.Join(t.TempDir(), "doc")
			session, err := document.New(document.Config[int]{Owner: a, Window: "main", Journal: j, Store: store, Path: path, Codec: document.Codec[int]{Capture: func() (int, error) { return value, nil }, Encode: func(v int) ([]byte, error) { return []byte(strconv.Itoa(v)), nil }, Decode: func(b []byte) (int, error) { return strconv.Atoi(string(b)) }, Install: func(v int) error { value = v; return nil }}, OnComplete: flow.complete})
			if err != nil {
				t.Fatal(err)
			}
			flow.session = session
			a.Do(func() {
				j.Apply(history.Change{Label: "Edit", Apply: func() error { value = 7; return nil }, Revert: func() error { value = 1; return nil }})
				if flow.requestClose() {
					t.Fatal("unsaved close accepted")
				}
				flow.saveDecision()
			})
			select {
			case <-store.entered:
			case <-time.After(2 * time.Second):
				t.Fatal("worker didn't start")
			}
			a.Do(func() {
				if lateEdit {
					j.Apply(history.Change{Label: "New edit", Apply: func() error { value = 9; return nil }, Revert: func() error { value = 7; return nil }})
				} else {
					flow.cancel()
				}
			})
			close(store.release)
			deadline := time.Now().Add(3 * time.Second)
			for session.State().Busy && time.Now().Before(deadline) {
				a.Drain()
				time.Sleep(time.Millisecond)
			}
			if session.State().Busy || a.Context().Err() != nil {
				t.Fatal("save closed despite canceled/stale decision")
			}
			if lateEdit {
				if !errors.Is(session.State().Last.Err, history.ErrStale) || !j.Snapshot().Dirty || value != 9 || flow.pending != "close" {
					t.Fatal("late edit lost")
				}
			} else {
				if session.State().Last.Err != nil || j.Snapshot().Dirty || flow.pending != "" {
					t.Fatal("canceling prompt corrupted save")
				}
			}
		})
	}
}
