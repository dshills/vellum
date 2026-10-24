package ui

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MenuButton opens a transient action menu. Options identify actions, not selected
// values: SelectedID stays empty. Change.SelectedID reports one enabled action on
// the owner. Reconciliation and disabled ancestry revoke an older open menu.
const MaxMenuOptions = 64

func validateMenuButton(n Node) error {
	if len(n.Options) < 1 || len(n.Options) > MaxMenuOptions || n.SelectedID != "" || n.Activate != nil || n.Open || n.Dismiss != nil {
		return fmt.Errorf("ui: menu button requires 1..64 actions, no selected value, and Change for action requests")
	}
	if strings.TrimSpace(n.Text) == "" && ((n.Icon == "" && n.IconImage == nil) || strings.TrimSpace(n.Accessibility.Name) == "") {
		return fmt.Errorf("ui: menu button requires text or a named icon")
	}
	return validateMenuTree(n.Options, 1, map[string]bool{}, new(int))
}

// MaxMenuDepth includes the root popup. The complete tree is limited to 64 items.
const MaxMenuDepth = 4

// MenuBranch snapshots a submenu. Retain the returned pointer across Views;
// replacing it deliberately retires the open chain. Items returns a copy.
type MenuBranch struct{ items []Option }

func NewMenuBranch(items []Option) (*MenuBranch, error) {
	if err := validateMenuTree(items, 1, map[string]bool{}, new(int)); err != nil {
		return nil, err
	}
	return &MenuBranch{items: slices.Clone(items)}, nil
}
func (b *MenuBranch) Items() []Option {
	if b == nil {
		return nil
	}
	return slices.Clone(b.items)
}
func validateMenuTree(items []Option, depth int, seen map[string]bool, count *int) error {
	if len(items) == 0 || depth > MaxMenuDepth {
		return fmt.Errorf("ui: empty or too-deep menu")
	}
	for _, o := range items {
		*count++
		if *count > MaxMenuOptions || o.ID == "" || seen[o.ID] || len(o.ID) > 1024 || len(o.Label) > 4096 || !utf8.ValidString(o.ID) || !utf8.ValidString(o.Label) || strings.ContainsRune(o.ID, 0) || strings.ContainsRune(o.Label, 0) {
			return fmt.Errorf("ui: invalid, duplicate or excessive menu identity/label")
		}
		seen[o.ID] = true
		if o.Closable || o.Badge != "" || len(o.Shortcut) > 128 || !utf8.ValidString(o.Shortcut) || strings.IndexFunc(o.Shortcut, unicode.IsControl) >= 0 {
			return fmt.Errorf("ui: invalid menu metadata")
		}
		if o.Separator {
			if o.Label != "" || o.Disabled || o.Checkable || o.Checked || o.Icon != "" || o.IconImage != nil || o.Branch != nil || o.Shortcut != "" {
				return fmt.Errorf("ui: separator permits only identity")
			}
			continue
		}
		if strings.TrimSpace(o.Label) == "" || o.Checked && !o.Checkable {
			return fmt.Errorf("ui: invalid menu label/check")
		}
		if err := validateIcon(Node{Kind: Button, Text: o.Label, Icon: o.Icon, IconImage: o.IconImage}); err != nil {
			return err
		}
		if o.Branch != nil {
			if o.Checkable || o.Checked || o.Shortcut != "" {
				return fmt.Errorf("ui: submenu cannot be a check or shortcut action")
			}
			if err := validateMenuTree(o.Branch.items, depth+1, seen, count); err != nil {
				return err
			}
		}
	}
	return nil
}

// MenuAction resolves an enabled leaf, including disabled ancestry.
func MenuAction(items []Option, id string) (Option, bool) {
	for _, o := range items {
		if o.Disabled || o.Separator {
			continue
		}
		if o.Branch != nil {
			if leaf, ok := MenuAction(o.Branch.items, id); ok {
				return leaf, true
			}
		} else if o.ID == id {
			return o, true
		}
	}
	return Option{}, false
}

// MenuEnabled reports whether an enabled leaf can reach a handler.
func MenuEnabled(n Node) bool {
	if n.Kind != MenuButton || n.Disabled || n.Change == nil {
		return false
	}
	var enabled func([]Option) bool
	enabled = func(items []Option) bool {
		for _, o := range items {
			if !o.Disabled && !o.Separator {
				if o.Branch == nil || enabled(o.Branch.items) {
					return true
				}
			}
		}
		return false
	}
	return enabled(n.Options)
}
