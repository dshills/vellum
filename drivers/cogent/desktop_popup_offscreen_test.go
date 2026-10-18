//go:build offscreen

package cogent

import (
	"errors"
	"strings"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestDesktopPopupOutputChargesGeneratedRepresentation(t *testing.T) {
	p := core.PopupInspection{ID: 1, Type: core.MenuStage, Items: []core.PopupItemInspection{{ID: 2, Name: strings.Repeat("x", 1024)}}}
	remaining := 1024 + 256 // enough for the original Core labels/records, not public IDs/records
	if out, err := desktopPopupInspection(p, 1, &remaining); !errors.Is(err, ErrDesktopBudget) || out.ID != "" {
		t.Fatal("unaccounted generated popup output", out, err)
	}
	remaining = 8 << 20
	out, err := desktopPopupInspection(p, 1, &remaining)
	if err != nil || len(out.Items) != 1 || remaining >= ((8<<20)-1024-1024) {
		t.Fatal("public popup strings/records not charged", remaining, err)
	}
}

// This exercises the actual software FIFO/widget route. desktopFrame remains an
// explicit test-only capability; it does not qualify a native composition.
func TestDesktopActualPopupTargetsCommitCancelAndRevoke(t *testing.T) {
	selected, calls := "a", 0
	s, m, _, w := desktopViewFixture(t, func() ui.Node {
		return ui.Node{ID: "selector", Kind: ui.Selector, SelectedID: selected, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "c", Label: "Charlie"}}, Change: func(e ui.ChangeEvent) { selected = e.SelectedID; calls++ }}
	})
	ready := func() DesktopSnapshot { desktopFrame(t, s, m, w); return desktopRead(t, s, m, w) }
	send := func(snap DesktopSnapshot, el DesktopElement, typ events.Types) error {
		in := DesktopInput{Frame: snap.Frame, Target: el.ID, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}
		r, err := s.submit(testContext, w.window.Ref(), 1, in)
		if err != nil {
			return err
		}
		return desktopPump(t, s, m, r).err
	}
	chooser := func(snap DesktopSnapshot) DesktopElement {
		for _, el := range snap.Elements {
			if el.DomainID == "selector" {
				return el
			}
		}
		t.Fatal("chooser absent")
		return DesktopElement{}
	}
	item := func(snap DesktopSnapshot, name string) DesktopElement {
		if len(snap.Popups) != 1 || !snap.Popups[0].Interactive {
			t.Fatal("actual popup missing", snap.Popups)
		}
		for _, el := range snap.Popups[0].Items {
			if el.Name == name {
				return el
			}
		}
		t.Fatal("actual item absent", name)
		return DesktopElement{}
	}
	open := func() {
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			snap := ready()
			if err := send(snap, chooser(snap), typ); err != nil {
				t.Fatal(err)
			}
		}
	}
	if chooser(ready()).Focused {
		t.Fatal("fixture must start without selector focus")
	}
	open()
	snap := ready()
	if len(snap.Popups[0].Items) != 2 {
		t.Fatal("disabled option entered actual popup", snap.Popups)
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		snap = ready()
		if err := send(snap, item(snap, "Charlie"), typ); err != nil {
			t.Fatal(err)
		}
	}
	snap = ready()
	if selected != "c" || calls != 1 || len(snap.Popups) != 0 {
		t.Fatal("placement duplicated or popup retained", selected, calls, snap.Popups)
	}

	if !chooser(snap).Focused || !findRecord(w.root, "selector").widget.AsWidget().ContainsFocus() {
		t.Fatal("pointer selection failed to retain actual chooser keyboard focus")
	}
	// The subsequent key targets the actual retained focus owner, not a forced
	// fixture focus request or a snapshot-only override.
	r, err := s.submit(testContext, w.window.Ref(), 1, DesktopInput{Frame: snap.Frame, Target: chooser(snap).ID, Type: events.KeyChord, Code: key.CodeReturnEnter})
	if err != nil {
		t.Fatal(err)
	}
	if reply := desktopPump(t, s, m, r); reply.err != nil {
		t.Fatal(reply.err)
	}
	snap = ready()
	_ = item(snap, "Alpha")
	sourceForCancel := w.body.Scene.RenderWindow().SystemWindow.Events()
	sourceForCancel.Key(events.KeyChord, 0, key.CodeEscape, 0)
	unitSettle(t, m)
	snap = ready()
	if len(snap.Popups) != 0 || !chooser(snap).Focused || calls != 1 {
		t.Fatal("keyboard reopen/Escape lost pointer focus or changed selection")
	}
	open()
	snap = ready()
	old := item(snap, "Alpha")
	// External ordinary input closes/reopens the popup without asking the test
	// session to issue another frame capability. A recycled Scene/index must not
	// authorize the old item even when Frame itself is unchanged.
	source := w.body.Scene.RenderWindow().SystemWindow.Events()
	source.Key(events.KeyChord, 0, key.CodeEscape, 0)
	unitSettle(t, m)
	ch := findRecord(w.root, "selector").widget.(*core.Chooser)
	ch.AsyncLock()
	ch.Send(events.Click)
	ch.AsyncUnlock()
	unitSettle(t, m)
	s.mu.Lock()
	binding := s.windows[w.window.Ref()]
	sameFrame := binding.frame == snap.Frame && binding.revision == binding.completedRevision
	s.mu.Unlock()
	if !sameFrame {
		t.Fatal("fixture did not preserve the same frame for incarnation check")
	}
	if err := send(snap, old, events.MouseDown); !errors.Is(err, ErrDesktopPending) {
		t.Fatal("stale same-frame popup target accepted", err)
	}
	snap = ready()
	if item(snap, "Alpha").ID == old.ID {
		t.Fatal("popup incarnation revived")
	}
	if calls != 1 || selected != "c" {
		t.Fatal("cancel or stale target changed model")
	}
	// Caller mutation cannot alter current popup labels/geometry.
	snap.Popups[0].Items[0].Name = "mutated"
	snap.Popups[0].Items[0].Bounds = old.Bounds
	if item(ready(), "Alpha").Name != "Alpha" {
		t.Fatal("borrowed popup values")
	}
}
