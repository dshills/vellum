//go:build offscreen

package replayconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/color"
	"io/fs"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

// This independent module uses only supported framework APIs. Its two panels
// intentionally repeat local IDs; actual routed input must reach the right one.
func TestPublicWorkbenchManualConsumer(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
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
			a, err := app.New(app.Config{QueueCapacity: 16})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "documents", Tabs: []string{"first", "second"}, Active: "first"}}, []string{"first", "second"})
			if err != nil {
				t.Fatal(err)
			}
			dock := &cogent.Workbench{Owner: a, Window: "main", Workspace: model}
			activated := map[string]int{}
			for _, panel := range []string{"first", "second"} {
				panel := panel
				dock.Panels = append(dock.Panels, cogent.Panel{ID: panel, Title: panel, View: func() ui.Node {
					return ui.Node{ID: "page", Kind: ui.Column, Children: []ui.Node{{ID: "action", Kind: ui.Button, Text: panel, Activate: func() { activated[panel]++; a.Invalidate("main") }}}}
				}})
			}
			if err = a.Open(app.Window{ID: "main", Title: "Public dock consumer", Width: 800, Height: 560, View: func() ui.Node { return ui.Node{ID: "workspace", Kind: ui.Extension, Custom: dock, Grow: true} }}); err != nil {
				t.Fatal(err)
			}
			config := cogent.ManualConfig{Clock: clock, Executor: executor, LogicalDPI: 160, OwnerBatch: batch, MaxWindows: 2, EventCapacity: 128, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
			err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
					return nil
				}
				data, err := fs.ReadFile(fonts.Default, path)
				if err != nil {
					return err
				}
				digest := sha256.Sum256(data)
				config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(digest[:])})
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			m, err := cogent.NewManual(nativeContext, a, config)
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close(nativeContext)
			settle := func() {
				for range 4096 {
					s, err := m.Snapshot(nativeContext)
					if err != nil {
						t.Fatal(err)
					}
					ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
					for _, w := range s.Windows {
						ready = ready || w.Events > 0 || w.ReadyPaint
						if !ready && w.LayoutPending {
							if err = m.RequestFrame(nativeContext, w.ID); err != nil {
								t.Fatal(err)
							}
							ready = true
						}
					}
					if !ready {
						return
					}
					work, err := m.Step(nativeContext, batch)
					if err != nil || work.Events+work.OwnerCallbacks > batch {
						t.Fatal("dispatch", work, err)
					}
				}
				t.Fatal("consumer exhausted explicit dispatch bound")
			}
			click := func(id string) {
				settle()
				s, err := m.Snapshot(nativeContext)
				if err != nil {
					t.Fatal(err)
				}
				w := s.Windows[0]
				for _, e := range w.Elements {
					if e.ID != id {
						continue
					}
					if !e.Visible || !e.Enabled || e.Bounds.Empty() {
						t.Fatalf("target unavailable: %+v", e)
					}
					for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
						if err = m.Input(nativeContext, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: kind, Button: events.Left}); err != nil {
							t.Fatal(err)
						}
					}
					settle()
					return
				}
				t.Fatal("missing public target", id)
			}
			click(cogent.ManualWorkbenchTabID("workspace", "second"))
			click(cogent.ManualWorkbenchControlID("workspace", "second", "action"))
			if model.Layout().Root.Active != "second" || activated["first"] != 0 || activated["second"] != 1 {
				t.Fatal("public routing or panel identity failed", activated)
			}
			if img, err := m.Capture(nativeContext, "main"); err != nil || img.Bounds().Dx() != 800 || img.Bounds().Dy() != 560 {
				t.Fatal("missing completed software composition", err)
			}
		})
	}
}
