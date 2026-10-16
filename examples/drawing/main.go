// Drawing is an independent public-SDK consumer. Drag pans; the wheel zooms.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func main() {
	smoke := flag.Bool("smoke", false, "run bounded native acceptance and close both fixture windows")
	flag.Parse()
	model, err := newDrawingModel()
	if err != nil {
		fail(err)
	}
	a, err := app.New(app.Config{})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	view := canvas.Viewport{Width: 800, Height: 600, Zoom: 1}
	g := &cogent.MeshCanvas{Owner: a, Window: "drawing", View: &view, Prepare: model.prepare}
	nativeErrors := make(chan error, 16)
	g.OnError = func(err error) {
		// Error callbacks already own model state and have released GPU locks.
		// In the live fixture, querying Stats here verifies that lock boundary.
		if *smoke && strings.Contains(err.Error(), "budget exceeded") {
			stats, queryErr := g.Stats()
			if queryErr != nil || !stats.Ready {
				err = fmt.Errorf("error callback could not query retained frame: ready=%v err=%v", stats.Ready, queryErr)
			}
		}
		select {
		case nativeErrors <- err:
		default:
		}
		fmt.Fprintln(os.Stderr, err)
	}
	status := "Triangle gradients, alpha overlap and a retained stroke"
	selectMode := func(mode string) {
		model.mode = mode
		g.Image = nil
		if mode == "image" {
			g.Image = model.imageLayer
		}
		view.Center = canvas.Point{}
		view.Zoom = 1
		if mode == "schematic" {
			view.Center = canvas.Point{X: 500, Y: 400}
		}
		g.Revision++
		status = map[string]string{"image": "4096×4096 RGBA; retained texture and changed 128-pixel tiles", "mesh": "Triangle gradients, alpha overlap and a retained stroke", "schematic": "100,000 source rectangles; visible geometry is queried from a retained index", "waveform": "1,000,000 source samples; min/max peaks reduced to visible columns"}[mode]
		a.Invalidate("drawing")
	}
	_, err = a.OpenWindow(app.Window{ID: "drawing", Title: "Vellum custom GPU drawing", Width: 800, Height: 600, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
			{ID: "modes", Kind: ui.Row, Children: []ui.Node{{ID: "mesh", Kind: ui.Button, Text: "Mesh and strokes", Activate: func() { selectMode("mesh") }}, {ID: "schematic", Kind: ui.Button, Text: "100k schematic", Activate: func() { selectMode("schematic") }}, {ID: "waveform", Kind: ui.Button, Text: "Million-sample waveform", Activate: func() { selectMode("waveform") }}, {ID: "image", Kind: ui.Button, Text: "4K tiled image", Activate: func() { selectMode("image") }}}},
			{ID: "status", Kind: ui.Label, Text: status}, {ID: "canvas", Kind: ui.Extension, Grow: true, Custom: g},
		}}
	}})
	if err != nil {
		fail(err)
	}
	var other *cogent.MeshCanvas
	var otherRef app.WindowRef
	if *smoke {
		otherView := canvas.Viewport{Width: 640, Height: 480, Zoom: 1}
		other = &cogent.MeshCanvas{Owner: a, Window: "drawing-other", View: &otherView, Mesh: model.fixture, OnError: g.OnError}
		if otherRef, err = a.OpenWindow(app.Window{ID: "drawing-other", Title: "Vellum independent GPU resources", Width: 640, Height: 480, View: func() ui.Node {
			return ui.Node{ID: "other-root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "other-canvas", Kind: ui.Extension, Grow: true, Custom: other}}}
		}}); err != nil {
			fail(err)
		}
	}
	result := make(chan error, 1)
	if *smoke {
		watchdog := time.AfterFunc(60*time.Second, func() { fmt.Fprintln(os.Stderr, "drawing smoke watchdog"); os.Exit(1) })
		defer watchdog.Stop()
		go func() {
			ctx, cancel := context.WithTimeout(a.Context(), 50*time.Second)
			defer cancel()
			result <- drawingSmoke(ctx, a, otherRef, g, other, model, nativeErrors, func() { status = "Chrome remains responsive" })
			a.Close()
		}()
	}
	if err := a.Run(context.Background(), cogent.Driver{}); err != nil {
		fail(err)
	}
	if *smoke {
		if err := <-result; err != nil {
			fail(err)
		}
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
