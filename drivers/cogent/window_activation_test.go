package cogent

import (
	"errors"
	"sync"
	"testing"

	"cogentcore.org/core/events"
)

func activationQueueLen(source *events.Source) int {
	source.Deque.Mu.Lock()
	defer source.Deque.Mu.Unlock()
	return len(source.Deque.Back) + len(source.Deque.Front)
}
func activationNext(t *testing.T, source *events.Source) *activationWake {
	t.Helper()
	e, ok := source.Deque.PollEvent()
	if !ok {
		t.Fatal("missing activation event")
	}
	return e.(*activationFIFO).FrameTraceEvent().(*events.CustomEvent).Data.(*activationWake)
}
func TestWindowActivationBoundedLatestPublication(t *testing.T) {
	var source, foreign events.Source
	v := &windowActivation{}
	if !v.active() {
		t.Fatal("unknown activation changed existing behavior")
	}
	v.observe(false)
	if !v.active() {
		t.Fatal("main-thread observation changed native appearance directly")
	}
	if err := v.bind(&source); err != nil {
		t.Fatal(err)
	}
	if err := v.bind(&foreign); !errors.Is(err, ErrDesktopSession) {
		t.Fatal("cross-window rebind", err)
	}
	for i := range 100 {
		v.observe(i%2 == 0)
	}
	if activationQueueLen(&source) != 1 {
		t.Fatal("unbounded focus notifications")
	}
	wake := activationNext(t, &source)
	if v.consume(&foreign, wake) || !v.consume(&source, wake) || v.consume(&source, wake) || v.active() {
		t.Fatal("source/exactly-once/latest-state publication")
	}
	v.observe(true)
	v.observe(false)
	if v.consume(&source, activationNext(t, &source)) {
		t.Fatal("unchanged latest state requested restyle")
	}
	v.observe(true)
	next := activationNext(t, &source)
	if v.consume(&source, wake) || v.active() || !v.consume(&source, next) || !v.active() {
		t.Fatal("stale wake consumed current transition")
	}
	v.observe(false)
	held := activationNext(t, &source)
	v.close()
	v.close()
	v.observe(true)
	if v.consume(&source, held) || v.source != nil || v.pending != nil || activationQueueLen(&source) != 0 {
		t.Fatal("retired notification revived state")
	}
	if err := v.bind(&source); !errors.Is(err, ErrDesktopClosed) {
		t.Fatal("retired rebind", err)
	}
}
func TestWindowActivationConcurrentCloseRemovesQueuedNode(t *testing.T) {
	var source events.Source
	v := &windowActivation{}
	if err := v.bind(&source); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for j := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				v.observe((i+j)%2 == 0)
				v.active()
				v.consume(&source, nil)
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); v.close() }()
	wg.Wait()
	if v.source != nil || v.pending != nil || activationQueueLen(&source) != 0 {
		t.Fatal("close retained source or queued node")
	}
}
func TestWindowActivationBridgeRetirement(t *testing.T) {
	b := newAccessibilityBridge(nil)
	var source events.Source
	if err := b.activation.bind(&source); err != nil {
		t.Fatal(err)
	}
	b.activation.observe(false)
	if activationQueueLen(&source) != 1 {
		t.Fatal("missing retirement precondition")
	}
	b.closeFrom("activation-test")
	if b.activation.pending != nil || b.activation.source != nil || activationQueueLen(&source) != 0 {
		t.Fatal("bridge retirement retained activation work")
	}
}
