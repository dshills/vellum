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
	"github.com/dshills/vellum/ui"
)

// Application-owner publication; native routing never reads the customizer.
type toolbarDragScope struct {
	handle *ui.ToolbarReorderHandle
	input  *workbenchInputScope
}
type toolbarDragBinding struct {
	r      *record
	button *core.Button
	scope  atomic.Pointer[toolbarDragScope]
	active *toolbarDragGesture // originating render ownership only
}
type toolbarDragGesture struct {
	scope    *toolbarDragScope
	scroll   *core.Frame
	rows     []*core.WidgetBase
	start    image.Point
	scale    float64
	dragging bool
	token    uint64
	stage    *core.Stage
	sprite   string
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
	if old != nil && old.handle == h && old.input.visible == visible && old.input.disabled == r.control.disabled && slices.Equal(old.input.scopes, scopes) {
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
	b.scope.Store(&toolbarDragScope{h, &workbenchInputScope{visible: visible, disabled: r.control.disabled, scopes: slices.Clone(scopes)}})
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
	if g.scroll.Scene != nil {
		g.scroll.Scene.Events.ReleasePointer(g.token)
	}
	if g.stage != nil {
		g.stage.Sprites.Delete(g.sprite)
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
func (g *toolbarDragGesture) preview(p image.Point) {
	if g.stage != nil {
		g.stage.Sprites.Delete(g.sprite)
	}
	g.stage = nil
	scene := g.scroll.Scene
	if scene == nil || scene.Stage == nil || scene.Stage.Main == nil {
		return
	}
	_, y, valid := g.destination(p)
	if !valid {
		return
	}
	clip := g.scroll.Geom.ContentBBox.Add(scene.SceneGeom.Pos)
	y += float32(scene.SceneGeom.Pos.Y)
	y = max(float32(clip.Min.Y)+1, min(float32(clip.Max.Y)-1, y))
	g.stage = scene.Stage.Main
	sprite := core.NewSprite(g.sprite, func(pc *paint.Painter) {
		pc.FillBox(math32.Vec2(float32(clip.Min.X), y-1), math32.Vec2(float32(clip.Dx()), 2), colors.Uniform(color.NRGBA{R: 0, G: 112, B: 245, A: 255}))
	})
	sprite.Active = true
	g.stage.Sprites.Add(sprite)
}
