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

	i := e.model.selectedIndex()
	if i < 0 || e.model.fileBusy() || e.model.prompt != "" || e.model.palette.State().Active {
		return
	}
	// Sync is under the UI owner. Draw reads detached values plus native layout
	// under the render lock, never mutable application model state.
	object, view, life := e.model.scene.Objects[i], e.model.scene.View, e.ref.Context()
	sprite := core.NewSprite(handleSpriteName, func(pc *paint.Painter) {
		if life.Err() != nil || target.This == nil || !target.IsVisible() || target.IsDisabled() {
			return
		}
		g, err := editGeometry(target, view)
		if err != nil {
			return
		}
		drawCornerHandles(pc, object, g)
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
