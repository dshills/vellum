package cogent

import (
	"image"

	"github.com/dshills/vellum/ui"
)

func gridAccessibility(r *record, parent string, b *accessibilityBridge, disabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	g := r.control.grid
	m := g.currentMeasurement()
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	if m == nil || m.Generation != g.projection.State.Generation {
		return nodes, handlers
	}
	rows := map[string]string{}
	for i, c := range m.Cells {
		if c.Bounds.Empty() {
			continue
		}
		rowID := accessibilityRowID(r.node.ID, c.RowKey)
		if rows[c.RowKey] == "" {
			rows[c.RowKey] = rowID
			nodes = append(nodes, accessibilityNode{ID: rowID, DomainID: gridRowID(r.node.ID, c.RowKey), Parent: parent, Role: "grid-row", Name: c.RowKey, RowIndex: c.Row + 1, RowCount: m.Stamp.Rows, Disabled: disabled, Actions: []string{}})
		}
		id := accessibilityManagedID(c.ID)
		x := accessibilityNode{pressFocusID: parent, ID: id, DomainID: c.ID, Parent: rowID, Role: "grid-cell", Name: c.Text, Value: c.Text, Disabled: disabled, Selected: c.Active, RowIndex: c.Row + 1, ColumnIndex: c.Column + 1, Actions: []string{}}
		if !disabled {
			x.Actions = []string{"press"}
			ref := g.projection.Cells[i].Ref
			handlers[id] = func(a accessibilityAction) bool {
				if a.Kind != "press" || !g.eligible() {
					return false
				}
				before, e := g.model.State()
				if e != nil {
					return false
				}
				if e = g.model.Activate(ref); e != nil {
					g.fail(e)
					return false
				}
				after, e := g.model.State()
				if e != nil {
					return false
				}
				if before.Generation != after.Generation {
					if cb := r.node.GridChange; cb != nil {
						cb(ui.GridEvent{ID: r.node.ID, Reason: ui.GridSelection, State: after})
					}
					_ = g.owner.Invalidate(g.window)
				}
				if b != nil {
					accessibilityFocus(b, parent)
				}
				return true
			}
		}
		nodes = append(nodes, x)
	}
	for _, h := range m.Headers {
		if h.Bounds.Empty() {
			continue
		}
		nodes = append(nodes, accessibilityNode{ID: accessibilityManagedID(h.ID), DomainID: h.ID, Parent: parent, Role: "grid-column", Name: h.Title, Value: h.Title, ColumnIndex: h.Column + 1, Disabled: disabled, Actions: []string{}})
	}
	return nodes, handlers
}
func gridAccessibilityBounds(r *record, out map[string][4]float32) {
	if r.control == nil || r.control.grid == nil {
		return
	}
	g := r.control.grid
	m := g.currentMeasurement()
	if m == nil || m.Generation != g.projection.State.Generation {
		return
	}
	put := func(id string, b image.Rectangle) {
		if !b.Empty() {
			out[id] = [4]float32{float32(b.Min.X), float32(b.Min.Y), float32(b.Dx()), float32(b.Dy())}
		}
	}
	rows := map[string]image.Rectangle{}
	for _, c := range m.Cells {
		put(accessibilityManagedID(c.ID), c.Bounds)
		rows[c.RowKey] = rows[c.RowKey].Union(c.Bounds)
	}
	for key, b := range rows {
		put(accessibilityRowID(r.node.ID, key), b)
	}
	for _, h := range m.Headers {
		put(accessibilityManagedID(h.ID), h.Bounds)
	}
}
