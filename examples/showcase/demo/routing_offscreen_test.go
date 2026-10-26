//go:build offscreen

package demo

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

func settleDemo(t *testing.T, m *cogent.Manual) {
	t.Helper()
	ctx := context.Background()
	for range 128 {
		s, e := m.Snapshot(ctx)
		if e != nil {
			t.Fatal(e)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
		}
		if ready {
			if _, e = m.Step(ctx, 8); e != nil {
				t.Fatal(e)
			}
			continue
		}
		layout := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				if _, e = m.Frame(ctx, w.ID); e != nil {
					t.Fatal(e)
				}
				layout = true
			}
		}
		if !layout {
			return
		}
	}
	t.Fatal("routing did not settle")
}

// controlsOnly preserves the managed controls and layout while replacing custom
// GPU extensions that Manual explicitly does not support. Canvas geometry is
// covered by model tests; native rendering is checked separately.
func controlsOnly(n ui.Node) ui.Node {
	if n.Kind == ui.Extension {
		n.Kind = ui.Column
		n.Custom = nil
	}
	for i := range n.Children {
		n.Children[i] = controlsOnly(n.Children[i])
	}
	return n
}

func mountDemo(t *testing.T, kind string, model Application) *cogent.Manual {
	t.Helper()
	ctx := context.Background()
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	ref, e := a.OpenWindow(app.Window{ID: kind, Title: kind, Width: 1400, Height: 880, View: func() ui.Node { return controlsOnly(model.View()) }, DisableRootScrolling: true})
	if e != nil {
		t.Fatal(e)
	}
	model.Bind(a, ref)
	clock, e := execution.NewManualClock(time.Unix(1, 0), 128, 128)
	if e != nil {
		t.Fatal(e)
	}
	executor, e := execution.NewManual(8)
	if e != nil {
		t.Fatal(e)
	}
	fonts := []cogent.ManualFont{}
	for _, f := range []struct {
		name string
		data []byte
	}{{"Go-Regular.ttf", goregular.TTF}, {"Go-Bold.ttf", gobold.TTF}} {
		sum := sha256.Sum256(f.data)
		fonts = append(fonts, cogent.ManualFont{Name: f.name, Data: f.data, SHA256: fmt.Sprintf("%x", sum)})
	}
	managedTheme := theme(kind == "pulse")
	managedTheme.Default.Background = ui.Some(managedTheme.Palette.Value.Window)
	manual, e := cogent.NewManual(ctx, a, cogent.ManualConfig{Clock: clock, Executor: executor, Fonts: fonts, Theme: managedTheme, Commands: model.Commands(), LogicalDPI: 160, MaxWindows: 1, EventCapacity: 128, OwnerBatch: 8})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := manual.Close(ctx); e != nil {
			t.Error(e)
		}
		a.Close()
		executor.Close()
		clock.Close()
	})
	settleDemo(t, manual)
	return manual
}
func clickDemo(t *testing.T, m *cogent.Manual, id string) {
	t.Helper()
	ctx := context.Background()
	s, e := m.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	w := s.Windows[0]
	for _, element := range w.Elements {
		if element.ID == id {
			if !element.Enabled || !element.Visible || element.Bounds.Empty() {
				t.Fatal("unusable control", id, element.Bounds)
			}
			p := element.Bounds.Min.Add(element.Bounds.Size().Div(2))
			for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
				if e = m.Input(ctx, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: kind, Button: events.Left, Position: p}); e != nil {
					t.Fatal(e)
				}
			}
			settleDemo(t, m)
			return
		}
	}
	t.Fatal("missing control", id)
}
func TestShowcaseRoutedControls(t *testing.T) {
	t.Run("forma", func(t *testing.T) {
		model, _ := NewForma()
		m := mountDemo(t, "forma", model)
		clickDemo(t, m, "forma-add-square")
		if len(model.state.Shapes) != 4 || model.journal.Snapshot().UndoCount != 1 {
			t.Fatal("routed add did not reach history")
		}
	})
	t.Run("luma", func(t *testing.T) {
		model, _ := NewLuma()
		m := mountDemo(t, "luma", model)
		clickDemo(t, m, "luma-favorites")
		if model.filter != "favorites" || len(model.visible()) != 3 {
			t.Fatal("routed favorite filter")
		}
	})
	t.Run("pulse", func(t *testing.T) {
		model, _ := NewPulse()
		m := mountDemo(t, "pulse", model)
		clickDemo(t, m, "pulse-degraded")
		if model.filter != "degraded" || model.source.RowCount() != 93 {
			t.Fatal("routed service filter")
		}
		clickDemo(t, m, "pulse-advance")
		if model.sample != 1 {
			t.Fatal("routed sample advance")
		}
	})
}
