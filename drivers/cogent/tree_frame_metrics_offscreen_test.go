//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func treeMetricsFixture(t *testing.T, count int, callback func(ui.TreeEvent)) (*Manual, *virtual.Outline, *FrameMetrics, *frameWindowMetrics) {
	t.Helper()
	outline := nativeOutline(t, count, true)
	m, a := gridManual(t, manualConfig(t, 1), func() ui.Node {
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: outline, Grow: true, TreeChange: callback}
	})
	unitSettle(t, m)
	metrics, err := NewFrameMetrics(2)
	if err != nil {
		t.Fatal(err)
	}
	b, err := metrics.bind(a, m.implementation.find("one").state)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(b.retire)
	return m, outline, metrics, b
}
func TestTreeFrameMetricsActualScrollPublishesGeneration(t *testing.T) {
	changes := 0
	m, o, metrics, b := treeMetricsFixture(t, 100_000, func(e ui.TreeEvent) {
		if e.Reason != ui.TreeViewport {
			t.Error(e)
		}
		changes++
	})
	changes = 0 // mounting also publishes its initial viewport
	before := treeState(t, b.key.owner, o)
	var inner error
	if err := b.key.owner.Do(func() { inner = EnqueueTreeScroll(testContext, b.key.owner, o, 12, 1) }); err != nil || inner != nil {
		t.Fatal(err, inner)
	}
	unitSettle(t, m)
	after := treeState(t, b.key.owner, o)
	if after.Viewport.ScrollY != before.Viewport.ScrollY+12 || after.Generation != before.Generation+1 || changes != 1 {
		t.Fatal("actual native scroll/owner notification missing", before, after, changes)
	}
	if b.trace.OwnerTag() != 1 {
		t.Fatal("owner transition was not linked to actual input")
	}
	// Manual has no native timing or GPU composition. Close its observer owner
	// terminal explicitly, then require native composition to remain unqualified.
	b.trace.OwnerWork(1, 0)
	poll := metrics.Poll(1)
	if poll.Found && poll.Sample.Completed {
		t.Fatal("manual paint qualified native composition", poll)
	}
	b.trace.Retire(1, "manual observer has no native composition")
	poll = metrics.Poll(1)
	if !poll.Found || poll.Sample.Completed || poll.Sample.Generation != after.Generation {
		t.Fatal(poll)
	}
}
func TestTreeFrameMetricsBoundaryScrollFailsTerminal(t *testing.T) {
	m, o, metrics, b := treeMetricsFixture(t, 1000, nil)
	var inner error
	if err := b.key.owner.Do(func() { inner = EnqueueTreeScroll(testContext, b.key.owner, o, -12, 1) }); err != nil || inner != nil {
		t.Fatal(err, inner)
	}
	unitSettle(t, m)
	b.trace.OwnerWork(1, 0)
	poll := metrics.Poll(1)
	if !poll.Found || poll.Pending != 0 || poll.Sample.Completed || poll.Sample.Failure != "traced tree input did not publish an owner transition" {
		t.Fatal(poll)
	}
	if treeState(t, b.key.owner, o).Viewport.ScrollY != 0 {
		t.Fatal("boundary changed model")
	}
}
func TestTreeFrameMetricsAdmissionAndCanceledDelivery(t *testing.T) {
	m, o, metrics, b := treeMetricsFixture(t, 1000, nil)
	check := func(dy float64, tag uint64, want error) {
		t.Helper()
		var got error
		if err := b.key.owner.Do(func() { got = EnqueueTreeScroll(testContext, b.key.owner, o, dy, tag) }); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(got, want) {
			t.Fatalf("dy=%g tag=%d got %v want %v", dy, tag, got, want)
		}
	}
	for _, dy := range []float64{0, math.NaN(), math.Inf(1), 1 << 21} {
		check(dy, 1, ErrFrameMetricsConfig)
	}
	check(12, 0, ErrFrameMetricsConfig)
	var request virtual.OutlineRequest
	var preparation error
	if err := b.key.owner.Do(func() { request, preparation = o.Prepare(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); err != nil || preparation != nil {
		t.Fatal(err, preparation)
	}
	check(12, 1, ErrTreeLayout)
	if err := b.key.owner.Do(func() { preparation = o.Cancel(request) }); err != nil || preparation != nil {
		t.Fatal(err, preparation)
	}
	ctx, cancel := context.WithCancel(testContext)
	defer cancel()
	var inner error
	if err := b.key.owner.Do(func() { inner = EnqueueTreeScroll(ctx, b.key.owner, o, 12, 1) }); err != nil || inner != nil {
		t.Fatal(err, inner)
	}
	check(12, 2, ErrFrameMetricsBusy)
	cancel()
	unitSettle(t, m)
	b.trace.OwnerWork(1, 0)
	poll := metrics.Poll(1)
	if !poll.Found || poll.Pending != 0 || poll.Sample.Completed || poll.Sample.Failure != context.Canceled.Error() {
		t.Fatal("cancellation bypassed input/owner terminals", poll)
	}
	if treeState(t, b.key.owner, o).Viewport.ScrollY != 12 {
		t.Fatal("cancellation incorrectly retracted admitted input")
	}
	check(12, 1, ErrFrameMetricsConfig)
}
func TestTreeFrameMetricsPriorOwnerDisablesTarget(t *testing.T) {
	o := nativeOutline(t, 1000, true)
	disabled := false
	m, a := gridManual(t, manualConfig(t, 1), func() ui.Node {
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true, Disabled: disabled}
	})
	unitSettle(t, m)
	metrics, _ := NewFrameMetrics(2)
	b, err := metrics.bind(a, m.implementation.find("one").state)
	if err != nil {
		t.Fatal(err)
	}
	defer b.retire()
	if err = a.PostWindow(fixtureWindowRef(a, "one"), func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	var inner error
	if err = a.Do(func() { inner = EnqueueTreeScroll(testContext, a, o, 12, 1) }); err != nil || inner != nil {
		t.Fatal(err, inner)
	}
	unitSettle(t, m)
	poll := metrics.Poll(1)
	if !poll.Found || poll.Pending != 0 || poll.Sample.Completed || poll.Sample.Failure != "traced input did not post owner work" {
		t.Fatal(poll)
	}
	if treeState(t, a, o).Viewport.ScrollY != 0 {
		t.Fatal("disabled input changed model")
	}
}
