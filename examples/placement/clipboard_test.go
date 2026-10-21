package main

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
)

func copiedText(t *testing.T, m *model) string {
	t.Helper()
	objects, err := m.copiedSelection()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(selectionClipboard{"vellum-placement-selection", 1, objects})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func TestCopyPasteDuplicateGeometryIdentityHistoryAndBranch(t *testing.T) {
	m := editModel(t)
	m.snap = false
	m.scene.Objects[0].Width = 87.25
	m.scene.Objects[2].Y = 13.5
	m.setSelection([]string{"block-1", "note-3"}, "block-1")
	_ = m.journal.MarkSaved()
	before := slices.Clone(m.scene.Objects)
	text := copiedText(t, m)
	if m.journal.Snapshot().Dirty {
		t.Fatal("copy dirtied drawing")
	}
	if err := m.pasteSelection(text); err != nil {
		t.Fatal(err)
	}
	after := slices.Clone(m.scene.Objects)
	for j, i := range []int{0, 2} {
		o := before[i]
		got := after[3+j]
		if got.ID == o.ID || got.X != o.X+20 || got.Y != o.Y+20 || got.Width != o.Width || got.Height != o.Height || got.Color != o.Color {
			t.Fatal("paste did not preserve relative geometry", got)
		}
	}
	if !slices.Equal(m.selectedIDs(), []string{"block-4", "note-5"}) || m.next != 5 || m.journal.Snapshot().UndoCount != 4 {
		t.Fatal("paste identity/selection/history")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) || m.scene.Selected != "block-1" {
		t.Fatal("paste Undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
		t.Fatal("paste Redo changed IDs/geometry", err)
	}
	if err := m.pasteSelection(text); err != nil || m.scene.Objects[5].X != before[0].X+40 {
		t.Fatal("paste cascade", err)
	}
	_ = m.journal.Undo()
	if err := m.duplicateSelected(); err != nil || m.next != 9 || m.scene.Objects[5].ID != "block-8" || m.scene.Objects[5].X != after[3].X+20 || m.journal.Snapshot().RedoCount != 0 || m.pasteCount != 2 {
		t.Fatal("duplicate branch/reused ID/clipboard state", err)
	}
	duplicated := slices.Clone(m.scene.Objects)
	_ = m.journal.Undo()
	_ = m.journal.Redo()
	if !slices.Equal(m.scene.Objects, duplicated) {
		t.Fatal("duplicate Redo")
	}
}

func TestClipboardRejectsAtomicallyAndBounded(t *testing.T) {
	m := editModel(t)
	m.setSelection([]string{"block-1", "note-3"}, "note-3")
	before := slices.Clone(m.scene.Objects)
	next := m.next
	text := copiedText(t, m)
	for _, bad := range []string{"", "ordinary text", text + " {}", strings.Replace(text, "\"version\":1", "\"version\":4", 1), strings.Replace(text, "\"kind\":\"block\"", "\"kind\":\"unknown\"", 1), strings.Repeat("x", drawingByteLimit+1)} {
		if m.pasteSelection(bad) == nil {
			t.Fatal("invalid clipboard accepted")
		}
	}
	objects, _ := m.copiedSelection()
	for _, failure := range []string{"position", "size", "capacity", "IDs", "locked", "notes", "active"} {
		t.Run(failure, func(t *testing.T) {
			m := editModel(t)
			m.setSelection([]string{"note-3"}, "note-3")
			data := slices.Clone(objects)
			switch failure {
			case "position":
				data[1].X = 9990
			case "size":
				data[1].Height = math.NaN()
			case "capacity":
				data = make([]copiedComponent, 511)
			case "IDs":
				m.next = math.MaxUint64 - 1
			case "locked":
				m.locked = true
			case "notes":
				m.rejectNotes = true
			case "active":
				m.move = &objectMove{}
			}
			prior, n := slices.Clone(m.scene.Objects), m.next
			if err := m.insertCopies(data, 20, "Paste"); err == nil || !slices.Equal(m.scene.Objects, prior) || m.next != n || m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("rejected insert partially applied", err)
			}
		})
	}
	if !slices.Equal(m.scene.Objects, before) || m.next != next || m.pasteCount != 0 {
		t.Fatal("malformed paste mutated state")
	}
}

type appClipboard struct {
	mu      sync.Mutex
	text    string
	fail    bool
	started chan struct{}
	done    chan struct{}
}

func (s *appClipboard) ReadClipboard(ctx context.Context) (string, error) {
	if s.started != nil {
		close(s.started)
		<-ctx.Done()
		close(s.done)
		return "", ctx.Err()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		return "", errors.New("clipboard read failure")
	}
	return s.text, nil
}
func (s *appClipboard) WriteClipboard(_ context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		return errors.New("clipboard write failure")
	}
	s.text = text
	return nil
}
func (*appClipboard) ChooseFile(context.Context, app.FileOptions) (string, error) {
	return "", app.ErrCanceled
}
func (*appClipboard) SetMenu(context.Context, []app.MenuGroup) error { return nil }
func waitClipboard(t *testing.T, f *fileFixture) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_ = f.a.Drain()
		done := false
		f.owner(t, func() { done = f.m.clipboardJob == nil })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("clipboard did not settle")
}
func TestClipboardWorkerCopyPasteFailuresAndClose(t *testing.T) {
	f := fileModel(t, false)
	s := &appClipboard{}
	f.a.SetServices(s)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 150, Y: 50})
	f.owner(t, func() {
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		_ = f.m.journal.MarkSaved()
		if err := f.m.commands.Invoke(f.ref.Context(), "copy"); err != nil {
			t.Fatal(err)
		}
		if f.m.commands.Enabled("paste") == nil || f.m.editAvailable() == nil {
			t.Fatal("worker did not gate edits")
		}
	})
	waitClipboard(t, f)
	f.owner(t, func() {
		if f.m.journal.Snapshot().Dirty {
			t.Fatal("async Copy dirtied")
		}
		f.m.setSelection(nil, "")
		if err := f.m.commands.Invoke(f.ref.Context(), "paste"); err != nil {
			t.Fatal(err)
		}
	})
	waitClipboard(t, f)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 4 || f.m.next != 4 || len(f.m.selectedIDs()) != 2 {
			t.Fatal("async Paste failed")
		}
	})
	s.mu.Lock()
	s.fail = true
	s.mu.Unlock()
	f.owner(t, func() {
		if err := f.m.commands.Invoke(f.ref.Context(), "paste"); err != nil {
			t.Fatal(err)
		}
	})
	waitClipboard(t, f)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 4 || f.m.clipboardJob != nil {
			t.Fatal("failed read changed document/busy")
		}
	})
	s.mu.Lock()
	s.fail = false
	s.text = strings.Repeat("x", drawingByteLimit+1)
	s.mu.Unlock()
	f.owner(t, func() {
		if err := f.m.commands.Invoke(f.ref.Context(), "paste"); err != nil {
			t.Fatal(err)
		}
	})
	waitClipboard(t, f)
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 4 || !strings.Contains(f.m.status, "byte limit exceeded") {
			t.Fatal("oversized worker read was not rejected", f.m.status)
		}
	})
	slow := &appClipboard{started: make(chan struct{}), done: make(chan struct{})}
	f.a.SetServices(slow)
	f.owner(t, func() {
		_ = f.m.journal.MarkSaved()
		if err := f.m.commands.Invoke(f.ref.Context(), "paste"); err != nil {
			t.Fatal(err)
		}
	})
	select {
	case <-slow.started:
	case <-time.After(time.Second):
		t.Fatal("worker not started")
	}
	f.owner(t, func() {
		if !f.m.requestClose(f.a) || f.m.clipboardJob != nil {
			t.Fatal("close did not retire clipboard")
		}
	})
	select {
	case <-slow.done:
	case <-time.After(time.Second):
		t.Fatal("close did not cancel worker")
	}
	_ = f.a.Drain()
	f.owner(t, func() {
		if len(f.m.scene.Objects) != 4 {
			t.Fatal("late completion pasted after close")
		}
	})
}

func TestCopiedDrawingFileRoundTrip(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200, Y: 50})
	var saved []canvas.Rect
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		if err := f.m.pasteSelection(copiedText(t, f.m)); err != nil {
			t.Fatal(err)
		}
		saved = slices.Clone(f.m.scene.Objects)
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if err := f.m.deleteSelected(); err != nil {
			t.Fatal(err)
		}
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if !slices.Equal(f.m.scene.Objects, saved) || f.m.next != 4 || f.m.scene.Selected != "note-4" || f.m.pasteCount != 0 || f.m.journal.Snapshot().UndoCount != 0 {
			t.Fatal("copied drawing roundtrip lost geometry/IDs")
		}
	})
}
