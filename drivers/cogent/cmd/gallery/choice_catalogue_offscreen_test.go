//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"io/fs"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

// An independently consumed public Manual adapter preflights the composed
// controls. It is software routing evidence, not native acceptance.
func TestChoiceCataloguePublicManualKeyboardAndDisableRoundTrip(t *testing.T) {
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
	cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, LogicalDPI: 160, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 32, G: 36, B: 42, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}, Kinds: map[ui.Kind]ui.Style{ui.Column: {Gap: ui.Some(float32(8))}}}}
	if err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		data, e := fs.ReadFile(fonts.Default, path)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(data)
		cfg.Fonts = append(cfg.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(sum[:])})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	g := newGallery()
	if err = a.Open(app.Window{ID: "gallery", Title: "Choices", Width: 900, Height: 700, View: g.choiceCatalogueView}); err != nil {
		t.Fatal(err)
	}
	m, err := cogent.NewManual(ctx, a, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(ctx)
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
		t.Fatal("manual readiness budget")
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
		t.Fatal("missing public control", id)
		return cogent.ManualElement{}
	}
	for _, id := range []string{"enable", "snapping", "pinned", "radio-preset", "segmented-preset", "slider", "progress", "progress-ring", "progress-gauge", "choice-status", "outside", "choice-icon", "choice-title"} {
		el := find(id)
		if !el.Visible || el.Bounds.Empty() || !el.Bounds.In(image.Rect(0, 0, 900, 700)) {
			t.Fatal("compact fixture clipped", id, el.Bounds)
		}
	}
	send := func(code key.Codes) {
		t.Helper()
		in := cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: events.KeyChord, Code: code}
		if code == key.CodeSpacebar {
			in.Rune = ' '
		}
		if e := m.Input(ctx, in); e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	expectFocus := func(id string) {
		t.Helper()
		if current.FocusedID != id {
			t.Fatal("actual public focus", current.FocusedID, id)
		}
	}
	expectFocus("enable")
	send(key.CodeTab)
	expectFocus("snapping")
	send(key.CodeSpacebar)
	send(key.CodeTab)
	expectFocus("pinned")
	send(key.CodeSpacebar)
	send(key.CodeTab)
	expectFocus("radio-preset")
	send(key.CodeRightArrow)
	send(key.CodeRightArrow)
	if g.selection != "basic" || g.selectionRequests != 2 {
		t.Fatal("radio wrap/disabled skip")
	}
	if err = a.Do(func() { g.selectionReject = true; _ = a.Invalidate("gallery") }); err != nil {
		t.Fatal(err)
	}
	current = settle()
	send(key.CodeRightArrow)
	if g.selection != "basic" || g.selectionRequests != 3 {
		t.Fatal("rejected shared choice")
	}
	if err = a.Do(func() { g.selectionReject = false; _ = a.Invalidate("gallery") }); err != nil {
		t.Fatal(err)
	}
	current = settle()
	send(key.CodeTab)
	expectFocus("segmented-preset")
	send(key.CodeRightArrow)
	send(key.CodeHome)
	send(key.CodeSpacebar)
	if g.selection != "basic" || g.selectionRequests != 5 {
		t.Fatal("segments selection/no redundant proposal", g.selection, g.selectionRequests)
	}
	send(key.CodeTab)
	expectFocus("slider")
	send(key.CodeRightArrow)
	if err = a.Do(func() { g.numberReject = true; _ = a.Invalidate("gallery") }); err != nil {
		t.Fatal(err)
	}
	current = settle()
	send(key.CodeRightArrow)
	if g.number != 26 || g.numberRequests != 2 {
		t.Fatal("slider owner rejection", g.number, g.numberRequests)
	}
	if err = a.Do(func() { g.numberReject = false; _ = a.Invalidate("gallery") }); err != nil {
		t.Fatal(err)
	}
	current = settle()
	send(key.CodeHome)
	send(key.CodeEnd)
	send(key.CodeRightArrow)
	if g.number != 100 || g.numberRequests != 5 {
		t.Fatal("bounded slider proposes once per native Change", g.number, g.numberRequests)
	}
	for _, id := range []string{"progress", "progress-ring", "progress-gauge"} {
		if find(id).Value != float32(1) {
			t.Fatal("progress controlled fraction", id, find(id).Value)
		}
	}
	send(key.CodeTab)
	expectFocus("outside")
	send(key.CodeSpacebar)
	if g.outsideClicks != 1 || g.snappingRequests != 1 || g.pinnedRequests != 1 {
		t.Fatal("ordinary action/toggle counts", g.outsideClicks, g.snappingRequests, g.pinnedRequests, g.snapping, g.pinned)
	}
	clickEnable := func() {
		t.Helper()
		el := find("enable")
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}); e != nil {
				t.Fatal(e)
			}
			current = settle()
		}
	}
	clickEnable()
	if g.enabled || find("slider").Enabled || find("radio-preset").Enabled {
		t.Fatal("ancestor disability not published")
	}
	send(key.CodeUpArrow)
	if g.numberRequests != 5 || g.selectionRequests != 5 {
		t.Fatal("disabled ordinary key mutated model")
	}
	clickEnable()
	if !g.enabled || g.enabledRequests != 2 {
		t.Fatal("checkbox re-enable")
	}
	for range 5 {
		send(key.CodeTab)
	}
	expectFocus("slider")
	send(key.CodeHome)
	if g.number != 0 || g.numberRequests != 6 {
		t.Fatal("reenabled slider not editable")
	}
}
