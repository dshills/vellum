package cogent

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/gpu/gpudraw"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/composer"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed mesh_canvas.wgsl
var meshShader string

// MeshCanvas is a shared-device retained 2D GPU extension. All fields and methods
// belong to Owner.Do/PostWait. Mesh is immutable; replace it to change geometry.
// View uses actual content pixels (not dp), world coordinates and world stroke widths.
// Prepare optionally derives visible geometry when View or Revision changes; it
// runs on the UI owner before GPU locking and must not call this adapter's methods.
// Build expensive indexes/reductions on a worker before publishing them to Prepare.
// Background must be opaque or transparent black (the zero default).
// OnError and Pick run synchronously on the UI owner. They must not nest Do/Drain.
// OnError runs after native/resource locks are released and may query this adapter.
// A mutable adapter mounts in only one live widget; meshes can be shared freely.
type MeshCanvas struct {
	Owner           *app.Application
	Window          string
	View            *canvas.Viewport
	Mesh            *canvas.Mesh
	Prepare         func(canvas.Viewport) (*canvas.Mesh, error)
	Revision        uint64
	Background      color.NRGBA
	MaxGPUBytes     uint64 // zero: 512 MiB, for this renderer's tracked resources only
	OnError         func(error)
	Pick            func(canvas.Point)
	widget          *meshWidget
	prepared        bool
	prepareView     canvas.Viewport
	prepareRevision uint64
}

// DrawingStats reports submitted work, not GPU timings or process-wide memory.
type DrawingStats struct {
	Ready                                                                                     bool
	RenderVisits, Frames, MeshUploads, MeshUploadBytes, UniformUploadBytes, SnapshotReadbacks uint64
	Generation, TargetBytes, MeshBytes, ReadbackBytes                                         uint64
	SubmittedTriangles                                                                        int
	Size                                                                                      image.Point
}

const defaultDrawingBudget = 512 << 20

func (g *MeshCanvas) validate() error {
	if g == nil || g.Owner == nil || g.View == nil {
		return fmt.Errorf("drawing: owner and viewport required")
	}
	if _, err := g.View.Bounds(); err != nil {
		return err
	}
	if g.widget != nil && !g.widget.retired.Load() && (g.Owner != g.widget.application || g.Window != g.widget.windowID) {
		return fmt.Errorf("drawing: mounted owner/window identity cannot change")
	}
	if g.Mesh == nil && g.Prepare == nil {
		return fmt.Errorf("drawing: mesh or preparation callback required")
	}
	if g.Background.A != 255 && g.Background != (color.NRGBA{}) {
		return fmt.Errorf("drawing: background must be opaque or transparent black")
	}
	return nil
}
func (g *MeshCanvas) Mount(parent tree.Node) core.Widget {
	if err := g.validate(); err != nil {
		g.report(err)
		return nil
	}
	if g.widget != nil && !g.widget.retired.Load() {
		g.report(errors.New("drawing: adapter is already mounted"))
		return nil
	}
	g.prepared = false
	g.widget = tree.New[meshWidget](parent)
	g.widget.owner = g
	g.widget.application = g.Owner
	g.widget.windowID = g.Window
	g.widget.Styler(func(s *styles.Style) {
		s.Min.Set(units.Dp(240), units.Dp(200))
		s.Grow.Set(1, 1)
		s.SetAbilities(true, abilities.Focusable, abilities.Clickable, abilities.Slideable)
	})
	mounted := g.widget
	mounted.OnFirst(events.Scroll, func(e events.Event) {
		s, ok := e.(*events.MouseScroll)
		if !ok {
			return
		}
		mounted.ownerEvent(func() {
			p := e.Pos().Sub(mounted.Geom.Pos.Content.ToPoint())
			if err := g.View.ZoomAt(canvas.Point{X: float64(p.X), Y: float64(p.Y)}, math.Exp(-float64(s.Delta.Y)*.003)); err != nil {
				g.report(err)
				return
			}
			g.Owner.Invalidate(g.Window)
		})
		e.SetHandled()
	})
	mounted.OnFirst(events.SlideMove, func(e events.Event) {
		mounted.ownerEvent(func() {
			p := e.PrevDelta()
			if err := g.View.Pan(float64(p.X), float64(p.Y)); err != nil {
				g.report(err)
				return
			}
			g.Owner.Invalidate(g.Window)
		})
		e.SetHandled()
	})
	mounted.OnFirst(events.MouseDown, func(e events.Event) {
		mounted.ownerEvent(func() {
			mounted.SetFocus()
			if g.Pick != nil {
				p := e.Pos().Sub(mounted.Geom.Pos.Content.ToPoint())
				g.Pick(g.View.World(canvas.Point{X: float64(p.X), Y: float64(p.Y)}))
			}
		})
		e.SetHandled()
	})
	mounted.OnFirst(events.KeyChord, func(e events.Event) {
		dx, dy := 0., 0.
		switch strings.ToLower(string(e.KeyChord())) {
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
		mounted.ownerEvent(func() {
			if err := g.View.Pan(dx, dy); err != nil {
				g.report(err)
				return
			}
			g.Owner.Invalidate(g.Window)
		})
		e.SetHandled()
	})
	return g.widget
}
func (g *MeshCanvas) Sync(core.Widget) error {
	if err := g.validate(); err != nil {
		return err
	}
	if g.widget == nil || g.widget.retired.Load() {
		return ErrFrameNotReady
	}
	g.widget.NeedsRender()
	return nil
}
func (g *MeshCanvas) report(err error) {
	if g != nil && g.OnError != nil {
		g.OnError(err)
	}
}
func (g *MeshCanvas) NativeWindow() system.Window {
	if g == nil || g.Owner == nil || g.widget == nil || g.widget.retired.Load() || g.Owner != g.widget.application || g.Window != g.widget.windowID {
		return nil
	}
	ctx, err := g.Owner.WindowContext(g.Window)
	if err != nil || ctx.Err() != nil {
		return nil
	}
	published := g.widget.native.Load()
	if published == nil {
		return nil
	}
	return published.window
}
func (g *MeshCanvas) budget() uint64 {
	if g.MaxGPUBytes != 0 {
		return g.MaxGPUBytes
	}
	return defaultDrawingBudget
}
func (g *MeshCanvas) RequestFrameAfter(delay time.Duration) error {
	if delay < 0 {
		return fmt.Errorf("drawing: negative frame delay")
	}
	w := g.NativeWindow()
	if w == nil {
		return ErrFrameNotReady
	}
	p, ok := w.(interface{ RequestPaintAfter(time.Duration) })
	if !ok {
		return fmt.Errorf("drawing: requested paint unavailable")
	}
	p.RequestPaintAfter(delay)
	return nil
}
func (g *MeshCanvas) RebuildResources() error {
	w := g.NativeWindow()
	if w == nil || !w.Lock() {
		return ErrFrameNotReady
	}
	defer w.Unlock()
	r := &g.widget.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := drawingDeviceError(w); err != nil {
		return err
	}
	r.release()
	g.widget.NeedsRender()
	g.Owner.Invalidate(g.Window)
	return nil
}
func (g *MeshCanvas) Stats() (DrawingStats, error) {
	w := g.NativeWindow()
	if w == nil || !w.Lock() {
		return DrawingStats{}, ErrFrameNotReady
	}
	defer w.Unlock()
	r := &g.widget.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	stats := r.stats
	stats.Ready = r.ready
	return stats, nil
}

// Snapshot copies the last completed frame. It does not flush pending layout,
// preparation or model changes and never reads widget geometry. Transparent GPU
// colors are unpremultiplied in linear space before returning image.NRGBA.
func (g *MeshCanvas) Snapshot() (*image.NRGBA, error) {
	w := g.NativeWindow()
	if w == nil || !w.Lock() {
		return nil, ErrFrameNotReady
	}
	defer w.Unlock()
	r := &g.widget.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := drawingDeviceError(w); err != nil {
		return nil, err
	}
	if !r.ready || r.frame == nil {
		return nil, ErrFrameNotReady
	}
	tex := r.frame.CurrentFrame()
	size := r.frame.Size()
	bytes := uint64((size.X*4+255)/256*256) * uint64(size.Y)
	if r.stats.TargetBytes+r.stats.MeshBytes+16+bytes > g.budget() {
		return nil, fmt.Errorf("drawing: readback exceeds GPU resource budget")
	}
	if bytes > r.limits.MaxBufferSize {
		return nil, fmt.Errorf("drawing: readback exceeds device buffer limit")
	}
	if err := tex.ConfigReadBuffer(); err != nil {
		return nil, err
	}
	r.stats.ReadbackBytes = bytes
	cmd, err := r.system.NewCommandEncoder()
	if err != nil {
		return nil, err
	}
	defer cmd.Release()
	if err := tex.CopyToReadBuffer(cmd); err != nil {
		return nil, err
	}
	buffer, err := cmd.TryFinish(nil)
	if err != nil {
		return nil, err
	}
	defer buffer.Release()
	r.system.Device().Queue.Submit(buffer)
	img, err := tex.ReadGoImage()
	if err != nil {
		return nil, err
	}
	straightSnapshot(img)
	r.stats.SnapshotReadbacks++
	return img, nil
}
func drawingDeviceError(w system.Window) error {
	d := nativeDevice(w)
	if d == nil {
		return ErrFrameNotReady
	}
	if err := d.LossError(); err != nil {
		return fmt.Errorf("%w: %v", ErrDeviceLost, err)
	}
	return nil
}

type drawingWindow struct{ window system.Window }

type meshWidget struct {
	core.WidgetBase
	owner       *MeshCanvas
	application *app.Application
	windowID    string
	retired     atomic.Bool
	resources   meshResources
	native      atomic.Pointer[drawingWindow]
}

func (w *meshWidget) Init()  { w.WidgetBase.Init() }
func (w *meshWidget) OnAdd() { w.WidgetBase.OnAdd(); w.Scene.AddDirectRender(w) }
func (w *meshWidget) Destroy() {
	w.retired.Store(true)
	var unlock func()
	if w.Scene != nil {
		if native := w.native.Load(); native != nil && native.window.Lock() {
			unlock = native.window.Unlock
		}
		w.Scene.DeleteDirectRender(w)
	}
	if unlock != nil {
		defer unlock()
	}
	w.resources.mu.Lock()
	w.resources.release()
	w.resources.mu.Unlock()
	w.WidgetBase.Destroy()
}
func (w *meshWidget) Render() {
	if w.owner == nil || w.retired.Load() {
		return
	}
	g := w.owner
	var renderErr error
	_ = w.application.Do(func() {
		// Registered before GPU lock defers: report only after those locks release,
		// while model state and callback replacement still belong to this owner.
		defer func() {
			if renderErr != nil {
				g.report(renderErr)
			}
		}()
		if w.retired.Load() || g.widget != w {
			return
		}
		ctx, err := w.application.WindowContext(w.windowID)
		if err != nil || ctx.Err() != nil {
			return
		}
		w.resources.mu.Lock()
		w.resources.stats.RenderVisits++
		w.resources.mu.Unlock()
		if err := g.validate(); err != nil {
			renderErr = err
			return
		}
		// Actual full content allocation, not the scroll-clipped ContentBBox.
		full := math32.RectFromPosSizeMax(w.Geom.Pos.Content, w.Geom.Size.Actual.Content)
		size := full.Size()
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		if err := g.View.SetSize(float64(size.X), float64(size.Y)); err != nil {
			renderErr = err
			return
		}
		if err := g.validate(); err != nil {
			renderErr = err
			return
		}
		if g.Prepare != nil && (!g.prepared || g.prepareView != *g.View || g.prepareRevision != g.Revision) {
			mesh, err := g.Prepare(*g.View)
			if err != nil {
				renderErr = err
				return
			}
			if mesh == nil {
				renderErr = errors.New("drawing: preparation returned nil mesh")
				return
			}
			g.Mesh = mesh
			g.prepared = true
			g.prepareView = *g.View
			g.prepareRevision = g.Revision
		}
		published := w.native.Load()
		if published == nil {
			rw := w.Scene.RenderWindow()
			if rw == nil || rw.SystemWindow == nil {
				return
			}
			published = &drawingWindow{rw.SystemWindow}
			w.native.Store(published)
		}
		window := published.window
		if window == nil || !window.Lock() {
			return
		}
		defer window.Unlock()
		if err := drawingDeviceError(window); err != nil {
			renderErr = err
			return
		}
		r := &w.resources
		r.mu.Lock()
		defer r.mu.Unlock()
		renderErr = r.render(nativeSurface(window), size, g.Mesh, *g.View, g.Background, g.budget())
	})
}
func (w *meshWidget) RenderSource(op draw.Op) composer.Source {
	if !w.IsVisible() || w.retired.Load() {
		return nil
	}
	r := &w.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.ready || r.frame == nil {
		return nil
	}
	full := math32.RectFromPosSizeMax(w.Geom.Pos.Content, w.Geom.Size.Actual.Content)
	dest, source := drawingClip(full, w.Geom.ContentBBox, w.Scene.SceneGeom.Pos, r.frame.Size())
	if source.Empty() {
		return nil
	}
	return &meshSource{resources: r, texture: r.frame.CurrentFrame(), generation: r.stats.Generation, dest: dest, source: source}
}
func drawingClip(full, clip image.Rectangle, scenePos, size image.Point) (image.Rectangle, image.Rectangle) {
	full.Max = full.Min.Add(size)
	visible := full.Intersect(clip)
	if visible.Empty() {
		return image.Rectangle{}, image.Rectangle{}
	}
	return visible.Add(scenePos), visible.Sub(full.Min)
}

type meshSource struct {
	resources    *meshResources
	texture      *gpu.Texture
	generation   uint64
	dest, source image.Rectangle
}

func (s *meshSource) Draw(c composer.Composer) {
	cd, ok := c.(*composer.ComposerDrawer)
	if !ok {
		return
	}
	drawer, ok := cd.Drawer.(*gpudraw.Drawer)
	if !ok {
		return
	}
	r := s.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.ready || r.frame == nil || r.stats.Generation != s.generation || r.frame.CurrentFrame() != s.texture {
		return
	}
	drawer.UseTexture(s.texture)
	drawer.CopyUsed(s.dest.Min, s.source, draw.Over, false)
}

type meshResources struct {
	origin                                canvas.Point
	limits                                wgpu.Limits
	haveLimits                            bool
	mu                                    sync.Mutex
	frame                                 *gpu.RenderTexture
	system                                *gpu.GraphicsSystem
	pipeline                              *gpu.GraphicsPipeline
	positions, colors, indices, transform *gpu.Value
	mesh                                  *canvas.Mesh
	view                                  canvas.Viewport
	background                            color.NRGBA
	ready                                 bool
	stats                                 DrawingStats
}

func (r *meshResources) release() {
	r.ready = false
	if r.system != nil {
		r.system.Release()
	}
	if r.frame != nil {
		r.frame.Release()
	}
	r.frame = nil
	r.system = nil
	r.pipeline = nil
	r.positions = nil
	r.colors = nil
	r.indices = nil
	r.transform = nil
	r.mesh = nil
	r.stats.Generation++
	r.stats.TargetBytes = 0
	r.stats.MeshBytes = 0
	r.stats.ReadbackBytes = 0
	r.stats.Size = image.Point{}
	r.stats.SubmittedTriangles = 0
}
func drawingBytes(size image.Point, m *canvas.Mesh) (target, mesh uint64, error error) {
	if size.X < 1 || size.Y < 1 || size.X > 16384 || size.Y > 16384 || m == nil {
		return 0, 0, fmt.Errorf("drawing: invalid GPU allocation")
	}
	n, ix := m.Counts()
	if ix == 0 {
		n = 0
	}
	return uint64(size.X) * uint64(size.Y) * 20, uint64(n)*24 + uint64(ix)*4, nil
}
func (r *meshResources) configure(surface *gpu.Surface, size image.Point) error {
	r.frame = gpu.NewRenderTexture(surface.GPU, surface.Device(), image.Pt(1, 1), 1, gpu.UndefinedType)
	r.frame.Format.SetMultisample(4)
	r.frame.Render().Format.SetMultisample(4)
	if err := resizeDrawingTarget(r.frame, size); err != nil {
		r.release()
		return err
	}
	r.system = gpu.NewGraphicsSystem(surface.GPU, "vellum.mesh", r.frame)
	r.pipeline = r.system.AddGraphicsPipeline("flat")
	r.pipeline.SetCullMode(wgpu.CullModeNone)
	r.pipeline.SetAlphaBlend(true)
	shader := r.pipeline.AddShader("mesh")
	if err := shader.OpenCode(meshShader); err != nil {
		r.release()
		return err
	}
	r.pipeline.AddEntry(shader, gpu.VertexShader, "vs_main")
	r.pipeline.AddEntry(shader, gpu.FragmentShader, "fs_main")
	vertices := r.system.Vars().AddVertexGroup()
	uniforms := r.system.Vars().AddGroup(gpu.Uniform)
	pos := vertices.Add("Pos", gpu.Float32Vector2, 0, gpu.VertexShader)
	clr := vertices.Add("Color", gpu.Float32Vector4, 0, gpu.VertexShader)
	ix := vertices.Add("Index", gpu.Uint32, 0, gpu.VertexShader)
	ix.Role = gpu.Index
	transform := uniforms.AddStruct("View", 16, 1, gpu.VertexShader)
	vertices.SetNValues(1)
	uniforms.SetNValues(1)
	// Config vars directly so pipeline configuration errors are not discarded.
	r.system.Vars().Config(r.system.Device())
	if err := r.pipeline.Config(true); err != nil {
		r.release()
		return err
	}
	r.positions = pos.Values.Values[0]
	r.colors = clr.Values.Values[0]
	r.indices = ix.Values.Values[0]
	r.transform = transform.Values.Values[0]
	r.stats.Generation++
	return nil
}
func (r *meshResources) render(surface *gpu.Surface, size image.Point, m *canvas.Mesh, view canvas.Viewport, bg color.NRGBA, budget uint64) error {
	target, mesh, err := drawingBytes(size, m)
	if err != nil {
		return err
	}
	// A previous snapshot staging buffer persists until resize/rebuild.
	readback := r.stats.ReadbackBytes
	if r.frame != nil && r.frame.Size() != size {
		readback = 0
	}
	if target+mesh+16+readback > budget {
		return fmt.Errorf("drawing: GPU resource budget exceeded")
	}
	if !r.haveLimits {
		r.limits = surface.Device().Device.GetLimits()
		r.haveLimits = true
	}
	if err := drawingAdmission(size, m, readback, r.limits); err != nil {
		return err
	}
	if r.frame == nil {
		if err := r.configure(surface, size); err != nil {
			return err
		}
	}
	resized := r.frame.Size() != size
	if resized {
		r.ready = false
		if err := resizeDrawingTarget(r.frame, size); err != nil {
			r.release()
			return err
		}
		r.stats.Generation++
		r.stats.ReadbackBytes = 0
	}
	changed := r.mesh != m
	if !changed && !resized && r.ready && r.view == view && r.background == bg {
		return nil
	}
	// Failed upload/submission must never expose a partially updated frame.
	r.ready = false
	if changed {
		r.mesh = nil
		ix := m.Indices()
		// Empty meshes clear normally without binding zero-length vertex buffers.
		if len(ix) > 0 {
			positions, colors, origin := drawingGeometry(m)
			r.origin = origin
			if err := gpu.SetValueFrom(r.positions, positions); err != nil {
				return err
			}
			if err := gpu.SetValueFrom(r.colors, colors); err != nil {
				return err
			}
			if err := gpu.SetValueFrom(r.indices, ix); err != nil {
				return err
			}
		} else {
			r.positions.Release()
			r.colors.Release()
			r.indices.Release()
		}
		if len(ix) > 0 {
			r.stats.MeshUploads++
			r.stats.MeshUploadBytes += mesh
		}
		r.mesh = m
	}
	transform := []float32{float32(2 * view.Zoom / view.Width), float32(-2 * view.Zoom / view.Height), float32(r.origin.X - view.Center.X), float32(r.origin.Y - view.Center.Y)}
	if err := gpu.SetValueFrom(r.transform, transform); err != nil {
		r.mesh = nil
		return err
	}
	r.stats.UniformUploadBytes += 16
	r.system.SetClearColor(bg)
	pass, err := r.system.BeginRenderPass()
	if err != nil {
		r.mesh = nil
		return err
	}
	if pass == nil {
		r.mesh = nil
		return ErrFrameNotReady
	}
	_, count := m.Counts()
	var bindErr error
	if count > 0 {
		bindErr = r.pipeline.BindPipeline(pass)
		if bindErr == nil {
			r.pipeline.BindDrawIndexed(pass)
		}
	}
	pass.End()
	encoder := r.system.CommandEncoder
	if err := r.system.SubmitRender(pass); err != nil {
		encoder.Release()
		r.mesh = nil
		return err
	}
	if bindErr != nil {
		r.mesh = nil
		return bindErr
	}
	r.view = view
	r.background = bg
	r.ready = true
	r.stats.Frames++
	r.stats.TargetBytes = target
	r.stats.MeshBytes = mesh
	r.stats.Size = size
	r.stats.SubmittedTriangles = count / 3
	return nil
}

var linearBytes = func() (table [256]float32) {
	for i := range table {
		v := float64(i) / 255
		if v <= .04045 {
			table[i] = float32(v / 12.92)
		} else {
			table[i] = float32(math.Pow((v+.055)/1.055, 2.4))
		}
	}
	return
}()

func linearByte(b uint8) float32 { return linearBytes[b] }

func srgbByte(v float64) uint8 {
	v = max(0, min(1, v))
	if v <= .0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - .055
	}
	return uint8(math.Round(v * 255))
}
func straightSnapshot(img *image.NRGBA) {
	for y := 0; y < img.Rect.Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			i := y*img.Stride + x*4
			a := img.Pix[i+3]
			if a == 255 {
				continue
			}
			if a == 0 {
				img.Pix[i] = 0
				img.Pix[i+1] = 0
				img.Pix[i+2] = 0
				continue
			}
			for c := 0; c < 3; c++ {
				img.Pix[i+c] = srgbByte(float64(linearByte(img.Pix[i+c])) * 255 / float64(a))
			}
		}
	}
}

// The one-pixel public constructor establishes the borrowed device. All actual
// target/MSAA allocation and resizing goes through checked public texture APIs;
// RenderTexture.SetSize/ConfigFrames otherwise discard those allocation errors.
func resizeDrawingTarget(rt *gpu.RenderTexture, size image.Point) error {
	rt.ReleaseFrames()
	rt.Format.Size = size
	rt.Render().Format.Size = size
	frame := gpu.NewTexture(rt.Device())
	if err := frame.ConfigRenderTexture(rt.Device(), &rt.Format); err != nil {
		frame.Release()
		return err
	}
	rt.Frames = []*gpu.Texture{frame}
	if err := rt.Render().Multi.ConfigMulti(rt.Device(), &rt.Format); err != nil {
		return err
	}
	rt.Render().HasMulti = true
	return nil
}
func drawingAdmission(size image.Point, m *canvas.Mesh, readback uint64, limits wgpu.Limits) error {
	if uint64(size.X) > uint64(limits.MaxTextureDimension2D) || uint64(size.Y) > uint64(limits.MaxTextureDimension2D) {
		return fmt.Errorf("drawing: target exceeds logical device texture limit")
	}
	vertices, indices := m.Counts()
	if indices == 0 {
		vertices = 0
	}
	for _, bytes := range []uint64{uint64(vertices) * 8, uint64(vertices) * 16, uint64(indices) * 4, 16, readback} {
		if bytes > limits.MaxBufferSize {
			return fmt.Errorf("drawing: buffer exceeds logical device limit")
		}
	}
	return nil
}
func drawingGeometry(m *canvas.Mesh) (positions, colors []float32, origin canvas.Point) {
	v := m.Vertices()
	if len(v) > 0 {
		origin = v[0].Position
	}
	positions = make([]float32, 0, len(v)*2)
	colors = make([]float32, 0, len(v)*4)
	for _, vertex := range v {
		positions = append(positions, float32(vertex.Position.X-origin.X), float32(vertex.Position.Y-origin.Y))
		c := vertex.Color
		a := float32(c.A) / 255
		colors = append(colors, linearByte(c.R)*a, linearByte(c.G)*a, linearByte(c.B)*a, a)
	}
	return
}

func (w *meshWidget) ownerEvent(fn func()) {
	_ = w.application.Do(func() {
		g := w.owner
		if w.retired.Load() || g == nil || g.widget != w || g.Owner != w.application || g.Window != w.windowID {
			return
		}
		ctx, err := w.application.WindowContext(w.windowID)
		if err != nil || ctx.Err() != nil {
			return
		}
		if err := g.validate(); err != nil {
			g.report(err)
			return
		}
		fn()
	})
}
