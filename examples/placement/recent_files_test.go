package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
)

func waitRecent(t *testing.T, r *recentDrawings) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		loading, _ := r.Status()
		if !loading {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("recent load did not finish")
}
func bindTestRecents(t *testing.T, f *fileFixture, directory string, store document.Store) *recentDrawings {
	t.Helper()
	// This helper owns the disposable fixture directory. Production refuses
	// shared directories rather than changing their existing permissions.
	if err := os.Chmod(directory, 0700); err != nil {
		t.Fatal(err)
	}
	var r *recentDrawings
	f.owner(t, func() {
		var err error
		r, err = newRecentDrawings(f.a, f.ref, recentConfig{Directory: directory, Store: store})
		if err != nil {
			t.Fatal(err)
		}
		f.m.recent = r
	})
	t.Cleanup(func() { _ = r.Stop() })
	waitRecent(t, r)
	return r
}

func TestRecentDrawingsBoundOrderDetachmentPersistenceAndPrivateMode(t *testing.T) {
	f := fileModel(t, false)
	directory := t.TempDir()
	r := bindTestRecents(t, f, directory, nil)
	for i := 0; i < 12; i++ {
		if err := r.Record(filepath.Join(directory, fmt.Sprintf("Drawing-%d.vellum", i))); err != nil {
			t.Fatal(err)
		}
	}
	latest := filepath.Join(directory, "Drawing-5.vellum")
	if err := r.Record(latest); err != nil {
		t.Fatal(err)
	}
	paths := r.Paths()
	if len(paths) != 10 || paths[0] != latest || paths[1] != filepath.Join(directory, "Drawing-11.vellum") {
		t.Fatal("recent bound/order", paths)
	}
	paths[0] = "changed"
	if r.Paths()[0] != latest {
		t.Fatal("Paths retained borrowed slice")
	}
	want := r.Paths()
	if err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := r.Record(latest); !errors.Is(err, app.ErrClosed) {
		t.Fatal("closed recent writer accepted update", err)
	}
	info, err := os.Stat(filepath.Join(directory, recentFilename))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("private cache", info, err)
	}
	g := fileModel(t, false)
	loaded := bindTestRecents(t, g, directory, nil)
	if !slices.Equal(want, loaded.Paths()) {
		t.Fatal("across launch order", loaded.Paths(), want)
	}
}

func TestRecentSuccessfulOperationsOnlyAndRememberedUnsavedOpen(t *testing.T) {
	f := fileModel(t, false)
	r := bindTestRecents(t, f, t.TempDir(), nil)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	if !slices.Equal(r.Paths(), []string{f.path}) {
		t.Fatal("successful Save not recorded", r.Paths())
	}
	f.place(t, "note", canvas.Point{X: 200})
	var before drawingFile
	f.owner(t, func() {
		var err error
		before, err = drawingCodec(f.m).Capture()
		if err != nil {
			t.Fatal(err)
		}
		if err = f.m.openRecent(f.a, f.ref, f.path); err != nil {
			t.Fatal(err)
		}
		if f.m.prompt != "open" || f.m.pendingOpenPath != f.path {
			t.Fatal("recent destination not remembered")
		}
		f.m.filePrompt(f.a, f.ref).Children[2].Activate()
		if f.m.prompt != "" || f.m.pendingOpenPath != "" {
			t.Fatal("cancel retained destination")
		}
	})
	f.owner(t, func() {
		after, err := drawingCodec(f.m).Capture()
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("cancel changed drawing", err)
		}
		if err = f.m.openRecent(f.a, f.ref, f.path); err != nil {
			t.Fatal(err)
		}
		f.m.filePrompt(f.a, f.ref).Children[1].Activate()
	})
	f.wait(t)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 1 || f.m.pendingOpenPath != "" || f.m.journal.Snapshot().Dirty {
			t.Fatal("explicit recent Open did not restore saved drawing")
		}
	})
	missing := filepath.Join(t.TempDir(), "missing.vellum")
	if err := r.Record(missing); err != nil {
		t.Fatal(err)
	}
	f.place(t, "note", canvas.Point{X: 200})
	var stateBefore drawingFile
	f.owner(t, func() {
		stateBefore, _ = drawingCodec(f.m).Capture()
		if err := f.m.openRecent(f.a, f.ref, missing); err != nil {
			t.Fatal(err)
		}
		if f.m.prompt != "open" {
			t.Fatal("missing recent skipped unsaved prompt")
		}
		f.m.filePrompt(f.a, f.ref).Children[1].Activate()
	})
	f.wait(t)
	f.owner(t, func() {
		after, err := drawingCodec(f.m).Capture()
		if err != nil || !reflect.DeepEqual(stateBefore, after) || f.m.files.State().Last.Err == nil || !f.m.journal.Snapshot().Dirty {
			t.Fatal("failed recent read changed drawing", err)
		}
	})
	if r.Paths()[0] != missing || len(r.Paths()) != 2 {
		t.Fatal("failed open changed recents", r.Paths())
	}
}

func TestDrawingChooserStripsOnlySuggestedRepeatedExtension(t *testing.T) {
	for _, name := range []string{"Plan.vellum", "Plan.VELLUM.vellum", "Plan", ".vellum.VeLlUm"} {
		f := fileModel(t, false)
		f.place(t, "block", canvas.Point{})
		chosen := filepath.Join(t.TempDir(), "confirmed.vellum.vellum")
		options := make(chan app.FileOptions, 1)
		f.owner(t, func() {
			f.m.files = nil
			var err error
			// Rebinding uses fresh registry entries in this fixture only.
			f.m.commands = command.New()
			err = f.m.bindFiles(f.a, f.ref, func(ctx context.Context, o app.FileOptions) (string, error) { options <- o; return chosen, nil })
			if err != nil {
				t.Fatal(err)
			}
			if err = f.m.files.Save(f.ref.Context(), document.SaveOptions{Path: filepath.Join(t.TempDir(), name)}); err != nil {
				t.Fatal(err)
			}
		})
		f.wait(t)
		f.owner(t, func() {
			if err := f.m.fileAction(f.a, f.ref, "save-as", false); err != nil {
				t.Fatal(err)
			}
		})
		f.wait(t)
		select {
		case option := <-options:
			if option.Filename != drawingSuggestion(name) || option.Mode != app.SaveFile {
				t.Fatal("wrong chooser suggestion", option)
			}
		default:
			t.Fatal("chooser did not run")
		}
		if _, err := os.Stat(chosen); err != nil {
			t.Fatal("confirmed destination was rewritten", err)
		}
		f.owner(t, func() {
			if f.m.files.State().Path != chosen {
				t.Fatal("session path rewritten")
			}
		})
	}
}

type recentBlockingStore struct {
	document.FileStore
	entered, release, done chan struct{}
	once                   sync.Once
}

func (s *recentBlockingStore) Read(ctx context.Context, path string, limit int64) ([]byte, error) {
	return nil, os.ErrNotExist
}
func (s *recentBlockingStore) Write(ctx context.Context, path string, data []byte, limit int64) (bool, error) {
	s.once.Do(func() { close(s.entered) })
	select {
	case <-s.release:
		return s.FileStore.Write(ctx, path, data, limit)
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func TestRecentCoalescedWorkerCloseJoinAndRetiredStartup(t *testing.T) {
	f := fileModel(t, false)
	store := &recentBlockingStore{entered: make(chan struct{}), release: make(chan struct{})}
	r := bindTestRecents(t, f, t.TempDir(), store)
	first := filepath.Join(t.TempDir(), "one.vellum")
	if err := r.Record(first); err != nil {
		t.Fatal(err)
	}
	select {
	case <-store.entered:
	case <-time.After(time.Second):
		t.Fatal("writer not started")
	}
	for i := 0; i < 25; i++ {
		if err := r.Record(filepath.Join(t.TempDir(), fmt.Sprintf("%d.vellum", i))); err != nil {
			t.Fatal(err)
		}
	}
	close(store.release)
	if err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-r.done:
	default:
		t.Fatal("Stop returned without joining")
	}
	if len(r.Paths()) != 10 {
		t.Fatal("unbounded recent queue")
	}
	if err := f.a.Drain(); err != nil {
		t.Fatal(err)
	}
	if f.a.Pending().ReadyCompletions != 0 {
		t.Fatal("retired startup callback remained queued")
	}
}

func TestRecentCacheRejectsMalformedAndRelativePaths(t *testing.T) {
	for _, data := range []string{`{}`, `{"version":2,"paths":[]}`, `{"version":1,"paths":["relative.vellum"]}`, `{"version":1,"paths":["/tmp/a","/tmp/a"]}`, `{"version":1,"paths":[]} {}`} {
		if _, err := decodeRecents([]byte(data)); err == nil {
			t.Fatal("bad cache accepted")
		}
	}
	for _, name := range []string{"a.vellum", "/tmp/bad\x00name"} {
		if _, err := recentPath(name); err == nil {
			t.Fatal("bad recent path accepted")
		}
	}
}
