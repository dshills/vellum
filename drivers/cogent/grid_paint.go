package cogent

import (
	"image"
	"image/color"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
)

// The reference switch is private evidence plumbing. Applications use the same
// managed grid API; individual Text repaints always retain their original box.
const gridReferenceBoxPaintProperty = "vellum.driver.reference_grid_box_paint"

// A pane paints its full opaque background before its disjoint cells. Within
// this full pane traversal, an undecorated passive cell's identical opaque box
// adds no pixels. Avoid constructing and rasterizing those repeated boxes, while
// retaining normal text, clipping, selected/cursor paint and widget ownership.
// Every temporary flag is restored even when a child panics or exits its goroutine.
func (p *gridPane) elideCoveredCellBoxes() func() {
	unchanged := func() {}
	g := p.binding
	if p.This != p || g == nil || !g.live() || !g.projectionReady || g.editing != nil || p.Scene == nil || p.Scene.Painter.State == nil {
		return unchanged
	}
	if reference, _ := p.Scene.Property(gridReferenceBoxPaintProperty).(bool); reference {
		return unchanged
	}
	if experimental, _ := p.Scene.Property(gridScrollReuseProperty).(bool); experimental {
		return unchanged // keep the separate verified-scroll protocol unchanged
	}
	if !plainGridBox(&p.Styles) || !p.Styles.RenderBox || !integerGridBox(&p.WidgetBase) {
		return unchanged
	}
	parent, ok := p.Parent.(core.Widget)
	if !ok {
		return unchanged
	}
	background := p.Styles.Background
	if background == nil {
		background = parent.AsWidget().Styles.ActualBackground
	}
	rgba, ok := opaqueGridUniform(background)
	if !ok {
		return unchanged
	}
	// An antialiased/rounded/masked parent edge can accumulate coverage when
	// painted again. Only exact integer rectangular clips justify elision.
	for _, context := range p.Scene.Painter.Stack {
		if context == nil || context.Cumulative != math32.Identity2() || context.Mask != nil || len(context.Bounds.Path) != 0 || len(context.ClipPath) != 0 || !sides.AreZero(context.Bounds.Radius.Sides) || context.Bounds.Rect != math32.B2FromRect(context.Bounds.Rect.ToRect()) {
			return unchanged
		}
	}
	managed := g.cells
	if p.part < 2 {
		managed = g.headers
	}
	cursor := 0
	var skipped []*core.Text
	var row image.Rectangle
	endX := 0
	for _, child := range p.Children {
		for cursor < len(managed) && managed[cursor].Parent != p {
			cursor++
		}
		tx, ok := child.(*core.Text)
		if !ok || cursor == len(managed) || managed[cursor] != tx || tx.This != tx || tx.Parts != nil || len(tx.Children) != 0 {
			return unchanged // foreign/reordered content retains ordinary painting
		}
		cursor++
		if !tx.IsVisible() || tx.Geom.TotalBBox.Empty() {
			continue
		}
		box := tx.Geom.TotalBBox
		if row.Empty() || box.Min.Y >= row.Max.Y {
			row, endX = box, box.Min.X
		} else if box.Min.Y != row.Min.Y || box.Max.Y != row.Max.Y {
			return unchanged
		}
		if box.Min.X < endX {
			return unchanged
		}
		endX = box.Max.X
		if !plainGridBox(&tx.Styles) || !tx.Styles.RenderBox || !integerGridBox(&tx.WidgetBase) {
			continue
		}
		cellBackground := tx.Styles.Background
		if cellBackground == nil {
			cellBackground = background
		}
		cellRGBA, uniform := opaqueGridUniform(cellBackground)
		if !uniform || cellRGBA != rgba {
			continue
		}
		skipped = append(skipped, tx)
	}
	// Admission completes before changing any widget. Later callers, including
	// isolated widget repaints, must still clear their old text normally.
	for _, tx := range skipped {
		tx.Styles.RenderBox = false
	}
	return func() {
		for _, tx := range skipped {
			tx.Styles.RenderBox = true
		}
	}
}

func plainGridBox(s *styles.Style) bool {
	return s.Opacity == 1 && s.StateLayer == 0 && len(s.BoxShadow) == 0 && len(s.MaxBoxShadow) == 0 &&
		sides.AreZero(s.TotalMargin().Sides) && sides.AreZero(s.Border.Width.Dots().Sides) && sides.AreZero(s.Border.Offset.Dots().Sides) && sides.AreZero(s.Border.Radius.Dots().Sides)
}

func integerGridBox(w *core.WidgetBase) bool {
	box := w.Geom.TotalBBox
	return w.Geom.Pos.Total == math32.FromPoint(box.Min) && w.Geom.Size.Actual.Total == math32.FromPoint(box.Size())
}

// Do not invoke arbitrary application color methods during an optimization proof.
func opaqueGridUniform(source image.Image) ([4]uint32, bool) {
	var rgba [4]uint32
	uniform, ok := source.(*image.Uniform)
	if !ok || uniform == nil {
		return rgba, false
	}
	switch uniform.C.(type) {
	case color.RGBA, color.RGBA64, color.NRGBA, color.NRGBA64, color.Gray, color.Gray16, color.Alpha, color.Alpha16:
	default:
		return rgba, false
	}
	rgba[0], rgba[1], rgba[2], rgba[3] = uniform.C.RGBA()
	return rgba, rgba[3] == 65535
}

func (p *gridPane) renderPaneFrame() {
	if p.Scene != nil {
		reference, _ := p.Scene.Property(gridReferenceBoxPaintProperty).(bool)
		experimental, _ := p.Scene.Property(gridScrollReuseProperty).(bool)
		if !reference && !experimental && p.Frame.RenderPassiveTextBatch() {
			return
		}
	}
	p.Frame.RenderWidget()
}
