//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestListBoxExplicitSelectedPointerProposesOnce(t *testing.T) {
	selected := "a"
	var proposals []string
	m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		n := listBoxDescriptor(func(e ui.ChangeEvent) { proposals = append(proposals, e.SelectedID); selected = e.SelectedID })
		n.SelectedID = selected
		return n
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	click := func(id string) {
		t.Helper()
		r := w.state.root
		row := r.control.choice.buttons[id]
		box := row.Geom.TotalBBox
		if box.Empty() || !box.In(r.widget.AsWidget().Geom.ContentBBox) {
			t.Fatal("row not actually visible", id, box)
		}
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: box.Min.Add(box.Size().Div(2))}); e != nil {
				t.Fatal(e)
			}
			unitSettle(t, m)
		}
	}
	click("a")
	if selected != "a" || len(proposals) != 1 || proposals[0] != "a" {
		t.Fatal("explicit accepted-row click lost or duplicated", selected, proposals)
	}
	click("b")
	if selected != "b" || len(proposals) != 2 || proposals[1] != "b" {
		t.Fatal("ordinary changed-row click lost or duplicated", selected, proposals)
	}
}
