package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
)

func styleSheetHost(r *record, s *styles.Style) {
	if !r.node.Sheet {
		return
	}
	s.Grow.Set(0, 0)
	applyStyle(s, r.node, r.theme)
}
func bindSheetDefault(r *record, body *core.Body) {
	if !r.node.Sheet {
		return
	}
	scope := r.control.dialog.tabScope.Load()
	body.Scene.OnFinal(events.KeyChord, func(e events.Event) {
		if e.IsHandled() || e.Modifiers() != 0 || e.KeyCode() != key.CodeReturnEnter {
			return
		}
		target := sheetRecord(r, r.node.DialogDefaultID)
		if target == nil || target.node.Activate == nil || controlDisabled(target) {
			return
		}
		e.SetHandled()
		postControl(r, func() {
			if scope == nil || scope.closed.Load() || r.control.retired || !r.node.Open || !r.node.Sheet || controlDisabled(r) || sheetRecord(r, r.node.DialogDefaultID) != target || controlDisabled(target) || target.node.Activate == nil {
				return
			}
			target.node.Activate()
			r.control.a.Invalidate(r.control.window)
		})
	})
}
func restoreSheetFocus(r *record) {
	if !r.node.Sheet || r.node.DialogReturnFocusID == "" || r.widget.AsWidget().Scene == nil {
		return
	}
	r.widget.AsWidget().Scene.WidgetWalkDown(func(w core.Widget, b *core.WidgetBase) bool {
		sem, ok := b.Property(SemanticsProperty).(Semantics)
		if ok && sem.ID == r.node.DialogReturnFocusID && !sem.Disabled && sem.Role == "button" && b.IsVisible() && !b.Geom.TotalBBox.Empty() {
			b.SetFocus()
			return false
		}
		return true
	})
}

func sheetRecord(r *record, id string) *record {
	if r == nil {
		return nil
	}
	if r.node.ID == id {
		return r
	}
	for _, child := range r.children {
		if found := sheetRecord(child, id); found != nil {
			return found
		}
	}
	return nil
}
