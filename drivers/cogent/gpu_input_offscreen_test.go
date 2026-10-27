//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/text/fonts"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/text/shaped/shapers/shapedgt"
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/oliverbestmann/webgpu/wgpu"
	"image"
	"io/fs"
	"testing"
	"time"
)

func TestGPUInputCapturedRoutingReleaseAndEscape(t *testing.T) {
	cfg := manualConfig(t, 7)
	ctx := context.Background()
	a, _ := app.New(app.Config{})
	defer a.Close()
	_ = a.Open(app.Window{ID: "gpu-route", Title: "route", Width: 600, Height: 400, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	var got []GPUEvent
	g := &GPUView{Owner: a, Window: "gpu-route", Program: &GPUProgram{Draw: func(GPUFrame, *wgpu.RenderPassEncoder) error { return nil }}, InputEvent: func(e GPUEvent) bool { got = append(got, e); return true }}
	b := core.NewBody("GPU input router")
	if g.Mount(b) == nil {
		t.Fatal("mount")
	}
	clicks := 0
	other := core.NewButton(b).SetText("Unrelated")
	other.OnClick(func(events.Event) { clicks++ })
	m, err := b.NewWindow().RunManual(ctx, core.ManualWindowConfig{Size: image.Pt(600, 400), LogicalDPI: 160, EventCapacity: 128, Now: cfg.Clock.Now, AfterFunc: func(c context.Context, d time.Duration, f func()) (core.ManualTimer, error) {
		return cfg.Clock.AfterFunc(c, d, f)
	}, NewShaper: func() (shaped.Shaper, error) { return shapedgt.NewShaperWithFonts([]fs.FS{fonts.Default}), nil }})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(ctx)
	if _, err = m.Frame(ctx); err != nil {
		t.Fatal(err)
	}
	inject := func(input core.ManualInput) {
		t.Helper()
		if e := m.Input(ctx, input); e != nil {
			t.Fatal(e)
		}
		if _, e := m.Step(ctx, 64); e != nil {
			t.Fatal(e)
		}
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
	}
	box := g.widget.Geom.ContentBBox
	pt := box.Min.Add(box.Size().Div(2))
	target := other.Geom.ContentBBox.Min.Add(other.Geom.ContentBBox.Size().Div(2))
	inject(core.ManualInput{Type: events.MouseDown, Button: events.Left, Position: pt})
	if g.widget.inputCapture == 0 {
		t.Fatal("not captured", box)
	}
	inject(core.ManualInput{Type: events.MouseMove, Position: pt.Add(image.Pt(20, 5))})
	inject(core.ManualInput{Type: events.MouseUp, Button: events.Left, Position: pt.Add(image.Pt(20, 5))})
	counts := map[GPUEventKind]int{}
	for _, e := range got {
		counts[e.Kind]++
	}
	if counts[GPUPointerDown] != 1 || counts[GPUPointerMove] != 1 || counts[GPUPointerUp] != 1 || g.widget.inputCapture != 0 {
		t.Fatal("duplicate/missing delivery", got)
	}
	got = nil
	inject(core.ManualInput{Type: events.MouseDown, Button: events.Left, Position: pt})
	inject(core.ManualInput{Type: events.KeyChord, Code: key.CodeEscape})
	inject(core.ManualInput{Type: events.MouseUp, Button: events.Left, Position: target})
	counts = map[GPUEventKind]int{}
	for _, e := range got {
		counts[e.Kind]++
	}
	if counts[GPUCancel] != 1 || counts[GPUPointerUp] != 0 || clicks != 0 || g.widget.inputCapture != 0 {
		t.Fatal("canceled release propagated", got, clicks)
	}
}
