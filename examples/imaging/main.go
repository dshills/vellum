// Imaging is an independent shared-device compute and floating-point consumer.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"os"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type imageWindow struct {
	application    *app.Application
	windowRef      app.WindowRef
	model          *imagingProgram
	view           *cogent.GPUView
	shown          atomic.Bool
	status         string // UI owner only
	captureBusy    bool   // includes optional worker save, beyond adapter callback
	captureResults chan captureResult
	interactive    bool
	nativeErrors   <-chan error
}
type captureResult struct {
	snapshot cogent.GPUSnapshot
	err      error
}
type shownView struct{ window *imageWindow }

func (v *shownView) Mount(parent tree.Node) core.Widget {
	widget := v.window.view.Mount(parent)
	if widget != nil {
		widget.AsWidget().OnShow(func(events.Event) { v.window.shown.Store(true) })
	}
	return widget
}
func (v *shownView) Sync(widget core.Widget) error { return v.window.view.Sync(widget) }

func newImageWindow(a *app.Application, id string, interactive bool, errors chan error) (*imageWindow, app.Window) {
	model, program := newImagingProgram()
	w := &imageWindow{application: a, model: model, interactive: interactive, status: "HDR compute: exposure 1.00", captureResults: make(chan captureResult, 1), nativeErrors: errors}
	w.view = &cogent.GPUView{Owner: a, Window: id, Program: program, OnError: func(err error) {
		select {
		case errors <- err:
		default:
		}
		fmt.Fprintln(os.Stderr, err)
	}}
	extension := &shownView{window: w}
	edit := func(value float32) {
		if err := model.setExposure(value); err != nil {
			w.status = err.Error()
		} else {
			w.view.Revision++
			w.status = fmt.Sprintf("HDR compute: exposure %.2f", value)
		}
		_ = a.Invalidate(id)
	}
	spec := app.Window{ID: id, Title: "Vellum compute / HDR imaging", Width: 800, Height: 640, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
			{ID: "tools", Kind: ui.Row, Children: []ui.Node{
				{ID: "exposure-down", Kind: ui.Button, Text: "Exposure / 2", Activate: func() {
					value := model.exposureValue() / 2
					if value < .125 {
						value = .125
					}
					edit(value)
				}},
				{ID: "exposure-up", Kind: ui.Button, Text: "Exposure × 2", Activate: func() {
					value := model.exposureValue() * 2
					if value > 16 {
						value = 16
					}
					edit(value)
				}},
				{ID: "reset", Kind: ui.Button, Text: "Reset", Activate: func() { edit(1) }},
				{ID: "capture", Kind: ui.Button, Text: "Capture", Disabled: w.captureBusy, Activate: w.capture},
			}},
			{ID: "status", Kind: ui.Label, MinHeight: 24, Text: w.status},
			{ID: "image", Kind: ui.Extension, Grow: true, Custom: extension},
		}}
	}}
	return w, spec
}

// reportCapture is a bounded smoke diagnostic mailbox. An unread prior result
// takes precedence; dropping a newer diagnostic must never block the UI owner.
func (w *imageWindow) reportCapture(result captureResult) {
	if w.interactive {
		return
	}
	select {
	case w.captureResults <- result:
	default:
	}
}
func (w *imageWindow) capture() {
	if w.captureBusy {
		return
	}
	ctx := w.windowRef.Context()
	if ctx.Err() != nil {
		w.reportCapture(captureResult{err: app.ErrWindowClosed})
		return
	}
	w.captureBusy = true
	err := w.view.SnapshotAsync(ctx, func(snapshot cogent.GPUSnapshot, err error) {
		w.completeCapture(ctx, snapshot, err)
	})
	if err != nil {
		w.completeCapture(ctx, cogent.GPUSnapshot{}, err)
	}
	_ = w.application.Invalidate(w.view.Window)
}

// completeCapture runs on the UI owner for both admission errors and callback
// completion; save/chooser work remains on a worker with the window lifetime.
func (w *imageWindow) completeCapture(ctx context.Context, snapshot cogent.GPUSnapshot, err error) {
	if err != nil {
		w.captureBusy = false
		w.status = err.Error()
		w.reportCapture(captureResult{err: err})
		_ = w.application.Invalidate(w.view.Window)
		return
	}
	w.status = fmt.Sprintf("Captured frame %d: %d × %d", snapshot.Frame, snapshot.Image.Bounds().Dx(), snapshot.Image.Bounds().Dy())
	if !w.interactive {
		w.captureBusy = false
		w.reportCapture(captureResult{snapshot: snapshot})
		_ = w.application.Invalidate(w.view.Window)
		return
	}
	// The callback owns a detached CPU image. Save/chooser work is off the UI
	// owner, with one retained image while the ordinary Capture button is busy.
	go w.save(ctx, snapshot.Image)
	_ = w.application.Invalidate(w.view.Window)
}
func (w *imageWindow) save(ctx context.Context, img *image.NRGBA) {
	service, err := w.application.Services()
	path := ""
	if err == nil {
		path, err = service.ChooseFile(ctx, app.FileOptions{Mode: app.SaveFile, Title: "Save HDR tone-mapped capture", Filename: "vellum-imaging.png", Extensions: []string{"png"}})
	}
	if err == nil {
		err = savePNG(ctx, path, img)
	}
	message := "Capture canceled"
	if err == nil {
		message = "PNG capture saved"
	} else if !errors.Is(err, app.ErrCanceled) && !errors.Is(err, context.Canceled) {
		message = err.Error()
	}
	_ = w.application.PostWait(ctx, func() {
		// The captured window lifetime may end after mailbox admission.
		if ctx.Err() != nil {
			return
		}
		w.captureBusy = false
		w.status = message
		_ = w.application.Invalidate(w.view.Window)
	})
}

// Result publication is distinct from complete worker/application cleanup.
// The outer smoke watchdog remains active if shutdown or a worker gets stuck.
func startSmokeWorker(stop func(), work func() error) (<-chan error, <-chan struct{}) {
	result := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer stop()
		result <- work()
	}()
	return result, done
}

func run(smoke bool) error {
	if smoke {
		// Keep fixture geometry isolated through application/native teardown.
		previousGeometrySaver := core.DebugSettings.DisableWindowGeometrySaver
		core.DebugSettings.DisableWindowGeometrySaver = true
		defer func() { core.DebugSettings.DisableWindowGeometrySaver = previousGeometrySaver }()
	}
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	nativeErrors := make(chan error, 16)
	first, spec := newImageWindow(a, "imaging", !smoke, nativeErrors)
	if first.windowRef, err = a.OpenWindow(spec); err != nil {
		return err
	}
	var result <-chan error
	if smoke {
		second, other := newImageWindow(a, "imaging-other", false, nativeErrors)
		if second.windowRef, err = a.OpenWindow(other); err != nil {
			return err
		}
		var workerDone <-chan struct{}
		result, workerDone = startSmokeWorker(a.Close, func() error {
			ctx, cancel := context.WithTimeout(a.Context(), 50*time.Second)
			defer cancel()
			return imagingSmoke(ctx, a, first, second, nativeErrors)
		})
		defer func() {
			// On native failure, cancel owner/window work before waiting. Joining
			// covers every post-launch return and precedes geometry restoration.
			a.Close()
			<-workerDone
		}()
	}
	runErr := a.Run(context.Background(), cogent.Driver{})
	if !smoke {
		return runErr
	}
	select {
	case workerErr := <-result:
		return errors.Join(runErr, workerErr)
	default:
	}
	if runErr != nil {
		return runErr
	}
	select {
	case err := <-result:
		return err
	case <-time.After(2 * time.Second):
		return errors.New("smoke result unavailable after native shutdown")
	}
}
func main() {
	smoke := flag.Bool("smoke", false, "run bounded native compute/HDR and async-capture acceptance")
	flag.Parse()
	var watchdog *time.Timer
	if *smoke {
		watchdog = time.AfterFunc(60*time.Second, func() { fmt.Fprintln(os.Stderr, "imaging smoke watchdog"); os.Exit(1) })
		defer watchdog.Stop()
	}
	if err := run(*smoke); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
