package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"image"
	"image/color"
	"math"
	"testing"
)

func TestGPUGeometryAndCameraCreatedBeforeUpdates(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	model, err := canvas.New([]canvas.Rect{{ID: "back", Width: 50, Height: 25, Color: color.RGBA{67, 197, 166, 255}}, {ID: "front", Width: 10, Height: 10, Color: color.RGBA{255, 190, 69, 255}}})
	if err != nil {
		t.Fatal(err)
	}
	model.View.SetSize(800, 400)
	g := &GPUCanvas{Owner: a, Model: model}
	host := core.NewBody("headless")
	native := g.Mount(host)
	if native == nil || len(g.solids) != 2 || g.solids[0] == nil || g.solids[1] == nil {
		t.Fatal("deferred geometry")
	}
	if err := g.Sync(native); err != nil {
		t.Fatal(err)
	}
	if g.solids[1].Pose.Pos.Z <= g.solids[0].Pose.Pos.Z {
		t.Fatal("paint order not reflected in depth")
	}
	expected := 400 / (2 * math.Tan(math.Pi/12))
	if math.Abs(float64(g.widget.XYZ.Camera.Far)-expected) > .001 {
		t.Fatal("projection size")
	}
	if err := model.View.ZoomAt(canvas.Point{X: 400, Y: 200}, 10); err != nil {
		t.Fatal(err)
	}
	if err := g.Sync(native); err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(g.widget.XYZ.Camera.Far)-expected/10) > .001 {
		t.Fatal("zoom/project mismatch")
	}
}

func TestCameraDepthAndAllocatedSize(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	model, _ := canvas.New([]canvas.Rect{{ID: "a", Width: 10, Height: 10}, {ID: "b", Width: 10, Height: 10}})
	g := &GPUCanvas{Owner: a, Model: model}
	g.Mount(core.NewBody("depth"))
	g.widget.Geom.ContentBBox = image.Rect(100, 50, 1500, 350)
	g.size()
	g.syncCamera()
	if model.View.Width != 1400 || model.View.Height != 300 {
		t.Fatal("viewport uses pre-layout dimensions")
	}
	for _, zoom := range []float64{.01, 1, 100} {
		model.View.Zoom = zoom
		g.syncCamera()
		cam := g.widget.XYZ.Camera
		for _, solid := range g.solids {
			point := math32.Vector4FromVector3(solid.Pose.Pos, 1).MulMatrix4(&cam.ViewMatrix).MulMatrix4(&cam.ProjectionMatrix)
			if z := point.Z / point.W; z <= 0 || z >= 1 {
				t.Fatal("object clipped by WebGPU depth", zoom, z)
			}
		}
	}
}

func TestCanvasPointerEventsUseOwnSelection(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	model, _ := canvas.New([]canvas.Rect{{ID: "a", Width: 20, Height: 20}, {ID: "b", X: 50, Width: 20, Height: 20}})
	model.View.SetSize(800, 400)
	model.View.Center = canvas.Point{X: 40, Y: 10}
	selected := ""
	g := &GPUCanvas{Owner: a, Model: model, Select: func(id string) { selected = id }}
	g.Mount(core.NewBody("pointer"))
	g.widget.Geom.ContentBBox = image.Rect(100, 50, 900, 450)
	point := model.View.Screen(canvas.Point{X: 60, Y: 10})
	event := events.NewMouse(events.MouseDown, events.Left, image.Pt(int(point.X)+100, int(point.Y)+50), 0)
	event.SetLocalOff(image.Point{})
	g.widget.HandleEvent(event)
	if selected != "b" || model.Selected != "b" || !event.IsHandled() {
		t.Fatal("native selection coordinate mismatch", selected, model.Selected)
	}
	double := events.NewMouse(events.DoubleClick, events.Left, event.WindowPos(), 0)
	double.SetLocalOff(image.Point{})
	g.widget.HandleEvent(double)
	if g.widget.CurrentSelected != nil {
		t.Fatal("inherited 3D manipulation selection leaked")
	}
}
