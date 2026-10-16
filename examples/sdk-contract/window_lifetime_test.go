package sdkcontract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
)

// This consumer retains a document's close command and an accepted worker while
// another controller opens a new document using the same logical window name.
func TestRetiredDocumentControllerCannotAffectReplacement(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 4, CompletionCapacity: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	spec := func(id string) app.Window {
		return app.Window{ID: id, Title: id, Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}
	}
	if err = a.Open(spec("survivor")); err != nil {
		t.Fatal(err)
	}
	original, err := a.OpenWindow(spec("document"))
	if err != nil {
		t.Fatal(err)
	}
	ex, err := execution.NewManual(2)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.Close()
	manager, err := tasks.NewWithConfig(tasks.Config{Owner: a, Window: original.ID(), Limit: 1, Executor: ex})
	if err != nil {
		t.Fatal(err)
	}
	commands := command.New()
	if err = commands.Register(command.Command{ID: "close", Label: "Close document", Execute: func(context.Context) error { return a.CloseWindow(original) }}); err != nil {
		t.Fatal(err)
	}
	commits, cleanups := 0, 0
	var originalTask *tasks.Task
	var startErr error
	if err = a.Do(func() {
		originalTask, startErr = tasks.Start(manager, original.Context(), func(ctx context.Context, _ func(float64, string)) (int, error) {
			defer func() { cleanups++ }()
			return 1, ctx.Err()
		}, func(int, error) { commits++ })
	}); err != nil || startErr != nil {
		t.Fatal(err, startErr)
	}
	if err = a.PostWindow(original, func() { t.Error("old document delivery ran") }); err != nil {
		t.Fatal(err)
	}
	if err = a.CloseWindow(original); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(spec("document"))
	if err != nil {
		t.Fatal(err)
	}
	if replacement == original || replacement.Context().Err() != nil || original.Context().Err() == nil {
		t.Fatal("opening identity lost")
	}
	if err = commands.Invoke(context.Background(), "close"); !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("retired close command affected new document", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err = ex.StepNext(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-originalTask.Done():
	default:
		t.Fatal("retired worker did not finish cleanup")
	}
	if err = a.Do(func() {
		_, startErr = tasks.Start(manager, ctx, func(context.Context, func(float64, string)) (int, error) {
			t.Error("retired manager submitted work")
			return 0, nil
		}, func(int, error) { t.Error("retired manager committed") })
	}); err != nil || !errors.Is(startErr, app.ErrWindowClosed) {
		t.Fatal("manager rebound", err, startErr)
	}
	current, err := tasks.NewWithConfig(tasks.Config{Owner: a, Window: replacement.ID(), Limit: 1, Executor: ex})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Do(func() {
		_, startErr = tasks.Start(current, replacement.Context(), func(context.Context, func(float64, string)) (int, error) { return 42, nil }, func(v int, err error) {
			if v != 42 || err != nil {
				t.Error(v, err)
			}
			commits++
		})
	}); err != nil || startErr != nil {
		t.Fatal(err, startErr)
	}
	if _, err = ex.StepNext(ctx); err != nil {
		t.Fatal(err)
	}
	if err = a.Drain(); err != nil {
		t.Fatal(err)
	}
	if commits != 1 || cleanups != 1 || replacement.Context().Err() != nil || !a.Pending().Idle() || !manager.Pending().Idle() || !current.Pending().Idle() {
		t.Fatal("replacement delivery or old cleanup failed", commits, cleanups, a.Pending())
	}
}
