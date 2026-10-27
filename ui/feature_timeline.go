package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
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
	// A single compact strip keeps the same controlled events and stable
	// action IDs. Item names move into tooltips/AX rather than wide body cards.
	root := Node{ID: id, Kind: Column, Disabled: p.Disabled, Accessibility: Accessibility{Name: p.Label}, Style: Style{Padding: Some(Insets{Top: 4, Bottom: 4, Left: 8, Right: 8}), Gap: Some(float32(0)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}.Overlay(p.Style)}
	quiet := Style{Padding: Some(Insets{}), Border: Some(Border{}), Background: Some(color.NRGBA{}), Typography: Some(MacTypography(11, 400)), MinSize: Some(Size{Width: 24, Height: 24}), MaxSize: Some(Size{Width: 24, Height: 24}), Grow: Some(float32(0)), CornerRadius: Some(float32(3))}
	controls := Node{ID: id + "/controls", Kind: Row, Style: Style{Padding: Some(Insets{}), Gap: Some(float32(2)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}}
	for _, a := range []struct {
		key, label string
		icon       Symbol
		delta      int
	}{{"back", "Previous feature", SymbolLeft, -1}, {"forward", "Next feature", SymbolRight, 1}} {
		delta := a.delta
		controls.Children = append(controls.Children, Node{ID: id + "/" + a.key, Kind: Button, Icon: a.icon, Appearance: ButtonTool, Tooltip: a.label, Accessibility: Accessibility{Name: a.label}, Style: quiet, Disabled: p.Change == nil || p.Marker+delta < 0 || p.Marker+delta > len(p.Items), Activate: func() { event("", "marker", p.Marker+delta) }})
	}
	selected := -1
	for j, i := range p.Items {
		if i.ID == p.SelectedID {
			selected = j
		}
	}
	for _, a := range []struct {
		key, label string
		icon       Symbol
		delta      int
	}{{"earlier", "Move feature earlier", SymbolUndo, -1}, {"later", "Move feature later", SymbolRedo, 1}} {
		delta := a.delta
		controls.Children = append(controls.Children, Node{ID: id + "/" + a.key, Kind: Button, Icon: a.icon, Appearance: ButtonTool, Tooltip: a.label, Accessibility: Accessibility{Name: a.label}, Style: quiet, Disabled: p.Change == nil || selected < 0 || selected+delta < 0 || selected+delta >= len(p.Items), Activate: func() {
			index := -1
			for j, i := range p.Items {
				if i.ID == p.SelectedID {
					index = j
				}
			}
			event(p.SelectedID, "move", index+delta)
		}})
	}
	track := Node{ID: id + "/track", Kind: Row, Style: Style{Padding: Some(Insets{}), Gap: Some(float32(2)), GrowY: Some(float32(0))}}
	marker := func() Node {
		return Node{ID: id + "/marker", Kind: Column, Accessibility: Accessibility{Name: "History marker", Help: fmt.Sprintf("%d of %d features active", p.Marker, len(p.Items))}, Style: Style{Padding: Some(Insets{}), MinSize: Some(Size{Width: 2, Height: 24}), MaxSize: Some(Size{Width: 2, Height: 24}), Grow: Some(float32(0)), Background: Some(color.NRGBA{R: 58, G: 122, B: 225, A: 255})}}
	}
	for j, item := range p.Items {
		if j == p.Marker {
			track.Children = append(track.Children, marker())
		}
		itemID := item.ID
		name := item.Label
		help := "Select feature"
		if j >= p.Marker {
			name += " · paused"
			help = "After history marker"
		}
		style := quiet
		n := Node{ID: id + "/feature/" + ribbonPart(itemID), Kind: Button, IconImage: item.IconImage, Disabled: item.Disabled || p.Change == nil, Appearance: ButtonTool, Tooltip: name, Accessibility: Accessibility{Name: name, Help: help}, Style: style, Activate: func() { event(itemID, "select", 0) }}
		if item.IconImage == nil {
			n.Text = item.Label
			n.Style.MinSize = Some(Size{Height: 24})
			n.Style.MaxSize = Some(Size{Height: 24})
			n.Style.Padding = Some(Insets{Left: 6, Right: 6})
		}
		if item.ID == p.SelectedID {
			n.Style.Background = Some(color.NRGBA{R: 216, G: 231, B: 251, A: 255})
			n.Style.Foreground = Some(color.NRGBA{R: 26, G: 48, B: 76, A: 255})
			n.Style.Border = Some(Border{Width: 1, Color: color.NRGBA{R: 58, G: 122, B: 225, A: 255}})
		}
		// A paused item keeps its gray fill even when the blue selection outline is present.
		if j >= p.Marker {
			n.Style.Background = Some(color.NRGBA{R: 237, G: 240, B: 244, A: 255})
			n.Style.Foreground = Some(color.NRGBA{R: 75, G: 86, B: 102, A: 255})
		}
		track.Children = append(track.Children, n)
	}
	if p.Marker == len(p.Items) {
		track.Children = append(track.Children, marker())
	}
	controls.Children = append(controls.Children, Node{ID: id + "/scroll", Kind: Scroll, Children: []Node{track}, Style: Style{Padding: Some(Insets{}), GrowX: Some(float32(1)), GrowY: Some(float32(0)), MinSize: Some(Size{Height: 28}), MaxSize: Some(Size{Height: 28})}})
	root.Children = append(root.Children, controls)
	return root, Validate(root)
}
