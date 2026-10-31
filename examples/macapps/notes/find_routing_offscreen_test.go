//go:build offscreen

package notes

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
)

// Use measured managed-widget input, not descriptor callbacks or AX setters.
func TestFindReplacementPointerFocusAndKeyboardUndo(t *testing.T) {
	notesViewport(t, image.Pt(1100, 740), func(model *Model, a *app.Application, m *cogent.Manual) {
		ctx := context.Background()
		must(t, a.Do(func() {
			must(t, model.Update("n1", "Acceptance", "B café 日本語\nalpha beta alpha"))
			must(t, model.openFind(true))
			a.Invalidate("notes")
		}))
		settleViewport(t, m)
		snapshot := func() cogent.ManualWindowSnapshot { s, e := m.Snapshot(ctx); must(t, e); return s.Windows[0] }
		click := func(id string) {
			t.Helper()
			w := snapshot()
			var target cogent.ManualElement
			for _, e := range w.Elements {
				if e.ID == id {
					target = e
					break
				}
			}
			if target.ID == "" || target.Bounds.Empty() {
				t.Fatalf("missing target %s", id)
			}
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				must(t, m.Input(ctx, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: id, TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}))
			}
			settleViewport(t, m)
		}
		replacement := findID + "/find/replacement"
		click(replacement)
		if w := snapshot(); w.FocusedID != replacement {
			t.Fatalf("replacement pointer focus: %q", w.FocusedID)
		}
		must(t, a.Do(func() { model.find.Query = "alpha"; model.find.Replacement = "gamma"; a.Invalidate("notes") }))
		settleViewport(t, m)
		click(findID + "/find/replace-all")
		if got := model.lookup("n1").value.Body; got != "B café 日本語\ngamma beta gamma" {
			t.Fatalf("replacement: %q", got)
		}
		send := func(code key.Codes, r rune, mods key.Modifiers) {
			t.Helper()
			w := snapshot()
			must(t, m.Input(ctx, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Rune: r, Modifiers: mods}))
			settleViewport(t, m)
		}
		send(key.CodeEscape, 0, 0)
		if w := snapshot(); w.FocusedID != "note-editor-n1" {
			t.Fatalf("dismiss focus: %q", w.FocusedID)
		}
		send(key.CodeZ, 'z', 1<<key.Meta)
		if got := model.lookup("n1").value.Body; got != "B café 日本語\nalpha beta alpha" {
			t.Fatalf("one Undo: %q", got)
		}
		send(key.CodeZ, 'z', 1<<key.Meta|1<<key.Shift)
		if got := model.lookup("n1").value.Body; got != "B café 日本語\ngamma beta gamma" {
			t.Fatalf("one Redo: %q", got)
		}
		if e := ui.Validate(model.View()); e != nil {
			t.Fatal(e)
		}
	})
}
