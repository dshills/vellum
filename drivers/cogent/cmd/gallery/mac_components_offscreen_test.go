//go:build offscreen

package main

import (
	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This fixture renders without opening a desktop window. It checks allocations
// and complete rasters; it does not establish physical-pointer/native evidence.
func TestMacComponentsOffscreenSurfacesAndJoinedSegments(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
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
			cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, LogicalDPI: 160, Theme: ui.MacTheme(dark)}
			cfg.Theme.Default.Background = ui.Some(cfg.Theme.Palette.Value.Window)
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			if err := fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
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
				sum := sha256.Sum256(data)
				cfg.Fonts = append(cfg.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(sum[:])})
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			model := newMacComponentsGallery(dark)
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if err := a.Open(app.Window{ID: "components", Title: "Offscreen components", Width: 660, Height: 740, View: func() ui.Node {
				n := model.view()
				// Manual does not emulate external OS file transport. Project away only
				// that annotation; production FileDrop routing has separate fixtures.
				var project func(*ui.Node)
				project = func(node *ui.Node) {
					if node.ID == ui.ImageWellTargetID("essentials/well") {
						node.FileDrop = nil
					}
					for i := range node.Children {
						project(&node.Children[i])
					}
				}
				project(&n)
				return n
			}}); err != nil {
				t.Fatal(err)
			}
			m, err := cogent.NewManual(ctx, a, cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := m.Close(ctx); err != nil {
					t.Error(err)
				}
			}()
			settle := func() cogent.ManualWindowSnapshot {
				t.Helper()
				for range 128 {
					snapshot, err := m.Snapshot(ctx)
					if err != nil {
						t.Fatal(err)
					}
					w := snapshot.Windows[0]
					if snapshot.Application.Ordinary+snapshot.Application.ReadyCompletions > 0 || w.Events > 0 || w.ReadyPaint {
						if _, err = m.Step(ctx, 7); err != nil {
							t.Fatal(err)
						}
						continue
					}
					if w.LayoutPending {
						if _, err = m.Frame(ctx, "components"); err != nil {
							t.Fatal(err)
						}
						continue
					}
					return w
				}
				t.Fatal("offscreen readiness bound")
				return cogent.ManualWindowSnapshot{}
			}
			for _, page := range []string{"theme", "menus", "split", "inspector", "browser", "rules", "image-well", "icon-picker", "icon-picker-open", "settings", "alerts", "calendar", "documents", "outline", "native-colors", "native-fonts"} {
				if err := a.Do(func() {
					model.page = page
					if page == "icon-picker-open" {
						model.page = "icon-picker"
						model.essentials.iconPickerView()
						model.essentials.picker.Open = true
					}
					if page == "calendar" {
						model.essentials.calendarView()
						date := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
						model.essentials.calendar.Value = date
						model.essentials.calendar.Month = date
						model.essentials.calendar.Today = date
						model.essentials.datePicker.Value = date
						model.essentials.datePicker.Today = date
					}
					if page == "browser" {
						model.browserPreview()
						model.browser.Path = []string{"drawings", "floor"}
					}
					_ = a.Invalidate("components")
				}); err != nil {
					t.Fatal(err)
				}
				w := settle()

				find := func(id string) cogent.ManualElement {
					t.Helper()
					for _, e := range w.Elements {
						if e.ID == id {
							return e
						}
					}
					t.Fatal("missing element", id)
					return cogent.ManualElement{}
				}
				for _, element := range w.Elements {
					if strings.HasSuffix(element.ID, "/error") || strings.HasSuffix(element.ID, "-error") {
						t.Fatal("gallery construction error", page, element.ID)
					}
				}
				if page == "image-well" && find(ui.ImageWellTargetID("essentials/well")).Bounds.Empty() {
					t.Fatal("image well has no geometry")
				}
				if page == "settings" && find("essentials/settings/name").Bounds.Empty() {
					t.Fatal("settings editor has no geometry")
				}
				if page == "calendar" && find(ui.CalendarDayID("essentials/calendar", time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC))).Bounds.Empty() {
					t.Fatal("calendar date has no geometry")
				}
				if page == "calendar" {
					previous := find("essentials/calendar/calendar/previous")
					next := find("essentials/calendar/calendar/next")
					if previous.Bounds.Empty() || next.Bounds.Empty() || next.Bounds.Min.X-previous.Bounds.Min.X > 270 {
						t.Fatal("calendar navigation clipped or stretched", previous.Bounds, next.Bounds)
					}
					for col := 0; col < 7; col++ {
						heading := find(fmt.Sprintf("essentials/calendar/calendar/weekday/%d", col)).Bounds
						for _, day := range []int{2 + col, 9 + col, 23 + col} {
							cell := find(ui.CalendarDayID("essentials/calendar", time.Date(2026, 3, day, 12, 0, 0, 0, time.UTC))).Bounds
							difference := heading.Min.X + heading.Max.X - cell.Min.X - cell.Max.X
							if cell.Empty() || difference < -2 || difference > 2 {
								t.Fatal("calendar column shifted", col, heading, cell)
							}
						}
					}
				}

				if page == "split" {
					for i := 0; i < 2; i++ {
						id := fmt.Sprintf("components/split/%d", i)
						primary := find(ui.SplitPrimaryID(id))
						menu := find(ui.SplitMenuID(id))
						if primary.Bounds.Empty() || menu.Bounds.Empty() {
							t.Fatal("split has no actual hit geometry")
						}
						gap := menu.Bounds.Min.X - primary.Bounds.Max.X
						if gap < -1 || gap > 2 {
							t.Fatal("split segments not joined", gap, primary.Bounds, menu.Bounds)
						}
					}
				}
				if page == "inspector" && find("components/inspector-name").Bounds.Empty() {
					t.Fatal("inspector editor clipped")
				}
				if page == "browser" && find("components/browser").Bounds.Dx() < 500 {
					t.Fatal("browser viewport failed to fill gallery")
				}
				if page == "rules" && find("components/rules").Bounds.Dx() < 500 {
					t.Fatal("rule viewport failed to fill gallery")
				}
				if page == "rules" && find(ui.RuleControlID("components/rules", "name-rule", "value")).Bounds.Empty() {
					t.Fatal("rule editor clipped")
				}
				capture, err := m.Capture(ctx, "components")
				if err != nil || capture.Bounds().Empty() {
					t.Fatal("no completed raster", page, err)
				}
				if dir := os.Getenv("VELLUM_COMPONENT_CAPTURE_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s-%t.png", page, dark)))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(file, capture)
					closeErr := file.Close()
					if err != nil || closeErr != nil {
						t.Fatal(err, closeErr)
					}
				}
			}
		})
	}
}
