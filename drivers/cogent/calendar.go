package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
)

func bindCalendarCell(r *record, button *core.Button) {
	button.OnFirst(events.KeyChord, func(e events.Event) {
		if e.IsHandled() || r.node.CalendarCell == nil || controlDisabled(r) || !button.IsVisible() || e.Modifiers() & ^((1<<key.CapsLock)|(1<<key.NumLock)) != 0 {
			return
		}
		delta := 0
		switch e.KeyCode() {
		case key.CodeLeftArrow:
			delta = -1
		case key.CodeRightArrow:
			delta = 1
		case key.CodeUpArrow:
			delta = -7
		case key.CodeDownArrow:
			delta = 7
		default:
			return
		}
		e.SetHandled()
		id := r.node.CalendarCell.Neighbor(delta)
		if id == "" || button.Scene == nil {
			return
		}
		button.Scene.WidgetWalkDown(func(w core.Widget, b *core.WidgetBase) bool {
			if sem, ok := b.Property(SemanticsProperty).(Semantics); ok && sem.ID == id && !sem.Disabled && b.IsVisible() {
				b.SetFocus()
				b.ScrollToThis()
				postControl(r, func() { _ = r.control.a.Invalidate(r.control.window) })
				return false
			}
			return true
		})
	})
}
