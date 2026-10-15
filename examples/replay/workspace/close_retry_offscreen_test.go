//go:build offscreen

package workspace

import (
	"context"
	"errors"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
)

func TestReviewerBusyDriverClosePreservesBorrowedResourcesForRetry(t *testing.T) {
	b, err := NewBackend(context.Background(), 10_000, 7)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := b.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	source := b.Fixture.Palette.Source("block")
	called, resourcesLive := false, false
	var closeErr error
	if err := b.Fixture.Owner.Post(func() {
		// Step holds Manual.busy, while owner callbacks run synchronously.
		// Close calls Application.Close (mu only), then Manual.Close refuses
		// at enter before trying a nested Application.Do. This is deterministic
		// and uses no callback, timer, sleep, or concurrent operation injection.
		called = true
		closeErr = b.Close(context.Background())
		resourcesLive = !b.Executor.Snapshot().Closed && !b.Clock.Snapshot().Closed && source.Valid()
	}); err != nil {
		t.Fatal(err)
	}
	_, stepErr := b.Driver.Step(context.Background(), 64)
	if !called || !errors.Is(closeErr, cogent.ErrManualBusy) {
		t.Fatal("did not exercise busy Manual retirement", called, closeErr, stepErr)
	}
	if !resourcesLive {
		t.Fatal("failed driver retirement closed still-borrowed resources")
	}
	if stepErr != nil && !errors.Is(stepErr, app.ErrClosed) && !errors.Is(stepErr, cogent.ErrManualClosed) {
		t.Fatal("unexpected active-step failure", stepErr)
	}
	// After Step relinquishes Manual.busy, closure can complete normally.
	if err := b.Close(context.Background()); err != nil {
		t.Fatal("retirement retry failed", err)
	}
	if !b.Executor.Snapshot().Closed || !b.Clock.Snapshot().Closed || source.Valid() {
		t.Fatal("successful retry did not retire owned resources")
	}
}
