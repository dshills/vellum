package main

import (
	"fmt"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

// This application-local extension adds object manipulation to the existing
// managed Canvas2D. It does not replace its rendering, placement, pan or zoom.
// The hidden sibling installs a first listener through the pinned Core API.
type canvasEditor struct {
	a     *app.Application
	ref   app.WindowRef
	model *model
}

func (e *canvasEditor) Mount(parent tree.Node) core.Widget {
	w := core.NewFrame(parent)
	w.Styler(func(s *styles.Style) { s.Display = styles.DisplayNone })
	return w
}
func (e *canvasEditor) Sync(w core.Widget) error {
	parent := w.AsTree().Parent
	if parent == nil {
		return fmt.Errorf("canvas editor has no host")
	}
	target := core.AsWidget(parent.AsTree().ChildByName("drawing"))
	if target == nil {
		return fmt.Errorf("canvas editor requires its managed canvas sibling")
	}
	const property = "vellum-placement-app-editor"
	if target.Property(property) == e {
		return nil
	}
	target.SetProperty(property, e)
	target.OnFirst(events.MouseDown, func(event events.Event) { e.press(target, event) })
	return nil
}
func editGeometry(w *core.WidgetBase, view canvas.Viewport) (placement.Geometry, error) {
	u := units.Dp(1)
	u.ToDots(&w.Styles.UnitContext)
	scale := float64(u.Dots)
	if w.Scene == nil || scale <= 0 || w.Geom.ContentBBox.Empty() {
		return placement.Geometry{}, fmt.Errorf("canvas not ready")
	}
	if err := view.SetSize(float64(w.Geom.Size.Actual.Content.X)/scale, float64(w.Geom.Size.Actual.Content.Y)/scale); err != nil {
		return placement.Geometry{}, err
	}
	origin, scene, clip := w.Geom.Pos.Content, w.Scene.SceneGeom.Pos, w.Geom.ContentBBox
	return placement.Geometry{Window: "placement", Origin: canvas.Point{X: float64(origin.X) + float64(scene.X), Y: float64(origin.Y) + float64(scene.Y)}, Clip: canvas.Bounds{Min: canvas.Point{X: float64(clip.Min.X + scene.X), Y: float64(clip.Min.Y + scene.Y)}, Max: canvas.Point{X: float64(clip.Max.X + scene.X), Y: float64(clip.Max.Y + scene.Y)}}, Scale: scale, View: view}, nil
}
func eventPoint(event events.Event) canvas.Point {
	p := event.WindowPos()
	return canvas.Point{X: float64(p.X), Y: float64(p.Y)}
}

func (e *canvasEditor) press(w *core.WidgetBase, event events.Event) {
	if event.Type() != events.MouseDown || event.MouseButton() != events.Left || w.IsDisabled() || e.ref.Context().Err() != nil {
		return
	}
	var move *objectMove
	var ticket *app.Completion
	_ = e.a.Do(func() {
		if e.ref.Context().Err() != nil {
			return
		}
		g, err := editGeometry(w, e.model.scene.View)
		if err != nil {
			return
		}
		move = e.model.beginMove(g, eventPoint(event))
		if move == nil {
			return
		}
		ticket, err = e.a.ReserveWindowCompletion(e.ref, nil)
		if err != nil {
			e.model.cancelMove(move)
			move = nil
		}
	})
	if move == nil {
		return
	} // Empty-space input retains ordinary canvas navigation.
	event.SetHandled()
	handle := func(input events.Event) {
		terminal := input.Type() == events.MouseUp && input.MouseButton() == events.Left
		if input.Type() != events.MouseMove && input.Type() != events.MouseDrag && !terminal {
			return
		}
		if terminal {
			ticket.Abort()
		}
		_ = e.a.Do(func() {
			if e.ref.Context().Err() != nil || e.model.move != move {
				return
			}
			g, err := editGeometry(w, e.model.scene.View)
			if err != nil {
				e.model.cancelMove(move)
			} else if !e.model.updateMoveGeometry(move, g) {
				// A changed camera/display mapping already canceled this move.
			} else if terminal {
				if err := e.model.finishMove(move, eventPoint(input)); err != nil {
					e.model.status = err.Error()
				}
			} else {
				e.model.previewMove(move, eventPoint(input))
			}
			_ = e.a.Invalidate("placement")
		})
	}
	lost := func() {
		_ = ticket.Publish(func() { e.model.cancelMove(move); _ = e.a.Invalidate("placement") })
	}
	_, err := w.Scene.Events.CapturePointer(w.This.(core.Widget), event, handle, lost)
	if err != nil {
		ticket.Abort()
		w.Scene.Events.ConsumePointerRelease()
		_ = e.a.Do(func() { e.model.cancelMove(move); _ = e.a.Invalidate("placement") })
		return
	}
	// Pointer capture must not scroll the already visible canvas into focus,
	// which changes its mapping while this gesture is in progress.
	w.SetFocusQuiet()
	_ = e.a.Invalidate("placement")
}
