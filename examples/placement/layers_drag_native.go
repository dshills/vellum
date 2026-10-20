package main

import (
	"fmt"
	"image/color"
	"math"
	"sync/atomic"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
)

type layerEditor struct {
	a     *app.Application
	ref   app.WindowRef
	model *model
}
type layerMount struct {
	core.Frame
	editor        *layerEditor
	active        *layerDrag
	owner         *core.WidgetBase
	ticket        *app.Completion
	token         uint64
	stage         *core.Stage
	sprite        *core.Sprite
	restoreQueued bool
	pulseScene    *core.Scene
	pulse         *layerPulse
}

// Completion callbacks own detached atomic state, never native mount fields.
// A native Custom event consumes readiness under render ownership.
type layerPulse struct {
	ready, canceled atomic.Bool
	ticket          *app.Completion
	timer           *time.Timer
}

const layerDragSprite = "vellum.placement-app.layer-insertion"

func (e *layerEditor) Mount(parent tree.Node) core.Widget {
	w := tree.New[layerMount](parent)
	w.editor = e
	w.Styler(func(s *styles.Style) { s.Display = styles.DisplayNone })
	return w
}
func (w *layerMount) clearPreview() {
	if w.stage != nil {
		w.stage.Sprites.Delete(layerDragSprite)
	}
	w.stage = nil
	w.sprite = nil
}
func (w *layerMount) endCapture() {
	w.stopPulse()
	if w.owner != nil && w.owner.Scene != nil {
		w.owner.Scene.Events.ReleasePointer(w.token)
	}
	if w.ticket != nil {
		w.ticket.Abort()
	}
	w.active = nil
	w.owner = nil
	w.ticket = nil
	w.token = 0
	w.clearPreview()
}
func (w *layerMount) Destroy() {
	if scene := w.pulseScene; scene != nil && scene.Property(layerPulseOwner) == w {
		scene.SetProperty(layerPulseOwner, nil)
	}
	w.pulseScene = nil
	// Reconciliation may already hold the non-reentrant app owner. Publish the
	// domain cleanup through its reserved completion instead of entering Do.
	if s, ticket := w.active, w.ticket; s != nil && ticket != nil {
		e := w.editor
		_ = ticket.Publish(func() { e.model.cancelLayerDrag(s); _ = e.a.Invalidate("placement") })
		w.ticket = nil // Transfer publication; endCapture must not withdraw it.
	}
	w.endCapture()
	w.Frame.Destroy()
}
func layerWidgets(parent tree.Node) (*core.WidgetBase, *core.WidgetBase) {
	if parent == nil {
		return nil, nil
	}
	panel := core.AsWidget(parent.AsTree().ChildByName("layers-panel"))
	if panel == nil {
		return nil, nil
	}
	scroll := core.AsWidget(panel.AsTree().ChildByName("layers-scroll"))
	if scroll == nil {
		return nil, nil
	}
	return scroll, core.AsWidget(scroll.AsTree().ChildByName("layers-rows"))
}
func layerGeometry(scroll, rows *core.WidgetBase, entries []layerRow) (layerLayout, error) {
	if scroll == nil || rows == nil || scroll.Scene == nil || scroll.Geom.ContentBBox.Empty() {
		return layerLayout{}, fmt.Errorf("layers not ready")
	}
	unit := units.Dp(1)
	unit.ToDots(&scroll.Styles.UnitContext)
	scale := float64(unit.Dots)
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		return layerLayout{}, fmt.Errorf("invalid layer scale")
	}
	offset := scroll.Scene.SceneGeom.Pos
	bounds := func(minX, minY, maxX, maxY int) canvas.Bounds {
		return canvas.Bounds{Min: canvas.Point{X: float64(minX + offset.X), Y: float64(minY + offset.Y)}, Max: canvas.Point{X: float64(maxX + offset.X), Y: float64(maxY + offset.Y)}}
	}
	b := scroll.Geom.ContentBBox
	g := layerLayout{clip: bounds(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y), scale: scale}
	for _, entry := range entries {
		row := core.AsWidget(rows.AsTree().ChildByName("layer-select-" + entry.id))
		if row == nil {
			return layerLayout{}, fmt.Errorf("layer row removed")
		}
		// TotalBBox is clipped, even empty for offscreen rows. Capture their
		// full layout rectangles so scrolling and clipped hits retain identity
		// and the real midpoint, independent of the visible portion.
		p, size := row.Geom.Pos.Total, row.Geom.Size.Actual.Total
		g.rows = append(g.rows, layerBox{entry.id, canvas.Bounds{Min: canvas.Point{X: float64(p.X) + float64(offset.X), Y: float64(p.Y) + float64(offset.Y)}, Max: canvas.Point{X: float64(p.X+size.X) + float64(offset.X), Y: float64(p.Y+size.Y) + float64(offset.Y)}}})
	}
	return g, nil
}
func (e *layerEditor) Sync(widget core.Widget) error {
	w := widget.(*layerMount)
	scroll, rows := layerWidgets(widget.AsTree().Parent)
	if scroll == nil || rows == nil {
		return fmt.Errorf("layers editor requires its panel sibling")
	}
	for _, entry := range e.model.layerRows() {
		row := core.AsWidget(rows.AsTree().ChildByName("layer-select-" + entry.id))
		if row == nil {
			continue
		}
		const prop = "vellum-placement-layer-drag"
		if row.Property(prop) != e {
			row.SetProperty(prop, e)
			row.OnFirst(events.MouseDown, func(event events.Event) { w.press(scroll, rows, row, entry.id, event) })
			// A tooltip popup would retire scoped capture during a held drag.
			// Keep normal help available when no drag is active.
			for _, typ := range []events.Types{events.LongHoverStart, events.LongPressStart} {
				row.OnFirst(typ, func(event events.Event) {
					if w.active != nil {
						event.SetHandled()
					}
				})
			}
		}
	}
	if w.active != nil && e.model.layerDrag != w.active {
		w.endCapture()
	}
	w.bindPulse(scroll.Scene)

	if e.model.layerDrag != nil {
		w.publishPreview(scroll)
	} else {
		w.clearPreview()
	}
	w.armPulse(scroll)
	const prop = "vellum-placement-layer-preview"
	if scroll.Property(prop) != w {
		scroll.SetProperty(prop, w)
		scroll.FinalStyler(func(*styles.Style) {
			if w.restoreQueued {
				return
			}
			w.restoreQueued = true
			scroll.Defer(func() {
				w.restoreQueued = false
				if w.This != nil && w.stage != nil && w.sprite != nil {
					w.stage.Sprites.Add(w.sprite)
					scroll.Scene.NeedsRender()
				}
			})
		})
	}
	return nil
}
func (w *layerMount) press(scroll, rows, row *core.WidgetBase, id string, event events.Event) {
	e := w.editor
	if event.Type() != events.MouseDown || event.MouseButton() != events.Left || row.IsDisabled() || e.ref.Context().Err() != nil {
		return
	}
	var s *layerDrag
	var ticket *app.Completion
	_ = e.a.Do(func() {
		g, err := layerGeometry(scroll, rows, e.model.layerRows())
		if err != nil {
			return
		}
		s = e.model.beginLayerDrag(id, eventPoint(event), g)
		if s == nil {
			return
		}
		ticket, err = e.a.ReserveWindowCompletion(e.ref, nil)
		if err != nil {
			e.model.cancelLayerDrag(s)
			s = nil
		}
	})
	if s == nil {
		return
	}
	event.SetHandled()
	lost := func() {
		w.stopPulse()
		w.active = nil
		w.owner = nil
		w.ticket = nil
		w.token = 0
		w.clearPreview()
		_ = ticket.Publish(func() { e.model.cancelLayerDrag(s); _ = e.a.Invalidate("placement") })
	}
	handle := func(input events.Event) {
		terminal := input.Type() == events.MouseUp && input.MouseButton() == events.Left
		wheel, scrolling := input.(*events.MouseScroll)
		if input.Type() != events.MouseMove && input.Type() != events.MouseDrag && !terminal && !scrolling {
			return
		}
		if terminal {
			w.stopPulse()
			ticket.Abort()
			w.active = nil
			w.owner = nil
			w.ticket = nil
			w.token = 0
			w.clearPreview()
		}
		_ = e.a.Do(func() {
			if e.ref.Context().Err() != nil || e.model.layerDrag != s {
				return
			}
			g, err := layerGeometry(scroll, rows, e.model.layerRows())
			if err != nil {
				e.model.cancelLayerDrag(s)
			} else if terminal {
				if err = e.model.finishLayerDrag(s, eventPoint(input), g); err != nil {
					e.model.status = err.Error()
				}
			} else {
				e.model.updateLayerDrag(s, eventPoint(input), g)
				if scrolling && s.dragging && e.model.layerDrag == s {
					// The captured wheel always scrolls this vertical layer list,
					// including with Shift held. Never route it to canvas zoom.
					w.scrollBy(scroll, rows, float64(wheel.Delta.Y))
				}
			}
			if terminal || e.model.layerDrag != s {
				if !terminal {
					w.endCapture()
				}
				_ = e.a.Invalidate("placement")
			} else {
				// Reconcile only at completion. Disabling/restyling the view
				// mid-capture moves the list and invalidates its own mapping.
				// Preview painting uses detached values and native render wake.
				w.publishPreview(scroll)
				scroll.Scene.NeedsRender()
				w.armPulse(scroll)
			}
		})
	}
	token, err := scroll.Scene.Events.CapturePointer(scroll.This.(core.Widget), event, handle, lost)
	if err != nil {
		ticket.Abort()
		scroll.Scene.Events.ConsumePointerRelease()
		_ = e.a.Do(func() { e.model.cancelLayerDrag(s); _ = e.a.Invalidate("placement") })
		return
	}
	w.active = s
	w.ticket = ticket
	w.owner = scroll
	w.token = token
	row.SetFocusQuiet()
}

const layerPulseOwner = "vellum-placement-layer-pulse-owner"
const layerPulseInstalled = "vellum-placement-layer-pulse-installed"

func (w *layerMount) bindPulse(scene *core.Scene) {
	w.pulseScene = scene
	scene.SetProperty(layerPulseOwner, w)
	if scene.Property(layerPulseInstalled) == true {
		return
	}
	scene.SetProperty(layerPulseInstalled, true)
	// The scene owns one handler. It looks up the current mount and widgets;
	// no retired mount/editor/row closures accumulate across reconciliation.
	scene.OnFirst(events.Custom, func(input events.Event) {
		current, _ := scene.Property(layerPulseOwner).(*layerMount)
		if current != nil && current.This != nil {
			current.handlePulse(input)
		}
	})
}
func (w *layerMount) handlePulse(input events.Event) {
	ce, ok := input.(*events.CustomEvent)
	if !ok {
		return
	}
	p, ok := ce.Data.(*layerPulse)
	if !ok || p != w.pulse || !p.ready.Load() {
		return
	}
	input.SetHandled()
	e := w.editor
	_ = e.a.Do(func() {
		w.stopPulse()
		scroll, rows := layerWidgets(w.AsTree().Parent)
		if w.active == nil || e.model.layerDrag != w.active {
			return
		}
		if scroll == nil || rows == nil || scroll.Scene == nil || !scroll.Scene.Events.PointerCaptured(w.token) {
			e.model.cancelLayerDrag(w.active)
			w.endCapture()
			_ = e.a.Invalidate("placement")
			return
		}
		w.scrollBy(scroll, rows, layerEdgeDelta(w.active))
		w.publishPreview(scroll)
		scroll.Scene.NeedsRender()
		w.armPulse(scroll)
	})
}

func (w *layerMount) stopPulse() {
	if p := w.pulse; p != nil {
		p.canceled.Store(true)
		if p.timer != nil {
			p.timer.Stop()
		}
		p.ticket.Abort()
	}
	w.pulse = nil
}

func layerScrollFrame(scroll *core.WidgetBase) *core.Frame {
	if scroll == nil || scroll.This == nil {
		return nil
	}
	if layout, ok := scroll.This.(core.Layouter); ok {
		return layout.AsFrame()
	}
	return nil
}

func layerScrollDestination(fr *core.Frame, delta float64) (float32, bool) {
	if fr == nil || !fr.HasScroll[math32.Y] || fr.Scrolls[math32.Y] == nil || math.IsNaN(delta) || math.IsInf(delta, 0) {
		return 0, false
	}
	sb := fr.Scrolls[math32.Y]
	maximum, visible, _ := fr.ScrollValues(math32.Y)
	limit := math.Max(0, float64(maximum-visible))
	// Bound even synthetic wheel records to one visible page per event.
	delta = math.Max(-float64(visible), math.Min(float64(visible), delta))
	next := float32(math.Max(0, math.Min(limit, float64(sb.Value)+delta)))
	return next, next != sb.Value
}

// Native mutations run only from event/Sync ownership, never a timer goroutine.
func (w *layerMount) scrollBy(scroll, rows *core.WidgetBase, delta float64) {
	e, s := w.editor, w.active
	if s == nil || e.model.layerDrag != s || !s.dragging {
		return
	}
	before, err := layerGeometry(scroll, rows, e.model.layerRows())
	if err != nil || !e.model.updateLayerDrag(s, s.pointer, before) {
		e.model.cancelLayerDrag(s)
		w.endCapture()
		_ = e.a.Invalidate("placement")
		return
	}
	fr := layerScrollFrame(scroll)
	next, changed := layerScrollDestination(fr, delta)
	if !changed {
		return
	}
	sb := fr.Scrolls[math32.Y]
	prior := sb.Value
	sb.SetValue(next)
	fr.ScrollChanged(math32.Y, sb)
	after, err := layerGeometry(scroll, rows, e.model.layerRows())
	if err != nil || !e.model.scrollLayerDrag(s, before, after, float64(sb.Value-prior)) {
		e.model.cancelLayerDrag(s)
		w.endCapture()
		_ = e.a.Invalidate("placement")
	}
}

// At most one timer and one reserved completion belong to this capture. The
// timer publishes detached readiness and queues one native wake. That wake
// performs the scroll without restyling/reconciling the captured list.
// Aborting cancels both unpublished and ready work on every terminal path.
func (w *layerMount) armPulse(scroll *core.WidgetBase) {
	s, e := w.active, w.editor
	if s == nil || e.model.layerDrag != s || e.ref.Context().Err() != nil {
		w.stopPulse()
		return
	}
	if _, can := layerScrollDestination(layerScrollFrame(scroll), layerEdgeDelta(s)); !can {
		w.stopPulse()
		return
	}
	if w.pulse != nil {
		return
	}
	ticket, err := e.a.ReserveWindowCompletion(e.ref, nil)
	if err != nil {
		return
	}
	p := &layerPulse{ticket: ticket}
	w.pulse = p
	source := scroll.SystemEvents()
	p.timer = time.AfterFunc(50*time.Millisecond, func() {
		_ = ticket.Publish(func() {
			if !p.canceled.Load() {
				p.ready.Store(true)
				input := &events.CustomEvent{}
				input.Typ, input.Data = events.Custom, p
				input.SetUnique()
				input.Init()
				source.Deque.Send(input)
			}
		})
	})
}
func (w *layerMount) publishPreview(scroll *core.WidgetBase) {
	w.clearPreview()
	s := w.editor.model.layerDrag
	if s == nil || !s.dragging || scroll.Scene == nil || scroll.Scene.Stage == nil || scroll.Scene.Stage.Main == nil {
		return
	}
	g, p, valid, y := s.layout, s.pointer, s.valid, s.line
	life := w.editor.ref.Context()
	sprite := core.NewSprite(layerDragSprite, func(pc *paint.Painter) {
		if life.Err() != nil || scroll.This == nil || !scroll.IsVisible() {
			return
		}
		fill := func(x1, y1, x2, y2 float64, col color.NRGBA) {
			x1 = math.Max(x1, g.clip.Min.X)
			y1 = math.Max(y1, g.clip.Min.Y)
			x2 = math.Min(x2, g.clip.Max.X)
			y2 = math.Min(y2, g.clip.Max.Y)
			if x2 > x1 && y2 > y1 {
				pc.FillBox(math32.Vec2(float32(x1), float32(y1)), math32.Vec2(float32(x2-x1), float32(y2-y1)), colors.Uniform(col))
			}
		}
		if valid {
			fill(g.clip.Min.X, y-2*g.scale, g.clip.Max.X, y+2*g.scale, color.NRGBA{R: 25, G: 90, B: 180, A: 255})
		}
		fill(p.X+8*g.scale, p.Y+8*g.scale, p.X+20*g.scale, p.Y+20*g.scale, color.NRGBA{R: 25, G: 90, B: 180, A: 100})
	})
	sprite.Active = true
	w.stage = scroll.Scene.Stage.Main
	w.sprite = sprite
	w.stage.Sprites.Add(sprite)
}
