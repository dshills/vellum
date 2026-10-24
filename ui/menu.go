package ui

import (
	"fmt"
	"strings"
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
	for _, o := range n.Options {
		if o.Separator {
			if o.Label != "" || o.Disabled || o.Checkable || o.Checked {
				return fmt.Errorf("ui: menu separator permits only a stable ID")
			}
		} else if o.Checked && !o.Checkable {
			return fmt.Errorf("ui: checked menu action requires Checkable")
		}
		if len(o.ID) > 1024 || len(o.Label) > 4096 || (!o.Separator && strings.TrimSpace(o.Label) == "") || !utf8.ValidString(o.ID) || !utf8.ValidString(o.Label) || strings.ContainsRune(o.ID, 0) || strings.ContainsRune(o.Label, 0) {
			return fmt.Errorf("ui: menu action requires bounded UTF-8 identity and label without NUL")
		}
	}
	return nil
}

// MenuEnabled reports whether the node has an enabled action handler. Ancestor
// disabling is handled by focus resolution and the driver.
func MenuEnabled(n Node) bool {
	if n.Kind != MenuButton || n.Disabled || n.Change == nil {
		return false
	}
	for _, o := range n.Options {
		if !o.Disabled && !o.Separator {
			return true
		}
	}
	return false
}
