//go:build offscreen

package cogent

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestReviewGridRowDomainCannotReplaceManagedCheckboxCapture(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
			{ID: "g", Kind: ui.DataGrid, Grid: g, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 240}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 240})}},
			{ID: "row-0", Kind: ui.Checkbox, Text: "Ordinary checkbox", Checked: true},
		}}
	}
	if err := ui.Validate(view()); err != nil {
		t.Fatal(err)
	}
	m, a := gridManual(t, manualConfig(t, 7), view)
	gridFrame(t, m)
	measured := gridInspection(t, a, g)
	found := false
	for _, c := range measured.Cells {
		if c.RowKey == "row-0" && !c.Bounds.Empty() {
			found = true
		}
	}
	if !found {
		t.Fatal("fixture has no actual visible row-0 cell")
	}
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range snap.Windows[0].Elements {
		if e.ID == "row-0" && e.Kind == ui.Checkbox {
			if checked, ok := e.Value.(bool); !ok || !checked {
				t.Errorf("ordinary checked checkbox captured foreign row Value %#v", e.Value)
			}
			return
		}
	}
	t.Fatal("ordinary checkbox missing")
}
