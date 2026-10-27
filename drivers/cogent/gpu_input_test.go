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
	"testing"
)

func TestGPUInputSnapshotCoordinatesAndModifiers(t *testing.T) {
	w := &gpuViewWidget{}
	w.Geom.Pos.Content = math32.Vec2(100, 50)
	w.Geom.Size.Actual.Content = math32.Vec2(400, 300)
	w.Geom.ContentBBox = image.Rect(120, 50, 500, 350)
	var mods key.Modifiers
	mods.SetFlag(true, key.Shift, key.Alt, key.CapsLock)
	e := events.NewMouse(events.MouseDown, events.Left, image.Pt(150, 90), mods)
	e.SetLocalOff(image.Point{})
	v := w.eventData(GPUPointerDown, e)
	if v.Position != image.Pt(50, 40) || v.Size != image.Pt(400, 300) || !v.Inside || v.Button != GPUButtonPrimary || v.Modifiers != GPUShift|GPUOption {
		t.Fatalf("snapshot: %+v", v)
	}
	e.Where = image.Pt(105, 90)
	e.SetLocalOff(image.Point{})
	if w.eventData(GPUPointerDown, e).Inside {
		t.Fatal("clipped position admitted")
	}
	if v.Position != image.Pt(50, 40) {
		t.Fatal("snapshot retained event pointer")
	}
}
func TestGPUInputKindsAndModifierLockBits(t *testing.T) {
	var m key.Modifiers
	m.SetFlag(true, key.Control, key.Meta, key.NumLock)
	if gpuModifiers(m) != GPUControl|GPUCommand {
		t.Fatal("modifier conversion")
	}
	w := &gpuViewWidget{}
	w.endGPUCapture(true)
	if w.inputCapture != 0 {
		t.Fatal("empty capture")
	}
	if GPUButton(events.Middle) != GPUButtonMiddle || GPUButton(events.Right) != GPUButtonSecondary {
		t.Fatal("button mapping")
	}
}

func TestGPUInputCancellationQueuedWithoutOwnerReentry(t *testing.T) {
	a, _ := app.New(app.Config{QueueCapacity: 8})
	defer a.Close()
	_ = a.Open(app.Window{ID: "typed", Title: "Typed", Width: 200, Height: 150, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	var received []GPUEvent
	g := &GPUView{Owner: a, Window: "typed", Program: &GPUProgram{Draw: func(GPUFrame, *wgpu.RenderPassEncoder) error { return nil }}, InputEvent: func(e GPUEvent) bool { received = append(received, e); return true }, Input: func(GPUInput) bool { t.Fatal("legacy input called in typed mode"); return false }}
	if g.Mount(core.NewBody("typed")) == nil {
		t.Fatal("mount")
	}
	w := g.widget
	w.inputCapture = 7
	w.lastInput = GPUEvent{Kind: GPUPointerDown, Sequence: 19}
	if err := a.Do(func() { w.endGPUCapture(true) }); err != nil {
		t.Fatal(err)
	}
	if len(received) != 0 || w.inputCapture != 0 {
		t.Fatal("cancel nested owner callback")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if len(received) != 1 || received[0].Kind != GPUCancel || received[0].Sequence != 19 {
		t.Fatal(received)
	}
	w.inputCapture = 8
	w.endGPUCapture(true)
	w.Destroy()
	_ = a.Drain()
	if len(received) != 1 {
		t.Fatal("callback survived retirement")
	}
}

func TestGPUInputTargetedGuard(t *testing.T) {
	if gpuTargetedInputAllowed(events.MouseUp, events.Left, 0) || gpuTargetedInputAllowed(events.MouseDrag, events.Left, 0) || gpuTargetedInputAllowed(events.Scroll, events.NoButton, 7) {
		t.Fatal("unowned/duplicate input admitted")
	}
	if !gpuTargetedInputAllowed(events.MouseMove, events.NoButton, 0) || !gpuTargetedInputAllowed(events.MouseUp, events.Right, 0) {
		t.Fatal("targeted input refused")
	}
}
