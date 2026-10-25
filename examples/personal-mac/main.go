// Personal Mac starter: existing public APIs only, with one native window.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
)

func toolbarArtwork() []byte {
	pixels := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			c := color.NRGBA{R: 35, G: 125, B: 235, A: 255}
			if x+y > 23 {
				c = color.NRGBA{R: 245, G: 175, B: 45, A: 255}
			}
			pixels.SetNRGBA(x, y, c)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, pixels); err != nil {
		panic(err)
	}
	return b.Bytes()
}
func run() error {
	exitAfter := flag.Duration("exit-after", 0, "optional owned smoke lifetime; zero keeps window open")
	flag.Parse()
	if *exitAfter < 0 || *exitAfter > time.Minute {
		return fmt.Errorf("exit-after must be zero to one minute")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	name := "My drawing"
	resets := 0
	ranges := ui.RangeSlider{ID: "distance", Label: "Distance", Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Value: ui.RangeValue{Lower: 20, Upper: 80}}
	ranges.Change = func(e ui.RangeEvent) {
		if next, err := e.Apply(ranges.Value, ranges.Numeric); err == nil {
			ranges.Value = next
		}
	}
	var ref app.WindowRef
	reset := func() {
		name = "My drawing"
		ranges.Value = ui.RangeValue{Lower: 20, Upper: 80}
		resets++
		_ = a.Invalidate(ref.ID())
	}
	ref, err = a.OpenWindow(app.Window{ID: "personal", Title: "Vellum Personal Starter", Width: 560, Height: 400, View: func() ui.Node {
		rangeNode, e := ranges.View()
		if e != nil {
			return ui.Node{ID: "error", Kind: ui.Label, Text: e.Error()}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(20))}, Children: []ui.Node{
			{ID: "title", Kind: ui.Label, Text: "A native Mac application from Go"},
			{ID: "name", Kind: ui.TextField, Text: name, Accessibility: ui.Accessibility{Name: "Drawing name"}, Change: func(e ui.ChangeEvent) { name = e.Text }},
			rangeNode, {ID: "reset", Kind: ui.Button, Text: "Reset", Activate: reset},
			{ID: "status", Kind: ui.Label, Text: fmt.Sprintf("Reset %d times", resets)},
		}}
	}})
	if err != nil {
		return err
	}
	if *exitAfter > 0 {
		timer := time.AfterFunc(*exitAfter, a.Close)
		defer timer.Stop()
	}
	toolbar := &cogent.TitlebarToolbar{Window: ref, Items: []cogent.TitlebarToolbarItem{{ID: "reset", Label: "Reset", Kind: ui.Button, PNG: toolbarArtwork(), Activate: reset}}}
	return a.Run(context.Background(), cogent.Driver{FollowSystemAppearance: true, TitlebarToolbar: toolbar})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
