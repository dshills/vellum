package main

import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func bindPrefs(t *testing.T, f *fileFixture, dir string, store document.Store) *editorPreferences {
	t.Helper()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	var r *editorPreferences
	f.owner(t, func() {
		var err error
		r, err = newEditorPreferences(f.a, f.ref, f.m, preferenceConfig{dir, store})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Cleanup(func() { _ = r.Stop() })
	return r
}
func waitPrefs(t *testing.T, f *fileFixture) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_ = f.a.Drain()
		loading := true
		f.owner(t, func() { loading = f.m.preferencesLoading })
		if !loading {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("preference load did not settle")
}
func TestPreferencesPersistSeparatelyWithPrivateModeAndNoHistory(t *testing.T) {
	f := fileModel(t, false)
	dir := t.TempDir()
	r := bindPrefs(t, f, dir, nil)
	waitPrefs(t, f)
	var want preferenceValues
	f.owner(t, func() {
		state := f.m.journal.Snapshot()
		path := f.m.files.State().Path
		f.m.gridHidden = true
		f.m.gridSpacing = 50
		f.m.aspectLocked = true
		f.m.selectionSubtract = true
		f.m.preferencesChanged(prefAll)
		want = f.m.currentPreferences()
		if state != f.m.journal.Snapshot() || path != f.m.files.State().Path {
			t.Fatal("preferences changed document")
		}
	})
	if err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(filepath.Join(dir, preferenceFilename))
	if err != nil || stat.Mode().Perm() != 0600 {
		t.Fatal("private preferences", err, stat)
	}
	g := fileModel(t, false)
	bindPrefs(t, g, dir, nil)
	waitPrefs(t, g)
	g.owner(t, func() {
		if g.m.currentPreferences() != want {
			t.Fatal("restart lost settings", g.m.currentPreferences(), want)
		}
	})
	if err := r.Record(want, prefAll); err != app.ErrClosed {
		t.Fatal("closed writer admitted", err)
	}
}

type delayedPreferenceStore struct {
	document.FileStore
	entered, release chan struct{}
}

func (s *delayedPreferenceStore) Read(ctx context.Context, path string, max int64) ([]byte, error) {
	close(s.entered)
	select {
	case <-s.release:
		return s.FileStore.Read(ctx, path, max)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func TestPreferenceStartupMergeGatesGesturesAndRetiresOnClose(t *testing.T) {
	dir := t.TempDir()
	_ = os.Chmod(dir, 0700)
	disk := defaultPreferences()
	disk.GridSpacing = 100
	disk.AspectLocked = true
	data, _ := json.Marshal(disk)
	_ = os.WriteFile(filepath.Join(dir, preferenceFilename), data, 0600)
	f := fileModel(t, false)
	store := &delayedPreferenceStore{entered: make(chan struct{}), release: make(chan struct{})}
	r := bindPrefs(t, f, dir, store)
	cutAwait(t, store.entered)
	f.owner(t, func() {
		if f.m.selectionAvailable() == nil || f.m.commands.Enabled("new") == nil {
			t.Fatal("loading did not gate interaction")
		}
		// A programmatic preference change before owner admission must win for its field.
		f.m.gridHidden = true
		f.m.preferencesChanged(prefGridVisible)
	})
	close(store.release)
	waitPrefs(t, f)
	f.owner(t, func() {
		if !f.m.gridHidden || f.m.spacing() != 100 || !f.m.aspectLocked {
			t.Fatal("late load replaced newer edits or lost untoggled fields")
		}
	})
	if err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	g := fileModel(t, false)
	blocked := &delayedPreferenceStore{entered: make(chan struct{}), release: make(chan struct{})}
	q := bindPrefs(t, g, t.TempDir(), blocked)
	cutAwait(t, blocked.entered)
	g.owner(t, func() {
		if !g.m.requestClose(g.a) || g.m.preferencesLoading {
			t.Fatal("startup close did not retire")
		}
	})
	_ = q.Stop()
	_ = g.a.Drain()
	g.owner(t, func() {
		if g.m.currentPreferences() != defaultPreferences() {
			t.Fatal("retired load changed model")
		}
	})
}
func TestPreferenceCorruptCacheDefaultsAndStrictSchema(t *testing.T) {
	for _, s := range []string{`{}`, `{"version":2,"grid_spacing":20}`, `{"version":1,"grid_spacing":7}`, `{"version":1,"grid_spacing":20,"extra":1}`, `{"version":1,"grid_spacing":20} {}`} {
		if _, err := decodePreferences([]byte(s)); err == nil {
			t.Fatal("bad schema accepted", s)
		}
	}
	f := fileModel(t, false)
	dir := t.TempDir()
	_ = os.Chmod(dir, 0700)
	_ = os.WriteFile(filepath.Join(dir, preferenceFilename), []byte(`{"broken":true}`), 0600)
	bindPrefs(t, f, dir, nil)
	waitPrefs(t, f)
	f.owner(t, func() {
		if f.m.currentPreferences() != defaultPreferences() || f.m.fileBusy() {
			t.Fatal("corrupt settings blocked startup")
		}
	})
}

func TestPreferenceCoalescedUpdatesFlushLatestAndJoin(t *testing.T) {
	f := fileModel(t, false)
	dir := t.TempDir()
	store := &recentBlockingStore{entered: make(chan struct{}), release: make(chan struct{})}
	r := bindPrefs(t, f, dir, store)
	waitPrefs(t, f)
	want := defaultPreferences()
	want.GridSpacing = 50
	if err := r.Record(want, prefAll); err != nil {
		t.Fatal(err)
	}
	cutAwait(t, store.entered)
	for i := 0; i < 100; i++ {
		want.GridHidden = i%2 == 0
		want.AspectLocked = i%3 == 0
		want.SelectionSubtract = i%5 == 0
		if err := r.Record(want, prefAll); err != nil {
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
		t.Fatal("Stop did not join")
	}
	data, err := os.ReadFile(filepath.Join(dir, preferenceFilename))
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodePreferences(data)
	if err != nil || got != want {
		t.Fatal("latest coalesced settings lost", err, got, want)
	}
}
