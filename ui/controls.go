package ui

import (
	"fmt"
	"math"
	"strings"
)

const (
	SplitPane        Kind = "split-pane"
	Toolbar          Kind = "toolbar"
	StatusBar        Kind = "status-bar"
	FormGroup        Kind = "form-group"
	Disclosure       Kind = "disclosure"
	Checkbox         Kind = "checkbox"
	Switch           Kind = "switch"
	ToggleButton     Kind = "toggle-button"
	ProgressBar      Kind = "progress-bar"
	ProgressRing     Kind = "progress-ring"
	ProgressGauge    Kind = "progress-gauge"
	Separator        Kind = "separator"
	Selector         Kind = "selector"
	MenuButton       Kind = "menu-button"
	RadioGroup       Kind = "radio-group"
	SegmentedControl Kind = "segmented-control"
	Slider           Kind = "slider"
	NumberField      Kind = "number-field"
	TextField        Kind = "text-field"
	MultilineText    Kind = "multiline-text"
	Tabs             Kind = "tabs"
	Scroll           Kind = "scroll"
	Dialog           Kind = "dialog"
	Canvas2D         Kind = "canvas-2d"
)

// Option has a stable domain ID independent of its displayed label. For Tabs,
// each option ID corresponds to the child page ID, in the same order.
type Option struct {
	ID, Label string
	Disabled  bool
	// MenuButton alone supports passive separators and controlled check actions.
	// Separators retain a stable ID, but have no label, value or input behavior.
	// Closable belongs only on Tabs and requires Node.TabClose.
	Closable  bool
	Separator bool
	Checkable bool
	Checked   bool
}

// Numeric uses finite float32 values supported by the native widgets. Step is
// the keyboard increment; pointer and typed values need not be step multiples.
type Numeric struct{ Min, Max, Step float32 }

// Orientation sets display/choice direction. Empty means horizontal except for
// RadioGroup, which defaults to vertical. Progress/separators are not focusable.
type Orientation string

const (
	Horizontal Orientation = "horizontal"
	Vertical   Orientation = "vertical"
)

// ChangeEvent requests a model edit. The relevant payload is Checked for
// Checkbox/Switch/ToggleButton/Disclosure, Number for Slider/NumberField, Text for TextField, and SelectedID
// for Selector/Tabs/RadioGroup/SegmentedControl/MenuButton. MenuButton check actions
// also use Checked for the proposed boolean. SplitPane uses Number for the
// first pane's fraction. Controlled nodes reflect the accepted value on redraw.
type ChangeEvent struct {
	ID               string
	Kind             Kind
	Checked          bool
	Number           float32
	Text, SelectedID string
}

// TabCloseEvent proposes closing one stable page. The application accepts by
// removing its option/page and selecting a remaining page in the next view.
// ID is the Tabs node ID; PageID is the option/page ID. No native deletion runs.
type TabCloseEvent struct{ ID, PageID string }

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
	case Row, Column, Tabs, Scroll, Dialog, Toolbar, StatusBar, FormGroup, SplitPane, Disclosure:
		return true
	}
	return false
}
func IsFocusable(kind Kind) bool {
	switch kind {
	case Button, Canvas2D, Checkbox, Switch, ToggleButton, MenuButton, Selector, RadioGroup, SegmentedControl, Slider, NumberField, TextField, MultilineText, DataGrid, DataTree, Tabs, SplitPane, Disclosure:
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
	if n.Kind == Canvas2D {
		if n.Canvas == nil || n.Canvas.Validate() != nil || !finite(n.CanvasGrid) || n.CanvasGrid < 0 || n.CanvasGrid > 1e6 {
			return fmt.Errorf("ui: Canvas2D requires a valid scene and bounded grid spacing")
		}
	} else if n.Canvas != nil || n.CanvasGrid != 0 || n.PlacementTarget != nil {
		return fmt.Errorf("ui: canvas fields belong only on Canvas2D")
	}
	if n.PlacementSource != nil && (n.Kind != Button || !n.PlacementSource.Valid() || n.Activate == nil) {
		return fmt.Errorf("ui: placement source requires a valid focusable Button and keyboard Activate path")
	}
	if n.PlacementTarget != nil && !n.PlacementTarget.Valid() {
		return fmt.Errorf("ui: invalid placement target")
	}
	if n.Kind != Tabs && n.TabClose != nil {
		return fmt.Errorf("ui: TabClose belongs only on tabs")
	}
	for _, option := range n.Options {
		if option.Closable && (n.Kind != Tabs || n.TabClose == nil) {
			return fmt.Errorf("ui: closable tab requires Tabs and TabClose")
		}
	}
	if n.Kind == Disclosure && (strings.TrimSpace(n.Text) == "" || n.Activate != nil) {
		return fmt.Errorf("ui: disclosure requires a heading and uses Checked/Change for expansion")
	}
	if err := validateImage(n); err != nil {
		return err
	}
	if err := validateIcon(n); err != nil {
		return err
	}
	if n.Kind == SplitPane {
		bounds := SplitRange(n.Numeric)
		if len(n.Children) != 2 || n.Activate != nil || bounds.Min <= 0 || bounds.Max >= 1 || bounds.Validate(n.Number) != nil {
			return fmt.Errorf("ui: split pane requires two children, a finite fraction within ordered bounds strictly inside 0..1, positive step, and no Activate callback")
		}
	}
	if (n.Kind == Toolbar || n.Kind == StatusBar || n.Kind == FormGroup) && (n.Change != nil || n.Activate != nil) {
		return fmt.Errorf("ui: container callbacks belong on their interactive children")
	}
	if n.Orientation != "" && n.Kind != ProgressBar && n.Kind != Separator && n.Kind != RadioGroup && n.Kind != SegmentedControl && n.Kind != Toolbar && n.Kind != SplitPane {
		return fmt.Errorf("ui: Orientation belongs only on progress, separator, choice, toolbar and split controls")
	}
	if n.Orientation != "" && n.Orientation != Horizontal && n.Orientation != Vertical {
		return fmt.Errorf("ui: invalid display orientation %q", n.Orientation)
	}
	if n.Kind == ProgressBar || n.Kind == ProgressRing || n.Kind == ProgressGauge {
		if !finite(n.Number) || n.Number < 0 || n.Number > 1 {
			return fmt.Errorf("ui: progress requires a finite fraction within 0..1")
		}
	}
	if (n.Kind == ProgressBar || n.Kind == ProgressRing || n.Kind == ProgressGauge || n.Kind == Separator) && (n.Change != nil || n.Activate != nil) {
		return fmt.Errorf("ui: display-only controls cannot have input callbacks")
	}
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
	if n.Kind == RadioGroup || n.Kind == SegmentedControl {
		if len(n.Options) == 0 || len(n.Options) > MaxChoiceOptions || n.SelectedID == "" {
			return fmt.Errorf("ui: choice group requires 1..%d options and one selected ID", MaxChoiceOptions)
		}
	}
	if n.Kind == MenuButton {
		if err := validateMenuButton(n); err != nil {
			return err
		}
	}
	if n.Kind != MenuButton {
		for _, o := range n.Options {
			if o.Separator || o.Checkable || o.Checked {
				return fmt.Errorf("ui: separator/check action metadata belongs only on menu options")
			}
		}
	}
	if n.Kind == Selector || n.Kind == Tabs || n.Kind == RadioGroup || n.Kind == SegmentedControl || n.Kind == MenuButton {
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
			if n.Kind == MenuButton && n.Change == nil {
				return n, false
			}
			if n.Kind == Tabs || n.Kind == RadioGroup || n.Kind == SegmentedControl || n.Kind == MenuButton {
				for _, option := range n.Options {
					if !option.Disabled && !option.Separator {
						return n, true
					}
				}
				return n, false
			}
			return n, true
		}
		if n.Kind == Dialog && !n.Open || n.Kind == Disclosure && !n.Checked {
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
