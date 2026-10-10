package eventqueue

import (
	"cogentcore.org/core/events"
	"testing"
)

func TestBurstCommandsPreserveOrder(t *testing.T) {
	var source events.Source
	for i := 0; i < 1000; i++ {
		Send(&source, i)
	}
	for i := 0; i < 1000; i++ {
		e, ok := source.Deque.PollEvent()
		if !ok {
			t.Fatalf("lost action %d", i)
		}
		if got := e.(*events.CustomEvent).Data.(int); got != i {
			t.Fatalf("got action %d, want %d", got, i)
		}
	}
	if _, ok := source.Deque.PollEvent(); ok {
		t.Fatal("duplicate action")
	}
}
