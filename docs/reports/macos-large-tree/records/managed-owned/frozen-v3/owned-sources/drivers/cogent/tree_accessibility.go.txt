package cogent

import (
	"image"

	"github.com/dshills/vellum/ui"
)

// The AX record is built from the accepted bounded projection and completed
// native geometry. It never asks TreeData or a pending index builder for rows.
func treeAccessibility(r *record, parent string, b *accessibilityBridge, disabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	t := r.control.outline
	m := t.currentMeasurement()
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	if m == nil || !t.live() || !t.eligible() && !disabled {
		return nodes, handlers
	}
	for i, row := range m.Rows {
		if row.Bounds.Empty() || i >= len(t.projection.Rows) {
			continue
		}
		captured := t.projection.Rows[i]
		id := accessibilityManagedID(row.ID)
		x := accessibilityNode{ID: id, DomainID: row.ID, Parent: parent, Role: "tree-row", Name: row.Label, Value: row.Label, Disabled: disabled, Selected: row.Selected, Level: row.Depth + 1, RowIndex: row.Index + 1, RowCount: m.VisibleRows, pressFocusID: parent, Actions: []string{}}
		if row.HasChildren {
			expanded := row.Expanded
			x.Expanded = &expanded
		}
		if !disabled {
			x.Actions = append(x.Actions, "press")
			if row.HasChildren && r.node.TreeDisclosure != nil {
				x.Actions = append(x.Actions, "expand", "collapse")
			}
			handlers[id] = func(a accessibilityAction) bool {
				if !t.eligible() {
					return false
				}
				current, err := t.model.Row(captured.Ref)
				if err != nil || current.ID != captured.ID {
					return false
				}
				switch a.Kind {
				case "press":
					t.apply(ui.TreeSelection, true, func() error { return t.model.SelectRef(captured.Ref) })
				case "expand", "collapse":
					if !current.HasChildren || r.node.TreeDisclosure == nil {
						return false
					}
					t.disclose(current, a.Kind == "expand")
				default:
					return false
				}
				if b != nil {
					accessibilityFocus(b, parent)
				}
				return true
			}
		}
		nodes = append(nodes, x)
	}
	for _, page := range m.Pages {
		if page.Bounds.Empty() {
			continue
		}
		id := accessibilityManagedID(page.ID)
		refused := disabled || !page.Enabled
		x := accessibilityNode{ID: id, DomainID: page.ID, Parent: parent, Role: "button", Name: page.Name, Disabled: refused, pressFocusID: parent, Actions: []string{}}
		if !refused {
			x.Actions = []string{"press"}
			generation := m.Generation
			direction := page.Direction
			handlers[id] = func(a accessibilityAction) bool {
				if a.Kind != "press" || !t.eligible() {
					return false
				}
				state, err := t.model.State()
				if err != nil || state.Generation != generation {
					return false
				}
				t.apply(ui.TreeViewport, true, func() error { return t.pageViewport(direction) })
				if b != nil {
					accessibilityFocus(b, parent)
				}
				return true
			}
		}
		nodes = append(nodes, x)
	}
	return nodes, handlers
}
func treeAccessibilityBounds(r *record, out map[string][4]float32) {
	if r.control == nil || r.control.outline == nil {
		return
	}
	t := r.control.outline
	m := t.currentMeasurement()
	if m == nil || !t.live() {
		return
	}
	put := func(id string, b image.Rectangle) {
		if !b.Empty() {
			out[accessibilityManagedID(id)] = [4]float32{float32(b.Min.X), float32(b.Min.Y), float32(b.Dx()), float32(b.Dy())}
		}
	}
	for _, row := range m.Rows {
		put(row.ID, row.Bounds)
	}
	for _, page := range m.Pages {
		put(page.ID, page.Bounds)
	}
}
