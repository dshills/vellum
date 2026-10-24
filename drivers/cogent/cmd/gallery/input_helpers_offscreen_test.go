//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func inputGalleryManual(t *testing.T) (*inputHelpersGallery, *app.Application, *cogent.Manual, func() cogent.ManualWindowSnapshot) {
	return inputGalleryManualWithTheme(t, false)
}
func inputGalleryManualWithTheme(t *testing.T, dark bool) (*inputHelpersGallery, *app.Application, *cogent.Manual, func() cogent.ManualWindowSnapshot) {
	t.Helper()
	ctx := context.Background()
	clock, err := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(clock.Close)
	executor, err := execution.NewManual(8)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(executor.Close)
	a, err := app.New(app.Config{QueueCapacity: 64})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	model := newInputHelpersGallery()
	model.setAppearance(dark)
	model.focus = ui.FocusRequest{ID: ui.TokenInputID("drawing-tags"), Sequence: 1}
	if err := a.Open(app.Window{ID: "inputs", Title: "Input recovery", Width: 660, Height: 760, View: model.windowView}); err != nil {
		t.Fatal(err)
	}
	theme := ui.MacTheme(dark)
	theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
	if dark {
		theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
	}
	theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
	cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, LogicalDPI: 160, Theme: theme}
	if err := fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, e error) error {
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
	m, err := cogent.NewManual(ctx, a, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	settle := func() cogent.ManualWindowSnapshot {
		t.Helper()
		for range 256 {
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
				if _, e = m.Frame(ctx, "inputs"); e != nil {
					t.Fatal(e)
				}
				continue
			}
			return w
		}
		t.Fatal("input gallery did not settle")
		return cogent.ManualWindowSnapshot{}
	}
	settle()
	return model, a, m, settle
}
func inputGalleryTypeTag(t *testing.T, m *cogent.Manual, settle func() cogent.ManualWindowSnapshot, text string) cogent.ManualWindowSnapshot {
	t.Helper()
	w := settle()
	for _, ch := range text {
		if ch < 'a' || ch > 'z' {
			t.Fatal("fixture expects lowercase ASCII")
		}
		if err := m.Input(context.Background(), cogent.ManualInput{Window: "inputs", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.Codes(int(key.CodeA) + int(ch-'a')), Rune: ch}); err != nil {
			t.Fatal(err)
		}
		w = settle()
	}
	if err := m.Input(context.Background(), cogent.ManualInput{Window: "inputs", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	return settle()
}
func TestInputGalleryRapidTagEntryAndValidationRecovery(t *testing.T) {
	model, _, m, settle := inputGalleryManual(t)
	for i, text := range []string{"alpha", "beta"} {
		w := inputGalleryTypeTag(t, m, settle, text)
		if len(model.tokens.Tokens) != i+2 || model.tokens.Tokens[i+1].Text != text || model.tokenError != "" {
			t.Fatal("tag did not commit exactly once", text, model.tokens.Tokens, model.tokenError)
		}
		if w.FocusedID != ui.TokenInputID("drawing-tags") {
			t.Fatal("accepted tag moved focus away from repeat entry", w.FocusedID)
		}
	}
	w := inputGalleryTypeTag(t, m, settle, "alpha")
	if len(model.tokens.Tokens) != 3 || !strings.Contains(model.tokenError, "already exists") || w.FocusedID != ui.TokenInputID("drawing-tags") {
		t.Fatal("duplicate recovery lost accepted values or entry focus", w.FocusedID, model.tokenError)
	}
	w = inputGalleryTypeTag(t, m, settle, "gamma")
	if len(model.tokens.Tokens) != 4 || model.tokens.Tokens[3].Text != "gamma" || model.tokenError != "" || w.FocusedID != ui.TokenInputID("drawing-tags") {
		t.Fatal("correction did not recover tag entry", w.FocusedID, model.tokenError)
	}
}
func TestInputGalleryTagFocusCapacityAndDisable(t *testing.T) {
	model, a, m, settle := inputGalleryManual(t)
	if err := a.Do(func() {
		model.tokens.Tokens = nil
		for i := range ui.MaxTokens - 1 {
			model.tokens.Tokens = append(model.tokens.Tokens, ui.Token{ID: fmt.Sprintf("seed-%d", i), Text: fmt.Sprintf("Existing%d", i)})
		}
		_ = a.Invalidate("inputs")
	}); err != nil {
		t.Fatal(err)
	}
	settle()
	w := inputGalleryTypeTag(t, m, settle, "final")
	if len(model.tokens.Tokens) != ui.MaxTokens || w.FocusedID != ui.ColorWellButtonID("drawing-color") {
		t.Fatal("capacity boundary did not advance to an enabled field", len(model.tokens.Tokens), w.FocusedID)
	}
	// Multiple accepted UI actions can share an owner batch before reconciliation.
	if err := a.Post(func() {
		model.tokens.Change(ui.TokenEvent{TokenID: "seed-0", Remove: true})
		model.setDisabled(true)
		_ = a.Invalidate("inputs")
	}); err != nil {
		t.Fatal(err)
	}
	settle()
	if model.windowView().Focus.ID != "" {
		t.Fatal("disabled demo kept a pending field focus request")
	}
	if err := a.Do(func() { model.setDisabled(false); _ = a.Invalidate("inputs") }); err != nil {
		t.Fatal(err)
	}
	settle()
	if len(model.tokens.Tokens) != ui.MaxTokens-1 || model.windowView().Focus.ID != "" {
		t.Fatal("re-enable replayed a canceled request")
	}
}

func TestInputGalleryQueuedTagRemovalThenDisable(t *testing.T) {
	model, a, _, settle := inputGalleryManual(t)
	if err := a.Post(func() {
		model.tokens.Change(ui.TokenEvent{TokenID: "tag-1", Remove: true})
		model.setDisabled(true)
		_ = a.Invalidate("inputs")
	}); err != nil {
		t.Fatal(err)
	}
	settle()
	if model.windowView().Focus.ID != "" {
		t.Fatal("disabled demo kept a pending field focus request")
	}
}

func TestInputGalleryTagFeedbackVisibleAtEntry(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			model, _, m, settle := inputGalleryManualWithTheme(t, dark)
			inputGalleryTypeTag(t, m, settle, "alpha")
			w := inputGalleryTypeTag(t, m, settle, "alpha")
			var feedback, entry cogent.ManualElement
			for _, e := range w.Elements {
				if e.ID == "token-feedback" {
					feedback = e
				}
				if e.ID == ui.TokenInputID("drawing-tags") {
					entry = e
				}
			}
			frame, err := m.Capture(context.Background(), "inputs")
			if err != nil {
				t.Fatal(err)
			}
			if w.FocusedID != entry.ID || !feedback.Visible || feedback.Bounds.Empty() || feedback.Bounds.Dy() < 12 || !feedback.Bounds.In(frame.Bounds()) || !entry.Bounds.In(frame.Bounds()) {
				t.Fatal("tag rejection feedback is outside the focused viewport", feedback.Bounds, entry.Bounds, frame.Bounds())
			}
			if strings.HasPrefix(model.tokenError, "ui: ") {
				t.Fatal("internal package prefix leaked into user feedback", model.tokenError)
			}
		})
	}
}
