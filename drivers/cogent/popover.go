package cogent

import (
	"cogentcore.org/core/colors"
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
	if d == nil || d.host == nil || d.body == nil || d.body.Scene == nil || anchor == nil {
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
	desired := d.host.Geom.Size.Actual.Total.ToPointCeil()
	if desired.X > 0 && desired.Y > 0 {
		size = image.Pt(min(desired.X, bounds.Dx()), min(desired.Y, bounds.Dy()))
	}
	pos := image.Pt(a.Min.X, a.Max.Y+4)
	pos.X = max(bounds.Min.X, min(pos.X, bounds.Max.X-size.X))
	pos.Y = max(bounds.Min.Y, min(pos.Y, bounds.Max.Y-size.Y))
	if scene.SceneGeom.Pos != pos || scene.SceneGeom.Size != size {
		scene.Resize(math32.Geom2DInt{Pos: pos, Size: size})
		scene.LayoutScene()
	}
}
func stylePopoverHost(r *record, s *styles.Style) {
	if r.node.PopoverAnchor == "" {
		return
	}
	s.Padding.Set(units.Dp(12))
	s.Gap.Set(units.Dp(6))
	s.Grow.Set(0, 0)
	s.Min.X = units.Dp(240)
	applyStyle(s, r.node, r.theme)
}

// Keep managed Dialog lifetime/focus semantics while removing its generic
// move handle, outer padding, minimum-size allowance and title side effect.
func preparePopoverSurface(r *record, body *core.Body) {
	body.FinalStyler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Padding.Zero(); s.Margin.Zero() })
	body.Scene.FinalStyler(func(s *styles.Style) {
		s.Padding.Zero()
		s.Grow.Set(1, 1)
		s.Gap.Zero()
		s.Border.Radius.Set(units.Dp(8))
		value, _ := ui.ResolveStyle(r.node, r.theme)
		if value.Background.Set {
			s.Background = colors.Uniform(value.Background.Value)
		}
		if value.Foreground.Set {
			s.Color = colors.Uniform(value.Foreground.Value)
		}
	})
}
func finishPopoverSurface(body *core.Body) {
	if body.Scene.Parts != nil {
		body.Scene.Parts.Destroy()
		body.Scene.Parts = nil
	}
	body.Scene.Update()
	body.Scene.LayoutScene()
}
