package document

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

func operationContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func manualSession(t *testing.T, store Store) (*fixture, *execution.Manual) {
	t.Helper()
	f := setup(t, store)
	e, err := execution.NewManual(2)
	if err != nil {
		t.Fatal(err)
	}
	f.s.config.Executor = e
	return f, e
}

func TestManualSavePublishesIntoFullOrdinaryQueueAndHasDistinctPendingStates(t *testing.T) {
	var bytes string
	f, e := manualSession(t, memoryStore{write: func(_ context.Context, b []byte) (bool, error) { bytes = string(b); return true, nil }})
	f.edit(t, 42)
	ctx := operationContext(t)
	if err := f.a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if err := f.s.Save(ctx, SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
		s := f.s.State()
		if !s.WorkerPending || !s.OwnerPending || !s.Busy {
			t.Fatalf("admission state %+v", s)
		}
	})
	if _, err := e.StepNext(ctx); err != nil {
		t.Fatalf("save blocked on full ordinary queue: %v", err)
	}
	f.owner(t, func() {
		s := f.s.State()
		if s.WorkerPending || !s.OwnerPending || !s.Busy {
			t.Fatalf("worker terminal state %+v", s)
		}
	})
	if bytes != "42" || len(f.results) != 0 {
		t.Fatalf("encode/write/owner split %q results%v", bytes, f.results)
	}
	if n, err := f.a.DrainN(1); err != nil || n != 1 {
		t.Fatalf("ordinary drain %d %v", n, err)
	}
	if len(f.results) != 0 {
		t.Fatal("save completed in ordinary input batch")
	}
	if n, err := f.a.DrainN(1); err != nil || n != 1 {
		t.Fatalf("completion drain %d %v", n, err)
	}
	f.owner(t, func() {
		s := f.s.State()
		if s.WorkerPending || s.OwnerPending || s.Busy || s.Dirty || len(f.results) != 1 {
			t.Fatalf("post-terminal lag %+v results%v", s, f.results)
		}
	})
}

type refusalExecutor struct{ err error }

func (r refusalExecutor) Submit(func()) error { return r.err }
func TestDocumentExecutorRefusalPreservesStateAndRefundsReservation(t *testing.T) {
	var calls atomic.Int32
	f := setup(t, memoryStore{write: func(context.Context, []byte) (bool, error) { calls.Add(1); return false, nil }, read: func(context.Context) ([]byte, error) { calls.Add(1); return nil, nil }})
	f.edit(t, 7)
	refusal := errors.New("executor refused")
	f.s.config.Executor = refusalExecutor{refusal}
	f.owner(t, func() {
		before := f.s.State()
		journal := f.j.Snapshot()
		if err := f.s.Save(context.Background(), SaveOptions{Path: f.path}); !errors.Is(err, refusal) {
			t.Fatalf("save refusal %v", err)
		}
		if err := f.s.Open(context.Background(), OpenOptions{Path: f.path, DiscardChanges: true}); !errors.Is(err, refusal) {
			t.Fatalf("open refusal %v", err)
		}
		if f.s.State() != before || f.j.Snapshot() != journal || calls.Load() != 0 || !f.a.Pending().Idle() {
			t.Fatalf("refusal changed state %+v before%+v calls%d queue%+v", f.s.State(), before, calls.Load(), f.a.Pending())
		}
	})
}

func TestDocumentCancelAndGoexitStillPublishOneTerminalResult(t *testing.T) {
	for _, mode := range []string{"cancel", "goexit"} {
		t.Run(mode, func(t *testing.T) {
			var cleanup atomic.Int32
			f, e := manualSession(t, memoryStore{write: func(ctx context.Context, _ []byte) (bool, error) {
				defer cleanup.Add(1)
				if mode == "goexit" {
					runtime.Goexit()
				}
				return false, ctx.Err()
			}})
			ctx := operationContext(t)
			f.owner(t, func() {
				if err := f.s.Save(ctx, SaveOptions{Path: f.path}); err != nil {
					t.Fatal(err)
				}
				if mode == "cancel" {
					f.s.Cancel()
				}
			})
			_, err := e.StepNext(ctx)
			if mode == "goexit" && !errors.Is(err, execution.ErrGoexit) {
				t.Fatalf("Goexit not observable %v", err)
			}
			if mode == "cancel" && err != nil {
				t.Fatal(err)
			}
			if cleanup.Load() != 1 {
				t.Fatal("accepted canceled operation skipped cleanup")
			}
			if err := f.a.Drain(); err != nil {
				t.Fatal(err)
			}
			f.owner(t, func() {
				s := f.s.State()
				if s.WorkerPending || s.OwnerPending || s.Busy || len(f.results) != 1 || f.results[0].Err == nil {
					t.Fatalf("terminal state %+v results%v", s, f.results)
				}
			})
		})
	}
}

func TestClosedDocumentRevokesResultWhileWorkerStillRuns(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	f, e := manualSession(t, memoryStore{write: func(ctx context.Context, _ []byte) (bool, error) { close(started); <-release; return false, ctx.Err() }})
	if err := f.a.Open(app.Window{ID: "survivor", Title: "Survivor", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "survivor", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	ctx := operationContext(t)
	f.owner(t, func() {
		if err := f.s.Save(ctx, SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	done := make(chan error, 1)
	go func() { _, err := e.StepNext(ctx); done <- err }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := f.a.CloseWindow(f.s.window); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		s := f.s.State()
		if !s.Closed || s.Busy || !s.WorkerPending || s.OwnerPending || len(f.results) != 0 {
			t.Fatalf("closed running state %+v results%v", s, f.results)
		}
	})
	if e.Snapshot().Idle() {
		t.Fatal("executor hid closed-document worker")
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	f.owner(t, func() {
		s := f.s.State()
		if s.WorkerPending || s.OwnerPending || len(f.results) != 0 {
			t.Fatalf("late state/callback %+v results%v", s, f.results)
		}
	})
	if !f.a.Pending().Idle() {
		t.Fatalf("closed operation leaked reserved delivery %+v", f.a.Pending())
	}
}

func TestDocumentSessionNeverRebindsToReopenedName(t *testing.T) {
	writes := 0
	f, e := manualSession(t, memoryStore{write: func(context.Context, []byte) (bool, error) { writes++; return true, nil }})
	if err := f.a.Open(app.Window{ID: "survivor", Title: "Survivor", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "survivor", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	old := f.s.window
	if err := f.a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	current, err := f.a.OpenWindow(app.Window{ID: "main", Title: "Reopened", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "main", Kind: ui.Label} }})
	if err != nil {
		t.Fatal(err)
	}
	ctx := operationContext(t)
	f.owner(t, func() {
		if err := f.s.Save(ctx, SaveOptions{Path: f.path}); !errors.Is(err, app.ErrWindowClosed) {
			t.Fatal("old save rebound", err)
		}
		if err := f.s.Open(ctx, OpenOptions{Path: f.path}); !errors.Is(err, app.ErrWindowClosed) {
			t.Fatal("old open rebound", err)
		}
	})
	if !e.Snapshot().Idle() || writes != 0 || current.Context().Err() != nil {
		t.Fatal("closed session admitted work")
	}
	fresh, err := New(f.s.config)
	if err != nil || fresh.window != current || fresh.window == old {
		t.Fatal("fresh session captured wrong lifetime", err)
	}
	f.owner(t, func() {
		if err := fresh.Save(ctx, SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
	})
	if _, err = e.StepNext(ctx); err != nil {
		t.Fatal(err)
	}
	if err = f.a.Drain(); err != nil {
		t.Fatal(err)
	}
	if writes != 1 || fresh.State().Closed || !f.s.State().Closed || len(f.results) != 1 {
		t.Fatal("session generations crossed")
	}
}
