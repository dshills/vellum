//go:build darwin && cgo

package main

import (
	"encoding/json"
	"example.com/vellum/foundationintegration/inputbridge"
	"fmt"
	"math"
)

type geometryQuery struct {
	RectRange  [2]int     `json:"rect_range"`
	Screen     [4]float64 `json:"screen_rect"`
	PointIndex int        `json:"point_index"`
}

func geometryContracts() map[string]any {
	pool := poolBegin()
	defer poolEnd(pool)
	view := bareViewNew()
	defer bareViewFree(view)
	var last inputbridge.Update
	accept := true
	in := inputbridge.Attach(view, func(u inputbridge.Update) bool {
		if accept {
			last = u
		}
		return accept
	}, func() {})
	if in == nil {
		panic("geometry attach")
	}
	defer in.Close()
	source := "A👩Z\nאבג"
	g := inputbridge.Geometry{Viewport: [4]float64{0, 0, 100, 100}, FocusLine: 0,
		Cells:  []inputbridge.TextCell{{Start: 0, End: 1, Line: 0, Rect: [4]float64{10, 10, 10, 20}}, {Start: 1, End: 3, Line: 0, Rect: [4]float64{20, 10, 20, 20}}, {Start: 3, End: 4, Line: 0, Rect: [4]float64{40, 10, 10, 20}}, {Start: 4, End: 5, Line: 0, Rect: [4]float64{50, 10, 0, 20}}, {Start: 5, End: 6, Line: 1, Rect: [4]float64{30, 30, 10, 20}}, {Start: 6, End: 7, Line: 1, Rect: [4]float64{20, 30, 10, 20}}, {Start: 7, End: 8, Line: 1, Rect: [4]float64{10, 30, 10, 20}}},
		Carets: []inputbridge.TextCaret{{Offset: 0, Line: 0, Rect: [4]float64{10, 10, 0, 20}}, {Offset: 1, Line: 0, Rect: [4]float64{20, 10, 0, 20}}, {Offset: 3, Line: 0, Rect: [4]float64{40, 10, 0, 20}}, {Offset: 4, Line: 0, Rect: [4]float64{50, 10, 0, 20}}, {Offset: 5, Line: 1, Rect: [4]float64{40, 30, 0, 20}}, {Offset: 6, Line: 1, Rect: [4]float64{30, 30, 0, 20}}, {Offset: 7, Line: 1, Rect: [4]float64{20, 30, 0, 20}}, {Offset: 8, Line: 1, Rect: [4]float64{10, 30, 0, 20}}}}
	publish := func(text string, ack uint64, geom inputbridge.Geometry) {
		if err := in.UpdateGeometry(text, 1, 1, -1, -1, true, ack, [4]float64{20, 10, 0, 20}, geom); err != nil {
			panic(err)
		}
	}
	query := func(start, length int, p [2]float64) geometryQuery {
		raw := in.QueryProbe(start, length, p)
		var q geometryQuery
		if err := json.Unmarshal([]byte(raw), &q); err != nil {
			panic(err)
		}
		return q
	}
	publish(source, 0, g)
	var cases []map[string]any
	for _, tc := range []struct {
		request, actual [2]int
		hit             int
	}{{[2]int{0, 2147483647}, [2]int{0, 5}, 0}, {[2]int{2, 1}, [2]int{1, 2}, 1}, {[2]int{5, 3}, [2]int{5, 3}, 6}, {[2]int{4, 3}, [2]int{4, 1}, -1}, {[2]int{3, 0}, [2]int{3, 0}, -1}, {[2]int{20, 1}, [2]int{-1, 0}, -1}} {
		q := query(tc.request[0], tc.request[1], [2]float64{-999, -999})
		if q.RectRange != tc.actual {
			panic(fmt.Sprintf("range %v: %+v", tc.request, q))
		}
		if q.Screen[2] > 0 {
			center := [2]float64{q.Screen[0] + q.Screen[2]/2, q.Screen[1] + q.Screen[3]/2}
			hit := query(tc.request[0], tc.request[1], center).PointIndex
			// The first full LTR fragment's center is inside the emoji cell.
			want := tc.hit
			if tc.request[0] == 0 {
				want = 1
			}
			if hit != want {
				panic(fmt.Sprintf("hit %d want %d", hit, want))
			}
			q.PointIndex = hit
		}
		cases = append(cases, map[string]any{"requested": tc.request, "observed": q})
	}
	original := query(1, 2, [2]float64{-999, -999})
	invalid := g
	invalid.Cells = append([]inputbridge.TextCell(nil), g.Cells...)
	invalid.Cells[0].Rect[0] = math.NaN()
	if in.UpdateGeometry(source, 1, 1, -1, -1, true, 0, [4]float64{}, invalid) == nil {
		panic("nonfinite geometry accepted")
	}
	if query(1, 2, [2]float64{-999, -999}) != original {
		panic("invalid publication changed native geometry")
	}
	// Refused preflight retains the old geometry; accepted edits invalidate it.
	accept = false
	in.Probe(1)
	if query(1, 2, [2]float64{-999, -999}) != original {
		panic("refusal invalidated geometry")
	}
	accept = true
	in.Probe(1)
	if query(1, 2, [2]float64{}).RectRange != [2]int{-1, 0} {
		panic("native edit retained stale geometry")
	}
	publish(source, 0, g)
	if query(1, 2, [2]float64{}).RectRange != [2]int{-1, 0} {
		panic("stale ack restored geometry")
	}
	// Application publication at the admitted sequence restores a matching layout.
	publish(source, last.Sequence, g)
	// Translate and scale all view geometry, as the UI does after scroll/DPI changes.
	moved := g
	moved.Cells = append([]inputbridge.TextCell(nil), g.Cells...)
	moved.Carets = append([]inputbridge.TextCaret(nil), g.Carets...)
	transform := func(r [4]float64) [4]float64 { return [4]float64{r[0]/2 + 7, r[1]/2 + 9, r[2] / 2, r[3] / 2} }
	for i := range moved.Cells {
		moved.Cells[i].Rect = transform(moved.Cells[i].Rect)
	}
	for i := range moved.Carets {
		moved.Carets[i].Rect = transform(moved.Carets[i].Rect)
	}
	publish(source, last.Sequence, moved)
	scaled := query(1, 2, [2]float64{-999, -999})
	if scaled.Screen[2] != 10 || scaled.Screen[3] != 10 {
		panic("scale conversion")
	}
	center := [2]float64{scaled.Screen[0] + 5, scaled.Screen[1] + 5}
	if query(1, 2, center).PointIndex != 1 {
		panic("translated point conversion")
	}
	// Cropped viewport clips the first fragment but retains its logical range.
	moved.Viewport = [4]float64{20, 0, 80, 100}
	publish(source, last.Sequence, moved)
	clipped := query(1, 2, [2]float64{-999, -999})
	if clipped.RectRange != [2]int{1, 2} || clipped.Screen[2] != 7 {
		panic(fmt.Sprintf("viewport clipping %+v", clipped))
	}
	whole := query(0, 2147483647, [2]float64{-999, -999})
	if whole.RectRange != [2]int{1, 4} {
		panic(fmt.Sprintf("fully offscreen cell advertised %+v", whole))
	}
	// A duplicated offset across wrapped lines has an explicit focus-line affinity.
	moved.Carets = append(moved.Carets, inputbridge.TextCaret{Offset: 1, Line: 1, Rect: [4]float64{30, 40, 0, 20}})
	moved.FocusLine = 1
	publish(source, last.Sequence, moved)
	affinity := query(1, 0, [2]float64{-999, -999})
	if affinity.RectRange != [2]int{1, 0} || affinity.Screen[3] != 20 {
		panic("wrap focus affinity")
	}
	return map[string]any{"queries": cases, "scaled": scaled, "clipped": clipped, "clipped_whole": whole, "affinity": affinity, "invalid_atomic": true, "refused_edit_preserves_geometry": true, "native_edit_invalidates_geometry": true, "stale_ack_rejected": true, "screen_point_roundtrip": true, "scope": "hidden Cocoa view queries using declared advance cells; actual rendered layout is checked separately"}
}
