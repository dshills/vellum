package cogent

import (
	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func accessibilityWorkbenchNamespace(extension, panel string) string {
	return "workbench:" + accessibilityIDPart(extension) + ":" + accessibilityIDPart(panel) + ":"
}
func accessibilityDockID(extension, group string) string {
	return "dock:" + accessibilityIDPart(extension) + ":" + accessibilityIDPart(group)
}
func accessibilityDockOptionID(extension, panel string) string {
	return "dock-option:" + accessibilityIDPart(extension) + ":" + accessibilityIDPart(panel)
}
func accessibilityWorkbenchLeaves(w *Workbench, visit func(*workbench.Group)) {
	if w == nil || w.layoutSnapshot == nil {
		return
	}
	var walk func(*workbench.Group)
	walk = func(g *workbench.Group) {
		if g == nil {
			return
		}
		if g.Axis != "" {
			walk(g.First)
			walk(g.Second)
			return
		}
		visit(g)
	}
	walk(w.layoutSnapshot)
}
func namespaceAccessibility(nodes []accessibilityNode, handlers map[string]accessibilityHandler, prefix string, b *accessibilityBridge) ([]accessibilityNode, map[string]accessibilityHandler) {
	if prefix == "" {
		return nodes, handlers
	}
	remapped := map[string]accessibilityHandler{}
	for i := range nodes {
		old := nodes[i].ID
		nodes[i].ID = prefix + old
		if nodes[i].GridScope != "" {
			nodes[i].GridScope = prefix + nodes[i].GridScope
		}
		if nodes[i].Parent != "" {
			nodes[i].Parent = prefix + nodes[i].Parent
		}
		if handler := handlers[old]; handler != nil {
			id := nodes[i].ID
			canFocus := nodes[i].Focusable && !nodes[i].Disabled && containsAccessibilityAction(nodes[i].Actions, "focus")
			handler := handler
			pressFocus := nodes[i].pressFocusID
			if pressFocus != "" {
				nodes[i].pressFocusID = prefix + pressFocus
			}
			remapped[id] = func(action accessibilityAction) bool {
				if action.Kind == "focus" {
					// Preserve the original permission/lifetime validation before
					// replacing its pending domain focus with the qualified ID.
					if !canFocus || !handler(action) {
						return false
					}
					accessibilityFocus(b, id)
					return true
				}
				accepted := handler(action)
				if accepted && action.Kind == "press" && pressFocus != "" && b != nil {
					// Only Grid cell metadata permits this rewrite, and only when
					// its validated handler published the exact raw host target.
					if pending := b.pendingFocus.Load(); pending != nil && *pending == pressFocus {
						accessibilityFocus(b, prefix+pressFocus)
					}
				}
				return accepted
			}
		}
	}
	return nodes, remapped
}
func workbenchAccessibility(r *record, b *accessibilityBridge, disabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	w, ok := r.node.Custom.(*Workbench)
	if !ok || w == nil {
		return nil, nil
	}
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	generation := w.generation.Load()
	accessibilityWorkbenchLeaves(w, func(group *workbench.Group) {
		groupID := accessibilityDockID(r.node.ID, group.ID)
		actions := []string{}
		if !disabled {
			actions = []string{"focus"}
		}
		focused := false
		for _, panel := range group.Tabs {
			if button := w.buttons[panel]; button != nil && button.ContainsFocus() {
				focused = true
				break
			}
		}
		nodes = append(nodes, accessibilityNode{ID: groupID, Parent: accessibilityManagedID(r.node.ID), Role: "tabs", DomainID: group.ID, Name: "Dock tabs", Focusable: true, Focused: focused, Disabled: disabled, Actions: actions})
		if !disabled {
			handlers[groupID] = func(action accessibilityAction) bool {
				if action.Kind != "focus" || w.generation.Load() != generation {
					return false
				}
				accessibilityFocus(b, groupID)
				return true
			}
		}
		for _, panelID := range group.Tabs {
			title := panelID
			for _, panel := range w.Panels {
				if panel.ID == panelID {
					title = panel.Title
					break
				}
			}
			id := accessibilityDockOptionID(r.node.ID, panelID)
			optionActions := []string{}
			if !disabled {
				optionActions = []string{"press"}
			}
			nodes = append(nodes, accessibilityNode{ID: id, Parent: groupID, Role: "option", DomainID: panelID, Name: title, Value: group.Active == panelID, Selected: group.Active == panelID, Disabled: disabled, Actions: optionActions})
			if !disabled {
				panelID := panelID
				groupID := group.ID
				handlers[id] = func(action accessibilityAction) bool {
					if action.Kind != "press" || w.generation.Load() != generation {
						return false
					}
					var err error
					if w.Group != nil {
						err = w.Group.Apply(func(model *workbench.Session) error { return model.Activate(w.SessionWindow, groupID, panelID) })
					} else if w.Workspace != nil {
						err = w.Workspace.Activate(groupID, panelID)
					} else {
						return false
					}
					if err != nil {
						if w.OnError != nil {
							w.OnError(err)
						}
						return false
					}
					return true
				}
			}
		}
		if active := w.records[group.Active]; active != nil {
			childNodes, childHandlers := accessibilitySnapshotInherited(active, b, disabled)
			prefix := accessibilityWorkbenchNamespace(r.node.ID, group.Active)
			childNodes, childHandlers = namespaceAccessibility(childNodes, childHandlers, prefix, b)
			for i := range childNodes {
				if childNodes[i].Parent == "" {
					childNodes[i].Parent = groupID
				}
				if disabled {
					childNodes[i].Disabled = true
					childNodes[i].Actions = []string{}
					delete(childHandlers, childNodes[i].ID)
				}
			}
			nodes = append(nodes, childNodes...)
			for id, handler := range childHandlers {
				handlers[id] = handler
			}
		}
	})
	return nodes, handlers
}
func workbenchAccessibilityBounds(r *record, bounds func(string, core.Widget), out map[string][4]float32, includeScoped bool) {
	w, ok := r.node.Custom.(*Workbench)
	if !ok || w == nil {
		return
	}
	accessibilityWorkbenchLeaves(w, func(group *workbench.Group) {
		bounds(accessibilityDockID(r.node.ID, group.ID), w.tabs[group.ID])
		for _, panel := range group.Tabs {
			bounds(accessibilityDockOptionID(r.node.ID, panel), w.buttons[panel])
		}
		if active := w.records[group.Active]; active != nil {
			prefix := accessibilityWorkbenchNamespace(r.node.ID, group.Active)
			for id, rect := range accessibilityBoundsIncludingScoped(active, includeScoped) {
				out[prefix+id] = rect
			}
		}
	})
}

// visibleAccessibilityScope includes managed records inside the active dock
// panels. Their IDs are qualified per extension/panel, including modal scopes.
func visibleAccessibilityScope(root *record) (*record, string) {
	selected := root
	namespace := ""
	var visit func(*record, string, bool)
	visit = func(r *record, prefix string, disabled bool) {
		if r == nil {
			return
		}
		disabled = disabled || r.node.Disabled
		if disabled || r.node.Kind == ui.Dialog && (!r.node.Open || r.control != nil && r.control.dialog != nil && r.control.dialog.dismissed) {
			return
		}
		if r.node.Kind == ui.Dialog || menuOpen(r) {
			selected = r
			namespace = prefix
		}
		if w, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok {
			accessibilityWorkbenchLeaves(w, func(g *workbench.Group) {
				visit(w.records[g.Active], prefix+accessibilityWorkbenchNamespace(r.node.ID, g.Active), disabled)
			})
		}
		for _, c := range r.children {
			if r.node.Kind == ui.Disclosure && !r.node.Checked || r.node.Kind == ui.Tabs && c.node.ID != r.node.SelectedID {
				continue
			}
			visit(c, prefix, disabled)
		}
	}
	visit(root, "", false)
	return selected, namespace
}

// focusProjection uses the same collision-free identifiers as the native tree.
// Portable FocusOrder can then include active dock content and modal isolation.
func accessibilityFocusProjection(root *record, prefix string) ui.Node {
	if root == nil {
		return ui.Node{}
	}
	n := root.node
	n.Options = append([]ui.Option(nil), n.Options...)
	n.ID = prefix + accessibilityManagedID(n.ID)
	if n.Kind == ui.Dialog && root.control != nil && root.control.dialog != nil && root.control.dialog.dismissed {
		n.Open = false
	}
	n.Children = nil
	if w, ok := root.node.Custom.(*Workbench); n.Kind == ui.Extension && ok {
		n.Kind = ui.Column
		accessibilityWorkbenchLeaves(w, func(g *workbench.Group) {
			tabs := ui.Node{ID: prefix + accessibilityDockID(root.node.ID, g.ID), Kind: ui.Tabs, Disabled: n.Disabled}
			for _, panel := range g.Tabs {
				id := prefix + accessibilityDockOptionID(root.node.ID, panel)
				tabs.Options = append(tabs.Options, ui.Option{ID: id})
				page := accessibilityFocusProjection(w.records[panel], prefix+accessibilityWorkbenchNamespace(root.node.ID, panel))
				tabs.Children = append(tabs.Children, ui.Node{ID: id, Kind: ui.Column, Children: []ui.Node{page}})
				if panel == g.Active {
					tabs.SelectedID = id
				}
			}
			n.Children = append(n.Children, tabs)
		})
		return n
	}
	if edit := gridEditingRecord(root); edit != nil {
		if !edit.node.Disabled {
			n.Kind = ui.Column
		}
		n.Children = append(n.Children, accessibilityFocusProjection(edit, prefix))
	}
	for _, child := range root.children {
		n.Children = append(n.Children, accessibilityFocusProjection(child, prefix))
	}
	if n.Kind == ui.Tabs {
		for i := range n.Options {
			n.Options[i].ID = prefix + accessibilityManagedID(n.Options[i].ID)
		}
		n.SelectedID = prefix + accessibilityManagedID(n.SelectedID)
	}
	return n
}

type accessibilityFocusTarget struct {
	widget  core.Widget
	focused bool
	scopeID string // Option targets validate visibility through their managed group.
	reveal  bool   // ListBox focus must reveal even an already-focused scrolled row.
}

func accessibilityNativeFocusTargets(root *record, prefix string, out map[string]accessibilityFocusTarget) {
	if root == nil {
		return
	}
	if ui.IsFocusable(root.node.Kind) {
		widget, err := focusWidget(root)
		if err == nil && !accessibilityTreeNil(widget) {
			focused := widget.AsWidget().ContainsFocus()
			if root.control != nil && root.control.tabs != nil {
				for _, button := range root.control.tabs.buttons {
					if !accessibilityTreeNil(button) && button.ContainsFocus() {
						focused = true
					}
				}
			}
			if root.control != nil && root.control.choice != nil {
				for _, button := range root.control.choice.buttons {
					focused = focused || button.ContainsFocus()
				}
				for _, option := range root.node.Options {
					if option.Disabled {
						continue
					}
					if button := root.control.choice.buttons[option.ID]; button != nil {
						out[prefix+accessibilityOptionID(root.node.ID, option.ID)] = accessibilityFocusTarget{widget: button, focused: button.ContainsFocus(), scopeID: prefix + accessibilityManagedID(root.node.ID), reveal: root.node.Kind == ui.ListBox}
					}
				}
			}
			out[prefix+accessibilityManagedID(root.node.ID)] = accessibilityFocusTarget{widget: widget, focused: focused, reveal: root.node.Kind == ui.ListBox}
			if root.control != nil && root.control.disclosure != nil {
				out[prefix+accessibilityDisclosureID(root.node.ID)] = accessibilityFocusTarget{widget: widget, focused: focused, scopeID: prefix + accessibilityManagedID(root.node.ID)}
			}
			if menuOpen(root) {
				for _, o := range menuDescriptor(root).Options {
					if !o.Disabled {
						if item := root.control.menu.popup.items[o.ID]; item != nil {
							out[prefix+menuActionID(root.node.ID, o.ID)] = accessibilityFocusTarget{widget: item, focused: item.ContainsFocus(), scopeID: prefix + accessibilityManagedID(root.node.ID)}
						}
					}
				}
			}
			if root.control != nil && root.control.split != nil {
				out[prefix+accessibilitySplitHandleID(root.node.ID)] = accessibilityFocusTarget{widget: widget, focused: focused, scopeID: prefix + accessibilityManagedID(root.node.ID)}
			}
		}
	}
	if w, ok := root.node.Custom.(*Workbench); root.node.Kind == ui.Extension && ok {
		accessibilityWorkbenchLeaves(w, func(g *workbench.Group) {
			focused := false
			for _, panel := range g.Tabs {
				if button := w.buttons[panel]; button != nil && button.ContainsFocus() {
					focused = true
				}
			}
			if button := w.buttons[g.Active]; button != nil {
				out[prefix+accessibilityDockID(root.node.ID, g.ID)] = accessibilityFocusTarget{widget: button, focused: focused}
			}
			accessibilityNativeFocusTargets(w.records[g.Active], prefix+accessibilityWorkbenchNamespace(root.node.ID, g.Active), out)
		})
	}
	if edit := gridEditingRecord(root); edit != nil {
		accessibilityNativeFocusTargets(edit, prefix, out)
	}
	for _, child := range root.children {
		accessibilityNativeFocusTargets(child, prefix, out)
	}
}

// The owner stamps modal hosts with their native namespace before render. A dock
// modal can then publish its own geometry without inspecting another Scene.
func accessibilityStampDialogNamespaces(root *record, prefix string) {
	if root == nil {
		return
	}
	if root.control != nil && root.control.dialog != nil {
		root.control.dialog.host.SetProperty("vellum.accessibility-namespace", prefix)
	}
	if w, ok := root.node.Custom.(*Workbench); root.node.Kind == ui.Extension && ok {
		accessibilityWorkbenchLeaves(w, func(g *workbench.Group) {
			accessibilityStampDialogNamespaces(w.records[g.Active], prefix+accessibilityWorkbenchNamespace(root.node.ID, g.Active))
		})
	}
	for _, child := range root.children {
		accessibilityStampDialogNamespaces(child, prefix)
	}
}

// activeWorkbenchRecords returns owner-cached visible panel records for control
// integration, including disabled panels that may need modal cleanup. It does
// not read the shared workspace or touch other windows.
func activeWorkbenchRecords(r *record) []*record {
	if r == nil || r.node.Kind != ui.Extension {
		return nil
	}
	w, ok := r.node.Custom.(*Workbench)
	if !ok || w == nil {
		return nil
	}
	records := []*record{}
	accessibilityWorkbenchLeaves(w, func(g *workbench.Group) {
		if active := w.records[g.Active]; active != nil {
			records = append(records, active)
		}
	})
	return records
}

// accessibilityDialogVisible follows native owner-cached visibility, including
// managed tabs and workbench panel frames, before opening or retaining a modal.
func accessibilityDialogVisible(r *record) bool {
	if r == nil || accessibilityTreeNil(r.widget) {
		return false
	}
	for parent := r.widget.AsTree().Parent; parent != nil; parent = parent.AsTree().Parent {
		if active, explicit := parent.AsTree().Property("vellum.workbench.active").(bool); explicit && !active {
			return false
		}
		if tabs, ok := parent.(*core.Tabs); ok {
			current, _ := tabs.CurrentTab()
			if accessibilityTreeNil(current) || r.widget.AsTree().ParentLevel(current) < 0 {
				return false
			}
		}
	}
	return true
}
