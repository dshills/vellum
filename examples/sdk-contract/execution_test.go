package sdkcontract_test

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"reflect"
	"testing"
	"time"
)

func TestPublicManualTaskReservedCompletionWithoutOwnerPump(t *testing.T) {
	a, e := app.New(app.Config{QueueCapacity: 1, CompletionCapacity: 2})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	manual, e := execution.NewManual(2)
	if e != nil {
		t.Fatal(e)
	}
	defer manual.Close()
	manager, e := tasks.NewWithConfig(tasks.Config{Owner: a, Limit: 2, Executor: manual})
	if e != nil {
		t.Fatal(e)
	}
	var got []string
	if e = a.Post(func() { got = append(got, "P") }); e != nil {
		t.Fatal(e)
	}
	var task *tasks.Task
	var startErr error
	if e = a.Do(func() {
		task, startErr = tasks.Start(manager, context.Background(), func(context.Context, func(float64, string)) (int, error) { return 42, nil }, func(value int, err error) {
			if err != nil || value != 42 {
				t.Errorf("result %d %v", value, err)
			}
			got = append(got, "result")
		})
	}); e != nil {
		t.Fatal(e)
	}
	if startErr != nil {
		t.Fatal(startErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, e = manual.StepNext(ctx); e != nil {
		t.Fatal(e)
	}
	select {
	case <-task.Done():
	default:
		t.Fatal("worker did not terminate")
	}
	if !task.Snapshot().AwaitingOwner || manager.Pending().Active != 1 || a.Pending().ReadyCompletions != 1 {
		t.Fatal("owner-pending result was reported terminal")
	}
	if count, e := a.DrainN(1); e != nil || count != 1 {
		t.Fatal(count, e)
	}
	if !reflect.DeepEqual(got, []string{"P"}) {
		t.Fatal(got)
	}
	if count, e := a.DrainN(7); e != nil || count != 1 {
		t.Fatal(count, e)
	}
	select {
	case <-task.OwnerDone():
	default:
		t.Fatal("owner delivery did not terminate")
	}
	if !reflect.DeepEqual(got, []string{"P", "result"}) || !manager.Pending().Idle() || !a.Pending().Idle() {
		t.Fatal(got, manager.Pending(), a.Pending())
	}
}
func TestPublicManualClockExplicitAdvanceAndFutureDeadline(t *testing.T) {
	epoch := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	c, e := execution.NewManualClock(epoch, 3, 3)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	var got []string
	_, e = c.AfterFunc(context.Background(), time.Second, func() {
		if c.Now() != epoch.Add(time.Second) {
			t.Error("callback did not observe deadline")
		}
		got = append(got, "A")
	})
	if e != nil {
		t.Fatal(e)
	}
	timer, e := c.AfterFunc(context.Background(), time.Second, func() { got = append(got, "B") })
	if e != nil {
		t.Fatal(e)
	}
	ready, e := c.ReadySnapshot()
	if e != nil || !ready.CurrentIdle() || ready.Idle() || ready.Future != 2 || ready.Due != 0 || !ready.HasDeadline || ready.NextDeadline != epoch.Add(time.Second) {
		t.Fatal("future timer inspection", ready, e)
	}
	if len(got) != 0 || c.Now() != epoch {
		t.Fatal("admission advanced time")
	}
	if e = timer.Reset(-1); !errors.Is(e, execution.ErrDuration) {
		t.Fatal(e)
	}
	if result, e := c.AdvanceContext(context.Background(), time.Second); e != nil || result.Fired != 2 {
		t.Fatal(result, e)
	}
	if !reflect.DeepEqual(got, []string{"A", "B"}) || !c.Snapshot().Idle() {
		t.Fatal(got, c.Snapshot())
	}
}
func TestPublicDocumentManualWorkerAndOwnerCompletion(t *testing.T) {
	a, e := app.New(app.Config{QueueCapacity: 1, CompletionCapacity: 1})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	if e = a.Open(app.Window{ID: "main", Title: "manual", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); e != nil {
		t.Fatal(e)
	}
	manual, _ := execution.NewManual(1)
	defer manual.Close()
	model, _ := editor.New("portable\r\n🙂\r\n", editor.Config{})
	session, e := document.New(document.Config[string]{Owner: a, Window: "main", Journal: model.Journal(), Executor: manual, Codec: document.Codec[string]{Capture: func() (string, error) { return model.Snapshot().Text, nil }, Encode: func(s string) ([]byte, error) { return []byte(s), nil }, Decode: func(b []byte) (string, error) { return string(b), nil }, Install: func(s string) error { return model.Install(s, editor.Selection{}) }}})
	if e != nil {
		t.Fatal(e)
	}
	getState := func() document.State {
		t.Helper()
		var state document.State
		if e := a.Do(func() { state = session.State() }); e != nil {
			t.Fatal(e)
		}
		return state
	}
	if e = a.Post(func() {}); e != nil {
		t.Fatal(e)
	}
	path := t.TempDir() + "/manual.txt"
	var saveErr error
	if e = a.Do(func() { saveErr = session.Save(context.Background(), document.SaveOptions{Path: path}) }); e != nil || saveErr != nil {
		t.Fatal(e, saveErr)
	}
	if !getState().WorkerPending || !getState().OwnerPending {
		t.Fatal(getState())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, e = manual.StepNext(ctx); e != nil {
		t.Fatal(e)
	}
	if getState().WorkerPending || !getState().OwnerPending || !getState().Busy {
		t.Fatal(getState())
	}
	if count, e := a.DrainN(7); e != nil || count != 2 {
		t.Fatal(count, e)
	}
	if getState().Busy || getState().OwnerPending || getState().Last.Err != nil {
		t.Fatal(getState())
	}
}
