package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"sync/atomic"
	"time"

	"cogentcore.org/core/system"
	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/go-gl/glfw/v3.4/glfw"
)

func owner(a *app.Application, fn func() error) error {
	var result error
	if err := a.Do(func() { result = fn() }); err != nil {
		return err
	}
	return result
}
func waitStats(ctx context.Context, a *app.Application, w *imageWindow, predicate func(cogent.GPUViewStats) bool) (cogent.GPUViewStats, error) {
	var stats cogent.GPUViewStats
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-w.nativeErrors:
			return stats, fmt.Errorf("native GPU program: %w", err)
		default:
		}
		if err := a.Drain(); err != nil {
			return stats, err
		}
		err := owner(a, func() error { var e error; stats, e = w.view.Stats(); return e })
		if err == nil && w.shown.Load() && predicate(stats) {
			return stats, nil
		}
		select {
		case <-ctx.Done():
			return stats, fmt.Errorf("GPU wait %+v: %w", stats, ctx.Err())
		case <-ticker.C:
		}
	}
}
func waitFrame(ctx context.Context, a *app.Application, w *imageWindow, after uint64) (cogent.GPUViewStats, error) {
	return waitStats(ctx, a, w, func(s cogent.GPUViewStats) bool { return s.Ready && s.Frames > after })
}

type asyncResult struct {
	snapshot cogent.GPUSnapshot
	current  cogent.GPUViewStats
	err      error
}

func startCapture(ctx context.Context, a *app.Application, w *imageWindow) (<-chan asyncResult, error) {
	result := make(chan asyncResult, 1)
	err := owner(a, func() error {
		return w.view.SnapshotAsync(ctx, func(snapshot cogent.GPUSnapshot, err error) {
			// Check callback-recursive admission first, without any intervening Stats.
			busy := w.view.SnapshotAsync(context.Background(), func(cogent.GPUSnapshot, error) {})
			if !errors.Is(busy, cogent.ErrSnapshotBusy) {
				err = errors.Join(err, fmt.Errorf("callback recursion admitted: %v", busy))
			}
			stats, statsErr := w.view.Stats()
			result <- asyncResult{snapshot: snapshot, current: stats, err: errors.Join(err, statsErr)}
		})
	})
	return result, err
}
func awaitCapture(ctx context.Context, a *app.Application, result <-chan asyncResult) (asyncResult, error) {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := a.Drain(); err != nil {
			return asyncResult{}, err
		}
		select {
		case got := <-result:
			return got, got.err
		case <-ctx.Done():
			return asyncResult{}, ctx.Err()
		case <-ticker.C:
		}
	}
}
func capture(ctx context.Context, a *app.Application, w *imageWindow, exposure float64) (cogent.GPUSnapshot, error) {
	result, err := startCapture(ctx, a, w)
	if err != nil {
		return cogent.GPUSnapshot{}, err
	}
	got, err := awaitCapture(ctx, a, result)
	if err != nil {
		return got.snapshot, err
	}
	if err := checkPixels(got.snapshot.Image, exposure); err != nil {
		return got.snapshot, err
	}
	return got.snapshot, nil
}
func checkPixels(img *image.NRGBA, exposure float64) error {
	if img == nil || img.Bounds().Dx() < 8 || img.Bounds().Dy() < 8 {
		return errors.New("missing imaging readback")
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	probes := []struct {
		x, y               int
		right, transparent bool
	}{{width / 4, height / 8, false, true}, {width / 4, 3 * height / 4, false, false}, {3 * width / 4, 3 * height / 4, true, false}}
	for _, p := range probes {
		got := img.NRGBAAt(p.x, p.y)
		want := referencePixel(exposure, p.right, p.transparent)
		if (p.transparent && got != want) || !nearPixel(got, want, 4) {
			return fmt.Errorf("HDR exposure=%g at (%d,%d): got %v want %v", exposure, p.x, p.y, got, want)
		}
	}
	return nil
}
func nearPixel(got, want color.NRGBA, tolerance int) bool {
	for _, v := range [][2]uint8{{got.R, want.R}, {got.G, want.G}, {got.B, want.B}, {got.A, want.A}} {
		difference := int(v[0]) - int(v[1])
		if difference < -tolerance || difference > tolerance {
			return false
		}
	}
	return true
}
func press(ctx context.Context, window, id string) error {
	for {
		ok, err := cogent.PerformAccessibilityAction(ctx, window, id, "press", "")
		if err == nil && ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("native button %s: %w", id, ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
}
func retained(before, after cogent.GPUViewStats) error {
	if after.ResourceGeneration != before.ResourceGeneration || after.BufferAllocations != before.BufferAllocations || after.TextureAllocations != before.TextureAllocations || after.BufferBytes != storageBytes+uniformBytes || after.TextureBytes != floatTextureBytes {
		return fmt.Errorf("retained HDR resources changed: before=%+v after=%+v", before, after)
	}
	if after.HelperUploadBytes-before.HelperUploadBytes != (after.Frames-before.Frames)*uniformBytes {
		return fmt.Errorf("changed frames uploaded more than 32 bytes: before=%+v after=%+v", before, after)
	}
	return nil
}

// Select a different bounded request even when the native adapter restores the
// size requested by an earlier fixture run.
func resizeRequest(current image.Point) image.Point {
	first, second := image.Pt(900, 700), image.Pt(1000, 760)
	if current == first {
		return second
	}
	return first
}
func nativeWindowSize(window system.Window) (image.Point, error) {
	var size image.Point
	var err error
	system.TheApp.RunOnMain(func() {
		native, ok := window.(*desktop.Window)
		if !ok || native.IsClosed() || native.Glw == nil {
			err = errors.New("native window size unavailable")
			return
		}
		x, y := native.Glw.GetSize()
		size = image.Pt(x, y)
	})
	return size, err
}
func minimized(window system.Window) bool {
	value := false
	system.TheApp.RunOnMain(func() {
		if native, ok := window.(*desktop.Window); ok && !native.IsClosed() && native.Glw != nil {
			value = native.Glw.GetAttrib(glfw.Iconified) == glfw.True
		}
	})
	return value
}
func restore(window system.Window) {
	system.TheApp.RunOnMain(func() {
		if native, ok := window.(*desktop.Window); ok && !native.IsClosed() && native.Glw != nil {
			native.Glw.Restore()
		}
	})
}

// Last-window close intentionally stops the application owner and its mailbox.
// Observe native/program teardown on a separate bounded context, without Do or
// Drain and without inferring the adapter's private snapshot reservation state.
func waitProgramTeardown(ctx context.Context, native system.Window, model *imagingProgram, after uint64) error {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		storage, texture, releases := model.resourceIdentity()
		if native.IsClosed() && storage == nil && texture == nil && releases > after {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("native/program teardown incomplete: nativeClosed=%v releases=%d after=%d: %w", native.IsClosed(), releases, after, ctx.Err())
		case <-ticker.C:
		}
	}
}
func waitCondition(ctx context.Context, a *app.Application, condition func() bool) error {
	for !condition() {
		if err := a.Drain(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
	return nil
}

func admissionProbes(ctx context.Context, a *app.Application, w *imageWindow) error {
	return owner(a, func() error {
		before, err := w.view.Stats()
		if err != nil {
			return err
		}
		if before.SnapshotReadbacks != 0 || before.AsyncSnapshotRequests != 0 {
			return errors.New("ordinary rendering read back pixels")
		}
		callback := func(cogent.GPUSnapshot, error) {}
		if w.view.SnapshotAsync(nil, callback) == nil || w.view.SnapshotAsync(ctx, nil) == nil {
			return errors.New("nil snapshot argument accepted")
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if !errors.Is(w.view.SnapshotAsync(canceled, callback), context.Canceled) {
			return errors.New("precanceled snapshot accepted")
		}
		w.view.MaxSnapshotBytes = 1
		limited := w.view.SnapshotAsync(ctx, callback)
		w.view.MaxSnapshotBytes = 0
		if limited == nil {
			return errors.New("snapshot byte limit admitted native staging")
		}
		padded := uint64((before.Size.X*4+255)/256*256) * uint64(before.Size.Y)
		w.view.MaxGPUBytes = before.TargetBytes + before.BufferBytes + before.TextureBytes + before.ReadbackBytes + padded - 1
		refused := w.view.SnapshotAsync(ctx, callback)
		w.view.MaxGPUBytes = 0
		if refused == nil {
			return errors.New("GPU budget admitted native staging over limit")
		}
		after, err := w.view.Stats()
		if err != nil {
			return err
		}
		if !after.Ready || after.Frames != before.Frames || after.AsyncSnapshotRequests != before.AsyncSnapshotRequests || after.AsyncReadbackBytes != 0 || after.SnapshotPending {
			return fmt.Errorf("refusal changed native frame/capture accounting: %+v %+v", before, after)
		}
		img, err := w.view.Snapshot()
		if err != nil {
			return err
		}
		return checkPixels(img, 1) // one explicitly requested sync staging coexists with async
	})
}

func imagingSmoke(ctx context.Context, a *app.Application, first, second *imageWindow, nativeErrors <-chan error) error {
	initial, err := waitFrame(ctx, a, first, 0)
	if err != nil {
		return err
	}
	independent, err := waitFrame(ctx, a, second, 0)
	if err != nil {
		return err
	}
	if initial.BufferBytes != storageBytes+uniformBytes || initial.TextureBytes != floatTextureBytes || initial.HelperUploadBytes != initial.Frames*uniformBytes {
		return fmt.Errorf("HDR byte accounting: %+v", initial)
	}
	firstStorage, firstFloat, _ := first.model.resourceIdentity()
	secondStorage, secondFloat, _ := second.model.resourceIdentity()
	if firstStorage == nil || firstFloat == nil || firstStorage == secondStorage || firstFloat == secondFloat {
		return errors.New("windows share program resources")
	}
	if err := admissionProbes(ctx, a, first); err != nil {
		return err
	}
	firstCapture, err := capture(ctx, a, first, 1)
	if err != nil {
		return err
	}
	if _, err := capture(ctx, a, second, 1); err != nil {
		return err
	}
	beforeChrome, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool { return !s.SnapshotPending })
	if err != nil {
		return err
	}
	if err := owner(a, func() error {
		var err error
		beforeChrome, err = first.view.Stats()
		if err != nil {
			return err
		}
		first.status = "Chrome update accepted"
		return a.Invalidate(first.view.Window)
	}); err != nil {
		return err
	}
	chrome, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool { return s.RenderVisits > beforeChrome.RenderVisits })
	if err != nil {
		return err
	}
	if chrome.Frames != beforeChrome.Frames || chrome.Prepares != beforeChrome.Prepares || chrome.HelperUploads != beforeChrome.HelperUploads {
		return fmt.Errorf("chrome update rerendered HDR: %+v %+v", beforeChrome, chrome)
	}
	if err := press(ctx, first.view.Window, "exposure-up"); err != nil {
		return err
	}
	exposed, err := waitFrame(ctx, a, first, chrome.Frames)
	if err != nil {
		return err
	}
	if err := retained(chrome, exposed); err != nil {
		return err
	}
	if _, err := capture(ctx, a, first, 2); err != nil {
		return err
	}
	if err := press(ctx, first.view.Window, "reset"); err != nil {
		return err
	}
	reset, err := waitFrame(ctx, a, first, exposed.Frames)
	if err != nil {
		return err
	}
	if err := retained(exposed, reset); err != nil {
		return err
	}
	if _, err := capture(ctx, a, first, 1); err != nil {
		return err
	}
	// Exercise the ordinary Capture button and its detached-image status flow.
	if err := press(ctx, first.view.Window, "capture"); err != nil {
		return err
	}
	if err := waitCondition(ctx, a, func() bool { return len(first.captureResults) > 0 }); err != nil {
		return err
	}
	buttonCapture := <-first.captureResults
	if buttonCapture.err != nil {
		return buttonCapture.err
	}
	if err := checkPixels(buttonCapture.snapshot.Image, 1); err != nil {
		return err
	}
	beforeResize, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool { return !s.SnapshotPending })
	if err != nil {
		return err
	}
	var native system.Window
	if err := owner(a, func() error { native = first.view.NativeWindow(); return nil }); err != nil {
		return err
	}
	if native == nil {
		return errors.New("native imaging window unavailable")
	}
	// Capture the completed frame, then edit the model and replace the normal
	// target. Readback owns its private staging allocation across that resize.
	result, err := startCapture(ctx, a, first)
	if err != nil {
		return err
	}
	var capturedRevision uint64
	if err := owner(a, func() error {
		capturedRevision = first.view.Revision
		_ = first.model.setExposure(2)
		first.view.Revision++
		return a.Invalidate(first.view.Window)
	}); err != nil {
		return err
	}
	currentNative, err := nativeWindowSize(native)
	if err != nil {
		return err
	}
	native.SetWinSize(resizeRequest(currentNative)) // outside the owner/main-lock cycle
	resized, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool {
		return s.Ready && s.Size != beforeResize.Size && s.TargetGeneration > beforeResize.TargetGeneration
	})
	if err != nil {
		return err
	}
	old, err := awaitCapture(ctx, a, result)
	if err != nil {
		return err
	}
	if old.snapshot.Frame != beforeResize.Frames || old.snapshot.Revision != capturedRevision || old.snapshot.TargetGeneration != beforeResize.TargetGeneration || old.snapshot.ResourceGeneration != beforeResize.ResourceGeneration || old.snapshot.Image.Bounds().Size() != beforeResize.Size {
		return fmt.Errorf("captured metadata changed across model/resize: %+v before=%+v", old.snapshot, beforeResize)
	}
	if err := checkPixels(old.snapshot.Image, 1); err != nil {
		return err
	}
	if err := retained(beforeResize, resized); err != nil {
		return err
	}
	storage, floatTex, _ := first.model.resourceIdentity()
	if storage != firstStorage || floatTex != firstFloat {
		return errors.New("normal resize replaced retained compute/float resources")
	}
	if _, err := capture(ctx, a, first, 2); err != nil {
		return err
	}
	// Cancellation and rebuild suppress delivery, with an observable released
	// reservation and canceled counter while this window is still alive.
	var dropped atomic.Int32
	canceledBefore, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool { return !s.SnapshotPending })
	if err != nil {
		return err
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	if err := owner(a, func() error {
		if e := first.view.SnapshotAsync(cancelCtx, func(cogent.GPUSnapshot, error) { dropped.Add(1) }); e != nil {
			return e
		}
		cancel()
		return nil
	}); err != nil {
		cancel()
		return err
	}
	canceled, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool {
		return !s.SnapshotPending && s.AsyncSnapshotCanceled > canceledBefore.AsyncSnapshotCanceled
	})
	if err != nil {
		return err
	}
	if dropped.Load() != 0 || canceled.AsyncReadbackBytes != 0 {
		return errors.New("canceled capture delivered or retained staging")
	}
	if err := owner(a, func() error {
		if e := first.view.SnapshotAsync(ctx, func(cogent.GPUSnapshot, error) { dropped.Add(1) }); e != nil {
			return e
		}
		return first.view.RebuildResources()
	}); err != nil {
		return err
	}
	rebuilt, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool {
		return s.Ready && !s.SnapshotPending && s.ResourceGeneration > resized.ResourceGeneration
	})
	if err != nil {
		return err
	}
	if dropped.Load() != 0 || rebuilt.AsyncReadbackBytes != 0 {
		return errors.New("rebuild capture delivered or retained staging")
	}
	if _, err := capture(ctx, a, first, 2); err != nil {
		return err
	}
	// Minimize uses native main-queue methods, never while holding Owner.Do.
	native.Minimize()
	if err := waitCondition(ctx, a, func() bool { return minimized(native) }); err != nil {
		return fmt.Errorf("native minimize unavailable: %w", err)
	}
	minimizedBefore, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool { return !s.SnapshotPending })
	if err != nil {
		return err
	}
	// All observations happen while still minimized; restore may legitimately
	// change the visible clip and trigger a later frame.
	defer restore(native)
	minimizedCapture, err := capture(ctx, a, first, 2)
	if err != nil {
		return err
	}
	minimizedAfter, err := waitStats(ctx, a, first, func(s cogent.GPUViewStats) bool { return !s.SnapshotPending })
	if err != nil {
		return err
	}
	if !minimized(native) || minimizedAfter.Frames != minimizedBefore.Frames || minimizedAfter.AsyncSnapshotPolls <= minimizedBefore.AsyncSnapshotPolls || minimizedCapture.Frame != minimizedBefore.Frames {
		return errors.New("minimized readback depended on new rendering")
	}
	restore(native)
	if err := owner(a, func() error {
		s, e := second.view.Stats()
		if e != nil {
			return e
		}
		if s.Frames != independent.Frames || s.BufferAllocations != independent.BufferAllocations {
			return errors.New("first window modified independent resources")
		}
		return nil
	}); err != nil {
		return err
	}
	// Direct native force-close bypasses CloseReq/normal scene-close initiation.
	// Closed Stats cannot expose private staging/reservation counters; the adapter
	// control-flow regression and owned-reference review cover that seam. Here
	// verify suppressed delivery
	// and usable surviving-window pixels, then complete domain cleanup normally.
	var otherNative system.Window
	otherLifetime, err := a.WindowContext(second.view.Window)
	if err != nil {
		return err
	}
	if err := owner(a, func() error {
		otherNative = second.view.NativeWindow()
		if otherNative == nil {
			return errors.New("second native window unavailable")
		}
		if err := second.view.SnapshotAsync(ctx, func(cogent.GPUSnapshot, error) { dropped.Add(1) }); err != nil {
			return err
		}
		// Direct native Close enqueues WinClose and native cleanup without
		// entering the owner. Keep delivery excluded through this admission/
		// close seam; CloseReq would run the scene-close policy instead.
		otherNative.Close()
		return nil
	}); err != nil {
		return err
	}
	if err := waitCondition(ctx, a, otherNative.IsClosed); err != nil {
		return err
	}
	if err := owner(a, func() error {
		if _, e := second.view.Stats(); !errors.Is(e, cogent.ErrFrameNotReady) {
			return fmt.Errorf("force-closed Stats: %v", e)
		}
		return nil
	}); err != nil {
		return err
	}
	if _, err := capture(ctx, a, first, 2); err != nil {
		return err
	}
	if dropped.Load() != 0 {
		return errors.New("closed capture delivered")
	}
	// Direct native Close bypasses the Scene/Body close policy and does not
	// promise domain removal. Complete that cleanup explicitly before testing
	// the normal close of the final registered window.
	if err := owner(a, func() error {
		if otherLifetime.Err() != nil {
			return nil
		}
		return a.CloseWindow(second.view.Window)
	}); err != nil {
		return err
	}
	if otherLifetime.Err() == nil {
		return errors.New("forced native window domain cleanup did not cancel its lifetime")
	}
	// Admit a capture on the last live domain window, then intentionally close
	// the application owner. The native handle and program counter stay usable
	// for a real teardown barrier even though public Stats no longer does.
	lifetime, err := a.WindowContext(first.view.Window)
	if err != nil {
		return err
	}
	_, _, releasesBeforeClose := first.model.resourceIdentity()
	if err := owner(a, func() error {
		windows := a.Windows()
		if len(windows) != 1 || windows[0].ID != first.view.Window {
			return errors.New("normal close is not the last registered window")
		}
		if e := first.view.SnapshotAsync(lifetime, func(cogent.GPUSnapshot, error) { dropped.Add(1) }); e != nil {
			return e
		}
		return a.CloseWindow(first.view.Window)
	}); err != nil {
		return err
	}
	if lifetime.Err() == nil || a.Context().Err() == nil || len(a.Windows()) != 0 {
		return errors.New("normal close did not cancel domain lifetimes and unregister windows")
	}
	terminalCtx, terminalCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer terminalCancel()
	if err := waitProgramTeardown(terminalCtx, native, first.model, releasesBeforeClose); err != nil {
		return err
	}
	// Desktop Lock acquires the teardown mutex, then rechecks its composer.
	// False after IsClosed establishes completed native teardown. This lock is
	// not context-cancelable; the process watchdog bounds a stuck native close.
	if native.Lock() {
		native.Unlock()
		return errors.New("closed native window still admitted drawing")
	}
	select {
	case e := <-nativeErrors:
		return e
	default:
	}
	if dropped.Load() != 0 {
		return errors.New("lifetime cancellation delivered capture")
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "pass", "compute": "256² storage -> single-sample RGBA16Float -> premultiplied tone-mapped sRGB", "exposure1": referencePixel(1, false, false), "exposure2": referencePixel(2, false, false), "transparentTop": true, "nativeAdmissionRefusals": true, "explicitSyncThenAsync": true, "partialAlpha": true, "initial": initial, "chrome": chrome, "exposure": exposed, "resized": resized, "capturedFrameBeforeResize": firstCapture.Frame, "oldCaptureMetadata": map[string]uint64{"Frame": old.snapshot.Frame, "Revision": old.snapshot.Revision, "ResourceGeneration": old.snapshot.ResourceGeneration, "TargetGeneration": old.snapshot.TargetGeneration}, "deliveryAfterResize": old.current.TargetGeneration > old.snapshot.TargetGeneration, "rebuilt": rebuilt, "minimized": minimizedAfter, "busyInsideCallback": true, "independentResources": true, "quietCancelRebuildClose": true, "normalClose": map[string]any{"nativeClosed": true, "nativeDrawingLockRefused": true, "windowLifetimeCanceled": true, "appLifetimeCanceled": true, "unregistered": true, "programReleased": true, "callbackCount": dropped.Load(), "privateSnapshotCleanup": "not measured by public consumer"}, "forceCloseDomainCleanup": "explicit CloseWindow after native closure", "forceClose": "no callback, closed Stats unavailable, survivor capture passes; private reservation control flow covered by adapter regression; buffer ownership reviewed"})
}
