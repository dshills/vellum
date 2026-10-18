//go:build offscreen

package main

import (
	"cogentcore.org/core/math32"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"math"
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
func TestStructureCataloguePublicManual(t *testing.T) {
	for _, dpi := range []float32{160, 162} {
		t.Run(fmt.Sprintf("dpi-%g", dpi), func(t *testing.T) { testStructureCataloguePublicManual(t, dpi) })
	}
}
func testStructureCataloguePublicManual(t *testing.T, dpi float32) {
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
	cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, LogicalDPI: dpi, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 32, G: 36, B: 42, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}, Kinds: map[ui.Kind]ui.Style{ui.Column: {Gap: ui.Some(float32(8))}}}}
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
	if e = a.Open(app.Window{ID: "gallery", Title: "Structure", Width: 900, Height: 700, View: g.structureCatalogueView}); e != nil {
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
		for _, e := range current.Elements {
			if e.ID == id {
				return e
			}
		}
		t.Fatal("missing actual manual target", id)
		return cogent.ManualElement{}
	}
	checkSplit := func() {
		t.Helper()
		if g.structurePage != "split-page" {
			return
		}
		pane := find("panes")
		h := find(structureID("split", "panes"))
		wantX := float64(pane.Bounds.Min.X) + float64(g.split)*float64(pane.Bounds.Dx()-h.Bounds.Dx())
		if h.Bounds.Empty() || !h.Bounds.In(pane.Bounds) || math.Abs(float64(h.Bounds.Min.X)-wantX) > 4 {
			t.Fatal("actual split geometry", g.split, pane.Bounds, h.Bounds, wantX)
		}
	}
	focus := func(id string) {
		t.Helper()
		if current.FocusedID != id {
			t.Fatal("manual actual focus", current.FocusedID, id)
		}
	}
	setup := func(fn func()) {
		t.Helper()
		if e := a.Do(func() { fn(); _ = a.Invalidate("gallery") }); e != nil {
			t.Fatal(e)
		}
		current = settle()
		checkSplit()
	}
	send := func(code key.Codes, r rune) {
		t.Helper()
		if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: events.KeyChord, Code: code, Rune: r}); e != nil {
			t.Fatal(e)
		}
		current = settle()
		checkSplit()
	}
	click := func(id string) {
		t.Helper()
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			el := find(id)
			if el.Bounds.Empty() {
				t.Fatal("empty actual pointer bounds", id)
			}
			if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: el.ID, TargetIncarnation: el.Incarnation, Type: typ, Button: events.Left}); e != nil {
				t.Fatal(e)
			}
			current = settle()
			checkSplit()
		}
	}
	requestFocus := func(id string) { setup(func() { g.focus.ID = id; g.focus.Sequence += 2 }) }
	heading, handle := structureID("disclosure", "details"), structureID("split", "panes")
	draft := find("section-draft")
	afterY := find("section-after").Bounds.Min.Y
	normalHeights := []int{find(heading).Bounds.Dy(), find("section-draft").Bounds.Dy(), find("section-action").Bounds.Dy()}
	focus("details")
	send(key.CodeLeftArrow, 0)
	if g.details || g.detailsRequests != 1 || find("section-after").Bounds.Min.Y >= afterY {
		t.Fatal("collapse state/layout", g.structureModel())
	}
	send(key.CodeTab, 0)
	focus("section-after")
	click(heading)
	if !g.details || g.detailsRequests != 2 || find("section-draft").Incarnation != draft.Incarnation {
		t.Fatal("reopen lost child identity", g.structureModel())
	}
	send(key.CodeTab, 0)
	focus("section-draft")
	click("section-action")
	if g.outsideClicks != 1 {
		t.Fatal("section action duplicated")
	}
	setup(func() { g.details = false })
	focus("details")
	send(key.CodeRightArrow, 0)
	send(key.CodeReturnEnter, 0)
	send(key.CodeSpacebar, ' ')
	setup(func() { g.detailsReject = true })
	send(key.CodeLeftArrow, 0)
	if !g.details || g.detailsRequests != 6 {
		t.Fatal("disclosure rejection", g.structureModel())
	}
	setup(func() { g.detailsReject = false; g.detailsReadOnly = true })
	send(key.CodeLeftArrow, 0)
	if !g.details || g.detailsRequests != 6 {
		t.Fatal("readonly disclosure mutated", g.structureModel())
	}
	setup(func() { g.detailsReadOnly = false; g.detailsDisabled = true })
	// Disabled Core buttons drop shadow margins; the inherited-disabled field
	// drops its reserved border. These are distinct style states, not a restore.
	disabledHeights := []int{find(heading).Bounds.Dy(), find("section-draft").Bounds.Dy(), find("section-action").Bounds.Dy()}
	for i, shrink := range []int{6, 2, 6} {
		if normalHeights[i]-disabledHeights[i] != shrink {
			t.Fatal("disabled style metric", dpi, i, normalHeights, disabledHeights)
		}
	}
	if find("section-after").Bounds.Min.Y != afterY-14 {
		t.Fatal("disabled section layout", dpi, afterY, find("section-after").Bounds)
	}
	if find(heading).Enabled || find("section-draft").Enabled {
		t.Fatal("inherited disclosure disabled state")
	}
	setup(func() { g.detailsDisabled = false })
	if find("section-after").Bounds.Min.Y != afterY {
		t.Fatal("same-style enabled restoration", dpi, afterY, find("section-after").Bounds)
	}
	click(heading)
	requestFocus("structure-pages")
	send(key.CodeRightArrow, 0)
	if g.structurePage != "split-page" || g.structurePageRequests != 1 {
		t.Fatal("tab disabled skip", g.structureModel())
	}
	click(handle)
	focus("panes")
	send(key.CodeRightArrow, 0)
	send(key.CodeLeftArrow, 0)
	send(key.CodeHome, 0)
	if math.Abs(float64(g.split-0.05)) > 1e-6 || g.splitRequests != 3 {
		t.Fatal("split keyboard", g.structureModel())
	}
	send(key.CodeEnd, 0)
	setup(func() { g.splitReject = true })
	send(key.CodeHome, 0)
	if g.split != 0.95 || g.splitRequests != 5 {
		t.Fatal("split rejection", g.structureModel())
	}
	setup(func() { g.splitReject = false; g.splitReadOnly = true })
	send(key.CodeHome, 0)
	if g.split != 0.95 || g.splitRequests != 5 {
		t.Fatal("readonly split mutated")
	}
	setup(func() { g.splitReadOnly = false })
	send(key.CodeHome, 0)
	click("split-reset")
	if g.split != 0.35 || g.splitRequests != 6 {
		t.Fatal("ordinary split reset", g.structureModel())
	}
	top := find("split-line-00").Bounds
	for i, delta := range []float32{100000, 100000, -100000} {
		el := find("split-workspace")
		if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: el.ID, TargetIncarnation: el.Incarnation, Type: events.Scroll, Delta: math32.Vec2(0, delta)}); e != nil {
			t.Fatal(e)
		}
		current = settle()
		checkSplit()
		first, last := find("split-line-00"), find("split-line-23")
		if i < 2 {
			if !first.Bounds.Empty() || last.Bounds.Empty() {
				t.Fatal("scroll bottom/boundary did not project endpoints", i, first, last)
			}
		} else if first.Bounds != top || !last.Bounds.Empty() {
			t.Fatal("scroll top not restored", first, last)
		}
	}
	requestFocus("structure-pages")
	send(key.CodeEnd, 0)
	if g.structurePage != "documents" {
		t.Fatal("document page not selected")
	}
	requestFocus("document-tabs")
	send(key.CodeEnd, 0)
	if g.documentSelected != "notes" || g.documentSelectRequests != 1 {
		t.Fatal("document selection", g.structureModel())
	}
	close := find(structureID("tab-close", "document-tabs", "notes"))
	setup(func() { g.documentReject = true })
	click(close.ID)
	send(key.CodeDelete, 0)
	if g.documentCloseRequests != 2 || len(g.documents) != 2 || g.documentSelectRequests != 1 {
		t.Fatal("rejected close changed selection/ownership", g.structureModel())
	}
	setup(func() { g.documentReject = false })
	send(key.CodeDelete, 0)
	if g.documentCloseRequests != 3 || len(g.documents) != 1 || g.documentSelected != "sketch" || g.documentSelectRequests != 1 {
		t.Fatal("accepted close duplicated or selected", g.structureModel())
	}
	if e := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: close.ID, TargetIncarnation: close.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(e, cogent.ErrManualStale) {
		t.Fatal("retired close admitted", e)
	}
	current = settle()
	checkSplit()
	requestFocus("structure-pages")
	send(key.CodeHome, 0)
	click(heading)
	if !g.details || g.detailsRequests != 8 || find("section-draft").Incarnation != draft.Incarnation || find("section-draft").Value != "Vellum" {
		t.Fatal("retained section after tab round trip", g.structureModel())
	}
}
