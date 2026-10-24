package cogent

import (
	"sync"

	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
)

const windowActivationProperty = "vellum.window-activation"

// Main-thread callbacks retain only the latest value and at most one FIFO node.
// Published appearance changes only on the originating native render owner, so
// every row in one style pass observes the same activation state.
type windowActivation struct {
	mu                               sync.Mutex
	source                           *events.Source
	pending                          *activationWake
	known, focused                   bool
	publishedKnown, publishedFocused bool
	closed                           bool
}
type activationWake struct{ state *windowActivation }
type activationFIFO struct {
	events.Event
	state *windowActivation
}

func (*activationFIFO) Ready() bool                     { return true }
func (*activationFIFO) Merge(events.DeferredEvent) bool { return false }
func (e *activationFIFO) FrameTraceEvent() events.Event { return e.Event }

func (v *windowActivation) observe(focused bool) {
	if v == nil {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed || v.known && v.focused == focused {
		return
	}
	v.known, v.focused = true, focused
	v.queueLocked()
}
func (v *windowActivation) bind(source *events.Source) error {
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
func (v *windowActivation) queueLocked() {
	if !v.known || v.source == nil || v.pending != nil {
		return
	}
	wake := &activationWake{state: v}
	e := &events.CustomEvent{}
	e.Typ = events.Custom
	e.Data = wake
	e.Init()
	v.pending = wake
	_ = v.source.Deque.SendDeferred(func() (events.DeferredEvent, error) { return &activationFIFO{Event: e, state: v}, nil })
}
func (v *windowActivation) consume(source *events.Source, wake *activationWake) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed || source == nil || source != v.source || wake == nil || wake != v.pending {
		return false
	}
	v.pending = nil
	changed := !v.publishedKnown || v.publishedFocused != v.focused
	v.publishedKnown, v.publishedFocused = v.known, v.focused
	return changed
}
func (v *windowActivation) active() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return !v.publishedKnown || v.publishedFocused
}
func (v *windowActivation) close() {
	if v == nil {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return
	}
	v.closed = true
	if source := v.source; source != nil {
		// Queue matchers never acquire v.mu; delivery releases the deque first.
		_ = source.Deque.UpdateDeferred(func() error { return nil }, func(e events.DeferredEvent) bool { node, ok := e.(*activationFIFO); return ok && node.state == v })
	}
	v.source = nil
	v.pending = nil
}
func bindWindowActivation(w *nativeWindow) error {
	b := windowAccessibilityBridge(w)
	if b == nil {
		return nil
	}
	return bindWindowActivationState(w, b.activation)
}
func bindWindowActivationState(w *nativeWindow, v *windowActivation) error {
	if w == nil || w.host == nil || w.body == nil || v == nil {
		return ErrDesktopSession
	}
	if previous := w.activation.Load(); previous != nil && previous != v {
		return ErrDesktopSession
	}
	if err := v.bind(w.body.SystemEvents()); err != nil {
		return err
	}
	w.activation.Store(v)
	if w.host.Property(windowActivationProperty) != v {
		w.host.SetProperty(windowActivationProperty, v)
	}
	return nil
}
func handleWindowActivation(w *nativeWindow, e events.Event) bool {
	ce, ok := e.(*events.CustomEvent)
	if !ok {
		return false
	}
	wake, ok := ce.Data.(*activationWake)
	if !ok || wake == nil || wake.state == nil {
		return false
	}
	e.SetHandled()
	if w != nil && w.host != nil && w.body != nil && !w.closing.Load() && w.window.Ref().Context().Err() == nil && w.activation.Load() == wake.state && wake.state.consume(w.body.SystemEvents(), wake) {
		w.desktop.invalidate(w)
		w.host.Restyle()
	}
	return true
}
func sourceListHasActiveFocus(r *record) bool {
	return r.widget.AsWidget().ContainsFocus() && nativeWindowActive(r)
}
func nativeWindowActive(r *record) bool {
	for node := r.widget.AsTree().Parent; !tree.IsNil(node); node = node.AsTree().Parent {
		if v, ok := node.AsTree().Property(windowActivationProperty).(*windowActivation); ok {
			return v.active()
		}
	}
	return true // offscreen/other platforms retain their established focus behavior
}
