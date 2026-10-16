package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

type memoryStore struct {
	read  func(context.Context) ([]byte, error)
	write func(context.Context, []byte) (bool, error)
}

func (s memoryStore) Read(ctx context.Context, _ string, _ int64) ([]byte, error) { return s.read(ctx) }
func (s memoryStore) Write(ctx context.Context, _ string, b []byte, _ int64) (bool, error) {
	return s.write(ctx, b)
}

type fixture struct {
	a          *app.Application
	w          *editorWindow
	ack        func(error)
	ackContext context.Context
	path       string
}

func setup(t *testing.T, store document.Store) *fixture {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	f := &fixture{a: a, path: filepath.Join(t.TempDir(), "document.txt")}
	f.w, err = newEditorWindow(a, "one", windowOptions{store: store, settle: func(ctx context.Context, _ *app.Application, _ string, _ *editor.Model, ack func(error)) error {
		if f.ack != nil {
			t.Fatal("overlapping acknowledgment")
		}
		f.ack = ack
		f.ackContext = ctx
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func (f *fixture) owner(t *testing.T, fn func()) {
	t.Helper()
	if err := f.a.Do(fn); err != nil {
		t.Fatal(err)
	}
}
func (f *fixture) edit(t *testing.T, text string) {
	t.Helper()
	f.owner(t, func() {
		if err := f.w.model.Replace(text, editor.Selection{}); err != nil {
			t.Fatal(err)
		}
	})
}
func (f *fixture) acknowledge(t *testing.T, err error) {
	t.Helper()
	f.owner(t, func() {
		if f.ack == nil {
			t.Fatal("no native acknowledgment pending")
		}
		ack := f.ack
		f.ack = nil
		ack(err)
	})
}
func (f *fixture) finish(t *testing.T, count uint64) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := f.a.Drain(); err != nil {
			t.Fatal(err)
		}
		done := false
		f.owner(t, func() { done = f.w.resultCount >= count })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("document completion timeout")
}
func waitEntered(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not enter")
	}
}

func TestSaveWaitsNativeAcknowledgmentAndPreservesExactBytes(t *testing.T) {
	f := setup(t, document.FileStore{})
	f.edit(t, "old")
	f.owner(t, func() {
		f.w.start(actionSave, f.path)
		if !f.w.settling || f.w.session.State().Busy {
			t.Fatal("save captured before native acknowledgment")
		}
	})
	if _, err := os.Stat(f.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("save started prematurely", err)
	}
	text := "a😀\r\nאבג العربية\r\n\r\n"
	f.edit(t, text)
	f.acknowledge(t, nil)
	f.finish(t, 1)
	got, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte(text)) {
		t.Fatalf("byte conversion: %q", got)
	}
	f.owner(t, func() {
		if f.w.last.Err != nil || f.w.session.State().Dirty {
			t.Fatal(f.w.last, f.w.session.State())
		}
	})
	f.edit(t, "discard me")
	f.owner(t, func() { f.w.start(actionOpen, f.path) })
	f.acknowledge(t, nil)
	f.owner(t, func() {
		if f.w.prompt != "open" || f.w.model.Snapshot().Text != "discard me" {
			t.Fatal("dirty open bypassed decision")
		}
		f.w.dismissPrompt()
		f.w.start(actionDiscardOpen, f.path)
	})
	f.acknowledge(t, nil)
	f.finish(t, 2)
	f.owner(t, func() {
		s := f.w.model.Snapshot()
		if s.Text != text || s.History.Dirty || s.History.UndoCount != 0 || s.History.RedoCount != 0 {
			t.Fatal(s)
		}
	})
}
func TestCancelWhileSettlingDoesNotStartIO(t *testing.T) {
	writes := make(chan struct{}, 1)
	f := setup(t, memoryStore{write: func(context.Context, []byte) (bool, error) { writes <- struct{}{}; return true, nil }})
	f.edit(t, "unsaved")
	f.owner(t, func() {
		f.w.start(actionSaveClose, f.path)
		f.w.cancel()
		if f.ackContext.Err() != context.Canceled {
			t.Fatal("Cancel did not cancel native settlement context")
		}
	})
	f.acknowledge(t, nil)
	f.owner(t, func() {
		if f.w.settling || f.w.session.State().Busy || !errors.Is(f.w.last.Err, context.Canceled) || f.w.lifetime.Err() != nil {
			t.Fatal(f.w.last, f.w.session.State())
		}
	})
	select {
	case <-writes:
		t.Fatal("cancel started worker")
	default:
	}
}
func TestStaleSaveCloseKeepsNewEditsAndWindow(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var saved string
	f := setup(t, memoryStore{write: func(ctx context.Context, b []byte) (bool, error) {
		saved = string(b)
		close(entered)
		select {
		case <-release:
			return true, nil
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}})
	f.edit(t, "captured")
	f.owner(t, func() { f.w.start(actionSaveClose, f.path) })
	f.acknowledge(t, nil)
	waitEntered(t, entered)
	f.edit(t, "newer")
	close(release)
	f.finish(t, 1)
	f.owner(t, func() {
		if saved != "captured" || !errors.Is(f.w.last.Err, history.ErrStale) || !f.w.last.Written || !f.w.session.State().Dirty || f.w.model.Snapshot().Text != "newer" || f.w.lifetime.Err() != nil {
			t.Fatal(saved, f.w.last, f.w.session.State())
		}
	})
}
func TestCanceledWorkerCompletesThroughFullMailbox(t *testing.T) {
	entered := make(chan struct{})
	f := setup(t, memoryStore{write: func(ctx context.Context, _ []byte) (bool, error) {
		close(entered)
		<-ctx.Done()
		return false, ctx.Err()
	}})
	f.edit(t, "changed")
	f.owner(t, func() { f.w.start(actionSave, f.path) })
	f.acknowledge(t, nil)
	waitEntered(t, entered)
	if err := f.a.PostWindow(f.w.windowRef, func() {}); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() { f.w.cancel() })
	f.finish(t, 1)
	f.owner(t, func() {
		if f.w.session.State().Busy || !errors.Is(f.w.last.Err, context.Canceled) || !f.w.session.State().Dirty {
			t.Fatal(f.w.last, f.w.session.State())
		}
	})
}
func TestOpenRefusalAndStaleReadAreAtomic(t *testing.T) {
	for _, stale := range []bool{false, true} {
		t.Run(map[bool]string{false: "invalid UTF8", true: "stale owner edit"}[stale], func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			f := setup(t, memoryStore{read: func(ctx context.Context) ([]byte, error) {
				close(entered)
				select {
				case <-release:
					if stale {
						return []byte("read result"), nil
					}
					return []byte{255}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}})
			f.edit(t, "keep")
			f.owner(t, func() { f.w.start(actionDiscardOpen, f.path) })
			f.acknowledge(t, nil)
			waitEntered(t, entered)
			want := "keep"
			if stale {
				want = "new owner text"
				f.edit(t, want)
			}
			var before editor.Snapshot
			f.owner(t, func() { before = f.w.model.Snapshot() })
			close(release)
			f.finish(t, 1)
			f.owner(t, func() {
				after := f.w.model.Snapshot()
				if after != before || after.Text != want || f.w.session.State().Path != "" || f.w.last.Err == nil {
					t.Fatal(before, after, f.w.last)
				}
			})
		})
	}
}
func TestCloseWaitsAcknowledgmentAndRequiresExplicitDiscard(t *testing.T) {
	f := setup(t, nil)
	f.edit(t, "dirty")
	f.owner(t, func() { f.w.start(actionClose, "") })
	if f.w.lifetime.Err() != nil {
		t.Fatal("closed before settle")
	}
	f.acknowledge(t, nil)
	f.owner(t, func() {
		if f.w.prompt != "close" || f.w.lifetime.Err() != nil {
			t.Fatal("dirty close accepted")
		}
		f.w.dismissPrompt()
		f.w.start(actionDiscardClose, "")
	})
	if f.w.lifetime.Err() != nil {
		t.Fatal("discard closed before settle")
	}
	f.acknowledge(t, nil)
	if f.w.lifetime.Err() == nil || len(f.a.Windows()) != 0 {
		t.Fatal("explicit discard did not close")
	}
}
func TestAdmissionFailureAndCompositionCaptureDoNotLoseState(t *testing.T) {
	f := setup(t, nil)
	f.edit(t, "base")
	f.owner(t, func() {
		f.w.settle = func(context.Context, *app.Application, string, *editor.Model, func(error)) error {
			return app.ErrQueueFull
		}
		f.w.start(actionSave, f.path)
		if f.w.settling || !errors.Is(f.w.last.Err, app.ErrQueueFull) {
			t.Fatal(f.w.last)
		}
		s := f.w.model.Snapshot()
		if err := f.w.model.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: editor.Range{Start: 4, End: 4}, Insert: "文", Selection: editor.Selection{Anchor: 5, Focus: 5}, Marked: editor.Range{Start: 4, End: 5}}); err != nil {
			t.Fatal(err)
		}
		if _, err := documentCodec(f.w.model).Capture(); !errors.Is(err, editor.ErrComposition) {
			t.Fatal("preedit persisted", err)
		}
		if err := f.w.model.CommitComposition(); err != nil {
			t.Fatal(err)
		}
		text, err := documentCodec(f.w.model).Capture()
		if err != nil || text != "base文" {
			t.Fatal(text, err)
		}
	})
}
func TestIndependentWindowsAndSoleJournal(t *testing.T) {
	f := setup(t, nil)
	other, err := newEditorWindow(f.a, "two", windowOptions{settle: func(_ context.Context, _ *app.Application, _ string, _ *editor.Model, ack func(error)) error {
		ack(nil)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	f.edit(t, "one")
	f.owner(t, func() {
		if err := other.model.Replace("two", editor.Selection{}); err != nil {
			t.Fatal(err)
		}
		f.w.start(actionUndo, "")
	})
	f.acknowledge(t, nil)
	f.owner(t, func() {
		if f.w.model.Snapshot().Text != "" || other.model.Snapshot().Text != "two" || other.model.Snapshot().History.UndoCount != 1 {
			t.Fatal("history leaked between windows")
		}
		f.w.start(actionRedo, "")
	})
	f.acknowledge(t, nil)
	f.owner(t, func() {
		if f.w.model.Snapshot().Text != "one" {
			t.Fatal("redo lost")
		}
		if err := f.w.closeNow(); err != nil {
			t.Fatal(err)
		}
		if other.lifetime.Err() != nil {
			t.Fatal("other lifetime canceled")
		}
	})
}
func TestCodecRejectsBoundedInvalidInput(t *testing.T) {
	m, err := editor.New("", editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	codec := documentCodec(m)
	if _, err = codec.Decode([]byte{0}); !errors.Is(err, editor.ErrInvalidText) {
		t.Fatal(err)
	}
	if _, err = codec.Decode([]byte{255}); !errors.Is(err, editor.ErrInvalidText) {
		t.Fatal(err)
	}
	if _, err = codec.Decode([]byte(strings.Repeat("a", editor.DefaultMaxBytes+1))); !errors.Is(err, document.ErrTooLarge) {
		t.Fatal(err)
	}
	if err = codec.Install("not inside journal replace"); !errors.Is(err, editor.ErrInstallContext) {
		t.Fatal(err)
	}
}

func TestSaveCloseChooserCancellationKeepsDocument(t *testing.T) {
	f := setup(t, nil)
	f.owner(t, func() {
		var err error
		f.w.session, err = document.New(document.Config[string]{Owner: f.a, Window: f.w.id, Journal: f.w.model.Journal(), Codec: documentCodec(f.w.model), Choose: func(context.Context, app.FileOptions) (string, error) { return "", app.ErrCanceled }, OnComplete: f.w.finish})
		if err != nil {
			t.Fatal(err)
		}
	})
	f.edit(t, "never discard me")
	f.owner(t, func() { f.w.start(actionSaveClose, "") })
	f.acknowledge(t, nil)
	f.finish(t, 1)
	f.owner(t, func() {
		if !errors.Is(f.w.last.Err, app.ErrCanceled) || f.w.closeAfterSave || f.w.session.State().Busy || !f.w.session.State().Dirty || f.w.lifetime.Err() != nil || f.w.model.Snapshot().Text != "never discard me" {
			t.Fatal(f.w.last, f.w.session.State())
		}
	})
}

func TestWindowLifetimeCancelsIOAndSuppressesLateCompletion(t *testing.T) {
	entered, canceled := make(chan struct{}), make(chan struct{})
	f := setup(t, memoryStore{write: func(ctx context.Context, _ []byte) (bool, error) {
		close(entered)
		<-ctx.Done()
		close(canceled)
		return false, ctx.Err()
	}})
	other, err := newEditorWindow(f.a, "survivor", windowOptions{settle: func(_ context.Context, _ *app.Application, _ string, _ *editor.Model, ack func(error)) error {
		ack(nil)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	f.edit(t, "captured")
	f.owner(t, func() { f.w.start(actionSave, f.path) })
	f.acknowledge(t, nil)
	waitEntered(t, entered)
	f.owner(t, func() {
		if err := f.w.closeNow(); err != nil {
			t.Fatal(err)
		}
	})
	waitEntered(t, canceled)
	if err := f.a.Drain(); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if f.w.resultCount != 0 || other.lifetime.Err() != nil || f.w.lifetime.Err() == nil {
			t.Fatal("late result or unrelated cancellation")
		}
	})
}

func TestSaveCloseFinalBarrierKeepsPreeditStartedDuringWrite(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	f := setup(t, memoryStore{write: func(ctx context.Context, _ []byte) (bool, error) {
		close(entered)
		select {
		case <-release:
			return true, nil
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}})
	f.edit(t, "saved")
	f.owner(t, func() { f.w.start(actionSaveClose, f.path) })
	f.acknowledge(t, nil)
	waitEntered(t, entered)
	f.owner(t, func() {
		s := f.w.model.Snapshot()
		if err := f.w.model.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: editor.Range{Start: 5, End: 5}, Insert: "文", Selection: editor.Selection{Anchor: 6, Focus: 6}, Marked: editor.Range{Start: 5, End: 6}}); err != nil {
			t.Fatal(err)
		}
	})
	close(release)
	f.finish(t, 1)
	f.owner(t, func() {
		if !f.w.settling || f.ack == nil || f.w.lifetime.Err() != nil || f.w.session.State().Dirty || !f.w.model.Snapshot().Composing {
			t.Fatal("successful earlier Save closed unacknowledged preedit", f.w.model.Snapshot())
		}
		if err := f.w.model.CommitComposition(); err != nil {
			t.Fatal(err)
		}
	})
	f.acknowledge(t, nil)
	f.owner(t, func() {
		s := f.w.model.Snapshot()
		if s.Text != "saved文" || !s.History.Dirty || f.w.lifetime.Err() != nil || f.w.settling {
			t.Fatal("newer committed preedit was discarded", s)
		}
	})
}

func TestSaveCloseFinalBarrierKeepsPendingNativeOwnerUpdate(t *testing.T) {
	f := setup(t, document.FileStore{})
	f.edit(t, "saved")
	f.owner(t, func() { f.w.start(actionSaveClose, f.path) })
	f.acknowledge(t, nil)
	f.finish(t, 1)
	f.owner(t, func() {
		if !f.w.settling || f.ack == nil || f.w.lifetime.Err() != nil {
			t.Fatal("no final native FIFO acknowledgment")
		}
	})
	// Simulate an already admitted native update being applied ahead of its FIFO
	// acknowledgment; the controller must inspect the revision after that barrier.
	f.edit(t, "queued native update")
	f.acknowledge(t, nil)
	f.owner(t, func() {
		if f.w.lifetime.Err() != nil || !f.w.session.State().Dirty || f.w.model.Snapshot().Text != "queued native update" {
			t.Fatal("pending native update discarded")
		}
	})
}

func TestSaveCloseFinalBarrierAllowsUnchangedDocument(t *testing.T) {
	f := setup(t, document.FileStore{})
	f.edit(t, "saved")
	f.owner(t, func() { f.w.start(actionSaveClose, f.path) })
	f.acknowledge(t, nil)
	f.finish(t, 1)
	f.owner(t, func() {
		if !f.w.settling || f.w.lifetime.Err() != nil {
			t.Fatal("closed before final native barrier")
		}
	})
	f.acknowledge(t, nil)
	if f.w.lifetime.Err() == nil {
		t.Fatal("unchanged saved document stayed open")
	}
}

func TestControlledFindFieldHasIndependentOwnerState(t *testing.T) {
	f := setup(t, document.FileStore{})
	f.edit(t, "日本 and 日本\r\n\r\n")
	f.owner(t, func() {
		before := f.w.model.Snapshot()
		f.w.searchChanged(ui.ChangeEvent{Text: "日本"})
		if f.w.search != "日本" || f.w.searchChanges != 1 || f.w.status != "Find: 2 matches" {
			t.Fatal("find field did not accept controlled Change")
		}
		if f.w.model.Snapshot() != before {
			t.Fatal("find field edited document/history")
		}
		f.w.focusID = "find"
		f.w.dismissPrompt()
		if f.w.focusID != f.w.editorID {
			t.Fatal("dialog dismissal failed to restore editor focus")
		}
	})
}

func TestRetiredEditorCloseKeepsSameNameReplacementOpen(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "survivor", Title: "Survivor", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	old, err := newEditorWindow(a, "document", windowOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = old.closeNow(); err != nil {
		t.Fatal(err)
	}
	current, err := newEditorWindow(a, "document", windowOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = old.closeNow(); !errors.Is(err, app.ErrWindowClosed) || current.lifetime.Err() != nil {
		t.Fatal("retired editor closed replacement", err)
	}
	if current.windowRef == old.windowRef || current.model == old.model || old.lifetime.Err() == nil {
		t.Fatal("editor opening or model identity reused")
	}
}
