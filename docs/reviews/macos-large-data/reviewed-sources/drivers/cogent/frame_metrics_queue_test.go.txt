package cogent

import (
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
