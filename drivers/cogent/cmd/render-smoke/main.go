// render-smoke checks demand wakeups and live GPU resource/window lifetimes.
package main

import (
	"cogentcore.org/core/system/driver/desktop"
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"fmt"
	"image"
	"image/color"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	watchdog := time.AfterFunc(20*time.Second, func() { fmt.Fprintln(os.Stderr, "render smoke timed out"); os.Exit(1) })
	defer watchdog.Stop()
	a, err := app.New(app.Config{})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	model, err := canvas.New([]canvas.Rect{{ID: "rect", Width: 100, Height: 60, Color: color.RGBA{67, 197, 166, 255}}})
	if err != nil {
		fail(err)
	}
	model.Fit()
	g := &cogent.GPUCanvas{Owner: a, Window: "render", Model: model}
	title := "Before idle"
	var updates atomic.Int32
	result := make(chan error, 1)
	err = a.Open(app.Window{ID: "render", Title: "Vellum rendering checks", Width: 640, Height: 480, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: title}, {ID: "canvas", Kind: ui.Extension, Grow: true, Custom: g}}}
	}})
	if err != nil {
		fail(err)
	}
	go func() {
		ctx, cancel := context.WithTimeout(a.Context(), 15*time.Second)
		defer cancel()
		var win *desktop.Window
		for win == nil {
			var probeErr error
			err := postAndWait(ctx, a, func() {
				_, probeErr = g.Snapshot()
				if probeErr == nil {
					if native, ok := g.NativeWindow().(*desktop.Window); ok && native != nil {
						win = native
					} else {
						probeErr = fmt.Errorf("render smoke: native desktop window unavailable")
					}
				}
			})
			if err != nil {
				result <- err
				a.Close()
				return
			}
			if probeErr != nil {
				select {
				case <-ctx.Done():
					result <- ctx.Err()
					a.Close()
					return
				case <-time.After(25 * time.Millisecond):
				}
			}
		}
		check := func() error {
			if !win.IsOnDemand() {
				return fmt.Errorf("requested painting not enabled")
			}
			time.Sleep(time.Second)
			before := win.Paint.Wakes.Load()
			time.Sleep(time.Second)
			idleWakes := win.Paint.Wakes.Load() - before
			if idleWakes != 0 {
				return fmt.Errorf("unconditional idle painting: %d wakes", idleWakes)
			}
			var e error
			if err := postAndWait(ctx, a, func() {
				title = "After idle"
				updates.Add(1)
				a.Invalidate("render")
				e = g.RebuildResources()
				if e == nil {
					_, e = g.Snapshot()
				}
			}); err != nil {
				return err
			}
			if e != nil {
				return e
			}
			win.SetSize(image.Pt(800, 600))
			time.Sleep(500 * time.Millisecond)
			if err := postAndWait(ctx, a, func() { _, e = g.Snapshot() }); err != nil {
				return err
			}
			if e != nil {
				return e
			}
			win.Minimize()
			time.Sleep(250 * time.Millisecond)
			win.Raise()
			time.Sleep(500 * time.Millisecond)
			if err := postAndWait(ctx, a, func() { _, e = g.Snapshot() }); err != nil {
				return err
			}
			if e != nil {
				return e
			}
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.render.smoke", "on_demand": true, "idle_paint_wakes": idleWakes, "owner_update_after_idle": updates.Load(), "resource_rebuild": true, "resize": true, "minimize_restore": true, "physical_device_loss_verified": false})
		}
		result <- check()
		a.Close()
	}()
	if err := a.Run(context.Background(), cogent.Driver{}); err != nil {
		fail(err)
	}
	if err := <-result; err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

func postAndWait(ctx context.Context, a *app.Application, fn func()) error {
	done := make(chan struct{})
	if err := a.PostWait(ctx, func() { fn(); close(done) }); err != nil {
		return err
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
