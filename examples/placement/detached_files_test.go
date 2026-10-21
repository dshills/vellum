package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func waitOutput(t *testing.T, f *fileFixture) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_ = f.a.Drain()
		done := false
		f.owner(t, func() { done = f.m.output == nil })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("detached save did not settle")
}
func TestSaveCopyPreservesActivePathDirtyHistoryAndRoundTrip(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	active := f.path
	original, err := os.ReadFile(active)
	if err != nil {
		t.Fatal(err)
	}
	f.place(t, "note", canvas.Point{X: 150, Y: 80})
	for _, dirty := range []bool{true, false} {
		f.path = filepath.Join(t.TempDir(), "copy.vellum")
		var before drawingFile
		var state document.State
		f.owner(t, func() {
			if !dirty {
				_ = f.m.journal.MarkSaved()
			}
			before, _ = drawingCodec(f.m).Capture()
			state = f.m.files.State()
			history := f.m.journal.Snapshot()
			if err := f.m.commands.Invoke(f.ref.Context(), "save-copy"); err != nil {
				t.Fatal(err)
			}
			if f.m.commands.Enabled("new") == nil {
				t.Fatal("output did not gate New")
			}
			if history != f.m.journal.Snapshot() {
				t.Fatal("output mutated journal synchronously")
			}
		})
		waitOutput(t, f)
		data, err := os.ReadFile(f.path)
		if err != nil {
			t.Fatal(err)
		}
		copy, err := drawingCodec(f.m).Decode(data)
		if err != nil || !reflect.DeepEqual(copy, before) {
			t.Fatal("copy roundtrip", err)
		}
		f.owner(t, func() {
			after, _ := drawingCodec(f.m).Capture()
			if !reflect.DeepEqual(after, before) || f.m.files.State() != state || f.m.journal.Snapshot().Dirty != dirty {
				t.Fatal("copy changed active drawing")
			}
		})
	}
	after, err := os.ReadFile(active)
	if err != nil || !reflect.DeepEqual(original, after) {
		t.Fatal("old file overwritten", err)
	}
}
func TestSaveCopyRejectsActiveDestinationAndAliases(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() { _ = f.m.fileAction(f.a, f.ref, "save", false) })
	f.wait(t)
	active := f.path
	f.place(t, "note", canvas.Point{X: 150})
	before, err := os.ReadFile(active)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{active, filepath.Join(filepath.Dir(active), "alias.vellum")} {
		if path != active {
			if err := os.Link(active, path); err != nil {
				t.Fatal(err)
			}
		}
		f.path = path
		f.owner(t, func() {
			if err := f.m.saveCopy(f.a, f.ref, nil); err != nil {
				t.Fatal(err)
			}
		})
		waitOutput(t, f)
		now, err := os.ReadFile(active)
		if err != nil || !reflect.DeepEqual(before, now) {
			t.Fatal("active overwritten", err)
		}
		f.owner(t, func() {
			if !f.m.journal.Snapshot().Dirty || f.m.files.State().Path != active {
				t.Fatal("failure changed active state")
			}
		})
	}
}

type failingOutputStore struct{ document.FileStore }

func (failingOutputStore) Write(context.Context, string, []byte, int64) (bool, error) {
	return false, errors.New("test write failure")
}
func TestSaveCopyCancelFailureAndCloseRetainDocument(t *testing.T) {
	f := fileModel(t, true)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.saveCopy(f.a, f.ref, nil); err != nil {
			t.Fatal(err)
		}
	})
	waitOutput(t, f)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 1 || !f.m.journal.Snapshot().Dirty {
			t.Fatal("canceled copy lost drawing")
		}
	})
	f.owner(t, func() {
		f.m.chooseFile = func(context.Context, app.FileOptions) (string, error) {
			return filepath.Join(t.TempDir(), "copy.vellum"), nil
		}
		if err := f.m.saveCopy(f.a, f.ref, failingOutputStore{}); err != nil {
			t.Fatal(err)
		}
	})
	waitOutput(t, f)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 1 || !f.m.journal.Snapshot().Dirty {
			t.Fatal("failed copy changed drawing")
		}
	})
	entered, done := make(chan struct{}), make(chan struct{})
	f.owner(t, func() {
		f.m.chooseFile = func(ctx context.Context, _ app.FileOptions) (string, error) {
			close(entered)
			<-ctx.Done()
			close(done)
			return "", ctx.Err()
		}
		if err := f.m.saveCopy(f.a, f.ref, nil); err != nil {
			t.Fatal(err)
		}
	})
	cutAwait(t, entered)
	f.owner(t, func() {
		if f.m.requestClose(f.a) {
			t.Fatal("busy close should first cancel")
		}
	})
	cutAwait(t, done)
	waitOutput(t, f)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 1 || !f.m.journal.Snapshot().Dirty {
			t.Fatal("close cancellation lost drawing")
		}
	})
}
