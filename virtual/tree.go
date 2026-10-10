package virtual

import "fmt"

type TreeNode struct{ ID, Parent, Label string }
type TreeRow struct {
	TreeNode
	Depth       int
	Expanded    bool
	HasChildren bool
}
type treeIndex struct {
	id    string
	depth int
}
type Tree struct {
	visible  []treeIndex
	dirty    bool
	nodes    map[string]TreeNode
	children map[string][]string
	expanded map[string]bool
	Selected string
}

func NewTree(nodes []TreeNode) (*Tree, error) {
	t := &Tree{dirty: true, nodes: map[string]TreeNode{}, children: map[string][]string{}, expanded: map[string]bool{}}
	for _, n := range nodes {
		if n.ID == "" || t.nodes[n.ID].ID != "" {
			return nil, fmt.Errorf("virtual: duplicate or empty tree ID")
		}
		t.nodes[n.ID] = n
		t.children[n.Parent] = append(t.children[n.Parent], n.ID)
	}
	// Parent chains detect cycles without recursion or quadratic work.
	colors := map[string]uint8{}
	for id := range t.nodes {
		var chain []string
		cur := id
		for cur != "" && colors[cur] == 0 {
			n, ok := t.nodes[cur]
			if !ok {
				return nil, fmt.Errorf("virtual: missing parent %q", cur)
			}
			colors[cur] = 1
			chain = append(chain, cur)
			cur = n.Parent
		}
		if cur != "" && colors[cur] == 1 {
			return nil, fmt.Errorf("virtual: parent cycle")
		}
		for _, p := range chain {
			colors[p] = 2
		}
	}
	return t, nil
}
func (t *Tree) Toggle(id string) error {
	if t.nodes[id].ID == "" {
		return fmt.Errorf("virtual: unknown node")
	}
	t.expanded[id] = !t.expanded[id]
	t.dirty = true
	return nil
}
func (t *Tree) Select(id string) error {
	if t.nodes[id].ID == "" {
		return fmt.Errorf("virtual: unknown node")
	}
	t.Selected = id
	return nil
}

// View materializes only the viewport. Expansion rebuilds an O(N) lightweight
// index; scrolling then performs O(count) work regardless of the offset.
func (t *Tree) View(start, count int) ([]TreeRow, error) {
	if start < 0 || count < 1 || count > MaxVisible {
		return nil, fmt.Errorf("virtual: invalid tree viewport")
	}
	t.VisibleLen()
	if start >= len(t.visible) {
		return nil, nil
	}
	end := start + count
	if end > len(t.visible) {
		end = len(t.visible)
	}
	out := make([]TreeRow, 0, end-start)
	for _, entry := range t.visible[start:end] {
		out = append(out, TreeRow{TreeNode: t.nodes[entry.id], Depth: entry.depth, Expanded: t.expanded[entry.id], HasChildren: len(t.children[entry.id]) > 0})
	}
	return out, nil
}
func (t *Tree) Len() int { return len(t.nodes) }
func (t *Tree) VisibleLen() int {
	if !t.dirty {
		return len(t.visible)
	}
	t.visible = t.visible[:0]
	type frame struct {
		ids          []string
		index, depth int
	}
	stack := []frame{{ids: t.children[""]}}
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		if top.index == len(top.ids) {
			stack = stack[:len(stack)-1]
			continue
		}
		id := top.ids[top.index]
		depth := top.depth
		top.index++
		t.visible = append(t.visible, treeIndex{id, depth})
		if t.expanded[id] && len(t.children[id]) > 0 {
			stack = append(stack, frame{ids: t.children[id], depth: depth + 1})
		}
	}
	t.dirty = false
	return len(t.visible)
}
