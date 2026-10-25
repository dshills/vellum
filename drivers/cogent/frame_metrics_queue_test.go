package cogent

import (
	"cogentcore.org/core/core"
	"context"
	"image"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
)

func TestFrameMetricsReservedScrollSurvivesOrdinaryCompression(t *testing.T) {
	var q events.Deque
	scroll := events.NewScroll(image.Pt(1, 2), math32.Vector2{Y: 12}, 0)
	reserved := &tracedScrollEvent{MouseScroll: scroll}
	var _ events.DeferredEvent = reserved
	q.Send(reserved)
	ordinary := events.NewScroll(image.Pt(1, 2), math32.Vector2{Y: 7}, 0)
	q.Send(ordinary)
	if first := q.NextEvent(); first != reserved || reserved.FrameTraceEvent() != scroll || scroll.Delta.Y != 12 {
		t.Fatalf("reserved input replaced/combined: %v", first)
	}
	if next := q.NextEvent(); next != ordinary || ordinary.Delta.Y != 7 {
		t.Fatal("ordinary input changed")
	}
}

// No AfterFunc is registered here: this deterministically models a cancellation
// whose asynchronous notification has not run before native dispatch.
func TestFrameMetricsDispatchObservesCancellationBeforeComposition(t *testing.T) {
	for _, prior := range []string{"", "earlier terminal failure"} {
		t.Run(prior, func(t *testing.T) {
			trace := &core.FrameTrace{}
			scroll := events.NewScroll(image.Pt(1, 2), math32.Vector2{Y: 12}, 0)
			if err := trace.Begin(1, scroll); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if prior != "" {
				trace.Fail(1, prior)
			}
			reserved := &tracedScrollEvent{MouseScroll: scroll, ctx: ctx, trace: trace, tag: 1}
			if got := reserved.FrameTraceEvent(); got != scroll || scroll.Delta.Y != 12 {
				t.Fatal("admitted input changed")
			}
			trace.Fail(1, "composition drawer does not support encoding timing")
			trace.Retire(1, "retired after dispatch")
			sample, ok := trace.Take()
			want := prior
			if want == "" {
				want = context.Canceled.Error()
			}
			if !ok || sample.Failure != want || sample.Completed {
				t.Fatal("terminal precedence", sample, ok)
			}
		})
	}
}
