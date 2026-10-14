package ui

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// MaxBreadcrumbItems bounds the retained path. Use a data view for larger lists.
const MaxBreadcrumbItems = 64

// BreadcrumbItem identifies one location independently of its display label.
// The final item is the current location and has no navigation action.
type BreadcrumbItem struct {
	ID, Label string
	Disabled  bool
}

// BreadcrumbEvent proposes navigation to an ancestor's stable domain ID.
// The application accepts by replacing Items in its next view, or rejects by
// leaving the accepted path unchanged. Index and display text are not identity.
type BreadcrumbEvent struct{ ID, ItemID string }

// Breadcrumb is an owner-held controlled path component. View composes ordinary
// managed controls; it owns no window, filesystem access, or background work.
// Do not mutate or render one instance concurrently from independent owners.
// Nil Navigate produces read-only labels. Disabled prevents all navigation.
type Breadcrumb struct {
	ID, Label string
	Items     []BreadcrumbItem
	Disabled  bool
	Navigate  func(BreadcrumbEvent)
	Style     Style
}

// BreadcrumbItemID returns the stable managed target for a domain item. The
// root owns root+"/breadcrumb/"; ordinary Node validation detects collisions.
// Encoding makes arbitrary domain IDs (including slashes) unambiguous.
func BreadcrumbItemID(root, item string) string {
	return root + "/breadcrumb/item/" + base64.RawURLEncoding.EncodeToString([]byte(item))
}

func (b *Breadcrumb) validate() error {
	if b == nil || b.ID == "" || len(b.Items) == 0 || len(b.Items) > MaxBreadcrumbItems {
		return fmt.Errorf("ui: breadcrumb requires an ID and 1..%d items", MaxBreadcrumbItems)
	}
	seen := make(map[string]bool, len(b.Items))
	for _, item := range b.Items {
		if item.ID == "" || seen[item.ID] || strings.TrimSpace(item.Label) == "" {
			return fmt.Errorf("ui: breadcrumb requires unique nonempty item IDs and nonblank labels")
		}
		seen[item.ID] = true
	}
	return b.Style.Validate()
}

// View returns detached nodes and captures the path's stable IDs. A queued
// proposal is refused if its root/path has changed, its target is now disabled,
// or its callback is removed. Label edits do not change location identity.
// The current location is a nonfocusable label; ancestors use quiet buttons.
// Toolbar's normal overflow/focus behavior keeps long paths reachable.
func (b *Breadcrumb) View() (Node, error) {
	if err := b.validate(); err != nil {
		return Node{}, err
	}
	label := b.Label
	if strings.TrimSpace(label) == "" {
		label = "Location"
	}
	style := b.Style
	if !style.Gap.Set {
		style.Gap = Some(float32(4))
	}
	root := Node{ID: b.ID, Kind: Toolbar, Disabled: b.Disabled, Style: style, Accessibility: Accessibility{Name: label}}
	ids := make([]string, len(b.Items))
	for i, item := range b.Items {
		ids[i] = item.ID
	}
	rootID := b.ID
	for i, item := range b.Items {
		if i > 0 {
			root.Children = append(root.Children, Node{ID: b.ID + "/breadcrumb/separator/" + base64.RawURLEncoding.EncodeToString([]byte(item.ID)), Kind: Icon, Icon: SymbolRight})
		}
		n := Node{ID: BreadcrumbItemID(b.ID, item.ID), Kind: Label, Text: item.Label, Disabled: item.Disabled, Accessibility: Accessibility{Name: item.Label}}
		if i == len(b.Items)-1 {
			n.Accessibility.Help = "Current location"
		} else if b.Navigate != nil {
			n.Kind = Button
			n.Appearance = ButtonQuiet
			n.Accessibility.Help = "Navigate to this location"
			target := item.ID
			n.Activate = func() {
				if b.validate() != nil || b.ID != rootID || b.Disabled || b.Navigate == nil || len(b.Items) != len(ids) {
					return
				}
				for j, id := range ids {
					if b.Items[j].ID != id {
						return
					}
				}
				for _, current := range b.Items[:len(b.Items)-1] {
					if current.ID == target && !current.Disabled {
						b.Navigate(BreadcrumbEvent{ID: rootID, ItemID: target})
						return
					}
				}
			}
		}
		root.Children = append(root.Children, n)
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
