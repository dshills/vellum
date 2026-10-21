package main

import (
	"github.com/dshills/vellum/canvas"
	"maps"
	"math"
	"slices"
	"testing"
)

func TestGeometryInspectorTransformsGroupsInOneHistoryEntry(t *testing.T) {
	m := groupedModel(t)
	before := slices.Clone(m.scene.Objects)
	groups := maps.Clone(m.groups)
	ids := m.selectedIDs()
	bounds, _ := selectionBounds(before, ids)
	n := m.journal.Snapshot().UndoCount
	if err := m.beginGeometry(); err != nil {
		t.Fatal(err)
	}
	d := m.geometry
	d.values = [4]string{"100.125", "-50.75", "400", "300"}
	m.prompt = ""
	if err := m.applyGeometry(d); err != nil {
		t.Fatal(err)
	}
	after := slices.Clone(m.scene.Objects)
	got, _ := selectionBounds(after, ids)
	if math.Abs(got.X-100.125) > 1e-9 || math.Abs(got.Y+50.75) > 1e-9 || math.Abs(got.Width-400) > 1e-9 || math.Abs(got.Height-300) > 1e-9 || !maps.Equal(m.groups, groups) || m.journal.Snapshot().UndoCount != n+1 {
		t.Fatal("geometry", bounds, got)
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("Undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
		t.Fatal("Redo", err)
	}
}
func TestGeometryInspectorRejectsInvalidOrStaleDraftAtomically(t *testing.T) {
	for _, v := range []string{"NaN", "+Inf", "1e999", "bad", "0", "-1", "0.0001", "1e100"} {
		t.Run(v, func(t *testing.T) {
			m := groupedModel(t)
			before := slices.Clone(m.scene.Objects)
			state := m.journal.Snapshot()
			if err := m.beginGeometry(); err != nil {
				t.Fatal(err)
			}
			d := m.geometry
			d.values[2] = v
			m.prompt = ""
			if m.applyGeometry(d) == nil || !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot() != state {
				t.Fatal("invalid edit mutated")
			}
		})
	}
	m := groupedModel(t)
	_ = m.beginGeometry()
	d := m.geometry
	m.prompt = ""
	m.scene.Objects[0].X++
	if m.applyGeometry(d) == nil {
		t.Fatal("stale draft")
	}
}
func TestGeometryInspectorNoopDoesNotAddHistory(t *testing.T) {
	m := groupedModel(t)
	n := m.journal.Snapshot().UndoCount
	_ = m.beginGeometry()
	d := m.geometry
	m.prompt = ""
	if err := m.applyGeometry(d); err != nil || m.journal.Snapshot().UndoCount != n {
		t.Fatal("no-op", err)
	}
}
func TestGeometryTransformPreservesUnselectedObjectsAndRejectsAllOnLimits(t *testing.T) {
	before := []canvas.Rect{{ID: "a", X: 10, Y: 20, Width: 80, Height: 40}, {ID: "b", X: 200, Y: 30, Width: 40, Height: 30}}
	after, err := transformedSelection(before, []string{"a"}, canvas.Rect{X: 1.25, Y: 2.5, Width: 100, Height: 50})
	if err != nil || after[1] != before[1] || after[0].ID != "a" {
		t.Fatal(err, after)
	}
	if _, err = transformedSelection(before, []string{"a", "b"}, canvas.Rect{X: 9999, Y: 0, Width: 300, Height: 80}); err == nil {
		t.Fatal("origin constraint")
	}
}

func TestGeometryExactSingletonLimitsAndFractionalNoop(t *testing.T) {
	for _, c := range []struct{ before, target float64 }{{49, 8}, {15, 2000}} {
		before := []canvas.Rect{{ID: "a", X: 1.25, Y: -3.75, Width: c.before, Height: 40}}
		after, err := transformedSelection(before, []string{"a"}, canvas.Rect{X: 1.25, Y: -3.75, Width: c.target, Height: 40})
		if err != nil || after[0].Width != c.target || after[0].Height != 40 {
			t.Fatal("exact singleton limit", c, after, err)
		}
	}
	before := []canvas.Rect{{ID: "a", X: -9999.1, Y: .1, Width: 80, Height: 40}, {ID: "b", X: 9999.3, Y: .3, Width: 40, Height: 30}}
	bounds, _ := selectionBounds(before, []string{"a", "b"})
	after, err := transformedSelection(before, []string{"a", "b"}, bounds)
	if err != nil || !slices.Equal(before, after) {
		t.Fatal("fractional unchanged geometry drift", err)
	}
}
