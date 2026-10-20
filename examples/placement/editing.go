package main

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
)

var errNoSelection = errors.New("select a component first")

// A move keeps one detached before image, never a history entry per pointer event.
type objectMove struct {
	id              string
	before          []canvas.Rect
	start, grab     canvas.Point
	geometry        placement.Geometry
	dragging, valid bool
	resize          bool
	corner          resizeCorner
	selection       []string
	selectionOnly   bool
	marquee         bool
	marqueeBounds   canvas.Rect
}

func (m *model) editAvailable() error {
	if m.fileBusy() || m.prompt != "" || m.move != nil || m.palette.State().Active {
		return errors.New("finish the current operation first")
	}
	if len(m.selectedIDs()) == 0 {
		return errNoSelection
	}
	return nil
}
func (m *model) selectedIndex() int {
	for i, o := range m.scene.Objects {
		if o.ID == m.scene.Selected {
			return i
		}
	}
	return -1
}

func (m *model) applyObjects(label string, after []canvas.Rect, selected string) error {
	priorIDs := m.selectedIDs()
	nextIDs := []string{selected}
	if selected == "" {
		nextIDs = nil
	} else if selected == m.scene.Selected {
		nextIDs = priorIDs
	}
	return m.applyDocument(label, after, prunedGroups(m.groups, after), nextIDs, selected)
}

func (m *model) deleteSelected() error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	ids := m.selectedIDs()
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	after := make([]canvas.Rect, 0, len(m.scene.Objects)-len(ids))
	for _, o := range m.scene.Objects {
		if !selected[o.ID] {
			after = append(after, o)
		}
	}
	if err := m.applyObjects("Delete selection", after, ""); err != nil {
		return err
	}
	m.status = fmt.Sprintf("Deleted %d selected components", len(ids))
	return nil
}

func (m *model) nudge(dx, dy float64) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	step := float64(1)
	if m.snap {
		step = 20
	}
	ids := m.selectedIDs()
	after, err := movedSelection(m.scene.Objects, ids, dx*step, dy*step)
	if err != nil {
		return err
	}
	if err = m.applyObjects("Move selection", after, m.scene.Selected); err != nil {
		return err
	}
	m.status = fmt.Sprintf("Moved %d selected components", len(ids))
	return nil
}
func validPosition(x, y float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsInf(x, 0) && !math.IsInf(y, 0) && math.Abs(x) <= 10000 && math.Abs(y) <= 10000
}

func (m *model) beginMove(g placement.Geometry, p canvas.Point) *objectMove {
	if m.fileBusy() || m.prompt != "" || m.move != nil || m.palette.State().Active {
		return nil
	}
	world, err := g.World(p)
	if err != nil {
		return nil
	}
	corner := noCorner
	id := ""
	ids := m.selectedIDs()
	if bounds, ok := selectionBounds(m.scene.Objects, ids); ok {
		corner = handleAt(bounds, g, p)
		if corner != noCorner {
			id = m.scene.Selected
		}
	}
	if id == "" {
		id = m.scene.ObjectAt(m.scene.View.Screen(world))
	}
	if id == "" {
		return nil
	}
	if !slices.Contains(ids, id) {
		ids = []string{id}
	}
	m.setSelection(ids, id)
	m.scene.View = g.View
	s := &objectMove{id: id, before: slices.Clone(m.scene.Objects), start: p, grab: world, geometry: g, valid: true, resize: corner != noCorner, corner: corner, selection: m.selectedIDs()}
	m.move = s
	return s
}
func (m *model) previewMove(s *objectMove, p canvas.Point) {
	if m.move != s || s.selectionOnly {
		return
	}
	world, err := s.geometry.World(p)
	s.valid = err == nil
	if !s.valid {
		return
	}
	if math.Hypot(p.X-s.start.X, p.Y-s.start.Y) < 6*s.geometry.Scale && !s.dragging {
		return
	}
	s.dragging = true
	if s.marquee {
		s.marqueeBounds = marqueeRect(s.grab, world)
		m.status = fmt.Sprintf("Marquee: %d components • release to select • Escape cancels", len(m.expandedSelection(marqueeIDs(s.before, s.selection, s.marqueeBounds))))
		return
	}
	// Keep the initial world grab point while mapping the current pointer with
	// the latest layout, preserving the offset inside the object.
	dx, dy := world.X-s.grab.X, world.Y-s.grab.Y
	if s.resize && len(s.selection) > 1 {
		after, err := resizedSelection(s.before, s.selection, s.corner, dx, dy, m.snap)
		if err != nil {
			s.valid = false
			return
		}
		m.scene.Objects = after
		m.status = fmt.Sprintf("Resizing %d selected components • Escape cancels", len(s.selection))
		return
	}
	after := slices.Clone(s.before)
	for i := range after {
		if after[i].ID != s.id {
			continue
		}
		if s.resize {
			after[i] = resizedCorner(after[i], s.corner, dx, dy, m.snap)
			if !validPosition(after[i].X, after[i].Y) {
				s.valid = false
				return
			}
			m.scene.Objects = after
			m.status = fmt.Sprintf("Resizing %s to %.0f × %.0f • Escape cancels", s.id, after[i].Width, after[i].Height)
			return
		}
		if m.snap {
			dx = math.Round((after[i].X+dx)/20)*20 - after[i].X
			dy = math.Round((after[i].Y+dy)/20)*20 - after[i].Y
		}
		moved, err := movedSelection(s.before, s.selection, dx, dy)
		if err != nil {
			s.valid = false
			return
		}
		m.scene.Objects = moved
		m.status = fmt.Sprintf("Moving %d selected components • Escape cancels", len(s.selection))
		return
	}
}
func (m *model) cancelMove(s *objectMove) {
	if m.move != s || s == nil {
		return
	}
	m.scene.Objects = slices.Clone(s.before)
	m.setSelection(s.selection, s.id)
	m.move = nil
	m.status = s.operation() + " canceled"
}
func (m *model) finishMove(s *objectMove, p canvas.Point) error {
	if m.move != s {
		return nil
	}
	m.previewMove(s, p)
	if !s.valid {
		m.cancelMove(s)
		return nil
	}
	if s.marquee {
		m.move = nil
		if s.dragging {
			m.setSelection(marqueeIDs(s.before, s.selection, s.marqueeBounds), s.id)
			m.status = fmt.Sprintf("Marquee selected %d components", len(m.selectedIDs()))
		}
		return nil
	}
	after := slices.Clone(m.scene.Objects)
	m.scene.Objects = slices.Clone(s.before)
	m.move = nil
	m.setSelection(s.selection, s.id)
	if s.selectionOnly {
		return nil
	}
	if !s.dragging {
		if !s.resize {
			m.setSelection([]string{s.id}, s.id)
		}
		return nil
	}
	if slices.Equal(after, s.before) {
		return nil
	}
	if err := m.applyObjects(s.operation()+" "+s.id, after, s.id); err != nil {
		return err
	}
	m.status = s.operation() + " completed: " + s.id
	return nil
}

// Layout can settle after focus/disabled styling. Refresh its mapping and clip
// while retaining the initial world grab point for translation.
// Camera or display-scale changes invalidate a gesture rather than mixing maps.
func (m *model) updateMoveGeometry(s *objectMove, g placement.Geometry) bool {
	if m.move != s || s == nil {
		return false
	}
	before := s.geometry
	if g.Window != before.Window || g.Scale != before.Scale || g.View.Zoom != before.View.Zoom || g.View.Center != before.View.Center {
		m.cancelMove(s)
		return false
	}
	s.geometry = g
	m.scene.View = g.View
	return true
}

func (s *objectMove) operation() string {
	if s.marquee {
		return "Marquee selection"
	}
	if s.resize {
		return "Resize"
	}
	return "Move"
}

const minComponentSize, maxComponentSize = 8.0, 2000.0

func validSize(width, height float64) bool {
	return !math.IsNaN(width) && !math.IsNaN(height) && width >= minComponentSize && height >= minComponentSize && width <= maxComponentSize && height <= maxComponentSize
}

// Only changed axes snap; resizing width must not change a component's height.
// Clamp to positive bounded sizes so crossing the anchor never flips an object.
func resizedDimensions(before canvas.Rect, dx, dy float64, snap bool) (float64, float64) {
	axis := func(size, delta float64) float64 {
		if delta == 0 {
			return size
		}
		size += delta
		if snap {
			size = math.Round(size/20) * 20
		}
		return math.Max(minComponentSize, math.Min(maxComponentSize, size))
	}
	return axis(before.Width, dx), axis(before.Height, dy)
}

func (m *model) resizeSelected(dw, dh float64) error {
	if err := m.resizeAvailable(); err != nil {
		return err
	}
	if math.IsNaN(dw) || math.IsNaN(dh) || math.IsInf(dw, 0) || math.IsInf(dh, 0) {
		return errors.New("invalid resize delta")
	}
	step := 1.0
	if m.snap {
		step = 20
	}
	i := m.selectedIndex()
	after := slices.Clone(m.scene.Objects)
	after[i].Width, after[i].Height = resizedDimensions(after[i], dw*step, dh*step, m.snap)
	if err := m.applyObjects("Resize "+after[i].ID, after, after[i].ID); err != nil {
		return err
	}
	m.status = fmt.Sprintf("Resized %s to %.0f × %.0f", after[i].ID, after[i].Width, after[i].Height)
	return nil
}
