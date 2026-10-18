//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/ui"
	"testing"
)

// Only core.Step pumps the critical wake. Manual.Step independently drains the
// app owner and would mask this native modal routing failure.
func TestOriginCustomWakeDrainsOwnerThroughActualPopup(t *testing.T) {
	for _, withSession := range []bool{false, true} {
		name := "ordinary"
		if withSession {
			name = "session"
		}
		t.Run(name, func(t *testing.T) {
			m, a := simpleManual(t, manualConfig(t, 712), func() ui.Node {
				return ui.Node{ID: "selector", Kind: ui.Selector, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, Change: func(ui.ChangeEvent) { t.Error("wake changed selection") }}
			})
			unitSettle(t, m)
			managed := m.implementation.find("one")
			w := managed.state
			var session *DesktopSession
			if withSession {
				session = NewDesktopSession()
				if err := session.start(a); err != nil {
					t.Fatal(err)
				}
				if err := session.bind(w); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { session.retire(w); session.close() })
			}
			// Actual Chooser construction/opening is setup; the critical wake below is
			// real posted work and real FIFO routing, with no outer application drain.
			ch := findRecord(w.root, "selector").widget.(*core.Chooser)
			ch.AsyncLock()
			ch.Send(events.Click)
			ch.AsyncUnlock()
			unitSettle(t, m)
			inspect := func() []core.PopupInspection {
				w.body.AsyncLock()
				defer w.body.AsyncUnlock()
				p, err := w.body.Scene.InspectPopups(8192, 8<<20)
				if err != nil {
					t.Fatal(err)
				}
				return p
			}
			before := inspect()
			if len(before) != 1 || len(before[0].Items) != 2 {
				t.Fatal("actual chooser menu missing", before)
			}
			bindWindowOriginCustom(w, session, func(n *nativeWindow, e events.Event) {
				n.pending.Store(false)
				a.Drain()
				if session != nil {
					session.handle(n, e)
				}
				e.SetHandled()
			})
			pump := func() {
				if _, err := managed.core.Step(testContext, 256); err != nil {
					t.Fatal(err)
				}
			}
			calls := 0
			if err := a.PostWindow(w.window.Ref(), func() { calls++ }); err != nil {
				t.Fatal(err)
			}
			pump()
			if calls != 1 {
				t.Fatalf("posted owner callback starved behind actual chooser popup: calls=%d", calls)
			}
			after := inspect()
			if len(after) != 1 || after[0].ID != before[0].ID {
				t.Fatal("wake dismissed/replaced actual popup")
			}
			pump()
			if calls != 1 {
				t.Fatal("wake duplicated callback")
			}

			// An unrelated Custom must still reach actual popup routing.
			popup := core.NewScene("unrelated-custom")
			unrelated := 0
			popup.OnFirst(events.Custom, func(e events.Event) { unrelated++; e.SetHandled() })
			core.NewButton(popup).SetText("Other")
			w.body.AsyncLock()
			core.NewPopupStage(core.MenuStage, popup, w.body).Run()
			w.body.AsyncUnlock()
			unitSettle(t, m)
			source := w.source.Load()
			source.Custom("application.custom")
			pump()
			if unrelated != 1 {
				t.Fatal("unrelated custom intercepted", unrelated)
			}
			// Retirement revokes the origin hook; a queued wake is then routed to the
			// popup without retaining authority to drain owner work.
			w.body.Scene.SetOriginCustomHandler(nil)
			if err := a.PostWindow(w.window.Ref(), func() { calls++ }); err != nil {
				t.Fatal(err)
			}
			pump()
			if calls != 1 || unrelated != 2 {
				t.Fatal("cleared hook still drained or stole custom", calls, unrelated)
			}
			a.Drain() // explicit test cleanup of the deliberately undelivered callback
		})
	}
}
