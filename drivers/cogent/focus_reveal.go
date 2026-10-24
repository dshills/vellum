package cogent

import (
	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
)

// One immutable intent per native window. It is consumed before the next host paint,
// after layout; atomics also permit shutdown to release the reference safely.
type focusRevealRequest struct {
	request ui.FocusRequest
	id      string
	target  core.Widget
}

// applyWindowFocus runs on the originating native owner with rootMu held.
func applyWindowFocus(w *nativeWindow, request ui.FocusRequest) error {
	if request.ID == "" {
		w.pendingFocusReveal.Store(nil)
	}
	previous := w.lastFocus
	if err := applyFocus(w.root, request, &w.lastFocus); err != nil {
		w.pendingFocusReveal.Store(nil)
		return err
	}
	if request.ID == "" || previous == w.lastFocus {
		return nil
	}
	w.pendingFocusReveal.Store(nil)
	id := accessibilityRequestedFocusID(w.root, request.ID)
	target, ok := accessibilityResolveFocus(w.root, id)
	if !ok || accessibilityTreeNil(target.widget) || w.host == nil || target.widget.AsWidget().Scene != w.host.Scene {
		return nil
	}
	w.pendingFocusReveal.Store(&focusRevealRequest{request: request, id: id, target: target.widget})
	if w.closing.Load() || w.window.Ref().Context().Err() != nil {
		w.pendingFocusReveal.Store(nil)
	} else {
		// Focus-only requests may not otherwise dirty native layout or paint.
		w.host.NeedsRender()
	}
	return nil
}

// Layout can move an already-focused field after SetFocus has used its old
// coordinates. Scroll once before child painting with completed positions. Do
// not change focus or replay during later manual scrolling. Scene/identity checks
// preserve the originating attachment and cancel replaced or hidden targets.
func finishWindowFocusReveal(w *nativeWindow, scene *core.Scene) {
	if w == nil {
		return
	}
	pending := w.pendingFocusReveal.Swap(nil)
	if pending == nil {
		return
	}
	w.rootMu.Lock()
	defer w.rootMu.Unlock()
	if w.closing.Load() || w.window.Ref().Context().Err() != nil || pending.request != w.lastFocus {
		return
	}
	target, ok := accessibilityResolveFocus(w.root, pending.id)
	if !ok || accessibilityTreeNil(target.widget) || target.widget != pending.target || !target.focused {
		return
	}
	widget := target.widget.AsWidget()
	if widget.Scene != scene {
		return
	}
	widget.ScrollToThis()
}
