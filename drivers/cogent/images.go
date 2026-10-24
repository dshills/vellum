package cogent

import (
	"image"
	"image/color"
	"math"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

// The source remains immutable and shared. Paint scales directly into the scene,
// avoiding a widget-owned scaled bitmap and unbounded cover-resize intermediates.
type managedImage struct {
	core.WidgetBase
	source *canvas.TiledImage
	fit    ui.ImageFit
}

func (w *managedImage) Destroy() { w.source = nil; w.WidgetBase.Destroy() }

func (w *managedImage) SizeUp() {
	w.WidgetBase.SizeUp()
	if w.source == nil {
		return
	}
	s := &w.Geom.Size
	source := math32.FromPoint(w.source.Size())
	want := s.Actual.Content
	switch {
	case want.X == 0 && want.Y == 0:
		scale := min(float32(1), 256/max(source.X, source.Y))
		x, y := units.Dp(max(1, source.X*scale)), units.Dp(max(1, source.Y*scale))
		x.ToDots(&w.Styles.UnitContext)
		y.ToDots(&w.Styles.UnitContext)
		want = math32.Vec2(x.Dots, y.Dots)
	case want.X == 0:
		want.X = want.Y * source.X / source.Y
	case want.Y == 0:
		want.Y = want.X * source.Y / source.X
	}
	want = want.Ceil()
	styles.SetClampMaxVector(&want, s.Max)
	s.Actual.Content = want
	s.SetTotalFromContent(&s.Actual)
}

func imagePlacement(source, box math32.Vector2, fit ui.ImageFit) (offset, size math32.Vector2) {
	if fit == ui.ImageStretch {
		return math32.Vector2{}, box
	}
	scale := min(box.X/source.X, box.Y/source.Y)
	if fit == ui.ImageCover {
		scale = max(box.X/source.X, box.Y/source.Y)
	}
	size = source.MulScalar(scale)
	return box.Sub(size).MulScalar(0.5), size
}

// imageViewportMask clips only visible content and composes the painter's
// existing destination mask. Coordinates remain global, including scrolling.
type imageViewportMask struct {
	bounds  image.Rectangle
	parent  image.Image
	parentP image.Point
	alpha   uint16
}

func (m imageViewportMask) ColorModel() color.Model { return color.Alpha16Model }
func (m imageViewportMask) Bounds() image.Rectangle { return m.bounds }
func (m imageViewportMask) At(x, y int) color.Color {
	if !image.Pt(x, y).In(m.bounds) {
		return color.Alpha16{}
	}
	a := uint32(m.alpha)
	if m.parent != nil {
		_, _, _, previous := m.parent.At(x+m.parentP.X, y+m.parentP.Y).RGBA()
		a = a * previous / 65535
	}
	return color.Alpha16{A: uint16(a)}
}

func (w *managedImage) Render() {
	w.WidgetBase.Render()
	if w.source == nil || w.Geom.ContentBBox.Empty() {
		return
	}
	box := w.Geom.Size.Actual.Content
	offset, size := imagePlacement(math32.FromPoint(w.source.Size()), box, w.fit)
	position := w.Geom.Pos.Content.Add(offset)
	for _, value := range []float32{position.X, position.Y, size.X, size.Y} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return
		}
	}
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	paint := &w.Scene.Painter
	operation := paint.DrawImageScaled(w.source, w.source, position.X, position.Y, size.X, size.Y)
	if operation == nil {
		return
	}
	alpha := uint16(max(0, min(w.Styles.Opacity, 1)) * 65535)
	operation.Mask = imageViewportMask{bounds: w.Geom.ContentBBox, parent: operation.Mask, parentP: operation.MaskPos, alpha: alpha}
	operation.MaskPos = image.Point{}
}
