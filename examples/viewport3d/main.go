// Viewport3d is an independent consumer of the advanced shared-device GPU API.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"math"
	"os"
	"strings"
	"time"
)

func main() {
	smoke := flag.Bool("smoke", false, "run bounded native depth/resource acceptance")
	flag.Parse()
	a, err := app.New(app.Config{})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	m, program := newCubeProgram()
	g := &cogent.GPUView{Owner: a, Window: "viewport3d", Program: program, Depth: true}
	nativeErrors := make(chan error, 32)
	g.OnError = func(err error) {
		if *smoke && strings.Contains(err.Error(), "budget exceeded") {
			stats, e := g.Stats()
			if e != nil || !stats.Ready {
				err = fmt.Errorf("OnError retained-frame query: %+v %v", stats, e)
			}
		}
		select {
		case nativeErrors <- err:
		default:
		}
		fmt.Fprintln(os.Stderr, err)
	}
	change := func() { g.Revision++; a.Invalidate(g.Window) }
	g.Input = func(in cogent.GPUInput) bool {
		consumed := false
		if in.Delta.X != 0 || in.Delta.Y != 0 {
			m.camera.orbit(float64(in.Delta.X)*.01, float64(in.Delta.Y)*.01)
			change()
			consumed = true
		}
		if in.Scroll != 0 {
			m.camera.dolly(math.Exp(float64(in.Scroll) * .003))
			change()
			consumed = true
		}
		switch strings.ToLower(in.Key) {
		case "leftarrow":
			consumed = true
			m.camera.orbit(-.1, 0)
			change()
		case "rightarrow":
			consumed = true
			m.camera.orbit(.1, 0)
			change()
		case "uparrow":
			consumed = true
			m.camera.orbit(0, -.1)
			change()
		case "downarrow":
			consumed = true
			m.camera.orbit(0, .1)
			change()
		}
		return consumed
	}
	status := "Drag or use arrows to orbit; scroll to dolly"
	primaryRef, err := a.OpenWindow(app.Window{ID: g.Window, Title: "Vellum shared-device 3D viewport", Width: 800, Height: 600, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
			{ID: "tools", Kind: ui.Row, Children: []ui.Node{{ID: "orbit", Kind: ui.Button, Text: "Orbit", Activate: func() { m.camera.orbit(.3, .1); change() }}, {ID: "dolly", Kind: ui.Button, Text: "Dolly", Activate: func() { m.camera.dolly(.85); change() }}, {ID: "reset", Kind: ui.Button, Text: "Reset camera", Activate: func() { m.camera = defaultCamera(); change() }}}},
			{ID: "status", Kind: ui.Label, Text: status}, {ID: "viewport", Kind: ui.Extension, Grow: true, Custom: g},
		}}
	}})
	if err != nil {
		fail(err)
	}
	var other *cogent.GPUView
	var otherModel *cubeProgram
	var otherRef app.WindowRef
	if *smoke {
		var p *cogent.GPUProgram
		otherModel, p = newCubeProgram()
		other = &cogent.GPUView{Owner: a, Window: "viewport3d-other", Program: p, Depth: true, OnError: g.OnError}
		if otherRef, err = a.OpenWindow(app.Window{ID: other.Window, Title: "Vellum independent 3D resources", Width: 640, Height: 480, View: func() ui.Node {
			return ui.Node{ID: "other", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "other-viewport", Kind: ui.Extension, Grow: true, Custom: other}}}
		}}); err != nil {
			fail(err)
		}
	}
	result := make(chan error, 1)
	if *smoke {
		watchdog := time.AfterFunc(60*time.Second, func() { fmt.Fprintln(os.Stderr, "viewport smoke watchdog"); os.Exit(1) })
		defer watchdog.Stop()
		go func() {
			ctx, cancel := context.WithTimeout(a.Context(), 50*time.Second)
			defer cancel()
			result <- viewportSmoke(ctx, a, primaryRef, otherRef, g, other, m, otherModel, nativeErrors, func() { status = "Chrome remains responsive" })
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
