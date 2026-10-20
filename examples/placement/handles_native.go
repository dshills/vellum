package main

import (
	"image/color"
	"math"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

const handleSpriteName = "vellum.placement-app.corner-handles"

// The mount owns one presentation sprite and drops it during native teardown.
// The sprite has no event box: capture still belongs to the managed canvas.
type editorMount struct {
	core.Frame
	handleStage   *core.Stage
	handleSprite  *core.Sprite
	restoreQueued bool
}

func (w *editorMount) clearHandles() {
	if w.handleStage != nil {
		w.handleStage.Sprites.Delete(handleSpriteName)
	}
	w.handleStage, w.handleSprite = nil, nil
}
func (w *editorMount) Destroy() {
	w.clearHandles()
	w.Frame.Destroy()
}

// Core resets sprites after restyling a resized scene. Reinstall after that
// pass, from the last detached snapshot, even if the app model is unchanged.
func (w *editorMount) deferHandles(target *core.WidgetBase) {
	if w.restoreQueued {
		return
	}
	w.restoreQueued = true
	target.Defer(func() {
		w.restoreQueued = false
		if w.This == nil || w.handleStage == nil || w.handleSprite == nil {
			return
		}
		w.handleStage.Sprites.Add(w.handleSprite)
		if w.handleStage.Scene != nil {
			w.handleStage.Scene.NeedsRender()
		}
	})
}

func (e *canvasEditor) publishHandles(target *core.WidgetBase, mount *editorMount) {
	mount.clearHandles()
	if target.Scene == nil || target.Scene.Stage == nil || target.Scene.Stage.Main == nil {
		return
	}
	stage := target.Scene.Stage.Main

	ids := e.model.selectedIDs()
	if len(ids) == 0 || e.model.fileBusy() || e.model.prompt != "" || e.model.palette.State().Active {
		return
	}
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	objects := make([]canvas.Rect, 0, len(ids))
	for _, o := range e.model.scene.Objects {
		if wanted[o.ID] {
			objects = append(objects, o)
		}
	}
	// Owner-detached values: native Draw reads no mutable application model.
	view, life := e.model.scene.View, e.ref.Context()
	sprite := core.NewSprite(handleSpriteName, func(pc *paint.Painter) {
		if life.Err() != nil || target.This == nil || !target.IsVisible() || target.IsDisabled() {
			return
		}
		g, err := editGeometry(target, view)
		if err != nil {
			return
		}
		if len(objects) == 1 {
			drawCornerHandles(pc, objects[0], g)
		} else {
			for _, object := range objects {
				drawSelectionOutline(pc, object, g)
			}
			if bounds, ok := selectionBounds(objects, ids); ok {
				drawSelectionOutline(pc, bounds, g)
				drawCornerHandles(pc, bounds, g)
			}
		}
	})
	sprite.Active = true
	mount.handleStage, mount.handleSprite = stage, sprite
	stage.Sprites.Add(sprite)
}

func drawCornerHandles(pc *paint.Painter, object canvas.Rect, g placement.Geometry) {
	fill := func(center canvas.Point, half float64, col color.NRGBA) {
		half *= g.Scale
		left, top := math.Max(g.Clip.Min.X, center.X-half), math.Max(g.Clip.Min.Y, center.Y-half)
		right, bottom := math.Min(g.Clip.Max.X, center.X+half), math.Min(g.Clip.Max.Y, center.Y+half)
		if right > left && bottom > top {
			pc.FillBox(math32.Vec2(float32(left), float32(top)), math32.Vec2(float32(right-left), float32(bottom-top)), colors.Uniform(col))
		}
	}
	for _, handle := range cornerHandles(object, g) {
		if !inCanvas(g, handle.center) {
			continue
		}
		fill(handle.center, 5, color.NRGBA{R: 25, G: 90, B: 180, A: 255})
		fill(handle.center, 3, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	}
}

func drawSelectionOutline(pc *paint.Painter, o canvas.Rect, g placement.Geometry) {
	p := g.View.Screen(canvas.Point{X: o.X, Y: o.Y})
	x, y := g.Origin.X+p.X*g.Scale, g.Origin.Y+p.Y*g.Scale
	right, bottom := x+o.Width*g.View.Zoom*g.Scale, y+o.Height*g.View.Zoom*g.Scale
	width := 2 * g.Scale
	fill := func(left, top, r, b float64) {
		left, top = math.Max(left, g.Clip.Min.X), math.Max(top, g.Clip.Min.Y)
		r, b = math.Min(r, g.Clip.Max.X), math.Min(b, g.Clip.Max.Y)
		if r > left && b > top {
			pc.FillBox(math32.Vec2(float32(left), float32(top)), math32.Vec2(float32(r-left), float32(b-top)), colors.Uniform(color.NRGBA{R: 25, G: 90, B: 180, A: 255}))
		}
	}
	fill(x-width, y-width, right+width, y)
	fill(x-width, bottom, right+width, bottom+width)
	fill(x-width, y, x, bottom)
	fill(right, y, right+width, bottom)
}
