//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestReviewGridDynamicIDCannotAliasOrdinaryButton(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	collision := GridCellID("g", "row-0", "c0")
	clicked := 0
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
			{ID: "g", Kind: ui.DataGrid, Grid: g, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 240}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 240})}},
			{ID: collision, Kind: ui.Button, Text: "Ordinary button", Activate: func() { clicked++ }},
		}}
	}
	if err := ui.Validate(view()); err != nil {
		t.Log("safe validation refusal", err)
		return
	} // Safe preallocation namespace refusal is allowed.
	m, a := gridManual(t, manualConfig(t, 7), view)
	gridFrame(t, m)
	measured := gridInspection(t, a, g)
	foundCell := false
	for _, cell := range measured.Cells {
		if !cell.Bounds.Empty() {
			foundCell = true
			if err := a.Do(func() { collision = cell.ID; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if !foundCell {
		t.Fatal("fixture has no actual visible grid cell")
	}
	gridFrame(t, m)
	snap, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	matches := 0
	var button ManualElement
	for _, e := range snap.Windows[0].Elements {
		if e.ID == collision {
			matches++
			if e.Kind == ui.Button {
				button = e
			}
		}
	}
	if matches != 1 {
		t.Errorf("managed snapshot has %d indistinguishable target IDs %q", matches, collision)
	}
	if button.ID == "" {
		t.Fatal("ordinary button missing")
	}
	window := snap.Windows[0]
	if err = m.Input(testContext, ManualInput{Window: window.ID, WindowIncarnation: window.Incarnation, Target: button.ID, TargetIncarnation: button.Incarnation, Type: events.MouseDown, Button: events.Left}); err != nil {
		t.Errorf("legitimate captured Button target refused because first same-ID element is a grid cell: %v", err)
	}
	if err = a.Do(func() {
		nodes, _ := accessibilitySnapshot(m.implementation.find("one").state.root, nil)
		ids := map[string]bool{}
		for _, n := range nodes {
			if ids[n.ID] {
				t.Errorf("AX semantic ID duplicated %q (role %s)", n.ID, n.Role)
			}
			ids[n.ID] = true
		}
	}); err != nil {
		t.Fatal(err)
	}
}
