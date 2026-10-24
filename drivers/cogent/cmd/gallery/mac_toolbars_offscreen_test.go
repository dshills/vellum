//go:build offscreen

package main

import (
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
	"io/fs"
	"strings"
	"testing"
	"time"
)

func TestMacToolbarsPublicManualSizingRetentionAndOverflow(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) { testMacToolbarsPublicManual(t, dark) })
	}
}
func testMacToolbarsPublicManual(t *testing.T, dark bool) {
	ctx := context.Background()
	clock, e := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
	if e != nil {
		t.Fatal(e)
	}
	defer clock.Close()
	executor, e := execution.NewManual(8)
	if e != nil {
		t.Fatal(e)
	}
	defer executor.Close()
	cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, LogicalDPI: 160, Theme: ui.MacTheme(dark)}
	cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
	cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
	if e = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		b, e := fs.ReadFile(fonts.Default, path)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(b)
		cfg.Fonts = append(cfg.Fonts, cogent.ManualFont{Name: path, Data: b, SHA256: hex.EncodeToString(sum[:])})
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()

	model := newMacToolbarGallery(dark)
	focus := ui.FocusRequest{ID: "toolbar-Small/New", Sequence: 1}
	if e = a.Open(app.Window{ID: "gallery", Title: "Toolbars", Width: 660, Height: 780, View: func() ui.Node { n := model.view(); n.Focus = focus; return n }}); e != nil {
		t.Fatal(e)
	}
	m, e := cogent.NewManual(ctx, a, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := m.Close(ctx); e != nil {
			t.Error(e)
		}
	}()
	settle := func() cogent.ManualWindowSnapshot {
		t.Helper()
		for range 128 {
			snap, e := m.Snapshot(ctx)
			if e != nil {
				t.Fatal(e)
			}
			w := snap.Windows[0]
			if snap.Application.Ordinary+snap.Application.ReadyCompletions > 0 || w.Events > 0 || w.ReadyPaint {
				if _, e = m.Step(ctx, 7); e != nil {
					t.Fatal(e)
				}
				continue
			}
			if w.LayoutPending {
				if _, e = m.Frame(ctx, "gallery"); e != nil {
					t.Fatal(e)
				}
				continue
			}
			return w
		}
		t.Fatal("manual readiness bound")
		return cogent.ManualWindowSnapshot{}
	}
	current := settle()
	find := func(id string) cogent.ManualElement {
		t.Helper()
		for _, el := range current.Elements {
			if el.ID == id {
				return el
			}
		}
		t.Fatal("missing public target", id)
		return cogent.ManualElement{}
	}

	if vertical, viewport := find("toolbar-vertical").Bounds, find("toolbars").Bounds; vertical.Empty() || !vertical.In(viewport) {
		t.Fatal("vertical example clipped in standard window", vertical, viewport)
	}
	heights := []int{}
	for _, size := range macToolbarSizes {
		box := find("toolbar-" + size.name + "/New").Bounds
		if box.Empty() {
			t.Fatal("empty action", size.name)
		}
		heights = append(heights, box.Dy())
	}
	if heights[0] >= heights[1] || heights[1] >= heights[2] {
		t.Fatal("sizes do not differ", heights)
	}
	original := find("toolbar-Regular/Save").Incarnation
	bothWidth := find("toolbar-Regular/Save").Bounds.Dx()
	setup := func(fn func()) {
		t.Helper()
		if e := a.Do(func() { fn(); _ = a.Invalidate("gallery") }); e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	for _, fullColor := range []bool{false, true} {
		setup(func() { model.fullColor = fullColor })
		if find("toolbar-Regular/Save").Incarnation != original {
			t.Fatal("color mode replaced native action")
		}
	}
	for _, mode := range []string{"icons", "text", "both", "above"} {
		setup(func() { model.display = mode })
		if find("toolbar-Regular/Save").Incarnation != original {
			t.Fatal("presentation replaced action", mode)
		}
		if mode == "icons" && find("toolbar-Regular/Save").Bounds.Dx() >= bothWidth {
			t.Fatal("icon-only action did not become compact")
		}
	}
	setup(func() { model.narrow = true; focus.Sequence++ })
	send := func(code key.Codes) {
		t.Helper()
		if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: events.KeyChord, Code: code}); e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	send(key.CodeEnd)
	if current.FocusedID != "toolbar-Small/Redo" {
		t.Fatal("End did not skip disabled Delete", current.FocusedID)
	}
	if box := find(current.FocusedID).Bounds; box.Empty() || !box.In(find("toolbar-Small").Bounds) {
		t.Fatal("focused overflowing action not revealed", box, find("toolbar-Small").Bounds)
	}
	send(key.CodeReturnEnter)
	if model.count != 1 || model.last != "toolbar-Small: Redo" {
		t.Fatal("activation not exactly once", model.count, model.last)
	}
	send(key.CodeHome)
	if current.FocusedID != "toolbar-Small/New" {
		t.Fatal("Home failed", current.FocusedID)
	}
	setup(func() { model.disabled = true })
	send(key.CodeReturnEnter)
	if model.count != 1 {
		t.Fatal("disabled toolbar activated")
	}
	if _, e := m.Capture(ctx, "gallery"); e != nil {
		t.Fatal(e)
	}
}
