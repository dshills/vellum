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
	if m.clipboardAvailable() != nil || g.scale <= 0 || math.IsNaN(g.scale) || math.IsInf(g.scale, 0) {
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
	if primary == "" {
		return nil
	}
	selection := m.selectedIDs()
	ids := m.expandedSelection([]string{primary})
	if slices.Contains(selection, primary) {
		ids = slices.Clone(selection)
		primary = m.scene.Selected
	}
	s := &layerDrag{source: source, primary: primary, ids: ids, selection: selection, priorPrimary: m.scene.Selected, before: slices.Clone(m.scene.Objects), groups: maps.Clone(m.groups), layout: layerLayout{clip: g.clip, scale: g.scale, rows: slices.Clone(g.rows)}, start: p, pointer: p}
	s.disabled = make(map[string]bool)
	for _, c := range m.commands.Search("") {
		s.disabled[c.ID] = m.commands.Enabled(c.ID) != nil
	}
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
	if g.clip != s.layout.clip || g.scale != s.layout.scale || !slices.Equal(g.rows, s.layout.rows) || !slices.Equal(m.scene.Objects, s.before) || !maps.Equal(m.groups, s.groups) || m.scene.Selected != s.priorPrimary || !slices.Equal(m.selectedIDs(), s.selection) {
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
		// A group marker's lower half targets its backmost member, not the
		// marker's own bottom. Require that actual boundary to be visible.
		if s.valid && !s.front {
			for _, o := range s.before {
				if s.groups[o.ID] != s.target {
					continue
				}
				s.valid = false
				for _, row := range g.rows {
					if row.id == o.ID && row.bounds.Min.Y >= g.clip.Min.Y && row.bounds.Max.Y <= g.clip.Max.Y {
						s.line, s.valid = row.bounds.Max.Y, true
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
