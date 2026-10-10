// Package ui describes retained, keyed UI records using ordinary Go values.
package ui

import "fmt"

type Kind string

const (
	Column    Kind = "column"
	Row       Kind = "row"
	Label     Kind = "label"
	Button    Kind = "button"
	Extension Kind = "extension"
)

// Node IDs are stable within one window. A View returns a fresh description;
// callbacks run on the application owner. Custom holds driver-specific content.
type Node struct {
	ID                  string
	Kind                Kind
	Text                string
	Disabled            bool
	Grow                bool
	MinWidth, MinHeight float32
	Activate            func()
	FileDrop            func(FileDropEvent) bool
	Children            []Node
	Custom              any
	Style               Style
	Checked             bool
	Number              float32
	Numeric             Numeric
	Options             []Option
	SelectedID          string
	Change              func(ChangeEvent)
	Accessibility       Accessibility
	Focus               FocusRequest
	Open                bool
	Dismiss             func()
}

func Validate(root Node) error {
	seen := map[string]bool{}
	var walk func(Node, int) error
	walk = func(n Node, depth int) error {
		if depth >= 128 {
			return fmt.Errorf("ui: hierarchy exceeds 128 levels")
		}
		if n.ID == "" || seen[n.ID] {
			return fmt.Errorf("ui: empty or duplicate ID %q", n.ID)
		}
		seen[n.ID] = true
		if depth > 0 && n.Focus.ID != "" {
			return fmt.Errorf("ui: focus requests belong on the root")
		}
		if err := ValidateControl(n); err != nil {
			return err
		}
		if _, err := ResolveStyle(n, Theme{}); err != nil {
			return err
		}
		switch n.Kind {
		case Column, Row, Tabs, Scroll, Dialog:
		case Label, Button, Extension, Checkbox, Selector, Slider, NumberField, TextField:
			if len(n.Children) != 0 {
				return fmt.Errorf("ui: leaf %q has children", n.ID)
			}
		default:
			return fmt.Errorf("ui: unknown kind %q", n.Kind)
		}
		for _, c := range n.Children {
			if err := walk(c, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, 0); err != nil {
		return err
	}
	return nil
}
