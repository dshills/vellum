package ui

// FocusScope selects the last open enabled dialog in visible descriptor order.
// A modal owns navigation until it closes; inactive tab pages are not searched.
func FocusScope(root Node) Node {
	var active *Node
	var visit func(Node, bool)
	visit = func(n Node, disabled bool) {
		disabled = disabled || n.Disabled
		if disabled || (n.Kind == Dialog && !n.Open) {
			return
		}
		if n.Kind == Dialog {
			copy := n
			active = &copy
		}
		if n.Kind == Disclosure && !n.Checked {
			return
		}
		for _, c := range n.Children {
			if n.Kind == Tabs && c.ID != n.SelectedID {
				continue
			}
			visit(c, disabled)
		}
	}
	visit(root, false)
	if active != nil {
		return *active
	}
	return root
}

// FocusOrder returns visible, enabled managed targets within the active modal.
// Tabs contribute one header target followed by their selected page. Extensions
// implement their own internal keyboard order; they are not descriptor targets.
func FocusOrder(root Node) []string {
	root = FocusScope(root)
	out := []string{}
	var visit func(Node, bool)
	visit = func(n Node, disabled bool) {
		disabled = disabled || n.Disabled
		if disabled || (n.Kind == Dialog && !n.Open) {
			return
		}
		if n.Kind == SplitPane && len(n.Children) == 2 {
			visit(n.Children[0], disabled)
			out = append(out, n.ID) // Divider is between the two pane tab orders.
			visit(n.Children[1], disabled)
			return
		}
		if _, ok := FocusTarget(n, n.ID); ok {
			out = append(out, n.ID)
		}
		if n.Kind == Disclosure && !n.Checked {
			return
		}
		for _, c := range n.Children {
			if n.Kind == Tabs && c.ID != n.SelectedID {
				continue
			}
			visit(c, disabled)
		}
	}
	visit(root, false)
	return out
}

// NextFocus wraps forward or backward through the active scope. An absent
// current target selects the first (forward) or last (backward) target.
func NextFocus(root Node, current string, reverse bool) (string, bool) {
	order := FocusOrder(root)
	if len(order) == 0 {
		return "", false
	}
	index := -1
	for i, id := range order {
		if id == current {
			index = i
			break
		}
	}
	if reverse {
		if index <= 0 {
			index = len(order)
		}
		return order[index-1], true
	}
	return order[(index+1)%len(order)], true
}
