package main

import (
	"errors"
	"math"
	"slices"

	"github.com/dshills/vellum/canvas"
)

// resizedAspectSelection uses the dimension with the larger relative pointer
// change (width wins ties). Only that world-space dimension snaps; a single
// shared scale derives the other dimension, every member size and spacing.
func resizedAspectSelection(before []canvas.Rect, ids []string, corner resizeCorner, dx, dy float64, snap bool) ([]canvas.Rect, error) {
	bounds, wanted, err := aspectSelection(before, ids)
	if err != nil || corner < northWest || corner > southWest || !finiteResize(dx) || !finiteResize(dy) {
		return nil, errors.New("invalid aspect resize")
	}
	if dx == 0 && dy == 0 {
		return slices.Clone(before), nil
	}
	west, north := corner == northWest || corner == southWest, corner == northWest || corner == northEast
	if west {
		dx = -dx
	}
	if north {
		dy = -dy
	}
	size, delta := bounds.Width, dx
	if math.Abs(dy)/bounds.Height > math.Abs(dx)/bounds.Width {
		size, delta = bounds.Height, dy
	}
	next := size + delta
	if snap {
		next = math.Round(next/20) * 20
	}
	ax, ay := bounds.X, bounds.Y
	if west {
		ax += bounds.Width
	}
	if north {
		ay += bounds.Height
	}
	return aspectScaleSelection(before, wanted, ax, ay, next/size)
}

// resizedSelectionDimension is the shared inspector/button transform. target
// is an absolute selection-bounds dimension in world units. Inspector exact
// values use snap=false; increment buttons may snap only the driving dimension.
// The top-left stays fixed, while an aspect lock scales both axes uniformly.
func resizedSelectionDimension(before []canvas.Rect, ids []string, dimension string, target float64, snap, aspectLocked bool) ([]canvas.Rect, error) {
	if !finiteResize(target) || (dimension != "width" && dimension != "height") {
		return nil, errors.New("invalid resize dimension")
	}
	bounds, ok := selectionBounds(before, ids)
	if !ok {
		return nil, errors.New("no resize selection")
	}
	dx, dy := 0.0, 0.0
	if dimension == "width" {
		dx = target - bounds.Width
	} else {
		dy = target - bounds.Height
	}
	return resizedSelection(before, ids, southEast, dx, dy, snap, aspectLocked)
}

func finiteResize(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func aspectSelection(before []canvas.Rect, ids []string) (canvas.Rect, map[string]bool, error) {
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		if wanted[id] {
			return canvas.Rect{}, nil, errors.New("duplicate aspect resize selection")
		}
		wanted[id] = true
	}
	seen := make(map[string]bool, len(wanted))
	for _, o := range before {
		if !wanted[o.ID] {
			continue
		}
		if seen[o.ID] || !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) {
			return canvas.Rect{}, nil, errors.New("invalid aspect resize component")
		}
		seen[o.ID] = true
	}
	bounds, ok := selectionBounds(before, ids)
	if !ok || len(seen) != len(wanted) {
		return canvas.Rect{}, nil, errors.New("missing aspect resize component")
	}
	return bounds, wanted, nil
}

// Intersect all member size and origin constraints before transforming any
// rectangle. Identity is always feasible for a validated original selection.
// Origins, unlike rectangle far edges, are restricted to [-10000,10000].
func aspectScaleSelection(before []canvas.Rect, wanted map[string]bool, ax, ay, requested float64) ([]canvas.Rect, error) {
	low, high := 0.0, math.Inf(1)
	constrainOrigin := func(anchor, original float64) {
		delta := original - anchor
		if delta == 0 {
			return
		}
		a, b := (-10000-anchor)/delta, (10000-anchor)/delta
		if a > b {
			a, b = b, a
		}
		low, high = math.Max(low, a), math.Min(high, b)
	}
	for _, o := range before {
		if !wanted[o.ID] {
			continue
		}
		low = math.Max(low, math.Max(minComponentSize/o.Width, minComponentSize/o.Height))
		high = math.Min(high, math.Min(maxComponentSize/o.Width, maxComponentSize/o.Height))
		constrainOrigin(ax, o.X)
		constrainOrigin(ay, o.Y)
	}
	if low > high || !finiteResize(requested) {
		return nil, errors.New("incompatible aspect resize constraints")
	}
	scale := math.Max(low, math.Min(high, requested))
	after := slices.Clone(before)
	// A ratio or affine coordinate at an exact limit can round outside by an
	// ulp. Move the whole scale toward identity, never clamp individual members.
	for attempt := 0; attempt < 64; attempt++ {
		valid := true
		for i, o := range before {
			if !wanted[o.ID] {
				continue
			}
			if scale == 1 {
				after[i] = o
				continue
			}
			after[i].X, after[i].Y = ax+(o.X-ax)*scale, ay+(o.Y-ay)*scale
			after[i].Width, after[i].Height = o.Width*scale, o.Height*scale
			valid = valid && validPosition(after[i].X, after[i].Y) && validSize(after[i].Width, after[i].Height)
		}
		if valid {
			return after, nil
		}
		next := math.Nextafter(scale, 1)
		if next == scale {
			break
		}
		scale = next
	}
	return nil, errors.New("aspect resize exceeds drawing bounds")
}
