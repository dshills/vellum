package main

import (
	"maps"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func TestMarqueeSelectionSubtractsWholeFlatUnitsInPaintOrder(t *testing.T) {
	objects := []canvas.Rect{
		{ID: "a", Width: 80, Height: 50},
		{ID: "b", X: 200, Width: 120, Height: 16},
		{ID: "c", X: 400, Width: 90, Height: 65},
		{ID: "d", X: 600, Width: 20, Height: 20},
	}
	groups := map[string]string{"a": "assembly", "c": "assembly"}
	prior := []string{"c", "missing", "b", "a", "c"}
	before, membership, priorCopy := slices.Clone(objects), maps.Clone(groups), slices.Clone(prior)
	// A one-unit overlap with a removes its distant/interleaved group member c.
	bounds := canvas.Rect{X: 79, Width: 21, Height: 1}
	if got := marqueeSelection(objects, groups, prior, bounds, true); !slices.Equal(got, []string{"b"}) {
		t.Fatal("partial contact split group or changed paint order", got)
	}
	if got := marqueeSelection(objects, groups, []string{"b"}, bounds, false); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatal("additive contact must also expand whole group", got)
	}
	for _, bounds := range []canvas.Rect{
		{X: 80, Width: 120, Height: 1}, // Edge-only contacts with a and b.
		{Width: 80, Height: 0},
		{Width: 0, Height: 50},
		{X: 900, Width: 10, Height: 10},
	} {
		if got := marqueeSelection(objects, groups, prior, bounds, true); !slices.Equal(got, []string{"a", "b", "c"}) {
			t.Fatal("empty/edge contact changed selection", bounds, got)
		}
	}
	if got := marqueeSelection(objects, groups, nil, bounds, true); len(got) != 0 {
		t.Fatal("subtraction added an unselected object", got)
	}
	if !slices.Equal(objects, before) || !maps.Equal(groups, membership) || !slices.Equal(prior, priorCopy) {
		t.Fatal("selection helper mutated its inputs")
	}
}

func TestSubtractMarqueeTransformsFrozenModePrimaryAndHistory(t *testing.T) {
	for _, direction := range []struct{ start, end canvas.Point }{
		{canvas.Point{X: -20, Y: -20}, canvas.Point{X: 100, Y: 70}},
		{canvas.Point{X: 100, Y: 70}, canvas.Point{X: -20, Y: -20}},
		{canvas.Point{X: 100, Y: -20}, canvas.Point{X: -20, Y: 70}},
		{canvas.Point{X: -20, Y: 70}, canvas.Point{X: 100, Y: -20}},
	} {
		for _, primary := range []string{"connector-2", "note-3"} {
			m := groupedModel(t)
			m.setSelection([]string{"block-1", "connector-2"}, primary)
			m.selectionSubtract = true
			_ = m.journal.MarkSaved()
			before, groups, state, next := slices.Clone(m.scene.Objects), maps.Clone(m.groups), m.journal.Snapshot(), m.next
			g := groupGeometry() // Nonzero origin, pan, zoom and 2x display scale.
			start, end := selectionPoint(g, direction.start), selectionPoint(g, direction.end)
			s := m.selectionToggle(g, start)
			if s == nil || !s.marquee || !s.marqueeSubtract || !maps.Equal(s.marqueeGroups, groups) {
				t.Fatal("subtractive marquee snapshot missing")
			}
			view := m.scene.View
			m.selectionSubtract = false // A gesture retains its captured operation.
			m.previewMove(s, end)
			if !s.dragging || s.marqueeBounds != marqueeRect(direction.start, direction.end) || len(m.selectedIDs()) != 3 {
				t.Fatal("transformed preview changed selection or mapping")
			}
			if err := m.finishMove(s, end); err != nil {
				t.Fatal(err)
			}
			_ = m.finishMove(s, end)
			if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) || m.scene.Selected != "connector-2" || m.move != nil || m.scene.View != view || m.next != next || !slices.Equal(m.scene.Objects, before) || !maps.Equal(m.groups, groups) || m.journal.Snapshot() != state {
				t.Fatal("subtraction changed document/history or retained a removed primary", m.selectedIDs())
			}
		}
	}
}

func TestSubtractShiftClickRemovesWholeGroupWithoutAdding(t *testing.T) {
	m := groupedModel(t)
	m.selectionSubtract = true
	g := groupGeometry()
	m.setSelection([]string{"block-1", "connector-2"}, "note-3")
	_ = m.journal.MarkSaved()
	state := m.journal.Snapshot()
	toggleSelection(t, m, g, canvas.Point{X: 40, Y: 25})
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) || m.scene.Selected != "connector-2" {
		t.Fatal("subtract click split group or kept removed primary")
	}
	toggleSelection(t, m, g, canvas.Point{X: 440, Y: 30})
	if !slices.Equal(m.selectedIDs(), []string{"connector-2"}) || m.scene.Selected != "connector-2" {
		t.Fatal("subtract click added an unselected group")
	}
	toggleSelection(t, m, g, canvas.Point{X: 240, Y: 8})
	if len(m.selectedIDs()) != 0 || m.scene.Selected != "" || m.journal.Snapshot() != state {
		t.Fatal("subtract click did not clear last member or changed history")
	}
	// Add remains the default and retains the existing toggle behavior.
	m.selectionSubtract = false
	toggleSelection(t, m, g, canvas.Point{X: 40, Y: 25})
	if !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) {
		t.Fatal("additive mode no longer expands/toggles groups")
	}
}

func TestSubtractMarqueeCancelThresholdNoopAndProtectedObjects(t *testing.T) {
	for _, how := range []string{"threshold", "empty", "zero-area", "escape", "capture-loss", "outside", "camera", "scale", "window", "close"} {
		t.Run(how, func(t *testing.T) {
			m := groupedModel(t)
			m.selectionSubtract = true
			m.setSelection([]string{"block-1", "connector-2"}, "note-3")
			_ = m.journal.MarkSaved()
			before, ids, state := slices.Clone(m.scene.Objects), m.selectedIDs(), m.journal.Snapshot()
			g := groupGeometry()
			p := selectionPoint(g, canvas.Point{X: -20, Y: -20})
			s := m.selectionToggle(g, p)
			end := selectionPoint(g, canvas.Point{X: 100, Y: 70})
			if s == nil {
				t.Fatal("begin")
			}
			switch how {
			case "threshold":
				_ = m.finishMove(s, canvas.Point{X: p.X + 3*g.Scale, Y: p.Y})
			case "empty":
				_ = m.finishMove(s, selectionPoint(g, canvas.Point{X: -100, Y: -100}))
			case "zero-area":
				_ = m.finishMove(s, selectionPoint(g, canvas.Point{X: 100, Y: -20}))
			case "outside":
				m.previewMove(s, end)
				_ = m.finishMove(s, canvas.Point{X: -1, Y: -1})
			case "camera", "scale", "window":
				m.previewMove(s, end)
				if how == "camera" {
					g.View.Center.X++
				} else if how == "scale" {
					g.Scale++
				} else {
					g.Window = "other"
				}
				m.updateMoveGeometry(s, g)
			case "close":
				if !m.requestClose(nil) {
					t.Fatal("clean subtract gesture blocked close")
				}
			default:
				m.previewMove(s, end)
				m.cancelMove(s)
			}
			_ = m.finishMove(s, end)
			if m.move != nil || !slices.Equal(m.selectedIDs(), ids) || m.scene.Selected != "note-3" || !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot() != state {
				t.Fatal("canceled/no-op subtraction changed selection/document/history", how)
			}
		})
	}
	for _, subtract := range []bool{false, true} {
		m := editModel(t)
		m.layerInfo = map[string]layerMetadata{"block-1": {Hidden: true}, "note-3": {Locked: true}}
		m.setSelection([]string{"block-1", "connector-2", "note-3"}, "connector-2")
		m.selectionSubtract = subtract
		g := groupGeometry()
		s := m.selectionToggle(g, selectionPoint(g, canvas.Point{X: -20, Y: -20}))
		_ = m.finishMove(s, selectionPoint(g, canvas.Point{X: 500, Y: 80}))
		want := []string{"connector-2"}
		if subtract {
			want = nil
		}
		if !slices.Equal(m.selectedIDs(), want) {
			t.Fatal("hidden/locked objects entered selection", subtract, m.selectedIDs())
		}
		if m.beginMove(g, selectionPoint(g, canvas.Point{X: -20, Y: -20})) != nil {
			t.Fatal("selection mode captured ordinary empty pan")
		}
	}
}
