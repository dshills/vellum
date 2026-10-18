//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"image"
	"reflect"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// This fixture exercises actual Core FIFO/hit-testing and composition on the
// manual software window. It is not native desktop or physical input evidence.
func desktopFixture(t *testing.T) (*DesktopSession, *Manual, *app.Application, *nativeWindow, *int) {
	t.Helper()
	clicks := new(int)
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: "Place", Activate: func() { *clicks++ }}}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one").state
	s := NewDesktopSession()
	if err := s.start(a); err != nil {
		t.Fatal(err)
	}
	if err := s.bind(w); err != nil {
		t.Fatal(err)
	}
	w.body.Scene.SetOriginCustomHandler(func(e events.Event) bool {
		if !s.owns(w, e) {
			return false
		}
		// The production driver drains/refreshed owner work before session handling.
		m.implementation.custom(w, e)
		if err := m.implementation.refresh(m.implementation.find("one")); err != nil {
			t.Fatal(err)
		}
		s.handle(w, e)
		return true
	})
	t.Cleanup(func() { s.retire(w); s.close() })
	return s, m, a, w, clicks
}
func desktopPump(t *testing.T, s *DesktopSession, m *Manual, r *desktopRequest) desktopReply {
	t.Helper()
	for range 256 {
		if _, err := m.Step(testContext, 7); err != nil {
			t.Fatal(err)
		}
		s.mu.Lock()
		s.reapLocked()
		s.mu.Unlock()
		select {
		case out := <-r.reply:
			return out
		default:
		}
	}
	t.Fatal("desktop request did not terminate")
	return desktopReply{}
}

// Install a test-only capability after an actual software frame. The offscreen
// compositor has no encoding timing and cannot positively prove native Frame.
// TestDesktopFrameRejectsUnsupportedComposition separately asserts that refusal.
func desktopFrame(t *testing.T, s *DesktopSession, m *Manual, w *nativeWindow) {
	t.Helper()
	unitSettle(t, m)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	b := s.windows[w.window.Ref()]
	b.frame = s.next
	b.completedRevision = b.revision
}
func TestDesktopFrameRejectsUnsupportedComposition(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	r, err := s.submit(testContext, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	out := desktopPump(t, s, m, r)
	if out.err == nil {
		t.Fatal("software compositor falsely qualified native composition")
	}
	s.mu.Lock()
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	if frame != 0 {
		t.Fatal("failed composition published capability")
	}
}
func desktopRead(t *testing.T, s *DesktopSession, m *Manual, w *nativeWindow) DesktopSnapshot {
	t.Helper()
	r, err := s.submit(testContext, w.window.Ref(), 0, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	out := desktopPump(t, s, m, r)
	if out.err != nil {
		t.Fatal(out.err)
	}
	return out.snapshot
}
func TestDesktopCompositionSnapshotAndActualClick(t *testing.T) {
	s, m, _, w, clicks := desktopFixture(t)
	desktopFrame(t, s, m, w)
	snap := desktopRead(t, s, m, w)
	var id string
	var bounds image.Rectangle
	for _, el := range snap.Elements {
		if el.DomainID == "button" {
			id, bounds = el.ID, el.Bounds
		}
	}
	if id == "" || bounds.Empty() || snap.Frame == 0 {
		t.Fatal("no completed button geometry")
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		desktopFrame(t, s, m, w)
		snap = desktopRead(t, s, m, w)
		in := DesktopInput{Frame: snap.Frame, Target: id, Type: typ, Button: events.Left, Position: bounds.Min.Add(bounds.Size().Div(2))}
		r, err := s.submit(testContext, w.window.Ref(), 1, in)
		if err != nil {
			t.Fatal(err)
		}
		if out := desktopPump(t, s, m, r); out.err != nil {
			t.Fatal(out.err)
		}
		stale, err := s.submit(testContext, w.window.Ref(), 0, DesktopInput{})
		if err != nil {
			t.Fatal(err)
		}
		if out := desktopPump(t, s, m, stale); !errors.Is(out.err, ErrDesktopPending) {
			t.Fatal("native-only input preserved stale frame", out.err)
		}
	}
	unitSettle(t, m)
	if *clicks != 1 {
		t.Fatalf("clicks=%d", *clicks)
	}
	// Mutating a returned snapshot cannot alter a later inspection.
	snap.Elements[0].ID = "changed"
	desktopFrame(t, s, m, w)
	fresh := desktopRead(t, s, m, w)
	if fresh.Elements[0].ID == "changed" {
		t.Fatal("borrowed snapshot")
	}
}
func TestDesktopCanceledQueuedInputRemovedAndCapacityReusable(t *testing.T) {
	s, m, _, w, clicks := desktopFixture(t)
	desktopFrame(t, s, m, w)
	snap := desktopRead(t, s, m, w)
	ctx, cancel := context.WithCancel(context.Background())
	r, err := s.submit(ctx, w.window.Ref(), 1, DesktopInput{Frame: snap.Frame, Type: events.MouseDown, Button: events.Left, Position: image.Pt(20, 20)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.submit(testContext, w.window.Ref(), 0, DesktopInput{}); !errors.Is(err, ErrDesktopBusy) {
		t.Fatalf("second admission: %v", err)
	}
	cancel()
	s.cancel(r)
	if out := <-r.reply; !errors.Is(out.err, context.Canceled) {
		t.Fatal(out.err)
	}
	source := w.source.Load()
	source.Deque.Mu.Lock()
	for _, e := range source.Deque.Back {
		if e, ok := e.(*desktopFIFO); ok && e.request == r {
			t.Error("canceled FIFO tombstone")
		}
	}
	source.Deque.Mu.Unlock()
	unitSettle(t, m)
	if *clicks != 0 {
		t.Fatal("canceled input delivered")
	}
	desktopRead(t, s, m, w)
}
func TestDesktopCanceledFramePreservesTraceTerminalAndOldFrameInvalid(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	desktopFrame(t, s, m, w)
	old := desktopRead(t, s, m, w)
	ctx, cancel := context.WithCancel(context.Background())
	r, err := s.submit(ctx, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	s.cancel(r)
	if s.pending != r || !r.binding.trace.Pending() {
		t.Fatal("cancellation prematurely retired undispatched trace")
	}
	out := desktopPump(t, s, m, r)
	if out.err == nil {
		t.Fatal("canceled frame succeeded")
	}
	in := DesktopInput{Frame: old.Frame, Type: events.MouseDown, Button: events.Left, Position: image.Pt(20, 20)}
	next, err := s.submit(testContext, w.window.Ref(), 1, in)
	if err != nil {
		t.Fatal(err)
	}
	if out := desktopPump(t, s, m, next); !errors.Is(out.err, ErrDesktopPending) {
		t.Fatalf("old frame admitted: %v", out.err)
	}
	desktopFrame(t, s, m, w)
	desktopRead(t, s, m, w)
}
func TestDesktopRevisionAndWindowRetirementRevokeCapabilities(t *testing.T) {
	s, m, a, w, _ := desktopFixture(t)
	desktopFrame(t, s, m, w)
	s.invalidate(w)
	r, err := s.submit(testContext, w.window.Ref(), 0, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	if out := desktopPump(t, s, m, r); !errors.Is(out.err, ErrDesktopPending) {
		t.Fatal(out.err)
	}
	desktopFrame(t, s, m, w)
	pending, err := s.submit(testContext, w.window.Ref(), 0, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	old := w.window.Ref()
	s.retire(w)
	if out := <-pending.reply; !errors.Is(out.err, ErrDesktopClosed) {
		t.Fatal(out.err)
	}
	if len(s.windows) != 0 || s.pending != nil || pending.binding.window != nil {
		t.Fatal("retained retired native window")
	}
	if _, err := s.submit(testContext, old, 0, DesktopInput{}); !errors.Is(err, ErrDesktopClosed) {
		t.Fatal(err)
	}
	if err := a.Open(app.Window{ID: "survivor", Title: "survivor", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "survivor", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	if err := a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	if err := a.Open(app.Window{ID: "one", Title: "replacement", Width: 400, Height: 240, View: func() ui.Node { return ui.Node{ID: "new", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.submit(testContext, old, 0, DesktopInput{}); !errors.Is(err, ErrDesktopClosed) {
		t.Fatal("same-name replacement revived ref", err)
	}
	s.close()
	if s.owner != nil || len(s.windows) != 0 {
		t.Fatal("retained application")
	}
}
func TestDesktopInputRefusesOSServiceShortcutsAndInvalidInput(t *testing.T) {
	for _, in := range []DesktopInput{
		{Frame: 1, Type: events.KeyChord, Code: key.CodeC, Rune: 'c', Modifiers: 1 << key.Meta},
		{Frame: 1, Type: events.KeyChord, Code: key.CodeV, Rune: 'v', Modifiers: 1 << key.Control},
		{Frame: 1, Type: events.MouseDown, Button: events.Right},
		{Type: events.MouseDown, Button: events.Left},
		{Frame: 1, Type: events.WindowPaint},
	} {
		if validateDesktopInput(in) == nil {
			t.Fatalf("admitted %+v", in)
		}
	}
	if err := validateDesktopInput(DesktopInput{Frame: 1, Type: events.KeyChord, Code: key.CodeA, Rune: 'a', Modifiers: 1 << key.Meta}); err != nil {
		t.Fatal("select all refused", err)
	}
	s := NewDesktopSession()
	copy := DesktopSession{self: s}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := copy.submit(ctx, app.WindowRef{}, 0, DesktopInput{}); !errors.Is(err, ErrDesktopSession) {
		t.Fatal("copied session accepted", err)
	}
}

func TestDesktopMoveScrollUseRequestedModifiersAndKeepOSSourceState(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	seen := map[events.Types]key.Modifiers{}
	for _, typ := range []events.Types{events.MouseMove, events.Scroll} {
		w.body.Scene.OnFirst(typ, func(e events.Event) { seen[e.Type()] = e.Modifiers() })
	}
	original := w.source.Load().Last
	for _, typ := range []events.Types{events.MouseMove, events.Scroll} {
		desktopFrame(t, s, m, w)
		snap := desktopRead(t, s, m, w)
		s.windows[w.window.Ref()].source.Last.Mods = 1 << key.Control
		r, err := s.submit(testContext, w.window.Ref(), 1, DesktopInput{Frame: snap.Frame, Type: typ, Modifiers: 1 << key.Shift, Position: image.Pt(20, 20), Delta: math32.Vec2(0, 3)})
		if err != nil {
			t.Fatal(err)
		}
		if out := desktopPump(t, s, m, r); out.err != nil {
			t.Fatal(out.err)
		}
		if seen[typ] != 1<<key.Shift {
			t.Fatalf("%v modifiers=%v", typ, seen[typ])
		}
	}
	if !reflect.DeepEqual(original, w.source.Load().Last) {
		t.Fatal("automation mutated OS event generation state")
	}
}
func TestDesktopFrameCanceledAfterDispatchReleasesOnlyTerminalTicket(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	r, err := s.submit(ctx, w.window.Ref(), 2, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	// Dispatch precisely the actual originating Custom; do not paint yet.
	step, stepErr := m.implementation.find("one").core.Step(testContext, 1)
	if stepErr != nil {
		t.Fatal(stepErr)
	}
	if step.Events != 1 || step.LastType != events.Custom || s.pending != r || r.kind != 2 || !r.binding.trace.Pending() || r.binding.trace.ReconcileTag() != 0 {
		t.Fatalf("not an in-flight dispatched originating frame: %+v", step)
	}
	cancel()
	s.cancel(r)
	out := desktopPump(t, s, m, r)
	if out.err == nil {
		t.Fatal("canceled frame succeeded")
	}
	s.mu.Lock()
	pending := s.pending
	frame := s.windows[w.window.Ref()].frame
	s.mu.Unlock()
	if pending != nil || frame != 0 || r.binding.trace.Pending() {
		t.Fatal("canceled terminal trace or capability retained")
	}
}

func TestDesktopCanvasPictureIsDetachedAndUnannotatedObjectsAreCounted(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	for range 4 {
		if _, e := f.manual.Frame(testContext, "main"); e != nil {
			t.Fatal(e)
		}
		manualDockSettle(t, f.manual, 7)
	}
	w := f.manual.implementation.find("main").state
	s := NewDesktopSession()
	if e := s.start(f.app); e != nil {
		t.Fatal(e)
	}
	if e := s.bind(w); e != nil {
		t.Fatal(e)
	}
	defer s.close()
	var first, second DesktopSnapshot
	var failure error
	capture := func(out *DesktopSnapshot) {
		if e := f.app.Do(func() { w.rootMu.Lock(); defer w.rootMu.Unlock(); *out, failure = desktopSnapshot(w, 1) }); e != nil {
			t.Fatal(e)
		}
		if failure != nil {
			for _, b := range placementWindows.windows[placementKey{f.app, "main", fixtureWindowRef(f.app, "main")}].bindings {
				if b.model != nil {
					frame := b.frame.Load()
					t.Logf("canvas frame=%v picture=%v pending=%v modelview=%+v", frame != nil, b.picture.Load() != nil, w.body.Scene.GlobalLayoutPending(), b.model.View)
					if frame != nil {
						t.Logf("frameview=%+v pictureeq=%v completed=%v", frame.geometry.View, frame.picture == b.picture.Load(), canvasCompletedFrame(b, frame))
					}
				}
			}
			t.Fatal(failure)
		}
	}
	capture(&first)
	if len(first.Canvases) != 1 || first.Canvases[0].TotalObjects != 0 || len(first.Canvases[0].Objects) != 1 || first.Canvases[0].Objects[0].ID != "existing" || first.Canvases[0].Selected != "existing" {
		t.Fatalf("incorrect unannotated completed picture: %+v", first.Canvases)
	}
	first.Canvases[0].Objects[0].X = 123456
	capture(&second)
	if second.Canvases[0].Objects[0].X == 123456 || f.model.Objects[0].X == 123456 {
		t.Fatal("desktop picture shares model or retained frame storage")
	}
	w.body.Scene.SetOriginCustomHandler(func(events.Event) bool { return true })
	s.close()
	if out := s.State(); !out.Closed || out.Pending || out.LiveWindows != 0 || s.owner != nil {
		t.Fatalf("session references retained: %+v", out)
	}
}

func TestDesktopSnapshotPublishesActualWorkbenchHeaderFocus(t *testing.T) {
	m, _, dock, _ := manualDockFixture(t, 7)
	manualDockClick(t, m, 7, "one", ManualWorkbenchTabID("dock", "right"))
	w := m.implementation.find("one").state
	snap, err := desktopSnapshot(w, 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, el := range snap.Elements {
		if el.ID == ManualWorkbenchTabID("dock", "right") {
			found = true
			if !el.Focused || !el.Selected || !dock.buttons["right"].ContainsFocus() {
				t.Fatal("actual focused right header not published")
			}
		}
		if el.ID == ManualWorkbenchTabID("dock", "left") && el.Focused {
			t.Fatal("inactive header falsely focused")
		}
	}
	if !found {
		t.Fatal("focused header missing")
	}
	// Focus moves to the retained panel's input while its tab stays selected.
	manualDockClick(t, m, 7, "one", accessibilityWorkbenchNamespace("dock", "right")+"name")
	snap, err = desktopSnapshot(w, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, el := range snap.Elements {
		if el.ID == ManualWorkbenchTabID("dock", "right") && el.Focused {
			t.Fatal("selected tab inferred as focused")
		}
	}
}
