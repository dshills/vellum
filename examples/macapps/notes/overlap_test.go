package notes

import (
	"context"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
)

func TestOperationOverlapSaveRefusesOpenQuitAndClose(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	target := writeNotebookFixture(t, 8)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	group, err := NewQuitGroup(x.a, x.f)
	if err != nil {
		t.Fatal(err)
	}
	var expected []Note
	x.owner(t, func() {
		if _, err := x.m.Add(); err != nil {
			t.Fatal(err)
		}
		expected = x.m.Items()
		replaceMainStore(t, x, failureStore{func(ctx context.Context, path string, data []byte, limit int64) (bool, error) {
			close(entered)
			select {
			case <-release:
				return (document.FileStore{}).Write(ctx, path, data, limit)
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}})
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("save worker did not start")
	}
	x.owner(t, func() {
		if err := x.f.OpenPath(target, false); !errors.Is(err, ErrPending) {
			t.Fatal("Open bypassed in-flight Save", err)
		}
		group.Request()
		if x.f.RequestClose() || group.running || x.f.quitHeld || x.f.prompt != "" || x.f.pendingExternal != nil || len(x.a.Windows()) != 1 {
			t.Fatal("in-flight Save admitted destructive action")
		}
		if !reflect.DeepEqual(x.m.Items(), expected) || !x.m.History().Dirty {
			t.Fatal("overlap changed saved snapshot")
		}
	})
	unblock()
	x.finish(t)
	data, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeNotebook(data)
	if err != nil || !reflect.DeepEqual(saved.Notes, expected) {
		t.Fatal("Save lost original snapshot", err)
	}
	x.owner(t, func() {
		if x.m.History().Dirty || x.f.main.State().Path != x.r.Path || x.f.openDestination != "" {
			t.Fatal("rejected actions replayed after Save")
		}
		if _, err := x.m.Add(); err != nil {
			t.Fatal(err)
		}
		group.Request()
		if x.f.prompt != "quit" {
			t.Fatal("Quit retry did not ask for dirty decision")
		}
		x.f.promptView().Dismiss()
		if group.running || x.f.quitHeld || len(x.a.Windows()) != 1 || !x.m.History().Dirty {
			t.Fatal("Quit Cancel lost live edits")
		}
	})
}

func TestOperationOverlapQueuedOpenPreservesDirtyDecisionAndRejectsDuplicates(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	if err := x.f.ConfigureRestoration(x.r); err != nil {
		t.Fatal(err)
	}
	target := writeNotebookFixture(t, 8)
	other := writeNotebookFixture(t, 9)
	group, err := NewQuitGroup(x.a, x.f)
	if err != nil {
		t.Fatal(err)
	}
	var before []Note
	x.owner(t, func() {
		if _, err := x.m.Add(); err != nil {
			t.Fatal(err)
		}
		before = x.m.Items()
		x.f.Activated()
		if !x.f.workspace.State().Busy {
			t.Fatal("no metadata operation")
		}
		if err := x.f.OpenPath(target, false); err != nil {
			t.Fatal(err)
		}
		if err := x.f.OpenPath(other, false); !errors.Is(err, document.ErrBusy) {
			t.Fatal("second queued Open admitted", err)
		}
		group.Request()
		if x.f.RequestClose() || group.running || x.f.quitHeld || x.f.pendingExternal == nil {
			t.Fatal("queued Open lost to Quit/close")
		}
	}) // Completion cannot run until the serialized owner drains.
	x.finish(t)
	x.owner(t, func() {
		if x.f.pendingExternal != nil || x.f.prompt != "open" || x.f.openDestination != target || !reflect.DeepEqual(before, x.m.Items()) || !x.m.History().Dirty {
			t.Fatal("queued Open bypassed dirty decision or changed destination")
		}
		x.f.promptView().Dismiss()
		x.f.Activated() // Unrelated metadata success must not replay the canceled Open.
	})
	x.finish(t)
	x.owner(t, func() {
		if x.f.prompt != "" || x.f.openDestination != "" || !reflect.DeepEqual(before, x.m.Items()) {
			t.Fatal("canceled request replayed")
		}
		if err := x.f.OpenPath(target, false); err != nil {
			t.Fatal(err)
		}
		x.f.settleIntent("open", true)
	})
	x.finish(t)
	data, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeNotebook(data)
	if err != nil || !reflect.DeepEqual(saved.Notes, before) {
		t.Fatal("Save before Open lost previous notebook", err)
	}
	x.owner(t, func() {
		if len(x.m.Items()) != 8 || x.m.History().Dirty || x.m.History().UndoCount != 0 || x.f.main.State().Path != target || x.f.pendingExternal != nil || x.f.resumeOpen {
			t.Fatal("explicit Open retry did not finish cleanly")
		}
	})
}

func TestOperationOverlapLateSettlementAfterWindowClosure(t *testing.T) {
	for _, action := range []string{"open", "save", "close", "delete"} {
		t.Run(action, func(t *testing.T) {
			x := filesFixture(t, privateDirectory(t))
			target := writeNotebookFixture(t, 8)
			x.owner(t, func() {
				if _, err := x.m.Add(); err != nil {
					t.Fatal(err)
				}
				before, history := x.m.Items(), x.m.History()
				var finish func(error)
				if err := x.m.SetSettlement(func(done func(error)) error { finish = done; return nil }); err != nil {
					t.Fatal(err)
				}
				var err error
				switch action {
				case "open":
					err = x.f.OpenPath(target, false)
				case "save":
					err = x.m.Request(func() error { return x.f.Save(false) })
				case "close":
					if x.f.RequestClose() {
						t.Fatal("unexpected synchronous Close approval")
					}
				case "delete":
					err = x.m.Request(func() error { return x.m.Delete(x.m.Selected()) })
				}
				if err != nil {
					t.Fatal(err)
				}
				if finish == nil || !x.m.pending {
					t.Fatal("action did not await settlement")
				}
				if _, err := x.a.OpenWindow(app.Window{ID: "survivor", Title: "survivor", Width: 600, Height: 400, View: New().View}); err != nil {
					t.Fatal(err)
				}
				if err := x.a.CloseWindow(x.f.ref); err != nil {
					t.Fatal(err)
				}
				replacement := New()
				if _, err := x.a.OpenWindow(app.Window{ID: x.f.ref.ID(), Title: "replacement", Width: 600, Height: 400, View: replacement.View}); err != nil {
					t.Fatal(err)
				}
				finish(nil)
				finish(nil)
				if x.f.prompt != "" || x.f.openDestination != "" || x.m.pending || !reflect.DeepEqual(x.m.Items(), before) || x.m.History() != history || len(x.a.Windows()) != 2 || len(replacement.Items()) != 6 {
					t.Fatal("late action resurrected retired state", x.f.prompt, x.f.openDestination)
				}
				called := false
				if err := x.m.Request(func() error { called = true; return nil }); !errors.Is(err, app.ErrWindowClosed) || called {
					t.Fatal("retired model admitted a new request", err)
				}
			})
			if _, err := os.Stat(x.r.Path); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("late callback wrote a notebook", err)
			}
			if _, err := os.Stat(x.r.RecoveryPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("late callback wrote recovery", err)
			}
		})
	}
}
