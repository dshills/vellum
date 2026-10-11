//go:build offscreen

package main

import (
	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"io/fs"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

// This verifies the actual routed/measured workload, without claiming native
// GPU or CPU timings. The existing fixture provides licensed embedded fonts,
// explicit clock/executor and a strictly bounded manual stepping loop.
func TestPerfActualViewportAndRoutedScroll(t *testing.T) {
	for _, variable := range []bool{false, true} {
		t.Run(fmt.Sprintf("variable%t", variable), func(t *testing.T) {

			a, e := app.New(app.Config{QueueCapacity: 64, CompletionCapacity: 8})
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(a.Close)

			d, e := generate(context.Background(), 10_000, nil)
			if e != nil {
				t.Fatal(e)
			}
			source, e := source("perf", d, 10_000)
			if e != nil {
				t.Fatal(e)
			}
			projection, e := virtual.BuildGridProjection(context.Background(), perfInput(source, variable), 2, virtual.GridQuery{}, limits())
			if e != nil {
				t.Fatal(e)
			}
			p := &perfProgram{owner: a, source: source, projection: projection}
			p.grid, e = virtual.NewGrid(projection, perfColumns(), virtual.GridConfig{RowHeight: 12, HeaderHeight: 24, FrozenColumns: 2, OverscanRows: 2, OverscanColumns: 1, MaxCells: 4096}, virtual.GridViewport{Width: 768, Height: 624})
			if e != nil {
				t.Fatal(e)
			}
			if e = a.Open(app.Window{ID: "perf", Title: "Perf actual fixture", Width: 1500, Height: 1250, View: p.describe}); e != nil {
				t.Fatal(e)
			}
			clock, e := execution.NewManualClock(time.Unix(1730000000, 0), 256, 256)
			if e != nil {
				t.Fatal(e)
			}
			executor, e := execution.NewManual(4)
			if e != nil {
				t.Fatal(e)
			}
			font, e := fs.ReadFile(fonts.Default, "noto/NotoSans-Regular.ttf")
			if e != nil {
				t.Fatal(e)
			}
			hash := sha256.Sum256(font)
			manual, e := cogent.NewManual(context.Background(), a, cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 64, LogicalDPI: 96, Theme: ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 15}), Foreground: ui.Some(color.NRGBA{R: 20, G: 20, B: 20, A: 255}), Background: ui.Some(color.NRGBA{R: 250, G: 250, B: 250, A: 255})}}, Fonts: []cogent.ManualFont{{Name: "noto/NotoSans-Regular.ttf", Data: font, SHA256: hex.EncodeToString(hash[:])}}})
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(func() {
				if e := manual.Close(context.Background()); e != nil {
					t.Error(e)
				}
				executor.Close()
				clock.Close()
			})

			settleManual(t, manual, 7)
			view, e := manual.Snapshot(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			var window cogent.ManualWindowSnapshot
			for _, w := range view.Windows {
				if w.ID == "perf" {
					window = w
				}
			}
			if len(window.Grids) != 1 {
				t.Fatalf("missing actual grid%+v", window)
			}
			g := window.Grids[0]
			var targetInc uint64
			for _, element := range window.Elements {
				if element.ID == "data" {
					targetInc = element.Incarnation
				}
			}
			if targetInc == 0 {
				t.Fatal("missing grid incarnation")
			}
			if _, e = checkPerfGeometry(g); e != nil {
				t.Fatal(e)
			}
			if g.Viewport.Width != 768 || g.Viewport.Height != 624 {
				t.Fatalf("actual viewport%+v", g.Viewport)
			}
			if e = manual.Input(context.Background(), cogent.ManualInput{Window: "perf", WindowIncarnation: window.Incarnation, Target: "data", TargetIncarnation: targetInc, Type: events.Scroll, Position: image.Pt((g.Bounds.Min.X+g.Bounds.Max.X)/2, (g.Bounds.Min.Y+g.Bounds.Max.Y)/2), Delta: math32.Vector2{Y: 12 * g.Scale}}); e != nil {
				t.Fatal(e)
			}
			settleManual(t, manual, 7)
			view, e = manual.Snapshot(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			for _, w := range view.Windows {
				if w.ID == "perf" {
					g = w.Grids[0]
				}
			}
			if _, e = checkPerfGeometry(g); e != nil {
				t.Fatal(e)
			}
			if g.Viewport.ScrollY != 12 || p.callbacks != 1 {
				t.Fatalf("real routed scroll%+v callbacks%d", g.Viewport, p.callbacks)
			}
		})
	}
}
