package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyz/xyzcore"
	"errors"
	"example.com/vellum/app"
	"example.com/vellum/canvas"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

// GPUCanvas renders rectangles through Cogent's public shared-device direct-render
// path. Native widget destruction releases the scene's GPU resources. Rectangles
// use one reusable mesh and per-object transforms/materials; no CPU readback.
type GPUCanvas struct {
	OnError func(error)
	Owner   *app.Application
	Window  string
	Model   *canvas.Scene
	Select  func(string)
	widget  *canvasWidget
	solids  []*xyz.Solid
	ids     []string
}

func (g *GPUCanvas) Mount(parent tree.Node) core.Widget {
	if g == nil || g.Owner == nil || g.Model == nil || g.Model.Validate() != nil {
		return nil
	}
	g.solids = nil
	g.ids = nil
	g.widget = tree.New[canvasWidget](parent)
	g.widget.owner = g
	g.widget.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(240), units.Dp(200)); s.Grow.Set(1, 1) })
	sc := g.widget.XYZ
	xyz.NewAmbient(sc, "vellum.ambient", 1, xyz.DirectSun)
	mesh := xyz.NewBox(sc, "vellum.rectangle", 1, 1, .000001)
	for i, o := range g.Model.Objects {
		name := fmt.Sprintf("object-%d", i)
		solid := xyz.NewSolid(sc)
		solid.SetName(name)
		solid.SetMesh(mesh).SetColor(o.Color).SetPos(float32(o.X+o.Width/2), float32(-o.Y-o.Height/2), 0)
		solid.Pose.Scale.Set(float32(o.Width), float32(o.Height), 1)
		g.solids = append(g.solids, solid)
		g.ids = append(g.ids, o.ID)
	}
	sc.Camera.Ortho = true
	sc.Camera.FOV = 30
	sc.Camera.Near = .01
	g.widget.OnFirst(events.Scroll, func(e events.Event) {
		if s, ok := e.(*events.MouseScroll); ok {
			g.Owner.Do(func() {
				g.size()
				p := e.Pos().Sub(g.widget.Geom.ContentBBox.Min)
				if err := g.Model.View.ZoomAt(canvas.Point{X: float64(p.X), Y: float64(p.Y)}, math.Exp(-float64(s.Delta.Y)*.003)); err != nil {
					g.report(err)
					return
				}
				g.updateCamera()
				g.Owner.Invalidate(g.Window)
			})
			e.SetHandled()
		}
	})
	g.widget.OnFirst(events.SlideMove, func(e events.Event) {
		g.Owner.Do(func() {
			g.size()
			d := e.PrevDelta()
			if err := g.Model.View.Pan(float64(d.X), float64(d.Y)); err != nil {
				g.report(err)
				return
			}
			g.updateCamera()
			g.Owner.Invalidate(g.Window)
		})
		e.SetHandled()
	})
	g.widget.OnFirst(events.MouseDown, func(e events.Event) {
		g.widget.SetFocus()
		g.Owner.Do(func() {
			g.size()
			p := e.Pos().Sub(g.widget.Geom.ContentBBox.Min)
			id := g.Model.Pick(canvas.Point{X: float64(p.X), Y: float64(p.Y)})
			if g.Select != nil {
				g.Select(id)
			}
			g.Owner.Invalidate(g.Window)
		})
		e.SetHandled()
	})
	g.widget.OnFirst(events.DoubleClick, func(e events.Event) { e.SetHandled() })
	g.widget.OnFirst(events.KeyChord, func(e events.Event) {
		chord := strings.ToLower(string(e.KeyChord()))
		dx, dy := 0., 0.
		switch chord {
		case "leftarrow":
			dx = 30
		case "rightarrow":
			dx = -30
		case "uparrow":
			dy = 30
		case "downarrow":
			dy = -30
		default:
			return
		}
		g.Owner.Do(func() {
			if err := g.Model.View.Pan(dx, dy); err != nil {
				g.report(err)
				return
			}
			g.updateCamera()
			g.Owner.Invalidate(g.Window)
		})
		e.SetHandled()
	})
	// Reconcile model changes during widget updates; Render handles post-layout size.

	return g.widget
}
func (g *GPUCanvas) size() {
	sz := g.widget.Geom.ContentBBox.Size()
	if sz.X > 0 && sz.Y > 0 {
		if err := g.Model.View.SetSize(float64(sz.X), float64(sz.Y)); err != nil {
			g.report(err)
		}
	}
}
func (g *GPUCanvas) updateCamera() { g.syncCamera(); g.widget.NeedsRender() }
func (g *GPUCanvas) syncCamera() {
	v := g.Model.View
	sc := g.widget.XYZ
	sc.Camera.Far = float32(v.Height / (v.Zoom * 2 * math.Tan(math.Pi/12)))
	if sc.Camera.Far < .04 {
		sc.Camera.Far = .04
	}
	// Cogent orthographic matrices retain the GL depth convention. Keeping
	// all object depths within the positive half also satisfies WebGPU clipping.
	sc.Camera.Pose.Pos.Set(float32(v.Center.X), float32(-v.Center.Y), sc.Camera.Far*.8)
	sc.Camera.LookAt(math32.Vec3(float32(v.Center.X), float32(-v.Center.Y), 0), math32.Vec3(0, 1, 0))
	sc.Camera.UpdateMatrix()
	for i, o := range g.Model.Objects {
		if i >= len(g.solids) {
			break
		}
		z := sc.Camera.Far * .2 * float32(i) / float32(len(g.solids))
		g.solids[i].SetPos(float32(o.X+o.Width/2), float32(-o.Y-o.Height/2), z)
		g.solids[i].Pose.Scale.Set(float32(o.Width), float32(o.Height), 1)
	}
	sc.Update()
}
func (g *GPUCanvas) Sync(core.Widget) error {
	if g == nil || g.widget == nil || g.Model == nil {
		return fmt.Errorf("canvas: mounted renderer and model required")
	}
	if err := g.Model.Validate(); err != nil {
		return err
	}
	if len(g.solids) != len(g.Model.Objects) {
		return fmt.Errorf("canvas: replace the extension for topology changes")
	}
	for i, o := range g.Model.Objects {
		if g.ids[i] != o.ID {
			return fmt.Errorf("canvas: stable object IDs required")
		}
		col := o.Color
		if o.ID == g.Model.Selected {
			col = color.RGBA{255, 190, 69, 255}
		}
		g.solids[i].SetColor(col)
	}
	g.updateCamera()
	return nil
}

func (g *GPUCanvas) report(err error) {
	if g.OnError != nil {
		g.OnError(err)
	}
}

// Snapshot explicitly reads back a live GPU frame. Call only on the UI owner;
// ordinary rendering and navigation do not perform readback.
func (g *GPUCanvas) Snapshot() (*image.NRGBA, error) {
	if g == nil || g.widget == nil || g.widget.XYZ.Frame == nil {
		return nil, ErrFrameNotReady
	}
	if err := g.Sync(g.widget); err != nil {
		return nil, err
	}
	g.size()
	g.syncCamera()
	img := g.widget.XYZ.RenderGrabImage()
	if img == nil {
		return nil, fmt.Errorf("canvas: GPU readback failed")
	}
	return img, nil
}

// canvasWidget synchronizes the viewport after native allocation, immediately
// before direct rendering. Updaters run before layout and cannot supply this size.
type canvasWidget struct {
	xyzcore.Scene
	owner *GPUCanvas
}

func (w *canvasWidget) OnAdd() { w.WidgetBase.OnAdd(); w.WidgetBase.Scene.AddDirectRender(w) }
func (w *canvasWidget) Destroy() {
	w.WidgetBase.Scene.DeleteDirectRender(w)
	w.XYZ.Destroy()
	w.WidgetBase.Destroy()
}
func (w *canvasWidget) Render() {
	if w.owner == nil {
		w.Scene.Render()
		return
	}
	if err := w.owner.Owner.Do(func() { w.owner.size(); w.owner.syncCamera(); w.Scene.Render() }); err != nil {
		w.owner.report(err)
	}
}

var ErrFrameNotReady = errors.New("canvas: native GPU frame not ready")
