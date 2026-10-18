package cogent

import (
	"errors"
	"math"
	"sync"
	"testing"

	"cogentcore.org/core/events"
)

func visibilityQueueLen(source *events.Source) int {
	source.Deque.Mu.Lock()
	defer source.Deque.Mu.Unlock()
	return len(source.Deque.Back) + len(source.Deque.Front)
}
func TestWindowVisibilityEarlyCoalescingAndLifetime(t *testing.T) {
	var source, other events.Source
	v := newWindowVisibility()
	v.observe(false)
	v.observe(true)
	if got, _ := v.snapshot(); got != (DesktopWindowVisibility{true, true, 2}) {
		t.Fatal(got)
	}
	if err := v.bind(&source); err != nil {
		t.Fatal(err)
	}
	if err := v.bind(&source); err != nil || visibilityQueueLen(&source) != 1 {
		t.Fatal(err)
	}
	if err := v.bind(&other); !errors.Is(err, ErrDesktopSession) {
		t.Fatal(err)
	}
	for range 100 {
		v.observe(false)
		v.observe(true)
		v.observe(true)
	}
	if visibilityQueueLen(&source) != 1 {
		t.Fatal("unbounded notifications")
	}
	if v.consume(&other, v.pending) {
		t.Fatal("cross-source accepted")
	}
	e, ok := source.Deque.PollEvent()
	if !ok {
		t.Fatal("missing early redraw")
	}
	node := e.(*visibilityFIFO)
	if node.state != v || node.FrameTraceEvent().(*events.CustomEvent).Data.(*visibilityWake).state != v {
		t.Fatal("identity lost")
	}
	wake := node.FrameTraceEvent().(*events.CustomEvent).Data.(*visibilityWake)
	if !v.consume(&source, wake) || v.consume(&source, wake) {
		t.Fatal("not exactly once")
	}
	// An ordinary custom event must never be replaced by our deferred node.
	ordinary := &events.CustomEvent{}
	ordinary.Data = "ordinary"
	ordinary.Typ = events.Custom
	ordinary.Init()
	source.Deque.Send(ordinary)
	v.observe(false)
	v.observe(true)
	if got, _ := source.Deque.PollEvent(); got != ordinary {
		t.Fatal("overwrote preceding custom event")
	}
	held, _ := source.Deque.PollEvent()
	v.observe(false)
	if v.consume(&source, v.pending) {
		t.Fatal("painted hidden latest state")
	}
	v.observe(true)
	if visibilityQueueLen(&source) != 1 {
		t.Fatal("lost later visible transition")
	}
	if v.consume(&source, wake) || v.pending == nil {
		t.Fatal("stale receipt consumed a newer wake")
	}
	v.close()
	v.close()
	v.observe(false)
	v.observe(true)
	if visibilityQueueLen(&source) != 0 || v.source != nil || v.pending != nil {
		t.Fatal("close retained source/node")
	}
	if held == nil || v.consume(&source, wake) {
		t.Fatal("dequeued node revived closed state")
	}
	if _, err := v.snapshot(); !errors.Is(err, ErrDesktopClosed) {
		t.Fatal(err)
	}
	if err := v.bind(&source); !errors.Is(err, ErrDesktopClosed) {
		t.Fatal(err)
	}
}
func TestWindowVisibilityUnknownOverflowAndRace(t *testing.T) {
	var absent *windowVisibility
	if got, err := absent.snapshot(); err != nil || got != (DesktopWindowVisibility{}) {
		t.Fatal(got, err)
	}
	var source events.Source
	v := newWindowVisibility()
	_ = v.bind(&source)
	if got, _ := v.snapshot(); got.Known || visibilityQueueLen(&source) != 0 {
		t.Fatal(got)
	}
	v.observe(true)
	held := v.pending
	if held == nil || visibilityQueueLen(&source) != 1 {
		t.Fatal("missing queued overflow precondition")
	}
	v.value.Sequence = math.MaxUint64
	v.observe(false)
	if _, err := v.snapshot(); !errors.Is(err, ErrDesktopBudget) || v.source != nil || v.pending != nil || visibilityQueueLen(&source) != 0 || v.consume(&source, held) {
		t.Fatal(err)
	}
	v = newWindowVisibility()
	_ = v.bind(&source)
	var wg sync.WaitGroup
	for j := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				v.observe((i+j)%2 == 0)
				v.snapshot()
				v.consume(&source, nil)
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); v.close() }()
	wg.Wait()
	if visibilityQueueLen(&source) != 0 || v.source != nil {
		t.Fatal("race retained nodes")
	}
}

func TestWindowVisibilityBridgeRetirement(t *testing.T) {
	v := newWindowVisibility()
	b := newAccessibilityBridgeWithVisibility(nil, v)
	accessibilityRegistry.Lock()
	published := accessibilityRegistry.bridges[b.token]
	initialized := published == b && published.visibility == v
	accessibilityRegistry.Unlock()
	if !initialized {
		t.Fatal("registry published incomplete visibility state")
	}
	var source events.Source
	_ = v.bind(&source)
	v.observe(true)
	b.closeFrom("visibility-test")
	if _, err := v.snapshot(); !errors.Is(err, ErrDesktopClosed) || visibilityQueueLen(&source) != 0 {
		t.Fatal(err)
	}
}
