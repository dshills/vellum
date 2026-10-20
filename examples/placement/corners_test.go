package main

import (
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

func cornerGeometry() placement.Geometry {
	return placement.Geometry{Window: "placement", Origin: canvas.Point{X: -30, Y: 25}, Clip: canvas.Bounds{Min: canvas.Point{X: 0, Y: 0}, Max: canvas.Point{X: 1800, Y: 1300}}, Scale: 2, View: canvas.Viewport{Width: 800, Height: 600, Zoom: 1.5, Center: canvas.Point{X: 30, Y: -20}}}
}
func TestAllCornerHandlesAnchorHistoryAndClamps(t *testing.T) {
	for _, c := range []resizeCorner{northWest, northEast, southEast, southWest} {
		t.Run(string(rune('0'+c)), func(t *testing.T) {
			m := editModel(t)
			m.snap = false
			m.scene.Selected = "block-1"
			g := cornerGeometry()
			before := slices.Clone(m.scene.Objects)
			o := before[0]
			start := cornerHandles(o, g)[c-1].center
			s := m.beginMove(g, start)
			if s == nil || !s.resize || s.corner != c {
				t.Fatal("corner did not capture selected object", c, s)
			}
			dx, dy := 40.0, 20.0
			if c == northWest || c == southWest {
				dx = -dx
			}
			if c == northWest || c == northEast {
				dy = -dy
			}
			end := canvas.Point{X: start.X + dx*g.Scale*g.View.Zoom, Y: start.Y + dy*g.Scale*g.View.Zoom}
			m.previewMove(s, end)
			if m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("preview added history")
			}
			if err := m.finishMove(s, end); err != nil {
				t.Fatal(err)
			}
			got := m.scene.Objects[0]
			if got.Width != 120 || got.Height != 70 || got.ID != o.ID || m.next != 3 || m.journal.Snapshot().UndoCount != 4 {
				t.Fatal("resize identity/dimensions/history", got)
			}
			wantX, wantY := o.X, o.Y
			if c == northWest || c == southWest {
				wantX -= 40
			}
			if c == northWest || c == northEast {
				wantY -= 20
			}
			if got.X != wantX || got.Y != wantY {
				t.Fatal("opposite corner moved", got)
			}
			after := slices.Clone(m.scene.Objects)
			_ = m.finishMove(s, end)
			if m.journal.Snapshot().UndoCount != 4 {
				t.Fatal("duplicate release")
			}
			if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) {
				t.Fatal("undo exact rectangle/order", err)
			}
			if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
				t.Fatal("redo", err)
			}
			// Crossing the opposite corner clamps, never flips, with anchor preserved.
			shrunk := resizedCorner(o, c, -dx*100, -dy*100, false)
			if shrunk.Width != 8 || shrunk.Height != 8 {
				t.Fatal("minimum", shrunk)
			}
			if (c == northWest || c == southWest) && shrunk.X+shrunk.Width != o.X+o.Width {
				t.Fatal("west clamp anchor")
			}
			if (c == northWest || c == northEast) && shrunk.Y+shrunk.Height != o.Y+o.Height {
				t.Fatal("north clamp anchor")
			}
		})
	}
}

func TestHandleHitPriorityClippingAndInteriorMove(t *testing.T) {
	m := editModel(t)
	m.snap = false
	m.scene.Selected = "block-1"
	g := cornerGeometry()
	center := cornerHandles(m.scene.Objects[0], g)[0].center
	// The outside half of a visible selected handle still starts resize.
	p := canvas.Point{X: center.X - 4*g.Scale, Y: center.Y - 4*g.Scale}
	s := m.beginMove(g, p)
	if s == nil || s.corner != northWest {
		t.Fatal("outside-half handle missed")
	}
	m.cancelMove(s)
	// Handles remain available over another object's fill.
	m.scene.Objects[1].X, m.scene.Objects[1].Y = -5, -5
	s = m.beginMove(g, center)
	if s == nil || s.id != "block-1" || s.corner != northWest {
		t.Fatal("handle priority")
	}
	m.cancelMove(s)
	// Interior input keeps ordinary movement, not resizing.
	middle := g.View.Screen(canvas.Point{X: 40, Y: 25})
	p = canvas.Point{X: g.Origin.X + middle.X*g.Scale, Y: g.Origin.Y + middle.Y*g.Scale}
	s = m.beginMove(g, p)
	if s == nil || s.resize {
		t.Fatal("interior became resize")
	}
	m.cancelMove(s)
	clipped := g
	clipped.Clip.Min.X = center.X + 1
	if handleAt(m.scene.Objects[0], clipped, canvas.Point{X: center.X + 3, Y: center.Y}) != noCorner {
		t.Fatal("hidden center handle hittable")
	}
	if handleAt(m.scene.Objects[0], g, canvas.Point{X: g.Clip.Max.X, Y: center.Y}) != noCorner {
		t.Fatal("outside clip hit")
	}
	m.scene.Selected = ""
	s = m.beginMove(g, center)
	if s != nil && s.resize {
		t.Fatal("unselected object has handles")
	}
	m.cancelMove(s)
}

func TestWestNorthResizeBoundsCancelAndSnap(t *testing.T) {
	m := editModel(t)
	m.scene.Selected = "block-1"
	m.snap = true
	g := cornerGeometry()
	o := m.scene.Objects[0]
	start := cornerHandles(o, g)[0].center
	s := m.beginMove(g, start)
	end := canvas.Point{X: start.X - 23*g.Scale*g.View.Zoom, Y: start.Y - 13*g.Scale*g.View.Zoom}
	if err := m.finishMove(s, end); err != nil {
		t.Fatal(err)
	}
	got := m.scene.Objects[0]
	if got.Width != 100 || got.Height != 60 || got.X != -20 || got.Y != -10 {
		t.Fatal("snapped opposite corner", got)
	}
	_ = m.journal.Undo()
	// Invalid west position may not commit or destroy the previous document.
	m.scene.Objects[0].X = -10000
	g.View.Center = canvas.Point{X: -9970}
	start = cornerHandles(m.scene.Objects[0], g)[0].center
	before := slices.Clone(m.scene.Objects)
	count := m.journal.Snapshot().UndoCount
	s = m.beginMove(g, start)
	if s == nil {
		t.Fatal("bound start")
	}
	if err := m.finishMove(s, canvas.Point{X: start.X - 60, Y: start.Y}); err != nil {
		t.Fatal(err)
	}
	if m.move != nil || !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot().UndoCount != count {
		t.Fatal("invalid origin committed")
	}
}
