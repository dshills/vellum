package acceptance

import (
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func TestSelectionModelsAndProposalsRemainPerWindow(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	first, err := New(a, "first", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(a, "second", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	view := first.selectionView()
	view.Children[0].Options[0].Label = "Detached description"
	if first.selectionView().Children[0].Options[0].Label != "Rejected first" || second.selectionView().Children[0].Options[0].Label != "Rejected first" {
		t.Fatal("option descriptions were shared")
	}
	for i, id := range []string{"last", "a", "last"} {
		if err = a.Do(func() { first.selectionView().Children[0].Change(ui.ChangeEvent{ID: SelectionListID, SelectedID: id}) }); err != nil {
			t.Fatal(err)
		}
		if first.State().ListSelection != "last" || first.State().Changes[SelectionListID] != i+1 {
			t.Fatal("accept/reject/repeat proposals lost", i, first.State())
		}
	}
	for i, id := range []string{SelectionLeftID, SelectionRightID} {
		var selectionErr error
		if err = a.Do(func() {
			if selectionErr = first.Drawing.Select(id); selectionErr == nil {
				first.selectionView().Children[1].Change(ui.ChangeEvent{ID: SelectionCanvasID, SelectedID: id})
			}
		}); err != nil || selectionErr != nil {
			t.Fatal(err, selectionErr)
		}
		if first.State().DrawingSelection != id || first.State().Changes[SelectionCanvasID] != i+1 {
			t.Fatal("canvas selection proposal lost", first.State())
		}
	}
	if second.State().ListSelection != "b" || second.State().DrawingSelection != "" || len(second.State().Changes) != 0 || first.Drawing == second.Drawing || &first.Drawing.Objects[0] == &second.Drawing.Objects[0] {
		t.Fatal("selection models or counters leaked into another window", second.State())
	}
}
