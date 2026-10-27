package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"image"
)

type GPUEventKind string

const (
	GPUPointerDown  GPUEventKind = "pointer-down"
	GPUPointerMove  GPUEventKind = "pointer-move"
	GPUPointerUp    GPUEventKind = "pointer-up"
	GPUPointerLeave GPUEventKind = "pointer-leave"
	GPUScroll       GPUEventKind = "scroll"
	GPUKey          GPUEventKind = "key"
	GPUCancel       GPUEventKind = "cancel"
	GPUContextMenu  GPUEventKind = "context-menu"
)

type GPUButton uint8

const (
	GPUButtonNone GPUButton = iota
	GPUButtonPrimary
	GPUButtonMiddle
	GPUButtonSecondary
)

type GPUModifiers uint8

const (
	GPUShift GPUModifiers = 1 << iota
	GPUControl
	GPUOption
	GPUCommand
)

// GPUEvent is a detached content-local physical-pixel input snapshot. Size is
// the full content size, Inside uses its visible clip. Scale converts dp to pixels.
// Primary sequences use existing same-window capture and consume their release.
// Secondary/middle events are delivered while targeted; they do not acquire
// cross-window or alternate-button capture. Use Shift/Option+primary for captured
// pan/orbit. Cancel is delivered while the owner/window is live; shutdown retires
// capture without running application callbacks after closure.
type GPUEvent struct {
	Kind GPUEventKind
	// Sequence identifies a primary gesture; ignore a queued Cancel for an older sequence.
	Sequence              uint64
	Position, Delta, Size image.Point
	Button                GPUButton
	Modifiers             GPUModifiers
	ScrollX, ScrollY      float32
	Key                   string
	Scale                 float64
	Inside                bool
}

func gpuModifiers(m key.Modifiers) GPUModifiers {
	var out GPUModifiers
	for _, p := range []struct {
		native key.Modifiers
		flag   GPUModifiers
	}{{key.Shift, GPUShift}, {key.Control, GPUControl}, {key.Alt, GPUOption}, {key.Meta, GPUCommand}} {
		if m.HasFlag(p.native) {
			out |= p.flag
		}
	}
	return out
}
func (w *gpuViewWidget) eventData(kind GPUEventKind, e events.Event) GPUEvent {
	p := e.Pos()
	v := GPUEvent{Kind: kind, Sequence: w.inputSequence, Position: p.Sub(w.Geom.Pos.Content.ToPoint()), Delta: e.PrevDelta(), Size: w.Geom.Size.Actual.Content.ToPoint(), Button: GPUButton(e.MouseButton()), Modifiers: gpuModifiers(e.Modifiers()), Scale: placementScale(&w.WidgetBase), Inside: p.In(w.Geom.ContentBBox)}
	if s, ok := e.(*events.MouseScroll); ok {
		v.ScrollX, v.ScrollY = s.Delta.X, s.Delta.Y
	}
	if kind == GPUKey {
		v.Key = string(e.KeyChord())
	}
	return v
}
func (w *gpuViewWidget) emitInput(v GPUEvent) bool {
	handled := false
	w.ownerEvent(func() {
		if w.onInput != nil {
			handled = w.onInput(v)
		}
	})
	return handled
}
func (w *gpuViewWidget) endGPUCapture(cancel bool) {
	token := w.inputCapture
	w.inputCapture = 0
	if token == 0 {
		return
	}
	if w.Scene != nil {
		w.Scene.Events.ReleasePointer(token)
	}
	if cancel {
		v := w.lastInput
		v.Kind = GPUCancel
		// Focus/popup transitions may run inside application ownership. Never nest Do.
		_ = w.application.PostWindow(w.windowRef, func() {
			if !w.retired.Load() && w.owner != nil && w.owner.widget == w && w.onInput != nil {
				w.onInput(v)
			}
		})
	}
}
func (w *gpuViewWidget) installGPUInput() {
	w.OnFirst(events.MouseDown, func(e events.Event) {
		if e.MouseButton() == events.Left {
			if w.inputSequence == ^uint64(0) {
				e.SetHandled()
				return
			}
			w.inputSequence++
		}
		v := w.eventData(GPUPointerDown, e)
		w.lastInput = v
		if e.MouseButton() == events.Left && w.Scene != nil {
			token, err := w.Scene.Events.CapturePointer(w, e, func(e events.Event) {
				switch e.Type() {
				case events.MouseMove, events.MouseDrag:
					v := w.eventData(GPUPointerMove, e)
					w.lastInput = v
					w.emitInput(v)
				case events.MouseUp:
					if e.MouseButton() == events.Left {
						v := w.eventData(GPUPointerUp, e)
						w.endGPUCapture(false)
						if !v.Inside {
							v.Kind = GPUCancel
						}
						w.emitInput(v)
					}
				case events.KeyChord:
					w.emitInput(w.eventData(GPUKey, e))
				case events.Scroll:
					w.emitInput(w.eventData(GPUScroll, e))
				}
			}, func() { w.endGPUCapture(true) })
			e.SetHandled()
			if err != nil {
				w.Scene.Events.ConsumePointerRelease()
				return
			}
			w.inputCapture = token
			w.SetFocus()
			w.emitInput(v)
			return
		}
		if w.emitInput(v) {
			e.SetHandled()
		}
	})
	for typ, kind := range map[events.Types]GPUEventKind{events.MouseMove: GPUPointerMove, events.MouseDrag: GPUPointerMove, events.MouseUp: GPUPointerUp, events.MouseLeave: GPUPointerLeave, events.Scroll: GPUScroll, events.KeyChord: GPUKey, events.ContextMenu: GPUContextMenu} {
		w.OnFirst(typ, func(e events.Event) {
			if !gpuTargetedInputAllowed(typ, e.MouseButton(), w.inputCapture) {
				return
			}
			if w.emitInput(w.eventData(kind, e)) {
				e.SetHandled()
			}
		})
	}
	w.OnFirst(events.FocusLost, func(events.Event) { w.endGPUCapture(true) })
}

// Captured events are intercepted before normal routing. This guard also
// refuses direct/synthetic targeted delivery during capture and primary releases
// from sequences that did not begin in this view.
func gpuTargetedInputAllowed(typ events.Types, button events.Buttons, capture uint64) bool {
	if capture != 0 {
		return false
	}
	return !((typ == events.MouseUp || typ == events.MouseDrag) && button == events.Left)
}
