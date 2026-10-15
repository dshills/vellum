package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// MaxListBoxOptions bounds an always-visible, single-selection list. Larger
// collections use virtual data views; Grid supports multi-selection.
const MaxListBoxOptions = 256

func validateListBox(n Node) error {
	if len(n.Options) > MaxListBoxOptions || n.Activate != nil {
		return fmt.Errorf("ui: list box requires at most %d options and uses Change for selection", MaxListBoxOptions)
	}
	seen := make(map[string]bool, len(n.Options))
	selected := n.SelectedID == ""
	for _, o := range n.Options {
		if strings.TrimSpace(o.ID) == "" || !utf8.ValidString(o.ID) || strings.ContainsRune(o.ID, 0) || seen[o.ID] || strings.TrimSpace(o.Label) == "" || !utf8.ValidString(o.Label) {
			return fmt.Errorf("ui: list box requires unique nonblank UTF-8 item IDs and labels")
		}
		seen[o.ID] = true
		selected = selected || o.ID == n.SelectedID
	}
	if !selected {
		return fmt.Errorf("ui: list box selected ID %q is absent", n.SelectedID)
	}
	return nil
}

// MoveListBoxOption moves by enabled items, clamping at either end. An absent
// cursor starts at the first/last enabled item. delta is nonzero and bounded by
// MaxListBoxOptions; it also supports page-sized movement. No model edit occurs.
func MoveListBoxOption(n Node, current string, delta int) (string, bool) {
	if n.Kind != ListBox || n.Disabled || delta == 0 || delta < -MaxListBoxOptions || delta > MaxListBoxOptions || ValidateControl(n) != nil {
		return "", false
	}
	index := -1
	for i, o := range n.Options {
		if o.ID == current {
			index = i
			break
		}
	}
	step := 1
	if delta < 0 {
		step = -1
		delta = -delta
		if index < 0 {
			index = len(n.Options)
		}
	}
	last := ""
	if index >= 0 && index < len(n.Options) && !n.Options[index].Disabled {
		last = n.Options[index].ID
	}
	for i := index + step; i >= 0 && i < len(n.Options); i += step {
		if n.Options[i].Disabled {
			continue
		}
		last = n.Options[i].ID
		delta--
		if delta == 0 {
			break
		}
	}
	return last, last != ""
}
