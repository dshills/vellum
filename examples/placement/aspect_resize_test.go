package main

import (
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func aspectClose(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}

func TestAspectAllCornersShareScaleAndOppositeAnchor(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		before := []canvas.Rect{{ID: "a", X: 10, Y: 20, Width: 80, Height: 50}, {ID: "b", X: 210, Y: 70, Width: 40, Height: 100}, {ID: "other", X: -500, Width: 20, Height: 20}}
		ids := []string{"a"}
		if multiple {
			ids = append(ids, "b")
		}
		bounds, _ := selectionBounds(before, ids)
		for _, corner := range []resizeCorner{northWest, northEast, southEast, southWest} {
			dx, dy := bounds.Width*.5, bounds.Height*.1
			ax, ay := bounds.X, bounds.Y
			if corner == northWest || corner == southWest {
				dx = -dx
				ax += bounds.Width
			}
			if corner == northWest || corner == northEast {
				dy = -dy
				ay += bounds.Height
			}
			after, err := resizedSelection(before, ids, corner, dx, dy, false, true)
			if err != nil {
				t.Fatal(multiple, corner, err)
			}
			for i, o := range before {
				if !slices.Contains(ids, o.ID) {
					if after[i] != o {
						t.Fatal("unselected member changed")
					}
					continue
				}
				if !aspectClose(after[i].X, ax+(o.X-ax)*1.5) || !aspectClose(after[i].Y, ay+(o.Y-ay)*1.5) || !aspectClose(after[i].Width, o.Width*1.5) || !aspectClose(after[i].Height, o.Height*1.5) || after[i].ID != o.ID || after[i].Color != o.Color {
					t.Fatal("nonuniform transform/anchor", multiple, corner, after[i], o)
				}
			}
			if !multiple {
				one := resizedCorner(before[0], corner, dx, dy, false, true)
				if one != after[0] {
					t.Fatal("single and selection aspect paths differ", one, after[0])
				}
			}
		}
		if before[0].Width != 80 || before[1].X != 210 {
			t.Fatal("detached transform mutated original")
		}
	}
}

func TestAspectSnapUsesOnlyDrivingWorldDimension(t *testing.T) {
	one := canvas.Rect{ID: "a", X: 10, Y: 20, Width: 80, Height: 50}
	width := resizedCorner(one, southEast, 23, 1, true, true)
	if width.Width != 100 || width.Height != 62.5 || width.X != 10 || width.Y != 20 {
		t.Fatal("width snapped independently from height", width)
	}
	height := resizedCorner(one, northWest, -1, -13, true, true)
	if height.Height != 60 || height.Width != 96 || !aspectClose(height.X, -6) || !aspectClose(height.Y, 10) {
		t.Fatal("height driver/opposite anchor", height)
	}
	before := []canvas.Rect{one, {ID: "b", X: 210, Y: 70, Width: 40, Height: 100}}
	after, err := resizedSelection(before, []string{"a", "b"}, southEast, 23, 1, true, true)
	if err != nil {
		t.Fatal(err)
	}
	bounds, _ := selectionBounds(after, []string{"a", "b"})
	if !aspectClose(bounds.Width, 260) || !aspectClose(bounds.Height, 162.5) || !aspectClose(after[1].X, 10+200*260.0/240) {
		t.Fatal("group driver snap/spacing", bounds, after)
	}
	// A pure handle click and an unchanged inspector value do not snap a shape.
	for _, snap := range []bool{false, true} {
		got, err := resizedSelection(before, []string{"a", "b"}, southEast, 0, 0, snap, true)
		if err != nil || !slices.Equal(got, before) {
			t.Fatal("zero motion changed geometry", got, err)
		}
		got, err = resizedSelectionDimension(before, []string{"a", "b"}, "width", 240, snap, true)
		if err != nil || !slices.Equal(got, before) {
			t.Fatal("same dimension changed geometry", got, err)
		}
	}
}

func TestAspectStrictestMemberLimitsAndNoFlips(t *testing.T) {
	before := []canvas.Rect{{ID: "a", X: 10, Y: 20, Width: 80, Height: 50}, {ID: "b", X: 210, Y: 70, Width: 40, Height: 100}}
	ids := []string{"a", "b"}
	for _, corner := range []resizeCorner{northWest, northEast, southEast, southWest} {
		dx, dy := -1e6, -1e6
		if corner == northWest || corner == southWest {
			dx = -dx
		}
		if corner == northWest || corner == northEast {
			dy = -dy
		}
		after, err := resizedSelection(before, ids, corner, dx, dy, false, true)
		if err != nil || after[0].Width != 16 || after[0].Height != 10 || after[1].Width != 8 || after[1].Height != 20 {
			t.Fatal("minimum shared scale/crossing", corner, after, err)
		}
	}
	after, err := resizedSelection(before, ids, southEast, 1e6, 1e6, true, true)
	if err != nil || after[0].Width != 1600 || after[0].Height != 1000 || after[1].Width != 800 || after[1].Height != 2000 {
		t.Fatal("maximum shared scale", after, err)
	}
}

func TestAspectOriginLimitsConstrainCommonScaleAtomically(t *testing.T) {
	before := []canvas.Rect{{ID: "a", X: -10000, Width: 10, Height: 10}, {ID: "b", X: 9990, Width: 10, Height: 10}, {ID: "other", X: 50, Y: 50, Width: 30, Height: 30}}
	after, err := resizedSelection(before, []string{"a", "b"}, southEast, 20000, 0, false, true)
	if err != nil || !aspectClose(after[1].X, 10000) || after[0].X != -10000 || after[2] != before[2] {
		t.Fatal("origin clamp", after, err)
	}
	scale := after[0].Width / before[0].Width
	if !aspectClose(after[1].Width, 10*scale) || !aspectClose(after[1].Height, 10*scale) || !aspectClose(after[1].X, -10000+19990*scale) {
		t.Fatal("origin clamp broke common scale")
	}
	one := canvas.Rect{ID: "a", X: -9999.5, Y: 10, Width: 80, Height: 50}
	got := resizedCorner(one, northWest, -100, -100, false, true)
	if !aspectClose(got.X, -10000) || !aspectClose(got.X+got.Width, one.X+one.Width) || !aspectClose(got.Width/one.Width, got.Height/one.Height) {
		t.Fatal("west origin anchor clamp", got)
	}
	// An anchor beyond the origin bound can prevent shrinking at that corner.
	one.X = 10000
	got = resizedCorner(one, northWest, 40, 10, false, true)
	if got != one {
		t.Fatal("shrink crossed origin boundary", got)
	}
}

func TestAspectInspectorDimensionAndWorldPointerAgree(t *testing.T) {
	before := []canvas.Rect{{ID: "a", X: 10, Y: 20, Width: 80, Height: 50}, {ID: "b", X: 210, Y: 70, Width: 40, Height: 100}}
	ids := []string{"a", "b"}
	exact, err := resizedSelectionDimension(before, ids, "width", 300, false, true)
	if err != nil {
		t.Fatal(err)
	}
	bounds, _ := selectionBounds(exact, ids)
	if !aspectClose(bounds.Width, 300) || !aspectClose(bounds.Height, 187.5) {
		t.Fatal("exact inspector aspect", bounds)
	}
	fromHeight, err := resizedSelectionDimension(before, ids, "height", 187.5, false, true)
	if err != nil || !slices.Equal(exact, fromHeight) {
		t.Fatal("dimension helper disagrees", exact, fromHeight, err)
	}
	g := groupGeometry()
	bounds, _ = selectionBounds(before, ids)
	start := cornerHandles(bounds, g)[southEast-1].center
	end := canvas.Point{X: start.X + 60*g.Scale*g.View.Zoom, Y: start.Y + 2*g.Scale*g.View.Zoom}
	a, err := g.World(start)
	if err != nil {
		t.Fatal("start mapping", err)
	}
	b, err := g.World(end)
	if err != nil {
		t.Fatal("end mapping", err)
	}
	fromPointer, err := resizedSelection(before, ids, southEast, b.X-a.X, b.Y-a.Y, false, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := range exact {
		if !aspectClose(exact[i].X, fromPointer[i].X) || !aspectClose(exact[i].Y, fromPointer[i].Y) || !aspectClose(exact[i].Width, fromPointer[i].Width) || !aspectClose(exact[i].Height, fromPointer[i].Height) {
			t.Fatal("window/world transform changed aspect result", exact, fromPointer)
		}
	}
	unlocked, err := resizedSelectionDimension(before, ids, "width", 300, false, false)
	if err != nil || unlocked[0].Height != 50 || unlocked[1].Height != 100 || unlocked[1].Y != 70 {
		t.Fatal("unlocked behavior changed", unlocked, err)
	}
}

func TestAspectRejectedInputLeavesOriginalAndNoopAddsNoHistory(t *testing.T) {
	before := []canvas.Rect{{ID: "a", Width: 80, Height: 50}}
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got, err := resizedSelection(before, []string{"a"}, southEast, bad, 0, false, true); err == nil || got != nil {
			t.Fatal("nonfinite motion admitted")
		}
		if got, err := resizedSelectionDimension(before, []string{"a"}, "width", bad, false, true); err == nil || got != nil {
			t.Fatal("nonfinite dimension admitted")
		}
	}
	for _, ids := range [][]string{nil, {"missing"}, {"a", "a"}, {"a", "missing"}} {
		if got, err := resizedSelection(before, ids, southEast, 1, 1, false, true); err == nil || got != nil {
			t.Fatal("invalid selection admitted", ids)
		}
	}
	if got, err := resizedSelection(before, []string{"a"}, noCorner, 1, 1, false, true); err == nil || got != nil {
		t.Fatal("invalid corner admitted")
	}
	if got, err := resizedSelectionDimension(before, []string{"a"}, "depth", 1, false, true); err == nil || got != nil {
		t.Fatal("unknown dimension admitted")
	}
	if before[0].Width != 80 || before[0].Height != 50 {
		t.Fatal("rejection mutated source")
	}
	m := editModel(t)
	m.setSelection([]string{"block-1"}, "block-1")
	original := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	after, err := resizedSelection(original, m.selectedIDs(), northWest, 0, 0, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.applyObjects("Aspect resize", after, "block-1"); err != nil || m.journal.Snapshot().UndoCount != count {
		t.Fatal("noop added history", err)
	}
}

func TestAspectFloatingBoundariesAlwaysUseOneValidScale(t *testing.T) {
	random := rand.New(rand.NewSource(17))
	for trial := 0; trial < 160; trial++ {
		before := make([]canvas.Rect, 3)
		ids := []string{"a", "b", "c"}
		for i := range before {
			before[i] = canvas.Rect{ID: ids[i], X: random.Float64()*20000 - 10000, Y: random.Float64()*20000 - 10000, Width: 8 + random.Float64()*1992, Height: 8 + random.Float64()*1992}
		}
		for _, corner := range []resizeCorner{northWest, northEast, southEast, southWest} {
			after, err := resizedSelection(before, ids, corner, random.Float64()*2e6-1e6, random.Float64()*2e6-1e6, trial%2 == 0, true)
			if err != nil {
				t.Fatal("valid original must have feasible identity scale", trial, corner, err)
			}
			scale := after[0].Width / before[0].Width
			for i, o := range after {
				if !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) || !aspectClose(o.Width, before[i].Width*scale) || !aspectClose(o.Height, before[i].Height*scale) {
					t.Fatal("invalid per-member clamp", trial, corner, o)
				}
			}
		}
	}
}
