package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
)

type FeatureItem struct {
	ID, Label string
	IconImage *canvas.TiledImage
	Disabled  bool
}
type FeatureTimelineEvent struct {
	ID, ItemID, Action string
	Index              int
}

// FeatureTimeline presents an application-owned feature history, independently
// from Undo/Redo. Marker is a count of active features (0..len Items). Selecting,
// moving or changing the marker only proposes an event; no solver is implied.
// Reorder is via earlier/later buttons and accessible actions, not a drag gesture.
type FeatureTimeline struct {
	ID, Label, SelectedID string
	Items                 []FeatureItem
	Marker                int
	Disabled              bool
	Change                func(FeatureTimelineEvent)
	Style                 Style
}

func (p *FeatureTimeline) valid() bool {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || len(p.Items) > 128 || p.Marker < 0 || p.Marker > len(p.Items) {
		return false
	}
	seen := map[string]bool{}
	for _, i := range p.Items {
		if !validTokenText(i.ID) || !validTokenText(i.Label) || seen[i.ID] {
			return false
		}
		seen[i.ID] = true
	}
	return p.SelectedID == "" || seen[p.SelectedID]
}
func (p *FeatureTimeline) propose(item, action string, index int) {
	if !p.valid() || p.Disabled || p.Change == nil {
		return
	}
	if action == "marker" {
		if index < 0 || index > len(p.Items) {
			return
		}
	} else {
		found := -1
		for j, i := range p.Items {
			if i.ID == item && !i.Disabled {
				found = j
			}
		}
		if found < 0 {
			return
		}
		if action == "move" && (index < 0 || index >= len(p.Items) || index == found) {
			return
		}
	}
	p.Change(FeatureTimelineEvent{p.ID, item, action, index})
}
func (p *FeatureTimeline) View() (Node, error) {
	if !p.valid() {
		return Node{}, fmt.Errorf("ui: invalid feature timeline")
	}
	id := p.ID
	event := func(item, action string, index int) {
		if p.ID == id {
			p.propose(item, action, index)
		}
	}
	root := Node{ID: id, Kind: Column, Disabled: p.Disabled, Style: Style{Gap: Some(float32(6)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}.Overlay(p.Style)}
	controls := Node{ID: id + "/controls", Kind: Row, Style: Style{Gap: Some(float32(8)), GrowY: Some(float32(0))}, Children: []Node{{ID: id + "/label", Kind: Label, Text: p.Label, Style: Style{Typography: Some(MacTypography(11, 600))}}}}
	for _, a := range []struct {
		key, label string
		delta      int
	}{{"back", "Previous feature", -1}, {"forward", "Next feature", 1}} {
		delta := a.delta
		icon := SymbolLeft
		if delta > 0 {
			icon = SymbolRight
		}
		controls.Children = append(controls.Children, Node{ID: id + "/" + a.key, Kind: Button, Icon: icon, Accessibility: Accessibility{Name: a.label}, Disabled: p.Change == nil || p.Marker+delta < 0 || p.Marker+delta > len(p.Items), Activate: func() { event("", "marker", p.Marker+delta) }})
	}
	selected := -1
	for j, i := range p.Items {
		if i.ID == p.SelectedID {
			selected = j
		}
	}
	for _, a := range []struct {
		key, label string
		delta      int
	}{{"earlier", "Move feature earlier", -1}, {"later", "Move feature later", 1}} {
		delta := a.delta
		controls.Children = append(controls.Children, Node{ID: id + "/" + a.key, Kind: Button, Text: a.label, Disabled: p.Change == nil || selected < 0 || selected+delta < 0 || selected+delta >= len(p.Items), Activate: func() {
			index := -1
			for j, i := range p.Items {
				if i.ID == p.SelectedID {
					index = j
				}
			}
			event(p.SelectedID, "move", index+delta)
		}})
	}
	root.Children = append(root.Children, controls)
	track := Node{ID: id + "/track", Kind: Row, Style: Style{Gap: Some(float32(6)), GrowY: Some(float32(0))}}
	for j, item := range p.Items {
		if j == p.Marker {
			track.Children = append(track.Children, Node{ID: id + "/marker", Kind: Label, Text: "│", Accessibility: Accessibility{Name: "History marker"}})
		}
		itemID := item.ID
		n := Node{ID: id + "/feature/" + ribbonPart(itemID), Kind: Button, Text: item.Label, IconImage: item.IconImage, Disabled: item.Disabled || p.Change == nil, Appearance: ButtonQuiet, Tooltip: item.Label, Style: Style{Padding: Some(Insets{Left: 10, Right: 10, Top: 6, Bottom: 6}), Typography: Some(MacTypography(10, 400)), MinSize: Some(Size{Width: 68, Height: 54}), GrowX: Some(float32(0))}, Activate: func() { event(itemID, "select", 0) }}
		if item.IconImage != nil {
			n.IconPlacement = IconAboveText
		}
		if item.ID == p.SelectedID {
			n.Appearance = ButtonPrimary
		}
		if j >= p.Marker {
			n.Text += " · paused"
		}
		track.Children = append(track.Children, n)
	}
	if p.Marker == len(p.Items) {
		track.Children = append(track.Children, Node{ID: id + "/marker", Kind: Label, Text: "│", Accessibility: Accessibility{Name: "History marker"}})
	}
	root.Children = append(root.Children, Node{ID: id + "/scroll", Kind: Scroll, Children: []Node{track}, Style: Style{GrowX: Some(float32(1)), GrowY: Some(float32(0)), MinSize: Some(Size{Height: 70}), MaxSize: Some(Size{Height: 70})}})
	return root, Validate(root)
}
