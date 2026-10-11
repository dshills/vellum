package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/oliverbestmann/webgpu/wgpu"
	"image"
	"math"
	"testing"
)

func TestGPUTrackedAdmissionOverflowAndWrites(t *testing.T) {
	for _, depth := range []bool{false, true} {
		bytes, err := gpuTargetBytes(image.Pt(100, 80), depth)
		want := uint64(160000)
		if depth {
			want = 288000
		}
		if err != nil || bytes != want {
			t.Fatal(bytes, err)
		}
	}
	if _, err := gpuTargetBytes(image.Pt(math.MaxInt, 2), true); err == nil {
		t.Fatal("unbounded target")
	}
	if gpuBudget(math.MaxUint64, math.MaxUint64, 1) == nil {
		t.Fatal("wrapped budget")
	}
	limits := wgpu.Limits{MaxTextureDimension2D: 16384}
	desc := GPUTexture{Width: 10, Height: 20, Format: wgpu.TextureFormatRGBA16Float, Usage: wgpu.TextureUsageRenderAttachment, Samples: 4}
	if bytes, err := gpuTextureBytes(desc, limits); err != nil || bytes != 6400 {
		t.Fatal(bytes, err)
	}
	desc.Usage |= wgpu.TextureUsageCopySrc
	if _, err := gpuTextureBytes(desc, limits); err == nil {
		t.Fatal("multisample copy")
	}
	desc.Width = math.MaxUint32
	desc.Height = math.MaxUint32
	desc.Usage = wgpu.TextureUsageRenderAttachment
	if _, err := gpuTextureBytes(desc, wgpu.Limits{MaxTextureDimension2D: math.MaxUint32}); err == nil {
		t.Fatal("texture byte overflow")
	}
	desc.Width = 1
	desc.Height = 1
	desc.Format = wgpu.TextureFormatUndefined
	if _, err := gpuTextureBytes(desc, limits); err == nil {
		t.Fatal("unsupported format")
	}
	b := wgpu.BufferDescriptor{Size: 16, Usage: wgpu.BufferUsageCopyDst}
	if gpuWriteBounds(b, 12, 4) != nil || gpuWriteBounds(b, 16, 0) != nil {
		t.Fatal("valid boundary refused")
	}
	for _, probe := range []struct {
		offset uint64
		size   int
	}{{12, 8}, {math.MaxUint64 - 3, 4}, {2, 4}, {0, 3}, {0, -4}} {
		if gpuWriteBounds(b, probe.offset, probe.size) == nil {
			t.Fatal("invalid upload accepted", probe)
		}
	}
	b.Usage = wgpu.BufferUsageVertex
	if gpuWriteBounds(b, 0, 4) == nil {
		t.Fatal("missing CopyDst")
	}
}
func TestGPUFrameHelperLeaseAndPanicCleanup(t *testing.T) {
	var r gpuViewResources
	r.pool.stats = &r.stats
	r.stats.BufferBytes = 24
	r.stats.TextureBytes = 32
	r.stats.TargetBytes = 99
	r.stats.ReadbackBytes = 8
	r.ready = true
	r.used = true
	r.pool.lease.Store(1)
	old := &GPUResources{pool: &r.pool, token: 1}
	if _, err := old.active(); err != nil {
		t.Fatal(err)
	}
	r.pool.lease.Store(2)
	if old.Drop("any") == nil {
		t.Fatal("old frame helper escaped lease")
	}
	releases := 0
	r.cleanup = func() { releases++; panic("program release") }
	func() {
		defer func() {
			if recover() == nil {
				t.Error("cleanup panic swallowed")
			}
		}()
		r.reset()
	}()
	if r.ready || r.used || r.stats.BufferBytes != 0 || r.stats.TextureBytes != 0 || r.stats.TargetBytes != 0 || r.stats.ReadbackBytes != 0 || r.pool.lease.Load() != 0 || releases != 1 {
		t.Fatalf("panic stranded resources: %+v releases=%d", r.stats, releases)
	}
	r.reset()
	if releases != 1 {
		t.Fatal("partial cleanup released twice")
	}
}
func TestGPUProgramMountCaptureRetirementAndIdentity(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(app.Window{ID: "gpu-test", Title: "test", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	calls := 0
	p := &GPUProgram{Draw: func(GPUFrame, *wgpu.RenderPassEncoder) error { calls++; return nil }}
	g := &GPUView{Owner: a, Window: "gpu-test", Program: p, Depth: true}
	host := core.NewBody("gpu-test")
	if g.Mount(host) == nil {
		t.Fatal("mount")
	}
	other := &GPUView{Owner: a, Window: "gpu-test", Program: p, Depth: true}
	if other.Mount(host) != nil {
		t.Fatal("program shared between mounts")
	}
	p.Draw = func(GPUFrame, *wgpu.RenderPassEncoder) error { t.Fatal("mutable callback reread"); return nil }
	g.widget.resources.draw(GPUFrame{}, nil)
	if calls != 1 {
		t.Fatal("callback not captured")
	}
	g.Depth = false
	if g.validate() == nil {
		t.Fatal("mounted depth changed")
	}
	g.Depth = true
	g.Program = &GPUProgram{Draw: p.Draw}
	if g.validate() == nil {
		t.Fatal("mounted program changed")
	}
	g.Program = p
	old := g.widget
	old.Destroy()
	if p.mounted.Load() {
		t.Fatal("retired mount retained program claim")
	}
	if g.NativeWindow() != nil {
		t.Fatal("retired window")
	}
	if g.Mount(host) == nil {
		t.Fatal("remount")
	}
	old.Render()
	old.Destroy()
	if !p.mounted.Load() {
		t.Fatal("stale Destroy released new mount claim")
	}
	if other.Mount(host) != nil {
		t.Fatal("stale Destroy permitted sharing")
	}
	g.widget.Destroy()
}
func TestGPUClosedAdapterAndInputCallbacks(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(app.Window{ID: "gpu-closed", Title: "test", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	called := false
	p := &GPUProgram{Draw: func(GPUFrame, *wgpu.RenderPassEncoder) error { return nil }}
	g := &GPUView{Owner: a, Window: "gpu-closed", Program: p}
	g.Mount(core.NewBody("gpu-closed"))
	a.CloseWindow(g.Window)
	g.Input = func(GPUInput) bool { called = true; return true }
	click := events.NewMouse(events.MouseDown, events.Left, image.Pt(20, 30), 0)
	click.SetLocalOff(image.Point{})
	g.widget.HandleEvent(click)
	g.widget.ownerEvent(func() { called = true })
	g.widget.Render()
	if called {
		t.Fatal("closed owner event")
	}
	if _, e := g.Stats(); e != ErrFrameNotReady {
		t.Fatal(e)
	}
	if _, e := g.Snapshot(); e != ErrFrameNotReady {
		t.Fatal(e)
	}
	g.widget.Destroy()
}

func TestGPUCallbackLeaseExpiresBetweenPrepareAndDraw(t *testing.T) {
	var r gpuViewResources
	var saved *GPUResources
	if err := r.callbackFrame(GPUFrame{}, func(f GPUFrame) error { saved = f.Resources; _, e := saved.active(); return e }); err != nil {
		t.Fatal(err)
	}
	if _, e := saved.active(); e == nil {
		t.Fatal("Prepare helper retained after return")
	}
	if err := r.callbackFrame(GPUFrame{}, func(f GPUFrame) error {
		if _, e := saved.active(); e == nil {
			t.Fatal("Prepare helper accepted in Draw")
		}
		_, e := f.Resources.active()
		return e
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGPUInputPropagatesUnconsumedKeysAndScroll(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(app.Window{ID: "gpu-input", Title: "test", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	g := &GPUView{Owner: a, Window: "gpu-input", Program: &GPUProgram{Draw: func(GPUFrame, *wgpu.RenderPassEncoder) error { return nil }}}
	g.Mount(core.NewBody("gpu-input"))
	defer g.widget.Destroy()
	emit := func(consumed bool) {
		g.Input = nil
		if consumed {
			g.Input = func(in GPUInput) bool { return in.Key == "Tab" || in.Scroll != 0 }
		}
		chord := events.NewKey(events.KeyChord, 0, key.CodeTab, 0)
		chord.SetLocalOff(image.Point{})
		g.widget.HandleEvent(chord)
		scroll := events.NewScroll(image.Pt(20, 30), math32.Vec2(0, 50), 0)
		scroll.SetLocalOff(image.Point{})
		g.widget.HandleEvent(scroll)
		if chord.IsHandled() != consumed || scroll.IsHandled() != consumed {
			t.Fatalf("consumed=%v chord=%v scroll=%v", consumed, chord.IsHandled(), scroll.IsHandled())
		}
	}
	emit(false)
	emit(true)
	g.Input = func(GPUInput) bool { return false }
	escape := events.NewKey(events.KeyChord, 0, key.CodeEscape, 0)
	escape.SetLocalOff(image.Point{})
	g.widget.HandleEvent(escape)
	if escape.IsHandled() {
		t.Fatal("unconsumed Escape swallowed")
	}
}
