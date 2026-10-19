//go:build offscreen

package calculator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
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
)

func settleViewport(t *testing.T, m *cogent.Manual) {
	t.Helper()
	ctx := context.Background()
	for range 128 {
		s, err := m.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
		}
		if ready {
			if _, err = m.Step(ctx, 7); err != nil {
				t.Fatal(err)
			}
			continue
		}
		pending := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				if _, err = m.Frame(ctx, w.ID); err != nil {
					t.Fatal(err)
				}
				pending = true
			}
		}
		if !pending {
			return
		}
	}
	t.Fatal("bounded viewport settlement exhausted")
}

// Real software layout and pointer routing establish that all keys are usable
// inside the fixed content viewport. This is not physical desktop-pointer proof.
func TestFixedViewportFitsEveryKeyAndRoutesCalculation(t *testing.T) {
	ctx := context.Background()
	model := New()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	// Offscreen Standard decoration keeps its existing 8dp scene inset. The
	// native Unified runner has no such inset: its full client is 240x420dp.
	if err = a.Open(app.Window{ID: "calculator", Title: "Calculator viewport", Width: 256, Height: 436, FixedSize: true, DisableRootScrolling: true, View: model.View}); err != nil {
		t.Fatal(err)
	}
	clock, err := execution.NewManualClock(time.Unix(1, 0), 256, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer clock.Close()
	executor, err := execution.NewManual(8)
	if err != nil {
		t.Fatal(err)
	}
	defer executor.Close()
	config := cogent.ManualConfig{Clock: clock, Executor: executor, Theme: Theme(), LogicalDPI: 160, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128}
	err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || (!strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf")) {
			return nil
		}
		data, err := fs.ReadFile(fonts.Default, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(sum[:])})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	m, err := cogent.NewManual(ctx, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := m.Close(ctx); err != nil {
			t.Error(err)
		}
	}()
	settleViewport(t, m)
	s, err := m.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	w := s.Windows[0]
	viewport := image.Rect(8, 8, 248, 428)
	keys := 0
	for _, e := range w.Elements {
		if e.Kind != ui.Button || !strings.HasPrefix(e.ID, "key-") {
			continue
		}
		keys++
		if !e.Visible || e.Bounds.Dx() < 46 || e.Bounds.Dy() < 46 || !e.Bounds.In(viewport) {
			t.Errorf("key clipped/hidden: %s %v inside %v", e.ID, e.Bounds, viewport)
		}
	}
	if keys != 20 {
		t.Fatal("key count", keys)
	}
	for _, id := range []string{"key-3-0", "key-3-3", "key-3-1", "key-4-3"} {
		s, err = m.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		w = s.Windows[0]
		var target cogent.ManualElement
		for _, e := range w.Elements {
			if e.ID == id {
				target = e
				break
			}
		}
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err = m.Input(ctx, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: id, TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}); err != nil {
				t.Fatal(err)
			}
		}
		settleViewport(t, m)
	}
	if model.Display() != "3" {
		t.Fatal("native routing result", model.Display())
	}
}
