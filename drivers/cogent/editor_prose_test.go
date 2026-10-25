package cogent

import (
	"cogentcore.org/core/text/textcore"
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
