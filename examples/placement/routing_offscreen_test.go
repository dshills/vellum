//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

// This separate module exercises the actual managed widgets through public APIs.
// It creates software windows only and is not physical pointer certification.
func TestPublicPaletteCanvasConsumer(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			ctx := context.Background()
			clock, err := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
			if err != nil {
				t.Fatal(err)
			}
			defer clock.Close()
			executor, err := execution.NewManual(8)
			if err != nil {
				t.Fatal(err)
			}
			defer executor.Close()
			m, err := newModel()
			if err != nil {
				t.Fatal(err)
			}
			defer m.palette.Close()
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if err := a.Open(app.Window{ID: "placement", Title: "Placement", Width: 1100, Height: 760, View: func() ui.Node { return view(a, m) }}); err != nil {
				t.Fatal(err)
			}
			config := cogent.ManualConfig{Clock: clock, Executor: executor, Commands: m.commands, LogicalDPI: 160, OwnerBatch: batch, MaxWindows: 2, EventCapacity: 128, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
			if err := fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || (!strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf")) {
					return nil
				}
				data, err := fs.ReadFile(fonts.Default, path)
				if err != nil {
					return err
				}
				digest := sha256.Sum256(data)
				config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(digest[:])})
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			driver, err := cogent.NewManual(ctx, a, config)
			if err != nil {
				t.Fatal(err)
			}
			defer driver.Close(ctx)
			settle := func() {
				for range 4096 {
					s, err := driver.Snapshot(ctx)
					if err != nil {
						t.Fatal(err)
					}
					ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
					for _, w := range s.Windows {
						ready = ready || w.Events > 0 || w.ReadyPaint
						if !ready && w.LayoutPending {
							if err := driver.RequestFrame(ctx, w.ID); err != nil {
								t.Fatal(err)
							}
							ready = true
						}
					}
					if !ready {
						return
					}
					if _, err := driver.Step(ctx, batch); err != nil {
						t.Fatal(err)
					}
				}
				t.Fatal("dispatch bound exhausted")
			}
			element := func(id string) (cogent.ManualWindowSnapshot, cogent.ManualElement) {
				s, err := driver.Snapshot(ctx)
				if err != nil {
					t.Fatal(err)
				}
				for _, w := range s.Windows {
					for _, e := range w.Elements {
						if e.ID == id {
							return w, e
						}
					}
				}
				t.Fatal("missing control", id)
				return cogent.ManualWindowSnapshot{}, cogent.ManualElement{}
			}
			input := func(i cogent.ManualInput) {
				w, _ := element("root")
				i.Window = w.ID
				i.WindowIncarnation = w.Incarnation
				if err := driver.Input(ctx, i); err != nil {
					t.Fatalf("input %+v: %v", i, err)
				}
				settle()
			}
			click := func(id string) {
				_, e := element(id)
				for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
					input(cogent.ManualInput{Type: typ, Button: events.Left, Position: e.Bounds.Min.Add(e.Bounds.Size().Div(2))})
				}
			}
			settle()
			for _, id := range []string{"block", "connector", "note"} {
				_, source := element("tool-" + id)
				_, target := element("drawing")
				input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: source.Bounds.Min.Add(source.Bounds.Size().Div(2))})
				pos := target.Bounds.Min.Add(target.Bounds.Size().Div(2)).Add(image.Pt(37, -19))
				input(cogent.ManualInput{Type: events.MouseMove, Position: pos})
				if !m.palette.State().Accepted || len(m.scene.Objects) != m.journal.Snapshot().UndoCount {
					t.Fatal("preview changed history or rejected valid drop")
				}
				input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: pos})
			}
			if len(m.scene.Objects) != 3 || m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("routed placements did not commit once")
			}
			original := slices.Clone(m.scene.Objects)
			click("undo")
			if len(m.scene.Objects) != 2 {
				t.Fatal("routed Undo failed")
			}
			click("redo")
			if !slices.Equal(original, m.scene.Objects) {
				t.Fatal("routed Redo changed identity")
			}
			// Source pointer clicks never place. They focus the native button; Return
			// follows the same validation and command path at the current world center.
			click("tool-block")
			if len(m.scene.Objects) != 3 {
				t.Fatal("palette click placed")
			}
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter, Rune: '\n'})
			if len(m.scene.Objects) != 4 || m.journal.Snapshot().UndoCount != 4 {
				t.Fatal("focused keyboard placement failed")
			}
			fourth := m.scene.Objects[3]
			input(cogent.ManualInput{Type: events.KeyChord, Rune: 'z', Code: key.CodeZ, Modifiers: key.Modifiers(1 << key.Meta)})
			if len(m.scene.Objects) != 3 {
				t.Fatal("Cmd+Z did not undo one placement")
			}
			input(cogent.ManualInput{Type: events.KeyChord, Rune: 'Z', Code: key.CodeZ, Modifiers: key.Modifiers(1<<key.Meta | 1<<key.Shift)})
			if len(m.scene.Objects) != 4 || m.scene.Objects[3] != fourth {
				t.Fatal("Cmd+Shift+Z changed placement identity")
			}
			click("reject-notes")
			click("tool-note")
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter, Rune: '\n'})
			if len(m.scene.Objects) != 4 {
				t.Fatal("keyboard bypassed validation")
			}
			_, target := element("drawing")
			center := target.Bounds.Min.Add(target.Bounds.Size().Div(2)).Add(image.Pt(20, 20))
			// Fresh canvas input still selects, pans and zooms after captured releases.
			input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: center})
			input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: center})
			if m.scene.Selected == "" {
				t.Fatal("canvas selection failed")
			}
			inspection, err := driver.Snapshot(ctx)
			if err != nil || len(inspection.Windows[0].Canvases) != 1 || len(inspection.Windows[0].Canvases[0].Objects) == 0 {
				t.Fatal("public canvas semantics missing", err)
			}
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeHome, Modifiers: key.Modifiers(1 << key.Alt)})
			if m.scene.Selected != m.scene.Objects[0].ID {
				t.Fatal("public keyboard object selection failed")
			}
			before := m.scene.View
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeRightArrow})
			if m.scene.View.Center == before.Center {
				t.Fatal("canvas keyboard pan failed")
			}
			before = m.scene.View
			input(cogent.ManualInput{Type: events.Scroll, Position: center, Delta: math32.Vec2(0, -40)})
			if m.scene.View.Zoom <= before.Zoom {
				t.Fatal("canvas wheel zoom failed")
			}
			before = m.scene.View
			input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: center})
			if _, err := clock.Advance(core.TimingSettings.SlideStartTime + time.Millisecond); err != nil {
				t.Fatal(err)
			}
			input(cogent.ManualInput{Type: events.MouseMove, Button: events.Left, Position: center.Add(image.Pt(20, 10))})
			input(cogent.ManualInput{Type: events.MouseMove, Button: events.Left, Position: center.Add(image.Pt(40, 20))})
			input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: center.Add(image.Pt(40, 20))})
			if m.scene.View.Center == before.Center {
				t.Fatal("canvas pointer pan failed")
			}
			if _, err := driver.Capture(ctx, "placement"); err != nil {
				t.Fatal(err)
			}
			if m.palette.State().Active || a.Pending().Reserved != 0 {
				t.Fatal("consumer retained drag state", a.Pending())
			}
		})
	}
}
