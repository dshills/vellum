package cogent

import (
	"math"
	"sync"

	"cogentcore.org/core/events"
)

// DesktopWindowVisibility is a detached observation of the Mac window's
// NSWindowOcclusionStateVisible bit. It is not a drawable or presentation fence.
// Sequence changes on the initial observation and each observed visibility change;
// it is scoped to the original WindowRef and never reused after retirement.
type DesktopWindowVisibility struct {
	Known    bool
	Visible  bool
	Sequence uint64
}

// A window retains at most one redraw node. No native pointer or widget is kept
// here. Main-thread notifications only publish values into the owning FIFO.
type windowVisibility struct {
	mu                sync.Mutex
	value             DesktopWindowVisibility
	source            *events.Source
	pending           *visibilityWake
	closed, exhausted bool
}

type visibilityWake struct{ state *windowVisibility }

type visibilityFIFO struct {
	events.Event
	state *windowVisibility
}

func (*visibilityFIFO) Ready() bool                     { return true }
func (*visibilityFIFO) Merge(events.DeferredEvent) bool { return false }
func (e *visibilityFIFO) FrameTraceEvent() events.Event { return e.Event }

func newWindowVisibility() *windowVisibility { return &windowVisibility{} }

// Binding is delayed until OnShow installs the owner/source. Early native
// samples are retained; a currently visible state is flushed exactly once.
func (v *windowVisibility) bind(source *events.Source) error {
	if v == nil || source == nil {
		return ErrDesktopSession
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return ErrDesktopClosed
	}
	if v.source != nil && v.source != source {
		return ErrDesktopSession
	}
	if v.source == source {
		return nil
	}
	v.source = source
	v.queueLocked()
	return nil
}
func (v *windowVisibility) observe(visible bool) {
	if v == nil {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed || v.value.Known && v.value.Visible == visible {
		return
	}
	if v.value.Sequence == math.MaxUint64 {
		v.exhausted = true
		v.closeLocked()
		return
	}
	v.value = DesktopWindowVisibility{Known: true, Visible: visible, Sequence: v.value.Sequence + 1}
	v.queueLocked()
}
func (v *windowVisibility) queueLocked() {
	if !v.value.Known || !v.value.Visible || v.pending != nil || v.source == nil {
		return
	}
	wake := &visibilityWake{state: v}
	e := &events.CustomEvent{}
	e.Data = wake
	e.Typ = events.Custom
	e.Init()
	v.pending = wake
	_ = v.source.Deque.SendDeferred(func() (events.DeferredEvent, error) {
		return &visibilityFIFO{Event: e, state: v}, nil
	})
}
func (v *windowVisibility) snapshot() (DesktopWindowVisibility, error) {
	if v == nil {
		return DesktopWindowVisibility{}, nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.exhausted {
		return DesktopWindowVisibility{}, ErrDesktopBudget
	}
	if v.closed {
		return DesktopWindowVisibility{}, ErrDesktopClosed
	}
	return v.value, nil
}
func (v *windowVisibility) consume(source *events.Source, wake *visibilityWake) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed || source == nil || source != v.source {
		return false
	}
	if wake == nil || v.pending != wake {
		return false
	}
	v.pending = nil
	return v.value.Known && v.value.Visible
}
func (v *windowVisibility) close() {
	if v == nil {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.closeLocked()
}
func (v *windowVisibility) closeLocked() {
	if v.closed {
		return
	}
	v.closed = true
	if source := v.source; source != nil {
		// Ready/Merge and this matcher never acquire v.mu: all producers use the
		// state -> deque lock order; delivery consumes only after releasing the deque.
		_ = source.Deque.UpdateDeferred(func() error { return nil }, func(e events.DeferredEvent) bool {
			node, ok := e.(*visibilityFIFO)
			return ok && node.state == v
		})
	}
	v.source = nil
	v.pending = nil
}

func bindWindowVisibility(w *nativeWindow) error {
	b := windowAccessibilityBridge(w)
	if b == nil || b.visibility == nil {
		return nil
	} // offscreen/other platforms
	w.visibility.Store(b.visibility)
	return b.visibility.bind(w.body.SystemEvents())
}

func handleWindowVisibility(w *nativeWindow, e events.Event) bool {
	ce, ok := e.(*events.CustomEvent)
	if !ok {
		return false
	}
	wake, ok := ce.Data.(*visibilityWake)
	if !ok || wake == nil || wake.state == nil {
		return false
	}
	v := wake.state
	e.SetHandled()
	if w != nil && w.body != nil && w.body.Scene.RenderWindow() != nil &&
		!w.closing.Load() && w.window.Ref().Context().Err() == nil &&
		v == w.visibility.Load() && v.consume(w.body.SystemEvents(), wake) {
		// Merely requesting paint can skip composition after failed acquisition
		// consumed dirty flags. This runs on the originating render owner.
		w.desktop.invalidate(w)
		w.body.Scene.NeedsRender()
	}
	return true
}
