//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"image"
	"strings"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/ui"
)

// Actual software FIFO/owner/widget routing with a controlled receipt clock.
// The software compositor must refuse successful native timing; these tests do
// not establish Metal callbacks or physical input/presentation.
func measuredKeyFixture(t *testing.T) (*DesktopSession, *Manual, *nativeWindow, *int) {
	t.Helper()
	s, m, a, w, count := desktopFixture(t)
	s.measurements = &desktopMeasurements{capacity: 8, samples: make([]DesktopInputMeasurement, 0, 8), clock: func() float64 { return 100 }}
	pm, err := NewPresentationMetrics(128)
	if err != nil {
		t.Fatal(err)
	}
	w.presentation, err = pm.bind(w.window.ID)
	if err != nil {
		t.Fatal(err)
	}
	w.body.Scene.OnFirst(events.KeyChord, func(e events.Event) {
		if e.Type() == events.KeyChord {
			if got := w.presentation.recorder.Snapshot(); len(got.Inputs) != 1 {
				t.Fatal("input effect preceded native receipt")
			}
			if err := a.PostWindow(w.window.Ref(), func() { *count++; _ = a.Invalidate(w.window.ID) }); err != nil {
				t.Fatal(err)
			}
			e.SetHandled()
		}
	})
	desktopFrame(t, s, m, w)
	return s, m, w, count
}
func measuredInput(t *testing.T, s *DesktopSession, w *nativeWindow, check DesktopInputExpectation) *desktopRequest {
	t.Helper()
	s.mu.Lock()
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	r, err := s.submitWithExpectation(testContext, w.window.Ref(), 5, DesktopInput{Frame: frame, Type: events.KeyChord, Code: key.CodeSpacebar, Rune: ' '}, check)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestMeasuredDesktopActualFIFOReceiptBeforeOwnerAndNoSyntheticSuccess(t *testing.T) {
	s, m, w, count := measuredKeyFixture(t)
	checks := 0
	r := measuredInput(t, s, w, func() (uint64, error) {
		checks++
		if *count != 1 {
			return 0, errors.New("expected owner action absent")
		}
		return 71, nil
	})
	out := desktopPump(t, s, m, r)
	if *count != 1 || checks != 1 {
		t.Fatalf("action/check count %d/%d; error %v", *count, checks, out.err)
	}
	if out.err == nil || out.measurement.Frame.Completed {
		t.Fatal("software frame falsely qualified native timing")
	}
	// Application state acceptance is independent of native composition. The
	// actual owner transition may validate once, but a software/unchanged final
	// frame must still fail rather than inventing a native composition result.
	if out.measurement.AcceptedRevision != 71 {
		t.Fatal("independently validated application revision was lost")
	}

	if out.measurement.Scope != w.presentation.scope || out.measurement.ReceiptStartedAt.IsZero() || out.measurement.ReceiptFinishedAt.Before(out.measurement.ReceiptStartedAt) {
		t.Fatal("missing or backwards actual receipt clock bracket")
	}
	if out.measurement.Frame.Tag != r.tag || r.measured.check != nil {
		t.Fatal("wrong tag or retained expectation")
	}
	snap := w.presentation.recorder.Snapshot()
	if len(snap.Inputs) != 1 || snap.Inputs[0].Tag != r.tag || snap.Inputs[0].Received != 100 {
		t.Fatalf("receipt %+v", snap)
	}
	retained := s.Measurements()
	if retained.Pending || len(retained.Samples) != 1 {
		t.Fatalf("retention %+v", retained)
	}
	retained.Samples[0].Frame.Window = "changed"
	if s.Measurements().Samples[0].Frame.Window != w.window.ID {
		t.Fatal("borrowed records")
	}
}
func TestMeasuredDesktopRejectsNoChangeAndExpectationFailure(t *testing.T) {
	t.Run("no transition", func(t *testing.T) {
		s, m, w, _ := measuredKeyFixture(t)
		s.mu.Lock()
		frame := s.windows[w.window.Ref()].frame
		s.mu.Unlock()
		checks := 0
		r, err := s.submitWithExpectation(testContext, w.window.Ref(), 5, DesktopInput{Frame: frame, Type: events.MouseMove, Position: image.Pt(1, 1)}, func() (uint64, error) { checks++; return 99, nil })
		if err != nil {
			t.Fatal(err)
		}
		out := desktopPump(t, s, m, r)
		if out.err == nil || checks != 0 || out.measurement.Frame.Completed || r.measured.check != nil {
			t.Fatalf("unchanged event admitted: %+v checks%d", out, checks)
		}
	})

}
func TestMeasuredDesktopBoundsStaleAndQueuedCancellation(t *testing.T) {
	s, m, w, _ := measuredKeyFixture(t)
	s.measurements.capacity = 1
	s.mu.Lock()
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	ctx, cancel := context.WithCancel(testContext)
	checks := 0
	r, err := s.submitWithExpectation(ctx, w.window.Ref(), 5, DesktopInput{Frame: frame, Type: events.KeyChord, Code: key.CodeSpacebar, Rune: ' '}, func() (uint64, error) { checks++; return 3, nil })
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	s.cancel(r)
	if !r.done.Load() || checks != 0 || r.measured.check != nil {
		t.Fatal("queued cancellation retained work")
	}
	if _, err := s.submitWithExpectation(testContext, w.window.Ref(), 5, DesktopInput{Frame: frame, Type: events.KeyChord, Code: key.CodeSpacebar}, func() (uint64, error) { return 3, nil }); !errors.Is(err, ErrDesktopBudget) {
		t.Fatal("capacity reused", err)
	}
	unitSettle(t, m)
	if got := s.Measurements(); got.Pending || len(got.Samples) != 1 || got.Samples[0].Frame.Failure == "" {
		t.Fatalf("failed record lost %+v", got)
	}
	for _, capacity := range []int{0, 65537} {
		if _, err := NewMeasuredDesktopSession(capacity); err == nil {
			t.Fatal("invalid capacity", capacity)
		}
	}
}
func TestMeasuredDesktopRetirementReleasesExpectationAndRetainsFailure(t *testing.T) {
	s, _, w, _ := measuredKeyFixture(t)
	r := measuredInput(t, s, w, func() (uint64, error) { t.Fatal("retired callback"); return 0, nil })
	s.retire(w)
	if !r.done.Load() || r.measured.check != nil || s.Measurements().Pending {
		t.Fatal("retired operation retained references")
	}
	if got := s.Measurements().Samples; len(got) != 1 || got[0].Frame.Completed || got[0].Frame.Failure == "" {
		t.Fatalf("retirement missing %+v", got)
	}
}

// Hand-edited view metadata cannot substitute for actual original Core trace
// terminals, nor authorize an application callback or native composition.
func TestMeasuredDesktopMissingCausalTraceSkipsExpectation(t *testing.T) {
	s, _, w, _ := measuredKeyFixture(t)
	calls := 0
	r := measuredInput(t, s, w, func() (uint64, error) { calls++; return 71, nil })
	s.mu.Lock()
	r.kind = 6
	r.measured.before = 1
	r.binding.revision = 2
	r.measured.reconciled = 2
	s.mu.Unlock()
	s.finishMeasuredInput(w, r)
	s.finishMeasuredInput(w, r)
	if calls != 0 || r.measured.check != nil || r.measured.accepted != 0 {
		t.Fatal("missing causal trace invoked or retained expectation", calls)
	}
	if r.binding.frame != 1 || r.done.Load() {
		t.Fatal("missing causal trace published frame")
	}
	s.retire(w)
}

func TestMeasuredDesktopStaleFrameCannotPublishReceiptOrInput(t *testing.T) {
	s, m, w, count := measuredKeyFixture(t)
	s.mu.Lock()
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	called := false
	r, err := s.submitWithExpectation(testContext, w.window.Ref(), 5, DesktopInput{Frame: frame + 1, Type: events.KeyChord, Code: key.CodeSpacebar}, func() (uint64, error) { called = true; return 1, nil })
	if err != nil {
		t.Fatal(err)
	}
	out := desktopPump(t, s, m, r)
	if !errors.Is(out.err, ErrDesktopPending) || called || *count != 0 || len(w.presentation.recorder.Snapshot().Inputs) != 0 {
		t.Fatalf("stale input escaped %+v", out)
	}
}
func TestMeasuredDesktopRoutedCancellationKeepsCausalReservation(t *testing.T) {
	s, m, w, count := measuredKeyFixture(t)
	s.mu.Lock()
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	ctx, cancel := context.WithCancel(testContext)
	r, err := s.submitWithExpectation(ctx, w.window.Ref(), 5, DesktopInput{Frame: frame, Type: events.KeyChord, Code: key.CodeSpacebar}, func() (uint64, error) { t.Fatal("canceled expectation"); return 1, nil })
	if err != nil {
		t.Fatal(err)
	}
	// Dispatch only the original admission node. Actual input remains in FIFO.
	ev, ok := w.source.Load().Deque.PollEvent()
	if !ok {
		t.Fatal("no originating node")
	}
	reserved, ok := ev.(*desktopFIFO)
	if !ok || !s.handle(w, reserved.Event) || r.kind != 6 {
		t.Fatal("input group not admitted")
	}
	cancel()
	s.cancel(r)
	if r.done.Load() || !s.Measurements().Pending || r.measured.check != nil {
		t.Fatal("routed cancellation reused trace or retained closure")
	}
	out := desktopPump(t, s, m, r)
	if !errors.Is(out.err, context.Canceled) || *count != 1 || out.measurement.Frame.Completed || s.Measurements().Pending {
		t.Fatalf("routed terminal %+v count%d", out, *count)
	}
}
func TestMeasuredDesktopFailurePayloadBound(t *testing.T) {
	message := strings.Repeat("x", 65536)
	if got := boundedDesktopMeasurementFailure(errors.New(message)); len(got) > 1024 {
		t.Fatal("unbounded error retained")
	}
}

func TestMeasuredDesktopViewInvalidationCannotRetagNewerState(t *testing.T) {
	s, _, w, _ := measuredKeyFixture(t)
	r := measuredInput(t, s, w, func() (uint64, error) { return 71, nil })
	s.mu.Lock()
	r.kind = 6
	r.measured.accepted = 71
	if err := r.binding.trace.BeginGroup(r.tag, []events.Event{&events.CustomEvent{}}); err != nil {
		s.mu.Unlock()
		t.Fatal(err)
	}
	s.mu.Unlock()
	s.invalidate(w)
	s.retire(w)
	got := s.Measurements().Samples[0]
	if got.Frame.Completed || got.Frame.Failure != "measured view invalidated after final arming" {
		t.Fatalf("newer view silently attributed: %+v", got)
	}
}

// A real KeyDown+KeyChord group uses Manual's between-event owner drain, not
// only the originating Custom ack drain. Intermediate reconciliation must be
// observed without claiming a native composition or completing the group early.
func TestMeasuredDesktopTwoEventGroupIncludesIntermediateManualDrain(t *testing.T) {
	s, m, w, count := measuredKeyFixture(t)
	firstOwnerRan := false
	w.body.Scene.OnFirst(events.KeyDown, func(e events.Event) {
		if err := s.owner.PostWindow(w.window.Ref(), func() { firstOwnerRan = true; *count++; _ = s.owner.Invalidate(w.window.ID) }); err != nil {
			t.Fatal(err)
		}
		e.SetHandled()
	})
	s.mu.Lock()
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	r, err := s.submitWithExpectation(testContext, w.window.Ref(), 5, DesktopInput{Frame: frame, Type: events.KeyDown, Code: key.CodeTab, Rune: '\t'}, func() (uint64, error) {
		if *count != 2 {
			return 0, errors.New("independently expected two-event state absent")
		}
		return 71, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Admission only. Its ordinary scheduled paint stays a separate FIFO event.
	if _, err := m.Step(testContext, 1); err != nil {
		t.Fatal(err)
	}
	if r.kind != 6 {
		t.Fatal("original two-event group not routed")
	}
	// The unchanged paint and first input consume two units; the first owner's
	// callback consumes the third, stopping before the second original input.
	step, err := m.Step(testContext, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !firstOwnerRan || *count != 1 || step.OwnerCallbacks != 1 || r.done.Load() {
		t.Fatalf("first input/drain boundary: %+v count%d ran%v", step, *count, firstOwnerRan)
	}
	if _, err := m.Step(testContext, 1); err != nil {
		t.Fatal(err)
	}
	if r.measured.reconciled <= r.measured.before {
		t.Fatal("intermediate changed refresh was omitted from group accounting")
	}
	out := desktopPump(t, s, m, r)
	if *count != 2 || out.err == nil || out.measurement.Frame.Completed || out.measurement.Frame.OwnerNS <= 0 || out.measurement.Frame.ReconcileNS <= 0 {
		t.Fatalf("incomplete group accounting or false native success %+v count%d", out, *count)
	}
}

func TestMeasuredDesktopEarlierSoftwareFailureSkipsStateObservation(t *testing.T) {
	count := 0
	s, m, a, w := desktopViewFixture(t, func() ui.Node {
		return ui.Node{ID: "visible-value", Kind: ui.Label, Text: strings.Repeat("x", count+1)}
	})
	s.measurements = &desktopMeasurements{capacity: 8, samples: make([]DesktopInputMeasurement, 0, 8), clock: func() float64 { return 100 }}
	pm, err := NewPresentationMetrics(128)
	if err != nil {
		t.Fatal(err)
	}
	w.presentation, err = pm.bind(w.window.ID)
	if err != nil {
		t.Fatal(err)
	}
	w.body.Scene.OnFirst(events.KeyChord, func(e events.Event) {
		if err := a.PostWindow(w.window.Ref(), func() { count++; _ = a.Invalidate(w.window.ID) }); err != nil {
			t.Fatal(err)
		}
		e.SetHandled()
	})
	desktopFrame(t, s, m, w)
	calls := 0
	r := measuredInput(t, s, w, func() (uint64, error) { calls++; return 71, nil })
	out := desktopPump(t, s, m, r)
	if count != 1 || calls != 0 || out.err == nil || out.measurement.Frame.Completed || out.measurement.AcceptedRevision != 0 || r.measured.check != nil {
		t.Fatalf("known earlier software failure reached app observation: %+v count%d calls%d", out, count, calls)
	}
}

// Controlled observer calls prove recording/classification only. Neither an
// ordinary drawable nor an unarmed tagged prefix claims the accepted response.
func TestMeasuredDesktopUnarmedPresentationStaysUnattributed(t *testing.T) {
	metrics, err := NewPresentationMetrics(8)
	if err != nil {
		t.Fatal(err)
	}
	sink, err := metrics.bind("main")
	if err != nil {
		t.Fatal(err)
	}
	sink.Receipt(5, 1)
	frames := []desktop.PresentationFrame{{Composition: 1}, {Tag: 5, Composition: 2}, {Tag: 5, Generation: 7, Composition: 3, Armed: true}}
	for idx, frame := range frames {
		sequence := sink.Acquired(frame, float64(idx*2+2))
		if sequence == 0 {
			t.Fatal("unarmed/ordinary acquisition rejected", idx)
		}
		sink.Presented(sequence, 0, float64(idx*2+3))
	}
	snapshot := metrics.Snapshot()
	if snapshot.Failed || snapshot.Pending != 0 || len(snapshot.Inputs) != 1 || len(snapshot.Samples) != 3 {
		t.Fatal("ordinary/unarmed drawables failed recorder", snapshot)
	}
	for idx, sample := range snapshot.Samples {
		if sample.Frame.Tag != frames[idx].Tag || sample.Frame.Generation != frames[idx].Generation || sample.Frame.Armed != frames[idx].Armed || sample.State != "presented" {
			t.Fatal("acquisition identity changed", idx, sample)
		}
	}
	if snapshot.Samples[0].Frame.Armed || snapshot.Samples[1].Frame.Armed || !snapshot.Samples[2].Frame.Armed {
		t.Fatal("unarmed prefix inherited final response")
	}
}
