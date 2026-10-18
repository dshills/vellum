//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
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
func TestMediaCataloguePublicManual(t *testing.T) {
	for _, dpi := range []float32{160, 162} {
		t.Run(fmt.Sprintf("dpi-%g", dpi), func(t *testing.T) { testMediaCataloguePublicManual(t, dpi) })
	}
}
func testMediaCataloguePublicManual(t *testing.T, dpi float32) {
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
	if e = a.Open(app.Window{ID: "gallery", Title: "Media", Width: 900, Height: 700, View: g.mediaCatalogueView}); e != nil {
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
	var stalePath, staleColor cogent.ManualInput
	var initialImage image.Rectangle
	var originalSample color.NRGBA
	originalSource := g.picture
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
				t.Fatal("target not refused", in, e)
			}
		} else if e != nil {
			t.Fatal(e)
		}
		current = settle()
	}
	for _, step := range g.mediaScript() {
		if step.Setup != nil {
			if e := a.Do(func() { step.Setup(); _ = a.Invalidate("gallery") }); e != nil {
				t.Fatal(e)
			}
			current = settle()
		}
		types := []events.Types{events.KeyChord}
		if step.Kind == "observe" {
			types = nil
		}
		if step.Kind == "click" || step.Kind == "text" {
			types = []events.Types{events.MouseDown, events.MouseUp}
		}
		if step.Kind == "refuse" {
			types = []events.Types{events.MouseDown}
		}
		// Manual exposes the managed Tabs/SegmentedControl, not native option nodes.
		// Exercise their independent public keyboard route; native uses pointer options.
		code := step.Code
		if step.Option != "" {
			if e := a.Do(func() { g.focus.ID = step.Target; g.focus.Sequence += 2; _ = a.Invalidate("gallery") }); e != nil {
				t.Fatal(e)
			}
			current = settle()
			types = []events.Types{events.KeyChord}
			code = key.CodeRightArrow
			if step.Option == "chrome" {
				code = key.CodeHome
			}
		}

		for _, typ := range types {
			in := cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: typ, Code: code}
			reject := step.Kind == "refuse" || step.Kind == "stale-path" || step.Kind == "stale-color"
			if step.Kind == "stale-path" {
				in = stalePath
				in.Type = events.MouseDown
			} else if step.Kind == "stale-color" {
				in = staleColor
				in.Type = events.MouseDown
			} else {
				id := step.Target

				el := find(id)
				in.Target = el.ID
				in.TargetIncarnation = el.Incarnation
				if typ == events.KeyChord {
					if current.FocusedID != step.Target {
						t.Fatalf("%s actual key focus %s want %s", step.Label, current.FocusedID, step.Target)
					}
				} else {
					in.Button = events.Left
				}
				if step.Target == ui.BreadcrumbItemID("project-path", "workspace") && typ == events.MouseDown && !reject {
					stalePath = in
				}
				if step.Target == ui.ColorFieldID("ink", ui.ColorHex) && typ == events.MouseDown {
					staleColor = in
				}
			}
			refusal := error(nil)
			if reject {
				refusal = cogent.ErrManualStale
				if step.Kind == "refuse" || step.Kind == "stale-color" {
					refusal = cogent.ErrManualDisabled
				}
			}
			send(in, refusal)
		}
		if step.Kind == "text" {
			for i, r := range append(append([]rune{'a'}, []rune(step.Text)...), '\n') {
				if current.FocusedID != step.Target {
					t.Fatalf("%s text focus %s", step.Label, current.FocusedID)
				}
				in := cogent.ManualInput{Window: "gallery", WindowIncarnation: current.Incarnation, Type: events.KeyChord, Rune: r}
				if i == 0 {
					in.Code = key.CodeA
					in.Modifiers = selectAll
				} else if r == '\n' {
					in.Rune = 0
					in.Code = key.CodeReturnEnter
				}
				send(in, nil)
			}
		}
		if step.Label == "hidden-color-refusal" {
			el := find(ui.ColorFieldID("ink", ui.ColorHex))
			if el.Visible || !el.Enabled {
				t.Fatal("hidden retained hex still visible", el)
			}
		}
		if g.mediaModel() != step.Want {
			t.Fatalf("%s actual model %+v want %+v", step.Label, g.mediaModel(), step.Want)
		}
		if step.Focus != "" && current.FocusedID != step.Focus {
			t.Fatalf("%s focus %s want %s", step.Label, current.FocusedID, step.Focus)
		}
		if step.Focus != "" {
			e := find(step.Focus)
			if e.Bounds.Empty() || !e.Bounds.In(find("media-catalogue").Bounds) {
				t.Fatal("focused control not visible", step.Label, e)
			}
		}
		if strings.HasPrefix(step.Focus, "media-tool-") && !find(step.Focus).Bounds.In(find("media-tools").Bounds) {
			t.Fatal("focused tool outside toolbar viewport", step.Label)
		}
		if step.Label == "initial" {
			last := find("media-tool-7")
			if !last.Bounds.Empty() && last.Bounds.In(find("media-tools").Bounds) {
				t.Fatal("fixture does not overflow")
			}
		}
		if step.Label == "image-contain" {
			im, e := m.Capture(ctx, "gallery")
			if e != nil {
				t.Fatal(e)
			}
			initialImage = find("image-preview").Bounds
			sample := image.Pt(initialImage.Min.X+initialImage.Dx()/6, initialImage.Min.Y+initialImage.Dy()/2)
			originalSample = color.NRGBAModel.Convert(im.At(sample.X, sample.Y)).(color.NRGBA)
			if originalSample != (color.NRGBA{R: 240, G: 64, B: 48, A: 255}) {
				t.Fatal("independent original strip pixel", sample, initialImage, originalSample)
			}
		}
		if step.Label == "image-replace" {
			im, e := m.Capture(ctx, "gallery")
			if e != nil {
				t.Fatal(e)
			}
			box := find("image-preview").Bounds
			if box != initialImage {
				t.Fatal("replacement changed image geometry")
			}
			sample := image.Pt(box.Min.X+box.Dx()/6, box.Min.Y+box.Dy()/2)
			actual := color.NRGBAModel.Convert(im.At(sample.X, sample.Y)).(color.NRGBA)
			if actual != (color.NRGBA{R: 240, G: 192, A: 255}) || actual == originalSample || g.picture == originalSource || originalSource.NRGBAAt(0, 0) != (color.NRGBA{R: 240, G: 64, B: 48, A: 255}) {
				t.Fatal("replacement pixels/immutable prior source", actual)
			}
		}

	}
	if g.mediaTools != 1 || g.breadcrumbRequests != 3 || g.imageReplacements != 1 || g.imageFitRequests != 3 || g.colorEdits != 6 || g.colorInvalid != 1 || g.colorPicker.Color != (color.NRGBA{R: 240, G: 1, B: 171, A: 0}) {
		t.Fatal("independent final callback/value oracle", g.mediaModel())
	}
}
