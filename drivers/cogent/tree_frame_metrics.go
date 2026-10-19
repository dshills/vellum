package cogent

import (
	"context"
	"image"
	"math"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// EnqueueTreeScroll must run on the application owner (Do/PostWait). It reserves
// an actual MouseScroll in the originating native FIFO using completed tree
// geometry and a logical-pixel delta. The existing tree handler and owner path
// perform the mutation; this function does not invoke descriptor callbacks.
// Only one traced event per window can be outstanding. Cancellation retires
// measurement but cannot retract input already delivered to the native handler.
func EnqueueTreeScroll(ctx context.Context, a *app.Application, model *virtual.Outline, dy float64, tag uint64) error {
	if ctx == nil || a == nil || model == nil || tag == 0 || math.IsNaN(dy) || math.IsInf(dy, 0) || math.Abs(dy) > 1<<20 || dy == 0 {
		return ErrFrameMetricsConfig
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	outlineRegistry.Lock()
	t := outlineRegistry.models[model]
	outlineRegistry.Unlock()
	if t == nil || t.owner != a || !t.live() {
		return ErrTreeClosed
	}
	if !t.eligible() {
		return ErrFrameMetricsDisabled
	}
	state, err := model.State()
	if err != nil {
		return err
	}
	measured := t.currentMeasurement()
	if state.Pending || measured == nil || measured.Generation != state.Generation || measured.Stamp != state.Stamp || measured.Viewport != state.Viewport || !validGridScale(measured.Scale) {
		return ErrTreeLayout
	}
	bounds := measured.Bounds
	body := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+int(math.Round(float64(state.Viewport.Width)*float64(measured.Scale))), bounds.Min.Y+int(math.Round(float64(state.Viewport.Height)*float64(measured.Scale)))).Intersect(bounds)
	if body.Empty() {
		return ErrTreeLayout
	}
	position := image.Pt((body.Min.X+body.Max.X)/2, (body.Min.Y+body.Max.Y)/2)
	frameMetricsRegistry.Lock()
	b := frameMetricsRegistry.windows[frameWindowKey{a, t.window, t.windowRef}]
	frameMetricsRegistry.Unlock()
	if b == nil || !b.metrics.valid() {
		return ErrFrameMetricsConfig
	}
	m := b.metrics
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectLocked()
	if b.retired || b.window.closing.Load() {
		return ErrTreeClosed
	}
	if b.tag != 0 {
		return ErrFrameMetricsBusy
	}
	if _, reserved := m.tags[tag]; reserved {
		return ErrFrameMetricsConfig
	}
	if len(m.tags) >= m.capacity {
		return ErrFrameMetricsBudget
	}
	source := b.window.source.Load()
	if source == nil {
		return ErrTreeLayout
	}
	event := events.NewScroll(position, math32.Vector2{Y: float32(dy) * measured.Scale}, 0)
	event.Init()
	if err := b.trace.Begin(tag, event); err != nil {
		return err
	}
	b.tag = tag
	m.tags[tag] = 0
	b.stop = context.AfterFunc(ctx, func() { b.trace.Fail(tag, ctx.Err().Error()) })
	recordPresentationReceipt(b.window, tag)
	source.Deque.Send(&tracedScrollEvent{MouseScroll: event})
	return nil
}

func tracedTreeIntent(t *outlineBinding) (*core.FrameTrace, uint64) {
	if t.host == nil || t.host.Scene == nil {
		return nil, 0
	}
	trace := t.host.Scene.FrameTrace()
	return trace, trace.CurrentTag()
}
func (t *outlineBinding) postTraced(trace *core.FrameTrace, tag uint64, reason ui.TreeChangeReason, fn func() error) {
	trace.ExpectOwner(tag)
	err := t.owner.PostWindow(t.windowRef, func() {
		returned := false
		defer func() {
			if !returned {
				trace.Fail(tag, "traced owner transition did not return")
				trace.Transition(tag, 0)
			}
		}()
		before, beforeErr := t.model.State()
		t.apply(reason, true, fn)
		after, afterErr := t.model.State()
		if beforeErr != nil || afterErr != nil || !t.live() || before.Generation == after.Generation {
			trace.Fail(tag, "traced tree input did not publish an owner transition")
			trace.Transition(tag, 0)
		} else {
			trace.Transition(tag, after.Generation)
		}
		returned = true
	})
	if err != nil {
		trace.Fail(tag, "traced tree owner posting refused")
		trace.Transition(tag, 0)
		trace.OwnerWork(tag, 0)
	}
}
