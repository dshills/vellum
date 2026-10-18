//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

// Independent public Manual routing preflight; this is not native evidence.
func TestNavigationCataloguePublicManual(t *testing.T) {
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
	cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, LogicalDPI: 160, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 32, G: 36, B: 42, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}, Kinds: map[ui.Kind]ui.Style{ui.Column: {Gap: ui.Some(float32(8))}}}}
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
	g := newGallery()
	if e = a.Open(app.Window{ID: "gallery", Title: "Navigation", Width: 900, Height: 700, View: g.navigationCatalogueView}); e != nil {
		t.Fatal(e)
	}
	m, e := cogent.NewManual(ctx, a, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := m.Close(ctx); e != nil {
			t.Error("manual cleanup", e)
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
	send := func(code key.Codes, r rune) {
		t.Helper()
		if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: events.KeyChord, Code: code, Rune: r}); e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	setup := func(fn func()) {
		t.Helper()
		if e := a.Do(func() { fn(); _ = a.Invalidate("gallery") }); e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	click := func(el cogent.ManualElement) {
		t.Helper()
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: el.ID, TargetIncarnation: el.Incarnation, Type: typ, Button: events.Left}); e != nil {
				t.Fatal(e)
			}
			current = settle()
		}
	}
	focus := func(id string) {
		t.Helper()
		if current.FocusedID != id {
			t.Fatal("actual public focus", current.FocusedID, id)
		}
	}
	focus("layers")
	send(key.CodeDownArrow, 0)
	send(key.CodeDownArrow, 0)
	if g.listSelection != "layer-03" || g.listRequests != 2 {
		t.Fatal("list disabled skip", g.navigationModel())
	}
	send(key.CodeSpacebar, ' ')
	if g.listSelection != "layer-03" || g.listRequests != 3 {
		t.Fatal("explicit repeat proposal lost or duplicated", g.navigationModel())
	}
	send(key.CodeEnd, 0)
	send(key.CodeDownArrow, 0)
	if g.listSelection != "layer-23" || g.listRequests != 4 {
		t.Fatal("list end boundary", g.navigationModel())
	}
	send(key.CodeHome, 0)
	send(key.CodeA, 'a')
	if g.listSelection != "layer-23" || g.listRequests != 6 {
		t.Fatal("list initial letter", g.navigationModel())
	}
	setup(func() { g.listReject = true })
	send(key.CodeHome, 0)
	if g.listSelection != "layer-23" || g.listRequests != 7 {
		t.Fatal("list rejection", g.navigationModel())
	}
	setup(func() { g.listReject = false; g.listReadOnly = true })
	send(key.CodeEnd, 0)
	send(key.CodeHome, 0)
	send(key.CodeSpacebar, ' ')
	if g.listSelection != "layer-23" || g.listRequests != 7 {
		t.Fatal("readonly focus changed selection", g.navigationModel())
	}
	send(key.CodeTab, 0)
	focus("menu-actions")
	send(key.CodeDownArrow, 0)
	var second cogent.ManualElement
	for _, el := range current.Elements {
		if el.Parent == "menu-actions" && el.Name == "Run action" {
			second = el
		}
	}
	if second.ID == "" || second.Bounds.Empty() {
		t.Fatal("duplicate-label real popup target absent")
	}
	click(second)
	if g.menuCount != 1 || g.menuLast != "second" {
		t.Fatal("duplicate-label stable action", g.navigationModel())
	}
	send(key.CodeReturnEnter, 0)
	if e = m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: second.ID, TargetIncarnation: second.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(e, cogent.ErrManualStale) {
		t.Fatal("retired menu target admitted", e)
	}
	current = settle()
	send(key.CodeDownArrow, 0)
	send(key.CodeReturnEnter, 0)
	if g.menuCount != 2 || g.menuLast != "second" {
		t.Fatal("menu keyboard action", g.navigationModel())
	}
	send(key.CodeReturnEnter, 0)
	send(key.CodeEscape, 0)
	if g.menuCount != 2 {
		t.Fatal("Escape activated menu")
	}
	send(key.CodeReturnEnter, 0)
	setup(func() { g.menuDisabled = true })
	if find("menu-actions").Enabled {
		t.Fatal("disabled anchor publication")
	}
	for _, el := range current.Elements {
		if el.Parent == "menu-actions" && el.Kind == ui.MenuButton {
			t.Fatal("disabled menu retained live popup")
		}
	}
	setup(func() { g.menuDisabled = false })
	click(find("menu-actions"))
	send(key.CodeTab, 0)
	focus("view-menu")
	send(key.CodeDownArrow, 0)
	send(key.CodeReturnEnter, 0)
	if !g.menuChecked || g.viewMenuRequests != 1 {
		t.Fatal("check action acceptance")
	}
	send(key.CodeReturnEnter, 0)
	for _, el := range current.Elements {
		if el.Parent == "view-menu" && el.Name == "Show grid" && el.Value != "true" {
			t.Fatal("accepted check semantics", el.Value)
		}
	}
	setup(func() { g.viewMenuReject = true })
	send(key.CodeReturnEnter, 0)
	if !g.menuChecked || g.viewMenuRequests != 2 {
		t.Fatal("check rejection")
	}
	setup(func() { g.viewMenuReject = false })
	send(key.CodeDownArrow, 0)
	send(key.CodeDownArrow, 0)
	send(key.CodeReturnEnter, 0)
	if g.menuChecked || g.viewMenuRequests != 3 {
		t.Fatal("separator skipped/reset", g.navigationModel())
	}
}
