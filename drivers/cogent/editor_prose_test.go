package cogent

import (
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/text/textcore"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestManagedProseRetainsBindingCompositionAndHistory(t *testing.T) {
	for _, prose := range []bool{false, true} {
		t.Run(map[bool]string{false: "code", true: "prose"}[prose], func(t *testing.T) {
			a, _, tw, r, model := editorFixture(t, "Draft é\r\n")
			widget := r.widget
			n := r.node
			n.Prose = prose
			n.SoftWrap = true
			if err := syncRecord(r, n); err != nil {
				t.Fatal(err)
			}
			widget.(*textcore.Editor).Style()
			initial := model.Snapshot().Text
			_ = a.Do(func() {
				tw.apply(fixtureInput(r, 1, "mark", initial, "に"+initial, 1, 1, 0, 1))
				tw.apply(fixtureInput(r, 2, "unmark", model.Snapshot().VisibleText, model.Snapshot().VisibleText, 1, 1, -1, -1))
			})
			if r.widget != widget || model.Snapshot().Composing || model.Snapshot().Text != "に"+initial || model.Snapshot().History.UndoCount != 1 {
				t.Fatal("presentation broke native composition or binding")
			}
			_ = a.Do(func() {
				if err := model.Journal().Undo(); err != nil {
					t.Fatal(err)
				}
				syncEditor(r)
			})
			if model.Snapshot().Text != initial {
				t.Fatal("undo changed source bytes")
			}
			_ = a.Do(func() {
				if err := model.Journal().Redo(); err != nil {
					t.Fatal(err)
				}
				syncEditor(r)
			})
			if model.Snapshot().Text != "に"+initial {
				t.Fatal("redo lost committed edit")
			}
			n.Prose = !prose
			if err := syncRecord(r, n); err != nil {
				t.Fatal(err)
			}
			widget.(*textcore.Editor).Style()
			if r.widget != widget || model.Snapshot().Text != "に"+initial {
				t.Fatal("mode transition replaced editor")
			}
		})
	}
}

func TestManagedProseFramelessFocusPreservesOtherIndicators(t *testing.T) {
	_, _, _, r, _ := editorFixture(t, "Draft")
	widget := r.widget
	for _, prose := range []bool{false, true, false, true} {
		n := r.node
		n.Prose = prose
		n.SoftWrap = true
		n.Style.Border = ui.Some(ui.Border{})
		if err := syncRecord(r, n); err != nil {
			t.Fatal(err)
		}
		for _, focused := range []bool{true, false, true} {
			widget.AsWidget().SetState(focused, states.Focused)
			widget.(*textcore.Editor).Style()
			width := widget.AsWidget().Styles.Border.Width.Top.Value
			if prose && width != 0 {
				t.Fatal("frameless prose gained an outer focus box", width)
			}
			if !prose && focused && width <= 0 {
				t.Fatal("code editor lost native focus indicator")
			}
			if r.widget != widget {
				t.Fatal("focus style replaced editor")
			}
		}
	}
}
