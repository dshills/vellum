package cogent

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakePaintMode struct {
	demand  bool
	changes int
}

func (p *fakePaintMode) SetOnDemand(on bool) { p.demand = on; p.changes++ }
func TestDefaultRenderingUsesInvalidation(t *testing.T) {
	p := &fakePaintMode{}
	if err := configureRendering(p, false); err != nil || !p.demand {
		t.Fatal(err)
	}
	if err := configureRendering(p, true); err != nil || p.demand {
		t.Fatal(err)
	}
	if configureRendering(struct{}{}, false) == nil {
		t.Fatal("unsupported painter accepted")
	}
}
func TestUnreadyCanvasRecoveryAndAnimation(t *testing.T) {
	g := &GPUCanvas{}
	if !errors.Is(g.RebuildResources(), ErrFrameNotReady) {
		t.Fatal("unready rebuild")
	}
	if !errors.Is(g.RequestFrameAfter(time.Millisecond), ErrFrameNotReady) {
		t.Fatal("unready animation")
	}
	if g.RequestFrameAfter(-time.Second) == nil {
		t.Fatal("negative delay accepted")
	}
}

func TestLossSubscriptionCannotOutliveClosingWindow(t *testing.T) {
	for range 100 {
		w := &nativeWindow{}
		var active atomic.Int32
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			installDeviceLoss(w, func() func() { active.Add(1); return func() { active.Add(-1) } })
		}()
		go func() { defer wg.Done(); w.closing.Store(true); stopDeviceLoss(w) }()
		wg.Wait()
		if active.Load() != 0 {
			t.Fatal("late loss subscription leaked")
		}
	}
	watchDeviceLoss(&nativeWindow{}, func(error) { t.Fatal("unready native window notified loss") })
	if nativeSurface(nil) != nil {
		t.Fatal("nil native surface")
	}
}

func TestAlreadyLostDeviceMayCloseDuringSubscribe(t *testing.T) {
	w := &nativeWindow{}
	canceled := false
	done := make(chan struct{})
	go func() {
		installDeviceLoss(w, func() func() { w.closing.Store(true); stopDeviceLoss(w); return func() { canceled = true } })
		close(done)
	}()
	select {
	case <-done:
		if !canceled {
			t.Fatal("immediate-loss subscription leaked")
		}
	case <-time.After(time.Second):
		t.Fatal("immediate loss deadlocked shutdown")
	}
}

func TestRetiredCanvasRejectsResourceAPIs(t *testing.T) {
	w := &canvasWidget{}
	w.retired.Store(true)
	g := &GPUCanvas{widget: w}
	if g.NativeWindow() != nil {
		t.Fatal("retired window exposed")
	}
	if !errors.Is(g.RequestFrameAfter(time.Millisecond), ErrFrameNotReady) {
		t.Fatal("retired animation accepted")
	}
	if !errors.Is(g.RebuildResources(), ErrFrameNotReady) {
		t.Fatal("retired rebuild accepted")
	}
	if _, err := g.Snapshot(); !errors.Is(err, ErrFrameNotReady) {
		t.Fatal("retired readback accepted")
	}
	if g.Sync(w) == nil {
		t.Fatal("retired sync accepted")
	}
}
