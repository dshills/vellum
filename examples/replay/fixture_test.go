package replayconsumer

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func consumer(t *testing.T) (*Fixture, *execution.Manual) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 8, CompletionCapacity: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	e, err := execution.NewManual(4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	f, err := New(Config{Owner: a, Executor: e})
	if err != nil {
		t.Fatal(err)
	}
	return f, e
}
func invoke(t *testing.T, f *Fixture, action Action) {
	t.Helper()
	var result error
	if err := f.Owner.Do(func() { result = f.Invoke(t.Context(), action) }); err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatal(result)
	}
}
func snapshot(t *testing.T, f *Fixture) Snapshot {
	t.Helper()
	var s Snapshot
	if err := f.Owner.Do(func() {
		s = f.Snapshot()
		if err := f.Validate(); err != nil {
			t.Error(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCommandWorkloadBatchOneSevenSixtyFourHasIdenticalModelAndDeclarations(t *testing.T) {
	var reference CommandProof
	for i, batch := range []int{1, 7, 64} {
		proof, err := RunCommands(t.Context(), batch)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			reference = proof
		} else if !reflect.DeepEqual(proof.Final, reference.Final) || !reflect.DeepEqual(proof.Checkpoints, reference.Checkpoints) {
			t.Fatalf("batch%d changed declared model/semantics", batch)
		}
		s := proof.Final
		if !s.Loaded || s.Property != "North Star" || s.UndoCount != 0 || s.RedoCount != 1 || s.Dirty || !s.Sorted || s.Selected != "r-alpha" || s.WorkerCleanups != 1 || s.Status != "Task canceled" {
			t.Fatalf("incorrect workload state %+v", s)
		}
		if !reflect.DeepEqual(s.VisibleIDs, []string{"r-alpha", "r-beta", "r-delta", "r-eta"}) || len(s.Results) != 1 || s.Results[0].Outcome != "canceled" || s.Results[0].Checksum != 0 || len(s.UIAttempts) != 0 {
			t.Fatalf("projection/result/UI scope %+v", s)
		}
		if !s.Task.WorkerDone || !s.Task.OwnerDone || s.Task.Running || s.Task.AwaitingOwner || !proof.Executor.Idle() || !proof.Mailbox.Idle() {
			t.Fatalf("terminal lag %+v", proof)
		}
		if proof.Mailbox.DirtyWindows != 1 {
			t.Fatal("command proof concealed required frame/reconciliation")
		}
	}
}

func TestCommandRefusalPreservesModelHistoryAndDescriptor(t *testing.T) {
	f, _ := consumer(t)
	var before Snapshot
	if err := f.Owner.Do(func() {
		before = f.Snapshot()
		if err := f.Invoke(t.Context(), Action{ID: EditProperty, Value: "Harbor"}); !errors.Is(err, ErrNotOpen) {
			t.Errorf("unopened edit %v", err)
		}
		if !reflect.DeepEqual(f.Snapshot(), before) {
			t.Error("unopened edit mutated state")
		}
	}); err != nil {
		t.Fatal(err)
	}
	invoke(t, f, Action{ID: OpenProject})
	for _, action := range []Action{{ID: EditProperty, Value: ""}, {ID: EditProperty, Value: strings.Repeat("x", 257)}, {ID: EditProperty, Value: string([]byte{0xff})}, {ID: EditProperty, Value: "bad\x00text"}, {ID: DisableProperty, Value: "yes"}, {ID: SelectRow, Value: "8"}, {ID: "unknown"}, {ID: SortTable, Value: "ignored"}, {ID: Undo, ExpectedRevision: 2}} {
		if err := f.Owner.Do(func() {
			before = f.Snapshot()
			if err := f.Invoke(t.Context(), action); !errors.Is(err, ErrAction) {
				t.Errorf("bad payload %+v error%v", action, err)
			}
			if !reflect.DeepEqual(before, f.Snapshot()) {
				t.Errorf("refusal mutated %+v", action)
			}
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Owner.Do(func() {
		before = f.Snapshot()
		if err := f.Commands.Invoke(t.Context(), EditProperty); !errors.Is(err, ErrAction) {
			t.Errorf("argument-less registry edit %v", err)
		}
		if err := f.Commands.Invoke(t.Context(), DisableProperty); !errors.Is(err, ErrAction) {
			t.Errorf("argument-less disabled mutation %v", err)
		}
		canceled, cancel := context.WithCancel(t.Context())
		cancel()
		if err := f.Invoke(canceled, Action{ID: SortTable}); !errors.Is(err, context.Canceled) {
			t.Errorf("canceled command %v", err)
		}
		if err := f.Invoke(nil, Action{ID: SortTable}); err == nil {
			t.Error("nil context accepted")
		}
		if !reflect.DeepEqual(before, f.Snapshot()) {
			t.Error("invalid registry/context mutation")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDisabledModalStaleRevisionAndRemountAreExplicitCommandContracts(t *testing.T) {
	f, _ := consumer(t)
	invoke(t, f, Action{ID: OpenProject})
	invoke(t, f, Action{ID: DisableProperty, Value: "true"})
	if err := f.Owner.Do(func() {
		before := f.Snapshot()
		if err := f.Invoke(t.Context(), Action{ID: EditProperty, Value: "Harbor"}); !errors.Is(err, ErrDisabled) {
			t.Errorf("disabled edit %v", err)
		}
		if !reflect.DeepEqual(before, f.Snapshot()) {
			t.Error("disabled mutation")
		}
	}); err != nil {
		t.Fatal(err)
	}
	invoke(t, f, Action{ID: DisableProperty, Value: "false"})
	invoke(t, f, Action{ID: OpenModal})
	if err := f.Owner.Do(func() {
		before := f.Snapshot()
		if err := f.Invoke(t.Context(), Action{ID: StartTask}); !errors.Is(err, ErrModal) {
			t.Errorf("background modal command %v", err)
		}
		if !reflect.DeepEqual(before, f.Snapshot()) {
			t.Error("modal refusal mutated")
		}
		if order := ui.FocusOrder(f.View()); !reflect.DeepEqual(order, []string{CloseModal}) {
			t.Errorf("modal declared focus order %v", order)
		}
		if field, ok := ui.FocusTarget(f.View(), f.PropertyID()); !ok || field.Disabled {
			t.Error("background field disabled masks router modality qualification")
		}
	}); err != nil {
		t.Fatal(err)
	}
	invoke(t, f, Action{ID: CloseModal})
	old := snapshot(t, f)
	invoke(t, f, Action{ID: EditProperty, Value: "Harbor"})
	if err := f.Owner.Do(func() {
		before := f.Snapshot()
		if err := f.Invoke(t.Context(), Action{ID: Undo, CheckRevision: true, ExpectedRevision: old.HistoryRevision}); !errors.Is(err, ErrStale) {
			t.Errorf("stale undo %v", err)
		}
		if !reflect.DeepEqual(before, f.Snapshot()) {
			t.Error("stale mutation")
		}
	}); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, f)
	invoke(t, f, Action{ID: RemountProperty})
	after := snapshot(t, f)
	if before.PropertyID == after.PropertyID || after.PropertyIncarnation != before.PropertyIncarnation+1 || after.Property != before.Property || after.HistoryRevision != before.HistoryRevision || after.UndoCount != before.UndoCount {
		t.Fatalf("remount modified model/history %+v / %+v", before, after)
	}
	if err := f.Owner.Do(func() {
		if _, ok := ui.FocusTarget(f.View(), before.PropertyID); ok {
			t.Error("retired declared target remains")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestManualChecksumStaleResultAndOrdinaryQueueSaturation(t *testing.T) {
	f, e := consumer(t)
	invoke(t, f, Action{ID: OpenProject})
	invoke(t, f, Action{ID: StartTask})
	invoke(t, f, Action{ID: EditProperty, Value: "Harbor"})
	for range 8 {
		if err := f.Owner.Post(func() {}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := e.StepNext(t.Context()); err != nil {
		t.Fatal(err)
	}
	if p := f.Owner.Pending(); p.Ordinary != 8 || p.ReadyCompletions != 1 {
		t.Fatalf("full ordinary queue lost result %+v", p)
	}
	if err := drainReady(t.Context(), f.Owner, 1); err != nil {
		t.Fatal(err)
	}
	s := snapshot(t, f)
	if s.Property != "Harbor" || s.UndoCount != 1 || len(s.Results) != 1 || s.Results[0].Outcome != "stale" || s.Results[0].Checksum != 42 || s.WorkerCleanups != 1 || s.Status != "Task stale" || !f.Owner.Pending().Idle() || !e.Snapshot().Idle() {
		t.Fatalf("stale result mutated current model %+v", s)
	}
}

func TestDeclaredSemanticSnapshotsAreDetachedAndNeverRunWidgetCallbacks(t *testing.T) {
	f, _ := consumer(t)
	invoke(t, f, Action{ID: OpenProject})
	first := snapshot(t, f)
	first.VisibleIDs[0] = "corrupt"
	first.VisibleCells[0][0] = "corrupt"
	first.Semantics[0].Children = append(first.Semantics[0].Children, "corrupt")
	after := snapshot(t, f)
	if after.VisibleIDs[0] != "r-delta" || after.VisibleCells[0][0] != "Delta" || len(after.UIAttempts) != 0 {
		t.Fatalf("snapshot alias/callback %+v", after)
	}
	if after.UndoCount != 0 || after.HistoryRevision != 2 {
		t.Fatalf("fixture open not clean %+v", after)
	}
	if err := f.Owner.Do(func() {
		if err := f.Journal.Undo(); !errors.Is(err, history.ErrEmpty) {
			t.Error("open fixture entered edit history")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWindowClosureDropsManualTaskResultButStillRunsAcceptedCleanup(t *testing.T) {
	f, e := consumer(t)
	if err := f.Owner.Open(app.Window{ID: "survivor", Title: "Survivor", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "survivor-root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	invoke(t, f, Action{ID: OpenProject})
	invoke(t, f, Action{ID: StartTask})
	if err := f.Owner.CloseWindow(f.windowRef); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, f)
	if !before.Closed || before.Task.WorkerDone || !before.Task.OwnerDone || !before.Task.Running || !f.Owner.Pending().Idle() || e.Snapshot().Idle() {
		t.Fatalf("close hid accepted manual work %+v", before)
	}
	if _, err := e.StepNext(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := drainReady(t.Context(), f.Owner, 7); err != nil {
		t.Fatal(err)
	}
	after := snapshot(t, f)
	if after.WorkerCleanups != 1 || !after.Task.WorkerDone || !after.Task.OwnerDone || len(after.Results) != 0 || len(after.UIAttempts) != 0 || !f.Owner.Pending().Idle() || !e.Snapshot().Idle() {
		t.Fatalf("closed result delivered/cleanup lost %+v", after)
	}
}

func TestRefusedTaskExecutorDoesNotChangeFixtureOrReserveCredit(t *testing.T) {
	f, e := consumer(t)
	invoke(t, f, Action{ID: OpenProject})
	e.Close()
	if err := f.Owner.Do(func() {
		before := f.Snapshot()
		if err := f.Invoke(t.Context(), Action{ID: StartTask}); !errors.Is(err, execution.ErrClosed) {
			t.Errorf("executor refusal %v", err)
		}
		if !reflect.DeepEqual(before, f.Snapshot()) || !f.Owner.Pending().Idle() || f.manager.Pending().Active != 0 {
			t.Error("refused task changed fixture/credit")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
