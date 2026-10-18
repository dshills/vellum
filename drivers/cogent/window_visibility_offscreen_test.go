//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/system/composer"
	"github.com/dshills/vellum/ui"
)

type visibilityDrawer struct {
	composer.Drawer
	calls int
}

func (d *visibilityDrawer) End() { d.calls++; d.Drawer.End() }

func TestWindowVisibilityOwnerRedrawCleanSceneAndPopup(t *testing.T) {
	for _, popup := range []bool{false, true} {
		t.Run(map[bool]string{false: "clean", true: "popup"}[popup], func(t *testing.T) {
			m, _ := simpleManual(t, manualConfig(t, 91), func() ui.Node {
				return ui.Node{ID: "selector", Kind: ui.Selector, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}}
			})
			unitSettle(t, m)
			managed := m.implementation.find("one")
			w := managed.state
			bindWindowOriginCustom(w, nil, func(*nativeWindow, events.Event) { t.Error("visibility drained unrelated owner work") })
			if popup {
				ch := findRecord(w.root, "selector").widget.(*core.Chooser)
				ch.AsyncLock()
				ch.Send(events.Click)
				ch.AsyncUnlock()
				unitSettle(t, m)
			}
			cd := w.body.Scene.RenderWindow().SystemWindow.Composer().(*composer.ComposerDrawer)
			old := cd.Drawer
			d := &visibilityDrawer{Drawer: old}
			cd.Drawer = d
			t.Cleanup(func() { cd.Drawer = old })
			pump := func() {
				t.Helper()
				if _, err := managed.core.Step(testContext, 128); err != nil {
					t.Fatal(err)
				}
			}
			pump()
			before := d.calls
			v := newWindowVisibility()
			w.visibility.Store(v)
			t.Cleanup(v.close)
			v.observe(false)
			v.observe(true)
			if err := v.bind(w.body.SystemEvents()); err != nil {
				t.Fatal(err)
			}
			// No Manual.Step owner drain or explicit NeedsRender can mask this path.
			pump()
			if d.calls != before+1 {
				t.Fatalf("clean scene not recomposed exactly once: %d -> %d", before, d.calls)
			}
			before = d.calls
			v.observe(true)
			pump()
			if d.calls != before {
				t.Fatal("duplicate notification redrew")
			}
			if popup {
				w.body.AsyncLock()
				items, err := w.body.Scene.InspectPopups(8192, 8<<20)
				w.body.AsyncUnlock()
				if err != nil || len(items) != 1 {
					t.Fatal("visibility dismissed popup", items, err)
				}
			}
			// A foreign/stale state cannot invalidate this scene or clear its own node.
			foreign := newWindowVisibility()
			t.Cleanup(foreign.close)
			_ = foreign.bind(w.body.SystemEvents())
			foreign.observe(true)
			pump()
			if d.calls != before || foreign.pending == nil {
				t.Fatal("foreign payload accepted")
			}
			v.observe(false)
			v.observe(true)
			v.close()
			pump()
			if d.calls != before {
				t.Fatal("retired state redrew")
			}
		})
	}
}

func TestDesktopVisibilityWithoutFrameAndCancellation(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	read := func() desktopReply {
		t.Helper()
		r, err := s.submit(testContext, w.window.Ref(), 4, DesktopInput{})
		if err != nil {
			t.Fatal(err)
		}
		return desktopPump(t, s, m, r)
	}
	if out := read(); out.err != nil || out.visibility.Known {
		t.Fatal(out)
	}
	v := newWindowVisibility()
	w.visibility.Store(v)
	t.Cleanup(v.close)
	v.observe(false)
	if out := read(); out.err != nil || out.visibility != (DesktopWindowVisibility{true, false, 1}) {
		t.Fatal(out)
	}
	b := s.windows[w.window.Ref()]
	if b.frame != 0 || b.trace.Pending() {
		t.Fatal("inspection armed frame")
	}
	desktopFrame(t, s, m, w)
	old := b.frame
	if out := read(); out.err != nil || b.frame != old {
		t.Fatal("inspection changed frame", out.err)
	}
	ctx, cancel := context.WithCancel(testContext)
	r, err := s.submit(ctx, w.window.Ref(), 4, DesktopInput{})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	s.cancel(r)
	if out := <-r.reply; !errors.Is(out.err, context.Canceled) {
		t.Fatal(out.err)
	}
	unitSettle(t, m)
	if s.pending != nil || b.frame != old {
		t.Fatal("cancel inspection leaked or changed frame")
	}
	v.close()
	if out := read(); !errors.Is(out.err, ErrDesktopClosed) {
		t.Fatal(out.err)
	}
}

func TestDesktopVisibilityPublicInspection(t *testing.T) {
	s, m, _, w, _ := desktopFixture(t)
	v := newWindowVisibility()
	w.visibility.Store(v)
	t.Cleanup(v.close)
	v.observe(true)
	ctx, cancel := context.WithTimeout(testContext, 3*time.Second)
	defer cancel()
	done := make(chan desktopReply, 1)
	go func() {
		value, err := s.Visibility(ctx, w.window.Ref())
		done <- desktopReply{visibility: value, err: err}
	}()
	for {
		select {
		case out := <-done:
			if out.err != nil || out.visibility != (DesktopWindowVisibility{true, true, 1}) || s.State().Pending {
				t.Fatal(out, s.State())
			}
			return
		case <-ctx.Done():
			t.Fatal("public visibility did not terminate", ctx.Err())
		default:
			if _, err := m.Step(ctx, 7); err != nil {
				t.Fatal(err)
			}
			runtime.Gosched()
		}
	}
}
