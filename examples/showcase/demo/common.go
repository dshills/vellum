// Package demo contains application-owned models using only public Vellum APIs.
package demo

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

var white = rgb(255, 255, 255)
var ink = rgb(33, 36, 43)
var muted = rgb(122, 127, 138)
var line = rgb(225, 227, 233)
var blue = rgb(43, 108, 244)

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{r, g, b, 255} }
func label(id, text string, size float32, weight int, tint color.NRGBA) ui.Node {
	return ui.Node{ID: id, Kind: ui.Label, Text: text, Style: ui.Style{Typography: ui.Some(ui.MacTypography(size, weight)), Foreground: ui.Some(tint), Padding: ui.Some(ui.Insets{}), GrowY: ui.Some(float32(0))}}
}
func col(id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Children: children, Style: ui.Style{Gap: ui.Some(float32(12)), Padding: ui.Some(ui.Insets{}), GrowX: ui.Some(float32(1))}}
}
func row(id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: ui.Row, Children: children, Style: ui.Style{Gap: ui.Some(float32(10)), Align: ui.Some(ui.AlignCenter), Padding: ui.Some(ui.Insets{}), GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0))}}
}
func spacer(id string) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(1))}}
}
func width(n ui.Node, w float32) ui.Node {
	n.Style.MinSize = ui.Some(ui.Size{Width: w})
	n.Style.MaxSize = ui.Some(ui.Size{Width: w})
	n.Style.GrowX = ui.Some(float32(0))
	return n
}
func height(n ui.Node, h float32) ui.Node {
	n.Style.MinSize = ui.Some(ui.Size{Height: h})
	n.Style.MaxSize = ui.Some(ui.Size{Height: h})
	n.Style.GrowY = ui.Some(float32(0))
	return n
}
func pad(n ui.Node, p float32) ui.Node { n.Style.Padding = ui.Some(ui.UniformInsets(p)); return n }
func section(id, title string, children ...ui.Node) ui.Node {
	n := col(id, append([]ui.Node{label(id+"-title", title, 11, 600, muted)}, children...)...)
	n.Style.Gap = ui.Some(float32(9))
	return n
}
func button(id, text string, run func()) ui.Node {
	return ui.Node{ID: id, Kind: ui.Button, Text: text, Activate: run, Style: ui.Style{GrowY: ui.Some(float32(0))}}
}
func quiet(id, text string, run func()) ui.Node {
	n := button(id, text, run)
	n.Appearance = ui.ButtonQuiet
	n.Style.Border = ui.Some(ui.Border{})
	n.Style.Background = ui.Some(color.NRGBA{})
	n.Style.TextAlign = ui.Some(ui.AlignStart)
	n.Style.GrowX = ui.Some(float32(1))
	return n
}
func rule(id string) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 1}), MaxSize: ui.Some(ui.Size{Height: 1}), GrowX: ui.Some(float32(1)), Background: ui.Some(line)}}
}
func pill(id, text string, tint, bg color.NRGBA) ui.Node {
	n := label(id, text, 11, 500, tint)
	n.Style.Padding = ui.Some(ui.Insets{Top: 4, Bottom: 4, Left: 8, Right: 8})
	n.Style.Background = ui.Some(bg)
	n.Style.CornerRadius = ui.Some(float32(6))
	return n
}
func imageNode(id, name string, img *canvas.TiledImage, w, h float32) ui.Node {
	n := ui.Node{ID: id, Kind: ui.Image, Image: img, Fit: ui.ImageContain, Text: name, Accessibility: ui.Accessibility{Name: name}, Style: ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(1))}}
	if w > 0 {
		n = width(n, w)
	}
	if h > 0 {
		n = height(n, h)
	}
	return n
}
func nav(id, text string, selected bool, tint color.NRGBA, run func()) ui.Node {
	n := quiet(id, text, run)
	n.IconImage = icon(tint, "folder")
	n.Style.Gap = ui.Some(float32(12))
	n.Style.Align = ui.Some(ui.AlignCenter)
	n = height(n, 22)
	n.Style.Padding = ui.Some(ui.Insets{Top: 5, Bottom: 5, Left: 8, Right: 8})
	n.Style.CornerRadius = ui.Some(float32(7))
	if selected {
		n.Style.Background = ui.Some(rgb(225, 233, 248))
		n.Style.Foreground = ui.Some(blue)
	}
	return n
}
func theme(dark bool) ui.Theme {
	t := ui.MacTheme(dark)
	for _, k := range []ui.Kind{ui.Column, ui.Row} {
		s := t.Kinds[k]
		s.Gap = ui.Some(float32(0))
		s.Padding = ui.Some(ui.Insets{})
		t.Kinds[k] = s
	}
	return t
}

var iconCache sync.Map

func icon(tint color.NRGBA, kind string) *canvas.TiledImage {
	key := fmt.Sprint(tint, kind)
	if cached, ok := iconCache.Load(key); ok {
		return cached.(*canvas.TiledImage)
	}
	p := image.NewNRGBA(image.Rect(0, 0, 48, 48))
	for y := 4; y < 44; y++ {
		for x := 4; x < 44; x++ {
			if kind == "circle" && (x-24)*(x-24)+(y-24)*(y-24) > 19*19 {
				continue
			}
			if (x < 9 || x > 38) && (y < 9 || y > 38) {
				continue
			}
			c := tint
			if y < 20 {
				c.R = uint8((int(c.R) + 255) / 2)
				c.G = uint8((int(c.G) + 255) / 2)
				c.B = uint8((int(c.B) + 255) / 2)
			}
			if kind == "add" && (x >= 22 && x < 26 && y >= 13 && y < 35 || y >= 22 && y < 26 && x >= 13 && x < 35) {
				c = white
			}
			if kind == "folder" && y >= 22 && y < 34 && x >= 13 && x < 35 {
				c = rgb(237, 246, 255)
			}
			if kind == "reset" && (x >= 14 && x <= 18 && y >= 15 && y <= 33 || y >= 15 && y <= 19 && x >= 14 && x <= 33) {
				c = white
			}
			p.SetNRGBA(x, y, c)
		}
	}
	out, e := canvas.NewTiledImage(p)
	if e != nil {
		panic(e)
	}
	actual, _ := iconCache.LoadOrStore(key, out)
	return actual.(*canvas.TiledImage)
}
func toolbarIcon(tint color.NRGBA, kind string) []byte {
	var b bytes.Buffer
	if e := png.Encode(&b, icon(tint, kind)); e != nil {
		panic(e)
	}
	return b.Bytes()
}

type Application interface {
	View() ui.Node
	Bind(*app.Application, app.WindowRef)
	Toolbar() []cogent.TitlebarToolbarItem
	Commands() *command.Registry
}

func Run(kind string) {
	exitAfter := flag.Duration("exit-after", 0, "optional smoke lifetime; 0 keeps the app open, maximum 1m")
	flag.Parse()
	if *exitAfter < 0 || *exitAfter > time.Minute {
		panic("exit-after outside 0..1m")
	}
	if e := run(kind, *exitAfter); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run(kind string, exitAfter time.Duration) error {
	var m Application
	var e error
	title := ""
	dark := false
	switch kind {
	case "forma":
		m, e = NewForma()
		title = "Forma — Design Studio"
	case "luma":
		m, e = NewLuma()
		title = "Luma — Image Library"
	case "pulse":
		m, e = NewPulse()
		title = "Pulse — Telemetry Explorer"
		dark = true
	default:
		return fmt.Errorf("unknown demo %q", kind)
	}
	if e != nil {
		return e
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	a, e := app.New(app.Config{})
	if e != nil {
		return e
	}
	defer a.Close()
	appearance := app.AppearanceLight
	if dark {
		appearance = app.AppearanceDark
	}
	ref, e := a.OpenWindow(app.Window{ID: kind, Title: title, Width: 1400, Height: 880, View: m.View, DisableRootScrolling: true, Chrome: app.WindowChrome{Style: app.ChromeStandard, Appearance: appearance}})
	if e != nil {
		return e
	}
	m.Bind(a, ref)
	if exitAfter > 0 {
		timer := time.AfterFunc(exitAfter, a.Close)
		defer timer.Stop()
	}
	toolbar := &cogent.TitlebarToolbar{Window: ref, Items: m.Toolbar(), Commands: m.Commands()}
	return a.Run(ctx, cogent.Driver{Theme: theme(dark), TitlebarToolbar: toolbar, Commands: m.Commands(), OnError: func(e error) { fmt.Fprintln(os.Stderr, e) }})
}

type owner struct {
	application *app.Application
	ref         app.WindowRef
	commands    *command.Registry
}

func (o *owner) bind(a *app.Application, r app.WindowRef) { o.application = a; o.ref = r }
func (o *owner) changed() {
	if o.application != nil {
		_ = o.application.Invalidate(o.ref.ID())
	}
}
func (o *owner) Commands() *command.Registry { return o.commands }
func (o *owner) register(id, title, shortcut string, available func() error, run func() error) {
	if o.commands == nil {
		o.commands = command.New()
	}
	if e := o.commands.Register(command.Command{ID: id, Label: title, Shortcut: shortcut, Available: available, Execute: func(context.Context) error { e := run(); o.changed(); return e }}); e != nil {
		panic(e)
	}
}
