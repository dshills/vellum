package main

import (
	"math"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

type resizeCorner uint8

const (
	noCorner resizeCorner = iota
	northWest
	northEast
	southEast
	southWest
)

type cornerHandle struct {
	corner resizeCorner
	center canvas.Point
}

// Centers and hit targets share the same current window-space transform.
// Handle size is logical pixels, independent of the drawing's zoom.
func cornerHandles(o canvas.Rect, g placement.Geometry) [4]cornerHandle {
	points := [4]canvas.Point{{X: o.X, Y: o.Y}, {X: o.X + o.Width, Y: o.Y}, {X: o.X + o.Width, Y: o.Y + o.Height}, {X: o.X, Y: o.Y + o.Height}}
	var handles [4]cornerHandle
	for i, p := range points {
		local := g.View.Screen(p)
		handles[i] = cornerHandle{corner: resizeCorner(i + 1), center: canvas.Point{X: g.Origin.X + local.X*g.Scale, Y: g.Origin.Y + local.Y*g.Scale}}
	}
	return handles
}
func inCanvas(g placement.Geometry, p canvas.Point) bool {
	return p.X >= g.Clip.Min.X && p.Y >= g.Clip.Min.Y && p.X < g.Clip.Max.X && p.Y < g.Clip.Max.Y
}
func handleAt(o canvas.Rect, g placement.Geometry, p canvas.Point) resizeCorner {
	if !inCanvas(g, p) {
		return noCorner
	}
	best, result := math.Inf(1), noCorner
	for _, h := range cornerHandles(o, g) {
		if !inCanvas(g, h.center) {
			continue
		}
		dx, dy := p.X-h.center.X, p.Y-h.center.Y
		if math.Abs(dx) <= 7*g.Scale && math.Abs(dy) <= 7*g.Scale && dx*dx+dy*dy < best {
			best, result = dx*dx+dy*dy, h.corner
		}
	}
	return result
}

// Resize from the original rectangle and hold its opposite corner fixed.
// Clamping keeps that anchor fixed even when the dragged corner crosses it.
func resizedCorner(before canvas.Rect, corner resizeCorner, dx, dy float64, snap bool, aspectLocked ...bool) canvas.Rect {
	return resizedCornerGrid(before, corner, dx, dy, snap, 20, len(aspectLocked) > 0 && aspectLocked[0])
}
func resizedCornerGrid(before canvas.Rect, corner resizeCorner, dx, dy float64, snap bool, spacing float64, aspectLocked bool) canvas.Rect {
	spacing = gridUnit(spacing)
	if aspectLocked {
		after, err := resizedAspectSelectionGrid([]canvas.Rect{before}, []string{before.ID}, corner, dx, dy, snap, spacing)
		if err == nil {
			return after[0]
		}
		return before
	}
	west, north := corner == northWest || corner == southWest, corner == northWest || corner == northEast
	if west {
		dx = -dx
	}
	if north {
		dy = -dy
	}
	after := before
	after.Width, after.Height = resizedDimensionsGrid(before, dx, dy, snap, spacing)
	if west {
		after.X = before.X + before.Width - after.Width
	}
	if north {
		after.Y = before.Y + before.Height - after.Height
	}
	return after
}
