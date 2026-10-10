// platform-smoke verifies target rendering, services and isolated window lifetime.
// A passing basic smoke is evidence for these checks, not full OS qualification.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/system"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

type report struct {
	Event             string   `json:"event"`
	OS                string   `json:"os"`
	Architecture      string   `json:"architecture"`
	Display           string   `json:"display"`
	Checks            []string `json:"checks"`
	QualificationGaps []string `json:"qualification_gaps"`
	Complete          bool     `json:"complete"`
}

// reportRecorder keeps failure reporting independent of worker termination.
// Snapshots own their slices, so JSON encoding never races with late progress.
type reportRecorder struct {
	mu    sync.Mutex
	value report
}

func (r *reportRecorder) add(checks ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value.Checks = append(r.value.Checks, checks...)
}
func (r *reportRecorder) snapshot() report {
	r.mu.Lock()
	defer r.mu.Unlock()
	value := r.value
	value.Checks = append([]string(nil), value.Checks...)
	value.QualificationGaps = append([]string(nil), value.QualificationGaps...)
	return value
}

func qualificationGaps(goos, display string) []string {
	switch goos {
	case "windows":
		return []string{"native IME/TSF text adapter", "native UI Automation accessibility bridge", "real hardware, DPI and interactive chooser/menu acceptance"}
	case "linux":
		gaps := []string{"native IBus/Fcitx composition adapter", "AT-SPI accessibility bridge", "real hardware, DPI and interactive chooser/menu acceptance"}
		if display != "x11" {
			gaps = append(gaps, "native Wayland backend; XWayland runtime must be identified separately")
		}
		return gaps
	default:
		return []string{"target-specific full qualification is outside this portable smoke"}
	}
}

// shownCanvasProbe publishes readiness from the actual native widget after
// scene/window initialization. Owner-posted readback work must not pump a window
// that is still constructing its native topology.
type shownCanvasProbe struct {
	renderer *cogent.GPUCanvas
	shown    *atomic.Bool
}

func (p *shownCanvasProbe) Mount(parent tree.Node) core.Widget {
	widget := p.renderer.Mount(parent)
	if widget != nil {
		widget.AsWidget().OnShow(func(events.Event) { p.shown.Store(true) })
	}
	return widget
}
func (p *shownCanvasProbe) Sync(widget core.Widget) error { return p.renderer.Sync(widget) }

func main() {
	clipboardWrite := flag.Bool("clipboard-write", false, "round trip and restore plain text clipboard (use disposable desktop)")
	requireComplete := flag.Bool("require-complete", false, "fail if any OS qualification gate remains")
	panel := flag.Bool("panel", true, "open a native file dialog and cancel it after one second")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	a, err := app.New(app.Config{})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	makeWindow := func(id string, views *atomic.Int32, shown *atomic.Bool) (app.Window, *cogent.GPUCanvas, *canvas.Scene) {
		model, err := canvas.New([]canvas.Rect{{ID: "probe", X: 0, Y: 0, Width: 100, Height: 60, Color: color.RGBA{67, 197, 166, 255}}})
		if err != nil {
			fail(err)
		}
		model.Fit()
		renderer := &cogent.GPUCanvas{Owner: a, Window: id, Model: model}
		return app.Window{ID: id, Title: "Vellum platform " + id, Width: 640, Height: 480, View: func() ui.Node {
			views.Add(1)
			return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Temporary platform qualification smoke"}, {ID: "canvas", Kind: ui.Extension, Grow: true, Custom: &shownCanvasProbe{renderer: renderer, shown: shown}}}}
		}}, renderer, model
	}
	var firstViews, secondViews atomic.Int32
	var firstShown, secondShown atomic.Bool
	first, firstGPU, firstModel := makeWindow("first", &firstViews, &firstShown)
	second, secondGPU, secondModel := makeWindow("second", &secondViews, &secondShown)
	if err := a.Open(first); err != nil {
		fail(err)
	}
	registry := command.New()
	if err := registry.Register(command.Command{ID: "probe", Label: "Platform probe", Execute: func(context.Context) error { return nil }}); err != nil {
		fail(err)
	}
	result := make(chan error, 1)
	display := "native"
	if runtime.GOOS == "linux" {
		display = "x11"
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			display = "x11_in_wayland_session"
		}
	}
	evidence := &reportRecorder{value: report{Event: "vellum.platform.smoke", OS: runtime.GOOS, Architecture: runtime.GOARCH, Display: display, QualificationGaps: qualificationGaps(runtime.GOOS, display)}}
	go func() {
		finish := func(err error) { result <- err; a.Close() }
		wait := func(test func() bool) error {
			ticker := time.NewTicker(25 * time.Millisecond)
			defer ticker.Stop()
			for !test() {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
				}
			}
			return nil
		}
		if err := wait(func() bool { _, err := a.Services(); return err == nil }); err != nil {
			finish(err)
			return
		}
		service, err := a.Services()
		if err != nil {
			finish(err)
			return
		}
		snapshot := func(renderer *cogent.GPUCanvas, model *canvas.Scene) error {
			for {
				measured := make(chan error, 1)
				if err := a.PostWait(ctx, func() {
					// Establish a narrow baseline projection, then change both
					// center and zoom and read back in this same owner callback.
					// No canvas Render can apply the camera between the mutation
					// and Snapshot: Snapshot's Sync must do it immediately.
					model.View.Zoom = 1
					model.View.Center = canvas.Point{X: 50, Y: 30}
					baseline, err := renderer.Snapshot()
					if err != nil {
						measured <- err
						return
					}
					if err := validateSnapshot(baseline, model.View); err != nil {
						measured <- err
						return
					}
					model.View.Zoom = 1.25
					model.View.Center = canvas.Point{X: 50 - model.View.Width/(4*model.View.Zoom), Y: 30 - model.View.Height/(4*model.View.Zoom)}
					p := model.View.Screen(canvas.Point{X: 50, Y: 30})
					if isProbePixel(baseline.NRGBAAt(int(p.X), int(p.Y))) {
						measured <- fmt.Errorf("camera baseline overlaps moved probe")
						return
					}
					img, err := renderer.Snapshot()
					if err != nil {
						measured <- err
						return
					}
					measured <- validateSnapshot(img, model.View)
				}); err != nil {
					return err
				}
				err := <-measured
				if !errors.Is(err, cogent.ErrFrameNotReady) {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(25 * time.Millisecond):
				}
			}
		}
		resizeAndVerify := func(renderer *cogent.GPUCanvas, model *canvas.Scene) error {
			// Native SetSize synchronously marshals to the main thread. Capture owner
			// handles first, then invoke it outside the owner to avoid lock inversion.
			for _, size := range []image.Point{image.Pt(800, 600), image.Pt(720, 540)} {
				var native system.Window
				var before canvas.Viewport
				if err := a.Do(func() { native = renderer.NativeWindow(); before = model.View }); err != nil {
					return err
				}
				beforeNative, err := nativePixelSize(native)
				if err != nil {
					return err
				}
				if beforeNative == size {
					return fmt.Errorf("resize qualification cannot establish transition: native size already %v", size)
				}
				native.SetSize(size)
				resizeCtx, stopResize := context.WithTimeout(ctx, 5*time.Second)
				ticker := time.NewTicker(25 * time.Millisecond)
				var observedNative image.Point
				var observed canvas.Viewport
				accepted := false
				for !accepted {
					observedNative, err = nativePixelSize(native)
					if err != nil {
						break
					}
					if err = a.Do(func() { observed = model.View }); err != nil {
						break
					}
					accepted = resizeAccepted(beforeNative, observedNative, size, before, observed)
					if accepted {
						break
					}
					select {
					case <-resizeCtx.Done():
						err = fmt.Errorf("resize qualification failed: requested %v from %v, native %v, viewport %gx%g -> %gx%g: %w", size, beforeNative, observedNative, before.Width, before.Height, observed.Width, observed.Height, resizeCtx.Err())
					case <-ticker.C:
					}
					if err != nil {
						break
					}
				}
				ticker.Stop()
				stopResize()
				if err != nil {
					return err
				}
				if err := snapshot(renderer, model); err != nil {
					return err
				}
			}
			return nil
		}

		if err := wait(firstShown.Load); err != nil {
			finish(err)
			return
		}
		if err := snapshot(firstGPU, firstModel); err != nil {
			finish(err)
			return
		}
		evidence.add("first window GPU dimensions and immediate camera readback")
		if err := resizeAndVerify(firstGPU, firstModel); err != nil {
			finish(err)
			return
		}
		evidence.add("first window GPU dimensions/camera after two resizes")
		if err := service.SetMenu(ctx, []app.MenuGroup{{Label: "Platform", Commands: []string{"probe"}}}); err != nil {
			finish(err)
			return
		}
		evidence.add("menu snapshot installation (activation requires interactive acceptance)")
		original, err := service.ReadClipboard(ctx)
		if err != nil {
			finish(err)
			return
		}
		evidence.add("plain text clipboard read")
		if *clipboardWrite {
			value := "Vellum clipboard 🧭 e\u0301\nfixture"
			if err := service.WriteClipboard(ctx, value); err != nil {
				finish(err)
				return
			}
			got, readErr := service.ReadClipboard(ctx)
			restoreErr := service.WriteClipboard(ctx, original)
			if readErr != nil || restoreErr != nil || got != value {
				finish(fmt.Errorf("clipboard roundtrip/restore: read=%v restore=%v equal=%v", readErr, restoreErr, got == value))
				return
			}
			evidence.add("Unicode text clipboard roundtrip and restore (text flavor only)")
		}
		if *panel {
			panelCtx, panelCancel := context.WithTimeout(ctx, time.Second)
			_, err := service.ChooseFile(panelCtx, app.FileOptions{Title: "Vellum temporary cancellation check", Extensions: []string{"txt"}})
			panelCancel()
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, app.ErrCanceled) {
				finish(fmt.Errorf("expected file chooser cancellation, got: %v", err))
				return
			}
			evidence.add("file chooser transport cancellation (display readiness/selection require acceptance)")
		}
		if err := a.Open(second); err != nil {
			finish(err)
			return
		}
		// The actual widget Show event publishes its initialized scene/window
		// before owner-posted Snapshot work can drain through another window.
		if err := wait(secondShown.Load); err != nil {
			finish(err)
			return
		}
		if err := snapshot(secondGPU, secondModel); err != nil {
			finish(err)
			return
		}
		if err := resizeAndVerify(secondGPU, secondModel); err != nil {
			finish(err)
			return
		}
		evidence.add("second window GPU dimensions/camera after two resizes before teardown")

		firstLife, err := a.WindowContext("first")
		if err != nil {
			finish(err)
			return
		}
		secondLife, err := a.WindowContext("second")
		if err != nil {
			finish(err)
			return
		}
		var firstNative system.Window
		if err := a.Do(func() { firstNative = firstGPU.NativeWindow() }); err != nil {
			finish(err)
			return
		}
		if firstNative == nil {
			finish(cogent.ErrFrameNotReady)
			return
		}
		if err := a.CloseWindow("first"); err != nil {
			finish(err)
			return
		}
		// Desktop IsClosed reads the atomic composer publication withdrawn
		// during native teardown, so this checks this exact window lifetime.
		if err := wait(firstNative.IsClosed); err != nil {
			finish(err)
			return
		}
		if firstLife.Err() == nil || secondLife.Err() != nil || a.Context().Err() != nil {
			finish(fmt.Errorf("window lifetime isolation failed"))
			return
		}
		if err := snapshot(secondGPU, secondModel); err != nil {
			finish(err)
			return
		}
		if err := resizeAndVerify(secondGPU, secondModel); err != nil {
			finish(err)
			return
		}
		evidence.add("second window GPU dimensions/camera after two resizes after first teardown", "isolated window lifetime")
		if err := a.CloseWindow("second"); err != nil {
			finish(err)
			return
		}
		if secondLife.Err() == nil || a.Context().Err() == nil {
			finish(fmt.Errorf("final window did not close application"))
			return
		}
		evidence.add("final window application shutdown")
		result <- nil
	}()
	err = collectRunResult(ctx, a.Run(ctx, cogent.Driver{Commands: registry}), result)
	finalEvidence := evidence.snapshot()
	if err != nil {
		finalEvidence.Event = "vellum.platform.failure"
		json.NewEncoder(os.Stdout).Encode(finalEvidence)
		fail(err)
	}
	finalEvidence.Complete = len(finalEvidence.QualificationGaps) == 0
	if err := json.NewEncoder(os.Stdout).Encode(finalEvidence); err != nil {
		fail(err)
	}
	if *requireComplete && !finalEvidence.Complete {
		fmt.Fprintln(os.Stderr, "full target qualification blocked by reported gaps")
		os.Exit(2)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

// nativePixelSize reads the native geometry publication under its own lock,
// outside the application owner. Size itself does not acquire that lock.
func nativePixelSize(window system.Window) (image.Point, error) {
	if window == nil || !window.Lock() {
		return image.Point{}, cogent.ErrFrameNotReady
	}
	defer window.Unlock()
	return window.Size(), nil
}

// resizeAccepted requires an actual requested native transition and completed
// canvas allocation changes on every resized axis. Canvas dimensions exclude
// labels/insets, so they need not equal the whole native window's dimensions.
func resizeAccepted(beforeNative, currentNative, requested image.Point, before, current canvas.Viewport) bool {
	if beforeNative == requested || currentNative != requested || current.Width <= 0 || current.Height <= 0 {
		return false
	}
	if beforeNative.X != requested.X && current.Width == before.Width {
		return false
	}
	if beforeNative.Y != requested.Y && current.Height == before.Height {
		return false
	}
	return true
}

// collectRunResult preserves a completed worker diagnostic alongside a native
// Run failure. A failed native loop must not wait for a worker that may still
// depend on it; only a successful Run waits for pending work or caller timeout.
func collectRunResult(ctx context.Context, runErr error, result <-chan error) error {
	select {
	case workerErr := <-result:
		return errors.Join(runErr, workerErr)
	default:
	}
	if runErr != nil {
		return runErr
	}
	select {
	case workerErr := <-result:
		return workerErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// validateSnapshot checks the completed owner viewport allocation and current
// model projection. A stale-size or stale-camera readback cannot pass this probe.
func validateSnapshot(img *image.NRGBA, view canvas.Viewport) error {
	if img == nil {
		return fmt.Errorf("GPU snapshot missing")
	}
	expected := image.Pt(int(view.Width), int(view.Height))
	if img.Bounds().Min != (image.Point{}) || img.Bounds().Size() != expected {
		return fmt.Errorf("GPU snapshot dimensions %v differ from owner viewport %v", img.Bounds(), expected)
	}
	p := view.Screen(canvas.Point{X: 50, Y: 30})
	point := image.Pt(int(p.X), int(p.Y))
	if !point.In(img.Bounds()) {
		return fmt.Errorf("probe outside GPU image: %v", point)
	}
	pixel := img.NRGBAAt(point.X, point.Y)
	if !isProbePixel(pixel) {
		return fmt.Errorf("GPU camera probe absent at %v: %v", point, pixel)
	}
	return nil
}

func isProbePixel(pixel color.NRGBA) bool {
	return pixel.G >= 100 && pixel.R <= 150 && pixel.B >= 80
}
