package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"slices"
)

type SettingsPage struct {
	ID, Label, Description string
	Icon                   Symbol
	IconImage              *canvas.TiledImage
	Disabled               bool
	Children               []Node
}
type SettingsEvent struct{ ID, PageID string }

// SettingsPane is a retained Mac-style sidebar/detail composition. Every page
// remains mounted in a headerless Tabs stack; inactive editors keep their values
// and native instances while leaving input/accessibility scope. Navigate proposes
// a stable page ID. Applications may embed the pane or put it in their own window.
type SettingsPane struct {
	ID, Title, SelectedID string
	Pages                 []SettingsPage
	Disabled              bool
	Navigate              func(SettingsEvent)
	NavigationWidth       float32
	Style                 Style
}

func SettingsNavigationID(id string) string { return id + "/settings/navigation" }
func SettingsPageID(id, page string) string {
	return id + "/settings/page/" + base64.RawURLEncoding.EncodeToString([]byte(page))
}
func (p *SettingsPane) View() (Node, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Title) || len(p.Pages) < 1 || len(p.Pages) > 16 {
		return Node{}, fmt.Errorf("ui: settings pane needs ID/title and 1..16 pages")
	}
	width := p.NavigationWidth
	if width == 0 {
		width = 160
	}
	if !finite(width) || width < 120 || width > 280 {
		return Node{}, fmt.Errorf("ui: invalid settings sidebar width")
	}
	id := p.ID
	seen := map[string]bool{}
	found := false
	options := []Option{}
	tabOptions := []Option{}
	pages := []Node{}
	for _, page := range p.Pages {
		if !validTokenText(page.ID) || !validTokenText(page.Label) || len(page.Description) > 2048 || seen[page.ID] {
			return Node{}, fmt.Errorf("ui: invalid settings page")
		}
		seen[page.ID] = true
		found = found || page.ID == p.SelectedID && !page.Disabled
		if err := validateIcon(Node{Kind: Button, Text: page.Label, Icon: page.Icon, IconImage: page.IconImage}); err != nil {
			return Node{}, err
		}
		options = append(options, Option{ID: page.ID, Label: page.Label, Icon: page.Icon, IconImage: page.IconImage, Disabled: page.Disabled})
		pageID := SettingsPageID(id, page.ID)
		tabOptions = append(tabOptions, Option{ID: pageID, Label: page.Label, Disabled: page.Disabled})
		header := Node{ID: pageID + "/heading", Kind: Row, Children: []Node{{ID: pageID + "/title", Kind: Label, Text: page.Label, Style: Style{Typography: Some(Typography{Size: 17, Weight: 600})}}}}
		if page.IconImage != nil {
			header.Children = append([]Node{{ID: pageID + "/image", Kind: Image, Image: page.IconImage, Text: page.Label, Fit: ImageContain, Style: Style{MinSize: Some(Size{Width: 28, Height: 28}), MaxSize: Some(Size{Width: 28, Height: 28})}}}, header.Children...)
		}
		children := []Node{header}
		if page.Description != "" {
			children = append(children, Node{ID: pageID + "/description", Kind: Label, Text: page.Description})
		}
		children = append(children, page.Children...)
		pages = append(pages, Node{ID: pageID, Kind: Scroll, Disabled: page.Disabled, Style: Style{Padding: Some(UniformInsets(12)), Gap: Some(float32(12)), GrowX: Some(float32(1)), GrowY: Some(float32(1))}, Children: children})
	}
	if !found {
		return Node{}, fmt.Errorf("ui: settings selected page must be enabled and present")
	}
	shown := slices.Clone(options)
	nav := Node{ID: SettingsNavigationID(id), Kind: ListBox, Multiple: true, SelectedIDs: []string{p.SelectedID}, Options: options, Disabled: p.Navigate == nil, Accessibility: Accessibility{Name: p.Title + " categories"}, Style: Style{MinSize: Some(Size{Width: width, Height: 260}), MaxSize: Some(Size{Width: width}), GrowX: Some(float32(0)), GrowY: Some(float32(1))}, SelectionChange: func(e ListSelectionEvent) {
		if p.ID != id || p.Disabled || p.Navigate == nil {
			return
		}
		for _, current := range p.Pages {
			if current.ID == e.TargetID && !current.Disabled {
				for _, old := range shown {
					if old.ID == current.ID && old.Label == current.Label && old.IconImage == current.IconImage && old.Icon == current.Icon && !old.Disabled {
						p.Navigate(SettingsEvent{ID: id, PageID: current.ID})
						return
					}
				}
			}
		}
	}}
	stack := Node{ID: id + "/settings/pages", Kind: Tabs, TabsHideHeader: true, SelectedID: SettingsPageID(id, p.SelectedID), Options: tabOptions, Children: pages, Grow: true}
	n := Node{ID: id, Kind: Column, Disabled: p.Disabled, Style: Style{GrowX: Some(float32(1)), MinSize: Some(Size{Height: 320})}.Overlay(p.Style), Children: []Node{{ID: id + "/settings/title", Kind: Label, Text: p.Title, Style: Style{Typography: Some(Typography{Size: 20, Weight: 600})}}, {ID: id + "/settings/body", Kind: Row, Grow: true, Style: Style{Align: Some(AlignStart)}, Children: []Node{nav, stack}}}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
