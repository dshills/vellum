//go:build offscreen

package main

import (
	"bytes"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/paint/pimage"
	"cogentcore.org/core/paint/render"
	"cogentcore.org/core/paint/renderers/rasterx"
	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"image/draw"
	"io/fs"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func runRoutedFixture(t testing.TB) {
	ctx := context.Background()
	a, p, e := newFixture()
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	clock, e := execution.NewManualClock(time.Unix(1730000000, 0), 256, 256)
	if e != nil {
		t.Fatal(e)
	}
	defer clock.Close()
	executor, e := execution.NewManual(4)
	if e != nil {
		t.Fatal(e)
	}
	defer executor.Close()
	font, e := fs.ReadFile(fonts.Default, "noto/NotoSans-Regular.ttf")
	if e != nil {
		t.Fatal(e)
	}
	hash := sha256.Sum256(font)
	m, e := cogent.NewManual(ctx, a, cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 64, LogicalDPI: 160, Commands: p.commands, Theme: ui.Theme{Default: ui.Style{Background: ui.Some(color.NRGBA{A: 255}), Foreground: ui.Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 15})}}, Fonts: []cogent.ManualFont{{Name: "noto/NotoSans-Regular.ttf", Data: font, SHA256: hex.EncodeToString(hash[:])}}})
	if e != nil {
		t.Fatal(e)
	}
	defer m.Close(ctx)
	settle := func() cogent.ManualWindowSnapshot {
		for range 128 {
			s, e := m.Snapshot(ctx)
			if e != nil {
				t.Fatal(e)
			}
			w := s.Windows[0]
			if s.Application.Ordinary+s.Application.ReadyCompletions > 0 || w.Events > 0 || w.ReadyPaint {
				if _, e = m.Step(ctx, 7); e != nil {
					t.Fatal(e)
				}
				continue
			}
			if w.LayoutPending {
				if _, e = m.Frame(ctx, windowID); e != nil {
					t.Fatal(e)
				}
				continue
			}
			return w
		}
		t.Fatal("bounded settle")
		return cogent.ManualWindowSnapshot{}
	}
	for i := uint64(0); i < 3; i++ {
		w := settle()
		if i > 0 {
			if e = m.Input(ctx, cogent.ManualInput{Window: windowID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Rune: ' ', Code: key.CodeSpacebar}); e != nil {
				t.Fatal(e)
			}
			w = settle()
			if _, e = clock.Advance(20 * time.Millisecond); e != nil {
				t.Fatal(e)
			}
			w = settle()
			if _, e = m.Frame(ctx, windowID); e != nil {
				t.Fatal(e)
			}
			w = settle()
		}
		if p.sequence != i {
			t.Fatalf("state %d/%d want%d", p.sequence, p.sequence, i)
		}
		image, e := m.Capture(ctx, windowID)
		if e != nil {
			t.Fatal(e)
		}
		var areaFound bool
		for _, el := range w.Elements {
			if el.ID != "patch" {
				continue
			}
			areaFound = true
			for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
				for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
					want := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
					if x >= el.Bounds.Min.X && x < el.Bounds.Max.X && y >= el.Bounds.Min.Y && y < el.Bounds.Max.Y {
						want = shade(i)
					}
					got := color.NRGBAModel.Convert(image.At(x, y)).(color.NRGBA)
					if got != want {
						t.Fatalf("pixel(%d,%d) got%v want%v;rect%v", x, y, got, want, el.Bounds)
					}
				}
			}
		}
		if !areaFound {
			t.Fatalf("patch geometry absent: %+v", w.Elements)
		}
	}
	if _, e = p.expect(99); e == nil {
		t.Fatal("no-op accepted")
	}
	w := settle()
	if e = m.Input(ctx, cogent.ManualInput{Window: windowID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Rune: 'x', Code: key.CodeX}); e != nil {
		t.Fatal(e)
	}
	settle()
	if p.sequence != 2 {
		t.Fatal("unrelated key mutated patch")
	}
}

// These fixtures own detached command copies. They are never installed in a
// native application, retained after the test, or used by shipping rendering.
type replayCapture struct {
	render.Renderer
	frames *[]render.Render
}

func (c *replayCapture) Release() {
	if rd, ok := c.Renderer.(interface{ Release() }); ok {
		rd.Release()
	}
}

func (c *replayCapture) Render(items render.Render) render.Renderer {
	c.Renderer.Render(items)
	if len(*c.frames) >= 8 || len(items) > 32 {
		panic("offscreen replay capture bound")
	}
	*c.frames = append(*c.frames, cloneReplay(items))
	return c
}
func cloneReplay(items render.Render) render.Render {
	dst := make(render.Render, 0, len(items))
	for _, item := range items {
		switch x := item.(type) {
		case *pimage.Params:
			if x.Cmd != pimage.Draw || x.Mask != nil || x.Original != nil {
				panic("unsupported image command")
			}
			u, ok := x.Source.(*image.Uniform)
			if !ok || u == nil {
				panic("unsupported image source")
			}
			switch u.C.(type) {
			case color.RGBA, color.NRGBA:
			default:
				panic("unsupported color")
			}
			y := *x
			y.Source = image.NewUniform(u.C)
			dst = append(dst, &y)
		case *render.Path:
			if x.Context.Style.HasFill() || x.Context.Style.HasStroke() || x.Context.Mask != nil {
				panic("painted or masked path outside fixture")
			}
			y := *x
			y.Path = append(y.Path[:0:0], x.Path...)
			y.Context.Bounds.Path = append(y.Context.Bounds.Path[:0:0], x.Context.Bounds.Path...)
			y.Context.ClipPath = append(y.Context.ClipPath[:0:0], x.Context.ClipPath...)
			dst = append(dst, &y)
		case *render.ContextPush:
			if x.Scroll != nil {
				panic("scroll outside fixture")
			}
			y := *x
			dst = append(dst, &y)
		case *render.ContextPop:
			y := *x
			dst = append(dst, &y)
		default:
			panic(fmt.Sprintf("unexpected command %T", item))
		}
	}
	return dst
}
func capturedReplay(t testing.TB) render.Render {
	t.Helper()
	old := paint.NewSourceRenderer
	var frames []render.Render
	paint.NewSourceRenderer = func(size math32.Vector2) render.Renderer { return &replayCapture{Renderer: old(size), frames: &frames} }
	defer func() { paint.NewSourceRenderer = old }()
	runRoutedFixture(t)
	if len(frames) != 4 {
		t.Fatalf("frames=%d want4", len(frames))
	}
	items := frames[len(frames)-1]
	counts := map[string]int{}
	for _, item := range items {
		counts[fmt.Sprintf("%T", item)]++
	}
	if !reflect.DeepEqual(counts, map[string]int{"*pimage.Params": 8, "*render.Path": 4, "*render.ContextPush": 4, "*render.ContextPop": 4}) {
		t.Fatalf("root shape %v", counts)
	}
	return items
}
func replayParts(items render.Render) (images, paths, final render.Render) {
	for _, item := range items {
		switch x := item.(type) {
		case *pimage.Params:
			images = append(images, item)
			final = render.Render{x}
		case *render.Path:
			paths = append(paths, item)
		}
	}
	return
}
func replayRenderer() render.Renderer { return rasterx.New(math32.Vec2(1519, 1266)) }
func legacyReplay(rs render.Renderer, items render.Render) {
	for _, item := range items {
		switch x := item.(type) {
		case *pimage.Params:
			x.Render(rs.Image().(*image.RGBA))
		case *render.Path:
			rs.(*rasterx.Renderer).RenderPath(x)
		}
	}
}
func replayPixels(t testing.TB, got, want image.Image) {
	t.Helper()
	a, b := got.(*image.RGBA), want.(*image.RGBA)
	if a.Rect != b.Rect || a.Stride != b.Stride || !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("full pixel image mismatch")
	}
}
func checkReplay(t testing.TB, items render.Render) {
	t.Helper()
	images, paths, final := replayParts(items)
	if len(images) != 8 || len(paths) != 4 || len(final) != 1 {
		t.Fatal("unexpected replay parts")
	}
	for _, item := range paths {
		p := item.(*render.Path)
		if len(p.Path) != 20 || p.Context.Cumulative != math32.Identity2() {
			t.Fatal("unexpected path")
		}
	}
	last := final[0].(*pimage.Params)
	if last.Rect != image.Rect(8, 8, 1511, 1258) || last.Op != draw.Over {
		t.Fatalf("last image rect=%v op=%v", last.Rect, last.Op)
	}
	before := cloneReplay(items)
	originalColor := last.Source.(*image.Uniform).C
	// A full pre-elision replay is an independent pixel oracle. Mutating just the
	// detached final uniform exercises successive colors without changing history.
	full, onlyImages, want, pathOnly := replayRenderer(), replayRenderer(), replayRenderer(), replayRenderer()
	for _, clr := range []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}, {R: 255, A: 255}} {
		last.Source.(*image.Uniform).C = clr
		legacyReplay(want, cloneReplay(items))
		full.Render(items)
		onlyImages.Render(images)
		replayPixels(t, full.Image(), want.Image())
		replayPixels(t, onlyImages.Image(), want.Image())
		blank := append([]byte(nil), pathOnly.Image().(*image.RGBA).Pix...)
		pathOnly.Render(paths)
		if !bytes.Equal(blank, pathOnly.Image().(*image.RGBA).Pix) {
			t.Fatal("unpainted paths wrote pixels")
		}
	}
	last.Source.(*image.Uniform).C = originalColor
	if !reflect.DeepEqual(items, before) {
		t.Fatal("replay changed command state")
	}
}
func TestRasterReplayPartsAndPixels(t *testing.T) {
	items := capturedReplay(t)
	checkReplay(t, items)
	rows := make([]map[string]any, 0, len(items))
	for i, item := range items {
		row := map[string]any{"Index": i, "Type": fmt.Sprintf("%T", item)}
		switch x := item.(type) {
		case *pimage.Params:
			u := x.Source.(*image.Uniform)
			r, g, b, a := u.C.RGBA()
			row["Rect"] = x.Rect
			row["SourcePos"] = x.SourcePos
			row["Command"] = x.Cmd
			row["Operation"] = x.Op
			row["ColorType"] = fmt.Sprintf("%T", u.C)
			row["UniformRGBA"] = [4]uint32{r, g, b, a}
		case *render.Path:
			row["PathLength"] = len(x.Path)
			row["Path"] = x.Path
			row["PathTransform"] = x.Context.Cumulative
			row["HasFill"] = x.Context.Style.HasFill()
			row["HasStroke"] = x.Context.Style.HasStroke()
		}
		rows = append(rows, row)
	}
	raw, e := json.Marshal(rows)
	if e != nil {
		t.Fatal(e)
	}
	t.Log("OFFSCREEN_REPLAY_SHAPE=" + string(raw))
}
func TestRasterReplayCaptureDetached(t *testing.T) {
	items := capturedReplay(t)
	copy := cloneReplay(items)
	imageIndex, pathIndex := -1, -1
	for i, item := range items {
		switch item.(type) {
		case *pimage.Params:
			imageIndex = i
		case *render.Path:
			pathIndex = i
		}
	}
	items[imageIndex].(*pimage.Params).Source.(*image.Uniform).C = color.NRGBA{B: 255, A: 255}
	items[pathIndex].(*render.Path).Path[1] = 1234
	if reflect.DeepEqual(copy[imageIndex], items[imageIndex]) || reflect.DeepEqual(copy[pathIndex], items[pathIndex]) {
		t.Fatal("command capture aliases mutable source/path")
	}
}
func BenchmarkRasterReplayBreakdown(b *testing.B) {
	if runtime.GOMAXPROCS(0) != 2 {
		b.Fatal("benchmark requires explicit -cpu=2")
	}
	items := capturedReplay(b)
	checkReplay(b, items)
	images, paths, final := replayParts(items)
	cases := []struct {
		name  string
		items render.Render
	}{{"full_replay", items}, {"image_commands", images}, {"no_paint_paths", paths}, {"last_solid_fill", final}, {"empty_replay", nil}}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			rs := replayRenderer()
			rs.Render(c.items)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				rs.Render(c.items)
			}
			b.StopTimer()
			runtime.KeepAlive(rs.Image())
		})
	}
}
