// This executable consumes both Vellum modules from its own go.mod.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"runtime"
	"time"

	"cogentcore.org/core/core"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func main() {
	benchmark := flag.Bool("benchmark", false, "emit native startup/readback metrics; close after idle period")
	idle := flag.Duration("idle", 3*time.Second, "quiet interval after benchmark GPU measurements")
	samples := flag.Int("samples", 10, "synchronous GPU render/readback samples")
	flag.Parse()
	if *idle <= 0 || *samples < 1 || *samples > 1000 {
		fail(errors.New("positive idle and 1..1000 samples required"))
	}
	// Include settings initialization in startup, as the former lazy Body path did.
	started := time.Now()
	if *benchmark {
		restoreGeometry, err := isolateBenchmarkGeometry()
		if err != nil {
			fail(err)
		}
		// Registered before app.Close, so the loaded policy is restored after teardown.
		defer restoreGeometry()
	}
	a, err := app.New(app.Config{})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	scene, err := canvas.New([]canvas.Rect{{ID: "rectangle", X: 0, Y: 0, Width: 100, Height: 60, Color: color.RGBA{67, 197, 166, 255}}})
	if err != nil {
		fail(err)
	}
	scene.Fit()
	renderer := &cogent.GPUCanvas{Owner: a, Window: "hello", Model: scene}
	model := &counter{}
	err = a.Open(app.Window{ID: "hello", Title: "Vellum Hello", Width: 640, Height: 480, View: func() ui.Node {
		return model.view(ui.Node{ID: "canvas", Kind: ui.Extension, Grow: true, Custom: renderer})
	}})
	if err != nil {
		fail(err)
	}
	result := make(chan error, 1)
	if *benchmark {
		go func() {
			ctx, cancel := context.WithTimeout(a.Context(), 30*time.Second)
			defer cancel()
			ticker := time.NewTicker(25 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					result <- ctx.Err()
					a.Close()
					return
				case <-ticker.C:
				}
				measured := make(chan error, 1)
				err := a.PostWait(ctx, func() {
					image, err := renderer.Snapshot()
					if err != nil {
						measured <- err
						return
					}
					point := scene.View.Screen(canvas.Point{X: 50, Y: 30})
					if point.X < 0 || point.Y < 0 || int(point.X) >= image.Bounds().Dx() || int(point.Y) >= image.Bounds().Dy() {
						measured <- errors.New("rectangle outside rendered image")
						return
					}
					pixel := image.NRGBAAt(int(point.X), int(point.Y))
					if pixel.G < 100 || pixel.R > 150 || pixel.B < 80 {
						measured <- fmt.Errorf("rectangle absent from GPU image: %v", pixel)
						return
					}
					startup := time.Since(started).Nanoseconds()
					durations := make([]int64, *samples)
					for i := range durations {
						start := time.Now()
						_, err := renderer.Snapshot()
						if err != nil {
							measured <- err
							return
						}
						durations[i] = time.Since(start).Nanoseconds()
					}
					var memory runtime.MemStats
					runtime.ReadMemStats(&memory)
					err = json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.native.ready", "geometry_persistence_disabled_in_fixture": core.DebugSettings.DisableWindowGeometrySaver, "startup_to_verified_gpu_ns": startup, "render_readback_ns": durations, "heap_alloc_bytes": memory.HeapAlloc, "width": image.Bounds().Dx(), "height": image.Bounds().Dy(), "go_version": runtime.Version()})
					measured <- err
				})
				if err != nil {
					result <- err
					a.Close()
					return
				}
				select {
				case <-ctx.Done():
					result <- ctx.Err()
					a.Close()
					return
				case err := <-measured:
					if errors.Is(err, cogent.ErrFrameNotReady) {
						continue
					}
					result <- err
					if err != nil {
						a.Close()
					} else {
						time.AfterFunc(*idle, a.Close)
					}
					return
				}
			}
		}()
	}
	if err := a.Run(context.Background(), cogent.Driver{}); err != nil {
		fail(err)
	}
	if *benchmark {
		select {
		case err := <-result:
			if err != nil {
				fail(err)
			}
		default:
			fail(errors.New("window closed before benchmark completed"))
		}
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

// isolateBenchmarkGeometry is local because generated applications cannot import
// the driver's internal nativefixture package. Call once before app construction.
// Load failures are terminal: Core may apply defaults before returning an error;
// this fixture does not roll them back and must not be retried or ignored.
func isolateBenchmarkGeometry() (restore func(), err error) {
	if core.AppearanceSettings.Zoom == 0 {
		if err := core.LoadAllSettings(); err != nil {
			return nil, err
		}
	}
	zoom := float64(core.AppearanceSettings.Zoom)
	if zoom <= 0 || math.IsNaN(zoom) || math.IsInf(zoom, 0) {
		return nil, errors.New("hello benchmark: positive initialized Core zoom required")
	}
	previous := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	return func() { core.DebugSettings.DisableWindowGeometrySaver = previous }, nil
}
