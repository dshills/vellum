package notes

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"testing"
	"time"
)

func quitFiles(t *testing.T) (*app.Application, []*Files, *QuitGroup) {
	t.Helper()
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	var files []*Files
	for _, id := range []string{"A", "B"} {
		r, e := OpenRepository(privateDirectory(t))
		if e != nil {
			t.Fatal(e)
		}
		m := New()
		ref, e := a.OpenWindow(app.Window{ID: id, Title: id, Width: 600, Height: 400, View: m.View})
		if e != nil {
			t.Fatal(e)
		}
		f, e := AttachFiles(a, ref, m, r, nil)
		if e != nil {
			t.Fatal(e)
		}
		files = append(files, f)
		t.Cleanup(func() {
			a.Close()
			if e := f.Stop(); e != nil {
				t.Error(e)
			}
			_ = r.Close()
		})
	}
	group, e := NewQuitGroup(a, files...)
	if e != nil {
		t.Fatal(e)
	}
	return a, files, group
}
func quitDrain(t *testing.T, a *app.Application, g *QuitGroup) {
	t.Helper()
	until := time.Now().Add(3 * time.Second)
	for time.Now().Before(until) {
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
		running := false
		_ = a.Do(func() { running = g.running })
		if !running {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("quit did not finish")
}
func TestQuitCancelKeepsCleanAndDirtyWindows(t *testing.T) {
	a, f, g := quitFiles(t)
	_ = a.Do(func() {
		_, _ = f[1].model.Add()
		g.Request()
		g.Request()
		if len(a.Windows()) != 2 || f[1].prompt != "quit" {
			t.Fatal("closed before decision")
		}
		if f[0].model.Commands().Enabled("notes.new") == nil {
			t.Fatal("clean window not held")
		}
		f[1].promptView().Dismiss()
		if g.running || len(a.Windows()) != 2 || f[0].quitHeld || f[1].quitHeld || len(f[1].model.Items()) != 7 {
			t.Fatal("Cancel changed workspace")
		}
	})
}
func TestQuitApprovalClearsAllThenCloses(t *testing.T) {
	a, f, g := quitFiles(t)
	_ = a.Do(func() {
		_, _ = f[0].model.Add()
		_, _ = f[1].model.Add()
		g.Request()
		f[0].settleIntent("quit", false)
		if len(a.Windows()) != 2 || f[1].prompt != "quit" {
			t.Fatal("first approval closed a window")
		}
		f[1].settleIntent("quit", true)
	})
	quitDrain(t, a, g)
	if len(a.Windows()) != 0 {
		t.Fatal("approved Quit left windows")
	}
}
func TestQuitStaleApprovalAndRecoveryFailureKeepWindows(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "stale", true: "recovery failure"}[failure], func(t *testing.T) {
			a, f, g := quitFiles(t)
			_ = a.Do(func() {
				_, _ = f[1].model.Add()
				g.Request()
				if failure {
					j, _ := history.New(1)
					session, e := document.New(document.Config[Recovery]{Owner: a, Window: f[0].ref.ID(), Journal: j, Path: "/irrelevant", MaxBytes: MaxNotebookBytes, Store: failureStore{write: func(context.Context, string, []byte, int64) (bool, error) { return false, errors.New("write refused") }}, Codec: document.Codec[Recovery]{Capture: func() (Recovery, error) { f[0].capturedClear = true; return Recovery{Version: 1}, nil }, Encode: encodeRecovery, Decode: decodeRecovery, Install: func(Recovery) error { return nil }}, OnComplete: f[0].recovered})
					if e != nil {
						t.Fatal(e)
					}
					f[0].recovery = session
				}
				if !failure {
					if err := f[0].withQuitAccess(func() error { _, e := f[0].model.Add(); return e }); err != nil {
						t.Fatal(err)
					}
				}
				f[1].settleIntent("quit", false)
			})
			quitDrain(t, a, g)
			if len(a.Windows()) != 2 {
				t.Fatal("failure partially closed workspace")
			}
		})
	}
}
func TestQuitSettlementFailureAndClosedRef(t *testing.T) {
	a, f, g := quitFiles(t)
	_ = a.Do(func() {
		_ = f[0].model.SetSettlement(func(done func(error)) error { done(errors.New("composition refused")); return nil })
		g.Request()
		if g.running || len(a.Windows()) != 2 {
			t.Fatal("settlement failure closed windows")
		}
		_ = f[0].model.SetSettlement(nil)
		_, _ = f[0].model.Add()
		g.Request()
		_ = a.CloseWindow(f[1].ref)
		f[0].promptView().Dismiss()
		if g.running || f[0].quitHeld {
			t.Fatal("retired ref retained Quit")
		}
	})

}

func TestQuitAsynchronousSettlementRefusalUnlocksWorkspace(t *testing.T) {
	a, f, g := quitFiles(t)
	var finish func(error)
	_ = a.Do(func() {
		_ = f[0].model.SetSettlement(func(done func(error)) error { finish = done; return nil })
		g.Request()
		if !g.running || !f[0].model.pending {
			t.Fatal("settlement not retained")
		}
		finish(errors.New("native composition refused"))
		if g.running || f[0].quitHeld || f[1].quitHeld || len(a.Windows()) != 2 {
			t.Fatal("late refusal retained or closed workspace")
		}
	})
}
func TestQuitSavedFirstThenCancelLastKeepsBoth(t *testing.T) {
	a, f, g := quitFiles(t)
	_ = a.Do(func() { _, _ = f[0].model.Add(); _, _ = f[1].model.Add(); g.Request(); f[0].settleIntent("quit", true) })
	until := time.Now().Add(3 * time.Second)
	for time.Now().Before(until) {
		_ = a.Drain()
		ready := false
		_ = a.Do(func() { ready = f[1].prompt == "quit" })
		if ready {
			break
		}
		time.Sleep(time.Millisecond)
	}
	_ = a.Do(func() {
		if f[1].prompt != "quit" {
			t.Fatal("second decision absent")
		}
		f[1].promptView().Dismiss()
		if g.running || len(a.Windows()) != 2 || f[0].model.History().Dirty || !f[1].model.History().Dirty {
			t.Fatal("Cancel changed saved/dirty workspace")
		}
	})
}

func TestQuitRequestAfterUnexpectedRetirementUnlocksSurvivor(t *testing.T) {
	a, f, g := quitFiles(t)
	_ = a.Do(func() {
		_, _ = f[0].model.Add()
		g.Request()
		_ = a.CloseWindow(f[1].ref)
		g.Request()
		if g.running || f[0].quitHeld {
			t.Fatal("retired window stranded Quit")
		}
	})
}
