package cogent

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"sync/atomic"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

const toolbarDragActiveProperty = "vellum.toolbar-drag-active"

// Suppress target-row help as well as source help. A tooltip is a native popup
// and would retire same-window capture during an otherwise valid held drag.
func bindToolbarDragTooltipGuard(widget core.Widget) {
	for _, typ := range []events.Types{events.LongHoverStart, events.LongPressStart} {
		widget.AsWidget().OnFirst(typ, func(e events.Event) {
			for p := widget.AsTree(); p != nil; {
				if g, ok := p.Property(toolbarDragActiveProperty).(*toolbarDragGesture); ok && g.scroll.Scene != nil && g.scroll.Scene.Events.PointerCaptured(g.token) {
					e.SetHandled()
					return
				}
				if tree.IsNil(p.Parent) {
					break
				}
				p = p.Parent.AsTree()
			}
		})
	}
}

// Application-owner publication; native routing never reads the customizer.
type toolbarDragScope struct {
	reducedMotion bool
	handle        *ui.ToolbarReorderHandle
	input         *workbenchInputScope
}
type toolbarDragBinding struct {
	r      *record
	button *core.Button
	scope  atomic.Pointer[toolbarDragScope]
	active *toolbarDragGesture // originating render ownership only
}
type toolbarDragGesture struct {
	scope                                            *toolbarDragScope
	scroll                                           *core.Frame
	rows                                             []*core.WidgetBase
	start                                            image.Point
	scale                                            float64
	dragging                                         bool
	token                                            uint64
	stage                                            *core.Stage
	sprite                                           string
	animation                                        *core.Animation
	animationScene                                   *core.Scene
	previewClip                                      image.Rectangle
	lineY, fromY, targetY, elapsed, alpha, fromAlpha float32
	hasLine                                          bool
}

func syncToolbarDrag(r *record) {
	if r.control.toolbarDrag == nil && r.node.ToolbarReorder != nil {
		if button, ok := r.widget.(*core.Button); ok {
			b := &toolbarDragBinding{r: r, button: button}
			r.control.toolbarDrag = b
			button.OnFirst(events.MouseDown, b.press)
			for _, typ := range []events.Types{events.LongHoverStart, events.LongPressStart} {
				button.OnFirst(typ, func(e events.Event) {
					if b.active != nil {
						e.SetHandled()
					}
				})
			}
		}
	}
}
func syncToolbarDragVisibility(r *record, visible bool, scopes []*tabCloseScope) {
	if r.control == nil || r.control.toolbarDrag == nil {
		return
	}
	b := r.control.toolbarDrag
	old := b.scope.Load()
	h := r.node.ToolbarReorder
	if old != nil && old.handle == h && old.reducedMotion == r.theme.ReducedMotion && old.input.visible == visible && old.input.disabled == r.control.disabled && slices.Equal(old.input.scopes, scopes) {
		return
	}
	if old != nil {
		old.input.revoked.Store(true)
	}
	b.end()
	if h == nil {
		b.scope.Store(nil)
		return
	}
	b.scope.Store(&toolbarDragScope{reducedMotion: r.theme.ReducedMotion, handle: h, input: &workbenchInputScope{visible: visible, disabled: r.control.disabled, scopes: slices.Clone(scopes)}})
}
func disposeToolbarDrag(r *record) {
	if b := r.control.toolbarDrag; b != nil {
		if old := b.scope.Swap(nil); old != nil {
			old.input.revoked.Store(true)
		}
		b.end()
	}
}
func (b *toolbarDragBinding) end() {
	g := b.active
	b.active = nil
	if g == nil {
		return
	}
	if g.scroll.Property(toolbarDragActiveProperty) == g {
		g.scroll.DeleteProperty(toolbarDragActiveProperty)
	}
	g.hasLine = false
	g.stopAnimation()
	if g.scroll.Scene != nil {
		g.scroll.Scene.Events.ReleasePointer(g.token)
	}
	if g.stage != nil {
		g.stage.Sprites.Delete(g.sprite)
		g.stage = nil
	}
	// No timers, application callbacks or captured row references survive exit.
}
func (b *toolbarDragBinding) press(e events.Event) {
	if e.Type() != events.MouseDown || e.MouseButton() != events.Left {
		return
	}
	s := b.scope.Load()
	if s == nil || !s.input.allowed() {
		return
	}
	row := core.AsWidget(b.button.Parent)
	if row == nil {
		return
	}
	scroll, ok := row.Parent.(*core.Frame)
	if !ok || scroll.Name != s.handle.ListID() || scroll.Scene == nil {
		return
	}
	g := &toolbarDragGesture{scope: s, scroll: scroll, start: e.Pos(), scale: placementScale(&scroll.WidgetBase), sprite: fmt.Sprintf("vellum.toolbar-reorder.%p", b)}
	if g.scale <= 0 || math.IsNaN(g.scale) || math.IsInf(g.scale, 0) {
		return
	}
	for _, id := range s.handle.RowIDs() {
		w := core.AsWidget(scroll.ChildByName(id))
		if w == nil || w.Scene != scroll.Scene {
			return
		}
		g.rows = append(g.rows, w)
	}
	e.SetHandled()
	token, err := scroll.Scene.Events.CapturePointer(scroll, e, func(event events.Event) { b.event(g, event) }, func() { b.end() })
	if err != nil {
		scroll.Scene.Events.ConsumePointerRelease()
		return
	}
	g.token = token
	b.active = g
	scroll.SetProperty(toolbarDragActiveProperty, g)
	b.button.SetFocusQuiet()
}

// Scoped capture sets e.LocalOff to the originating SceneGeom.Pos before
// delivery. e.Pos() is already scene-local, matching ContentBBox/full row bounds;
// only the detached main-stage sprite adds the window scene origin back.
func (b *toolbarDragBinding) event(g *toolbarDragGesture, e events.Event) {
	if b.active != g {
		return
	}
	if b.scope.Load() != g.scope || !g.scope.input.allowed() {
		b.end()
		return
	}
	if e.Type() == events.MouseUp && e.MouseButton() == events.Left {
		before, _, valid := g.destination(e.Pos())
		commit := g.dragging && valid
		b.end() // revoke capture and preview before queuing the single proposal
		if commit {
			s := g.scope
			postControl(b.r, func() {
				if b.scope.Load() == s && s.input.allowed() && b.r.node.ToolbarReorder == s.handle {
					s.handle.DropBefore(before)
					_ = b.r.control.a.Invalidate(b.r.control.window)
				}
			})
		}
		return
	}
	switch e.Type() {
	case events.MouseMove, events.MouseDrag:
		delta := e.Pos().Sub(g.start)
		if math.Hypot(float64(delta.X), float64(delta.Y)) >= 6*g.scale {
			g.dragging = true
		}
	case events.Scroll:
		if !g.dragging {
			return
		}
		// Captured wheel events scroll only this list; they cannot reach the toolbar.
		if wheel, ok := e.(*events.MouseScroll); ok {
			g.scrollBy(float64(wheel.Delta.Y))
		}
	default:
		return
	}
	if !g.dragging {
		return
	}
	// Moving at an edge scrolls a bounded step. Wheel/trackpad scrolling also works
	// during capture. There is deliberately no background held-pointer timer.
	clip := g.scroll.Geom.ContentBBox
	if e.Type() != events.Scroll && e.Pos().In(clip) {
		if e.Pos().Y < clip.Min.Y+int(18*g.scale) {
			g.scrollBy(-10 * g.scale)
		}
		if e.Pos().Y > clip.Max.Y-int(18*g.scale) {
			g.scrollBy(10 * g.scale)
		}
	}
	g.preview(e.Pos())
	g.scroll.Scene.NeedsRender()
}
func (g *toolbarDragGesture) scrollBy(delta float64) {
	fr := g.scroll
	if !fr.HasScroll[math32.Y] || fr.Scrolls[math32.Y] == nil || math.IsNaN(delta) || math.IsInf(delta, 0) {
		return
	}
	maximum, visible, _ := fr.ScrollValues(math32.Y)
	delta = math.Max(-float64(visible), math.Min(float64(visible), delta))
	next := float32(math.Max(0, math.Min(float64(maximum-visible), float64(fr.Scrolls[math32.Y].Value)+delta)))
	if next == fr.Scrolls[math32.Y].Value {
		return
	}
	fr.Scrolls[math32.Y].SetValue(next)
	fr.ScrollChanged(math32.Y, fr.Scrolls[math32.Y])
}
func (g *toolbarDragGesture) destination(p image.Point) (string, float32, bool) {
	clip := g.scroll.Geom.ContentBBox
	if !p.In(clip) || len(g.rows) == 0 {
		return "", 0, false
	}
	for _, row := range g.rows {
		// Full rectangles retain the true midpoint of clipped/offscreen rows.
		if row.Scene != g.scroll.Scene || row.Parent != g.scroll {
			return "", 0, false
		}
		y, h := row.Geom.Pos.Total.Y, row.Geom.Size.Actual.Total.Y
		if float32(p.Y) < y+h/2 {
			return row.Name, y, true
		}
	}
	last := g.rows[len(g.rows)-1]
	return "", last.Geom.Pos.Total.Y + last.Geom.Size.Actual.Total.Y, true
}

// Native scene animations run only on originating render ownership. There is
// at most one animation per captured gesture; exit removes it synchronously.
func (g *toolbarDragGesture) stopAnimation() {
	if a := g.animation; a != nil {
		a.Done = true
		a.Func = nil
		a.Widget = nil
		if g.animationScene != nil {
			g.animationScene.Animations = slices.DeleteFunc(g.animationScene.Animations, func(other *core.Animation) bool { return other == a })
		}
	}
	g.animation = nil
	g.animationScene = nil
}
func (g *toolbarDragGesture) preview(p image.Point) {
	scene := g.scroll.Scene
	if scene == nil || scene.Stage == nil {
		return
	}
	_, y, valid := g.destination(p)
	if !valid {
		g.stopAnimation()
		g.hasLine = false
		if g.stage != nil {
			g.stage.Sprites.Delete(g.sprite)
		}
		g.stage = nil
		return
	}
	clip := g.scroll.Geom.ContentBBox.Add(scene.SceneGeom.Pos)
	y += float32(scene.SceneGeom.Pos.Y)
	y = max(float32(clip.Min.Y)+1, min(float32(clip.Max.Y)-1, y))
	if g.hasLine && g.targetY == y && g.previewClip == clip {
		return
	}
	g.previewClip = clip
	// DialogStage is itself a main stage. Its parent Main is underneath the
	// sheet, and only the current top main stage's sprite layer is composed.
	switch scene.Stage.Type {
	case core.WindowStage, core.DialogStage:
		g.stage = scene.Stage
	case core.MenuStage, core.TooltipStage, core.SnackbarStage, core.CompleterStage:
		g.stage = scene.Stage.Main
	default:
		return
	}
	if g.stage == nil {
		return
	}
	if !g.hasLine {
		g.lineY = y
		g.hasLine = true
		g.alpha = .4
	}
	g.fromY = g.lineY
	g.fromAlpha = g.alpha
	g.targetY = y
	g.elapsed = 0
	if g.scope.reducedMotion {
		g.stopAnimation()
		g.lineY = y
		g.alpha = 1
		g.publishLine()
		scene.NeedsRender()
		return
	}
	g.publishLine()
	if g.animation == nil {
		a := &core.Animation{Widget: &g.scroll.WidgetBase}
		a.Func = func(a *core.Animation) {
			if !g.scope.input.allowed() || !g.hasLine || g.scroll.Scene != scene {
				a.Done = true
				return
			}
			// An explicit/manual clock may paint twice without advancing.
			// Do not keep requesting redraws when interpolation cannot progress.
			if a.Dt <= 0 {
				return
			}
			g.elapsed += a.Dt
			f := min(float32(1), g.elapsed/100)
			eased := 1 - (1-f)*(1-f)*(1-f)
			g.lineY = g.fromY + (g.targetY-g.fromY)*eased
			g.alpha = g.fromAlpha + (1-g.fromAlpha)*eased
			g.publishLine()
			scene.NeedsRender()
			if f >= 1 {
				a.Done = true
				g.animation = nil
				g.animationScene = nil
			}
		}
		g.animation = a
		g.animationScene = scene
		scene.Animations = append(scene.Animations, a)
	}
	scene.NeedsRender()
}
func (g *toolbarDragGesture) publishLine() {
	if g.stage == nil || !g.hasLine || !g.scope.input.allowed() {
		return
	}
	clip, y, scale, alpha := g.previewClip, g.lineY, float32(g.scale), g.alpha
	col := color.NRGBA{R: 0, G: 112, B: 245, A: uint8(255 * alpha)}
	thickness := max(float32(2), 2*scale)
	line := max(float32(clip.Min.Y)+thickness/2, min(float32(clip.Max.Y)-thickness/2, y))
	sprite := core.NewSprite(g.sprite, func(pc *paint.Painter) {
		// Immutable detached values: painting never mutates captures or follows refs.
		pc.FillBox(math32.Vec2(float32(clip.Min.X)+2*scale, line-thickness/2), math32.Vec2(max(float32(0), float32(clip.Dx())-4*scale), thickness), colors.Uniform(col))
		pc.FillBox(math32.Vec2(float32(clip.Min.X)+2*scale, line-3*scale), math32.Vec2(3*scale, 6*scale), colors.Uniform(col))
	})
	sprite.Active = true
	g.stage.Sprites.Add(sprite)
}
