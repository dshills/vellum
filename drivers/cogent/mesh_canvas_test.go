package cogent

import (
	"errors"
	"image"
	"image/color"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/gpu/gpudraw"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/system/composer"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"github.com/oliverbestmann/webgpu/wgpu"
	"math"
)

func TestDrawingClipEachEdgeAtUnequalOrigin(t *testing.T) {
	full := image.Rect(37, 81, 137, 161)
	for _, test := range []struct{ clip, dest, source image.Rectangle }{
		{image.Rect(50, 95, 120, 145), image.Rect(57, 104, 127, 154), image.Rect(13, 14, 83, 64)},
		{image.Rect(0, 0, 70, 100), image.Rect(44, 90, 77, 109), image.Rect(0, 0, 33, 19)},
		{image.Rect(200, 200, 300, 300), image.Rectangle{}, image.Rectangle{}},
	} {
		dest, source := drawingClip(full, test.clip, image.Pt(7, 9), full.Size())
		if dest != test.dest || source != test.source {
			t.Fatal(dest, source, test)
		}
	}
}
func TestSnapshotLinearUnpremultiply(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	c := color.NRGBA{180, 70, 30, 128}
	img.Pix[0] = srgbByte(float64(linearByte(c.R)) * 128 / 255)
	img.Pix[1] = srgbByte(float64(linearByte(c.G)) * 128 / 255)
	img.Pix[2] = srgbByte(float64(linearByte(c.B)) * 128 / 255)
	img.Pix[3] = 128
	copy(img.Pix[4:8], []byte{255, 90, 40, 0})
	copy(img.Pix[8:12], []byte{20, 30, 40, 255})
	straightSnapshot(img)
	got := img.NRGBAAt(0, 0)
	for _, pair := range [][2]uint8{{got.R, c.R}, {got.G, c.G}, {got.B, c.B}} {
		if abs(int(pair[0])-int(pair[1])) > 2 {
			t.Fatal("not linear unpremultiply", got)
		}
	}
	if img.NRGBAAt(1, 0) != (color.NRGBA{}) || img.NRGBAAt(2, 0) != (color.NRGBA{20, 30, 40, 255}) {
		t.Fatal("zero/opaque conversion")
	}
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func TestMeshCanvasFullAllocationAndMountRetirement(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	m, _ := canvas.NewMesh(nil, nil)
	view := canvas.Viewport{Width: 200, Height: 100, Zoom: 1}
	g := &MeshCanvas{Owner: a, View: &view, Mesh: m}
	a.Open(app.Window{ID: "drawing", Title: "test", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	g.Window = "drawing"
	host := core.NewBody("drawing")
	native := g.Mount(host)
	if native == nil {
		t.Fatal("mount")
	}
	if g.Mount(host) != nil {
		t.Fatal("mutable adapter mounted twice")
	}
	g.widget.Geom.Pos.Content = math32.Vec2(37, 81)
	g.widget.Geom.Size.Actual.Content = math32.Vec2(100, 80)
	g.widget.Geom.ContentBBox = image.Rect(50, 95, 120, 145)
	g.widget.Render()
	if view.Width != 100 || view.Height != 80 {
		t.Fatal("viewport incorrectly shrank to clipped size", view)
	}
	g.widget.Destroy()
	if g.NativeWindow() != nil {
		t.Fatal("retired native window")
	}
	if g.Mount(host) == nil {
		t.Fatal("remount after retirement")
	}
}
func TestDrawingResourceBudgetAndEmptyMesh(t *testing.T) {
	m, _ := canvas.NewMesh([]canvas.Vertex{{Position: canvas.Point{}}}, nil)
	target, mesh, err := drawingBytes(image.Pt(100, 100), m)
	if err != nil || target != 200000 || mesh != 0 {
		t.Fatal(target, mesh, err)
	}
	var r meshResources
	if err := r.render(nil, image.Pt(100, 100), m, canvas.Viewport{}, color.NRGBA{}, 1); err == nil || r.frame != nil {
		t.Fatal("budget must refuse before touching GPU")
	}
}

func TestRetiredOrClosedWindowDoesNotPrepare(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	for _, id := range []string{"one", "two"} {
		a.Open(app.Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	}
	m, _ := canvas.NewMesh(nil, nil)
	view := canvas.Viewport{Width: 200, Height: 100, Zoom: 1}
	calls := 0
	g := &MeshCanvas{Owner: a, Window: "one", View: &view, Prepare: func(canvas.Viewport) (*canvas.Mesh, error) { calls++; return m, nil }}
	g.Mount(core.NewBody("one"))
	old := g.widget
	old.Geom.Size.Actual.Content = math32.Vec2(900, 700)
	a.CloseWindow(fixtureWindowRef(a, "one"))
	old.Render()
	if calls != 0 || view.Width != 200 {
		t.Fatal("closed-window preparation/layout escaped", calls, view)
	}
	old.Destroy()
	g.Window = "two"
	g.Mount(core.NewBody("two"))
	old.Render()
	if calls != 0 || view.Width != 200 {
		t.Fatal("stale renderer changed remounted model")
	}
}
func TestMeshLocalGeometryTranslationAndFractionalPan(t *testing.T) {
	var outputs [][]float32
	for _, translation := range []float64{0, 999999} {
		m, err := canvas.NewMesh([]canvas.Vertex{{Position: canvas.Point{X: translation + .01}}, {Position: canvas.Point{X: translation + .02}}, {Position: canvas.Point{X: translation + .03, Y: .01}}}, []uint32{0, 1, 2})
		if err != nil {
			t.Fatal(err)
		}
		positions, _, origin := drawingGeometry(m)
		center := translation + .015
		screen := make([]float32, len(positions))
		for i := 0; i < len(positions); i += 2 {
			screen[i] = (positions[i] + float32(origin.X-center)) * 100
			screen[i+1] = positions[i+1] * 100
		}
		outputs = append(outputs, screen)
	}
	for i, v := range outputs[0] {
		if math.Abs(float64(v-outputs[1][i])) > 1e-5 {
			t.Fatal("translated geometry/fractional camera lost", outputs)
		}
	}
}
func TestLogicalDeviceAdmissionBeforeAllocation(t *testing.T) {
	m, _ := canvas.NewMesh([]canvas.Vertex{{}, {}, {}}, []uint32{0, 1, 2})
	limits := wgpu.DefaultLimits()
	limits.MaxTextureDimension2D = 1024
	limits.MaxBufferSize = 64
	if drawingAdmission(image.Pt(1025, 10), m, 0, limits) == nil || drawingAdmission(image.Pt(10, 10), m, 65, limits) == nil {
		t.Fatal("device limit ignored")
	}
	if err := drawingAdmission(image.Pt(10, 10), m, 64, limits); err != nil {
		t.Fatal(err)
	}
}

func TestLateCanvasInputIsBoundToItsMountedLifetime(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	for _, id := range []string{"one", "two"} {
		a.Open(app.Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	}
	m, _ := canvas.NewMesh(nil, nil)
	view := canvas.Viewport{Width: 200, Height: 100, Zoom: 1}
	picked := 0
	g := &MeshCanvas{Owner: a, Window: "one", View: &view, Mesh: m, Pick: func(canvas.Point) { picked++ }}
	g.Mount(core.NewBody("one"))
	old := g.widget
	before := view
	emit := func() {
		s := events.NewScroll(image.Pt(20, 30), math32.Vec2(0, 50), 0)
		s.SetLocalOff(image.Point{})
		old.HandleEvent(s)
		click := events.NewMouse(events.MouseDown, events.Left, image.Pt(20, 30), 0)
		click.SetLocalOff(image.Point{})
		old.HandleEvent(click)
	}
	a.CloseWindow(fixtureWindowRef(a, "one"))
	emit()
	if view != before || picked != 0 {
		t.Fatal("cancelled-window input escaped", view, picked)
	}
	old.Destroy()
	g.Window = "two"
	g.Mount(core.NewBody("two"))
	old.ownerEvent(func() { picked++; view.Zoom = 2 })
	if view != before || picked != 0 {
		t.Fatal("retired callback changed remounted model")
	}
	g.Window = "one"
	if g.NativeWindow() != nil || g.Sync(g.widget) == nil {
		t.Fatal("mounted identity change accepted")
	}
}
func TestBorrowedMeshSourceRefusesRetiredGenerationAndTexture(t *testing.T) {
	old, newTexture := &gpu.Texture{}, &gpu.Texture{}
	r := &meshResources{ready: true, frame: &gpu.RenderTexture{Frames: []*gpu.Texture{newTexture}}, stats: DrawingStats{Generation: 2}}
	// An unconfigured drawer cannot submit any texture; stale sources must return
	// before touching it. This exercises a real Source.Draw boundary without GPU.
	drawer := &composer.ComposerDrawer{Drawer: &gpudraw.Drawer{}}
	source := &meshSource{resources: r, texture: old, generation: 1}
	source.Draw(drawer)
	source.generation = 2
	source.Draw(drawer)
	source.texture = newTexture
	r.ready = false
	source.Draw(drawer)
	r.ready = true
	r.frame = nil
	source.Draw(drawer)
}

// Render preparation failures must use the same owner as native-input errors.
// In particular, replacing the callback cannot overtake an admitted report.
func TestDrawingErrorCallbackSharesOwnerAndReleasesResources(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(app.Window{ID: "drawing", Title: "test", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	expected := errors.New("prepare fixture")
	view := canvas.Viewport{Width: 200, Height: 100, Zoom: 1}
	g := &MeshCanvas{Owner: a, Window: "drawing", View: &view, Prepare: func(canvas.Viewport) (*canvas.Mesh, error) { return nil, expected }}
	g.Mount(core.NewBody("drawing"))
	g.widget.Geom.Size.Actual.Content = math32.Vec2(100, 80)
	entered, release, rendered := make(chan struct{}), make(chan struct{}), make(chan struct{})
	callbackErr := make(chan error, 1)
	g.OnError = func(err error) {
		if !errors.Is(err, expected) {
			callbackErr <- err
		}
		if !g.widget.resources.mu.TryLock() {
			callbackErr <- errors.New("resource lock held during error callback")
		} else {
			g.widget.resources.mu.Unlock()
		}
		close(entered)
		<-release
	}
	go func() { g.widget.Render(); close(rendered) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("error callback not delivered")
	}
	attempted, replaced := make(chan struct{}), make(chan struct{})
	go func() { close(attempted); a.Do(func() { g.OnError = nil }); close(replaced) }()
	<-attempted
	escaped := false
	select {
	case <-replaced:
		escaped = true
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	for _, done := range []<-chan struct{}{rendered, replaced} {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("owner/callback did not finish")
		}
	}
	if escaped {
		t.Fatal("callback replacement overtook owner-held delivery")
	}
	select {
	case err := <-callbackErr:
		t.Fatal(err)
	default:
	}
}
