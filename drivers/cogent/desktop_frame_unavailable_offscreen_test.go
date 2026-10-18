//go:build offscreen

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"cogentcore.org/core/system/composer"
)

// Actual Manual FIFO/reconciliation/composition uses clearly synthetic measured
// drawer output. These tests prove protocol behavior, never native GPU success.
type desktopMeasuredFixture struct {
	composer.Drawer
	result composer.ComposeTiming
	calls  int
}

func (d *desktopMeasuredFixture) EndMeasured() composer.ComposeTiming {
	d.calls++
	d.Drawer.End()
	return d.result
}
func installDesktopMeasuredFixture(t *testing.T, w *nativeWindow) *desktopMeasuredFixture {
	t.Helper()
	cd := w.body.Scene.RenderWindow().SystemWindow.Composer().(*composer.ComposerDrawer)
	original := cd.Drawer
	d := &desktopMeasuredFixture{Drawer: original}
	cd.Drawer = d
	t.Cleanup(func() { cd.Drawer = original })
	return d
}
func desktopPaintWithoutReap(t *testing.T, m *Manual, d *desktopMeasuredFixture) {
	t.Helper()
	before := d.calls
	for range 256 {
		if _, err := m.Step(testContext, 1); err != nil {
			t.Fatal(err)
		}
		if d.calls != before {
			return
		}
	}
	t.Fatal("scripted measured composition not reached")
}
func TestDesktopFrameUnavailableTerminalAndExplicitRecovery(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	desktopFrame(t, s, m, w)
	old := s.windows[w.window.Ref()].frame
	d := installDesktopMeasuredFixture(t, w)
	d.result = composer.ComposeTiming{Supported: true, Unavailable: true, Failure: "scripted occluded"}
	r, err := s.submit(testContext, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	out := desktopPump(t, s, m, r)
	var classified *DesktopCompositionError
	if !errors.Is(out.err, ErrDesktopFrameUnavailable) || !errors.As(out.err, &classified) || classified.Failure != "scripted occluded" || !classified.Unavailable {
		t.Fatal("checked terminal lost", out.err)
	}
	if s.pending != nil || r.binding.trace.Pending() || r.binding.frame != 0 || d.calls != 1 {
		t.Fatal("failed frame retained capacity/token", s.State(), d.calls)
	}
	issued := s.State().Issued
	if _, err := m.Step(testContext, 128); err != nil {
		t.Fatal(err)
	}
	if d.calls != 1 || s.State().Issued != issued || s.State().Pending {
		t.Fatal("unavailable auto-requeued", s.State(), d.calls)
	}
	// Explicit application decision only; no availability notification is inferred.
	d.result = composer.ComposeTiming{Supported: true, Submitted: true}
	r, err = s.submit(testContext, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	if out = desktopPump(t, s, m, r); out.err != nil || r.binding.frame != r.tag || r.binding.frame == old || d.calls != 2 {
		t.Fatal("new explicit request failed/reset incorrectly", out.err, d.calls)
	}
	in, err := s.submit(testContext, w.window.Ref(), 1, DesktopInput{Frame: old})
	if err != nil {
		t.Fatal(err)
	}
	if out = desktopPump(t, s, m, in); !errors.Is(out.err, ErrDesktopPending) {
		t.Fatal("stale frame revived", out.err)
	}
	d.result = composer.ComposeTiming{Supported: true, Failure: "scripted backend error"}
	r, err = s.submit(testContext, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	if out = desktopPump(t, s, m, r); out.err == nil || errors.Is(out.err, ErrDesktopFrameUnavailable) || r.binding.frame != 0 || s.pending != nil || d.calls != 3 {
		t.Fatal("fatal error became unavailable/success", out.err)
	}
}
func TestDesktopFrameCancellationPreservesContextIdentity(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := s.submit(ctx, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	s.cancel(r)
	if s.pending != r || !r.binding.trace.Pending() {
		t.Fatal("queued input/owner terminal released prematurely")
	}
	if _, err = s.submit(testContext, w.window.Ref(), 2, DesktopInput{}); !errors.Is(err, ErrDesktopBusy) {
		t.Fatal("pending cancellation lost capacity", err)
	}
	if out := desktopPump(t, s, m, r); !errors.Is(out.err, context.Canceled) || errors.Is(out.err, ErrDesktopFrameUnavailable) {
		t.Fatal("canceled Frame lost context identity", out.err)
	}
	if s.pending != nil || r.binding.trace.Pending() || r.binding.frame != 0 {
		t.Fatal("canceled terminal leaked")
	}
}
func TestDesktopFrameCancellationWinsOnlyAfterCompositionTerminal(t *testing.T) {
	for _, result := range []composer.ComposeTiming{{Supported: true, Unavailable: true, Failure: "yield"}, {Supported: true, Failure: "fatal"}, {Supported: true, Submitted: true}} {
		t.Run(result.Failure, func(t *testing.T) {
			s, m, _, w, _ := desktopFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			d := installDesktopMeasuredFixture(t, w)
			d.result = result
			r, err := s.submit(ctx, w.window.Ref(), 2, DesktopInput{})
			if err != nil {
				t.Fatal(err)
			}
			desktopPaintWithoutReap(t, m, d)
			cancel()
			s.mu.Lock()
			s.reapLocked()
			s.mu.Unlock()
			out := <-r.reply
			if !errors.Is(out.err, context.Canceled) || errors.Is(out.err, ErrDesktopFrameUnavailable) || s.pending != nil || r.binding.trace.Pending() || r.binding.frame != 0 || d.calls != 1 {
				t.Fatal("canceled terminal published/retained/misclassified", out.err, s.State(), d.calls)
			}
		})
	}
}
func TestDesktopFrameUnavailableRetirementAndMetricsCopy(t *testing.T) {
	t.Run("close-pending", func(t *testing.T) {
		s, _, _, w, _ := desktopFixture(t)
		d := installDesktopMeasuredFixture(t, w)
		r, err := s.submit(testContext, w.window.Ref(), 2, DesktopInput{})
		if err != nil {
			t.Fatal(err)
		}
		s.retire(w)
		out := <-r.reply
		if !errors.Is(out.err, ErrDesktopClosed) || errors.Is(out.err, ErrDesktopFrameUnavailable) || s.pending != nil || r.binding.trace.Pending() || d.calls != 0 {
			t.Fatal("retirement retained/misclassified pending request", out.err, s.State())
		}
	})
	t.Run("metrics-copy", func(t *testing.T) {
		s, m, _, w, _ := desktopFixture(t)
		d := installDesktopMeasuredFixture(t, w)
		d.result = composer.ComposeTiming{Supported: true, Unavailable: true, Failure: "scripted yield"}
		r, err := s.submit(testContext, w.window.Ref(), 2, DesktopInput{})
		if err != nil {
			t.Fatal(err)
		}
		desktopPaintWithoutReap(t, m, d)
		// At this actual trace terminal, hand the value to the recorder solely to
		// exercise production copying. This is not simultaneous observer binding;
		// retire the session without granting it a completed frame afterward.
		metrics, _ := NewFrameMetrics(1)
		metrics.windows = []*frameWindowMetrics{{trace: r.binding.trace, key: frameWindowKey{window: "one"}}}
		metrics.tags[r.tag] = 0
		got := metrics.Snapshot()
		if len(got.Samples) != 1 || !got.Samples[0].Unavailable || got.Samples[0].Completed || got.Samples[0].Failure != "scripted yield" || got.Pending != 0 {
			t.Fatal("metrics lost/misqualified terminal", got)
		}
		wire, err := json.Marshal(got)
		if err != nil || !strings.Contains(string(wire), `"Unavailable":true`) {
			t.Fatal("availability missing from failure wire", string(wire), err)
		}
		got.Samples[0].Unavailable = false
		if !metrics.Poll(r.tag).Sample.Unavailable {
			t.Fatal("caller mutated retained availability")
		}
		s.retire(w)
		if out := <-r.reply; !errors.Is(out.err, ErrDesktopClosed) {
			t.Fatal(out.err)
		}
	})
}
