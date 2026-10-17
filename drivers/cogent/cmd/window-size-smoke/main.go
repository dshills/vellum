//go:build darwin && !offscreen

// window-size-smoke observes initial native client geometry and a later shrink.
// It disables geometry persistence only within its own process. It does not
// synthesize input or read/modify other applications or macOS preferences.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"math"
	"os"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type observation struct {
	Event                 string           `json:"event"`
	PID                   int              `json:"pid"`
	RequestedDP           [2]float32       `json:"requested_dp"`
	Screen                system.Screen    `json:"screen"`
	PixelSize             image.Point      `json:"pixel_size"`
	WindowSize            image.Point      `json:"window_size"`
	RenderGeometry        math32.Geom2DInt `json:"render_geometry"`
	GeometrySaverDisabled bool             `json:"geometry_saver_disabled"`
}

func emit(v any) error { return json.NewEncoder(os.Stdout).Encode(v) }
func main() {
	width := flag.Float64("width", 320, "initial width in Dp")
	height := flag.Float64("height", 240, "initial height in Dp")
	flag.Parse()
	for _, v := range []float64{*width, *height} {
		if !(v >= 100 && v <= 1200) || math.IsInf(v, 0) {
			fmt.Fprintln(os.Stderr, "fixture dimensions must be 100..1200 Dp")
			os.Exit(2)
		}
	}
	watchdog := time.AfterFunc(25*time.Second, func() { fmt.Fprintln(os.Stderr, "window sizing watchdog"); os.Exit(1) })
	defer watchdog.Stop()
	core.DebugSettings.DisableWindowGeometrySaver = true
	a, err := app.New(app.Config{})
	if err != nil {
		panic(err)
	}
	defer a.Close()
	title := "Vellum initial window size check"
	if err = a.Open(app.Window{ID: "size", Title: title, Width: float32(*width), Height: float32(*height), View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column, Grow: true} }}); err != nil {
		panic(err)
	}
	result := make(chan error, 1)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		var err error
		var native system.Window
		read := func(event string) (observation, error) {
			var value observation
			var queryErr error
			executed := false
			desktop.TheApp.RunOnMainQueued(func() {
				executed = true
				if desktop.TheApp.NWindows() != 1 {
					queryErr = fmt.Errorf("fixture window not ready")
					return
				}
				w := desktop.TheApp.Window(0)
				if w.Title() != title {
					queryErr = fmt.Errorf("fixture window title mismatch")
					return
				}
				screen := w.Screen()
				if screen == nil || !w.Lock() {
					queryErr = fmt.Errorf("fixture screen/window unavailable")
					return
				}
				value = observation{Event: event, PID: os.Getpid(), RequestedDP: [2]float32{float32(*width), float32(*height)}, Screen: *screen, PixelSize: w.Size(), WindowSize: w.WinSize(), RenderGeometry: w.RenderGeom(), GeometrySaverDisabled: core.DebugSettings.DisableWindowGeometrySaver}
				w.Unlock()
				native = w
			})
			if !executed {
				return value, fmt.Errorf("native main queue stopped")
			}
			return value, queryErr
		}
		finish := func(e error) { result <- e; a.Close() }
		deadline := time.Now().Add(8 * time.Second)
		var initial observation
		for time.Now().Before(deadline) {
			initial, err = read("vellum.window.initial")
			if err == nil && initial.PixelSize.X > 0 && initial.PixelSize.Y > 0 {
				break
			}
			time.Sleep(25 * time.Millisecond)
		}
		if err != nil || initial.PixelSize.X <= 0 || initial.PixelSize.Y <= 0 {
			finish(fmt.Errorf("initial client unavailable: %v", err))
			return
		}
		if err = emit(initial); err != nil {
			finish(err)
			return
		}
		// A separate bounded, read-only AX client may observe the outer frame here.
		time.Sleep(4 * time.Second)
		target := image.Pt(max(1, initial.WindowSize.X/2), max(1, initial.WindowSize.Y/2))
		// Native resize is outside the application owner/render locks.
		native.SetWinSize(target)
		deadline = time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			var resized observation
			resized, err = read("vellum.window.resized")
			if err == nil && resized.WindowSize == target && resized.PixelSize.X < initial.PixelSize.X && resized.PixelSize.Y < initial.PixelSize.Y {
				if err = emit(resized); err != nil {
					finish(err)
					return
				}
				finish(nil)
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
		finish(fmt.Errorf("later resize did not reach %v", target))
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 22*time.Second)
	defer cancel()
	runErr := a.Run(ctx, cogent.Driver{})
	var checkErr error
	select {
	case checkErr = <-result:
	case <-ctx.Done():
		checkErr = ctx.Err()
	}
	select {
	case <-workerDone:
	case <-ctx.Done():
		checkErr = ctx.Err()
	}
	if runErr != nil || checkErr != nil {
		fmt.Fprintln(os.Stderr, runErr, checkErr)
		os.Exit(1)
	}
	if err = emit(map[string]any{"event": "vellum.window.complete", "pid": os.Getpid(), "physical_input_verified": false}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
