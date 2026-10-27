package cogent

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	"github.com/oliverbestmann/webgpu/wgpu"
)

const GPUContextVersion = 1

// GPUProgram supplies advanced shared-device rendering. Callbacks are captured at
// mount; one program instance can mount in only one live widget. Prepare encodes
// uploads/offscreen/compute work before the framework opens the main pass. Draw
// records into that pass. Neither callback may call UI/adapter/window methods,
// nest Do/Drain, launch concurrent GPU work, or finish/end/release borrowed handles.
// Release frees program-owned pipelines/layouts/groups/views; helper-owned buffers
// and textures are released by the framework afterwards. It also runs after a
// failed initial Prepare, and must tolerate partial initialization.
type GPUProgram struct {
	Prepare func(GPUFrame) error
	Draw    func(GPUFrame, *wgpu.RenderPassEncoder) error
	Release func()
	mounted atomic.Bool
}

// GPUFrame describes a callback-scoped borrowed rendering context. Color output
// is linear premultiplied RGBA, stored in an sRGB target. Clip is content-local
// physical pixels; Size is the full allocation, not its visible clip. Resource
// generation changes on rebuild/failure/device change; target generation also
// changes on resize. Encoder is supplied only to Prepare. Device/Queue belong to the window: do not destroy, release,
// present or independently submit them. Raw allocations/writes bypass accounting.
// Encoder and Resources expire on callback return; persistent helper handles are
// borrowed for this resource generation and must not be released by the program.
type GPUFrame struct {
	Version                              int
	Size                                 image.Point
	Clip                                 image.Rectangle
	DpScale                              float32
	Format, DepthFormat                  wgpu.TextureFormat
	Samples                              uint32
	ResourceGeneration, TargetGeneration uint64
	Device                               *wgpu.Device
	Queue                                *wgpu.Queue
	Limits                               wgpu.Limits
	Encoder                              *wgpu.CommandEncoder
	Resources                            *GPUResources
}

// GPUInput runs on the UI owner, after GPU locks have been released. Input returns
// true only for consumed events; false preserves key/scroll propagation. It may change
// the model, advance Revision and invalidate the window, but must not nest Do.
type GPUInput struct {
	Position, Delta image.Point
	Scroll          float32
	Key             string
}

// GPUView mounts an advanced GPUProgram as an ordinary UI extension, using the
// window device and a retained 4x-MSAA target. Fields/methods belong to Owner.Do.
// Owner, Window, Program and Depth are fixed while mounted. Revision must advance
// for every program/model change. OnError can query Stats after locks release.
// MaxGPUBytes covers targets, readback and helper resources, not raw allocations,
// pipeline/driver overhead or process-wide memory. Background is opaque or zero.
type GPUView struct {
	Owner            *app.Application
	Window           string
	Program          *GPUProgram
	Revision         uint64
	Depth            bool
	Background       color.NRGBA
	MaxGPUBytes      uint64
	MaxSnapshotBytes uint64
	OnError          func(error)
	Input            func(GPUInput) bool
	// InputEvent opts into typed input in place of legacy Input; captured at mount.
	InputEvent     func(GPUEvent) bool
	widget         *gpuViewWidget
	snapshotActive atomic.Bool
}

type GPUViewStats struct {
	Ready                                                                                                        bool
	SnapshotPending                                                                                              bool
	AsyncReadbackBytes, AsyncSnapshotRequests, AsyncSnapshotCompleted, AsyncSnapshotCanceled, AsyncSnapshotPolls uint64
	RenderVisits, Frames, Prepares, Draws, SnapshotReadbacks                                                     uint64
	ResourceGeneration, TargetGeneration                                                                         uint64
	TargetBytes, BufferBytes, TextureBytes, ReadbackBytes                                                        uint64
	HelperUploads, HelperUploadBytes, BufferAllocations, TextureAllocations                                      uint64
	Size                                                                                                         image.Point
}

func (g *GPUView) validate() error {
	if g == nil || g.Owner == nil || g.Program == nil {
		return errors.New("GPU view: owner and program required")
	}
	if w := g.widget; w != nil && !w.retired.Load() {
		if g.Owner != w.application || g.Window != w.windowID || g.Program != w.program || g.Depth != w.depth {
			return errors.New("GPU view: mounted identity/depth cannot change")
		}
	} else if g.Program.Draw == nil {
		return errors.New("GPU view: Draw callback required")
	}
	if g.Background.A != 255 && g.Background != (color.NRGBA{}) {
		return errors.New("GPU view: background must be opaque or transparent black")
	}
	return nil
}
func (g *GPUView) report(err error) {
	if g != nil && g.OnError != nil {
		g.OnError(err)
	}
}
func (g *GPUView) budget() uint64 {
	if g.MaxGPUBytes != 0 {
		return g.MaxGPUBytes
	}
	return defaultDrawingBudget
}
func (g *GPUView) Mount(parent tree.Node) core.Widget {
	if err := g.validate(); err != nil {
		g.report(err)
		return nil
	}
	ref, err := captureMountScope(parent, g.Owner, g.Window)
	if err != nil {
		g.report(err)
		return nil
	}
	if g.widget != nil && !g.widget.retired.Load() {
		g.report(errors.New("GPU view: already mounted"))
		return nil
	}
	if !g.Program.mounted.CompareAndSwap(false, true) {
		g.report(errors.New("GPU view: program already mounted"))
		return nil
	}
	mounted := false
	defer func() {
		if !mounted {
			g.Program.mounted.Store(false)
		}
	}()
	w := tree.New[gpuViewWidget](parent)
	w.owner, w.application, w.windowID, w.program, w.depth = g, g.Owner, g.Window, g.Program, g.Depth
	w.windowRef = ref
	w.resources.prepare, w.resources.draw, w.resources.cleanup = g.Program.Prepare, g.Program.Draw, g.Program.Release
	g.widget = w
	w.Styler(func(s *styles.Style) {
		s.Min.Set(units.Dp(240), units.Dp(200))
		s.Grow.Set(1, 1)
		s.SetAbilities(true, abilities.Focusable, abilities.Clickable, abilities.Slideable)
	})
	if g.InputEvent != nil {
		w.onInput = g.InputEvent
		w.installGPUInput()
		mounted = true
		return w
	}
	dispatch := func(e events.Event, input func() GPUInput) {
		handled := false
		w.ownerEvent(func() {
			if e.Type() == events.MouseDown {
				w.SetFocus()
				handled = true
			}
			if g.Input != nil {
				handled = g.Input(input()) || handled
			}
		})
		if handled {
			e.SetHandled()
		}
	}
	w.OnFirst(events.MouseDown, func(e events.Event) {
		dispatch(e, func() GPUInput { return GPUInput{Position: e.Pos().Sub(w.Geom.Pos.Content.ToPoint())} })
	})
	w.OnFirst(events.SlideMove, func(e events.Event) {
		dispatch(e, func() GPUInput {
			return GPUInput{Position: e.Pos().Sub(w.Geom.Pos.Content.ToPoint()), Delta: e.PrevDelta()}
		})
	})
	w.OnFirst(events.Scroll, func(e events.Event) {
		if s, ok := e.(*events.MouseScroll); ok {
			dispatch(e, func() GPUInput { return GPUInput{Scroll: s.Delta.Y} })
		}
	})
	w.OnFirst(events.KeyChord, func(e events.Event) { dispatch(e, func() GPUInput { return GPUInput{Key: string(e.KeyChord())} }) })
	mounted = true
	return w
}
func (g *GPUView) Sync(core.Widget) error {
	if err := g.validate(); err != nil {
		return err
	}
	if g.widget == nil || g.widget.retired.Load() {
		return ErrFrameNotReady
	}
	g.widget.NeedsRender()
	return nil
}
func (g *GPUView) NativeWindow() system.Window {
	if g == nil || g.Owner == nil || g.widget == nil || g.widget.retired.Load() || g.Owner != g.widget.application || g.Window != g.widget.windowID {
		return nil
	}
	ctx, err := scopedWindowContext(g.Owner, g.widget.windowRef)
	if err != nil || ctx.Err() != nil {
		return nil
	}
	if p := g.widget.native.Load(); p != nil {
		return p.window
	}
	return nil
}
func (g *GPUView) Stats() (GPUViewStats, error) {
	w := g.NativeWindow()
	if w == nil || !w.Lock() {
		return GPUViewStats{}, ErrFrameNotReady
	}
	defer w.Unlock()
	r := &g.widget.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.stats
	s.Ready = r.ready
	s.SnapshotPending = g.snapshotActive.Load()
	return s, nil
}
func (g *GPUView) RebuildResources() error {
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
	r.reset()
	g.widget.NeedsRender()
	return g.Owner.Invalidate(g.Window)
}
func (g *GPUView) RequestFrameAfter(delay time.Duration) error {
	if delay < 0 {
		return errors.New("GPU view: negative delay")
	}
	w := g.NativeWindow()
	if w == nil {
		return ErrFrameNotReady
	}
	p, ok := w.(interface{ RequestPaintAfter(time.Duration) })
	if !ok {
		return errors.New("GPU view: requested paint unavailable")
	}
	p.RequestPaintAfter(delay)
	return nil
}

// Snapshot synchronously reads the last completed offscreen frame; it does not
// flush pending layout/model updates and is never part of normal composition.
func (g *GPUView) Snapshot() (*image.NRGBA, error) {
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
	if !r.ready || r.target == nil {
		return nil, ErrFrameNotReady
	}
	size := r.target.Size()
	bytes := uint64((size.X*4+255)/256*256) * uint64(size.Y)
	if err := gpuBudget(g.budget(), r.stats.TargetBytes, r.stats.BufferBytes, r.stats.TextureBytes, r.stats.AsyncReadbackBytes, bytes); err != nil {
		return nil, err
	}
	if bytes > r.limits.MaxBufferSize {
		return nil, errors.New("GPU view: readback exceeds device buffer limit")
	}
	tex := r.target.CurrentFrame()
	if err := tex.ConfigReadBuffer(); err != nil {
		return nil, err
	}
	r.stats.ReadbackBytes = bytes
	cmd, err := r.device.Device.TryCreateCommandEncoder(nil)
	if err != nil {
		return nil, err
	}
	defer cmd.Release()
	if err := tex.CopyToReadBuffer(cmd); err != nil {
		return nil, err
	}
	buf, err := cmd.TryFinish(nil)
	if err != nil {
		return nil, err
	}
	defer buf.Release()
	r.device.Queue.Submit(buf)
	img, err := tex.ReadGoImage()
	if err != nil {
		return nil, err
	}
	straightSnapshot(img)
	r.stats.SnapshotReadbacks++
	return img, nil
}

type gpuViewWidget struct {
	core.WidgetBase
	owner         *GPUView
	application   *app.Application
	windowID      string
	windowRef     app.WindowRef
	program       *GPUProgram
	depth         bool
	retired       atomic.Bool
	native        atomic.Pointer[drawingWindow]
	resources     gpuViewResources
	onInput       func(GPUEvent) bool
	inputCapture  uint64
	inputSequence uint64
	lastInput     GPUEvent
}

func (w *gpuViewWidget) Init()  { w.WidgetBase.Init() }
func (w *gpuViewWidget) OnAdd() { w.WidgetBase.OnAdd(); w.Scene.AddDirectRender(w) }
func (w *gpuViewWidget) Destroy() {
	w.endGPUCapture(false)
	if w.retired.Swap(true) {
		return
	}
	defer w.WidgetBase.Destroy()
	defer func() {
		if w.program != nil {
			w.program.mounted.Store(false)
		}
	}()
	if w.Scene != nil {
		if p := w.native.Load(); p != nil && p.window.Lock() {
			defer p.window.Unlock()
		}
		w.Scene.DeleteDirectRender(w)
	}
	w.resources.mu.Lock()
	defer w.resources.mu.Unlock()
	w.resources.reset()
}
func (w *gpuViewWidget) ownerEvent(fn func()) {
	_ = w.application.Do(func() {
		g := w.owner
		if w.retired.Load() || g == nil || g.widget != w {
			return
		}
		ctx, err := scopedWindowContext(w.application, w.windowRef)
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
func (w *gpuViewWidget) Render() {
	if w.owner == nil || w.retired.Load() {
		return
	}
	g := w.owner
	_ = w.application.Do(func() {
		var renderErr error
		defer func() {
			if renderErr != nil {
				g.report(renderErr)
			}
		}()
		if w.retired.Load() || g.widget != w {
			return
		}
		ctx, err := scopedWindowContext(w.application, w.windowRef)
		if err != nil || ctx.Err() != nil {
			return
		}
		r := &w.resources
		r.mu.Lock()
		r.stats.RenderVisits++
		r.mu.Unlock()
		if err := g.validate(); err != nil {
			renderErr = err
			return
		}
		full := math32.RectFromPosSizeMax(w.Geom.Pos.Content, w.Geom.Size.Actual.Content)
		if full.Empty() {
			return
		}
		p := w.native.Load()
		if p == nil {
			rw := w.Scene.RenderWindow()
			if rw == nil || rw.SystemWindow == nil {
				return
			}
			p = &drawingWindow{rw.SystemWindow}
			w.native.Store(p)
		}
		window := p.window
		if !window.Lock() {
			return
		}
		defer window.Unlock()
		if err := drawingDeviceError(window); err != nil {
			renderErr = err
			return
		}
		_, clip := drawingClip(full, w.Geom.ContentBBox, image.Point{}, full.Size())
		key := gpuFrameKey{size: full.Size(), clip: clip, scale: w.Styles.UnitContext.DPI / units.DpPerInch, revision: g.Revision, background: g.Background}
		r.mu.Lock()
		defer r.mu.Unlock()
		renderErr = r.render(nativeSurface(window), key, w.depth, g.budget())
	})
}
func (w *gpuViewWidget) RenderSource(op draw.Op) composer.Source {
	if !w.IsVisible() || w.retired.Load() {
		return nil
	}
	r := &w.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.ready || r.target == nil {
		return nil
	}
	full := math32.RectFromPosSizeMax(w.Geom.Pos.Content, w.Geom.Size.Actual.Content)
	dest, source := drawingClip(full, w.Geom.ContentBBox, w.Scene.SceneGeom.Pos, r.target.Size())
	if source.Empty() {
		return nil
	}
	return &gpuViewSource{r: r, texture: r.target.CurrentFrame(), generation: r.stats.TargetGeneration, dest: dest, source: source}
}

type gpuViewSource struct {
	r            *gpuViewResources
	texture      *gpu.Texture
	generation   uint64
	dest, source image.Rectangle
}

func (s *gpuViewSource) Draw(c composer.Composer) {
	cd, ok := c.(*composer.ComposerDrawer)
	if !ok {
		return
	}
	drawer, ok := cd.Drawer.(*gpudraw.Drawer)
	if !ok {
		return
	}
	r := s.r
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.ready || r.target == nil || r.stats.TargetGeneration != s.generation || r.target.CurrentFrame() != s.texture {
		return
	}
	drawer.UseTexture(s.texture)
	drawer.CopyUsed(s.dest.Min, s.source, draw.Over, false)
}

type gpuFrameKey struct {
	size       image.Point
	clip       image.Rectangle
	scale      float32
	revision   uint64
	background color.NRGBA
}
type gpuViewResources struct {
	mu           sync.Mutex
	prepare      func(GPUFrame) error
	draw         func(GPUFrame, *wgpu.RenderPassEncoder) error
	cleanup      func()
	used         bool
	device       *gpu.Device
	limits       wgpu.Limits
	target       *gpu.RenderTexture
	depth        *wgpu.Texture
	depthView    *wgpu.TextureView
	pool         gpuResourcePool
	leaseCounter uint64
	ready        bool
	key          gpuFrameKey
	stats        GPUViewStats
	capture      *gpuSnapshotJob
}

func (r *gpuViewResources) releaseTarget() {
	if r.depthView != nil {
		r.depthView.Release()
		r.depthView = nil
	}
	if r.depth != nil {
		r.depth.Release()
		r.depth = nil
	}
	if r.target != nil {
		r.target.Release()
		r.target = nil
	}
	r.ready = false
	r.stats.TargetBytes = 0
	r.stats.ReadbackBytes = 0
	r.stats.Size = image.Point{}
}
func (r *gpuViewResources) reset() {
	r.cancelSnapshot()
	r.ready = false
	r.pool.lease.Store(0)
	r.stats.ResourceGeneration++
	r.stats.TargetGeneration++
	defer func() { r.pool.release(); r.releaseTarget(); r.device = nil }()
	used := r.used
	r.used = false
	if used && r.cleanup != nil {
		r.cleanup()
	}
}
func gpuTargetBytes(size image.Point, depth bool) (uint64, error) {
	if size.X < 1 || size.Y < 1 || size.X > 16384 || size.Y > 16384 {
		return 0, errors.New("GPU view: invalid target size")
	}
	bytes := uint64(20)
	if depth {
		bytes = 36
	}
	return uint64(size.X) * uint64(size.Y) * bytes, nil
}
func gpuBudget(budget uint64, parts ...uint64) error {
	remaining := budget
	for _, part := range parts {
		if part > remaining {
			return errors.New("GPU view: resource budget exceeded")
		}
		remaining -= part
	}
	return nil
}
func (r *gpuViewResources) configure(surface *gpu.Surface, size image.Point, depth bool) error {
	r.releaseTarget()
	r.stats.TargetGeneration++
	r.target = gpu.NewRenderTexture(surface.GPU, surface.Device(), image.Pt(1, 1), 1, gpu.UndefinedType)
	r.target.Format.SetMultisample(4)
	r.target.Render().Format.SetMultisample(4)
	if err := resizeDrawingTarget(r.target, size); err != nil {
		return err
	}
	if depth {
		var err error
		r.depth, err = r.device.Device.TryCreateTexture(&wgpu.TextureDescriptor{Label: "vellum.gpu.depth", Dimension: wgpu.TextureDimension2D, Size: wgpu.Extent3D{Width: uint32(size.X), Height: uint32(size.Y), DepthOrArrayLayers: 1}, MipLevelCount: 1, SampleCount: 4, Format: wgpu.TextureFormatDepth32Float, Usage: wgpu.TextureUsageRenderAttachment})
		if err != nil {
			return err
		}
		r.depthView, err = r.depth.TryCreateView(nil)
		if err != nil {
			return err
		}
	}
	r.stats.TargetBytes, _ = gpuTargetBytes(size, depth)
	r.stats.Size = size
	return nil
}
func (r *gpuViewResources) render(surface *gpu.Surface, key gpuFrameKey, depth bool, budget uint64) (err error) {
	if surface == nil || surface.Device() == nil {
		return ErrFrameNotReady
	}
	if r.device != nil && r.device.Device != surface.Device().Device {
		r.reset()
	}
	if r.device == nil {
		r.device = surface.Device()
		r.limits = r.device.Device.GetLimits()
	}
	targetBytes, e := gpuTargetBytes(key.size, depth)
	if e != nil {
		return e
	}
	if uint64(key.size.X) > uint64(r.limits.MaxTextureDimension2D) || uint64(key.size.Y) > uint64(r.limits.MaxTextureDimension2D) {
		return errors.New("GPU view: target exceeds logical device limit")
	}
	readback := r.stats.ReadbackBytes
	if r.target == nil || r.target.Size() != key.size {
		readback = 0
	}
	if e := gpuBudget(budget, targetBytes, r.stats.BufferBytes, r.stats.TextureBytes, readback, r.stats.AsyncReadbackBytes); e != nil {
		return e
	}
	if r.ready && r.key == key {
		return nil
	}
	// From here any failure (including a callback panic) withdraws the frame and
	// cleans partially initialized program/helper resources before another attempt.
	complete := false
	defer func() {
		if !complete {
			r.reset()
		}
	}()
	r.ready = false
	if r.target == nil || r.target.Size() != key.size {
		if e := r.configure(surface, key.size, depth); e != nil {
			return e
		}
	}
	r.pool.device = r.device
	r.pool.limits = r.limits
	r.pool.stats = &r.stats
	r.pool.budget = budget
	encoder, e := r.device.Device.TryCreateCommandEncoder(nil)
	if e != nil {
		return e
	}
	defer encoder.Release()
	frame := GPUFrame{Version: GPUContextVersion, Size: key.size, Clip: key.clip, DpScale: key.scale, Format: wgpu.TextureFormatRGBA8UnormSrgb, Samples: 4, ResourceGeneration: r.stats.ResourceGeneration, TargetGeneration: r.stats.TargetGeneration, Device: r.device.Device, Queue: r.device.Queue, Limits: r.limits, Encoder: encoder}
	if depth {
		frame.DepthFormat = wgpu.TextureFormatDepth32Float
	}
	r.used = true
	if r.prepare != nil {
		r.stats.Prepares++
		if e := r.callbackFrame(frame, r.prepare); e != nil {
			return fmt.Errorf("GPU view prepare: %w", e)
		}
	}
	r.target.Render().ClearColor = key.background
	view, e := r.target.GetCurrentTexture()
	if e != nil {
		return e
	}
	desc := r.target.Render().ClearRenderPass(view)
	if depth {
		desc.DepthStencilAttachment = &wgpu.RenderPassDepthStencilAttachment{View: r.depthView, DepthLoadOp: wgpu.LoadOpClear, DepthStoreOp: wgpu.StoreOpStore, DepthClearValue: 1, StencilReadOnly: true}
	}
	pass, e := encoder.TryBeginRenderPass(desc)
	if e != nil {
		return e
	}
	e = func() error {
		defer pass.Release()
		defer pass.End()
		r.stats.Draws++
		return r.callbackFrame(frame, func(drawFrame GPUFrame) error { drawFrame.Encoder = nil; return r.draw(drawFrame, pass) })
	}()
	if e != nil {
		return fmt.Errorf("GPU view draw: %w", e)
	}
	buffer, e := encoder.TryFinish(nil)
	if e != nil {
		return e
	}
	defer buffer.Release()
	r.device.Queue.Submit(buffer)
	if e := r.device.LossError(); e != nil {
		return fmt.Errorf("%w: %v", ErrDeviceLost, e)
	}
	r.key = key
	r.ready = true
	r.stats.Frames++
	complete = true
	return nil
}

// Each callback gets a fresh lease, even when Prepare and Draw share one frame.
func (r *gpuViewResources) callbackFrame(frame GPUFrame, callback func(GPUFrame) error) error {
	r.leaseCounter++
	if r.leaseCounter == 0 {
		r.leaseCounter++
	}
	token := r.leaseCounter
	r.pool.lease.Store(token)
	defer r.pool.lease.Store(0)
	frame.Resources = &GPUResources{pool: &r.pool, token: token}
	return callback(frame)
}
