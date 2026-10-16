package cogent

import (
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/gpu/gpudraw"
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/composer"
	"errors"
	"fmt"
	"time"
)

// ErrDeviceLost requires a fresh process/device. Domain models are not modified.
var ErrDeviceLost = errors.New("cogent: GPU device lost")

type demandPainter interface{ SetOnDemand(bool) }

func configureRendering(window any, continuous bool) error {
	painter, ok := window.(demandPainter)
	if !ok {
		return fmt.Errorf("cogent: native window lacks requested-paint support")
	}
	painter.SetOnDemand(!continuous)
	return nil
}
func nativeSurface(window system.Window) *gpu.Surface {
	if window == nil {
		return nil
	}
	cd, ok := window.Composer().(*composer.ComposerDrawer)
	if !ok {
		return nil
	}
	drawer, ok := cd.Drawer.(*gpudraw.Drawer)
	if !ok {
		return nil
	}
	surface, ok := drawer.Renderer().(*gpu.Surface)
	if !ok {
		return nil
	}
	return surface
}
func nativeDevice(window system.Window) *gpu.Device {
	surface := nativeSurface(window)
	if surface == nil {
		return nil
	}
	return surface.Device()
}

func watchDeviceLoss(w *nativeWindow, fail func(error)) {
	if w == nil || w.body == nil || w.body.Scene == nil {
		return
	}
	rw := w.body.Scene.RenderWindow()
	if rw == nil || rw.SystemWindow == nil {
		return
	}
	device := nativeDevice(rw.SystemWindow)
	if device == nil {
		return
	}
	installDeviceLoss(w, func() func() {
		return device.ObserveLoss(func(err error) {
			if !w.closing.Load() {
				fail(fmt.Errorf("%w: window %s: %v", ErrDeviceLost, w.window.ID, err))
			}
		})
	})
}

// The loss mutex is independent of rootMu: startup and teardown can overlap
// while cleanup already owns the widget root. Closing suppresses late installs.
func installDeviceLoss(w *nativeWindow, subscribe func() func()) {
	w.lossMu.Lock()
	unavailable := w.closing.Load() || w.stopLoss != nil
	w.lossMu.Unlock()
	if unavailable {
		return
	}
	// Subscribe may replay an already-retained loss synchronously; no lifecycle
	// lock is held while that callback triggers shutdown.
	stop := subscribe()
	w.lossMu.Lock()
	if w.closing.Load() || w.stopLoss != nil {
		w.lossMu.Unlock()
		if stop != nil {
			stop()
		}
		return
	}
	w.stopLoss = stop
	w.lossMu.Unlock()
}
func stopDeviceLoss(w *nativeWindow) {
	w.lossMu.Lock()
	stop := w.stopLoss
	w.stopLoss = nil
	w.lossMu.Unlock()
	if stop != nil {
		stop()
	}
}

// RequestFrameAfter schedules one native paint, for application-owned animations.
// Call from the window's owner/native event context. Application.Invalidate is
// sufficient for model changes; this method never starts a continuous ticker.
func (g *GPUCanvas) RequestFrameAfter(delay time.Duration) error {
	if delay < 0 {
		return fmt.Errorf("canvas: negative frame delay")
	}
	if g == nil || g.widget == nil || g.widget.retired.Load() || g.widget.WidgetBase.Scene == nil {
		return ErrFrameNotReady
	}
	window := g.NativeWindow()
	if window == nil {
		return ErrFrameNotReady
	}
	painter, ok := window.(interface{ RequestPaintAfter(time.Duration) })
	if !ok {
		return fmt.Errorf("canvas: requested paint unavailable")
	}
	painter.RequestPaintAfter(delay)
	return nil
}

// RebuildResources recreates canvas meshes/material bindings on the live device.
// It does not attempt to reuse a lost physical device. Call on the window owner.
func (g *GPUCanvas) RebuildResources() error {
	if g == nil || g.widget == nil || g.widget.retired.Load() {
		return ErrFrameNotReady
	}
	if g.Model == nil {
		return fmt.Errorf("canvas: model required")
	}
	if err := g.Model.Validate(); err != nil {
		return err
	}
	window := g.NativeWindow()
	if window == nil || !window.Lock() {
		return ErrFrameNotReady
	}
	defer window.Unlock()
	if g.widget == nil || g.widget.retired.Load() || g.widget.XYZ.Frame == nil {
		return ErrFrameNotReady
	}
	device := g.widget.XYZ.Frame.Device()
	if err := device.LossError(); err != nil {
		return fmt.Errorf("%w: %v", ErrDeviceLost, err)
	}
	g.widget.XYZ.Rebuild()
	g.updateCamera()
	return nil
}

// NativeWindow exposes the originating native window for custom native integrations.
// Call on the UI owner; nil means it has not been shown yet.
func (g *GPUCanvas) NativeWindow() system.Window {
	if g == nil || g.widget == nil || g.widget.retired.Load() || g.widget.WidgetBase.Scene == nil {
		return nil
	}
	if g.Owner == nil || g.Owner != g.widget.application || g.Window != g.widget.windowRef.ID() {
		return nil
	}
	ctx, err := scopedWindowContext(g.Owner, g.widget.windowRef)
	if err != nil || ctx.Err() != nil {
		return nil
	}
	rw := g.widget.WidgetBase.Scene.RenderWindow()
	if rw == nil {
		return nil
	}
	return rw.SystemWindow
}
