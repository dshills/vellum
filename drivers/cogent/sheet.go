package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
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

// Center the measured card inside its originating client area. Use its retained
// content size, rather than the generic dialog minimum/title allowance.
func positionSheet(r *record) {
	d := r.control.dialog
	if d == nil || d.body == nil || d.body.Scene == nil || r.widget.AsWidget().Scene == nil {
		return
	}
	scene, parent := d.body.Scene, r.widget.AsWidget().Scene
	bounds := parent.SceneGeom.Bounds()
	size := d.host.Geom.Size.Actual.Total.Add(scene.Geom.Size.Space).Add(d.body.Geom.Size.Space).ToPointCeil()
	size.X = max(1, min(size.X, bounds.Dx()))
	size.Y = max(1, min(size.Y, bounds.Dy()))
	pos := bounds.Min.Add(bounds.Size().Sub(size).Div(2))
	if scene.SceneGeom.Pos != pos || scene.SceneGeom.Size != size {
		scene.Resize(math32.Geom2DInt{Pos: pos, Size: size})
		scene.LayoutScene()
	}
}

func prepareSheetSurface(r *record, body *core.Body) {
	preparePopoverSurface(r, body)
	body.FinalStyler(func(s *styles.Style) { s.Overflow.Set(styles.OverflowHidden) })
	body.Scene.FinalStyler(func(s *styles.Style) { s.Overflow.Set(styles.OverflowHidden) })
}
