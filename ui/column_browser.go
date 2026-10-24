package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"reflect"
	"slices"
	"strings"
)

const MaxColumnBrowserDepth = 4

type BrowserItem struct {
	ID, Label string
	Icon      Symbol
	Thumbnail *canvas.TiledImage
	Disabled  bool
	Children  []BrowserItem
}
type BrowserEvent struct {
	ID   string
	Path []string
}
type BrowserResizeEvent struct {
	ID, ColumnID string
	Fraction     float32
}

// ColumnBrowser is an owner-held, controlled hierarchy, not a filesystem view.
// Up/Down and typeahead select rows; Tab traverses columns and resize dividers.
// Scroll horizontally when the columns exceed the viewport. At most 256 items
// and four levels are accepted. Preview is application-provided managed content.
type ColumnBrowser struct {
	ID, Label string
	Items     []BrowserItem
	Path      []string
	Disabled  bool
	Change    func(BrowserEvent)
	Fractions map[string]float32
	Resize    func(BrowserResizeEvent)
	Preview   *Node
	Style     Style
}

func BrowserColumnID(root string, parent []string) string {
	out := root + "/browser/column"
	for _, id := range parent {
		out += "/" + base64.RawURLEncoding.EncodeToString([]byte(id))
	}
	return out
}
func browserChildren(items []BrowserItem, path []string) ([]BrowserItem, bool) {
	for _, id := range path {
		found := false
		for _, item := range items {
			if item.ID == id && !item.Disabled {
				items = item.Children
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return items, true
}
func (b *ColumnBrowser) validate() error {
	if b == nil || !validTokenText(b.ID) || len(b.Path) > MaxColumnBrowserDepth {
		return fmt.Errorf("ui: invalid column browser identity/path")
	}
	total := 0
	seen := map[string]bool{}
	var walk func([]BrowserItem, int) error
	walk = func(items []BrowserItem, depth int) error {
		if depth > MaxColumnBrowserDepth && len(items) > 0 {
			return fmt.Errorf("ui: browser hierarchy too deep")
		}
		for _, item := range items {
			total++
			if total > MaxListBoxOptions || !validTokenText(item.ID) || !validTokenText(item.Label) || len(item.Label) > 512 || seen[item.ID] {
				return fmt.Errorf("ui: invalid/duplicate/excessive browser items")
			}
			seen[item.ID] = true
			if item.Thumbnail != nil && item.Thumbnail.TileCount() == 0 || item.Thumbnail != nil && item.Icon != "" {
				return fmt.Errorf("ui: browser thumbnail requires pixels without an icon")
			}
			if item.Icon != "" {
				if err := validateIcon(Node{Kind: Icon, Icon: item.Icon}); err != nil {
					return err
				}
			}
			if err := walk(item.Children, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(b.Items, 1); err != nil {
		return err
	}
	if _, ok := browserChildren(b.Items, b.Path); !ok {
		return fmt.Errorf("ui: browser path missing or disabled")
	}
	if len(b.Fractions) > MaxColumnBrowserDepth+1 {
		return fmt.Errorf("ui: excessive browser fractions")
	}
	for _, fraction := range b.Fractions {
		if SplitRange(Numeric{}).Validate(fraction) != nil {
			return fmt.Errorf("ui: invalid browser fraction")
		}
	}
	return nil
}
func cloneBrowserItems(items []BrowserItem) []BrowserItem {
	out := slices.Clone(items)
	for i := range out {
		out[i].Children = cloneBrowserItems(out[i].Children)
	}
	return out
}
func (b *ColumnBrowser) View() (Node, error) {
	if err := b.validate(); err != nil {
		return Node{}, err
	}
	id := b.ID
	shown := cloneBrowserItems(b.Items)
	path := slices.Clone(b.Path)
	available := func() bool { return b.ID == id && !b.Disabled && b.Change != nil && b.validate() == nil }
	var columns []Node
	items := b.Items
	for depth := 0; depth <= len(path) && len(items) > 0; depth++ {
		parent := slices.Clone(path[:depth])
		columnID := BrowserColumnID(id, parent)
		options := []Option{}
		selected := []string{}
		for _, item := range items {
			label := item.Label
			if len(item.Children) > 0 {
				label += "  ›"
			}
			options = append(options, Option{ID: item.ID, Label: label, Icon: item.Icon, IconImage: item.Thumbnail, Disabled: item.Disabled})
			if depth < len(path) && path[depth] == item.ID {
				selected = append(selected, item.ID)
			}
		}
		column := Node{ID: columnID, Kind: ListBox, Multiple: true, SelectedIDs: selected, Options: options, Disabled: b.Change == nil, Accessibility: Accessibility{Name: fmt.Sprintf("%s column %d", b.Label, depth+1), Help: "Up/Down selects; Tab moves between columns and dividers."}, Style: Style{MinSize: Some(Size{Width: 160, Height: 220}), MaxSize: Some(Size{Height: 280}), GrowX: Some(float32(1)), GrowY: Some(float32(1)), Padding: Some(UniformInsets(4))}}
		column.SelectionChange = func(e ListSelectionEvent) {
			if !available() || !slices.Equal(parent, pathPrefix(b.Path, len(parent))) {
				return
			}
			old, oldOK := browserChildren(shown, parent)
			current, ok := browserChildren(b.Items, parent)
			if !oldOK || !ok {
				return
			}
			for _, item := range current {
				if item.ID == e.TargetID && !item.Disabled {
					for _, previous := range old {
						if previous.ID == item.ID && reflect.DeepEqual(previous, item) {
							next := append(slices.Clone(parent), item.ID)
							b.Change(BrowserEvent{ID: id, Path: next})
							return
						}
					}
				}
			}
		}
		columns = append(columns, column)
		if depth == len(path) {
			break
		}
		items, _ = browserChildren(b.Items, path[:depth+1])
	}
	if len(columns) == 0 {
		columns = append(columns, Node{ID: BrowserColumnID(id, nil), Kind: Label, Text: "No items", MinWidth: 160})
	}
	if b.Preview != nil {
		columns = append(columns, Node{ID: id + "/browser/preview", Kind: Column, MinWidth: 180, Children: []Node{*b.Preview}})
	}
	content := columns[len(columns)-1]
	for i := len(columns) - 2; i >= 0; i-- {
		columnID := columns[i].ID
		fraction := float32(1) / float32(len(columns)-i)
		if f, ok := b.Fractions[columnID]; ok {
			fraction = f
		}
		content = Node{ID: columnID + "/divider", Kind: SplitPane, Number: fraction, MinWidth: float32(len(columns)-i) * 160, MinHeight: 220, Children: []Node{columns[i], content}, Change: func(e ChangeEvent) {
			if b.ID != id || b.Disabled || b.Resize == nil || !slices.Equal(b.Path, path) || b.validate() != nil || SplitRange(Numeric{}).Validate(e.Number) != nil {
				return
			}
			b.Resize(BrowserResizeEvent{ID: id, ColumnID: columnID, Fraction: e.Number})
		}}
	}
	label := strings.TrimSpace(b.Label)
	if label == "" {
		label = "Browser"
	}
	root := Node{ID: id, Kind: Scroll, Disabled: b.Disabled, Accessibility: Accessibility{Name: label}, Style: Style{MinSize: Some(Size{Height: 240}), MaxSize: Some(Size{Height: 320}), GrowX: Some(float32(1)), Padding: Some(Insets{})}.Overlay(b.Style), Children: []Node{content}}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
func pathPrefix(path []string, n int) []string {
	if len(path) < n {
		return nil
	}
	return path[:n]
}
