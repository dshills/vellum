package ui

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/canvas"
)

// CollectionItem has stable application identity. Thumbnail is immutable shared
// image data; Icon is used when Thumbnail is nil. Neither loads external files.
type CollectionItem struct {
	ID, Label string
	Icon      Symbol
	Thumbnail *canvas.TiledImage
	Disabled  bool
}

type CollectionEvent struct{ ID, ItemID string }

// CollectionLayout is a detached native presentation snapshot produced by View.
// Applications normally configure CollectionView, rather than constructing it.
// Items and Options must have matching order and eligibility.
type CollectionLayout struct {
	source    *CollectionView
	Items     []CollectionItem
	ItemWidth float32
	Activate  func(CollectionEvent)
}

// SameSource compares immutable component identity, without reading owner state.
// Detached snapshots from the same retained CollectionView are compatible.
func (c *CollectionLayout) SameSource(other *CollectionLayout) bool {
	if c == nil || other == nil {
		return c == other
	}
	if c.source == nil || other.source == nil {
		return c == other
	}
	return c.source == other.source
}

// CollectionView is a bounded, controlled Finder-style icon collection. Retain
// it on one application owner. Change proposes selection edits; the owner accepts
// using ListSelectionEvent.Apply. Return/double-click requests Activate separately.
// Command toggles membership; Shift extends from the keyboard/pointer anchor.
// No drag rearrangement, marquee selection or filesystem behavior is implied.
type CollectionView struct {
	ID, Label   string
	Items       []CollectionItem
	SelectedIDs []string
	Disabled    bool
	ItemWidth   float32
	Change      func(ListSelectionEvent)
	Activate    func(CollectionEvent)
	Style       Style
}

func (f *CollectionView) validate() error {
	if f == nil || strings.TrimSpace(f.ID) == "" || !utf8.ValidString(f.ID) || strings.ContainsRune(f.ID, 0) || len(f.Items) > MaxListBoxOptions {
		return fmt.Errorf("ui: collection requires ID and at most %d items", MaxListBoxOptions)
	}
	if f.ItemWidth != 0 && (!finite(f.ItemWidth) || f.ItemWidth < 80 || f.ItemWidth > 256) {
		return fmt.Errorf("ui: collection item width must be 80..256 logical pixels")
	}
	seen := map[string]bool{}
	for _, item := range f.Items {
		if strings.TrimSpace(item.ID) == "" || !utf8.ValidString(item.ID) || strings.ContainsRune(item.ID, 0) || seen[item.ID] || strings.TrimSpace(item.Label) == "" || !utf8.ValidString(item.Label) || len(item.Label) > 512 || strings.ContainsAny(item.Label, "\x00\r\n") {
			return fmt.Errorf("ui: collection requires unique IDs and nonblank single-line labels of at most 512 bytes")
		}
		seen[item.ID] = true
		if item.Icon != "" {
			if err := validateIcon(Node{Kind: Icon, Icon: item.Icon}); err != nil {
				return err
			}
		}
		if item.Thumbnail != nil && item.Thumbnail.TileCount() == 0 {
			return fmt.Errorf("ui: collection thumbnail must contain pixels")
		}
	}
	selected := map[string]bool{}
	for _, id := range f.SelectedIDs {
		if !seen[id] || selected[id] {
			return fmt.Errorf("ui: collection selection has absent or duplicate ID")
		}
		selected[id] = true
	}
	return nil
}

// ValidateCollectionNode validates the detached native grid metadata.
func ValidateCollectionNode(n Node) error {
	if n.Collection == nil {
		return nil
	}
	c := n.Collection
	if n.Kind != ListBox || !n.Multiple || len(n.Options) != len(c.Items) {
		return fmt.Errorf("ui: collection presentation belongs on a matching multiple ListBox")
	}
	f := CollectionView{ID: n.ID, Items: c.Items, SelectedIDs: n.SelectedIDs, ItemWidth: c.ItemWidth}
	if err := f.validate(); err != nil {
		return err
	}
	for i, item := range c.Items {
		o := n.Options[i]
		if o.ID != item.ID || o.Label != item.Label || o.Disabled != item.Disabled || o.Icon != item.Icon || o.Badge != "" || o.IconImage != nil {
			return fmt.Errorf("ui: collection presentation/options disagree")
		}
	}
	return nil
}

func (f *CollectionView) View() (Node, error) {
	if err := f.validate(); err != nil {
		return Node{}, err
	}
	id := f.ID
	items := slices.Clone(f.Items)
	available := func() bool { return f.ID == id && !f.Disabled && f.validate() == nil }
	eligible := func(target string) bool {
		if !available() {
			return false
		}
		for _, item := range f.Items {
			if item.ID == target {
				return !item.Disabled
			}
		}
		return false
	}
	n := Node{ID: id, Kind: ListBox, Multiple: true, SelectedIDs: slices.Clone(f.SelectedIDs), Disabled: f.Disabled, Style: f.Style, Accessibility: Accessibility{Name: f.Label, Help: "Icon collection. Command-click toggles; Shift extends; arrow keys navigate; Return opens."}}
	if n.Accessibility.Name == "" {
		n.Accessibility.Name = "Items"
	}
	n.Collection = &CollectionLayout{source: f, Items: items, ItemWidth: f.ItemWidth}
	if f.Change != nil {
		n.SelectionChange = func(e ListSelectionEvent) {
			if f.Change == nil || !eligible(e.TargetID) {
				return
			}
			for _, target := range e.IDs {
				if !eligible(target) {
					return
				}
			}
			e.ID = id
			e.IDs = slices.Clone(e.IDs)
			f.Change(e)
		}
	}
	if f.Activate != nil {
		n.Collection.Activate = func(e CollectionEvent) {
			if f.Activate != nil && eligible(e.ItemID) {
				f.Activate(CollectionEvent{ID: id, ItemID: e.ItemID})
			}
		}
	}
	for _, item := range items {
		n.Options = append(n.Options, Option{ID: item.ID, Label: item.Label, Icon: item.Icon, Disabled: item.Disabled})
	}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}

// MoveCollectionItem returns a grid neighbor, clamping at the collection edges
// and skipping disabled cells. Vertical moves stay in the same column; if an
// incomplete final row has no cell in that column, focus stays put. columns is
// the measured, responsive column count.
// dx/dy must describe one cardinal arrow. No selection or model change occurs.
func MoveCollectionItem(items []CollectionItem, current string, columns, dx, dy int) (string, bool) {
	if len(items) == 0 || len(items) > MaxListBoxOptions || columns < 1 || columns > MaxListBoxOptions || (dx == 0) == (dy == 0) || dx < -1 || dx > 1 || dy < -1 || dy > 1 {
		return "", false
	}
	index := slices.IndexFunc(items, func(i CollectionItem) bool { return i.ID == current })
	if index < 0 {
		for _, item := range items {
			if !item.Disabled {
				return item.ID, true
			}
		}
		return "", false
	}
	step := dx + dy*columns
	if dy != 0 {
		for i := index + step; i >= 0 && i < len(items); i += step {
			if !items[i].Disabled {
				return items[i].ID, true
			}
		}
		return current, !items[index].Disabled
	}
	target := max(0, min(len(items)-1, index+step))
	if target == index {
		return current, !items[index].Disabled
	}
	direction := 1
	if step < 0 {
		direction = -1
	}
	for i := target; i >= 0 && i < len(items); i += direction {
		if !items[i].Disabled {
			return items[i].ID, true
		}
	}
	return current, !items[index].Disabled
}
