// Package ui describes retained, keyed UI records using ordinary Go values.
package ui

import (
	"fmt"
	"strings"

	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/virtual"
)

type Kind string

const (
	Column    Kind = "column"
	Row       Kind = "row"
	Label     Kind = "label"
	Button    Kind = "button"
	Extension Kind = "extension"
)

// Node IDs are stable within one window. The grid: prefix is reserved for
// generated grid cell/row/header identities and is refused in descriptions.
// A View returns a fresh description;
// callbacks run on the application owner. Custom holds driver-specific content.
type Node struct {
	ID   string
	Kind Kind
	Text string
	// Editor owns multiline document state and history; one live native binding is supported.
	Editor *editor.Model
	// Grid is caller-owned. One live managed binding per Grid is supported.
	Grid       *virtual.Grid
	GridChange func(GridEvent)
	// TabInserts makes Tab insert a literal tab in MultilineText; otherwise Tab moves focus.
	TabInserts bool
	// SoftWrap wraps visual lines without changing document bytes.
	SoftWrap            bool
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
	models := map[*editor.Model]bool{}
	grids := map[*virtual.Grid]bool{}
	var walk func(Node, int) error
	walk = func(n Node, depth int) error {
		if depth >= 128 {
			return fmt.Errorf("ui: hierarchy exceeds 128 levels")
		}
		if n.ID == "" || seen[n.ID] {
			return fmt.Errorf("ui: empty or duplicate ID %q", n.ID)
		}
		if strings.HasPrefix(n.ID, "grid:") {
			return fmt.Errorf("ui: ID %q uses reserved grid: prefix", n.ID)
		}
		seen[n.ID] = true
		if n.Kind == MultilineText {
			if models[n.Editor] {
				return fmt.Errorf("ui: editor model has multiple bindings")
			}
			models[n.Editor] = true
		}
		if n.Kind == DataGrid {
			if grids[n.Grid] {
				return fmt.Errorf("ui: grid model has multiple bindings")
			}
			grids[n.Grid] = true
		}
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
		case Label, Button, Extension, Checkbox, Selector, Slider, NumberField, TextField, MultilineText, DataGrid:
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
