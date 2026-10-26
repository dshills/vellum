package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func bindPaletteInput(r *record) {
	field, ok := r.widget.(*core.TextField)
	if !ok || r.node.Kind != ui.TextField {
		return
	}
	update := func(e events.Event, delta int, accept bool) bool {
		input := r.control.paletteInput.Load()
		if input == nil || r.scope.Context().Err() != nil {
			return false
		}
		if _, _, composing := field.NativeCompositionRange(); composing {
			return false
		}
		text := field.EditingState().Text
		handled := false
		_ = r.control.a.Do(func() {
			if r.control.retired || r.control.paletteInput.Load() != input || controlDisabled(r) || !fileDropTargetActive(r, r.widget) {
				return
			}
			input.Change(ui.PaletteKeyEvent{Text: text, Delta: delta, Accept: accept})
			r.control.a.Invalidate(r.control.window)
			handled = true
		})
		return handled
	}
	field.OnInput(func(e events.Event) { update(e, 0, false) })
	field.OnFirst(events.KeyChord, func(e events.Event) {
		if e.IsHandled() || key.HasAnyModifier(e.Modifiers(), key.Control, key.Meta, key.Alt, key.Shift) {
			return
		}
		delta, accept := 0, false
		switch e.KeyCode() {
		case key.CodeUpArrow:
			delta = -1
		case key.CodeDownArrow:
			delta = 1
		case key.CodeReturnEnter, key.CodeKeypadEnter:
			accept = true
		default:
			return
		}
		if update(e, delta, accept) {
			e.SetHandled()
		}
	})
}
