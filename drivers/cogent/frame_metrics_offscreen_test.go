//go:build offscreen

package cogent

import (
	"testing"

	"cogentcore.org/core/events"
	"github.com/dshills/vellum/ui"
)

// This test exercises real event routing and owner refusal/unchanged terminals.
// Manual has no GPU timing; it explicitly closes the synthetic observer owner
// interval and never qualifies a performance/composition sample.
func TestFrameMetricsActualUnchangedScrollIsFailedTerminal(t *testing.T) {
	grid, _ := testNativeGrid(t, 1000)
	m, a := gridManual(t, manualConfig(t, 1), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: grid, Grow: true} })
	gridFrame(t, m)
	metrics, _ := NewFrameMetrics(2)
	w := m.implementation.find("one").state
	b, err := metrics.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	defer b.retire()
	if err := a.Do(func() { err = EnqueueGridScroll(testContext, a, grid, 0, -12, 1) }); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	b.trace.OwnerWork(1, 0) // explicit manual owner terminal, no performance claim
	got := metrics.Snapshot()
	if got.Pending != 0 || len(got.Samples) != 1 || got.Samples[0].Completed || got.Samples[0].Failure != "traced grid input did not publish an owner transition" {
		t.Fatalf("unchanged input wedged/qualified: %+v", got)
	}
	state, _ := grid.State()
	if state.Viewport.ScrollY != 0 {
		t.Fatal("boundary input changed viewport")
	}
	// Reserved dispatch still unwraps to the real Scroll expected by the widget;
	// the trace cannot substitute a Custom callback invocation.
	var _ events.DeferredEvent = &tracedScrollEvent{}
}

func TestFrameMetricsPriorOwnerWakeDisablesTargetBeforeDispatch(t *testing.T) {
	grid, _ := testNativeGrid(t, 1000)
	disabled := false
	m, a := gridManual(t, manualConfig(t, 1), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: grid, Grow: true, Disabled: disabled} })
	gridFrame(t, m)
	metrics, _ := NewFrameMetrics(2)
	w := m.implementation.find("one").state
	b, err := metrics.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	defer b.retire()
	if err := a.PostWindow(fixtureWindowRef(a, "one"), func() { disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	if outer := a.Do(func() { err = EnqueueGridScroll(testContext, a, grid, 0, 12, 1) }); outer != nil || err != nil {
		t.Fatal(outer, err)
	}
	gridFrame(t, m)
	got := metrics.Snapshot()
	if got.Pending != 0 || len(got.Samples) != 1 || got.Samples[0].Completed || got.Samples[0].Failure != "traced input did not post owner work" {
		t.Fatalf("unhandled disabled input wedged/qualified: %+v", got)
	}
	state, _ := grid.State()
	if state.Viewport.ScrollY != 0 {
		t.Fatal("disabled input changed viewport")
	}
}
