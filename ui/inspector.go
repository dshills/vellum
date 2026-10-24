package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"strings"
)

type InspectorProperty struct {
	ID, Label string
	// Editor keeps its own stable ID and application-owned callbacks/value.
	Editor Node
	// Mixed adds an explicit hint without substituting a false numeric value.
	Mixed bool
}
type InspectorSection struct {
	ID, Title  string
	Expanded   bool
	Properties []InspectorProperty
}
type InspectorDisclosureEvent struct {
	ID, SectionID string
	Expanded      bool
}

// Inspector composes retained disclosures and aligned property rows. It owns
// ID+"/inspector/"; editor IDs must not use this namespace. Retain on one owner.
type Inspector struct {
	ID, Title  string
	IconImage  *canvas.TiledImage
	Sections   []InspectorSection
	LabelWidth float32
	Disabled   bool
	Disclosure func(InspectorDisclosureEvent)
	Style      Style
}

func InspectorSectionID(id, section string) string { return id + "/inspector/section/" + section }
func (p *Inspector) View() (Node, error) {
	if p == nil || p.ID == "" || p.Title == "" || len(p.Sections) > 32 {
		return Node{}, fmt.Errorf("ui: inspector requires ID, title and at most 32 sections")
	}
	width := p.LabelWidth
	if width == 0 {
		width = 112
	}
	if !finite(width) || width < 48 || width > 320 {
		return Node{}, fmt.Errorf("ui: invalid inspector label width")
	}
	id := p.ID
	prefix := id + "/inspector/"
	root := Node{ID: id, Kind: Column, Disabled: p.Disabled, Style: Style{GrowX: Some(float32(1)), Gap: Some(float32(10)), Padding: Some(UniformInsets(12))}.Overlay(p.Style)}
	title := Node{ID: prefix + "title", Kind: Label, Text: p.Title, Style: Style{Typography: Some(Typography{Size: 14, Weight: 600})}}
	heading := Node{ID: prefix + "heading", Kind: Row, Children: []Node{title}}
	if p.IconImage != nil {
		heading.Children = append([]Node{{ID: prefix + "image", Kind: Image, Image: p.IconImage, Text: p.Title, Fit: ImageContain, Style: Style{MinSize: Some(Size{Width: 24, Height: 24}), MaxSize: Some(Size{Width: 24, Height: 24})}}}, heading.Children...)
	}
	root.Children = append(root.Children, heading)
	seen := map[string]bool{}
	total := 0
	for _, section := range p.Sections {
		if section.ID == "" || section.Title == "" || seen[section.ID] {
			return Node{}, fmt.Errorf("ui: invalid inspector section")
		}
		seen[section.ID] = true
		sectionID := section.ID
		group := Node{ID: InspectorSectionID(id, section.ID), Kind: Disclosure, Text: section.Title, Checked: section.Expanded, Change: func(e ChangeEvent) {
			if p.ID != id || p.Disabled || p.Disclosure == nil {
				return
			}
			for _, current := range p.Sections {
				if current.ID == sectionID {
					p.Disclosure(InspectorDisclosureEvent{ID: id, SectionID: sectionID, Expanded: e.Checked})
					return
				}
			}
		}}
		props := map[string]bool{}
		for _, property := range section.Properties {
			total++
			if total > 128 || property.ID == "" || property.Label == "" || props[property.ID] {
				return Node{}, fmt.Errorf("ui: invalid/excessive inspector properties")
			}
			props[property.ID] = true
			editor := property.Editor
			seenEditors := map[string]bool{}
			var namespaceOK func(Node, int) bool
			namespaceOK = func(n Node, depth int) bool {
				if depth >= 128 || n.ID == "" || seenEditors[n.ID] || strings.HasPrefix(n.ID, prefix) {
					return false
				}
				seenEditors[n.ID] = true
				for _, c := range n.Children {
					if !namespaceOK(c, depth+1) {
						return false
					}
				}
				return true
			}
			if !namespaceOK(editor, 0) {
				return Node{}, fmt.Errorf("ui: inspector editors require IDs outside %s", prefix)
			}
			if editor.Accessibility.Name == "" {
				editor.Accessibility.Name = property.Label
			}
			if property.Mixed {
				editor.Accessibility.Help = strings.TrimSpace(editor.Accessibility.Help + " Multiple values")
			}

			rowID := group.ID + "/row/" + property.ID
			label := Node{ID: rowID + "/label", Kind: Label, Text: property.Label, Style: Style{MinSize: Some(Size{Width: width}), MaxSize: Some(Size{Width: width}), TextAlign: Some(AlignEnd)}}
			children := []Node{label, editor}
			if property.Mixed {
				children = append(children, Node{ID: rowID + "/mixed", Kind: Label, Text: "Multiple values", Accessibility: Accessibility{Name: property.Label + ": multiple values"}, Style: Style{MinSize: Some(Size{Width: 104}), MaxSize: Some(Size{Width: 104}), Typography: Some(Typography{Size: 11, Italic: true})}})
			}
			group.Children = append(group.Children, Node{ID: rowID, Kind: Row, Style: Style{GrowX: Some(float32(1)), Gap: Some(float32(8)), Align: Some(AlignCenter)}, Children: children})
		}
		root.Children = append(root.Children, group)
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
