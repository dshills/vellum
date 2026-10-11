package cogent

import (
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"github.com/dshills/vellum/workbench"
)

// Extensions expose only materialized rows, plus previous/next page actions.
// IDs contain stable domain IDs and never expose reusable native slot identity.
func extensionAccessibility(r *record, b *accessibilityBridge, disabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	parent := accessibilityManagedID(r.node.ID)
	addPage := func(name string, delta int, fn func(int) bool) {
		id := accessibilityPageID(r.node.ID, name)
		actions := []string{}
		if !disabled {
			actions = []string{"press"}
		}
		nodes = append(nodes, accessibilityNode{ID: id, Parent: parent, Role: "button", Name: strings.ToUpper(name[:1]) + name[1:] + " page", Disabled: disabled, Actions: actions})
		if !disabled {
			handlers[id] = func(action accessibilityAction) bool {
				if action.Kind != "press" {
					return false
				}
				return fn(delta)
			}
		}
	}
	switch v := r.node.Custom.(type) {
	case *Workbench:
		return workbenchAccessibility(r, b, disabled)
	case *VirtualTable:
		if v == nil || v.Model == nil {
			return nodes, handlers
		}
		rows := v.Model.View()
		start := v.Model.StartRow()
		count := v.Model.VisibleCount()
		length := v.Model.Len()
		page := func(delta int) bool {
			if v.Model == nil || v.Model.VisibleCount() != count {
				return false
			}
			next := v.Model.StartRow() + delta
			maximum := v.Model.Len() - count
			if maximum < 0 {
				maximum = 0
			}
			if next < 0 {
				next = 0
			}
			if next > maximum {
				next = maximum
			}
			return v.Model.SetViewport(next, count) == nil
		}
		addPage("previous", -count, page)
		addPage("next", count, page)
		for index, row := range rows {
			row := row
			id := accessibilityRowID(r.node.ID, row.ID)
			actions := []string{}
			if !disabled {
				actions = []string{"press"}
			}
			nodes = append(nodes, accessibilityNode{ID: id, Parent: parent, Role: "row", DomainID: row.ID, Name: strings.Join(row.Cells, ", "), Value: strings.Join(row.Cells, ", "), Selected: v.Model.Selected == row.ID, Disabled: disabled, Actions: actions, RowIndex: start + index + 1, RowCount: length})
			if !disabled {
				handlers[id] = func(action accessibilityAction) bool {
					if action.Kind != "press" {
						return false
					}
					// Match the captured domain ID against current viewport before selecting.
					for slot, current := range v.Model.View() {
						if current.ID == row.ID {
							selected, err := v.Model.SelectVisible(slot)
							if err != nil {
								return false
							}
							if v.Select != nil {
								v.Select(selected)
							}
							return true
						}
					}
					return false
				}
			}
		}
	case *VirtualTree:
		if v == nil || v.Model == nil {
			return nodes, handlers
		}
		rows, err := v.Model.View(v.Start, v.Count)
		if err != nil {
			return nodes, handlers
		}
		page := func(delta int) bool {
			if v.Model == nil {
				return false
			}
			accessibilityTreePage(v, delta)
			return true
		}
		addPage("previous", -v.Count, page)
		addPage("next", v.Count, page)
		for index, row := range rows {
			row := row
			id := accessibilityRowID(r.node.ID, row.ID)
			actions := []string{}
			if !disabled {
				actions = []string{"press"}
				if row.HasChildren {
					actions = append(actions, "expand", "collapse")
				}
			}
			var expanded *bool
			if row.HasChildren {
				value := row.Expanded
				expanded = &value
			}
			nodes = append(nodes, accessibilityNode{ID: id, Parent: parent, Role: "tree-row", DomainID: row.ID, Name: row.Label, Value: row.Label, Selected: v.Model.Selected == row.ID, Disabled: disabled, Actions: actions, RowIndex: v.Start + index + 1, RowCount: v.Model.VisibleLen(), Level: row.Depth + 1, Expanded: expanded})
			if !disabled {
				allowed := append([]string(nil), actions...)
				handlers[id] = func(action accessibilityAction) bool {
					if !containsAccessibilityAction(allowed, action.Kind) {
						return false
					}
					return applyAccessibilityTreeAction(v, row.ID, action.Kind)
				}
			}
		}
	default:
		_ = b // Custom extensions without a semantic adapter remain named groups.
	}
	return nodes, handlers
}
func extensionAccessibilityBounds(r *record, out map[string][4]float32, bounds func(string, core.Widget), includeGrids bool) {
	switch v := r.node.Custom.(type) {
	case *Workbench:
		workbenchAccessibilityBounds(r, bounds, out, includeGrids)
	case *VirtualTable:
		if v == nil || v.host == nil {
			return
		}
		// Sync wrote stable row IDs to the property below only when publishing.
		for _, widget := range v.rows {
			if id, ok := widget.Property("vellum.accessibility-domain").(string); ok {
				bounds(accessibilityRowID(r.node.ID, id), widget)
			}
		}
		for _, child := range v.host.Children {
			for _, item := range child.AsTree().Children {
				if button, ok := item.(*core.Button); ok {
					switch button.Text {
					case "Previous page":
						bounds(accessibilityPageID(r.node.ID, "previous"), button)
					case "Next page":
						bounds(accessibilityPageID(r.node.ID, "next"), button)
					}
				}
			}
		}
	case *VirtualTree:
		if v == nil || v.host == nil {
			return
		}
		for _, widget := range v.rows {
			if id, ok := widget.Property("vellum.accessibility-domain").(string); ok {
				bounds(accessibilityRowID(r.node.ID, id), widget)
			}
		}
		for _, child := range v.host.Children {
			for _, item := range child.AsTree().Children {
				if button, ok := item.(*core.Button); ok {
					if page, ok := button.Property("vellum.accessibility-page").(string); ok {
						bounds(accessibilityPageID(r.node.ID, page), button)
					}
				}
			}
		}
	}
	_ = out
}

// accessibilityStampVirtualGeometry is called under the owner after each sync.
func accessibilityStampVirtualGeometry(root *record) {
	if root == nil {
		return
	}
	switch v := root.node.Custom.(type) {
	case *Workbench:
		accessibilityWorkbenchLeaves(v, func(g *workbench.Group) { accessibilityStampVirtualGeometry(v.records[g.Active]) })
	case *VirtualTable:
		if v != nil && v.Model != nil {
			rows := v.Model.View()
			for i, w := range v.rows {
				id := ""
				if i < len(rows) {
					id = rows[i].ID
				}
				w.SetProperty("vellum.accessibility-domain", id)
			}
		}
	case *VirtualTree:
		if v != nil {
			for i, w := range v.rows {
				id := ""
				if i < len(v.visible) {
					id = v.visible[i].ID
				}
				w.SetProperty("vellum.accessibility-domain", id)
			}
		}
	}
	for _, child := range root.children {
		accessibilityStampVirtualGeometry(child)
	}
}

// mountAccessibilityTreePager gives virtual outline paging a real native target.
// Mount invokes this after constructing the host and before its fixed row pool.
func mountAccessibilityTreePager(v *VirtualTree) {
	bar := core.NewFrame(v.host)
	bar.Styler(func(s *styles.Style) { s.Direction = styles.Row })
	for _, direction := range []int{-1, 1} {
		page, title := "previous", "Previous page"
		if direction > 0 {
			page, title = "next", "Next page"
		}
		delta := direction * v.Count
		button := core.NewButton(bar).SetText(title)
		button.SetProperty("vellum.accessibility-page", page)
		button.OnClick(func(events.Event) {
			if err := v.Owner.Do(func() {
				accessibilityTreePage(v, delta)
				_ = v.Owner.Invalidate(v.Window)
			}); err != nil && v.OnError != nil {
				v.OnError(err)
			}
		})
	}
}
func accessibilityTreePage(v *VirtualTree, delta int) {
	next := v.Start + delta
	maximum := v.Model.VisibleLen() - v.Count
	if maximum < 0 {
		maximum = 0
	}
	if next < 0 {
		next = 0
	}
	if next > maximum {
		next = maximum
	}
	v.Start = next
}

// Selection and disclosure are independent native actions. Read the current
// materialized row so stale callbacks cannot toggle a replaced leaf or slot.
func applyAccessibilityTreeAction(v *VirtualTree, id, kind string) bool {
	if v == nil || v.Model == nil {
		return false
	}
	rows, err := v.Model.View(v.Start, v.Count)
	if err != nil {
		return false
	}
	for _, row := range rows {
		if row.ID != id {
			continue
		}
		switch kind {
		case "press":
			if err := v.Model.Select(id); err != nil {
				return false
			}
			if v.Select != nil {
				v.Select(id)
			}
			return true
		case "expand", "collapse":
			if !row.HasChildren {
				return false
			}
			if row.Expanded == (kind == "expand") {
				return true
			}
			return v.Model.Toggle(id) == nil
		}
		return false
	}
	return false
}
