package cogent

import (
	"image"
	"slices"
)

// Frame ordinals alone do not change accessibility. Include all other completed
// tree provenance, data and geometry, even for same-generation native layout.
func sameTreeAccessibilityMeasurement(a, b *TreeInspection) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Window != b.Window || a.ID != b.ID ||
		a.Incarnation != b.Incarnation || a.RowIncarnation != b.RowIncarnation ||
		a.Generation != b.Generation || a.Stamp != b.Stamp || a.Viewport != b.Viewport ||
		a.Selected != b.Selected || a.SelectedVisible != b.SelectedVisible ||
		a.Pending != b.Pending || a.VisibleRows != b.VisibleRows ||
		a.TotalNodes != b.TotalNodes || a.RetainedRows != b.RetainedRows ||
		a.Scale != b.Scale || a.Bounds != b.Bounds {
		return false
	}
	return (a.Rows == nil) == (b.Rows == nil) && (a.Pages == nil) == (b.Pages == nil) &&
		slices.Equal(a.Rows, b.Rows) && slices.Equal(a.Pages, b.Pages)
}

// The existing Grid-prefixed protocol is a bounded virtual-view transaction.
// Outlines share its detached pending cache and native action/identity policy.
func stampTreeAccessibility(r *record, nodes []accessibilityNode) {
	if len(nodes) == 0 {
		return
	}
	t := r.control.outline
	m := t.currentMeasurement()
	pending := m == nil || !t.live() || m.Incarnation != t.incarnation || m.RowIncarnation != t.epoch
	frames := map[string][4]float32{}
	if !pending {
		put := func(id string, rect image.Rectangle) {
			frames[id] = [4]float32{float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy())}
		}
		put(nodes[0].ID, m.Bounds)
		for _, row := range m.Rows {
			if !row.Bounds.Empty() {
				put(accessibilityManagedID(row.ID), row.Bounds)
			}
		}
		for _, page := range m.Pages {
			if !page.Bounds.Empty() {
				put(accessibilityManagedID(page.ID), page.Bounds)
			}
		}
		for _, node := range nodes {
			if _, ok := frames[node.ID]; !ok {
				pending = true
				break
			}
		}
	}
	for i := range nodes {
		n := &nodes[i]
		n.GridScope = nodes[0].ID
		n.GridIncarnation = t.incarnation
		n.GridPending = pending
		n.GridTargetGeneration = t.projection.State.Generation
		n.GridGeneration = 0
		if !pending {
			n.GridGeneration = m.Generation
			frame := frames[n.ID]
			n.GridFrame = &frame
		}
	}
}

// Admission may precede a later layout or model change. Recheck completed
// geometry at owner delivery, including the host's generic focus action.
func guardTreeAccessibility(t *outlineBinding, nodes []accessibilityNode, handlers map[string]accessibilityHandler) {
	for _, node := range nodes {
		if handler := handlers[node.ID]; handler != nil {
			generation, incarnation, pending := node.GridGeneration, node.GridIncarnation, node.GridPending
			handlers[node.ID] = func(action accessibilityAction) bool {
				if pending || !t.eligible() {
					return false
				}
				state, err := t.model.State()
				if err != nil || state.Generation != generation {
					return false
				}
				m := t.currentMeasurement()
				if m == nil || m.Incarnation != incarnation || m.RowIncarnation != t.epoch || m.Generation != generation || m.Stamp != state.Stamp || m.Viewport != state.Viewport || m.Pending != state.Pending {
					return false
				}
				return handler(action)
			}
		}
	}
}
