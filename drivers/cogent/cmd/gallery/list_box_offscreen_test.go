//go:build offscreen

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// The runnable gallery consumes only public Vellum APIs, including actual
// offscreen key routing. This opens no native application window.
func TestGalleryListBoxActualPublicRouting(t *testing.T) {
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
	a, err := app.New(app.Config{QueueCapacity: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	g := newGallery()
	g.tab = "lists"
	g.focus = ui.FocusRequest{ID: "layers", Sequence: 1}
	if err := a.Open(app.Window{ID: "gallery", Title: "Lists", Width: 1000, Height: 760, View: g.view}); err != nil {
		t.Fatal(err)
	}
	config := cogent.ManualConfig{Clock: clock, Executor: executor, LogicalDPI: 160, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
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
		config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(sum[:])})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	m, err := cogent.NewManual(ctx, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(ctx)
	settle := func() {
		t.Helper()
		for range 1024 {
			s, err := m.Snapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
			for _, w := range s.Windows {
				ready = ready || w.Events > 0 || w.ReadyPaint
				if !ready && w.LayoutPending {
					if err := m.RequestFrame(ctx, w.ID); err != nil {
						t.Fatal(err)
					}
					ready = true
				}
			}
			if !ready {
				return
			}
			if _, err := m.Step(ctx, 7); err != nil {
				t.Fatal(err)
			}
		}
		t.Fatal("gallery did not settle")
	}
	settle()
	initial, err := m.Capture(ctx, "gallery")
	if err != nil {
		t.Fatal(err)
	}
	s, err := m.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if s.Windows[0].FocusedID != "layers" {
		t.Fatal("list initial focus missing", s.Windows[0].FocusedID)
	}
	incarnation := s.Windows[0].Incarnation
	send := func(code key.Codes) {
		t.Helper()
		if err := m.Input(ctx, cogent.ManualInput{Window: "gallery", WindowIncarnation: incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		settle()
	}
	send(key.CodeEnd)
	if g.listSelection != "layer-23" || g.listRequests != 1 {
		t.Fatal("gallery End request failed", g.listSelection, g.listRequests)
	}
	completed, err := m.Capture(ctx, "gallery")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(initial.Pix, completed.Pix) {
		t.Fatal("gallery selection/scroll produced no pixels")
	}
	g.listReject = true
	a.Invalidate("gallery")
	settle()
	send(key.CodeHome)
	if g.listSelection != "layer-23" || g.listRequests != 2 {
		t.Fatal("gallery rejection bypassed model")
	}
	g.listReadOnly = true
	a.Invalidate("gallery")
	settle()
	send(key.CodeEnd)
	if g.listRequests != 2 {
		t.Fatal("readonly gallery changed selection")
	}
	send(key.CodeTab)
	s, err = m.Snapshot(ctx)
	if err != nil || s.Windows[0].FocusedID != "list-reject" {
		t.Fatal("public list Tab traversal failed", s, err)
	}
}
