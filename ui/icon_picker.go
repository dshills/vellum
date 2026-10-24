package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"slices"
	"strings"
)

type IconPickerItem struct {
	ID, Label, Category string
	Icon                Symbol
	Image               *canvas.TiledImage
	Disabled            bool
}
type IconPickerEvent struct{ ID, ItemID string }
type iconPickerEpoch struct{ marker byte }

// IconPicker is a retained, same-window searchable asset chooser. Arrow keys
// highlight; Return/double-click or Choose proposes selection. Only accepted
// selection updates bounded recents and closes. Keep one instance on one owner.
type IconPicker struct {
	ID, Label, SelectedID                string
	Items                                []IconPickerItem
	Open, Disabled                       bool
	Change                               func(IconPickerEvent)
	Style, PopoverStyle                  Style
	query, category, highlight, identity string
	recent                               []string
	epoch                                *iconPickerEpoch
	collection                           *CollectionView
}

func IconPickerButtonID(id string) string     { return id + "/icon-picker/button" }
func IconPickerCollectionID(id string) string { return id + "/icon-picker/items" }
func (p *IconPicker) validate() error {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || len(p.Items) > MaxListBoxOptions {
		return fmt.Errorf("ui: icon picker requires ID, label and <=256 items")
	}
	seen, categories := map[string]bool{}, map[string]bool{}
	for _, item := range p.Items {
		if !validTokenText(item.ID) || !validTokenText(item.Label) || item.Category != "" && !validTokenText(item.Category) || seen[item.ID] || item.Image != nil && item.Icon != "" {
			return fmt.Errorf("ui: invalid icon picker item")
		}
		seen[item.ID] = true
		categories[item.Category] = true
		if item.Image != nil && item.Image.TileCount() == 0 {
			return fmt.Errorf("ui: icon picker image requires pixels")
		}
		if item.Icon != "" {
			if err := validateIcon(Node{Kind: Icon, Icon: item.Icon}); err != nil {
				return err
			}
		}
	}
	if len(categories) > 32 || p.SelectedID != "" && !seen[p.SelectedID] {
		return fmt.Errorf("ui: absent picker selection or excessive categories")
	}
	return nil
}
func pickerCategory(s string) string {
	return "category-" + base64.RawURLEncoding.EncodeToString([]byte(s))
}
func (p *IconPicker) visible(item IconPickerItem) bool {
	category := p.category == "" || p.category == pickerCategory(item.Category) || p.category == "recent" && slices.Contains(p.recent, item.ID)
	return category && strings.Contains(strings.ToLower(item.Label), strings.ToLower(p.query))
}
func (p *IconPicker) View() (Node, error) {
	if err := p.validate(); err != nil {
		return Node{}, err
	}
	if p.identity != p.ID {
		p.recent = nil
		p.query = ""
		p.category = ""
		p.highlight = ""
	}
	present := map[string]bool{}
	for _, item := range p.Items {
		present[item.ID] = true
	}
	p.recent = slices.DeleteFunc(p.recent, func(id string) bool { return !present[id] })
	if p.identity != p.ID || !p.Open || p.Disabled {
		p.epoch = nil
		p.identity = p.ID
	}
	if p.Open && !p.Disabled && p.epoch == nil {
		p.epoch = &iconPickerEpoch{}
		p.query = ""
		p.category = ""
		p.highlight = p.SelectedID
	}
	id, epoch := p.ID, p.epoch
	available := func() bool { return p.ID == id && !p.Disabled && p.Change != nil && p.validate() == nil }
	live := func() bool { return available() && p.Open && epoch != nil && p.epoch == epoch }
	shown := slices.Clone(p.Items)
	query, category := p.query, p.category
	eligible := func(target string) bool {
		if !live() || p.query != query || p.category != category {
			return false
		}
		for _, item := range p.Items {
			if item.ID == target && !item.Disabled && p.visible(item) {
				for _, old := range shown {
					if item == old {
						return true
					}
				}
			}
		}
		return false
	}
	close := func() { p.Open = false; p.epoch = nil }
	choose := func(target string) {
		if !eligible(target) {
			return
		}
		p.Change(IconPickerEvent{ID: id, ItemID: target})
		if p.SelectedID == target {
			p.recent = append([]string{target}, slices.DeleteFunc(slices.Clone(p.recent), func(s string) bool { return s == target })...)
			p.recent = p.recent[:min(8, len(p.recent))]
			close()
		}
	}
	button := Node{ID: IconPickerButtonID(id), Kind: Button, Text: p.Label, Disabled: !available(), Activate: func() {
		if available() {
			p.Open = true
			p.highlight = p.SelectedID
			p.query = ""
			p.category = ""
		}
	}}
	for _, item := range p.Items {
		if item.ID == p.SelectedID {
			button.Text = item.Label
			button.Icon = item.Icon
			button.IconImage = item.Image
		}
	}
	if p.collection == nil {
		p.collection = &CollectionView{ID: IconPickerCollectionID(id), Label: p.Label}
	}
	collection := p.collection
	collection.ID = IconPickerCollectionID(id)
	collection.Items = nil
	collection.SelectedIDs = nil
	collection.ItemWidth = 96
	collection.Style = Style{MinSize: Some(Size{Width: 340, Height: 160}), MaxSize: Some(Size{Height: 240}), GrowX: Some(float32(1))}
	for _, item := range p.Items {
		if p.visible(item) {
			collection.Items = append(collection.Items, CollectionItem{ID: item.ID, Label: item.Label, Icon: item.Icon, Thumbnail: item.Image, Disabled: item.Disabled})
			if item.ID == p.highlight {
				collection.SelectedIDs = []string{item.ID}
			}
		}
	}
	collection.Change = func(e ListSelectionEvent) {
		if eligible(e.TargetID) {
			p.highlight = e.TargetID
		}
	}
	collection.Activate = func(e CollectionEvent) { choose(e.ItemID) }
	grid, err := collection.View()
	if err != nil {
		return Node{}, err
	}
	search := &SearchBox{ID: id + "/icon-picker/search", Label: "Search icons", Text: p.query, Change: func(e SearchEvent) {
		if live() && len(e.Text) <= 256 {
			p.query = e.Text
			p.highlight = ""
		}
	}}
	searchNode, err := search.View()
	if err != nil {
		return Node{}, err
	}
	options := []Option{{ID: "all", Label: "All categories"}}
	selectedCategory := "all"
	if p.category != "" {
		selectedCategory = p.category
	}
	if len(p.recent) > 0 {
		options = append(options, Option{ID: "recent", Label: "Recently used"})
	}
	cats := map[string]bool{}
	for _, item := range p.Items {
		if item.Category != "" && !cats[item.Category] {
			cats[item.Category] = true
			options = append(options, Option{ID: pickerCategory(item.Category), Label: item.Category})
		}
	}
	foundCategory := false
	for _, o := range options {
		foundCategory = foundCategory || o.ID == selectedCategory
	}
	if !foundCategory {
		selectedCategory = "all"
		p.category = ""
	}
	children := []Node{searchNode, {ID: id + "/icon-picker/category", Kind: Selector, SelectedID: selectedCategory, Options: options, Accessibility: Accessibility{Name: "Icon category"}, Change: func(e ChangeEvent) {
		if !live() {
			return
		}
		for _, o := range options {
			if o.ID == e.SelectedID {
				p.category = e.SelectedID
				if p.category == "all" {
					p.category = ""
				}
				p.highlight = ""
				return
			}
		}
	}}, grid}
	if len(collection.Items) == 0 {
		children = append(children, Node{ID: id + "/icon-picker/empty", Kind: Label, Text: "No matching icons"})
	}
	children = append(children, Node{ID: id + "/icon-picker/actions", Kind: Row, Style: Style{Justify: Some(AlignEnd)}, Children: []Node{{ID: id + "/icon-picker/cancel", Kind: Button, Text: "Cancel", Activate: func() {
		if live() {
			close()
		}
	}}, {ID: id + "/icon-picker/choose", Kind: Button, Text: "Choose", Disabled: !eligible(p.highlight), Activate: func() { choose(p.highlight) }}}})
	panel, err := (Popover{ID: id + "/icon-picker/popup", AnchorID: button.ID, Label: p.Label, Open: p.Open, Disabled: !available(), Dismiss: func() {
		if live() {
			close()
		}
	}, Style: Style{MinSize: Some(Size{Width: 380}), MaxSize: Some(Size{Width: 480}), Padding: Some(UniformInsets(12)), Gap: Some(float32(8))}.Overlay(p.PopoverStyle), Children: children}).View()
	if err != nil {
		return Node{}, err
	}
	n := Node{ID: id, Kind: Column, Style: p.Style, Children: []Node{button, panel}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
