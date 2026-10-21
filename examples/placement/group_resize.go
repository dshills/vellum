package main

import (
	"errors"
	"math"
	"slices"

	"github.com/dshills/vellum/canvas"
)

// Bounds include only selected objects, in world coordinates. This value is
// shared by handle painting, hit testing and the original resize transform.
func selectionBounds(objects []canvas.Rect, ids []string) (canvas.Rect, bool) {
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, o := range objects {
		if wanted[o.ID] {
			left, top = math.Min(left, o.X), math.Min(top, o.Y)
			right, bottom = math.Max(right, o.X+o.Width), math.Max(bottom, o.Y+o.Height)
		}
	}
	if !(right > left && bottom > top) {
		return canvas.Rect{}, false
	}
	return canvas.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}, true
}

// Scale every member from the original snapshot about the opposite group
// corner. One shared scale per axis preserves spacing and relative sizes.
// Member limits constrain the shared scale, never individual rectangles.
func resizedSelection(before []canvas.Rect, ids []string, corner resizeCorner, dx, dy float64, snap bool, aspectLocked ...bool) ([]canvas.Rect, error) {
	return resizedSelectionGrid(before, ids, corner, dx, dy, snap, 20, len(aspectLocked) > 0 && aspectLocked[0])
}
func resizedSelectionGrid(before []canvas.Rect, ids []string, corner resizeCorner, dx, dy float64, snap bool, spacing float64, aspectLocked bool) ([]canvas.Rect, error) {
	spacing = gridUnit(spacing)
	if aspectLocked {
		return resizedAspectSelectionGrid(before, ids, corner, dx, dy, snap, spacing)
	}
	bounds, ok := selectionBounds(before, ids)
	if !ok || corner < northWest || corner > southWest || math.IsNaN(dx) || math.IsNaN(dy) || math.IsInf(dx, 0) || math.IsInf(dy, 0) {
		return nil, errors.New("invalid group resize")
	}
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	minX, minY, maxX, maxY := 0.0, 0.0, math.Inf(1), math.Inf(1)
	for _, o := range before {
		if wanted[o.ID] {
			if !validSize(o.Width, o.Height) {
				return nil, errors.New("invalid component dimensions")
			}
			minX, minY = math.Max(minX, minComponentSize/o.Width), math.Max(minY, minComponentSize/o.Height)
			maxX, maxY = math.Min(maxX, maxComponentSize/o.Width), math.Min(maxY, maxComponentSize/o.Height)
		}
	}
	if minX > maxX || minY > maxY {
		return nil, errors.New("incompatible component size limits")
	}
	west, north := corner == northWest || corner == southWest, corner == northWest || corner == northEast
	if west {
		dx = -dx
	}
	if north {
		dy = -dy
	}
	scale := func(size, delta, min, max float64) float64 {
		if delta == 0 {
			return 1
		}
		next := size + delta
		if snap {
			next = math.Round(next/spacing) * spacing
		}
		return math.Max(min, math.Min(max, next/size))
	}
	sx, sy := scale(bounds.Width, dx, minX, maxX), scale(bounds.Height, dy, minY, maxY)
	// At a shared size limit, multiplication can round one ulp outside it.
	// Move the shared scale toward identity rather than clamp any member.
	for _, o := range before {
		if !wanted[o.ID] {
			continue
		}
		if w := o.Width * sx; w < minComponentSize || w > maxComponentSize {
			sx = math.Nextafter(sx, 1)
		}
		if h := o.Height * sy; h < minComponentSize || h > maxComponentSize {
			sy = math.Nextafter(sy, 1)
		}
	}
	ax, ay := bounds.X, bounds.Y
	if west {
		ax += bounds.Width
	}
	if north {
		ay += bounds.Height
	}
	after := slices.Clone(before)
	for i, o := range before {
		if !wanted[o.ID] {
			continue
		}
		if sx != 1 {
			after[i].X = ax + (o.X-ax)*sx
			after[i].Width = o.Width * sx
		}
		if sy != 1 {
			after[i].Y = ay + (o.Y-ay)*sy
			after[i].Height = o.Height * sy
		}
		if !validPosition(after[i].X, after[i].Y) || !validSize(after[i].Width, after[i].Height) {
			return nil, errors.New("all resized components must remain within drawing bounds")
		}
	}
	return after, nil
}
