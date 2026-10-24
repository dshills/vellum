package ui

import "testing"

func TestSplitButtonIndependentActionsAndStaleOptions(t *testing.T) {
	primary, secondary := 0, 0
	b := &SplitButton{ID: "split", Label: "Save", Options: []Option{{ID: "as", Label: "Save As…"}}, Primary: func() { primary++ }, Change: func(e ChangeEvent) {
		if e.ID != "split" {
			t.Fatal(e)
		}
		secondary++
	}}
	n, err := b.View()
	if err != nil {
		t.Fatal(err)
	}
	n.Children[0].Activate()
	if primary != 1 || secondary != 0 {
		t.Fatal(primary, secondary)
	}
	n.Children[1].Change(ChangeEvent{SelectedID: "as"})
	if primary != 1 || secondary != 1 {
		t.Fatal(primary, secondary)
	}
	b.Options[0].Disabled = true
	n.Children[1].Change(ChangeEvent{SelectedID: "as"})
	if secondary != 1 {
		t.Fatal("stale menu admitted")
	}
	b.Disabled = true
	n.Children[0].Activate()
	if primary != 1 {
		t.Fatal("disabled primary admitted")
	}
	b.Disabled = false
	b.Small = true
	n, err = b.View()
	if err != nil || n.Children[0].Style.MinSize.Value.Height != 24 {
		t.Fatal(n, err)
	}
	b.ID = "replacement"
	n.Children[0].Activate()
	if primary != 1 {
		t.Fatal("retired identity admitted")
	}
}
