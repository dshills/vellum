package cogent

import (
	"image"

	"github.com/dshills/vellum/virtual"
)

// The native action may already be queued when render withdraws geometry but
// before the owner publishes pending. Check readiness again at owner delivery.
func guardGridAccessibility(g *gridBinding, nodes []accessibilityNode, handlers map[string]accessibilityHandler) {
	for _, node := range nodes {
		if handler := handlers[node.ID]; handler != nil {
			generation, incarnation := node.GridGeneration, node.GridIncarnation
			pending, cell := node.GridPending, node.Role == "grid-cell"
			handlers[node.ID] = func(action accessibilityAction) bool {
				if pending || !g.eligible() {
					return false
				}
				state, err := g.model.State()
				if err != nil {
					return false
				}
				if state.Generation != generation {
					// Cell refs already reported this typed refusal before the
					// publication guard was added. Preserve that error callback.
					if cell {
						g.fail(virtual.ErrGridStale)
					}
					return false
				}
				m := g.currentMeasurement()
				if m == nil || m.Generation != generation || m.Incarnation != incarnation || m.CellIncarnation != g.epoch {
					return false
				}
				if !currentGridAccessibilityEditor(g, m) {
					return false
				}
				if state.Stamp != m.Stamp || state.Viewport != m.Viewport {
					return false
				}
				return handler(action)
			}
		}
	}
}

// Private marked previews can change the owner-held field cache without a new
// portable grid generation. They also revoke actions from the older frame.
func currentGridAccessibilityEditor(g *gridBinding, m *GridInspection) bool {
	edit := gridEditingRecord(g.r)
	if (edit == nil) != (m.Editor == nil) {
		return false
	}
	if edit == nil {
		return true
	}
	return g.editing != nil && m.Editor.ID == edit.node.ID && m.Editor.Session == g.editing.session && edit.control.fieldState != nil && edit.control.fieldState.Text == m.Editor.Text
}

// stampGridAccessibility runs on the owner. The frame was copied by render;
// a pending layout never reads live widget rectangles or accepts old actions.
func stampGridAccessibility(r *record, nodes []accessibilityNode) {
	if len(nodes) == 0 {
		return
	}
	g := r.control.grid
	m := g.currentMeasurement()
	pending := m == nil || !g.live() || !g.projectionReady || m.Incarnation != g.incarnation || m.CellIncarnation != g.epoch
	if !pending {
		edit := gridEditingRecord(r)
		if !currentGridAccessibilityEditor(g, m) {
			pending = true
		} else if edit != nil {
			for _, node := range nodes {
				if node.ID == accessibilityManagedID(edit.node.ID) && (node.Value != m.Editor.Text || node.Focused != m.Editor.Focused) {
					pending = true
				}
			}
		}
	}
	frames := map[string][4]float32{}
	if !pending {
		put := func(id string, rect image.Rectangle) {
			frames[id] = [4]float32{float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy())}
		}
		put(nodes[0].ID, m.Bounds)
		rows := map[string]image.Rectangle{}
		for _, c := range m.Cells {
			put(accessibilityManagedID(c.ID), c.Bounds)
			rows[c.RowKey] = rows[c.RowKey].Union(c.Bounds)
		}
		for key, bounds := range rows {
			put(accessibilityRowID(r.node.ID, key), bounds)
		}
		for _, h := range m.Headers {
			put(accessibilityManagedID(h.ID), h.Bounds)
		}
		if m.Editor != nil {
			put(accessibilityManagedID(m.Editor.ID), m.Editor.Bounds)
		}
		// An editor mounted after capture also needs a completed frame. Avoid
		// combining current semantics with an earlier editor's geometry.
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
		n.GridIncarnation = g.incarnation
		n.GridPending = pending
		n.GridTargetGeneration = g.projection.State.Generation
		n.GridGeneration = 0
		if !pending {
			n.GridGeneration = m.Generation
			frame := frames[n.ID]
			n.GridFrame = &frame
		}
	}
}

func cloneAccessibilityNodes(nodes []accessibilityNode) []accessibilityNode {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]accessibilityNode, len(nodes))
	for i := range out {
		out[i] = cloneAccessibilityNode(nodes[i])
	}
	return out
}

func cloneAccessibilityNode(n accessibilityNode) accessibilityNode {
	n.Actions = append([]string{}, n.Actions...)
	if n.GridFrame != nil {
		value := *n.GridFrame
		n.GridFrame = &value
	}
	if n.TextSelection != nil {
		value := *n.TextSelection
		n.TextSelection = &value
	}
	if n.Expanded != nil {
		value := *n.Expanded
		n.Expanded = &value
	}
	if n.Min != nil {
		value := *n.Min
		n.Min = &value
	}
	if n.Max != nil {
		value := *n.Max
		n.Max = &value
	}
	return n
}

// prepareGridPublication retains one completed subtree per currently visible
// grid, scoped by binding incarnation and parent. It never retains handlers or
// a grid/model pointer. Actual removal drops the cache immediately.
func (b *accessibilityBridge) prepareGridPublication(nodes []accessibilityNode, handlers map[string]accessibilityHandler) ([]accessibilityNode, map[string]accessibilityHandler) {
	return b.prepareGridPublicationSnapshot(nodes, handlers, false)
}

// owned admits only freshly created snapshot containers whose caller relinquishes
// them. Borrowed callers still receive separate output and handler containers.
func (b *accessibilityBridge) prepareGridPublicationSnapshot(nodes []accessibilityNode, handlers map[string]accessibilityHandler, owned bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed.Load() {
		return nil, nil
	}
	// Group input indices, rather than repeatedly copying large node records
	// while each scope grows. Completed cache records are still fully detached.
	scopes := map[string][]int{}
	for index, node := range nodes {
		if node.GridScope != "" {
			scopes[node.GridScope] = append(scopes[node.GridScope], index)
		}
	}
	cache := map[string][]accessibilityNode{}
	replacements := map[string][]accessibilityNode{}
	actions := handlers
	if !owned {
		actions = make(map[string]accessibilityHandler, len(handlers))
		for id, handler := range handlers {
			actions[id] = handler
		}
	}
	for scope, indices := range scopes {
		host := nodes[indices[0]]
		if host.ID != scope {
			continue // Framework traversal always emits the table host first.
		}
		if !host.GridPending {
			group := make([]accessibilityNode, len(indices))
			for i, index := range indices {
				group[i] = cloneAccessibilityNode(nodes[index])
			}
			cache[scope] = group
			continue
		}
		previous := b.gridPublications[scope]
		retained := nodes[indices[0] : indices[0]+1]
		if len(previous) > 0 && previous[0].GridIncarnation == host.GridIncarnation && previous[0].Parent == host.Parent && previous[0].Role == host.Role {
			cache[scope] = previous
			retained = previous
		}
		retained = cloneAccessibilityNodes(retained)
		for i := range retained {
			n := &retained[i]
			n.GridPending = true
			n.GridTargetGeneration = host.GridTargetGeneration
			n.Disabled, n.Focused = true, false
			n.Actions = []string{}
			n.TextClient, n.TextSelection, n.TextFocus = 0, nil, 0
			delete(actions, n.ID)
		}
		// Descriptor labels/help remain current even while cell data are read-only.
		retained[0].Name, retained[0].Help = host.Name, host.Help
		for _, index := range indices {
			delete(actions, nodes[index].ID)
		}
		replacements[scope] = retained
	}
	if owned && len(replacements) == 0 {
		b.gridPublications = cache
		return nodes, actions
	}
	out := make([]accessibilityNode, 0, len(nodes))
	for _, node := range nodes {
		if replacement, pending := replacements[node.GridScope]; pending {
			if node.ID == node.GridScope {
				out = append(out, replacement...)
			}
			continue
		}
		out = append(out, node)
	}
	b.gridPublications = cache
	return out, actions
}
