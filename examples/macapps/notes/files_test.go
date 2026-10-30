package notes

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

type fileFixture struct {
	a *app.Application
	m *Model
	f *Files
	r *Repository
}

func filesFixture(t *testing.T, dir string) *fileFixture {
	t.Helper()
	r, err := OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := New()
	offer, err := r.Load(m)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	a, err := app.New(app.Config{QueueCapacity: 4})
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	ref, err := a.OpenWindow(app.Window{ID: "notes", Title: "Notes test", Width: 1100, Height: 740, View: m.View})
	if err != nil {
		t.Fatal(err)
	}
	f, err := AttachFiles(a, ref, m, r, offer)
	if err != nil {
		t.Fatal(err)
	}
	x := &fileFixture{a, m, f, r}
	t.Cleanup(func() {
		a.Close()
		if err := f.Stop(); err != nil {
			t.Error(err)
		}
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	return x
}
func (x *fileFixture) owner(t *testing.T, fn func()) {
	t.Helper()
	if err := x.a.Do(fn); err != nil {
		t.Fatal(err)
	}
}
func (x *fileFixture) finish(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := x.a.Drain(); err != nil {
			t.Fatal(err)
		}
		done := false
		x.owner(t, func() { done = !x.f.busy() && !x.f.clear })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("file operation did not settle")
}
func TestFilesSaveLoadUndoAndAtomicReplacement(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	var expected []Note
	x.owner(t, func() {
		id, _ := x.m.Add()
		if err := x.m.Update(id, "Saved title", "Exact 😀\r\nbody"); err != nil {
			t.Fatal(err)
		}
		expected = x.m.Items()
		if err := x.m.Request(func() error { return x.f.Save(false) }); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if x.m.History().Dirty {
			t.Fatal("save not acknowledged")
		}
		if err := x.m.Undo(); err != nil || !x.m.History().Dirty {
			t.Fatal("undo after save", err)
		}
		if err := x.m.Redo(); err != nil || x.m.History().Dirty {
			t.Fatal("redo saved revision", err)
		}
	})
	data, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeNotebook(data)
	if err != nil || !reflect.DeepEqual(b.Notes, expected) {
		t.Fatal("file round trip", err)
	}
	marker, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	recovery, err := decodeRecovery(marker)
	if err != nil || recovery.Notebook != nil {
		t.Fatal("recovery not cleared", err)
	}
	x.owner(t, func() {
		if err := x.m.Update(x.m.Selected(), "Later", "later"); err != nil {
			t.Fatal(err)
		}
		if err := x.f.main.Open(x.f.ref.Context(), document.OpenOptions{Path: x.r.Path}); !errors.Is(err, document.ErrUnsaved) {
			t.Fatal("dirty open admitted", err)
		}
		if err := x.f.main.Open(x.f.ref.Context(), document.OpenOptions{Path: x.r.Path, DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if !reflect.DeepEqual(expected, x.m.Items()) || x.m.History().UndoCount != 0 || x.m.History().Dirty {
			t.Fatal("open did not replace cleanly")
		}
	})
}
func TestFilesRecoveryRestoreDirtyWithoutPersistedUndo(t *testing.T) {
	dir := privateDirectory(t)
	x := filesFixture(t, dir)
	var expected []Note
	x.owner(t, func() {
		id, _ := x.m.Add()
		_ = x.m.Update(id, "Recovered", "body 😀")
		expected = x.m.Items()
		x.f.Checkpoint()
	})
	x.finish(t)
	x.owner(t, func() {
		if !x.m.History().Dirty {
			t.Fatal("recovery marked main saved")
		}
	})
	x.a.Close()
	if err := x.f.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := x.r.Close(); err != nil {
		t.Fatal(err)
	}
	other := filesFixture(t, dir)
	other.owner(t, func() {
		if other.f.prompt != "recovery" {
			t.Fatal("no recovery offer")
		}
		if err := other.m.titleChanged(other.m.Selected(), "Earlier admitted edit"); err != nil {
			t.Fatal(err)
		}
		if other.m.History().UndoCount == 0 {
			t.Fatal("test requires prior history")
		}
		view := other.f.promptView()
		view.Children[1].Activate()
		if !other.m.History().Dirty || other.m.History().UndoCount != 0 || !reflect.DeepEqual(other.m.Items(), expected) {
			t.Fatal("recovery state/history")
		}
		if err := other.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	other.finish(t)
	data, err := os.ReadFile(other.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeNotebook(data)
	if err != nil || !reflect.DeepEqual(b.Notes, expected) {
		t.Fatal("recovered save", err)
	}
}

type failureStore struct {
	write func(context.Context, string, []byte, int64) (bool, error)
}

func (s failureStore) Read(ctx context.Context, p string, n int64) ([]byte, error) {
	return (document.FileStore{}).Read(ctx, p, n)
}
func (s failureStore) Write(ctx context.Context, p string, b []byte, n int64) (bool, error) {
	return s.write(ctx, p, b, n)
}
func replaceMainStore(t *testing.T, x *fileFixture, store document.Store) {
	t.Helper()
	session, err := document.New(document.Config[Notebook]{Owner: x.a, Window: "notes", Journal: x.m.journal, Path: x.r.Path, MaxBytes: MaxNotebookBytes, Store: store, Codec: document.Codec[Notebook]{Capture: x.m.CaptureNotebook, Encode: EncodeNotebook, Decode: DecodeNotebook, Install: x.m.installNotebook}, OnComplete: x.f.completed})
	if err != nil {
		t.Fatal(err)
	}
	x.f.main = session
}
func TestFailedSaveKeepsOriginalAndDirtyRevision(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	before, _ := os.ReadFile(x.r.Path)
	x.owner(t, func() {
		replaceMainStore(t, x, failureStore{func(context.Context, string, []byte, int64) (bool, error) {
			return false, errors.New("disk unavailable")
		}})
		_ = x.m.Update(x.m.Selected(), "Unsaved", "edited")
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	after, _ := os.ReadFile(x.r.Path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed save replaced original")
	}
	x.owner(t, func() {
		if !x.m.History().Dirty || x.f.main.State().Last.Err == nil {
			t.Fatal("failed save marked clean")
		}
	})
}
func TestLateEditRefusesSavedRevisionAndRetainsNewText(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	started := make(chan struct{})
	release := make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	x.owner(t, func() {
		replaceMainStore(t, x, failureStore{func(ctx context.Context, p string, b []byte, n int64) (bool, error) {
			close(started)
			select {
			case <-release:
				return (document.FileStore{}).Write(ctx, p, b, n)
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}})
		_ = x.m.Update(x.m.Selected(), "Captured", "old")
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker not started")
	}
	x.owner(t, func() {
		if err := x.m.titleChanged(x.m.Selected(), "Later native commit"); err != nil {
			t.Fatal(err)
		}
	})
	close(release)
	x.finish(t)
	x.owner(t, func() {
		if !x.m.History().Dirty || x.m.lookup(x.m.Selected()).value.Title != "Later native commit" || x.f.main.State().Last.Err == nil {
			t.Fatal("stale save erased later edit")
		}
	})
}
func TestCloseCancelDiscardAndRecoveryMarker(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		_ = x.m.Update(x.m.Selected(), "Edited", "text")
		if x.f.RequestClose() {
			t.Fatal("dirty close admitted")
		}
		if x.f.prompt != "close" {
			t.Fatal("missing close prompt")
		}
		x.f.promptView().Dismiss()
		if x.f.prompt != "" || !x.m.History().Dirty {
			t.Fatal("cancel changed document")
		}
		x.f.RequestClose()
		x.f.settleIntent("close", false)
	})
	deadline := time.Now().Add(3 * time.Second)
	for x.f.ref.Context().Err() == nil && time.Now().Before(deadline) {
		_ = x.a.Drain()
		time.Sleep(time.Millisecond)
	}
	if x.f.ref.Context().Err() == nil {
		t.Fatal("discard close did not finish")
	}
	data, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := decodeRecovery(data)
	if err != nil || marker.Notebook != nil {
		t.Fatal("discard left recovery content", err)
	}
}

func TestSaveWaitsForNativeSettlementAndDeduplicatesCompletion(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	var complete *func(error)
	x.owner(t, func() {
		complete, _ = deferSettlement(t, x.m)
		if err := x.m.Request(func() error { return x.f.Save(false) }); err != nil {
			t.Fatal(err)
		}
		if x.f.main.State().Busy {
			t.Fatal("save began before input settled")
		}
		if err := x.m.titleChanged(x.m.Selected(), "Native title before Save"); err != nil {
			t.Fatal(err)
		}
		(*complete)(nil)
		(*complete)(nil)
	})
	x.finish(t)
	data, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeNotebook(data)
	if err != nil || b.Notes[0].Title != "Native title before Save" {
		t.Fatal("missed native title", err)
	}
	x.owner(t, func() {
		if x.m.History().Dirty {
			t.Fatal("settled save stayed dirty")
		}
	})
}
func TestCorruptOpenPreservesLiveHistoryAndFile(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	bad := x.r.Directory + "/invalid.vellumnotes"
	data := []byte(`{"version":1,"notes":[{"ID":"n1"}]}`)
	if err := os.WriteFile(bad, data, 0600); err != nil {
		t.Fatal(err)
	}
	var items []Note
	var state history.State
	x.owner(t, func() {
		_ = x.m.Update(x.m.Selected(), "Keep me", "exact content")
		items = x.m.Items()
		state = x.m.History()
		if err := x.f.main.Open(x.f.ref.Context(), document.OpenOptions{Path: bad, DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if !reflect.DeepEqual(items, x.m.Items()) || x.m.History() != state || x.f.main.State().Last.Err == nil {
			t.Fatal("corrupt open changed live state")
		}
	})
	after, _ := os.ReadFile(bad)
	if !bytes.Equal(data, after) {
		t.Fatal("bad input modified")
	}
}
func TestRecoverySkipsMarkedTextWithoutSettlingIt(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		_ = x.m.Update(x.m.Selected(), "Committed", "committed")
		_, body := x.m.CurrentInput()
		s := body.Snapshot()
		if err := body.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Insert: "あ", Selection: editor.Selection{Anchor: 1, Focus: 1}, Marked: editor.Range{Start: 0, End: 1}}); err != nil {
			t.Fatal(err)
		}
		x.f.Checkpoint()
		if x.f.recovery.State().Busy || !body.Snapshot().Composing {
			t.Fatal("recovery settled native preedit")
		}
		if err := body.CancelComposition(); err != nil {
			t.Fatal(err)
		}
		x.f.Checkpoint()
	})
	x.finish(t)
	data, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	b, err := decodeRecovery(data)
	if err != nil || b.Notebook == nil || b.Notebook.Notes[0].Body != "committed" {
		t.Fatal("recovery leaked preedit", err)
	}
}
func TestWindowCloseCancelsPendingWorkerAndRefusesLateInstall(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	started := make(chan struct{})
	workerDone := make(chan struct{})
	x.owner(t, func() {
		replaceMainStore(t, x, failureStore{func(ctx context.Context, _ string, _ []byte, _ int64) (bool, error) {
			defer close(workerDone)
			close(started)
			<-ctx.Done()
			return false, ctx.Err()
		}})
		_ = x.m.Update(x.m.Selected(), "Keep in memory", "body")
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker not started")
	}
	x.owner(t, func() {
		if x.f.RequestClose() || x.f.ref.Context().Err() != nil {
			t.Fatal("normal close lost pending save")
		}
	})
	x.a.Close()
	if err := x.f.Stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-workerDone:
	default:
		t.Fatal("worker retained after stop")
	}
	if _, err := os.Stat(x.r.Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("canceled save wrote a file", err)
	}
}

func TestAttachCommandCollisionLeavesModelAndRegistryUntouched(t *testing.T) {
	r, err := OpenRepository(privateDirectory(t))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	m := New()
	ref, err := a.OpenWindow(app.Window{ID: "notes", Title: "Attach", Width: 1100, Height: 740, View: m.View})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Commands().Register(command.Command{ID: "other-save", Label: "Other", Shortcut: "Meta+Shift+S", Execute: func(context.Context) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	before := len(m.Commands().Search(""))
	if f, err := AttachFiles(a, ref, m, r, nil); err == nil || f != nil {
		t.Fatal("collision admitted")
	}
	if m.files != nil || len(m.Commands().Search("")) != before {
		t.Fatal("partial model/registry attachment")
	}
	if _, ok := m.Commands().Lookup("notes.save"); ok {
		t.Fatal("first command leaked before later collision")
	}
}
func TestRecoveryWindowClosePreservesOfferWithoutWriting(t *testing.T) {
	dir := privateDirectory(t)
	x := filesFixture(t, dir)
	x.owner(t, func() { _ = x.m.Update(x.m.Selected(), "Interrupted", "body"); x.f.Checkpoint() })
	x.finish(t)
	before, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	x.a.Close()
	if err := x.f.Stop(); err != nil {
		t.Fatal(err)
	}
	_ = x.r.Close()
	next := filesFixture(t, dir)
	next.owner(t, func() {
		if next.f.prompt != "recovery" {
			t.Fatal("missing offer")
		}
		if next.f.RequestClose() {
			t.Fatal("must close through exact owner reference")
		}
	})
	if next.f.ref.Context().Err() == nil {
		t.Fatal("native close ignored during recovery decision")
	}
	after, err := os.ReadFile(next.r.RecoveryPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("closing offer changed recovery", err)
	}
}

func TestRecoveryMarkerFailureKeepsSavedBytesAndExplainsRefusedClose(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	if err := os.Mkdir(x.r.RecoveryPath, 0700); err != nil {
		t.Fatal(err)
	}
	x.owner(t, func() {
		_ = x.m.Update(x.m.Selected(), "Save succeeded", "exact latest body")
		x.f.RequestClose()
		x.f.settleIntent("close", true)
	})
	x.finish(t)
	data, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeNotebook(data)
	if err != nil || b.Notes[0].Title != "Save succeeded" {
		t.Fatal("primary saved bytes lost", err)
	}
	x.owner(t, func() {
		if x.f.ref.Context().Err() != nil || x.m.History().Dirty || !strings.HasPrefix(x.m.status, "Recovery:") {
			t.Fatal("failed marker did not preserve saved notebook/report error", x.m.status)
		}
	})
	if err := os.Remove(x.r.RecoveryPath); err != nil {
		t.Fatal(err)
	}
	x.owner(t, func() { x.f.RequestClose() })
	deadline := time.Now().Add(3 * time.Second)
	for x.f.ref.Context().Err() == nil && time.Now().Before(deadline) {
		_ = x.a.Drain()
		time.Sleep(time.Millisecond)
	}
	if x.f.ref.Context().Err() == nil {
		t.Fatal("close did not retry after storage repaired")
	}
}

func TestAttachedNotesTreeHasUniqueIDsWithEveryFileDialog(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		b, err := x.m.CaptureNotebook()
		if err != nil {
			t.Fatal(err)
		}
		x.f.offer = &Recovery{Version: 1, Notebook: &b}
		for _, prompt := range []string{"", "open", "close", "recovery"} {
			x.f.prompt = prompt
			if err := ui.Validate(x.m.View()); err != nil {
				t.Fatalf("prompt %q: %v", prompt, err)
			}
		}
		x.f.prompt = ""
	})
}

func TestRedoToSavedRevisionClearsEarlierDirtyRecovery(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		_ = x.m.Update(x.m.Selected(), "Saved", "saved body")
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if err := x.m.Undo(); err != nil {
			t.Fatal(err)
		}
		x.f.Checkpoint()
	})
	x.finish(t)
	x.owner(t, func() {
		if err := x.m.Redo(); err != nil || x.m.History().Dirty {
			t.Fatal("redo saved state", err)
		}
		x.f.Checkpoint()
	})
	x.finish(t)
	data, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	r, err := decodeRecovery(data)
	if err != nil || r.Notebook != nil {
		t.Fatal("clean redo retained obsolete recovery", err)
	}
}

func TestSaveReplacementErrorPreservesRecoveryAndRetry(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if err := x.m.Update(x.m.Selected(), "Durability uncertain", "Retain recovery 😀"); err != nil {
			t.Fatal(err)
		}
		x.f.Checkpoint()
	})
	x.finish(t)
	recoveryBefore, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("injected post-replacement directory sync failure")
	x.owner(t, func() {
		replaceMainStore(t, x, failureStore{func(ctx context.Context, p string, b []byte, n int64) (bool, error) {
			written, err := (document.FileStore{}).Write(ctx, p, b, n)
			if err != nil {
				return written, err
			}
			return written, failure
		}})
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	data, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeNotebook(data)
	if err != nil || saved.Notes[0].Title != "Durability uncertain" {
		t.Fatal("replacement bytes lost", err)
	}
	recoveryAfter, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil || !bytes.Equal(recoveryBefore, recoveryAfter) {
		t.Fatal("uncertain save cleared recovery", err)
	}
	x.owner(t, func() {
		state := x.f.main.State()
		if !state.Last.Written || !errors.Is(state.Last.Err, failure) || !x.m.History().Dirty || x.f.ref.Context().Err() != nil {
			t.Fatal("post-replacement error was acknowledged", state)
		}
		replaceMainStore(t, x, document.FileStore{})
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if x.m.History().Dirty || x.f.main.State().Last.Err != nil {
			t.Fatal("retry did not acknowledge save")
		}
		if err := x.m.Undo(); err != nil || !x.m.History().Dirty {
			t.Fatal("retry lost undo", err)
		}
		if err := x.m.Redo(); err != nil || x.m.History().Dirty {
			t.Fatal("retry lost saved revision", err)
		}
	})
	recoveryAfter, err = os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := decodeRecovery(recoveryAfter)
	if err != nil || marker.Notebook != nil {
		t.Fatal("successful retry did not clear recovery", err)
	}
}

func TestFailedRecoveryCheckpointRetriesWithoutChangingSavedNotebook(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	x.owner(t, func() {
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	primaryBefore, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(x.r.RecoveryPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(x.r.RecoveryPath, 0700); err != nil {
		t.Fatal(err)
	}
	x.owner(t, func() {
		if err := x.m.Update(x.m.Selected(), "Recovery retry", "New body"); err != nil {
			t.Fatal(err)
		}
		x.f.Checkpoint()
	})
	x.finish(t)
	x.owner(t, func() {
		if !x.m.History().Dirty || x.f.lastRecovery != 0 || !strings.HasPrefix(x.m.status, "Recovery:") {
			t.Fatal("failed checkpoint acknowledged or stranded", x.m.status)
		}
	})
	if err := os.Remove(x.r.RecoveryPath); err != nil {
		t.Fatal(err)
	}
	x.owner(t, func() { x.f.Checkpoint() })
	x.finish(t)
	data, err := os.ReadFile(x.r.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := decodeRecovery(data)
	if err != nil || recovered.Notebook == nil || recovered.Notebook.Notes[0].Title != "Recovery retry" {
		t.Fatal("recovery retry lost edits", err)
	}
	primaryAfter, err := os.ReadFile(x.r.Path)
	if err != nil || !bytes.Equal(primaryBefore, primaryAfter) {
		t.Fatal("checkpoint changed main file", err)
	}
	x.owner(t, func() {
		if !x.m.History().Dirty || x.f.lastRecovery != x.m.History().Revision || x.f.clear || x.f.busy() {
			t.Fatal("successful checkpoint changed saved state or retained work")
		}
	})
}
