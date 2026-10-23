package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
	"image"
)

func popoverAnchor(r *record) core.Widget {
	scene := r.widget.AsWidget().Scene
	if scene == nil {
		return nil
	}
	var found core.Widget
	scene.WidgetWalkDown(func(w core.Widget, b *core.WidgetBase) bool {
		sem, ok := b.Property(SemanticsProperty).(Semantics)
		if ok && sem.ID == r.node.PopoverAnchor && sem.Role == ui.Button && !sem.Disabled && b.IsVisible() && !b.Geom.TotalBBox.Empty() {
			found = w
			return false
		}
		return true
	})
	return found
}
func positionPopover(r *record, anchor core.Widget) {
	d := r.control.dialog
	if d.body == nil || d.body.Scene == nil || anchor == nil {
		return
	}
	scene := d.body.Scene
	parent := r.widget.AsWidget().Scene
	if parent == nil {
		return
	}
	bounds := parent.SceneGeom.Bounds()
	a := anchor.AsWidget().Geom.TotalBBox
	size := scene.SceneGeom.Size
	pos := image.Pt(a.Min.X, a.Max.Y+4)
	pos.X = max(bounds.Min.X, min(pos.X, bounds.Max.X-size.X))
	pos.Y = max(bounds.Min.Y, min(pos.Y, bounds.Max.Y-size.Y))
	if scene.SceneGeom.Pos != pos {
		scene.Resize(math32.Geom2DInt{Pos: pos, Size: size})
	}
}
func stylePopoverHost(r *record, s *styles.Style) {
	if r.node.PopoverAnchor == "" {
		return
	}
	s.Padding.Set(units.Dp(12))
	s.Gap.Set(units.Dp(6))
	s.Grow.Set(0, 0)
	applyStyle(s, r.node, r.theme)
}
