package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/paint/render"
	"image"
	"math"
)

const gridScrollReuseProperty = "vellum.driver.experimental_grid_scroll_reuse"

func setGridScrollReuse(scene *core.Scene, enabled bool) {
	scene.SetProperty(gridScrollReuseProperty, enabled)
}

// Keep each pane at its original painter position. The renderer checks the
// actual paint operations before reusing pixels; model generations, selection,
// fonts and theme identities alone do not establish translation equivalence.
func (p *gridPane) RenderWidget() {
	if p.renderCellSurface() {
		return
	}
	restore := p.elideCoveredCellBoxes()
	defer restore()
	if p.Scene == nil || p.Scene.Painter.State == nil {
		p.renderPaneFrame()
		return
	}
	enabled, _ := p.Scene.Property(gridScrollReuseProperty).(bool)
	if !enabled {
		p.renderPaneFrame()
		return
	}
	items := &p.Scene.Painter.Render
	start := len(*items)
	p.renderPaneFrame()
	g := p.binding
	if g == nil || g.closed.Load() || !g.projectionReady || g.geometryRefused || start >= len(*items) {
		return
	}
	push, ok := (*items)[start].(*render.ContextPush)
	if !ok || push == nil {
		return
	}
	bounds := p.Geom.TotalBBox
	scale := p.Styles.UnitContext.Px(1)
	if bounds.Empty() || !validGridScale(scale) {
		return
	}
	viewport := g.projection.State.Viewport
	offset := image.Point{}
	if p.part == 1 || p.part == 3 {
		offset.X = int(math.Round(float64(viewport.ScrollX) * float64(scale)))
	}
	if p.part == 2 || p.part == 3 {
		offset.Y = int(math.Round(float64(viewport.ScrollY) * float64(scale)))
	}
	// Incarnation is a nonwrapping lifetime ID. The constant epoch only identifies
	// this checked protocol: every overlap independently compares detached paints.
	push.Scroll = &render.ScrollSurface{ID: g.incarnation, Part: p.part, Epoch: 1, Position: offset, Bounds: bounds, VerifyPaint: true}
}
