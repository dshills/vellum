package main

import (
	"math"

	"github.com/dshills/vellum/canvas"
)

func marqueeRect(a, b canvas.Point) canvas.Rect {
	return canvas.Rect{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y), Width: math.Abs(b.X - a.X), Height: math.Abs(b.Y - a.Y)}
}

// Shift-marquee adds positive-area intersections to the captured selection.
// Paint order determines result order; edge-only contact does not select.
func marqueeIDs(objects []canvas.Rect, prior []string, bounds canvas.Rect) []string {
	wanted := make(map[string]bool, len(prior))
	for _, id := range prior {
		wanted[id] = true
	}
	ids := make([]string, 0, len(objects))
	for _, o := range objects {
		intersects := bounds.Width > 0 && bounds.Height > 0 && o.X < bounds.X+bounds.Width && o.X+o.Width > bounds.X && o.Y < bounds.Y+bounds.Height && o.Y+o.Height > bounds.Y
		if wanted[o.ID] || intersects {
			ids = append(ids, o.ID)
		}
	}
	return ids
}
