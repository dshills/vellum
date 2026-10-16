package document

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

type memoryStore struct {
	read  func(context.Context) ([]byte, error)
	write func(context.Context, []byte) (bool, error)
}

func (s memoryStore) Read(c context.Context, _ string, _ int64) ([]byte, error) { return s.read(c) }
func (s memoryStore) Write(c context.Context, _ string, b []byte, _ int64) (bool, error) {
	return s.write(c, b)
}

type fixture struct {
	a       *app.Application
	j       *history.Journal
	s       *Session[int]
	value   int
	results []Result
	path    string
}

func setup(t *testing.T, store Store) *fixture {
	t.Helper()
	f := &fixture{path: filepath.Join(t.TempDir(), "document.vellum")}
	f.a, _ = app.New(app.Config{QueueCapacity: 1})
	t.Cleanup(f.a.Close)
	if err := f.a.Open(app.Window{ID: "main", Title: "Document", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	f.j, _ = history.New(8)
	var err error
	f.s, err = New(Config[int]{Owner: f.a, Window: "main", Journal: f.j, Store: store, Codec: Codec[int]{Capture: func() (int, error) { return f.value, nil }, Encode: func(v int) ([]byte, error) { return []byte(strconv.Itoa(v)), nil }, Decode: func(b []byte) (int, error) { return strconv.Atoi(string(b)) }, Install: func(v int) error { f.value = v; return nil }}, OnComplete: func(r Result) { f.results = append(f.results, r) }})
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
func (f *fixture) edit(t *testing.T, n int) {
	t.Helper()
	f.owner(t, func() {
		before := f.value
		if err := f.j.Apply(history.Change{Label: "Value", Apply: func() error { f.value = n; return nil }, Revert: func() error { f.value = before; return nil }}); err != nil {
			t.Fatal(err)
		}
	})
}
func (f *fixture) finish(t *testing.T, n int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := f.a.Drain(); err != nil {
			t.Fatal(err)
		}
		done := false
		f.owner(t, func() { done = len(f.results) >= n })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("completion timed out")
}
func wait(t *testing.T, c <-chan struct{}) {
	t.Helper()
	select {
	case <-c:
	case <-time.After(3 * time.Second):
		t.Fatal("worker timed out")
	}
}
func TestSaveLoadRevisionAndAtomicInstall(t *testing.T) {
	f := setup(t, FileStore{})
	f.edit(t, 42)
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
		if !f.s.State().Busy {
			t.Fatal("not busy")
		}
		if !errors.Is(f.s.Open(context.Background(), OpenOptions{Path: f.path}), ErrBusy) {
			t.Fatal("overlapping IO")
		}
	})
	f.finish(t, 1)
	if f.results[0].Err != nil || !f.results[0].Written || f.j.Snapshot().Dirty || f.s.State().Path != f.path {
		t.Fatal(f.results, f.s.State())
	}
	saved := f.j.Snapshot().Revision
	f.edit(t, 84)
	f.owner(t, func() {
		if !errors.Is(f.s.Open(context.Background(), OpenOptions{Path: f.path}), ErrUnsaved) {
			t.Fatal("unsaved load admitted")
		}
		if err := f.s.Open(context.Background(), OpenOptions{Path: f.path, DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 2)
	if f.results[1].Err != nil || f.value != 42 || f.j.Snapshot().Dirty || f.j.Snapshot().UndoCount != 0 || f.j.Snapshot().Revision == saved {
		t.Fatal(f.results, f.value, f.j.Snapshot())
	}
}
func TestStaleSaveAndLoadPreserveCurrentModel(t *testing.T) {
	for _, action := range []Action{Save, Open} {
		t.Run(string(action), func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			var written string
			store := memoryStore{write: func(c context.Context, b []byte) (bool, error) {
				written = string(b)
				close(entered)
				<-release
				return true, nil
			}, read: func(c context.Context) ([]byte, error) { close(entered); <-release; return []byte("90"), nil }}
			f := setup(t, store)
			f.edit(t, 42)
			f.owner(t, func() {
				var err error
				if action == Save {
					err = f.s.Save(context.Background(), SaveOptions{Path: f.path})
				} else {
					err = f.s.Open(context.Background(), OpenOptions{Path: f.path, DiscardChanges: true})
				}
				if err != nil {
					t.Fatal(err)
				}
			})
			wait(t, entered)
			f.edit(t, 84)
			state := f.j.Snapshot()
			close(release)
			f.finish(t, 1)
			if !errors.Is(f.results[0].Err, history.ErrStale) || f.value != 84 || f.j.Snapshot() != state || f.s.State().Path != "" {
				t.Fatal(f.results, f.value, f.j.Snapshot())
			}
			if action == Save && (written != "42" || !f.results[0].Written) {
				t.Fatal("snapshot not persisted", written)
			}
		})
	}
}
func TestCanceledCompletionClearsBusyEvenWithFullMailbox(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	f := setup(t, memoryStore{write: func(c context.Context, b []byte) (bool, error) { close(entered); <-release; return true, nil }})
	f.edit(t, 7)
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	wait(t, entered)
	if err := f.a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() { f.s.Cancel() })
	close(release)
	f.finish(t, 1)
	if !errors.Is(f.results[0].Err, context.Canceled) || f.s.State().Busy || !f.s.State().Dirty || f.s.State().Path != "" {
		t.Fatal(f.results, f.s.State())
	}
}
func TestWindowCloseDropsLateCallbacksAndCancelsWorker(t *testing.T) {
	entered, ended := make(chan struct{}), make(chan struct{})
	f := setup(t, memoryStore{read: func(c context.Context) ([]byte, error) {
		close(entered)
		<-c.Done()
		close(ended)
		return []byte("90"), nil
	}})
	// Keep another window alive so the application queue can still be drained.
	if err := f.a.Open(app.Window{ID: "other", Title: "Other", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "other", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if err := f.s.Open(context.Background(), OpenOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	wait(t, entered)
	if err := f.a.CloseWindow(f.s.window); err != nil {
		t.Fatal(err)
	}
	wait(t, ended)
	f.a.Drain()
	f.owner(t, func() {
		if !f.s.State().Closed || f.s.State().Busy || f.value != 0 || len(f.results) != 0 {
			t.Fatal("late result installed")
		}
	})
}
func TestWorkerPanicAndGoexitClearBusy(t *testing.T) {
	for _, exit := range []bool{false, true} {
		t.Run(fmt.Sprint(exit), func(t *testing.T) {
			f := setup(t, memoryStore{read: func(c context.Context) ([]byte, error) {
				if exit {
					runtime.Goexit()
				}
				panic("decode worker")
			}})
			f.owner(t, func() {
				if err := f.s.Open(context.Background(), OpenOptions{Path: f.path}); err != nil {
					t.Fatal(err)
				}
			})
			f.finish(t, 1)
			if f.results[0].Err == nil || f.s.State().Busy || f.value != 0 {
				t.Fatal(f.results, f.s.State())
			}
		})
	}
}
func TestFailedInstallAndWrittenErrorDoNotAcknowledge(t *testing.T) {
	failure := errors.New("failed")
	f := setup(t, memoryStore{read: func(context.Context) ([]byte, error) { return []byte("90"), nil }, write: func(context.Context, []byte) (bool, error) { return true, failure }})
	f.edit(t, 7)
	before := f.j.Snapshot()
	f.s.config.Codec.Install = func(int) error { return failure }
	f.owner(t, func() {
		if err := f.s.Open(context.Background(), OpenOptions{Path: f.path, DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 1)
	if !errors.Is(f.results[0].Err, failure) || f.value != 7 || f.j.Snapshot() != before {
		t.Fatal("failed install altered model/history")
	}
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 2)
	if !f.results[1].Written || !errors.Is(f.results[1].Err, failure) || f.s.State().Path != "" || f.j.Snapshot() != before {
		t.Fatal("partial save acknowledged")
	}
}
func TestCompletionCanChainSaveAndClose(t *testing.T) {
	f := setup(t, FileStore{})
	f.edit(t, 1)
	f.s.config.OnComplete = func(r Result) {
		f.results = append(f.results, r)
		if r.Err != nil {
			t.Error(r.Err)
			return
		}
		if len(f.results) == 1 {
			if err := f.s.Open(context.Background(), OpenOptions{Path: f.path}); err != nil {
				t.Error(err)
			}
		} else {
			if err := f.a.CloseWindow(f.s.window); err != nil {
				t.Error(err)
			}
		}
	}
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	deadline := time.Now().Add(3 * time.Second)
	for f.a.Context().Err() == nil && time.Now().Before(deadline) {
		f.a.Drain()
		time.Sleep(time.Millisecond)
	}
	if f.a.Context().Err() == nil || len(f.results) != 2 || f.results[1].Err != nil {
		t.Fatal(f.results)
	}
}

func TestCancelAfterAdmissionAndRetryOnLiveWindow(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	calls := 0
	f := setup(t, memoryStore{write: func(c context.Context, b []byte) (bool, error) {
		calls++
		if calls == 1 {
			close(entered)
			<-release
		}
		return true, nil
	}})
	f.edit(t, 7)
	wake := make(chan struct{}, 4)
	f.a.SetWake(func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	})
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	wait(t, entered)
	for len(wake) > 0 {
		<-wake
	}
	close(release)
	wait(t, wake) // PostWait admission wake; completion has not executed.
	f.owner(t, func() {
		if !f.s.State().Busy || len(f.results) != 0 {
			t.Fatal("admission counted as execution")
		}
		f.s.Cancel()
	})
	f.finish(t, 1)
	if !errors.Is(f.results[0].Err, context.Canceled) || !f.j.Snapshot().Dirty || f.s.State().Path != "" {
		t.Fatal(f.results, f.s.State())
	}
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 2)
	if f.results[1].Err != nil || f.j.Snapshot().Dirty || f.s.State().Path != f.path || calls != 2 {
		t.Fatal("retry failed", f.results)
	}
}
func TestOwnerInstallPanicInvalidatesAndDropsCompletion(t *testing.T) {
	f := setup(t, memoryStore{read: func(context.Context) ([]byte, error) { return []byte("7"), nil }})
	f.s.config.Codec.Install = func(int) error { panic("bad install") }
	admitted := make(chan struct{}, 4)
	f.a.SetWake(func() {
		select {
		case admitted <- struct{}{}:
		default:
		}
	})
	f.owner(t, func() {
		if err := f.s.Open(context.Background(), OpenOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	deadline := time.Now().Add(3 * time.Second)
	panicked := false
	for !panicked && time.Now().Before(deadline) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if r != "bad install" {
						t.Errorf("panic identity: %v", r)
					}
					panicked = true
				}
			}()
			f.a.Drain()
		}()
		if !panicked {
			time.Sleep(time.Millisecond)
		}
	}
	if !panicked || !f.j.Snapshot().Invalid || !f.j.Snapshot().Dirty || f.a.Context().Err() == nil || len(f.results) != 0 {
		t.Fatal("invalid install did not stop owner", f.j.Snapshot(), f.results)
	}
}
func TestChooserFailureAndByteBoundsPreserveDocument(t *testing.T) {
	f := setup(t, memoryStore{read: func(context.Context) ([]byte, error) { return []byte("123456789"), nil }, write: func(context.Context, []byte) (bool, error) {
		t.Error("oversize write reached store")
		return false, nil
	}})
	f.edit(t, 1234)
	f.s.config.MaxBytes = 3
	f.s.config.Choose = func(context.Context, app.FileOptions) (string, error) { return "", app.ErrCanceled }
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Choose: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 1)
	if !errors.Is(f.results[0].Err, app.ErrCanceled) {
		t.Fatal(f.results)
	}
	f.owner(t, func() {
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 2)
	if !errors.Is(f.results[1].Err, ErrTooLarge) {
		t.Fatal(f.results)
	}
	f.owner(t, func() {
		if err := f.s.Open(context.Background(), OpenOptions{Path: f.path, DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.finish(t, 3)
	if !errors.Is(f.results[2].Err, ErrTooLarge) || f.value != 1234 || !f.j.Snapshot().Dirty || f.s.State().Path != "" {
		t.Fatal(f.results, f.value, f.s.State())
	}
}
