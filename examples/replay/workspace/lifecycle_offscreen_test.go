//go:build offscreen

package workspace

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/replay"
)

func TestWorkspaceManifestAndStaleRefusals(t *testing.T) {
	b, err := NewBackend(context.Background(), 10000, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := b.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	r, _ := replay.NewRunner(replay.RunnerConfig{EventBatch: 1}, b)
	manifest := b.Manifest()
	manifest.FixtureSHA256 = manifest.ThemeSHA256
	if got, err := r.Run(context.Background(), replay.Record{Manifest: manifest}); !errors.Is(err, replay.ErrBackend) || got.Applied != 0 {
		t.Fatal("fixture mismatch admitted", got, err)
	}
	manifest = b.Manifest()
	if err := b.CheckManifest(context.Background(), manifest); err != nil {
		t.Fatal(err)
	}
	w := manifest.Windows[0]
	if _, err := b.Capture(context.Background(), replay.Operation{Kind: replay.Capture, Window: w.ID, WindowIncarnation: w.Incarnation, Label: "unrendered"}); !errors.Is(err, cogent.ErrManualLayout) {
		t.Fatal("implicit render", err)
	}
	before, err := b.Readiness(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	idle, err := r.RunUntilIdle(context.Background())
	if err != nil || !idle.QuiescentNow || !idle.Residual.Now.Equal(before.Now) {
		t.Fatal("initial eligible paints did not settle at the same time", idle, err)
	}
	op := replay.Operation{Kind: replay.Command, Window: "workspace", WindowIncarnation: 999, Action: "open"}
	if err := b.Apply(context.Background(), op); !errors.Is(err, cogent.ErrManualStale) {
		t.Fatal("stale window admitted", err)
	}
	op.WindowIncarnation = 2
	for _, v := range manifest.Windows {
		if v.ID == "workspace" {
			op.WindowIncarnation = v.Incarnation
		}
	}
	op.CheckRevision = true
	op.ExpectedRevision = 999
	if err := b.Apply(context.Background(), op); !errors.Is(err, replay.ErrBackend) {
		t.Fatal("stale revision admitted", err)
	}
	// Opening a project admits exactly one worker, but idle/step never selects it.
	op.CheckRevision = false
	op.ExpectedRevision = 0
	if err := b.Apply(context.Background(), op); err != nil {
		t.Fatal(err)
	}
	ids := b.Executor.PendingIDs()
	if len(ids) != 1 {
		t.Fatal(ids)
	}
	start := b.Clock.Now()
	for range 3 {
		if _, err := r.RunUntilIdle(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	model, err := b.Model()
	if err != nil || model.Loaded || !reflect.DeepEqual(ids, b.Executor.PendingIDs()) || !b.Clock.Now().Equal(start) {
		t.Fatal("implicit worker/time advancement", err)
	}
	// Retirement rejects stale operations while owned teardown still executes the
	// canceled admitted callback and releases its reservation.
	b.Fixture.Owner.Close()
	if err := b.Apply(context.Background(), replay.Operation{Kind: replay.Worker, WorkerID: uint64(ids[0])}); !errors.Is(err, app.ErrClosed) {
		t.Fatal("closed-owner worker accepted", err)
	}
	if err := b.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !b.Executor.Snapshot().Idle() || !b.Fixture.Owner.Pending().Idle() {
		t.Fatal("closure leaked admitted work")
	}
}
func TestBackendCanceledCloseCanRetry(t *testing.T) {
	b, err := NewBackend(context.Background(), 10000, 1)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := b.Close(cleanup); err != nil {
		t.Fatal(err)
	}
}
