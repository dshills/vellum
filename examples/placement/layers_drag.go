package main

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
)

type layerBox struct {
	id     string
	bounds canvas.Bounds
}
type layerLayout struct {
	clip  canvas.Bounds
	scale float64
	rows  []layerBox
}
type layerDrag struct {
	source, primary        string
	ids, selection         []string
	priorPrimary           string
	before                 []canvas.Rect
	groups                 map[string]string
	layout                 layerLayout
	start, pointer         canvas.Point
	dragging, valid, front bool
	target                 string
	line                   float64
	disabled               map[string]bool
}

func (g layerLayout) rowAt(p canvas.Point) (string, bool, float64) {
	if p.X < g.clip.Min.X || p.X >= g.clip.Max.X || p.Y < g.clip.Min.Y || p.Y >= g.clip.Max.Y {
		return "", false, 0
	}
	for _, r := range g.rows {
		b := r.bounds
		if p.X >= b.Min.X && p.X < b.Max.X && p.Y >= b.Min.Y && p.Y < b.Max.Y {
			front := p.Y < (b.Min.Y+b.Max.Y)/2
			y := b.Max.Y
			if front {
				y = b.Min.Y
			}
			return r.id, front, math.Max(g.clip.Min.Y, math.Min(g.clip.Max.Y, y))
		}
	}
	return "", false, 0
}
func (m *model) beginLayerDrag(source string, p canvas.Point, g layerLayout) *layerDrag {
	if m.layerFilterActive() || m.clipboardAvailable() != nil || g.scale <= 0 || math.IsNaN(g.scale) || math.IsInf(g.scale, 0) {
		return nil
	}
	hit, _, _ := g.rowAt(p)
	if hit != source {
		return nil
	}
	primary := ""
	for i := len(m.scene.Objects) - 1; i >= 0; i-- {
		o := m.scene.Objects[i]
		if o.ID == source || m.groups[o.ID] == source {
			primary = o.ID
			break
		}
	}
	if primary == "" || m.layerLocked(primary) {
		return nil
	}
	selection := m.selectedIDs()
	ids := m.layerUnitIDs(primary)
	if slices.Contains(selection, primary) {
		ids = slices.Clone(selection)
		primary = m.scene.Selected
	}
	s := &layerDrag{source: source, primary: primary, ids: ids, selection: selection, priorPrimary: m.scene.Selected, before: slices.Clone(m.scene.Objects), groups: maps.Clone(m.groups), layout: layerLayout{clip: g.clip, scale: g.scale, rows: slices.Clone(g.rows)}, start: p, pointer: p}
	s.disabled = make(map[string]bool)
	for _, c := range m.commands.Search("") {
		s.disabled[c.ID] = m.commands.Enabled(c.ID) != nil
	}
	s.disabled["control:geometry"] = m.editAvailable() != nil
	s.disabled["control:interaction"] = m.clipboardAvailable() != nil
	s.disabled["control:recent"] = m.files == nil || len(m.recent.Paths()) == 0 || m.clipboardAvailable() != nil
	m.layerDrag = s
	return s
}
func (m *model) cancelLayerDrag(s *layerDrag) {
	if s != nil && m.layerDrag == s {
		m.layerDrag = nil
		m.status = "Layer drag canceled"
	}
}
func (m *model) updateLayerDrag(s *layerDrag, p canvas.Point, g layerLayout) bool {
	if s == nil || m.layerDrag != s {
		return false
	}
	if m.layerFilterActive() || g.clip != s.layout.clip || g.scale != s.layout.scale || !slices.Equal(g.rows, s.layout.rows) || !slices.Equal(m.scene.Objects, s.before) || !maps.Equal(m.groups, s.groups) || m.scene.Selected != s.priorPrimary || !slices.Equal(m.selectedIDs(), s.selection) {
		m.cancelLayerDrag(s)
		return false
	}
	s.pointer = p
	if math.Hypot(p.X-s.start.X, p.Y-s.start.Y) >= 6*g.scale {
		s.dragging = true
	}
	s.target, s.front, s.line = g.rowAt(p)
	s.valid = false
	if s.dragging && s.target != "" {
		_, err := reorderedLayers(s.before, s.groups, s.ids, s.target, s.front)
		s.valid = err == nil
		// A group marker targets the outer group boundary, including members
		// below the viewport. Clamp its indicator to the visible list edge.
		if s.valid && !s.front && !m.collapsedGroups[s.target] {
			for _, o := range s.before {
				if s.groups[o.ID] != s.target {
					continue
				}
				s.valid = false
				for _, row := range g.rows {
					if row.id == o.ID {
						s.line, s.valid = math.Max(g.clip.Min.Y, math.Min(g.clip.Max.Y, row.bounds.Max.Y)), true
						break
					}
				}
				break
			}
		}
	}
	if s.dragging {
		m.status = "Layer drag: release over another row; Escape cancels"
		if s.valid {
			side := "below"
			if s.front {
				side = "above"
			}
			m.status = fmt.Sprintf("Drop %d components %s %s", len(s.ids), side, s.target)
		}
	}
	return true
}

// Adopt only a translation produced by our native list scroll operation. All
// other geometry changes keep the ordinary cancellation path. The stationary
// pointer is re-hit-tested against the new full (unclipped) row positions.
func (m *model) scrollLayerDrag(s *layerDrag, before, after layerLayout, delta float64) bool {
	if s == nil {
		return false
	}
	if !m.updateLayerDrag(s, s.pointer, before) {
		return false
	}
	valid := !math.IsNaN(delta) && !math.IsInf(delta, 0) && before.clip == after.clip && before.scale == after.scale && len(before.rows) == len(after.rows)
	// ApplyScenePos can remove an initial subpixel layout alignment on the
	// first scroll. Accept at most one dot of uniform translation rounding,
	// then verify every row against the measured displacement.
	if valid && len(before.rows) > 0 {
		actual := before.rows[0].bounds.Min.Y - after.rows[0].bounds.Min.Y
		valid = math.Abs(actual-delta) <= 1
		delta = actual
	}
	if valid {
		for i, a := range before.rows {
			b := after.rows[i]
			if a.id != b.id || a.bounds.Min.X != b.bounds.Min.X || a.bounds.Max.X != b.bounds.Max.X || !layerTranslatedY(a.bounds.Min.Y, b.bounds.Min.Y, delta) || !layerTranslatedY(a.bounds.Max.Y, b.bounds.Max.Y, delta) {
				valid = false
				break
			}
		}
	}
	if !valid {
		m.cancelLayerDrag(s)
		return false
	}
	s.layout.rows = slices.Clone(after.rows)
	return m.updateLayerDrag(s, s.pointer, after)
}

func layerTranslatedY(before, after, delta float64) bool {
	// Native row/scroll geometry is float32. Distant offscreen rows can have
	// larger ULPs than a fixed epsilon, especially at high display scaling.
	ulp := func(v float64) float64 {
		f := float32(v)
		return math.Abs(float64(math.Nextafter32(f, float32(math.Inf(1)))) - float64(f))
	}
	tolerance := math.Max(0.001, 2*(ulp(before)+ulp(after)+ulp(delta)))
	return math.Abs(before-delta-after) <= tolerance
}

// A bounded, vertical edge band. Outside release is still invalid; no scrolling
// continues while the pointer is over canvas/chrome or before the threshold.
func layerEdgeDelta(s *layerDrag) float64 {
	if s == nil || !s.dragging {
		return 0
	}
	g, p := s.layout, s.pointer
	if p.X < g.clip.Min.X || p.X >= g.clip.Max.X || p.Y < g.clip.Min.Y || p.Y >= g.clip.Max.Y {
		return 0
	}
	band := math.Min(24*g.scale, (g.clip.Max.Y-g.clip.Min.Y)/3)
	if p.Y < g.clip.Min.Y+band {
		return -12 * g.scale
	}
	if p.Y > g.clip.Max.Y-band {
		return 12 * g.scale
	}
	return 0
}
func (m *model) finishLayerDrag(s *layerDrag, p canvas.Point, g layerLayout) error {
	if !m.updateLayerDrag(s, p, g) {
		return nil
	}
	if !s.dragging {
		source, _, _ := g.rowAt(p)
		m.cancelLayerDrag(s)
		if source == s.source {
			return m.selectLayer(source)
		}
		return nil
	}
	if !s.valid {
		m.cancelLayerDrag(s)
		return nil
	}
	after, err := reorderedLayers(s.before, s.groups, s.ids, s.target, s.front)
	m.layerDrag = nil // Completion must unlock other app commands before publishing.
	if err != nil {
		return err
	}
	if err = m.applyDocument("Reorder layers", after, s.groups, s.ids, s.primary); err != nil {
		return err
	}
	m.status = "Reordered layers"
	return nil
}

// Move the chosen set as one stable block around a live target. A group marker
// targets the whole group's outer paint boundary; a member row targets that
// exact component. Every unselected pair keeps its original relative order.
func reorderedLayers(before []canvas.Rect, groups map[string]string, ids []string, target string, front bool) ([]canvas.Rect, error) {
	if target == "" {
		return nil, history.ErrStale
	}
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	boundary := ""
	moving, rest := make([]canvas.Rect, 0, len(ids)), make([]canvas.Rect, 0, len(before))
	for _, o := range before {
		if o.ID == target || groups[o.ID] == target {
			if selected[o.ID] {
				return nil, errors.New("cannot drop onto the moving selection")
			}
			if boundary == "" || front {
				boundary = o.ID
			}
		}
		if selected[o.ID] {
			moving = append(moving, o)
		} else {
			rest = append(rest, o)
		}
	}
	if len(moving) == 0 || len(moving) != len(selected) || boundary == "" {
		return nil, history.ErrStale
	}
	index := slices.IndexFunc(rest, func(o canvas.Rect) bool { return o.ID == boundary })
	if front {
		index++
	}
	after := make([]canvas.Rect, 0, len(before))
	after = append(after, rest[:index]...)
	after = append(after, moving...)
	after = append(after, rest[index:]...)
	return after, nil
}

// Keep the pre-gesture presentation stable even if native services request a
// refresh during capture. Execution still checks live availability every time.
func (m *model) commandDisabled(id string) bool {
	if s := m.layerDrag; s != nil {
		if value, ok := s.disabled[id]; ok {
			return value
		}
	}
	return m.commands.Enabled(id) != nil
}

// Keep presentation stable during a layer capture; callbacks still check live admission.
func (m *model) controlDisabled(id string, current bool) bool {
	if s := m.layerDrag; s != nil {
		if value, ok := s.disabled["control:"+id]; ok {
			return value
		}
	}
	return current
}
