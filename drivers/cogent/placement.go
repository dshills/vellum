package cogent

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"sync"
	"sync/atomic"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

type placementKey struct {
	owner  *app.Application
	window string
}

var placementWindows = struct {
	sync.Mutex
	windows map[placementKey]*placementWindow
	models  map[any]*placementBinding
}{windows: map[placementKey]*placementWindow{}, models: map[any]*placementBinding{}}

type placementWindow struct {
	key      placementKey
	bindings []*placementBinding
	active   *nativePlacement
}
type placementBinding struct {
	domain          context.Context
	r               *record
	window          *placementWindow
	source          *placement.Source
	target          *placement.Target
	model           *canvas.Scene
	scope           atomic.Pointer[workbenchInputScope]
	frame           atomic.Pointer[placementFrame]
	picture         atomic.Pointer[canvasPicture]
	viewportPending atomic.Pointer[placementFrame]
	viewportQueued  atomic.Bool
	retired         bool
}
type placementFrame struct {
	geometry placement.Geometry
	scene    *core.Scene
}
type canvasPicture struct {
	objects  []canvas.Rect
	view     canvas.Viewport
	selected string
	grid     float64
	preview  placement.State
}
type nativePlacement struct {
	source, target *placementBinding
	scope          *workbenchInputScope
	gesture        *placement.Gesture
	scene          *core.Scene
	token          uint64
	sprite         string
	cancellation   *app.Completion
}

func syncPlacement(r *record) error {
	c := r.control
	if c == nil {
		return nil
	}
	old := c.placement
	needed := r.node.Kind == ui.Canvas2D || r.node.PlacementSource != nil
	if old != nil && (!needed || old.source != r.node.PlacementSource || old.target != r.node.PlacementTarget || old.model != r.node.Canvas) {
		disposePlacement(r)
		old = nil
	}
	if !needed {
		return nil
	}
	if old == nil {
		domain, err := c.a.WindowContext(c.window)
		if err != nil || domain.Err() != nil {
			return app.ErrWindowClosed
		}
		binding := &placementBinding{domain: domain, r: r, source: r.node.PlacementSource, target: r.node.PlacementTarget, model: r.node.Canvas}
		placementWindows.Lock()
		for _, model := range binding.models() {
			if placementWindows.models[model] != nil {
				placementWindows.Unlock()
				return fmt.Errorf("placement: model or endpoint already has a native binding")
			}
		}
		key := placementKey{c.a, c.window}
		window := placementWindows.windows[key]
		if window == nil {
			window = &placementWindow{key: key}
			placementWindows.windows[key] = window
		}
		if len(window.bindings) >= 512 {
			placementWindows.Unlock()
			return fmt.Errorf("placement: native binding budget exceeded")
		}
		binding.window = window
		window.bindings = append(window.bindings, binding)
		for _, model := range binding.models() {
			placementWindows.models[model] = binding
		}
		placementWindows.Unlock()
		c.placement = binding
		old = binding
		setPlacementScope(old, true, c.disabled, nil)
	}
	if old.model != nil {
		old.applyPendingViewport()
		picture := &canvasPicture{objects: slices.Clone(old.model.Objects), view: old.model.View, selected: old.model.Selected, grid: float64(r.node.CanvasGrid)}
		if old.target != nil {
			picture.preview = old.target.State()
		}
		old.picture.Store(picture)
		r.widget.AsWidget().NeedsRender()
	}
	return nil
}
func (b *placementBinding) models() []any {
	out := []any{}
	if b.source != nil {
		out = append(out, b.source)
	}
	if b.target != nil {
		out = append(out, b.target)
	}
	if b.model != nil {
		out = append(out, b.model)
	}
	return out
}

// Disposal already owns the application and native render context. It never
// invokes a native lost callback that might try to acquire the owner again.
func disposePlacement(r *record) {
	if r == nil || r.control == nil || r.control.placement == nil {
		return
	}
	b := r.control.placement
	r.control.placement = nil
	b.retired = true
	if old := b.scope.Swap(nil); old != nil {
		old.revoked.Store(true)
	}
	if cap := b.window.active; cap != nil && (cap.source == b || cap.target == b) {
		b.window.cancelOwned(cap)
	}
	if b.source != nil {
		b.source.Cancel()
	}
	if b.target != nil {
		b.target.Cancel()
	}
	placementWindows.Lock()
	for _, model := range b.models() {
		delete(placementWindows.models, model)
	}
	b.window.bindings = slices.DeleteFunc(b.window.bindings, func(other *placementBinding) bool { return other == b })
	if len(b.window.bindings) == 0 {
		delete(placementWindows.windows, b.window.key)
	}
	placementWindows.Unlock()
	b.frame.Store(nil)
	b.viewportPending.Store(nil)
	b.picture.Store(nil)
	b.source = nil
	b.target = nil
	b.model = nil
}
func setPlacementScope(b *placementBinding, visible, disabled bool, scopes []*tabCloseScope) {
	old := b.scope.Load()
	if old != nil && !old.revoked.Load() && old.visible == visible && old.disabled == disabled && slices.Equal(old.scopes, scopes) {
		return
	}
	if old != nil {
		old.revoked.Store(true)
	}
	b.scope.Store(&workbenchInputScope{visible: visible, disabled: disabled, scopes: slices.Clone(scopes)})
	if cap := b.window.active; cap != nil && (cap.source == b || cap.target == b) {
		b.window.cancelOwned(cap)
	}
}
func (b *placementBinding) live() bool {
	return b != nil && b.domain != nil && b.domain.Err() == nil && !b.retired && b.r.control != nil && !b.r.control.retired && b.r.control.placement == b && b.scope.Load().allowed()
}

// Retain the originating binding/scope across queued button activation. This
// also rejects an ordinary button becoming a placement source before delivery.
// The returned predicate is evaluated only on the application owner.
func placementActivationGuard(r *record) func() bool {
	c := r.control
	b := c.placement
	if b == nil {
		return func() bool { return r.control == c && !c.retired && c.placement == nil }
	}
	scope := b.scope.Load()
	return func() bool {
		return r.control == c && c.placement == b && b.source != nil && b.live() && scope.allowed()
	}
}

func placementScale(w *core.WidgetBase) float64 {
	u := units.Dp(1)
	u.ToDots(&w.Styles.UnitContext)
	return float64(u.Dots)
}

func beginPlacement(r *record, e events.Event) {
	if e.Type() != events.MouseDown || e.MouseButton() != events.Left || r.control == nil {
		return
	}
	b := r.control.placement
	if b == nil || b.source == nil {
		return
	}
	e.SetHandled()
	cap := &nativePlacement{source: b, scene: r.widget.AsWidget().Scene, sprite: fmt.Sprintf("vellum.placement.%p", b)}
	token, err := cap.scene.Events.CapturePointer(r.widget, e, func(event events.Event) { b.window.event(cap, event) }, func() { b.window.lost(cap) })
	if err != nil {
		cap.scene.Events.ConsumePointerRelease()
		return
	}
	cap.token = token
	err = r.control.a.Do(func() {
		if !b.live() || b.window.active != nil {
			return
		}
		life, err := r.control.a.WindowContext(r.control.window)
		if err != nil || life.Err() != nil {
			return
		}
		p := e.WindowPos()
		g, err := b.source.Begin(r.control.window, canvas.Point{X: float64(p.X), Y: float64(p.Y)}, 6*placementScale(r.widget.AsWidget()))
		if err != nil {
			placementRefusal(r, err)
			return
		}
		ticket, err := r.control.a.ReserveCompletion(r.control.window, nil)
		if err != nil {
			g.Cancel()
			placementRefusal(r, err)
			return
		}
		cap.scope, cap.gesture, cap.cancellation = b.scope.Load(), g, ticket
		b.window.active = cap
	})
	if err != nil || cap.gesture == nil {
		cap.scene.Events.ReleasePointer(cap.token)
		return
	}
	r.widget.AsWidget().SetFocus()
}
func placementRefusal(r *record, err error) {
	if err == nil || errors.Is(err, placement.ErrClosed) || errors.Is(err, placement.ErrBusy) || errors.Is(err, placement.ErrRejected) || errors.Is(err, app.ErrCompletionCapacity) || errors.Is(err, app.ErrClosed) || errors.Is(err, app.ErrWindowClosed) {
		return
	}
	r.control.refused.Store(&err)
	_ = r.control.a.Invalidate(r.control.window)
}
func (w *placementWindow) resolve(p image.Point) (*placementBinding, *placementFrame) {
	for i := len(w.bindings) - 1; i >= 0; i-- {
		b := w.bindings[i]
		if b.target == nil || !b.live() {
			continue
		}
		frame := b.frame.Load()
		if frame == nil || frame.scene.GlobalLayoutPending() {
			continue
		}
		point := canvas.Point{X: float64(p.X), Y: float64(p.Y)}
		if _, err := frame.geometry.World(point); err == nil {
			return b, frame
		}
	}
	return nil, nil
}
func (w *placementWindow) event(cap *nativePlacement, e events.Event) {
	if w.active != cap {
		return
	}
	if e.Type() != events.MouseDrag && e.Type() != events.MouseMove && !(e.Type() == events.MouseUp && e.MouseButton() == events.Left) {
		return
	}
	p := e.WindowPos()
	target, frame := w.resolve(p)
	var targetScope *workbenchInputScope
	if target != nil {
		targetScope = target.scope.Load()
	}
	terminal := e.Type() == events.MouseUp
	// Core already revoked the terminal capture, before application callbacks.
	if terminal {
		w.active = nil
		cap.cancellation.Abort()
		w.clearSprite(cap)
	}
	err := w.key.owner.Do(func() {
		if !cap.source.live() || !cap.scope.allowed() {
			cap.gesture.Cancel()
			return
		}
		life, err := w.key.owner.WindowContext(w.key.window)
		if err != nil || life.Err() != nil {
			cap.gesture.Cancel()
			return
		}
		var endpoint *placement.Target
		var geometry placement.Geometry
		if target != nil && target.live() && targetScope.allowed() && target.frame.Load() == frame {
			current := target.model.View
			current.Width = frame.geometry.View.Width
			current.Height = frame.geometry.View.Height
			if current == frame.geometry.View {
				endpoint = target.target
				geometry = frame.geometry
				cap.target = target
			}
		}
		pos := canvas.Point{X: float64(p.X), Y: float64(p.Y)}
		if terminal {
			_, err = cap.gesture.Release(life, endpoint, geometry, pos)
		} else {
			err = cap.gesture.Move(endpoint, geometry, pos)
		}
		placementRefusal(cap.source.r, err)
		_ = w.key.owner.Invalidate(w.key.window)
	})
	if err != nil {
		// A stopped owner has no concurrent model callbacks; mirror driver cleanup.
		if errors.Is(err, app.ErrClosed) {
			cap.gesture.Cancel()
		}
		if !terminal {
			cap.cancellation.Abort()
			w.active = nil
			w.clearSprite(cap)
			cap.scene.Events.ReleasePointer(cap.token)
		}
		return
	}
	if !terminal {
		state := cap.gesture.State()
		if !state.Active {
			cap.cancellation.Abort()
			w.active = nil
			w.clearSprite(cap)
			cap.scene.Events.ReleasePointer(cap.token)
			return
		}
		w.preview(cap, state)
	}
}
func (w *placementWindow) clearSprite(cap *nativePlacement) {
	if cap.scene != nil && cap.scene.Stage != nil && cap.scene.Stage.Main != nil {
		cap.scene.Stage.Main.Sprites.Delete(cap.sprite)
		cap.scene.NeedsRender()
	}
}
func (w *placementWindow) cancelOwned(cap *nativePlacement) {
	if w.active != cap {
		return
	}
	w.active = nil
	cap.cancellation.Abort()
	cap.gesture.Cancel()
	w.clearSprite(cap)
	cap.scene.Events.ReleasePointer(cap.token)
}
func (w *placementWindow) lost(cap *nativePlacement) {
	if w.active != cap {
		return
	}
	w.active = nil
	w.clearSprite(cap)
	// Loss can occur during deletion on an already-held owner. A reserved
	// completion guarantees cancellation admission even when the ordinary queue
	// is full. Lifetime revocation is also covered by binding disposal.
	gesture := cap.gesture
	_ = cap.cancellation.Publish(func() { gesture.Cancel(); _ = w.key.owner.Invalidate(w.key.window) })
}
func (w *placementWindow) preview(cap *nativePlacement, state placement.State) {
	if !state.Dragging {
		return
	}
	size := math32.Vec2(72, 38)
	pos := math32.Vec2(float32(state.Pointer.X+12), float32(state.Pointer.Y+12))
	col := color.NRGBAModel.Convert(state.Preview.Color).(color.NRGBA)
	col.A = 180
	sprite := core.NewSprite(cap.sprite, func(pc *paint.Painter) { pc.FillBox(pos, size, colors.Uniform(col)) })
	sprite.Active = true
	cap.scene.Stage.Main.Sprites.Add(sprite)
	cap.scene.NeedsRender()
}

// managedCanvas paints detached model/preview values using ordinary retained
// scene drawing. Native geometry is published only after actual layout/render.
type managedCanvas struct {
	core.WidgetBase
	r *record
}

func (c *managedCanvas) Init() {
	c.WidgetBase.Init()
	c.Styler(func(s *styles.Style) {
		s.Min.Set(units.Dp(240), units.Dp(180))
		s.Grow.Set(1, 1)
		s.SetAbilities(true, abilities.Focusable, abilities.Clickable, abilities.Slideable, abilities.ScrollableUnattended)
	})
}
func (c *managedCanvas) Render() {
	c.WidgetBase.Render()
	if c.r == nil || c.r.control == nil {
		return
	}
	b := c.r.control.placement
	if b == nil || b.retired {
		return
	}
	picture := b.picture.Load()
	if picture == nil {
		return
	}
	scale := placementScale(&c.WidgetBase)
	size := c.Geom.Size.Actual.Content
	if scale <= 0 || size.X <= 0 || size.Y <= 0 || c.Geom.ContentBBox.Empty() {
		b.frame.Store(nil)
		return
	}
	view := picture.view
	if err := view.SetSize(float64(size.X)/scale, float64(size.Y)/scale); err != nil {
		return
	}
	origin := c.Geom.Pos.Content
	sceneOrigin := c.Scene.SceneGeom.Pos
	clip := c.Geom.ContentBBox
	geometry := placement.Geometry{Window: b.window.key.window, Origin: canvas.Point{X: float64(origin.X) + float64(sceneOrigin.X), Y: float64(origin.Y) + float64(sceneOrigin.Y)}, Clip: canvas.Bounds{Min: canvas.Point{X: float64(clip.Min.X + sceneOrigin.X), Y: float64(clip.Min.Y + sceneOrigin.Y)}, Max: canvas.Point{X: float64(clip.Max.X + sceneOrigin.X), Y: float64(clip.Max.Y + sceneOrigin.Y)}}, Scale: scale, View: view}
	b.frame.Store(&placementFrame{geometry: geometry, scene: c.Scene})
	if picture.view.Width != view.Width || picture.view.Height != view.Height {
		b.requestViewport(b.frame.Load())
	}
	pc := &c.Scene.Painter
	fill := func(rect canvas.Rect, col color.Color) {
		p := view.Screen(canvas.Point{X: rect.X, Y: rect.Y})
		x, y := float64(origin.X)+p.X*scale, float64(origin.Y)+p.Y*scale
		right, bottom := x+rect.Width*view.Zoom*scale, y+rect.Height*view.Zoom*scale
		x = math.Max(x, float64(clip.Min.X))
		y = math.Max(y, float64(clip.Min.Y))
		right = math.Min(right, float64(clip.Max.X))
		bottom = math.Min(bottom, float64(clip.Max.Y))
		if right > x && bottom > y {
			pc.FillBox(math32.Vec2(float32(x), float32(y)), math32.Vec2(float32(right-x), float32(bottom-y)), colors.Uniform(col))
		}
	}
	if picture.grid > 0 {
		bounds, err := view.Bounds()
		if err == nil {
			spacing := picture.grid
			for (bounds.Max.X-bounds.Min.X)/spacing > 512 || (bounds.Max.Y-bounds.Min.Y)/spacing > 512 {
				spacing *= 2
			}
			width := 1 / (view.Zoom * scale)
			for x := math.Ceil(bounds.Min.X/spacing) * spacing; x < bounds.Max.X; x += spacing {
				fill(canvas.Rect{X: x, Y: bounds.Min.Y, Width: width, Height: bounds.Max.Y - bounds.Min.Y}, color.RGBA{222, 226, 232, 255})
			}
			for y := math.Ceil(bounds.Min.Y/spacing) * spacing; y < bounds.Max.Y; y += spacing {
				fill(canvas.Rect{X: bounds.Min.X, Y: y, Width: bounds.Max.X - bounds.Min.X, Height: width}, color.RGBA{222, 226, 232, 255})
			}
		}
	}
	outline := func(rect canvas.Rect, col color.Color) {
		width := 2 / (view.Zoom * scale)
		fill(canvas.Rect{X: rect.X - width, Y: rect.Y - width, Width: rect.Width + 2*width, Height: width}, col)
		fill(canvas.Rect{X: rect.X - width, Y: rect.Y + rect.Height, Width: rect.Width + 2*width, Height: width}, col)
		fill(canvas.Rect{X: rect.X - width, Y: rect.Y, Width: width, Height: rect.Height}, col)
		fill(canvas.Rect{X: rect.X + rect.Width, Y: rect.Y, Width: width, Height: rect.Height}, col)
	}
	for _, rect := range picture.objects {
		fill(rect, rect.Color)
		if rect.ID == picture.selected {
			outline(rect, color.RGBA{235, 150, 30, 255})
		}
	}
	if picture.preview.Active && picture.preview.Dragging {
		col := color.RGBA{200, 60, 60, 255}
		if picture.preview.Accepted {
			col = color.RGBA{25, 160, 90, 255}
			fill(picture.preview.Preview, color.NRGBA{R: 25, G: 160, B: 90, A: 70})
		}
		outline(picture.preview.Preview, col)
		// Target acceptance is also visible at the canvas boundary.
		p := view.World(canvas.Point{})
		outline(canvas.Rect{X: p.X, Y: p.Y, Width: view.Width / view.Zoom, Height: view.Height / view.Zoom}, col)
	}
}

// applyPendingViewport runs only under the application owner. The window's
// original lifetime must still be current, including when an ID is reused.
func (b *placementBinding) applyPendingViewport() bool {
	frame := b.viewportPending.Swap(nil)
	if !b.live() || frame == nil || b.frame.Load() != frame || b.model == nil {
		return false
	}
	life, err := b.r.control.a.WindowContext(b.r.control.window)
	if err != nil || life != b.domain || life.Err() != nil {
		return false
	}
	if b.model.View.Center != frame.geometry.View.Center || b.model.View.Zoom != frame.geometry.View.Zoom {
		return false
	}
	if b.model.View.Width == frame.geometry.View.Width && b.model.View.Height == frame.geometry.View.Height {
		return false
	}
	b.model.View.Width, b.model.View.Height = frame.geometry.View.Width, frame.geometry.View.Height
	return true
}

// Coalesce completed viewport feedback. If the ordinary queue is full, a root
// invalidation applies the retained latest frame during owner reconciliation.
// Painting never writes the mutable model or launches a worker.
func (b *placementBinding) requestViewport(frame *placementFrame) {
	b.viewportPending.Store(frame)
	if !b.viewportQueued.CompareAndSwap(false, true) {
		return
	}
	err := b.r.control.a.PostWindow(b.r.control.window, func() {
		// Clear before consuming, so a frame published during execution can post
		// its own callback rather than becoming stranded behind this one.
		b.viewportQueued.Store(false)
		if b.applyPendingViewport() {
			_ = b.r.control.a.Invalidate(b.r.control.window)
		}
	})
	if err != nil {
		b.viewportQueued.Store(false)
		if errors.Is(err, app.ErrQueueFull) {
			_ = b.r.control.a.Invalidate(b.r.control.window)
		} else {
			placementRefusal(b.r, err)
		}
	}
}

func mountManagedCanvas(parent tree.Node, r *record) *managedCanvas {
	w := tree.New[managedCanvas](parent)
	w.r = r
	w.OnFirst(events.MouseDown, func(e events.Event) {
		if e.MouseButton() != events.Left {
			return
		}
		w.SetFocus()
		canvasOwnerEvent(r, e, func(b *placementBinding, frame *placementFrame) {
			world, err := frame.geometry.World(canvas.Point{X: float64(e.WindowPos().X), Y: float64(e.WindowPos().Y)})
			if err != nil {
				return
			}
			b.model.View = frame.geometry.View
			b.model.Selected = ""
			b.model.Pick(frame.geometry.View.Screen(world))
			canvasChanged(r)
		})
		e.SetHandled()
	})
	w.OnFirst(events.Scroll, func(e events.Event) {
		scroll, ok := e.(*events.MouseScroll)
		if !ok {
			return
		}
		canvasOwnerEvent(r, e, func(b *placementBinding, frame *placementFrame) {
			b.model.View = frame.geometry.View
			local := canvas.Point{X: (float64(e.WindowPos().X) - frame.geometry.Origin.X) / frame.geometry.Scale, Y: (float64(e.WindowPos().Y) - frame.geometry.Origin.Y) / frame.geometry.Scale}
			placementRefusal(r, b.model.View.ZoomAt(local, math.Exp(-float64(scroll.Delta.Y)*.003)))
			canvasChanged(r)
		})
		e.SetHandled()
	})
	w.OnFirst(events.SlideMove, func(e events.Event) {
		canvasOwnerEvent(r, e, func(b *placementBinding, frame *placementFrame) {
			b.model.View = frame.geometry.View
			delta := e.PrevDelta()
			placementRefusal(r, b.model.View.Pan(float64(delta.X)/frame.geometry.Scale, float64(delta.Y)/frame.geometry.Scale))
			canvasChanged(r)
		})
		e.SetHandled()
	})
	w.OnFirst(events.DoubleClick, func(e events.Event) { e.SetHandled() })
	w.OnFirst(events.KeyChord, func(e events.Event) {
		dx, dy := 0., 0.
		switch e.KeyCode() {
		case key.CodeLeftArrow:
			dx = 30
		case key.CodeRightArrow:
			dx = -30
		case key.CodeUpArrow:
			dy = 30
		case key.CodeDownArrow:
			dy = -30
		default:
			return
		}
		canvasOwnerEvent(r, e, func(b *placementBinding, frame *placementFrame) {
			b.model.View = frame.geometry.View
			placementRefusal(r, b.model.View.Pan(dx, dy))
			canvasChanged(r)
		})
		e.SetHandled()
	})
	return w
}
func canvasOwnerEvent(r *record, e events.Event, change func(*placementBinding, *placementFrame)) {
	if r.control == nil || r.control.placement == nil {
		return
	}
	b := r.control.placement
	frame := b.frame.Load()
	scope := b.scope.Load()
	if frame == nil || frame.scene.GlobalLayoutPending() {
		return
	}
	err := r.control.a.Do(func() {
		life, lifeErr := r.control.a.WindowContext(r.control.window)
		if lifeErr != nil || life != b.domain || life.Err() != nil || !b.live() || !scope.allowed() || b.frame.Load() != frame {
			return
		}
		current := b.model.View
		current.Width = frame.geometry.View.Width
		current.Height = frame.geometry.View.Height
		if current != frame.geometry.View {
			return
		}
		change(b, frame)
	})
	placementRefusal(r, err)
}
func canvasChanged(r *record) {
	if r.node.Change != nil {
		r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: r.node.Kind, SelectedID: r.node.Canvas.Selected, Number: float32(r.node.Canvas.View.Zoom)})
	}
	_ = r.control.a.Invalidate(r.control.window)
}
