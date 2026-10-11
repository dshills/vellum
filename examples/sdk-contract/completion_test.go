package sdkcontract_test

import (
	"context"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
)

func TestPublicExecutorAndReservedCompletionContract(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 1, CompletionCapacity: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	if err := a.Open(app.Window{ID: "document", Title: "SDK", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	ex, err := execution.NewManual(2)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := tasks.NewWithConfig(tasks.Config{Owner: a, Window: "document", Limit: 1, Executor: ex})
	if err != nil {
		t.Fatal(err)
	}
	journal, err := history.New(4)
	if err != nil {
		t.Fatal(err)
	}
	_, err = document.New(document.Config[string]{Owner: a, Window: "document", Journal: journal, Executor: ex, Codec: document.Codec[string]{Capture: func() (string, error) { return "", nil }, Encode: func(s string) ([]byte, error) { return []byte(s), nil }, Decode: func(b []byte) (string, error) { return string(b), nil }, Install: func(string) error { return nil }}})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	var task *tasks.Task
	var committed bool
	if err := a.Do(func() {
		task, err = tasks.Start(manager, t.Context(), func(context.Context, func(float64, string)) (int, error) { return 7, nil }, func(value int, failure error) { committed = value == 7 && failure == nil })
	}); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ex.StepNext(t.Context()); err != nil {
		t.Fatal(err)
	}
	<-task.Done()
	if state := task.Snapshot(); !state.WorkerDone || state.OwnerDone || !state.AwaitingOwner {
		t.Fatalf("worker/admission state %+v", state)
	}
	if pending := a.Pending(); pending.Ordinary != 1 || pending.ReadyCompletions != 1 {
		t.Fatalf("saturated mailbox %+v", pending)
	}
	if n, err := a.DrainN(7); err != nil || n != 2 {
		t.Fatalf("drain %d %v", n, err)
	}
	<-task.OwnerDone()
	if !committed || !manager.Pending().Idle() || !a.Pending().Idle() || !ex.Snapshot().Idle() {
		t.Fatalf("terminal contract committed%v task%+v app%+v executor%+v", committed, task.Snapshot(), a.Pending(), ex.Snapshot())
	}
}
