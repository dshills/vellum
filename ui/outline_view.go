package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/virtual"
	"slices"
	"strings"
	"unicode/utf8"
)

type OutlineItem struct {
	ID, Parent, Label, Badge string
	Icon                     Symbol
	Image                    *canvas.TiledImage
	Disabled                 bool
}
type OutlineDisclosureEvent struct {
	ID, ItemID string
	Expanded   bool
}

func (e OutlineDisclosureEvent) Apply(current []string) []string {
	next := slices.Clone(current)
	i := slices.Index(next, e.ItemID)
	if e.Expanded && i < 0 {
		return append(next, e.ItemID)
	}
	if !e.Expanded && i >= 0 {
		return append(next[:i], next[i+1:]...)
	}
	return next
}

type OutlineEvent struct{ ID, ItemID string }
type OutlineViewRow struct {
	virtual.TreeRow
	Icon     Symbol
	Image    *canvas.TiledImage
	Badge    string
	Disabled bool
}
type OutlineViewLayout struct {
	source     *OutlineView
	Rows       []OutlineViewRow
	Disclosure func(OutlineDisclosureEvent)
	Activate   func(OutlineEvent)
}

func (l *OutlineViewLayout) SameSource(other *OutlineViewLayout) bool {
	if l == nil || other == nil {
		return l == other
	}
	return l.source != nil && l.source == other.source
}

// OutlineView composes the bounded virtual.Tree projection and source-list
// routing. Retain it on one owner. Hidden selected IDs remain caller-owned.
// Larger worker-built outlines continue to use DataTree/virtual.Outline.
type OutlineView struct {
	ID, Label                string
	Items                    []OutlineItem
	ExpandedIDs, SelectedIDs []string
	Disabled                 bool
	Change                   func(ListSelectionEvent)
	Disclosure               func(OutlineDisclosureEvent)
	Activate                 func(OutlineEvent)
	ContextMenu              *ContextMenu
	Style                    Style
}

func (p *OutlineView) projection() ([]OutlineViewRow, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || len(p.Items) > MaxListBoxOptions {
		return nil, fmt.Errorf("ui: outline needs ID/label and <=256 items")
	}
	nodes := make([]virtual.TreeNode, len(p.Items))
	items := map[string]OutlineItem{}
	parents := map[string]bool{}
	for i, item := range p.Items {
		if !validTokenText(item.ID) || !validTokenText(item.Label) || item.Parent != "" && !validTokenText(item.Parent) || items[item.ID].ID != "" || len(item.Badge) > 32 || !utf8.ValidString(item.Badge) || strings.ContainsAny(item.Badge, "\x00\r\n") {
			return nil, fmt.Errorf("ui: invalid outline item")
		}
		if err := validateIcon(Node{Kind: Button, Text: item.Label, Icon: item.Icon, IconImage: item.Image}); err != nil {
			return nil, err
		}
		items[item.ID] = item
		nodes[i] = virtual.TreeNode{ID: item.ID, Parent: item.Parent, Label: item.Label}
		parents[item.Parent] = true
	}
	tree, err := virtual.NewTree(nodes)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, id := range p.ExpandedIDs {
		if items[id].ID == "" || !parents[id] || seen[id] {
			return nil, fmt.Errorf("ui: invalid outline expansion")
		}
		seen[id] = true
		_ = tree.Toggle(id)
	}
	seen = map[string]bool{}
	for _, id := range p.SelectedIDs {
		if items[id].ID == "" || seen[id] {
			return nil, fmt.Errorf("ui: invalid outline selection")
		}
		seen[id] = true
	}
	for _, item := range p.Items {
		depth := 0
		for parent := item.Parent; parent != ""; parent = items[parent].Parent {
			depth++
			if depth > 16 {
				return nil, fmt.Errorf("ui: bounded outline depth exceeds 16")
			}
		}
	}
	visible, err := tree.View(0, MaxListBoxOptions)
	if err != nil {
		return nil, err
	}
	rows := make([]OutlineViewRow, 0, len(visible))
	for _, row := range visible {
		item := items[row.ID]
		disabled := item.Disabled
		for parent := item.Parent; parent != ""; parent = items[parent].Parent {
			disabled = disabled || items[parent].Disabled
		}
		rows = append(rows, OutlineViewRow{TreeRow: row, Icon: item.Icon, Image: item.Image, Badge: item.Badge, Disabled: disabled})
	}
	return rows, nil
}
func (p *OutlineView) View() (Node, error) {
	rows, err := p.projection()
	if err != nil {
		return Node{}, err
	}
	id := p.ID
	current := func() map[string]OutlineViewRow {
		if p.ID != id || p.Disabled {
			return nil
		}
		next, err := p.projection()
		if err != nil {
			return nil
		}
		m := make(map[string]OutlineViewRow, len(next))
		for _, row := range next {
			if !row.Disabled {
				m[row.ID] = row
			}
		}
		return m
	}
	shown := map[string]OutlineViewRow{}
	for _, row := range rows {
		shown[row.ID] = row
	}
	eligible := func(m map[string]OutlineViewRow, target string) bool {
		row, ok := m[target]
		old, wasShown := shown[target]
		return ok && wasShown && row.Parent == old.Parent
	}
	layout := &OutlineViewLayout{source: p, Rows: rows}
	n := Node{ID: id, Kind: ListBox, Multiple: true, Disabled: p.Disabled, Outline: layout, ContextMenu: p.ContextMenu, Style: Style{GrowX: Some(float32(1)), MinSize: Some(Size{Height: 200}), MaxSize: Some(Size{Height: 400})}.Overlay(p.Style), Accessibility: Accessibility{Name: p.Label, Help: "Outline. Right expands or enters; Left collapses or moves to parent. Command/Shift selects; Return opens."}}
	for _, row := range rows {
		n.Options = append(n.Options, Option{ID: row.ID, Label: row.Label, Disabled: row.Disabled, Badge: row.Badge})
		if slices.Contains(p.SelectedIDs, row.ID) {
			n.SelectedIDs = append(n.SelectedIDs, row.ID)
		}
	}
	if p.Change != nil {
		n.SelectionChange = func(e ListSelectionEvent) {
			m := current()
			if p.Change == nil || !eligible(m, e.TargetID) {
				return
			}
			for _, target := range e.IDs {
				if !eligible(m, target) {
					return
				}
			}
			e.ID = id
			e.IDs = slices.Clone(e.IDs)
			p.Change(e)
		}
	}
	if p.Disclosure != nil {
		layout.Disclosure = func(e OutlineDisclosureEvent) {
			m := current()
			if !eligible(m, e.ItemID) || p.Disclosure == nil {
				return
			}
			row := m[e.ItemID]
			if !row.HasChildren || row.Expanded == e.Expanded {
				return
			}
			p.Disclosure(OutlineDisclosureEvent{ID: id, ItemID: e.ItemID, Expanded: e.Expanded})
		}
	}
	if p.Activate != nil {
		layout.Activate = func(e OutlineEvent) {
			if p.Activate != nil && eligible(current(), e.ItemID) {
				p.Activate(OutlineEvent{ID: id, ItemID: e.ItemID})
			}
		}
	}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
func ValidateOutlineViewNode(n Node) error {
	l := n.Outline
	if l == nil {
		return nil
	}
	if l.source == nil || n.Kind != ListBox || !n.Multiple || n.Collection != nil || n.DocumentList != nil || len(l.Rows) != len(n.Options) {
		return fmt.Errorf("ui: invalid outline presentation")
	}
	for i, row := range l.Rows {
		if err := validateIcon(Node{Kind: Button, Text: row.Label, Icon: row.Icon, IconImage: row.Image}); err != nil {
			return err
		}
		if row.Label != n.Options[i].Label || row.Badge != n.Options[i].Badge || row.ID != n.Options[i].ID || row.Disabled != n.Options[i].Disabled || row.Depth < 0 || row.Depth > 16 {
			return fmt.Errorf("ui: outline rows/options disagree")
		}
	}
	return nil
}
