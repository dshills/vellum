package main

import (
	"math"

	"github.com/dshills/vellum/canvas"
)

func marqueeRect(a, b canvas.Point) canvas.Rect {
	return canvas.Rect{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y), Width: math.Abs(b.X - a.X), Height: math.Abs(b.Y - a.Y)}
}

// Shift-marquee adds positive-area intersections to the captured selection.
// Keep this ungrouped additive entry point for existing callers.
func marqueeIDs(objects []canvas.Rect, prior []string, bounds canvas.Rect) []string {
	return marqueeSelection(objects, nil, prior, bounds, false)
}

// marqueeSelection computes a detached selection from one gesture snapshot.
// Any intersected member affects its entire flat group. Results follow paint
// order, contain only live IDs, and never mutate the snapshot or prior set.
// The model's canonical setter separately excludes hidden and locked units.
func marqueeSelection(objects []canvas.Rect, groups map[string]string, prior []string, bounds canvas.Rect, subtract bool) []string {
	unit := func(id string) string {
		if group := groups[id]; group != "" {
			return group
		}
		return id
	}
	wanted := make(map[string]bool, len(prior))
	for _, id := range prior {
		wanted[unit(id)] = true
	}
	touched := make(map[string]bool)
	for _, o := range objects {
		if bounds.Width > 0 && bounds.Height > 0 && o.X < bounds.X+bounds.Width && o.X+o.Width > bounds.X && o.Y < bounds.Y+bounds.Height && o.Y+o.Height > bounds.Y {
			touched[unit(o.ID)] = true
		}
	}
	ids := make([]string, 0, len(objects))
	for _, o := range objects {
		key := unit(o.ID)
		keep := wanted[key] || touched[key]
		if subtract {
			keep = wanted[key] && !touched[key]
		}
		if keep {
			ids = append(ids, o.ID)
		}
	}
	return ids
}
