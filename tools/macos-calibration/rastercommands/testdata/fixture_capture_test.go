//go:build offscreen

package main

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/paint/renderers/rasterx"
	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"image/color"
	"io/fs"
	"testing"
	"time"
)

func TestRoutedStateAndExactPatchPixels(t *testing.T) {
	rasterx.VellumCommandReset()
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
	m, e := cogent.NewManual(ctx, a, cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 64, LogicalDPI: 96, Commands: p.commands, Theme: ui.Theme{Default: ui.Style{Background: ui.Some(color.NRGBA{A: 255}), Foreground: ui.Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 15})}}, Fonts: []cogent.ManualFont{{Name: "noto/NotoSans-Regular.ttf", Data: font, SHA256: hex.EncodeToString(hash[:])}}})
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
	capture := rasterx.VellumCommandSnapshot()
	if capture.Overflow || len(capture.Frames) == 0 {
		t.Fatalf("missing/overflow command capture: %+v", capture)
	}
	raw, err := json.Marshal(capture)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("RASTER_COMMANDS_JSON=" + string(raw))

}
