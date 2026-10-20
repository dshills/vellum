package main

import (
	"fmt"
	"image/color"
	"math"

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
		b := row.Geom.TotalBBox
		g.rows = append(g.rows, layerBox{entry.id, bounds(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y)})
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
		}
	}
	if w.active != nil && e.model.layerDrag != w.active {
		w.endCapture()
	}
	if e.model.layerDrag != nil {
		w.publishPreview(scroll)
	} else {
		w.clearPreview()
	}
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
		w.active = nil
		w.owner = nil
		w.ticket = nil
		w.token = 0
		w.clearPreview()
		_ = ticket.Publish(func() { e.model.cancelLayerDrag(s); _ = e.a.Invalidate("placement") })
	}
	handle := func(input events.Event) {
		terminal := input.Type() == events.MouseUp && input.MouseButton() == events.Left
		if input.Type() != events.MouseMove && input.Type() != events.MouseDrag && !terminal {
			return
		}
		if terminal {
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
