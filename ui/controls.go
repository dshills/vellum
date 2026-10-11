package ui

import (
	"fmt"
	"math"
)

const (
	Checkbox      Kind = "checkbox"
	Selector      Kind = "selector"
	Slider        Kind = "slider"
	NumberField   Kind = "number-field"
	TextField     Kind = "text-field"
	MultilineText Kind = "multiline-text"
	Tabs          Kind = "tabs"
	Scroll        Kind = "scroll"
	Dialog        Kind = "dialog"
)

// Option has a stable domain ID independent of its displayed label. For Tabs,
// each option ID corresponds to the child page ID, in the same order.
type Option struct {
	ID, Label string
	Disabled  bool
}

// Numeric uses finite float32 values supported by the native widgets. Step is
// the keyboard increment; pointer and typed values need not be step multiples.
type Numeric struct{ Min, Max, Step float32 }

// ChangeEvent requests a model edit. The relevant payload is Checked for
// Checkbox, Number for Slider/NumberField, Text for TextField, and SelectedID
// for Selector/Tabs. Controlled nodes reflect the accepted value on redraw.
type ChangeEvent struct {
	ID               string
	Kind             Kind
	Checked          bool
	Number           float32
	Text, SelectedID string
}

// Accessibility carries semantic labels independently of display text. Driver
// publication to platform accessibility APIs requires separate certification.
type Accessibility struct{ Name, Help string }

// FocusRequest appears only on the root. Change Sequence to request the same ID
// again; ordinary redraws do not repeatedly steal keyboard focus.
type FocusRequest struct {
	ID       string
	Sequence uint64
}

func IsContainer(kind Kind) bool {
	switch kind {
	case Row, Column, Tabs, Scroll, Dialog:
		return true
	}
	return false
}
func IsFocusable(kind Kind) bool {
	switch kind {
	case Button, Checkbox, Selector, Slider, NumberField, TextField, MultilineText, DataGrid, DataTree, Tabs:
		return true
	}
	return false
}
func finite(value float32) bool { return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) }
func (r Numeric) Validate(value float32) error {
	span := float64(r.Max) - float64(r.Min)
	if !finite(r.Min) || !finite(r.Max) || !finite(r.Step) || !finite(value) || span <= 0 || span > math.MaxFloat32 || r.Step <= 0 || float64(r.Step) > span || value < r.Min || value > r.Max {
		return fmt.Errorf("ui: numeric control requires finite ordered bounds, positive step within range, and value within bounds")
	}
	return nil
}

// ClampNumeric rejects nonfinite native values and clamps finite requests.
func ClampNumeric(value float32, r Numeric) (float32, bool) {
	if !finite(value) {
		return 0, false
	}
	if value < r.Min {
		value = r.Min
	}
	if value > r.Max {
		value = r.Max
	}
	return value, true
}

func ValidateControl(n Node) error {
	if n.Kind == DataTree {
		if n.Tree == nil {
			return fmt.Errorf("ui: data tree requires an Outline")
		}
		if _, err := n.Tree.State(); err != nil {
			return fmt.Errorf("ui: invalid outline: %w", err)
		}
	} else if n.Tree != nil || n.TreeChange != nil || n.TreeDisclosure != nil {
		return fmt.Errorf("ui: Tree/TreeChange/TreeDisclosure belong only on data trees")
	}

	if n.Kind == DataGrid {
		if n.Grid == nil {
			return fmt.Errorf("ui: data grid requires a Grid")
		}
		if _, err := n.Grid.State(); err != nil {
			return fmt.Errorf("ui: invalid grid: %w", err)
		}
	} else if n.Grid != nil || n.GridChange != nil || n.GridCommit != nil {
		return fmt.Errorf("ui: Grid/GridChange/GridCommit belong only on data grids")
	}

	if n.Kind == MultilineText && (n.Editor == nil || n.Editor.Validate() != nil) {
		return fmt.Errorf("ui: multiline node %q requires an editor Model", n.ID)
	}
	if n.Kind != MultilineText && n.Editor != nil {
		return fmt.Errorf("ui: Editor belongs only on multiline nodes")
	}
	if n.Kind != MultilineText && n.SoftWrap {
		return fmt.Errorf("ui: SoftWrap belongs only on multiline nodes")
	}
	if n.Kind != MultilineText && n.TabInserts {
		return fmt.Errorf("ui: TabInserts belongs only on multiline nodes")
	}
	if n.Kind == Slider || n.Kind == NumberField {
		if err := n.Numeric.Validate(n.Number); err != nil {
			return fmt.Errorf("ui: node %q: %w", n.ID, err)
		}
	}
	if n.Kind == Selector || n.Kind == Tabs {
		seen := map[string]bool{}
		selected := n.SelectedID == ""
		if n.Kind == Tabs && (len(n.Options) == 0 || len(n.Options) != len(n.Children)) {
			return fmt.Errorf("ui: tabs %q requires one option per child page", n.ID)
		}
		for i, o := range n.Options {
			if o.ID == "" || seen[o.ID] {
				return fmt.Errorf("ui: node %q has empty/duplicate option ID %q", n.ID, o.ID)
			}
			seen[o.ID] = true
			selected = selected || o.ID == n.SelectedID
			if n.Kind == Tabs && n.Children[i].ID != o.ID {
				return fmt.Errorf("ui: tabs %q option order must match child IDs", n.ID)
			}
		}
		if !selected {
			return fmt.Errorf("ui: node %q selected ID %q is absent", n.ID, n.SelectedID)
		}
		if n.Kind == Tabs && n.SelectedID == "" {
			return fmt.Errorf("ui: tabs %q requires a selected page", n.ID)
		}
	}
	return nil
}

// FocusTarget resolves a request through visible, enabled descendants. Tabs
// only expose their selected page; closed dialogs expose no descendants.
func FocusTarget(root Node, id string) (Node, bool) {
	var find func(Node, bool) (Node, bool)
	find = func(n Node, disabled bool) (Node, bool) {
		disabled = disabled || n.Disabled
		if n.ID == id {
			if disabled || !IsFocusable(n.Kind) {
				return n, false
			}
			if n.Kind == Tabs {
				for _, option := range n.Options {
					if !option.Disabled {
						return n, true
					}
				}
				return n, false
			}
			return n, true
		}
		if n.Kind == Dialog && !n.Open {
			return Node{}, false
		}
		for _, c := range n.Children {
			if n.Kind == Tabs && c.ID != n.SelectedID {
				continue
			}
			if found, ok := find(c, disabled); ok {
				return found, true
			}
		}
		return Node{}, false
	}
	return find(root, false)
}
