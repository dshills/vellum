package sdkcontract_test

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

// An independent application consumes only exported root-module APIs.
func TestPublicListBoxStableSelectionContract(t *testing.T) {
	selected := ""
	reject := false
	requests := 0
	view := func() ui.Node {
		return ui.Node{ID: "layers", Kind: ui.ListBox, Options: []ui.Option{{ID: "rear", Label: "Background"}, {ID: "guide", Label: "Guides", Disabled: true}, {ID: "front", Label: "Foreground"}}, SelectedID: selected, Change: func(e ui.ChangeEvent) {
			requests++
			if !reject {
				selected = e.SelectedID
			}
		}}
	}
	n := view()
	if ui.MaxListBoxOptions != 256 {
		t.Fatal("unexpected public bound")
	}
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	target, ok := ui.MoveListBoxOption(n, "rear", 1)
	if !ok || target != "front" || selected != "" {
		t.Fatal("navigation mutated model or selected disabled row")
	}
	n.Change(ui.ChangeEvent{ID: n.ID, Kind: ui.ListBox, SelectedID: target})
	if view().SelectedID != "front" {
		t.Fatal("accepted edit")
	}
	reject = true
	view().Change(ui.ChangeEvent{ID: n.ID, Kind: ui.ListBox, SelectedID: "rear"})
	if view().SelectedID != "front" || requests != 2 {
		t.Fatal("rejected edit")
	}
	n = view()
	n.Options = []ui.Option{n.Options[2], n.Options[0]}
	if err := ui.Validate(n); err != nil {
		t.Fatal("stable selection failed reorder", err)
	}
}
