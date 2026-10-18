//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

// Independent public Manual routing preflight; this is not native evidence.
func TestTextCataloguePublicManual(t *testing.T) {
	for _, dpi := range []float32{160, 162} {
		t.Run(fmt.Sprintf("dpi-%g", dpi), func(t *testing.T) { testTextCataloguePublicManual(t, dpi) })
	}
}
func testTextCataloguePublicManual(t *testing.T, dpi float32) {
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
	if e = a.Open(app.Window{ID: "gallery", Title: "Text", Width: 900, Height: 700, View: g.textCatalogueView}); e != nil {
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
		t.Fatal("actual target absent", id)
		return cogent.ManualElement{}
	}

	var background, retired, stale cogent.ManualInput
	selectAll := catalogueSelectAllModifiers()
	if keymap.Of(key.NewChord('a', key.CodeA, selectAll)) != keymap.SelectAll {
		selectAll = 0
		selectAll.SetFlag(true, key.Control)
	}
	if keymap.Of(key.NewChord('a', key.CodeA, selectAll)) != keymap.SelectAll {
		t.Fatal("no tested Select All binding")
	}
	send := func(in cogent.ManualInput, refusal error) {
		t.Helper()
		e := m.Input(ctx, in)
		if refusal != nil {
			if !errors.Is(e, refusal) {
				t.Fatal("expected refusal", in, e)
			}
		} else if e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	click := func(label, id string) {
		t.Helper()
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			el := find(id)
			if !el.Visible || el.Bounds.Empty() {
				t.Fatal(label, "pointer target hidden", el)
			}
			in := cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: el.ID, TargetIncarnation: el.Incarnation, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}
			if id == "text-preferences" && typ == events.MouseDown {
				background = in
			}
			if id == "dialog-name" && typ == events.MouseDown {
				retired = in
			}
			if id == "apply" && typ == events.MouseDown {
				stale = in
			}
			send(in, nil)
		}
	}
	sendKey := func(label, id string, code key.Codes, r rune, mods key.Modifiers) {
		t.Helper()
		el := find(id)
		if current.FocusedID != id {
			t.Fatal(label, "actual key focus", current.FocusedID, "want", id)
		}
		send(cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: el.ID, TargetIncarnation: el.Incarnation, Type: events.KeyChord, Code: code, Rune: r, Modifiers: mods}, nil)
	}
	for _, step := range g.textScript() {
		if step.Setup != nil {
			if e = a.Do(func() { step.Setup(); _ = a.Invalidate("gallery") }); e != nil {
				t.Fatal(e)
			}
			current = settle()
		}
		switch step.Kind {
		case "observe":
		case "click":
			click(step.Label, step.Target)
		case "key":
			mods := step.Modifiers
			if step.Code == key.CodeA && mods != 0 {
				mods = selectAll
			}
			sendKey(step.Label, step.Target, step.Code, step.Rune, mods)
		case "replace":
			click(step.Label, step.Target)
			sendKey(step.Label+"/select-all", step.Target, key.CodeA, 'a', selectAll)
			for _, r := range step.Text {
				sendKey(step.Label, step.Target, 0, r, 0)
			}
			sendKey(step.Label+"/commit", step.Target, key.CodeReturnEnter, 0, 0)
		case "refuse", "background", "retired", "stale":
			in := background
			if step.Kind == "refuse" {
				el := find(step.Target)
				in = cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Target: el.ID, TargetIncarnation: el.Incarnation, Type: events.MouseDown, Button: events.Left}
			}
			if step.Kind == "retired" {
				in = retired
			}
			if step.Kind == "stale" {
				in = stale
			}
			if in.Target == "" {
				t.Fatal("missing refusal capability")
			}
			send(in, cogent.ErrManualDisabled)
		default:
			t.Fatal("unknown authored step", step.Kind)
		}
		if e = a.Do(func() {
			if actual := g.textModel(); actual != step.Want {
				t.Errorf("%s model %+v want %+v", step.Label, actual, step.Want)
			}
		}); e != nil {
			t.Fatal(e)
		}
		if t.Failed() {
			return
		}
		if step.Focus != "" && current.FocusedID != step.Focus {
			t.Fatal(step.Label, "actual focus", current.FocusedID, "want", step.Focus)
		}
		if step.Want.Modal {
			for _, id := range []string{"preferences", "dialog-name", "apply"} {
				el := find(id)
				if !el.Visible || el.Bounds.Empty() {
					t.Fatal(step.Label, "modal visible geometry", id, el)
				}
			}
			if find("dialog-name").Value != step.Want.Name {
				t.Fatal("modal accepted text publication")
			}
		} else {
			for id, want := range map[string]string{"text-name": step.Want.Name, "text-notes": step.Want.Note, "text-status": fmt.Sprintf("Project %s / edits %d / undo %d / redo %d", step.Want.Name, step.Want.Edits, step.Want.Undo, step.Want.Redo)} {
				el := find(id)
				if !el.Visible || el.Bounds.Empty() || el.Value != want {
					t.Fatal(step.Label, "accepted publication", id, el, want)
				}
			}
			if find("text-name").Enabled == step.Want.DisableName || find("text-notes").Enabled == step.Want.DisableEditor {
				t.Fatal("disabled text publication")
			}
			if find("dialog-name").Visible || find("apply").Visible {
				t.Fatal("retired modal children still visible")
			}
		}
	}
}
