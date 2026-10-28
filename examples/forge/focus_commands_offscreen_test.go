//go:build offscreen

package main

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/matcolor"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"image"
	"io/fs"
	"strings"
	"testing"
	"time"
)

// Geometry/routing of the exact app descriptions. The GPU viewport alone is
// a fixed placeholder; this fixture does not claim GPU/native pointer evidence.
func TestFocusedBackspacePreservesBodies(t *testing.T) {
	// Serial fixture: no t.Parallel. All Manual windows close before restoration.
	// Process-local native palette only; never changes macOS appearance.
	previousDark := matcolor.SchemeIsDark
	defer colors.SetScheme(previousDark)

	var fixtureFonts []cogent.ManualFont
	if e := fs.WalkDir(fonts.Default, ".", func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() || !strings.HasSuffix(p, ".ttf") && !strings.HasSuffix(p, ".otf") {
			return nil
		}
		data, e := fs.ReadFile(fonts.Default, p)
		if e != nil {
			return e
		}
		h := sha256.Sum256(data)
		fixtureFonts = append(fixtureFonts, cogent.ManualFont{Name: p, Data: data, SHA256: hex.EncodeToString(h[:])})
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	for _, dpi := range []float32{160} {
		for _, size := range [][2]int{{1440, 920}} {
			t.Run(fmt.Sprintf("%g-%dx%d", dpi, size[0], size[1]), func(t *testing.T) {
				ctx := context.Background()
				clock, e := execution.NewManualClock(time.Unix(1730000000, 0), 256, 256)
				if e != nil {
					t.Fatal(e)
				}
				defer clock.Close()
				executor, e := execution.NewManual(8)
				if e != nil {
					t.Fatal(e)
				}
				defer executor.Close()
				theme := forgeTheme()
				theme.Default.Background = theme.Kinds[ui.FormGroup].Background
				theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				cfg := cogent.ManualConfig{Clock: clock, Executor: executor, Theme: theme, LogicalDPI: dpi, OwnerBatch: 7, MaxWindows: 2, EventCapacity: 128}
				cfg.Fonts = fixtureFonts
				a, e := app.New(app.Config{})
				if e != nil {
					t.Fatal(e)
				}
				defer a.Close()
				model := newWorkbench()
				cfg.Commands = model.commands
				var placeholder func(ui.Node) ui.Node
				placeholder = func(n ui.Node) ui.Node {
					if n.ID == "scene" {
						return ui.Node{ID: "scene", Kind: ui.Column, Grow: true, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 100, Height: 80})}}
					}
					for i := range n.Children {
						n.Children[i] = placeholder(n.Children[i])
					}
					return n
				}
				scale := dpi / 160
				ref, e := a.OpenWindow(app.Window{ID: "forge", Title: "Forge geometry", Width: float32(size[0]) * scale, Height: float32(size[1]) * scale, DisableRootScrolling: true, View: func() ui.Node {
					n := placeholder(model.View())
					n.Focus = ui.FocusRequest{ID: "width/quantity/input", Sequence: 1}
					return n
				}})
				if e != nil {
					t.Fatal(e)
				}
				model.application = a
				model.ref = ref
				m, e := cogent.NewManual(ctx, a, cfg)
				if e != nil {
					t.Fatal(e)
				}
				defer m.Close(ctx)
				colors.SetScheme(true) // Reproduce host palette after native initialization.
				settle := func() {
					t.Helper()
					for range 128 {
						s, e := m.Snapshot(ctx)
						if e != nil {
							t.Fatal(e)
						}
						ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
						for _, w := range s.Windows {
							ready = ready || w.Events > 0 || w.ReadyPaint
						}
						if ready {
							if _, e = m.Step(ctx, 7); e != nil {
								t.Fatal(e)
							}
							continue
						}
						pending := false
						for _, w := range s.Windows {
							if w.LayoutPending {
								if _, e = m.Frame(ctx, w.ID); e != nil {
									t.Fatal(e)
								}
								pending = true
							}
						}
						if !pending {
							return
						}
					}
					t.Fatal("settle budget")
				}
				settle()
				snapshot, e := m.Snapshot(ctx)
				if e != nil {
					t.Fatal(e)
				}

				if snapshot.Windows[0].FocusedID != "width/quantity/input" {
					t.Fatal("focus not established", snapshot.Windows[0].FocusedID)
				}
				if e := m.Input(ctx, cogent.ManualInput{Window: "forge", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.KeyChord, Rune: 'a', Code: key.CodeA, Modifiers: 1 << key.Control}); e != nil {
					t.Fatal(e)
				}
				settle()
				fieldText := func() string {
					snap, e := m.Snapshot(ctx)
					if e != nil {
						t.Fatal(e)
					}
					for _, el := range snap.Windows[0].Elements {
						if el.ID == "width/quantity/input" {
							v, _ := el.Value.(string)
							return v
						}
					}
					t.Fatal("missing field")
					return ""
				}
				originalText := fieldText()
				if originalText == "" {
					t.Fatal("empty fixture text")
				}
				count := len(model.state.Objects)
				draft := model.draft
				if e := m.Input(ctx, cogent.ManualInput{Window: "forge", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.KeyChord, Code: key.CodeBackspace}); e != nil {
					t.Fatal(e)
				}
				settle()
				t.Logf("Focused width field, Backspace: bodies %d -> %d; draft %q -> %q", count, len(model.state.Objects), draft, model.draft)
				if len(model.state.Objects) != count {
					t.Fatal("Backspace in width editor deleted a body")
				}

				if len(fieldText()) >= len(originalText) {
					t.Fatal("Backspace did not edit text", fieldText())
				}
				for _, mods := range []key.Modifiers{1 << key.Meta, 1<<key.Meta | 1<<key.Shift} {
					if e := m.Input(ctx, cogent.ManualInput{Window: "forge", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.KeyChord, Rune: 'z', Code: key.CodeZ, Modifiers: mods}); e != nil {
						t.Fatal(e)
					}
					settle()
					if len(model.state.Objects) != count {
						t.Fatal("text history changed bodies")
					}
				}
				elements := map[string]cogent.ManualElement{}
				for _, el := range snapshot.Windows[0].Elements {
					elements[el.ID] = el
				}
				bounds := image.Rect(0, 0, int(float32(size[0])*scale), int(float32(size[1])*scale))
				for _, id := range []string{"ribbon", "scene", "width/quantity/input", "width/quantity/apply", "nudge", "remove", "orientation-scroll", "timeline", "timeline/back", "status"} {
					el, ok := elements[id]
					if !ok || el.Bounds.Empty() || !el.Bounds.In(bounds) {
						t.Fatal("control clipped/missing", id, el.Bounds, bounds)
					}
				}
				if dpi == 160 && size[0] == 1440 {
					frame, e := m.Capture(ctx, "forge")
					if e != nil {
						t.Fatal(e)
					}
					bar := elements["ribbon/ribbon/c29saWQ"].Bounds
					if bar.Empty() {
						t.Fatal("missing ribbon page")
					}
					sample := frame.RGBAAt(bar.Max.X-20, bar.Min.Y+bar.Dy()/2)
					if sample.R < 230 || sample.G < 230 || sample.B < 230 {
						t.Fatal("ribbon inherited dark native surface", sample)
					}
					action := elements[ui.RibbonActionID("ribbon", "solid", "add")].Bounds
					ink := 0
					for y := action.Min.Y + action.Dy()/2; y < action.Max.Y; y++ {
						for x := action.Min.X; x < action.Max.X; x++ {
							p := frame.RGBAAt(x, y)
							if p.R < 90 && p.G < 100 && p.B < 110 {
								ink++
							}
						}
					}
					if ink < 3 {
						t.Fatal("ribbon caption lost dark readable ink", ink)
					}
				}
				input, apply := elements["width/quantity/input"].Bounds, elements["width/quantity/apply"].Bounds
				if input.Overlaps(apply) {
					t.Fatal("inspector controls overlap", input, apply)
				}
				if float32(elements["timeline"].Bounds.Dy()) > 48*scale {
					t.Fatal("timeline expanded")
				}
			})
		}
	}
}
